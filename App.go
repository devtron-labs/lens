package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/devtron-labs/lens/api"
	"github.com/devtron-labs/lens/client"
	"github.com/devtron-labs/lens/pkg"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type App struct {
	MuxRouter        *api.MuxRouter
	Logger           *zap.SugaredLogger
	IngestionService pkg.IngestionService
	server           *http.Server
	db               *pg.DB
	natsSubscription *client.NatsSubscriptionImpl
}

func NewApp(MuxRouter *api.MuxRouter, Logger *zap.SugaredLogger, db *pg.DB, IngestionService pkg.IngestionService, natsSubscription *client.NatsSubscriptionImpl) *App {
	return &App{
		MuxRouter:        MuxRouter,
		Logger:           Logger,
		db:               db,
		natsSubscription: natsSubscription,
		IngestionService: IngestionService,
	}
}

func (app *App) Start() {
	port := 8080 //TODO: extract from environment variable
	app.Logger.Infow("starting server on ", "port", port)
	app.MuxRouter.Init()
	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: app.MuxRouter.Router}
	app.server = server
	err := server.ListenAndServe()
	if err != nil {
		app.Logger.Errorw("error in startup", "err", err)
		os.Exit(2)
	}
}

//TODO : adhiran : Why stopping nats connection code was commented out here?
func (app *App) Stop() {
	app.Logger.Infow("lens shutdown initiating")
	timeoutContext, _ := context.WithTimeout(context.Background(), 5*time.Second)

	app.Logger.Infow("closing router")
	err := app.server.Shutdown(timeoutContext)
	if err != nil {
		app.Logger.Errorw("error in mux router shutdown", "err", err)
	}
	app.Logger.Infow("closing db connection")
	err = app.db.Close()
	if err != nil {
		app.Logger.Errorw("error in closing db connection", "err", err)
	}

	app.Logger.Infow("housekeeping done. exiting now")
}
