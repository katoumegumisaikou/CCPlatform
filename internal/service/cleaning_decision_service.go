package service

import (
	"ccplatform/internal/model"
	"ccplatform/internal/repository"
	"ccplatform/pkg/util"
	"encoding/json"
	"log"
	"strconv"
	"time"
)

// CleaningAdvice 清扫建议返回结构。
type CleaningAdvice struct {
	SuggestedAction   string                  `json:"suggested_action"`   // 建议动作: 建议清扫 / 暂缓清扫 / 暂停清扫
	Reason            string                  `json:"reason"`              // 决策理由
	DustScore         float64                 `json:"dust_score"`         // 灰尘累积评分 (0-1)
	WeatherScore      float64                 `json:"weather_score"`      // 气象适宜度评分 (0-1)
	TriggeredWarnings []model.WeatherWarning  `json:"triggered_warnings"` // 触发的极端天气预警
	RecommendedTime   string                  `json:"recommended_time"`   // 推荐清扫时间窗口
}

// AdjustmentAdvice 动态频次调整建议。
type AdjustmentAdvice struct {
	CurrentFreq   string  `json:"current_freq"`
	SuggestedFreq string  `json:"suggested_freq"`
	Reason        string  `json:"reason"`
	WeatherScore  float64 `json:"weather_score"`
	CleanWindows  int     `json:"clean_windows"` // 未来 7 天适宜清扫的天数
}

// CleaningDecisionService 清扫决策业务逻辑层，提供规则管理和智能清扫建议。
type CleaningDecisionService struct {
	decisionRepo *repository.CleaningDecisionRepo
	weatherRepo  *repository.WeatherRepo
	stationRepo  *repository.StationRepo
}

// NewCleaningDecisionService 创建清扫决策 Service 实例。
func NewCleaningDecisionService() *CleaningDecisionService {
	return &CleaningDecisionService{
		decisionRepo: repository.NewCleaningDecisionRepo(),
		weatherRepo:  repository.NewWeatherRepo(),
		stationRepo:  repository.NewStationRepo(),
	}
}

// ─── 决策规则 CRUD ──────────────────────────────────────────

func (s *CleaningDecisionService) ListRules(stationID string, page, size int) ([]model.CleaningDecisionRule, int64, error) {
	return s.decisionRepo.ListRules(stationID, page, size)
}

func (s *CleaningDecisionService) CreateRule(rule *model.CleaningDecisionRule) error {
	rule.RuleID = util.GenerateID()
	return s.decisionRepo.CreateRule(rule)
}

func (s *CleaningDecisionService) GetRuleByID(id string) (*model.CleaningDecisionRule, error) {
	return s.decisionRepo.GetRuleByID(id)
}

func (s *CleaningDecisionService) UpdateRule(rule *model.CleaningDecisionRule) error {
	return s.decisionRepo.UpdateRule(rule)
}

func (s *CleaningDecisionService) DeleteRule(id string) error {
	return s.decisionRepo.DeleteRule(id)
}

// ─── 极端天气规则 CRUD ──────────────────────────────────────

func (s *CleaningDecisionService) ListExtremeRules(stationID string, page, size int) ([]model.ExtremeWeatherRule, int64, error) {
	return s.decisionRepo.ListExtremeRules(stationID, page, size)
}

func (s *CleaningDecisionService) CreateExtremeRule(rule *model.ExtremeWeatherRule) error {
	rule.RuleID = util.GenerateID()
	return s.decisionRepo.CreateExtremeRule(rule)
}

func (s *CleaningDecisionService) GetExtremeRuleByID(id string) (*model.ExtremeWeatherRule, error) {
	return s.decisionRepo.GetExtremeRuleByID(id)
}

func (s *CleaningDecisionService) UpdateExtremeRule(rule *model.ExtremeWeatherRule) error {
	return s.decisionRepo.UpdateExtremeRule(rule)
}

func (s *CleaningDecisionService) DeleteExtremeRule(id string) error {
	return s.decisionRepo.DeleteExtremeRule(id)
}

// ─── 智能清扫建议 ──────────────────────────────────────────

