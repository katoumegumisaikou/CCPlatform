package repository

import (
	"ccplatform/internal/model"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ComponentHealthRepo 机器人部件健康指数当前状态访问层。
type ComponentHealthRepo struct {
	db *gorm.DB
}

func NewComponentHealthRepo() *ComponentHealthRepo {
	return &ComponentHealthRepo{db: DB}
}

// Upsert 按 (robot_id, component) 写入或更新健康指数当前状态。
func (r *ComponentHealthRepo) Upsert(h *model.RobotComponentHealth) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "robot_id"}, {Name: "component"}},
		DoUpdates: clause.AssignmentColumns([]string{"health_score", "health_level", "degrade_rate", "metrics", "update_time"}),
	}).Create(h).Error
}

// GetByRobotID 查询指定机器人所有部件的当前健康状态。
func (r *ComponentHealthRepo) GetByRobotID(robotID string) ([]model.RobotComponentHealth, error) {
	var list []model.RobotComponentHealth
	err := r.db.Where("robot_id = ?", robotID).Order("component ASC").Find(&list).Error
	return list, err
}

// GetByRobotIDs 批量查询多个机器人的当前健康状态，用于电站级汇总。
func (r *ComponentHealthRepo) GetByRobotIDs(robotIDs []string) ([]model.RobotComponentHealth, error) {
	var list []model.RobotComponentHealth
	if len(robotIDs) == 0 {
		return list, nil
	}
	err := r.db.Where("robot_id IN ?", robotIDs).Find(&list).Error
	return list, err
}

// HealthHistoryRepo 机器人部件健康指数历史快照访问层。
type HealthHistoryRepo struct {
	db *gorm.DB
}

func NewHealthHistoryRepo() *HealthHistoryRepo {
	return &HealthHistoryRepo{db: DB}
}

// Create 追加一条健康指数历史快照。
func (r *HealthHistoryRepo) Create(h *model.RobotHealthHistory) error {
	return r.db.Create(h).Error
}

// GetRecent 查询指定机器人部件最近 N 条历史快照（按时间倒序），用于退化速率回归。
func (r *HealthHistoryRepo) GetRecent(robotID, component string, limit int) ([]model.RobotHealthHistory, error) {
	var list []model.RobotHealthHistory
	err := r.db.Where("robot_id = ? AND component = ?", robotID, component).
		Order("record_time DESC").Limit(limit).Find(&list).Error
	return list, err
}

// GetByRange 查询指定机器人部件在时间范围内的历史快照（按时间正序），用于趋势曲线展示。
func (r *HealthHistoryRepo) GetByRange(robotID, component string, startTime, endTime time.Time) ([]model.RobotHealthHistory, error) {
	var list []model.RobotHealthHistory
	err := r.db.Where("robot_id = ? AND component = ? AND record_time BETWEEN ? AND ?", robotID, component, startTime, endTime).
		Order("record_time ASC").Find(&list).Error
	return list, err
}

// SensorDataRepo PHM 传感器数据访问层。
type SensorDataRepo struct {
	db *gorm.DB
}

func NewSensorDataRepo() *SensorDataRepo {
	return &SensorDataRepo{db: DB}
}

// Create 写入一条传感器数据记录。
func (r *SensorDataRepo) Create(data *model.RobotSensorData) error {
	return r.db.Create(data).Error
}

// GetLatest 查询指定机器人部件最新一条传感器数据。
func (r *SensorDataRepo) GetLatest(robotID, component string) (*model.RobotSensorData, error) {
	var data model.RobotSensorData
	err := r.db.Where("robot_id = ? AND component = ?", robotID, component).
		Order("record_time DESC").First(&data).Error
	return &data, err
}

// GetRecent 查询指定机器人部件最近 N 条传感器数据，用于振动趋势跟踪。
func (r *SensorDataRepo) GetRecent(robotID, component string, limit int) ([]model.RobotSensorData, error) {
	var list []model.RobotSensorData
	err := r.db.Where("robot_id = ? AND component = ?", robotID, component).
		Order("record_time DESC").Limit(limit).Find(&list).Error
	return list, err
}

// FaultPatternRepo 故障模式库访问层。
type FaultPatternRepo struct {
	db *gorm.DB
}

func NewFaultPatternRepo() *FaultPatternRepo {
	return &FaultPatternRepo{db: DB}
}

// Count 统计故障模式库记录数，用于启动时判断是否需要做种子初始化。
func (r *FaultPatternRepo) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.FaultPattern{}).Count(&count).Error
	return count, err
}

