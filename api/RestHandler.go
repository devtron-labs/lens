package api

import (
	"github.com/devtron-labs/lens/pkg"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type RestHandler interface {
	GetDeploymentMetrics(w http.ResponseWriter, r *http.Request)
	ProcessDeploymentEvent(w http.ResponseWriter, r *http.Request)
	ResetApplication(w http.ResponseWriter, r *http.Request)
}

func NewRestHandlerImpl(logger *zap.SugaredLogger,
	deploymentMetricService pkg.DeploymentMetricService,
	ingestionService pkg.IngestionService) *RestHandlerImpl {
	return &RestHandlerImpl{logger: logger,
		deploymentMetricService: deploymentMetricService,
		ingestionService:        ingestionService}
}

type RestHandlerImpl struct {
	logger                  *zap.SugaredLogger
	deploymentMetricService pkg.DeploymentMetricService
	ingestionService        pkg.IngestionService
}
type Response struct {
	Code   int         `json:"code,omitempty"`
	Status string      `json:"status,omitempty"`
	Result interface{} `json:"result,omitempty"`
	Errors []*ApiError `json:"errors,omitempty"`
}
type ApiError struct {
	HttpStatusCode    int         `json:"-"`
	Code              string      `json:"code,omitempty"`
	InternalMessage   string      `json:"internalMessage,omitempty"`
	UserMessage       interface{} `json:"userMessage,omitempty"`
	UserDetailMessage string      `json:"userDetailMessage,omitempty"`
}

func (impl RestHandlerImpl) writeJsonResp(w http.ResponseWriter, err error, respBody interface{}, status int) {
	response := Response{}
	response.Code = status
	response.Status = http.StatusText(status)
	if err == nil {
		response.Result = respBody
	} else {
		apiErr := &ApiError{}
		apiErr.Code = "000" // 000=unknown
		apiErr.InternalMessage = err.Error()
		apiErr.UserMessage = respBody
		response.Errors = []*ApiError{apiErr}

	}
	b, err := json.Marshal(response)
	if err != nil {
		impl.logger.Error("error in marshaling err object", err)
		status = 500
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(b)
}

/*
func (handler RestHandlerImpl) SaveGitProvider(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	gitProvider := &sql.GitProvider{}
	err := decoder.Decode(gitProvider)
	if err != nil {
		handler.logger.Error(err)
		handler.writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Infow("update gitProvider request ", "req", gitProvider)
	res, err := handler.repositoryManager.SaveGitProvider(gitProvider)
	if err != nil {
		handler.writeJsonResp(w, err, nil, http.StatusBadRequest)
	} else {
		handler.writeJsonResp(w, err, res, http.StatusOK)
	}
}*/

func (impl *RestHandlerImpl) GetDeploymentMetrics(w http.ResponseWriter, r *http.Request) {
	//decoder := json.NewDecoder(r.Body)
	v := r.URL.Query()
	impl.logger.Infow("metrics request", "req", v)
	metricRequest := &pkg.MetricRequest{}
	if v.Get("env_id") != "" {
		envId, err := strconv.Atoi(v.Get("env_id"))
		if err != nil {
			impl.writeJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		metricRequest.EnvId = envId
	}
	if v.Get("app_id") != "" {
		appId, err := strconv.Atoi(v.Get("app_id"))
		if err != nil {
			impl.writeJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		metricRequest.AppId = appId
	}
	if v.Get("from") != "" {
		from := v.Get("from")
		metricRequest.From = from
	}
	if v.Get("to") != "" {
		to := v.Get("to")
		metricRequest.To = to
	}

	//err := decoder.Decode(metricRequest)
	//if err != nil {
	//	impl.logger.Error(err)
	//	impl.writeJsonResp(w, err, nil, http.StatusBadRequest)
	//	return
	//}
	metrics, err := impl.deploymentMetricService.GetDeploymentMetrics(metricRequest)
	impl.logger.Infof("metrics %+v", metrics)
	impl.writeJsonResp(w, err, metrics, 200)
}

func (impl *RestHandlerImpl) ProcessDeploymentEvent(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	deploymentEvent := &pkg.DeploymentEvent{}
	err := decoder.Decode(deploymentEvent)
	if err != nil {
		impl.logger.Error(err)
		impl.writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	release, err := impl.ingestionService.ProcessDeploymentEvent(deploymentEvent)
	impl.logger.Infow("release saved", "release", release)
	impl.writeJsonResp(w, err, release, 200)
}

type ResetRequest struct {
	AppId         int `json:"appId"`
	EnvironmentId int `json:"environmentId"`
}

func (impl *RestHandlerImpl) ResetApplication(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	resetRequest := &ResetRequest{}
	err := decoder.Decode(resetRequest)
	if err != nil {
		impl.logger.Error(err)
		impl.writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	status, err := impl.ingestionService.CleanAppDataForEnvironment(resetRequest.AppId, resetRequest.EnvironmentId)
	impl.logger.Infow("save", "status", status)
	impl.writeJsonResp(w, err, status, 200)
}
