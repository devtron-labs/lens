package pkg

import (
	"time"

	"github.com/devtron-labs/lens/internal/sql"
	pg "github.com/go-pg/pg/v10"
	"go.uber.org/zap"
)

const (
	layout = "2006-01-02T15:04:05.000Z"
)

type DeploymentMetricService interface {
	GetDeploymentMetrics(request *MetricRequest) (*Metrics, error)
}

type Metrics struct {
	Series                 []*Metric `json:"series"`
	AverageCycleTime       float64   `json:"average_cycle_time"`
	AverageLeadTime        float64   `json:"average_lead_time"`
	ChangeFailureRate      float64   `json:"change_failure_rate"`
	AverageRecoveryTime    float64   `json:"average_recovery_time"`
	AverageDeploymentSize  float32   `json:"average_deployment_size"`
	AverageLineAdded       float32   `json:"average_line_added"`
	AverageLineDeleted     float32   `json:"average_line_deleted"`
	LastFailedTime         string    `json:"last_failed_time"`
	RecoveryTimeLastFailed float64   `json:"recovery_time_last_failed"`
}

type Metric struct {
	ReleaseType           sql.ReleaseType   `json:"release_type"`
	ReleaseStatus         sql.ReleaseStatus `json:"release_status"`
	ReleaseTime           time.Time         `json:"release_time"`
	ChangeSizeLineAdded   int               `json:"change_size_line_added"`
	ChangeSizeLineDeleted int               `json:"change_size_line_deleted"`
	DeploymentSize        int               `json:"deployment_size"`
	CommitHash            string            `json:"commit_hash"`
	CommitTime            time.Time         `json:"commit_time"`
	LeadTime              float64           `json:"lead_time"`
	CycleTime             float64           `json:"cycle_time"`
	RecoveryTime          float64           `json:"recovery_time"`
}

type MetricRequest struct {
	AppId int    `json:"app_id"`
	EnvId int    `json:"env_id"`
	From  string `json:"from"`
	To    string `json:"to"`
}

type DeploymentMetricServiceImpl struct {
	logger                     *zap.SugaredLogger
	appReleaseRepository       sql.AppReleaseRepository
	pipelineMaterialRepository sql.PipelineMaterialRepository
	leadTimeRepository         sql.LeadTimeRepository
}

func NewDeploymentMetricServiceImpl(
	logger *zap.SugaredLogger,
	appReleaseRepository sql.AppReleaseRepository,
	pipelineMaterialRepository sql.PipelineMaterialRepository,
	leadTimeRepository sql.LeadTimeRepository) *DeploymentMetricServiceImpl {
	return &DeploymentMetricServiceImpl{
		logger:                     logger,
		appReleaseRepository:       appReleaseRepository,
		pipelineMaterialRepository: pipelineMaterialRepository,
		leadTimeRepository:         leadTimeRepository,
	}
}

