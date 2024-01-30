package gitSensor

import "encoding/json"

type GitSensorResponse struct {
	Code   int                  `json:"code,omitempty"`
	Status string               `json:"status,omitempty"`
	Result json.RawMessage      `json:"result,omitempty"`
	Errors []*GitSensorApiError `json:"errors,omitempty"`
}

type GitSensorApiError struct {
	HttpStatusCode    int    `json:"-"`
	Code              string `json:"code,omitempty"`
	InternalMessage   string `json:"internalMessage,omitempty"`
	UserMessage       string `json:"userMessage,omitempty"`
	UserDetailMessage string `json:"userDetailMessage,omitempty"`
}

type GitChanges struct {
	Commits   []*Commit
	FileStats FileStats
}
type FileStats []FileStat

// FileStat stores the status of changes in content of a file.
type FileStat struct {
	Name     string
	Addition int
	Deletion int
}

type ReleaseChangesRequest struct {
	PipelineMaterialId int    `json:"pipelineMaterialId"`
	OldCommit          string `json:"oldCommit"`
	NewCommit          string `json:"newCommit"`
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
