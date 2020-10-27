package internal

import (
	"github.com/caarlos0/env"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan"
	"math/rand"
	"strconv"
	"time"
)

const (
	POLL_CD_SUCCESS         = "ORCHESTRATOR.CD.TRIGGER"
	POLL_CD_SUCCESS_GRP     = "ORCHESTRATOR.CD.TRIGGER_GRP1"
	POLL_CD_SUCCESS_DURABLE = "ORCHESTRATOR.CD.TRIGGER_DURABLE1"
)

type PubSubConfig struct {
	NatsServerHost string `env:"NATS_SERVER_HOST" envDefault:"nats://localhost:4222"`
	ClusterId      string `env:"CLUSTER_ID" envDefault:"devtron-stan"`
	ClientId       string `env:"CLIENT_ID" envDefault:"lens"`
}

func NewNatsConnection() (stan.Conn, error) {
	cfg := &PubSubConfig{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	nc, err := nats.Connect(cfg.NatsServerHost, nats.ReconnectWait(10*time.Second), nats.MaxReconnects(100))
	if err != nil {
		return nil, err
	}
	s := rand.NewSource(time.Now().UnixNano())
	uuid := rand.New(s)
	uniqueClientId := cfg.ClientId + strconv.Itoa(uuid.Int())

	sc, err := stan.Connect(cfg.ClusterId, uniqueClientId, stan.NatsConn(nc))
	if err != nil {
		return nil, err
	}
	return sc, nil
}