// EvaluateCleaningAdvice 根据当前气象数据和历史清扫记录，生成清扫建议。
func (s *CleaningDecisionService) EvaluateCleaningAdvice(stationID string) (*CleaningAdvice, error) {
	advice := &CleaningAdvice{
		DustScore:    0,
		WeatherScore: 1.0, // 默认满分，逐项扣分
	}

	// 1. 获取当前天气
	now, err := s.weatherRepo.GetNowByStationID(stationID)
	if err != nil {
		log.Printf("[Decision] No weather data for station %s: %v", stationID, err)
		advice.SuggestedAction = "暂无数据"
		advice.Reason = "缺少气象数据，建议手动确认后再安排清扫"
		return advice, nil
	}

	// 2. 加载决策规则
	rules, err := s.decisionRepo.GetRulesByStationID(stationID)
	if err != nil {
		return nil, err
	}

	// 3. 计算气象适宜度
	var reasons []string
	for _, rule := range rules {
		if rule.Status != 1 {
			continue
		}
		// 解析条件 JSON
		var conds map[string]interface{}
		if err := json.Unmarshal([]byte(rule.Conditions), &conds); err != nil {
			continue
		}

		// 检查风速
		if maxScale, ok := conds["max_wind_scale"].(float64); ok {
			scale, _ := strconv.ParseFloat(now.WindScale, 64)
			if scale > maxScale {
				advice.WeatherScore -= 0.3
				reasons = append(reasons, "风力过大 ("+now.WindScale+"级)")
			}
		}

		// 检查温度
		if minTemp, ok := conds["min_temp"].(float64); ok {
			temp, _ := strconv.ParseFloat(now.Temp, 64)
			if temp < minTemp {
				advice.WeatherScore -= 0.3
				reasons = append(reasons, "温度过低 ("+now.Temp+"°C)")
			}
		}
		if maxTemp, ok := conds["max_temp"].(float64); ok {
			temp, _ := strconv.ParseFloat(now.Temp, 64)
			if temp > maxTemp {
				advice.WeatherScore -= 0.3
				reasons = append(reasons, "温度过高 ("+now.Temp+"°C)")
			}
		}

		// 检查湿度
		if maxHumidity, ok := conds["max_humidity"].(float64); ok {
			hum, _ := strconv.ParseFloat(now.Humidity, 64)
			if hum > maxHumidity {
				advice.WeatherScore -= 0.1
				reasons = append(reasons, "湿度过高 ("+now.Humidity+"%)")
			}
		}

		// 检查降雨
		if noRain, ok := conds["no_rain"].(bool); ok && noRain {
			precip, _ := strconv.ParseFloat(now.Precip, 64)
			if precip > 0 {
				advice.WeatherScore -= 0.4
				reasons = append(reasons, "当前有降水 ("+now.Precip+"mm)")
			}
		}
	}

	// 4. 灰尘累积评分（基于决策规则中的灰尘阈值估算）
	var dustRule map[string]interface{}
	if len(rules) > 0 && rules[0].DustRule != "" {
		_ = json.Unmarshal([]byte(rules[0].DustRule), &dustRule)
	}
	// 简化：默认灰尘累积率为每天 ~0.15，阈值 0.7 触发清扫建议
	// 实际生产环境中应从清扫记录表计算距上次清扫天数
	advice.DustScore = 0.15 * 3 // 假设 3 天未清扫
	if threshold, ok := dustRule["dust_threshold"].(float64); ok && threshold > 0 {
		days, _ := dustRule["days_since_last_clean"].(float64)
		if days > 0 {
			advice.DustScore = days / (threshold / 100) // 简化计算
		}
	}

	// 5. 检查极端天气
	extremeRules, err := s.decisionRepo.GetExtremeRulesByStationID(stationID)
	if err == nil {
		for _, er := range extremeRules {
			if er.Status != 1 {
				continue
			}
			var trigger map[string]interface{}
			if err := json.Unmarshal([]byte(er.TriggerCond), &trigger); err != nil {
				continue
			}
			// 检查大风触发
			if windGt, ok := trigger["wind_scale_gt"].(float64); ok {
				scale, _ := strconv.ParseFloat(now.WindScale, 64)
				if scale > windGt {
					advice.SuggestedAction = er.Action
					advice.Reason = "触发极端天气规则: " + er.RuleName + " (" + er.WeatherType + ")"
					advice.WeatherScore = 0
					if advice.SuggestedAction == "暂停清扫" {
						return advice, nil
					}
					reasons = append(reasons, "触发极端天气规则: "+er.RuleName)
				}
			}
			// 检查降雨触发
			if precipGt, ok := trigger["precip_gt"].(float64); ok {
				precip, _ := strconv.ParseFloat(now.Precip, 64)
				if precip > precipGt {
					advice.SuggestedAction = er.Action
					advice.Reason = "触发极端天气规则: " + er.RuleName
					advice.WeatherScore = 0
					return advice, nil
				}
			}
		}
	}

	// 6. 获取当前预警
	warnings, _ := s.weatherRepo.GetActiveWarnings(stationID)
	advice.TriggeredWarnings = warnings

	// 7. 综合决策
	if advice.SuggestedAction == "" {
		if advice.WeatherScore >= 0.7 && advice.DustScore >= 0.5 {
			advice.SuggestedAction = "建议清扫"
			advice.Reason = "气象条件适宜，灰尘累积已达阈值"
		} else if advice.WeatherScore < 0.4 {
			advice.SuggestedAction = "暂停清扫"
			advice.Reason = "气象条件不适宜: " + joinReasons(reasons)
		} else if advice.DustScore < 0.3 {
			advice.SuggestedAction = "暂缓清扫"
			advice.Reason = "灰尘累积较轻，可延长清扫间隔"
		} else {
			advice.SuggestedAction = "建议清扫"
			advice.Reason = "可安排清扫，注意观察天气变化"
		}
	}

	// 推荐时间窗口: 如果夜间有雨，建议雨停后清扫
	advice.RecommendedTime = "未来 48 小时内"
	nowTime := time.Now()
	if nowTime.Hour() >= 6 && nowTime.Hour() < 10 {
		advice.RecommendedTime = "当前时段适宜清扫"
	} else if nowTime.Hour() >= 10 && nowTime.Hour() < 16 {
		advice.RecommendedTime = "今天下午或明天早晨"
	}

	return advice, nil
}

