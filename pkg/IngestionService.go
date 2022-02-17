package pkg

import (
	"time"

	"github.com/devtron-labs/lens/client/gitSensor"
	"github.com/devtron-labs/lens/internal/sql"
	pg "github.com/go-pg/pg/v10"
	"go.uber.org/zap"
)

type IngestionService interface {
	ProcessDeploymentEvent(deploymentEvent *DeploymentEvent) (*sql.AppRelease, error)
	CleanAppDataForEnvironment(appId, environmentId int) (bool, error)
}
type IngestionServiceImpl struct {
	logger                     *zap.SugaredLogger
	appReleaseRepository       sql.AppReleaseRepository
	PipelineMaterialRepository sql.PipelineMaterialRepository
	leadTimeRepository         sql.LeadTimeRepository
	gitSensorClient            gitSensor.GitSensorClient
}

func NewIngestionServiceImpl(logger *zap.SugaredLogger,
	appReleaseRepository sql.AppReleaseRepository,
	PipelineMaterialRepository sql.PipelineMaterialRepository,
	leadTimeRepository sql.LeadTimeRepository,
	gitSensorClient gitSensor.GitSensorClient) *IngestionServiceImpl {
	return &IngestionServiceImpl{
		logger:                     logger,
		appReleaseRepository:       appReleaseRepository,
		PipelineMaterialRepository: PipelineMaterialRepository,
		leadTimeRepository:         leadTimeRepository,
		gitSensorClient:            gitSensorClient,
	}
}

type DeploymentEvent struct {
	ApplicationId      int
	EnvironmentId      int
	ReleaseId          int
	PipelineOverrideId int
	TriggerTime        time.Time
	PipelineMaterials  []*PipelineMaterialInfo
	CiArtifactId       int
}
type PipelineMaterialInfo struct {
	PipelineMaterialId int
	CommitHash         string
}

// 1.save AppRelease
// 2. save PipelineMaterial with release status
// 4. check for first commit and rollback
// 5. fetch changes from git
// 6. save LeadTime and commit size
func (impl *IngestionServiceImpl) ProcessDeploymentEvent(deploymentEvent *DeploymentEvent) (*sql.AppRelease, error) {
	impl.logger.Infow("processing release trigger", "request", deploymentEvent)
	appRelease, err := impl.saveAppRelease(deploymentEvent)
	if err != nil {
		return nil, err
	}
	materials, err := impl.savePipelineMaterial(deploymentEvent, appRelease)
	if err != nil {
		return nil, err
	}
	//--------
	appRelease, err = impl.checkAndUpdateReleaseType(appRelease)
	if err != nil {
		return nil, err
	}
	if appRelease.ReleaseType == sql.RollBack {
		//no need to fetch git detail return
		return appRelease, nil //FIXME
	}
	//mark previous pipeline fail
	err = impl.markPreviousTriggerFail(appRelease)
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	}

	//TODO handle in separate worker/scheduler
	err = impl.fetchAndSaveChangesFromGit(appRelease, materials)
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	} else if err == pg.ErrNoRows {
		return appRelease, nil
	}
	return appRelease, nil
}

func (impl *IngestionServiceImpl) markPreviousTriggerFail(release *sql.AppRelease) error {
	impl.logger.Infow("markPreviousTriggerFail", "release", release)
	previousAppRelease, err := impl.appReleaseRepository.GetPreviousReleaseWithinTime(release.AppId, release.EnvironmentId, release.TriggerTime.Add(time.Hour*time.Duration(-2)), release.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting previous release", "app", release.AppId, "err", err)
		return err
	} else if err == pg.ErrNoRows {
		return err //no pipeline fail
	}
	if previousAppRelease != nil {
		impl.logger.Infow("pipeline failure detected", "PreviousappRelease", previousAppRelease)
		previousAppRelease.ReleaseStatus = sql.Failure
		previousAppRelease.UpdatedTime = time.Now()
		_, err = impl.appReleaseRepository.Update(previousAppRelease)
		if err != nil {
			impl.logger.Errorw("error in updating pipeline status", "PreviousappRelease", previousAppRelease, "err", err)
			return err
		}
		//mark this release as patch
		release.ReleaseType = sql.Patch
		release.UpdatedTime = time.Now()
		_, err = impl.appReleaseRepository.Update(release)
		if err != nil {
			impl.logger.Errorw("error in updating  patch status", "release", release, "err", err)
			return err
		}
	}
	return nil
}

