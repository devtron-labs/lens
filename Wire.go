//+build wireinject

package main

import (
	"github.com/devtron-labs/lens/api"
	"github.com/devtron-labs/lens/client"
	"github.com/devtron-labs/lens/client/gitSensor"
	"github.com/devtron-labs/lens/internal"
	"github.com/devtron-labs/lens/internal/logger"
	"github.com/devtron-labs/lens/internal/sql"
	"github.com/devtron-labs/lens/pkg"
	"github.com/google/wire"
)

func InitializeApp() (*App, error) {
	wire.Build(
		NewApp,
		api.NewMuxRouter,
		logger.NewSugardLogger,
		sql.GetConfig,
		sql.NewDbConnection,
		api.NewRestHandlerImpl,
		wire.Bind(new(api.RestHandler), new(*api.RestHandlerImpl)),
		pkg.NewIngestionServiceImpl,
		wire.Bind(new(pkg.IngestionService), new(*pkg.IngestionServiceImpl)),
		sql.NewAppReleaseRepositoryImpl,
		wire.Bind(new(sql.AppReleaseRepository), new(*sql.AppReleaseRepositoryImpl)),
		sql.NewLeadTimeRepositoryImpl,
		wire.Bind(new(sql.LeadTimeRepository), new(*sql.LeadTimeRepositoryImpl)),
		sql.NewPipelineMaterialRepositoryImpl,
		wire.Bind(new(sql.PipelineMaterialRepository), new(*sql.PipelineMaterialRepositoryImpl)),
		pkg.NewDeploymentMetricServiceImpl,
		wire.Bind(new(pkg.DeploymentMetricService), new(*pkg.DeploymentMetricServiceImpl)),
		gitSensor.GetGitSensorConfig,
		gitSensor.NewGitSensorSession,
		wire.Bind(new(gitSensor.GitSensorClient), new(*gitSensor.GitSensorClientImpl)),
		internal.NewNatsConnection,
		client.NewNatsSubscription,
	)
	return &App{}, nil
}
