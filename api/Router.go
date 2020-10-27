package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
)

type MuxRouter struct {
	logger      *zap.SugaredLogger
	Router      *mux.Router
	restHandler RestHandler
}

func NewMuxRouter(logger *zap.SugaredLogger, restHandler RestHandler) *MuxRouter {
	return &MuxRouter{logger: logger, Router: mux.NewRouter(), restHandler: restHandler}
}

func (r MuxRouter) Init() {
	r.Router.StrictSlash(true)
	//r.Router.Handle("/metrics", promhttp.Handler())
	r.Router.Path("/health").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(200)
		response := Response{}
		response.Code = 200
		response.Result = "OK"
		b, err := json.Marshal(response)
		if err != nil {
			b = []byte("OK")
			r.logger.Errorw("Unexpected error in apiError", "err", err)
		}
		_, _ = writer.Write(b)
	})

	r.Router.Path("/deployment-metrics").HandlerFunc(r.restHandler.GetDeploymentMetrics).
		Queries("app_id", "{app_id}", "env_id", "{env_id}", "from", "{from}", "to", "{to}").
		Methods("GET", "OPTIONS")
	r.Router.Path("/new-deployment-event").HandlerFunc(r.restHandler.ProcessDeploymentEvent).Methods("POST")
	r.Router.Path("/reset-app-environment").HandlerFunc(r.restHandler.ResetApplication).Methods("POST")

}
