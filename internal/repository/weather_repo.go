package repository

import (
	"ccplatform/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// WeatherRepo 气象数据访问层，封装和风天气缓存表的所有数据库操作。
type WeatherRepo struct {
	db *gorm.DB
}

// NewWeatherRepo 创建 WeatherRepo 实例，使用全局 DB 连接。
func NewWeatherRepo() *WeatherRepo {
	return &WeatherRepo{db: DB}
}

// ─── 实时天气 ─────────────────────────────────────────────────

// UpsertNow 写入/更新实时天气缓存（主键 location_id 冲突时覆盖）。
func (r *WeatherRepo) UpsertNow(cache *model.WeatherNowCache) error {
	return r.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(cache).Error
}

// GetNowByStationID 按电站 ID 查询实时天气缓存。
func (r *WeatherRepo) GetNowByStationID(stationID string) (*model.WeatherNowCache, error) {
	var now model.WeatherNowCache
	err := r.db.Where("station_id = ?", stationID).First(&now).Error
	if err != nil {
		return nil, err
	}
	return &now, nil
}

// GetNowByLocationID 按和风天气 LocationID 查询实时天气缓存。
func (r *WeatherRepo) GetNowByLocationID(locID string) (*model.WeatherNowCache, error) {
	var now model.WeatherNowCache
	err := r.db.Where("location_id = ?", locID).First(&now).Error
	if err != nil {
		return nil, err
	}
	return &now, nil
}

// ─── 7 天预报 ─────────────────────────────────────────────────

// BatchUpsertDaily 批量写入/更新逐天预报缓存（联合主键冲突时覆盖）。
func (r *WeatherRepo) BatchUpsertDaily(caches []model.WeatherDailyCache) error {
	if len(caches) == 0 {
		return nil
	}
	return r.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&caches).Error
}

// GetDailyByStationID 按电站 ID 查询逐天预报，按预报日期升序排列。
// days 限制返回条数，<=0 时默认返回 7 条。
func (r *WeatherRepo) GetDailyByStationID(stationID string, days int) ([]model.WeatherDailyCache, error) {
	var list []model.WeatherDailyCache
	q := r.db.Where("station_id = ?", stationID).Order("fx_date ASC")
	if days > 0 {
		q = q.Limit(days)
	}
	err := q.Find(&list).Error
	return list, err
}

// ─── 逐小时预报 ───────────────────────────────────────────────

// BatchUpsertHourly 批量写入/更新逐小时预报缓存。
func (r *WeatherRepo) BatchUpsertHourly(caches []model.WeatherHourlyCache) error {
	if len(caches) == 0 {
		return nil
	}
	return r.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&caches).Error
}

// GetHourlyByStationID 按电站 ID 查询逐小时预报，按预报时间升序排列。
func (r *WeatherRepo) GetHourlyByStationID(stationID string, hours int) ([]model.WeatherHourlyCache, error) {
	var list []model.WeatherHourlyCache
	q := r.db.Where("station_id = ?", stationID).Order("fx_time ASC")
	if hours > 0 {
		q = q.Limit(hours)
	}
	err := q.Find(&list).Error
	return list, err
}

// ─── 天气预警 ─────────────────────────────────────────────────

// UpsertWarning 写入/更新单条预警记录（联合主键冲突时覆盖）。
func (r *WeatherRepo) UpsertWarning(warning *model.WeatherWarning) error {
	return r.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(warning).Error
}

// GetActiveWarnings 查询当前仍在生效的预警（end_time > 当前预警 begin 时间）。
// 返回按发布时间倒序排列的预警列表。
func (r *WeatherRepo) GetActiveWarnings(stationID string) ([]model.WeatherWarning, error) {
	var list []model.WeatherWarning
	q := r.db.Where("station_id = ? AND status = ?", stationID, "active").Order("pub_time DESC")
	err := q.Find(&list).Error
	return list, err
}

// ListWarnings 分页查询历史预警，支持按日期范围筛选。
func (r *WeatherRepo) ListWarnings(stationID string, page, size int, startDate, endDate string) ([]model.WeatherWarning, int64, error) {
	var list []model.WeatherWarning
	var total int64
	q := r.db.Model(&model.WeatherWarning{}).Where("station_id = ?", stationID)
	if startDate != "" {
		q = q.Where("start_time >= ?", startDate)
	}
	if endDate != "" {
		q = q.Where("end_time <= ?", endDate)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	err := q.Offset((page - 1) * size).Limit(size).Order("pub_time DESC").Find(&list).Error
	return list, total, err
}

// ─── 空气质量 ─────────────────────────────────────────────────

// UpsertAirQuality 写入/更新空气质量缓存（主键 location_id 冲突时覆盖）。
func (r *WeatherRepo) UpsertAirQuality(air *model.WeatherAirQuality) error {
	return r.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(air).Error
}

// GetAirQualityByStationID 按电站 ID 查询最新空气质量缓存。
func (r *WeatherRepo) GetAirQualityByStationID(stationID string) (*model.WeatherAirQuality, error) {
	var air model.WeatherAirQuality
	err := r.db.Where("station_id = ?", stationID).First(&air).Error
	if err != nil {
		return nil, err
	}
	return &air, nil
}
