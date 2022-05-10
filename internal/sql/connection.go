package sql

import (
	"context"

	"github.com/caarlos0/env"
	"github.com/devtron-labs/lens/internal/logger"
	pg "github.com/go-pg/pg/v10"
	"go.uber.org/zap"
)

type Config struct {
	Addr            string `env:"PG_ADDR" envDefault:"127.0.0.1"`
	Port            string `env:"PG_PORT" envDefault:"5432"`
	User            string `env:"PG_USER" envDefault:""`
	Password        string `env:"PG_PASSWORD" envDefault:""`
	Database        string `env:"PG_DATABASE" envDefault:"lens"`
	ApplicationName string `env:"APP" envDefault:"lens"`
	LogQuery        bool   `env:"PG_LOG_QUERY" envDefault:"true"`
}

func (d dbLogger) BeforeQuery(c context.Context, q *pg.QueryEvent) (context.Context, error) {
	return c, nil
}

func (d dbLogger) AfterQuery(c context.Context, q *pg.QueryEvent) error {
	query, err := q.FormattedQuery()
	logger.NewSugardLogger().Debugw("Printing formatted query", "query", query)
	return err
}

type dbLogger struct {
	beforeQueryMethod func(context.Context, *pg.QueryEvent) (context.Context, error)
	afterQueryMethod  func(context.Context, *pg.QueryEvent) error
}

func GetConfig() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)
	return cfg, err
}

func NewDbConnection(cfg *Config, logger *zap.SugaredLogger) (*pg.DB, error) {
	options := pg.Options{
		Addr:            cfg.Addr + ":" + cfg.Port,
		User:            cfg.User,
		Password:        cfg.Password,
		Database:        cfg.Database,
		ApplicationName: cfg.ApplicationName,
	}
	dbConnection := pg.Connect(&options)
	//check db connection
	var test string
	_, err := dbConnection.QueryOne(pg.Scan(&test), "SELECT 1")

	if err != nil {
		logger.Errorw("error in connecting db ", "db", cfg, "err", err)
		return nil, err
	} else {
		logger.Infow("connected with db", "db", cfg)
	}
	if cfg.LogQuery {
		dbConnection.AddQueryHook(dbLogger{})
	}
	return dbConnection, err
}
