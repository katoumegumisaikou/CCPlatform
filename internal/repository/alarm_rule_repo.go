package repository

import (
	"ccplatform/internal/model"

	"gorm.io/gorm"
)

// AlarmRuleRepo 告警规则数据访问层。
type AlarmRuleRepo struct {
	db *gorm.DB
}

func NewAlarmRuleRepo() *AlarmRuleRepo {
	return &AlarmRuleRepo{db: DB}
}

func (r *AlarmRuleRepo) Create(rule *model.AlarmRule) error {
	return r.db.Create(rule).Error
}

func (r *AlarmRuleRepo) GetByID(id string) (*model.AlarmRule, error) {
	var rule model.AlarmRule
	err := r.db.Where("rule_id = ?", id).First(&rule).Error
	return &rule, err
}

func (r *AlarmRuleRepo) Update(rule *model.AlarmRule) error {
	return r.db.Save(rule).Error
}

func (r *AlarmRuleRepo) Delete(id string) error {
	return r.db.Where("rule_id = ?", id).Delete(&model.AlarmRule{}).Error
}

func (r *AlarmRuleRepo) List(page, size int, alarmType string) ([]model.AlarmRule, int64, error) {
	var rules []model.AlarmRule
	var total int64
	query := r.db.Model(&model.AlarmRule{})
	if alarmType != "" {
		query = query.Where("alarm_type = ?", alarmType)
	}
	query.Count(&total)
	err := query.Offset((page - 1) * size).Limit(size).Order("create_time DESC").Find(&rules).Error
	return rules, total, err
}

// GetActiveRules 查询所有启用状态的规则。
func (r *AlarmRuleRepo) GetActiveRules() ([]model.AlarmRule, error) {
	var rules []model.AlarmRule
	err := r.db.Where("status = 1").Find(&rules).Error
	return rules, err
}

// GetByAlarmType 查询指定告警类型的启用规则。
func (r *AlarmRuleRepo) GetByAlarmType(alarmType string) ([]model.AlarmRule, error) {
	var rules []model.AlarmRule
	err := r.db.Where("alarm_type = ? AND status = 1", alarmType).Find(&rules).Error
	return rules, err
}
