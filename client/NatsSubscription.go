package client

import (
	"github.com/devtron-labs/lens/internal"
	"github.com/devtron-labs/lens/pkg"
	"encoding/json"
	"github.com/nats-io/stan"
	"go.uber.org/zap"
	"time"
)

type NatsSubscription interface {
}

type NatsSubscriptionImpl struct {
	nats             stan.Conn
	logger           *zap.SugaredLogger
	ingestionService pkg.IngestionService
}

func NewNatsSubscription(nats stan.Conn,
	logger *zap.SugaredLogger,
	ingestionService pkg.IngestionService, ) (*NatsSubscriptionImpl, error) {
	ns := &NatsSubscriptionImpl{
		nats:             nats,
		logger:           logger,
		ingestionService: ingestionService,
	}
	return ns, ns.Subscribe()
}

func (impl NatsSubscriptionImpl) Subscribe() error {
	aw, _ := time.ParseDuration("20s")
	_, err := impl.nats.QueueSubscribe(internal.POLL_CD_SUCCESS, internal.POLL_CD_SUCCESS_GRP, func(msg *stan.Msg) {
		impl.logger.Debugw("received msg", "msg", msg)
		defer msg.Ack()
		deploymentEvent := &pkg.DeploymentEvent{}
		err := json.Unmarshal(msg.Data, deploymentEvent)
		if err != nil {
			impl.logger.Errorw("err in reading msg", "err", err, "msg", string(msg.Data))
			return
		}
		impl.logger.Debugw("deploymentEvent", "id", deploymentEvent, )
		release, err := impl.ingestionService.ProcessDeploymentEvent(deploymentEvent)
		if err != nil {
			impl.logger.Errorw("err in processing deploymentEvent", "deploymentEvent", deploymentEvent, "err", err)
			return
		}
		impl.logger.Infow("app release saved ", "apprelease", release)
	}, stan.DurableName(internal.POLL_CD_SUCCESS_DURABLE), stan.StartWithLastReceived(), stan.SetManualAckMode(), stan.AckWait(aw), stan.MaxInflight(1))
	//s.Close()
	return err
}
