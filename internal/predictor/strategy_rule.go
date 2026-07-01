package predictor

import "fmt"

// RuleBasedStrategyOptimizer 基于规则的策略优化器。
// 未来可替换为强化学习或线性规划模型，只需实现 StrategyOptimizer 接口。
type RuleBasedStrategyOptimizer struct{}

func NewRuleBasedStrategyOptimizer() *RuleBasedStrategyOptimizer {
	return &RuleBasedStrategyOptimizer{}
}

func (o *RuleBasedStrategyOptimizer) Optimize(snap StationSnapshot) []StrategyItem {
	items := make([]StrategyItem, 0, 5)
	id := 1

	onlineRate := 0.0
	if snap.TotalRobots > 0 {
		onlineRate = float64(snap.OnlineCount) / float64(snap.TotalRobots) * 100
	}

	// 规则1：在线率低
	if onlineRate < 60 {
		items = append(items, StrategyItem{
			ID:                  fmt.Sprintf("%d", id),
			Title:               "提升机器人在线率",
			Description:         fmt.Sprintf("当前在线率 %.0f%%，低于建议阈值 60%%。建议检查离线机器人网络连接状态，优先恢复通信中断设备。", onlineRate),
			Priority:            "high",
			ExpectedImprovement: "可提升整体清扫覆盖率 20-30%",
		})
		id++
	}

	// 规则2：充电中机器人占比高（充电仓可能成为瓶颈）
	chargingRate := 0.0
	if snap.OnlineCount > 0 {
		chargingRate = float64(snap.ChargingCount) / float64(snap.OnlineCount) * 100
	}
	if chargingRate > 40 {
		items = append(items, StrategyItem{
			ID:                  fmt.Sprintf("%d", id),
			Title:               "错峰充电，释放作业资源",
			Description:         fmt.Sprintf("在线机器人中 %.0f%% 处于充电状态，建议按电量梯度错峰充电，优先派遣电量充足设备出勤，减少空闲等待。", chargingRate),
			Priority:            "medium",
			ExpectedImprovement: "可增加同时作业机器人数量 15-25%",
		})
		id++
	}

	// 规则3：存在高风险故障预测
	if snap.HighFaultRobots > 0 {
		items = append(items, StrategyItem{
			ID:                  fmt.Sprintf("%d", id),
			Title:               "预防性维护优先调度",
			Description:         fmt.Sprintf("当前有 %d 台机器人存在高风险故障预测，活跃预测共 %d 条。建议降低风险设备作业强度，优先调度健康状态良好的设备，并安排预防性维护窗口。", snap.HighFaultRobots, snap.ActiveFaultPredictions),
			Priority:            "high",
			ExpectedImprovement: "可降低非计划停机率 40-50%",
		})
		id++
	} else if snap.MediumFaultRobots > 0 {
		items = append(items, StrategyItem{
			ID:                  fmt.Sprintf("%d", id),
			Title:               "关注中风险设备负载",
			Description:         fmt.Sprintf("当前有 %d 台机器人存在中风险故障预测，活跃预测共 %d 条。建议减少连续高负载任务，结合维护窗口进行巡检。", snap.MediumFaultRobots, snap.ActiveFaultPredictions),
			Priority:            "medium",
			ExpectedImprovement: "可降低故障升级概率 20-30%",
		})
		id++
	}

	// 规则4：平均电量低
	if snap.AvgBattery < 50 {
		items = append(items, StrategyItem{
			ID:                  fmt.Sprintf("%d", id),
			Title:               "优化充电调度策略",
			Description:         fmt.Sprintf("在线机器人平均电量 %.0f%%，建议调整充电调度策略：电量低于 30%% 时立即返回充电，电量高于 80%% 时优先出勤，避免电量波动影响任务完成率。", snap.AvgBattery),
			Priority:            "medium",
			ExpectedImprovement: "可提升任务完成率 10-15%",
		})
		id++
	}

	// 规则5：空闲机器人多（作业效率浪费）
	idleRate := 0.0
	if snap.OnlineCount > 0 {
		idleRate = float64(snap.IdleCount) / float64(snap.OnlineCount) * 100
	}
	if idleRate > 50 && snap.IdleCount > 2 {
		items = append(items, StrategyItem{
			ID:                  fmt.Sprintf("%d", id),
			Title:               "增加任务派发密度",
			Description:         fmt.Sprintf("当前有 %d 台机器人空闲（占在线设备 %.0f%%），建议加大任务派发频率，或启用自动巡检模式，充分利用设备产能。", snap.IdleCount, idleRate),
			Priority:            "low",
			ExpectedImprovement: "可提升设备利用率 20-35%",
		})
		id++
	}

	// 兜底：状态良好时给正向反馈
	if len(items) == 0 {
		items = append(items, StrategyItem{
			ID:                  "1",
			Title:               "设备运行状态良好",
			Description:         fmt.Sprintf("电站共 %d 台机器人，在线率 %.0f%%，平均电量 %.0f%%，当前无需重大调度调整，维持现有运行策略。", snap.TotalRobots, onlineRate, snap.AvgBattery),
			Priority:            "low",
			ExpectedImprovement: "保持当前清扫效率",
		})
	}

	return items
}
