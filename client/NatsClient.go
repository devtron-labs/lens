package client

//type PubSubClient struct {
//	Logger     *zap.SugaredLogger
//	JetStrCtxt nats.JetStreamContext
//	Conn       *nats.Conn
//}
//
//type PubSubConfig struct {
//	NatsServerHost string `env:"NATS_SERVER_HOST" envDefault:"nats://localhost:4222"`
//}
//
//func NewPubSubClient(logger *zap.SugaredLogger) (*PubSubClient, error) {
//	cfg := &PubSubConfig{}
//	err := env.Parse(cfg)
//	if err != nil {
//		return nil, err
//	}
//
//	nc, err := nats.Connect(cfg.NatsServerHost,
//		nats.ReconnectWait(10*time.Second), nats.MaxReconnects(100),
//		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
//			logger.Errorw("Nats Connection got disconnected!", "Reason", err)
//		}),
//		nats.ReconnectHandler(func(nc *nats.Conn) {
//			logger.Infow("Nats Connection got reconnected", "url", nc.ConnectedUrl())
//		}),
//		nats.ClosedHandler(func(nc *nats.Conn) {
//			logger.Errorw("Nats Client Connection closed!", "Reason", nc.LastError())
//		}))
//	if err != nil {
//		logger.Error("err", err)
//		return &PubSubClient{}, err
//	}
//
//	//Create a jetstream context
//	js, err := nc.JetStream()
//
//	if err != nil {
//		logger.Errorw("Error while creating jetstream context", "error", err)
//		return nil, err
//	}
//
//	natsClient := &PubSubClient{
//		Logger:     logger,
//		JetStrCtxt: js,
//	}
//	return natsClient, nil
//}
