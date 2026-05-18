package repository

import (
	"ccplatform/internal/model"
	"time"

	"gorm.io/gorm"
)

// AlarmRepo 告警数据访问层，封装 alarms 表的所有数据库操作。
// 提供告警的增删改查、处理状态更新、按级别统计、趋势分析等功能。
type AlarmRepo struct {
	db *gorm.DB
}

// NewAlarmRepo 创建 AlarmRepo 实例，使用全局 DB 连接。
func NewAlarmRepo() *AlarmRepo {
	return &AlarmRepo{db: DB}
}

// Create 创建新告警记录，由 MQTT 告警消息处理器调用。
func (r *AlarmRepo) Create(alarm *model.Alarm) error {
	return r.db.Create(alarm).Error
}

// GetByID 根据告警 ID 查询单条记录。
func (r *AlarmRepo) GetByID(id string) (*model.Alarm, error) {
	var alarm model.Alarm
	err := r.db.Where("alarm_id = ?", id).First(&alarm).Error
	return &alarm, err
}

// Update 更新告警信息（全量更新）。
func (r *AlarmRepo) Update(alarm *model.Alarm) error {
	return r.db.Save(alarm).Error
}

// List 分页查询告警列表，支持按电站、机器人、告警级别、处理状态筛选。
// alarmLevel=0 表示不筛选级别，handleStatus<0 表示不筛选处理状态。
func (r *AlarmRepo) List(page, size int, stationID, robotID string, alarmLevel, handleStatus int) ([]model.Alarm, int64, error) {
	var alarms []model.Alarm
	var total int64
	query := r.db.Model(&model.Alarm{})
	if stationID != "" {
		query = query.Where("station_id = ?", stationID)
	}
	if robotID != "" {
		query = query.Where("robot_id = ?", robotID)
	}
	if alarmLevel > 0 {
		query = query.Where("alarm_level = ?", alarmLevel)
	}
	if handleStatus >= 0 {
		query = query.Where("handle_status = ?", handleStatus)
	}
	query.Count(&total)
	err := query.Offset((page - 1) * size).Limit(size).Order("alarm_time DESC").Find(&alarms).Error
	return alarms, total, err
}

// Handle 处理告警，记录处理人和处理时间，更新处理状态。
// status: 1已处理 2已忽略
func (r *AlarmRepo) Handle(id string, handlerID string, status int8) error {
	now := time.Now()
	return r.db.Model(&model.Alarm{}).Where("alarm_id = ?", id).Updates(map[string]interface{}{
		"handle_status": status,
		"handle_time":   now,
		"handler_id":    handlerID,
	}).Error
}

// CountUnhandled 统计未处理告警总数（handle_status=0），用于仪表盘展示。
func (r *AlarmRepo) CountUnhandled() (int64, error) {
	var count int64
	err := r.db.Model(&model.Alarm{}).Where("handle_status = 0").Count(&count).Error
	return count, err
}

// CountByLevel 按告警级别统计未处理告警数量。
// 返回 info/warning/error/critical/total 五个维度，用于告警统计面板。
func (r *AlarmRepo) CountByLevel() (map[string]int64, error) {
	result := make(map[string]int64)
	levels := []struct {
		Level int8
		Count int64
	}{}
	r.db.Model(&model.Alarm{}).Where("handle_status = 0").
		Select("alarm_level as level, count(*) as count").
		Group("alarm_level").Scan(&levels)
	for _, l := range levels {
		switch l.Level {
		case 1:
			result["info"] = l.Count    // 提示级别
		case 2:
			result["warning"] = l.Count // 一般告警
		case 3:
			result["error"] = l.Count   // 严重告警
		case 4:
			result["critical"] = l.Count // 紧急告警
		}
	}
	var total int64
	r.db.Model(&model.Alarm{}).Where("handle_status = 0").Count(&total)
	result["total"] = total
	return result, nil
}

// GetRecentAlarms 查询最近的告警记录，用于监控中心实时告警展示。
func (r *AlarmRepo) GetRecentAlarms(limit int) ([]model.Alarm, error) {
	var alarms []model.Alarm
	err := r.db.Order("alarm_time DESC").Limit(limit).Find(&alarms).Error
	return alarms, err
}

// GetTrendByType 按告警类型统计数量，用于趋势分析。
func (r *AlarmRepo) GetTrendByType(startTime, endTime string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	query := r.db.Model(&model.Alarm{}).Select("alarm_type, count(*) as count").Group("alarm_type")
	if startTime != "" {
		query = query.Where("alarm_time >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("alarm_time <= ?", endTime)
	}
	err := query.Order("count DESC").Find(&results).Error
	return results, err
}

// GetTrendByTime 按时间聚合告警趋势。
func (r *AlarmRepo) GetTrendByTime(granularity, startTime, endTime string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	dateFmt := "%Y-%m-%d"
	switch granularity {
	case "month":
		dateFmt = "%Y-%m"
	case "hour":
		dateFmt = "%Y-%m-%d %H:00"
	}
	query := r.db.Model(&model.Alarm{}).Select("DATE_FORMAT(alarm_time, ?) as period, count(*) as count", dateFmt).Group("period")
	if startTime != "" {
		query = query.Where("alarm_time >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("alarm_time <= ?", endTime)
	}
	err := query.Order("period ASC").Find(&results).Error
	return results, err
}

// GetTopAlarmTypes 获取高频告警类型 Top N。
func (r *AlarmRepo) GetTopAlarmTypes(limit int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	if limit <= 0 {
		limit = 10
	}
	err := r.db.Model(&model.Alarm{}).Select("alarm_type, count(*) as count, station_id").
		Group("alarm_type, station_id").Order("count DESC").Limit(limit).Find(&results).Error
	return results, err
}

// CountRecentByType 统计指定机器人在最近 N 分钟内同类型告警的数量，用于升级规则评估。
func (r *AlarmRepo) CountRecentByType(robotID, alarmType string, withinMin int) (int64, error) {
	var count int64
	cutoff := time.Now().Add(-time.Duration(withinMin) * time.Minute)
	err := r.db.Model(&model.Alarm{}).
		Where("robot_id = ? AND alarm_type = ? AND alarm_time >= ?", robotID, alarmType, cutoff).
		Count(&count).Error
	return count, err
}

// GetLatestAlarmTime 获取指定机器人同类型告警的最近一条记录时间，用于抑制规则评估。
func (r *AlarmRepo) GetLatestAlarmTime(robotID, alarmType string) (*time.Time, error) {
	var alarm model.Alarm
	err := r.db.Where("robot_id = ? AND alarm_type = ?", robotID, alarmType).
		Order("alarm_time DESC").First(&alarm).Error
	if err != nil {
		return nil, err
	}
	return &alarm.AlarmTime, nil
}

// CountByStation 统计指定电站的未处理告警数量。
func (r *AlarmRepo) CountByStation(stationID string) (int64, error) {
	var count int64
	err := r.db.Model(&model.Alarm{}).Where("station_id = ? AND handle_status = 0", stationID).Count(&count).Error
	return count, err
}
