package repository

import (
	"ccplatform/internal/model"

	"gorm.io/gorm"
)

// CleaningDecisionRepo 清扫决策数据访问层。
type CleaningDecisionRepo struct {
	db *gorm.DB
}

// NewCleaningDecisionRepo 创建决策规则 Repo 实例。
func NewCleaningDecisionRepo() *CleaningDecisionRepo {
	return &CleaningDecisionRepo{db: DB}
}

// ─── 清扫决策规则 ────────────────────────────────────────────

func (r *CleaningDecisionRepo) CreateRule(rule *model.CleaningDecisionRule) error {
	return r.db.Create(rule).Error
}

func (r *CleaningDecisionRepo) GetRuleByID(id string) (*model.CleaningDecisionRule, error) {
	var rule model.CleaningDecisionRule
	err := r.db.Where("rule_id = ?", id).First(&rule).Error
	return &rule, err
}

func (r *CleaningDecisionRepo) UpdateRule(rule *model.CleaningDecisionRule) error {
	return r.db.Save(rule).Error
}

func (r *CleaningDecisionRepo) DeleteRule(id string) error {
	return r.db.Delete(&model.CleaningDecisionRule{}, "rule_id = ?", id).Error
}

// ListRules 分页查询决策规则，支持按电站筛选。
func (r *CleaningDecisionRepo) ListRules(stationID string, page, size int) ([]model.CleaningDecisionRule, int64, error) {
	var list []model.CleaningDecisionRule
	var total int64
	q := r.db.Model(&model.CleaningDecisionRule{})
	if stationID != "" {
		q = q.Where("station_id = ? OR station_id = ''", stationID)
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
	err := q.Offset((page - 1) * size).Limit(size).Order("priority DESC, create_time DESC").Find(&list).Error
	return list, total, err
}

// GetRulesByStationID 获取适用于指定电站的规则（电站专属 + 全局规则）。
func (r *CleaningDecisionRepo) GetRulesByStationID(stationID string) ([]model.CleaningDecisionRule, error) {
	var list []model.CleaningDecisionRule
	err := r.db.Where("status = 1 AND (station_id = ? OR station_id = '')", stationID).
		Order("priority DESC").Find(&list).Error
	return list, err
}

// ─── 极端天气规则 ────────────────────────────────────────────

func (r *CleaningDecisionRepo) CreateExtremeRule(rule *model.ExtremeWeatherRule) error {
	return r.db.Create(rule).Error
}

func (r *CleaningDecisionRepo) GetExtremeRuleByID(id string) (*model.ExtremeWeatherRule, error) {
	var rule model.ExtremeWeatherRule
	err := r.db.Where("rule_id = ?", id).First(&rule).Error
	return &rule, err
}

func (r *CleaningDecisionRepo) UpdateExtremeRule(rule *model.ExtremeWeatherRule) error {
	return r.db.Save(rule).Error
}

func (r *CleaningDecisionRepo) DeleteExtremeRule(id string) error {
	return r.db.Delete(&model.ExtremeWeatherRule{}, "rule_id = ?", id).Error
}

func (r *CleaningDecisionRepo) ListExtremeRules(stationID string, page, size int) ([]model.ExtremeWeatherRule, int64, error) {
	var list []model.ExtremeWeatherRule
	var total int64
	q := r.db.Model(&model.ExtremeWeatherRule{})
	if stationID != "" {
		q = q.Where("station_id = ? OR station_id = ''", stationID)
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
	err := q.Offset((page - 1) * size).Limit(size).Order("create_time DESC").Find(&list).Error
	return list, total, err
}

// GetExtremeRulesByStationID 获取适用于指定电站的极端天气规则。
func (r *CleaningDecisionRepo) GetExtremeRulesByStationID(stationID string) ([]model.ExtremeWeatherRule, error) {
	var list []model.ExtremeWeatherRule
	err := r.db.Where("status = 1 AND (station_id = ? OR station_id = '')", stationID).Find(&list).Error
	return list, err
}

// ─── 动态频次调整 ────────────────────────────────────────────

func (r *CleaningDecisionRepo) CreateAdjustment(adj *model.CleaningScheduleAdjustment) error {
	return r.db.Create(adj).Error
}

func (r *CleaningDecisionRepo) ListAdjustments(stationID string, page, size int) ([]model.CleaningScheduleAdjustment, int64, error) {
	var list []model.CleaningScheduleAdjustment
	var total int64
	q := r.db.Model(&model.CleaningScheduleAdjustment{}).Where("station_id = ?", stationID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	err := q.Offset((page - 1) * size).Limit(size).Order("create_time DESC").Find(&list).Error
	return list, total, err
}

func (r *CleaningDecisionRepo) UpdateAdjustment(adj *model.CleaningScheduleAdjustment) error {
	return r.db.Save(adj).Error
}
