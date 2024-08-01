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

package client

import (
	"encoding/json"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/common-lib/pubsub-lib/model"
	"github.com/devtron-labs/lens/pkg"
	"go.uber.org/zap"
)

type NatsSubscription interface {
}

type NatsSubscriptionImpl struct {
	pubSubClient     *pubsub.PubSubClientServiceImpl
	logger           *zap.SugaredLogger
	ingestionService pkg.IngestionService
}

func NewNatsSubscription(pubSubClient *pubsub.PubSubClientServiceImpl,
	logger *zap.SugaredLogger,
	ingestionService pkg.IngestionService) (*NatsSubscriptionImpl, error) {
	ns := &NatsSubscriptionImpl{
		pubSubClient:     pubSubClient,
		logger:           logger,
		ingestionService: ingestionService,
	}
	callback := func(msg *model.PubSubMsg) {
		ns.logger.Debugw("received msg", "msg", msg)
		// defer msg.Ack()
		deploymentEvent := &pkg.DeploymentEvent{}
		err := json.Unmarshal([]byte(msg.Data), deploymentEvent)
		if err != nil {
			ns.logger.Errorw("err in reading msg", "err", err, "msg", string(msg.Data))
			return
		}
		ns.logger.Debugw("deploymentEvent", "id", deploymentEvent)
		release, err := ns.ingestionService.ProcessDeploymentEvent(deploymentEvent)
		if err != nil {
			ns.logger.Errorw("err in processing deploymentEvent", "deploymentEvent", deploymentEvent, "err", err)
			return
		}
		ns.logger.Infow("app release saved ", "apprelease", release)
	}

	var loggerFunc pubsub.LoggerFunc = func(msg model.PubSubMsg) (string, []interface{}) {
		deploymentEvent := &pkg.DeploymentEvent{}
		err := json.Unmarshal([]byte(msg.Data), &deploymentEvent)
		if err != nil {
			return "error while unmarshalling deploymentEvent object", []interface{}{"err", err, "msg", msg.Data}
		}
		return "got message for deployment stage completion", []interface{}{"envId", deploymentEvent.EnvironmentId, "appId", deploymentEvent.ApplicationId, "ciArtifactId", deploymentEvent.CiArtifactId}
	}

	err := pubSubClient.Subscribe(pubsub.CD_SUCCESS, callback, loggerFunc)
	if err != nil {
		ns.logger.Errorw("Error while subscribing to pubsub client", "topic", pubsub.CD_SUCCESS, "error", err)
	}
	return ns, err
}