func (impl *IngestionServiceImpl) fetchAndSaveChangesFromGit(appRelease *sql.AppRelease, materials []*sql.PipelineMaterial) error {
	impl.logger.Infow("fetchAndSaveChangesFromGit", "appRelease", appRelease, "materials", materials)

	//fetch previous released gitHash
	previousAppRelease, err := impl.appReleaseRepository.GetPreviousRelease(appRelease.AppId, appRelease.EnvironmentId, appRelease.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting previous release for", "appRelease", appRelease.Id, "err", err)
		return err
	} else if err == pg.ErrNoRows {
		return err
	}
	previousPipelineMaterials, err := impl.PipelineMaterialRepository.FindByAppReleaseId(previousAppRelease.Id)
	if err != nil {
		impl.logger.Errorw("error in fetching previous pipeline material", "appReleaseId", previousAppRelease.Id, "err", err)
		return err
	}

	oldMaterialCommitHash := make(map[int]string)
	for _, pipelineMaterial := range previousPipelineMaterials {
		oldMaterialCommitHash[pipelineMaterial.PipelineMaterialId] = pipelineMaterial.CommitHash
	}

	//fetch data from git sensor
	lineAdded := 0
	lineRemoved := 0
	oldestTime := time.Now()
	now := oldestTime
	var oldest *gitSensor.Commit
	oldestId := 0
	for _, pipelineMaterial := range materials {
		oldHash, ok := oldMaterialCommitHash[pipelineMaterial.PipelineMaterialId]
		if ok && oldHash != pipelineMaterial.CommitHash {
			request := &gitSensor.ReleaseChangesRequest{
				PipelineMaterialId: pipelineMaterial.PipelineMaterialId,
				OldCommit:          oldHash,
				NewCommit:          pipelineMaterial.CommitHash,
			}
			changes, err := impl.gitSensorClient.GetReleaseChanges(request)
			if err != nil {
				impl.logger.Errorw("error in fetching git data", "err", err)
				return err
			}
			for _, change := range changes.FileStats {
				//change.Name	//TODO apply file filter
				lineRemoved = lineRemoved + change.Deletion
				lineAdded = lineAdded + change.Addition
			}
			for _, d := range changes.Commits {
				if oldestTime.After(d.Author.Date) {
					oldestTime = d.Author.Date
					oldest = d
					oldestId = pipelineMaterial.PipelineMaterialId
				}
			}
		}
	}
	if !now.Equal(oldestTime) {

		leadTime := &sql.LeadTime{
			AppReleaseId:       appRelease.Id,
			CommitTime:         oldest.Committer.Date,                          //
			CommitHash:         oldest.Hash.Long,                               //
			PipelineMaterialId: oldestId,                                       //
			LeadTime:           appRelease.TriggerTime.Sub(oldest.Author.Date), //
		}
		_, err = impl.leadTimeRepository.Save(leadTime)
		if err != nil {
			impl.logger.Errorw("error in saving leadtime", "leadtime", leadTime, "err", err)
			return err
		}
	}

	appRelease.UpdatedTime = time.Now()
	appRelease.ProcessStage = sql.LeadTimeFetch
	appRelease.ChangeSizeLineAdded = lineAdded
	appRelease.ChangeSizeLineDeleted = lineRemoved
	appRelease, err = impl.appReleaseRepository.Update(appRelease)
	if err != nil {
		impl.logger.Errorw("error in updating releaseTime", "appRelease", appRelease, "err", err)
		return err
	}
	return nil
}

func (impl *IngestionServiceImpl) saveAppRelease(deploymentEvent *DeploymentEvent) (*sql.AppRelease, error) {
	impl.logger.Infow("save appRelease", "deploymentEvent", deploymentEvent)
	appRelease := &sql.AppRelease{
		AppId:              deploymentEvent.ApplicationId,
		CiArtifactId:       deploymentEvent.CiArtifactId,
		TriggerTime:        deploymentEvent.TriggerTime,
		EnvironmentId:      deploymentEvent.EnvironmentId,
		CreatedTime:        time.Now(),
		UpdatedTime:        time.Now(),
		PipelineOverrideId: deploymentEvent.PipelineOverrideId,
		ReleaseId:          deploymentEvent.ReleaseId,
		ProcessStage:       sql.Init,
		ReleaseType:        sql.Unknown,
	}
	appRelease, err := impl.appReleaseRepository.Save(appRelease)
	if err != nil {
		impl.logger.Errorw("error in saving initial event ", "event", appRelease, "err", err)
		return nil, err
	}
	return appRelease, nil
}

func (impl *IngestionServiceImpl) savePipelineMaterial(deploymentEvent *DeploymentEvent, appRelease *sql.AppRelease) (materials []*sql.PipelineMaterial, err error) {
	impl.logger.Infow("save pipeline material ", "deploymentEvent", deploymentEvent, "appRelease", appRelease)
	for _, pipelineMaterialInfo := range deploymentEvent.PipelineMaterials {
		material := &sql.PipelineMaterial{
			PipelineMaterialId: pipelineMaterialInfo.PipelineMaterialId,
			CommitHash:         pipelineMaterialInfo.CommitHash,
			AppReleaseId:       appRelease.Id,
		}
		materials = append(materials, material)
	}
	err = impl.PipelineMaterialRepository.Save(materials...)
	if err != nil {
		impl.logger.Errorw("error in saving pipeline material", "material", materials, "err", err)
		return nil, err
	}
	return materials, nil
}

func (impl *IngestionServiceImpl) checkAndUpdateReleaseType(appRelease *sql.AppRelease) (*sql.AppRelease, error) {
	impl.logger.Infow("check and update release type ", "appRelease", appRelease)
	duplicate, err := impl.appReleaseRepository.CheckDuplicateRelease(appRelease.AppId, appRelease.EnvironmentId, appRelease.CiArtifactId)
	if err != nil {
		impl.logger.Errorw("eror in determining rollback", "pipelineOverrideId", appRelease.PipelineOverrideId, "err", err)
		return appRelease, err
	}
	if duplicate {
		appRelease.ReleaseType = sql.RollBack
	} else {
		appRelease.ReleaseType = sql.RollForward
	}
	appRelease.ProcessStage = sql.ReleaseTypeDetermined
	appRelease.UpdatedTime = time.Now()
	appRelease, err = impl.appReleaseRepository.Update(appRelease)

	if err != nil {
		impl.logger.Errorw("error in updating release status", "appRelease", appRelease, "err", err)
		return appRelease, err
	}
	return appRelease, err
}

func (impl *IngestionServiceImpl) CleanAppDataForEnvironment(appId, environmentId int) (bool, error) {
	impl.logger.Infow("cleaning app data for ", "app", appId, "env", environmentId)
	err := impl.appReleaseRepository.CleanAppDataForEnvironment(appId, environmentId)
	if err != nil {
		impl.logger.Errorw("error in cleaning data", "err", err)
		return false, err
	}
	return true, err

}
