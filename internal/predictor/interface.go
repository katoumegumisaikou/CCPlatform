package predictor

// EfficiencyPoint 单日效率数据点，用于前端时序折线图。
type EfficiencyPoint struct {
	Date      string  `json:"date"`
	Actual    float64 `json:"actual"`    // 历史实际效率（%），未来日期为 0
	Predicted float64 `json:"predicted"` // 预测效率（%）
	Upper     float64 `json:"upper"`     // 置信区间上限
	Lower     float64 `json:"lower"`     // 置信区间下限
}

// StrategyItem 单条策略优化建议。
type StrategyItem struct {
	ID                  string `json:"id"`
	Title               string `json:"title"`
	Description         string `json:"description"`
	Priority            string `json:"priority"` // high / medium / low
	ExpectedImprovement string `json:"expected_improvement"`
}

// StationSnapshot 传给预测器的电站当前状态快照。
type StationSnapshot struct {
	StationID              string
	TotalRobots            int
	OnlineCount            int
	IdleCount              int
	ChargingCount          int
	AvgBattery             float64
	HighFaultRobots        int // 当前存在高风险活跃故障预测的机器人数量
	MediumFaultRobots      int // 当前存在中风险活跃故障预测的机器人数量
	ActiveFaultPredictions int // 当前活跃故障预测总数
}

// DailyRecord 历史每日记录。
type DailyRecord struct {
	Date      string
	Total     int64
	Completed int64
}

// EfficiencyPredictor 效率预测算法接口，新算法（LSTM等）实现此接口即可替换。
type EfficiencyPredictor interface {
	// Predict 接收过去 N 天的每日任务记录，返回历史+预测的时序数据点列表。
	Predict(history []DailyRecord, forecastDays int) []EfficiencyPoint
}

// StrategyOptimizer 策略优化算法接口。
type StrategyOptimizer interface {
	// Optimize 根据电站快照生成策略建议列表。
	Optimize(snap StationSnapshot) []StrategyItem
}
