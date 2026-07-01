package service

import (
	"ccplatform/internal/predictor"
	"ccplatform/internal/repository"
)

// PredictionService 趋势预测服务。
// effPredictor 和 stratOptimizer 通过接口注入，换算法只需替换实现，service 层不动。
type PredictionService struct {
	stationRepo    *repository.StationRepo
	robotRepo      *repository.RobotRepo
	taskRepo       *repository.TaskRepo
	predRepo       *repository.FaultPredictionRepo
	effPredictor   predictor.EfficiencyPredictor
	stratOptimizer predictor.StrategyOptimizer
}

func NewPredictionService() *PredictionService {
	return &PredictionService{
		stationRepo:    repository.NewStationRepo(),
		robotRepo:      repository.NewRobotRepo(),
		taskRepo:       repository.NewTaskRepo(),
		predRepo:       repository.NewFaultPredictionRepo(),
		effPredictor:   predictor.NewMovingAveragePredictor(),
		stratOptimizer: predictor.NewRuleBasedStrategyOptimizer(),
	}
}

// PredictEfficiency 效率预测：返回过去 30 天历史 + 未来 7 天预测的时序数据点。
func (s *PredictionService) PredictEfficiency(stationID string) ([]predictor.EfficiencyPoint, error) {
	rows, err := s.taskRepo.GetDailyStatsByStation(stationID, 30)
	if err != nil {
		return nil, err
	}

	history := make([]predictor.DailyRecord, len(rows))
	for i, r := range rows {
		history[i] = predictor.DailyRecord{
			Date:      r.Date,
			Total:     r.Total,
			Completed: r.Completed,
		}
	}

	return s.effPredictor.Predict(history, 7), nil
}

// OptimizeStrategy 策略优化：返回基于当前电站状态的策略建议列表。
func (s *PredictionService) OptimizeStrategy(stationID string) ([]predictor.StrategyItem, error) {
	robots, err := s.robotRepo.GetByStationID(stationID)
	if err != nil {
		return nil, err
	}

	snap := predictor.StationSnapshot{
		StationID:   stationID,
		TotalRobots: len(robots),
	}
	for _, r := range robots {
		if r.OnlineStatus == 1 {
			snap.OnlineCount++
			snap.AvgBattery += float64(r.BatteryLevel)
			if r.WorkStatus == 0 {
				snap.IdleCount++
			}
			if r.WorkStatus == 2 { // 充电中
				snap.ChargingCount++
			}
		}
	}
	if snap.OnlineCount > 0 {
		snap.AvgBattery /= float64(snap.OnlineCount)
	}

	// 活跃故障预测风险统计。GetByStationID 已限定 status IN (0,1)，
	// 已生成工单的预测仍代表设备风险，不能从调度策略里排除。
	preds, _ := s.predRepo.GetByStationID(stationID)
	snap.ActiveFaultPredictions = len(preds)
	highRiskRobots := make(map[string]struct{})
	mediumRiskRobots := make(map[string]struct{})
	for _, p := range preds {
		switch p.RiskLevel {
		case "high":
			highRiskRobots[p.RobotID] = struct{}{}
		case "medium":
			mediumRiskRobots[p.RobotID] = struct{}{}
		}
	}
	snap.HighFaultRobots = len(highRiskRobots)
	snap.MediumFaultRobots = len(mediumRiskRobots)

	return s.stratOptimizer.Optimize(snap), nil
}
