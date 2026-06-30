package repository

import (
	"ccplatform/internal/model"
	"time"

	"gorm.io/gorm"
)

// CleaningRecordRepo 清扫记录数据访问层。
type CleaningRecordRepo struct {
	db *gorm.DB
}

func NewCleaningRecordRepo() *CleaningRecordRepo {
	return &CleaningRecordRepo{db: DB}
}

// Create 写入清扫记录。
func (r *CleaningRecordRepo) Create(rec *model.CleaningRecord) error {
	return r.db.Create(rec).Error
}

// SumCleanAreaByStation 按电站和时间范围汇总清扫面积。
func (r *CleaningRecordRepo) SumCleanAreaByStation(stationID, startTime, endTime string) (float64, error) {
	var sum float64
	query := r.db.Model(&model.CleaningRecord{}).Select("COALESCE(SUM(clean_area), 0)")
	if stationID != "" {
		query = query.Where("station_id = ?", stationID)
	}
	if startTime != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", startTime); err == nil {
			query = query.Where("record_time >= ?", t)
		}
	}
	if endTime != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", endTime); err == nil {
			query = query.Where("record_time <= ?", t)
		}
	}
	err := query.Scan(&sum).Error
	return sum, err
}

// GetByStation 按电站查询清扫记录。
func (r *CleaningRecordRepo) GetByStation(stationID, startTime, endTime string) ([]model.CleaningRecord, error) {
	var records []model.CleaningRecord
	query := r.db.Model(&model.CleaningRecord{})
	if stationID != "" {
		query = query.Where("station_id = ?", stationID)
	}
	if startTime != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", startTime); err == nil {
			query = query.Where("record_time >= ?", t)
		}
	}
	if endTime != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", endTime); err == nil {
			query = query.Where("record_time <= ?", t)
		}
	}
	err := query.Order("record_time ASC").Find(&records).Error
	return records, err
}

// CountByStation 按电站统计清扫记录数。
func (r *CleaningRecordRepo) CountByStation(stationID string) (int64, error) {
	var count int64
	err := r.db.Model(&model.CleaningRecord{}).Where("station_id = ?", stationID).Count(&count).Error
	return count, err
}
