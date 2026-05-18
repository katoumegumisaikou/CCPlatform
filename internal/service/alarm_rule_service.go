package service

import (
	"ccplatform/internal/model"
	"ccplatform/internal/repository"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// EscalationEntry 升级规则条目：在 within_min 分钟内发生 count 次同类型告警，则将级别升级到 to_level。
type EscalationEntry struct {
	Count     int `json:"count"`
	WithinMin int `json:"within_min"`
	ToLevel   int8 `json:"to_level"`
}

// SuppressionRule 抑制规则：同类型告警在 within_sec 秒内只生成一条。
type SuppressionRule struct {
	WithinSec int `json:"within_sec"`
}

// EvalResult 告警规则评估结果。
type EvalResult struct {
	ShouldCreate bool  // false 表示被抑制，不应创建告警
	FinalLevel   int8  // 升级后的最终级别（可能与原始级别相同）
	Escalated    bool  // 是否触发了升级
}

// AlarmRuleService 告警规则业务逻辑层。
type AlarmRuleService struct {
	repo      *repository.AlarmRuleRepo
	alarmRepo *repository.AlarmRepo
}

func NewAlarmRuleService() *AlarmRuleService {
	return &AlarmRuleService{
		repo:      repository.NewAlarmRuleRepo(),
		alarmRepo: repository.NewAlarmRepo(),
	}
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

// EvaluateRules 评估告警规则，返回是否需要创建告警以及最终级别。
// scopeMatching: robot → station → global，越具体优先级越高。
// 评估流程: 1.作用域匹配 → 2.抑制检查 → 3.升级检查 → 4.阈值检查
func (s *AlarmRuleService) EvaluateRules(alarmType, robotID, stationID string, originalLevel int8) (*EvalResult, error) {
	rules, err := s.repo.GetByAlarmType(alarmType)
	if err != nil || len(rules) == 0 {
		return &EvalResult{ShouldCreate: true, FinalLevel: originalLevel}, nil
	}

	rule := pickBestRule(rules, robotID, stationID)
	if rule == nil {
		return &EvalResult{ShouldCreate: true, FinalLevel: originalLevel}, nil
	}

	result := &EvalResult{ShouldCreate: true, FinalLevel: originalLevel}

	// 1. 抑制检查：同类型告警在抑制窗口内不再重复创建
	if rule.SuppressionRule != "" {
		var suppr SuppressionRule
		if err := json.Unmarshal([]byte(rule.SuppressionRule), &suppr); err == nil && suppr.WithinSec > 0 {
			latestTime, err := s.alarmRepo.GetLatestAlarmTime(robotID, alarmType)
			if err == nil && latestTime != nil {
				elapsed := time.Since(*latestTime)
				if elapsed < time.Duration(suppr.WithinSec)*time.Second {
					log.Printf("[AlarmRule] Suppressed: %s for robot %s (last: %v, %.0fs ago, window: %ds)",
						alarmType, robotID, latestTime, elapsed.Seconds(), suppr.WithinSec)
					result.ShouldCreate = false
					return result, nil
				}
			}
		}
	}

	// 2. 升级检查：最近 N 分钟内同类型告警次数达到阈值，自动升级级别
	if rule.EscalationRule != "" {
		var entries []EscalationEntry
		if err := json.Unmarshal([]byte(rule.EscalationRule), &entries); err == nil {
			for _, entry := range entries {
				count, err := s.alarmRepo.CountRecentByType(robotID, alarmType, entry.WithinMin)
				if err == nil && int(count) >= entry.Count && entry.ToLevel > result.FinalLevel {
					oldLevel := result.FinalLevel
					result.FinalLevel = entry.ToLevel
					result.Escalated = true
					log.Printf("[AlarmRule] Escalated: %s for robot %s level %d→%d (count=%d, threshold=%d, window=%dmin)",
						alarmType, robotID, oldLevel, result.FinalLevel, count, entry.Count, entry.WithinMin)
					break
				}
			}
		}
	}

	// 3. 阈值检查：解析阈值 JSON，与当前状态数据比对
	//    由 MQTT handleStatus 调用时传入状态数据后生效，此入口留作后续扩展

	return result, nil
}

// pickBestRule 按 robot > station > global 优先级选取最匹配的规则。
func pickBestRule(rules []model.AlarmRule, robotID, stationID string) *model.AlarmRule {
	var globalRule, stationRule, robotRule *model.AlarmRule
	for i := range rules {
		r := &rules[i]
		switch r.ScopeType {
		case "robot":
			if r.ScopeID == robotID {
				robotRule = r
			}
		case "station":
			if r.ScopeID == stationID {
				stationRule = r
			}
		default:
			globalRule = r
		}
	}
	if robotRule != nil {
		return robotRule
	}
	if stationRule != nil {
		return stationRule
	}
	return globalRule
}
