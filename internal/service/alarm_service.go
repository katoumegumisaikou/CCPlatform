package service

import (
	"ccplatform/internal/model"
	"ccplatform/internal/repository"
)

// AlarmService 告警业务逻辑层，处理告警的创建、查询、处理和统计。
type AlarmService struct {
	repo *repository.AlarmRepo
}

// NewAlarmService 创建 AlarmService 实例。
func NewAlarmService() *AlarmService {
	return &AlarmService{repo: repository.NewAlarmRepo()}
}

// Create 创建新告警记录，通常由 MQTT 告警消息触发。
func (s *AlarmService) Create(alarm *model.Alarm) error {
	return s.repo.Create(alarm)
}

// GetByID 根据 ID 查询告警详情。
func (s *AlarmService) GetByID(id string) (*model.Alarm, error) {
	return s.repo.GetByID(id)
}

// List 分页查询告警列表，支持多维度筛选。
func (s *AlarmService) List(page, size int, stationID, robotID string, alarmLevel, handleStatus int) ([]model.Alarm, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	return s.repo.List(page, size, stationID, robotID, alarmLevel, handleStatus)
}

// Handle 处理告警，记录处理人和处理状态（1已处理/2已忽略）。
func (s *AlarmService) Handle(alarmID, handlerID string, status int8) error {
	return s.repo.Handle(alarmID, handlerID, status)
}

// GetStats 获取告警统计信息，按级别分组统计未处理告警数量。
func (s *AlarmService) GetStats() (map[string]int64, error) {
	return s.repo.CountByLevel()
}

// GetRecentAlarms 获取最近的告警记录，用于监控中心实时告警展示。
func (s *AlarmService) GetRecentAlarms(limit int) ([]model.Alarm, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.repo.GetRecentAlarms(limit)
}

// GetTrendAnalysis 获取告警趋势分析数据。
func (s *AlarmService) GetTrendAnalysis(granularity, startTime, endTime string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	byType, _ := s.repo.GetTrendByType(startTime, endTime)
	result["by_type"] = byType
	byTime, _ := s.repo.GetTrendByTime(granularity, startTime, endTime)
	result["by_time"] = byTime
	topTypes, _ := s.repo.GetTopAlarmTypes(10)
	result["top_types"] = topTypes
	return result, nil
}

// CountUnhandled 统计未处理告警总数，用于仪表盘红色角标。
func (s *AlarmService) CountUnhandled() (int64, error) {
	return s.repo.CountUnhandled()
}
