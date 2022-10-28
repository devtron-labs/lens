package client

import (
	"encoding/json"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
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
	callback := func(msg *pubsub.PubSubMsg) {
		ns.logger.Debugw("received msg", "msg", msg)
		//defer msg.Ack()
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

	err := pubSubClient.Subscribe(pubsub.CD_SUCCESS, callback)
	if err != nil {
		ns.logger.Errorw("Error while subscribing to pubsub client", "topic", pubsub.CD_SUCCESS, "error", err)
	}
	return ns, nil
}
