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

package sql

import (
	"time"

	pg "github.com/go-pg/pg/v10"
	"go.uber.org/zap"
)

type LeadTime struct {
	tableName          struct{}      `pg:"lead_time"`
	Id                 int           `pg:"id"`
	AppReleaseId       int           `pg:"app_release_id"`
	PipelineMaterialId int           `pg:"pipeline_material_id"`
	CommitHash         string        `pg:"commit_hash"`
	CommitTime         time.Time     `pg:"commit_time"`
	LeadTime           time.Duration `pg:"lead_time"`
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
