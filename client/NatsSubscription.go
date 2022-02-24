package client

import (
	"encoding/json"

	"github.com/devtron-labs/lens/pkg"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type NatsSubscription interface {
}

type NatsSubscriptionImpl struct {
	pubSubClient     *PubSubClient
	logger           *zap.SugaredLogger
	ingestionService pkg.IngestionService
}

func NewNatsSubscription(pubSubClient *PubSubClient,
	logger *zap.SugaredLogger,
	ingestionService pkg.IngestionService) (*NatsSubscriptionImpl, error) {
	ns := &NatsSubscriptionImpl{
		pubSubClient:     pubSubClient,
		logger:           logger,
		ingestionService: ingestionService,
	}
	return ns, ns.Subscribe()
}

//TODO : adhiran : Work with Nishant to see how we can bind to a specific stream
func (impl NatsSubscriptionImpl) Subscribe() error {
	_, err := impl.pubSubClient.JetStrCtxt.QueueSubscribe(POLL_CD_SUCCESS, POLL_CD_SUCCESS_GRP, func(msg *nats.Msg) {
		impl.logger.Debugw("received msg", "msg", msg)
		defer msg.Ack()
		deploymentEvent := &pkg.DeploymentEvent{}
		err := json.Unmarshal(msg.Data, deploymentEvent)
		if err != nil {
			impl.logger.Errorw("err in reading msg", "err", err, "msg", string(msg.Data))
			return
		}
		impl.logger.Debugw("deploymentEvent", "id", deploymentEvent)
		release, err := impl.ingestionService.ProcessDeploymentEvent(deploymentEvent)
		if err != nil {
			impl.logger.Errorw("err in processing deploymentEvent", "deploymentEvent", deploymentEvent, "err", err)
			return
		}
		impl.logger.Infow("app release saved ", "apprelease", release)
	}, nats.Durable(POLL_CD_SUCCESS_DURABLE), nats.DeliverLast(), nats.ManualAck(), nats.BindStream(""))
	return err
}
