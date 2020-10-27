package sql

import (
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type LeadTime struct {
	tableName          struct{}      `sql:"lead_time"`
	Id                 int           `sql:"id"`
	AppReleaseId       int           `sql:"app_release_id"`
	PipelineMaterialId int           `sql:"pipeline_material_id"`
	CommitHash         string        `sql:"commit_hash"`
	CommitTime         time.Time     `sql:"commit_time"`
	LeadTime           time.Duration `sql:"lead_time"`
	AppRelease         *AppRelease
}

type LeadTimeRepository interface {
	Save(leadTime *LeadTime) (*LeadTime, error)
	FindByIds(ids []int) ([]LeadTime, error)
	CleanAppDataForEnvironment(appId, environmentId int, tx *pg.Tx) error
}

type LeadTimeRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewLeadTimeRepositoryImpl(dbConnection *pg.DB,
	logger *zap.SugaredLogger) *LeadTimeRepositoryImpl {
	return &LeadTimeRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

func (impl *LeadTimeRepositoryImpl) Save(leadTime *LeadTime) (*LeadTime, error) {
	_, err := impl.dbConnection.Model(leadTime).Insert()
	return leadTime, err
}

func (impl *LeadTimeRepositoryImpl) FindByIds(ids []int) ([]LeadTime, error) {
	var leadTimes []LeadTime
	err := impl.dbConnection.
		Model(&leadTimes).
		Where("app_release_id in (?)", pg.In(ids)).
		Select()
	return leadTimes, err
}

func (impl *LeadTimeRepositoryImpl) CleanAppDataForEnvironment(appId, environmentId int, tx *pg.Tx) error {
	r, err := tx.Model(&LeadTime{}).
		Table("app_release").
		Where("app_release.app_id =?", appId).
		Where("app_release.environment_id = ?", environmentId).
		Where("app_release.id = lead_time.app_release_id").
		Delete()
	if err != nil {
		return err
	} else {
		impl.logger.Infow("leadtime deleted for ", "app", appId, "env", environmentId, "count", r.RowsAffected())
		return nil
	}
}
