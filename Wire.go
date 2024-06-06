//go:build wireinject
// +build wireinject

/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/lens/api"
	"github.com/devtron-labs/lens/client"
	"github.com/devtron-labs/lens/client/gitSensor"
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
		gitSensor.GetConfig,
		gitSensor.NewGitSensorGrpcClientImpl,
		wire.Bind(new(gitSensor.GitSensorGrpcClient), new(*gitSensor.GitSensorGrpcClientImpl)),
		pubsub.NewPubSubClientServiceImpl,
		client.NewNatsSubscription,
	)
	return &App{}, nil
}