func (impl DeploymentMetricServiceImpl) GetDeploymentMetrics(request *MetricRequest) (*Metrics, error) {
	from, err := time.Parse(layout, request.From)
	if err != nil {
		return nil, err
	}
	to, err := time.Parse(layout, request.To)
	if err != nil {
		return nil, err
	}
	releases, err := impl.appReleaseRepository.GetReleaseBetween(request.AppId, request.EnvId, from, to)
	if err != nil {
		impl.logger.Errorf("error getting data from db ", "err", err)
		return nil, err
	}
	if len(releases) == 0 {
		return &Metrics{Series: []*Metric{}}, nil
	}
	var ids []int
	for _, v := range releases {
		ids = append(ids, v.Id)
	}
	materials, err := impl.pipelineMaterialRepository.FindByAppReleaseIds(ids)
	if err != nil {
		impl.logger.Errorf("error getting material from db ", "err", err)
		return nil, err
	}
	leadTimes, err := impl.leadTimeRepository.FindByIds(ids)
	if err != nil {
		impl.logger.Errorf("error getting lead time from db ", "err", err)
		return nil, err
	}
	lastId := releases[len(releases)-1].Id
	lastRelease, err := impl.appReleaseRepository.GetPreviousRelease(request.AppId, request.EnvId, lastId)
	if err != nil {
		if err != pg.ErrNoRows {
			impl.logger.Errorf("error getting data from db ", "err", err)
		}
		lastRelease = nil
	}
	metrics, err := impl.populateMetrics(releases, materials, leadTimes, lastRelease)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func (impl DeploymentMetricServiceImpl) populateMetrics(appReleases []sql.AppRelease, materials []*sql.PipelineMaterial, leadTimes []sql.LeadTime, lastRelease *sql.AppRelease) (*Metrics, error) {
	releases := impl.transform(appReleases, materials, leadTimes)
	leadTimesCount := 0
	totalLeadTime := float64(0)
	for _, r := range releases {
		if r.LeadTime != float64(0) {
			totalLeadTime += r.LeadTime
			leadTimesCount++
		}
	}

	totalCycleTime := float64(0)
	cycleTimeCount := len(releases)
	for i := 0; i < len(releases)-1; i++ {
		releases[i].CycleTime = releases[i].ReleaseTime.Sub(releases[i+1].ReleaseTime).Minutes()
		totalCycleTime += releases[i].CycleTime
	}
	if lastRelease != nil {
		releases[len(releases)-1].CycleTime = releases[len(releases)-1].ReleaseTime.Sub(lastRelease.TriggerTime).Minutes()
		totalCycleTime += releases[len(releases)-1].CycleTime
	} else if len(releases) > 0 {
		releases[len(releases)-1].CycleTime = 0
		cycleTimeCount -= 1
	}
	averageCycleTime := float64(0)
	if cycleTimeCount > 0 {
		averageCycleTime = totalCycleTime / float64(cycleTimeCount)
	}

	metrics := &Metrics{
		Series: releases,
		//ChangeFailureRate: changeFailureRate,
		AverageCycleTime: averageCycleTime,
	}

	if leadTimesCount > 0 {
		metrics.AverageLeadTime = totalLeadTime / float64(leadTimesCount)
	}

	impl.calculateChangeFailureRateAndRecoveryTime(metrics)
	if len(metrics.Series) > 0 {
		impl.calculateChangeSize(metrics)
	}
	return metrics, nil
}

func (impl DeploymentMetricServiceImpl) calculateChangeFailureRateAndRecoveryTime(metrics *Metrics) {
	releases := metrics.Series
	failed := 0
	success := 0
	recoveryTime := float64(0)
	recovered := 0
	for _, v := range releases {
		if v.ReleaseStatus == sql.Failure {
			if metrics.LastFailedTime == "" {
				metrics.LastFailedTime = v.ReleaseTime.Format(layout)
			}
			//if i != 0 {
			//	releases[i].RecoveryTime = releases[i].ReleaseTime.Sub(releases[i+1].ReleaseTime)
			//	recoveryTime += int(releases[i].RecoveryTime.Hours())
			//}
			failed++
		}
		if v.ReleaseStatus == sql.Success {
			success++
		}
	}
	for i := 0; i < len(releases); i++ {
		if releases[i].ReleaseStatus == sql.Failure {
			if i < len(releases)-1 && releases[i+1].ReleaseStatus == sql.Failure {
				continue
			}
			for j := i - 1; j >= 0; j-- {
				if releases[j].ReleaseStatus == sql.Success {
					releases[i].RecoveryTime = releases[j].ReleaseTime.Sub(releases[i].ReleaseTime).Minutes()
					recoveryTime += releases[i].RecoveryTime
					recovered++
					if metrics.RecoveryTimeLastFailed == 0 {
						metrics.RecoveryTimeLastFailed = releases[i].RecoveryTime
					}
					break
				}
			}
		}
	}
	changeFailureRate := float64(0)
	averageRecoveryTime := float64(0)
	if success+failed > 0 {
		changeFailureRate = float64(failed) * float64(100) / float64(failed+success)
	}
	if failed > 0 && recovered > 0 {
		averageRecoveryTime = recoveryTime / float64(recovered)
	}
	metrics.ChangeFailureRate = changeFailureRate
	metrics.AverageRecoveryTime = averageRecoveryTime
}

func (impl DeploymentMetricServiceImpl) calculateChangeSize(metrics *Metrics) {
	releases := metrics.Series
	lineAdded := 0
	lineDeleted := 0
	deploymentSize := 0
	for _, v := range releases {
		lineAdded += v.ChangeSizeLineAdded
		lineDeleted += v.ChangeSizeLineDeleted
		deploymentSize += v.DeploymentSize
	}
	metrics.AverageDeploymentSize = float32(deploymentSize) / float32(len(releases))
	metrics.AverageLineAdded = float32(lineAdded) / float32(len(releases))
	metrics.AverageLineDeleted = float32(lineDeleted) / float32(len(releases))
}

func (impl DeploymentMetricServiceImpl) transform(releases []sql.AppRelease, materials []*sql.PipelineMaterial, leadTimes []sql.LeadTime) []*Metric {
	pm := make(map[int]*sql.PipelineMaterial)
	for _, v := range materials {
		pm[v.AppReleaseId] = v
	}
	lt := make(map[int]sql.LeadTime)
	for _, v := range leadTimes {
		lt[v.AppReleaseId] = v
	}

	impl.logger.Errorw("materials ", "mat", pm)

	metrics := make([]*Metric, 0)
	for _, v := range releases {
		metric := &Metric{
			ReleaseType:           v.ReleaseType,
			ReleaseStatus:         v.ReleaseStatus,
			ReleaseTime:           v.TriggerTime,
			ChangeSizeLineAdded:   v.ChangeSizeLineAdded,
			ChangeSizeLineDeleted: v.ChangeSizeLineDeleted,
			DeploymentSize:        v.ChangeSizeLineDeleted + v.ChangeSizeLineAdded,
			LeadTime:              0,
			CycleTime:             0,
			RecoveryTime:          0,
		}
		if p, ok := pm[v.Id]; ok {
			metric.CommitHash = p.CommitHash
		} else {
			impl.logger.Errorf("not found appId: %d", v.AppId)
		}
		if l, ok := lt[v.Id]; ok {
			metric.LeadTime = l.LeadTime.Minutes()
		} else {
			impl.logger.Errorf("not found appId: %d", v.AppId)
		}
		metrics = append(metrics, metric)
	}
	return metrics
}