// Create 创建一条故障模式记录。
func (r *FaultPatternRepo) Create(p *model.FaultPattern) error {
	return r.db.Create(p).Error
}

// GetByCode 根据故障模式编码查询单条记录。
func (r *FaultPatternRepo) GetByCode(code string) (*model.FaultPattern, error) {
	var p model.FaultPattern
	err := r.db.Where("pattern_code = ?", code).First(&p).Error
	return &p, err
}

// ListByComponent 查询指定部件的所有故障模式，用于匹配预测结果对应的知识库条目。
func (r *FaultPatternRepo) ListByComponent(component string) ([]model.FaultPattern, error) {
	var list []model.FaultPattern
	err := r.db.Where("component = ?", component).Find(&list).Error
	return list, err
}

// List 查询全部故障模式库条目。
func (r *FaultPatternRepo) List() ([]model.FaultPattern, error) {
	var list []model.FaultPattern
	err := r.db.Find(&list).Error
	return list, err
}

// FaultPredictionRepo 故障预测结果访问层。
type FaultPredictionRepo struct {
	db *gorm.DB
}

func NewFaultPredictionRepo() *FaultPredictionRepo {
	return &FaultPredictionRepo{db: DB}
}

// Create 创建一条故障预测记录。
func (r *FaultPredictionRepo) Create(p *model.FaultPrediction) error {
	return r.db.Create(p).Error
}

// Update 更新故障预测记录（全量更新）。
func (r *FaultPredictionRepo) Update(p *model.FaultPrediction) error {
	return r.db.Save(p).Error
}

// GetByID 根据预测 ID 查询单条记录。
func (r *FaultPredictionRepo) GetByID(id string) (*model.FaultPrediction, error) {
	var p model.FaultPrediction
	err := r.db.Where("prediction_id = ?", id).First(&p).Error
	return &p, err
}

// GetOpenByRobotComponent 查询指定机器人部件当前未处理(status=0)的预测记录，用于去重更新而非重复创建。
func (r *FaultPredictionRepo) GetOpenByRobotComponent(robotID, component, patternCode string) (*model.FaultPrediction, error) {
	var p model.FaultPrediction
	err := r.db.Where("robot_id = ? AND component = ? AND pattern_code = ? AND status = 0", robotID, component, patternCode).
		Order("create_time DESC").First(&p).Error
	return &p, err
}

// GetByRobotID 查询指定机器人的故障预测记录（按创建时间倒序）。
func (r *FaultPredictionRepo) GetByRobotID(robotID string) ([]model.FaultPrediction, error) {
	var list []model.FaultPrediction
	err := r.db.Where("robot_id = ?", robotID).Order("create_time DESC").Find(&list).Error
	return list, err
}

// GetByStationID 查询指定电站下所有机器人的故障预测记录（按创建时间倒序）。
func (r *FaultPredictionRepo) GetByStationID(stationID string) ([]model.FaultPrediction, error) {
	var list []model.FaultPrediction
	err := r.db.Where("station_id = ?", stationID).Order("create_time DESC").Find(&list).Error
	return list, err
}

// GetOpenByStatus 查询处于 open(status=0) 状态的预测记录，可选按电站过滤，用于备件预测和维护窗口推荐聚合。
func (r *FaultPredictionRepo) GetOpenByStatus(stationID string) ([]model.FaultPrediction, error) {
	var list []model.FaultPrediction
	query := r.db.Where("status = 0")
	if stationID != "" {
		query = query.Where("station_id = ?", stationID)
	}
	err := query.Order("risk_level ASC, create_time DESC").Find(&list).Error
	return list, err
}

// List 分页查询故障预测记录，支持按机器人/电站/风险等级筛选。
func (r *FaultPredictionRepo) List(page, size int, robotID, stationID, riskLevel string) ([]model.FaultPrediction, int64, error) {
	var list []model.FaultPrediction
	var total int64
	query := r.db.Model(&model.FaultPrediction{})
	if robotID != "" {
		query = query.Where("robot_id = ?", robotID)
	}
	if stationID != "" {
		query = query.Where("station_id = ?", stationID)
	}
	if riskLevel != "" {
		query = query.Where("risk_level = ?", riskLevel)
	}
	query.Count(&total)
	err := query.Offset((page - 1) * size).Limit(size).Order("create_time DESC").Find(&list).Error
	return list, total, err
}
