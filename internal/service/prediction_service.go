package service

import (
	"ccplatform/internal/repository"
	"math"
)

// PredictionService 趋势预测服务，提供简化的效率预测、故障预测和策略优化。
type PredictionService struct {
	stationRepo *repository.StationRepo
	robotRepo   *repository.RobotRepo
	taskRepo    *repository.TaskRepo
}

func NewPredictionService() *PredictionService {
	return &PredictionService{
		stationRepo: repository.NewStationRepo(),
		robotRepo:   repository.NewRobotRepo(),
		taskRepo:    repository.NewTaskRepo(),
	}
}

// PredictEfficiency 基于移动平均法预测下一期清扫效率。
func (s *PredictionService) PredictEfficiency(stationID string) (map[string]interface{}, error) {
	robots, err := s.robotRepo.GetByStationID(stationID)
	if err != nil {
		return nil, err
	}

	// 简化的移动平均预测
	var totalBattery float64
	onlineCount := 0
	for _, r := range robots {
		if r.OnlineStatus == 1 {
			onlineCount++
			totalBattery += float64(r.BatteryLevel)
		}
	}

	avgBattery := 0.0
	if onlineCount > 0 {
		avgBattery = totalBattery / float64(onlineCount)
	}

	predictedEfficiency := 2000.0 * (avgBattery / 100.0) // 基准效率 × 电量系数

	return map[string]interface{}{
		"station_id":             stationID,
		"predicted_efficiency":   math.Round(predictedEfficiency*100) / 100,
		"online_robot_count":     onlineCount,
		"avg_battery_level":      math.Round(avgBattery*100) / 100,
		"confidence":             "moderate",
	}, nil
}

// PredictFaultProbability 预测机器人故障概率（简化线性模型）。
func (s *PredictionService) PredictFaultProbability(robotID string) (map[string]interface{}, error) {
	robot, err := s.robotRepo.GetByID(robotID)
	if err != nil {
		return nil, err
	}

	// 简化模型: 电量越低故障风险越高 + 工作状态
	faultProb := 0.0
	if robot.OnlineStatus == 1 {
		faultProb = math.Max(0, (100-float64(robot.BatteryLevel))/100*0.3)
		if robot.WorkStatus == 1 {
			faultProb += 0.1
		}
	}
	faultProb = math.Min(faultProb, 1.0)

	riskLevel := "low"
	if faultProb > 0.5 {
		riskLevel = "high"
	} else if faultProb > 0.2 {
		riskLevel = "medium"
	}

	return map[string]interface{}{
		"robot_id":       robotID,
		"fault_probability": math.Round(faultProb*100) / 100,
		"risk_level":     riskLevel,
	}, nil
}

// OptimizeStrategy 最优清扫策略推荐（简化版）。
func (s *PredictionService) OptimizeStrategy(stationID string) (map[string]interface{}, error) {
	robots, err := s.robotRepo.GetByStationID(stationID)
	if err != nil {
		return nil, err
	}

	// 策略: 选择电量最高且在线的机器人优先清扫
	var bestRobotID string
	var maxBattery int8
	for _, r := range robots {
		if r.OnlineStatus == 1 && r.WorkStatus == 0 && r.BatteryLevel > maxBattery {
			maxBattery = r.BatteryLevel
			bestRobotID = r.RobotID
		}
	}

	strategy := map[string]interface{}{
		"station_id": stationID,
		"total_robots": len(robots),
	}
	if bestRobotID != "" {
		strategy["recommended_robot"] = bestRobotID
		strategy["reason"] = "highest battery among idle online robots"
	} else {
		strategy["recommended_robot"] = nil
		strategy["reason"] = "no idle online robot with sufficient battery"
	}

	return strategy, nil
}
