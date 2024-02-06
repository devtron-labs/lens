package gitSensor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"time"

	"github.com/caarlos0/env"
	"go.uber.org/zap"
)

type GitSensorClient interface {
	GetReleaseChanges(request *ReleaseChangesRequest) (*GitChanges, error)
}

type GitSensorClientImpl struct {
	httpClient *http.Client
	logger     *zap.SugaredLogger
	baseUrl    *url.URL
}

func GetGitSensorConfig() (*GitSensorConfig, error) {
	cfg := &GitSensorConfig{}
	err := env.Parse(cfg)
	return cfg, err
}

// ----------------------impl
type GitSensorConfig struct {
	Url     string `env:"GIT_SENSOR_URL" envDefault:"http://localhost:9999"`
	Timeout int    `env:"GIT_SENSOR_TIMEOUT" envDefault:"0"` // in seconds
}

type StatusCode int

func (code StatusCode) IsSuccess() bool {
	return code >= 200 && code <= 299
}

type ClientRequest struct {
	Method       string
	Path         string
	RequestBody  interface{}
	ResponseBody interface{}
}

func (session *GitSensorClientImpl) doRequest(clientRequest *ClientRequest) (resBody []byte, resCode *StatusCode, err error) {
	if clientRequest.ResponseBody == nil {
		return nil, nil, fmt.Errorf("responce body cant be nil")
	}
	if reflect.ValueOf(clientRequest.ResponseBody).Kind() != reflect.Ptr {
		return nil, nil, fmt.Errorf("responsebody non pointer")
	}
	rel, err := session.baseUrl.Parse(clientRequest.Path)
	if err != nil {
		return nil, nil, err
	}
	var body io.Reader
	if clientRequest.RequestBody != nil {
		if req, err := json.Marshal(clientRequest.RequestBody); err != nil {
			return nil, nil, err
		} else {
			session.logger.Infow("argo req with body", "body", string(req))
			body = bytes.NewBuffer(req)
		}

	}
	httpReq, err := http.NewRequest(clientRequest.Method, rel.String(), body)
	if err != nil {
		return nil, nil, err
	}
	httpRes, err := session.httpClient.Do(httpReq)
	if err != nil {
		return nil, nil, err
	}
	defer httpRes.Body.Close()
	resBody, err = ioutil.ReadAll(httpRes.Body)
	if err != nil {
		session.logger.Errorw("error in git communication ", "err", err)
		return nil, nil, err
	}
	status := StatusCode(httpRes.StatusCode)
	if status.IsSuccess() {
		apiRes := &GitSensorResponse{}
		err = json.Unmarshal(resBody, apiRes)
		if apiStatus := StatusCode(apiRes.Code); apiStatus.IsSuccess() {
			err = json.Unmarshal(apiRes.Result, clientRequest.ResponseBody)
			return resBody, &apiStatus, err
		} else {
			session.logger.Infow("api err", "res", apiRes.Errors)
			return resBody, &apiStatus, fmt.Errorf("err in api res")
		}
	} else {
		session.logger.Infow("api err", "res", string(resBody))
		return resBody, &status, fmt.Errorf("res not success, code: %d ", status)
	}
	return resBody, &status, err
}

func NewGitSensorSession(config *GitSensorConfig, logger *zap.SugaredLogger) (session *GitSensorClientImpl, err error) {
	baseUrl, err := url.Parse(config.Url)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: time.Duration(config.Timeout)}
	return &GitSensorClientImpl{httpClient: client, logger: logger, baseUrl: baseUrl}, nil
}

func (session GitSensorClientImpl) GetReleaseChanges(req *ReleaseChangesRequest) (changes *GitChanges, err error) {
	changes = new(GitChanges)
	request := &ClientRequest{ResponseBody: changes, Method: "POST", RequestBody: req, Path: "release/changes"}
	_, _, err = session.doRequest(request)
	return changes, err
}
