package sql

import (
	pg "github.com/go-pg/pg/v10"
	"go.uber.org/zap"
)

type PipelineMaterial struct {
	tableName          struct{} `pg:"pipeline_material"`
	PipelineMaterialId int      `pg:"pipeline_material_id"`
	CommitHash         string   `pg:"commit_hash"`
	AppReleaseId       int      `pg:"app_release_id"`
	AppRelease         *AppRelease
}
type PipelineMaterialRepository interface {
	Save(pipelineMaterial ...*PipelineMaterial) error
	FindByAppReleaseId(appReleaseId int) ([]*PipelineMaterial, error)
	FindByAppReleaseIds(appReleaseIds []int) ([]*PipelineMaterial, error)
	CleanAppDataForEnvironment(appId, environmentId int, tx *pg.Tx) error
}

type PipelineMaterialRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewPipelineMaterialRepositoryImpl(dbConnection *pg.DB,
	logger *zap.SugaredLogger) *PipelineMaterialRepositoryImpl {
	return &PipelineMaterialRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}
func (impl *PipelineMaterialRepositoryImpl) FindByAppReleaseId(appReleaseId int) ([]*PipelineMaterial, error) {
	var pipelineMaterials []*PipelineMaterial
	err := impl.dbConnection.Model(&pipelineMaterials).Where("app_release_id = ?", appReleaseId).Select()
	return pipelineMaterials, err
}

func (impl *PipelineMaterialRepositoryImpl) FindByAppReleaseIds(appReleaseIds []int) ([]*PipelineMaterial, error) {
	var pipelineMaterials []*PipelineMaterial
	err := impl.dbConnection.Model(&pipelineMaterials).Where("app_release_id in (?)", pg.In(appReleaseIds)).Select()
	return pipelineMaterials, err
}

func (impl *PipelineMaterialRepositoryImpl) Save(pipelineMaterial ...*PipelineMaterial) error {
	_, err := impl.dbConnection.Model(&pipelineMaterial).Insert()
	return err
}

func (impl *PipelineMaterialRepositoryImpl) CleanAppDataForEnvironment(appId, environmentId int, tx *pg.Tx) error {
	r, err := tx.Model(&PipelineMaterial{}).
		Table("app_release").
		Where("app_release.app_id =?", appId).
		Where("app_release.environment_id = ?", environmentId).
		Where("app_release.id = pipeline_material.app_release_id").
		Delete()
	if err != nil {
		return err
	} else {
		impl.logger.Infow("pipelineMaterial deleted for ", "app", appId, "env", environmentId, "count", r.RowsAffected())
		return nil
	}
}