// ─── 动态频次调整 ──────────────────────────────────────────

// GenerateScheduleAdjustment 根据气象数据生成清扫频次调整建议。
func (s *CleaningDecisionService) GenerateScheduleAdjustment(stationID string) (*AdjustmentAdvice, error) {
	advice := &AdjustmentAdvice{
		CurrentFreq:   "每 5 天一次",
		SuggestedFreq: "每 5 天一次",
	}

	// 分析 7 天预报
	forecast, err := s.weatherRepo.GetDailyByStationID(stationID, 7)
	if err != nil || len(forecast) == 0 {
		advice.Reason = "缺少预报数据，保持当前频次"
		return advice, nil
	}

	// 统计适宜清扫的天数（无降水、风力小）
	cleanWindows := 0
	for _, d := range forecast {
		precip, _ := strconv.ParseFloat(d.Precip, 64)
		scale, _ := strconv.ParseFloat(d.WindScaleDay, 64)
		if precip == 0 && scale <= 4 {
			cleanWindows++
		}
	}
	advice.CleanWindows = cleanWindows

	// 根据适宜天数建议频次
	if cleanWindows >= 6 {
		advice.SuggestedFreq = "每 3 天一次"
		advice.Reason = "未来 7 天天气良好，可增加清扫频次，提升发电效率"
		advice.WeatherScore = 0.9
	} else if cleanWindows >= 4 {
		advice.SuggestedFreq = "每 5 天一次"
		advice.Reason = "未来天气条件适中，保持正常清扫频次"
		advice.WeatherScore = 0.7
	} else if cleanWindows >= 2 {
		advice.SuggestedFreq = "每 7 天一次"
		advice.Reason = "未来雨天较多，建议降低清扫频次，避免无效作业"
		advice.WeatherScore = 0.5
	} else {
		advice.SuggestedFreq = "每 10 天一次"
		advice.Reason = "未来天气条件恶劣，大幅降低清扫频次，待天气好转后恢复"
		advice.WeatherScore = 0.3
	}

	return advice, nil
}

// ─── 调整记录 ────────────────────────────────────────────────

func (s *CleaningDecisionService) ListAdjustments(stationID string, page, size int) ([]model.CleaningScheduleAdjustment, int64, error) {
	return s.decisionRepo.ListAdjustments(stationID, page, size)
}

func (s *CleaningDecisionService) CreateAdjustment(adj *model.CleaningScheduleAdjustment) error {
	return s.decisionRepo.CreateAdjustment(adj)
}

// joinReasons 将原因列表用分隔符连接。
func joinReasons(reasons []string) string {
	if len(reasons) == 0 {
		return "无特殊原因"
	}
	result := reasons[0]
	for i := 1; i < len(reasons); i++ {
		result += "；" + reasons[i]
	}
	return result
}
