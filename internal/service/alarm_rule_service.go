package service

import (
	"ccplatform/internal/model"
	"ccplatform/internal/repository"
	"fmt"
)

// AlarmRuleService 告警规则业务逻辑层。
type AlarmRuleService struct {
	repo *repository.AlarmRuleRepo
}

func NewAlarmRuleService() *AlarmRuleService {
	return &AlarmRuleService{repo: repository.NewAlarmRuleRepo()}
}

func (s *AlarmRuleService) Create(rule *model.AlarmRule) error {
	rule.RuleID = generateID()
	return s.repo.Create(rule)
}

func (s *AlarmRuleService) Update(rule *model.AlarmRule) error {
	_, err := s.repo.GetByID(rule.RuleID)
	if err != nil {
		return fmt.Errorf("rule not found: %w", err)
	}
	return s.repo.Update(rule)
}

func (s *AlarmRuleService) Delete(id string) error {
	return s.repo.Delete(id)
}

func (s *AlarmRuleService) List(page, size int, alarmType string) ([]model.AlarmRule, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	return s.repo.List(page, size, alarmType)
}

// EvaluateRules 评估告警规则，返回是否需要升级。
func (s *AlarmRuleService) EvaluateRules(alarmType string) (*model.AlarmRule, error) {
	rules, err := s.repo.GetByAlarmType(alarmType)
	if err != nil || len(rules) == 0 {
		return nil, nil
	}
	for _, rule := range rules {
		if rule.Status == 1 {
			return &rule, nil
		}
	}
	return nil, nil
}
