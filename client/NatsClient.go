package client

import (
	"time"

	"github.com/caarlos0/env"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

const (
	POLL_CD_SUCCESS         = "ORCHESTRATOR.CD.TRIGGER"
	POLL_CD_SUCCESS_GRP     = "ORCHESTRATOR.CD.TRIGGER_GRP1"
	POLL_CD_SUCCESS_DURABLE = "ORCHESTRATOR.CD.TRIGGER_DURABLE1"
)

type PubSubClient struct {
	Logger     *zap.SugaredLogger
	JetStrCtxt nats.JetStreamContext
	Conn       *nats.Conn
}

type PubSubConfig struct {
	NatsServerHost string `env:"NATS_SERVER_HOST" envDefault:"nats://localhost:4222"`
}

func NewPubSubClient(logger *zap.SugaredLogger) (*PubSubClient, error) {
	cfg := &PubSubConfig{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	nc, err := nats.Connect(cfg.NatsServerHost, nats.ReconnectWait(10*time.Second), nats.MaxReconnects(100))
	if err != nil {
		logger.Error("err", err)
		return &PubSubClient{}, err
	}

	//Create a jetstream context
	js, err := nc.JetStream()

	if err != nil {
		logger.Errorw("Error while creating jetstream context", "error", err)
		return nil, err
	}

	natsClient := &PubSubClient{
		Logger:     logger,
		JetStrCtxt: js,
	}
	return natsClient, nil
}
