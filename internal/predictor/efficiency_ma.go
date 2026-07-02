package predictor

import (
	"fmt"
	"math"
	"time"
)

// MovingAveragePredictor 基于加权移动平均的效率预测器。
// 历史效率 = 当日完成任务数 / 当日总任务数 × 100
// 预测值   = 最近 window 天的加权移动平均（越近权重越高）
// 置信区间 = 预测值 ± 1.5 × 历史标准差
type MovingAveragePredictor struct {
	Window int // 移动平均窗口，默认 7
}

func NewMovingAveragePredictor() *MovingAveragePredictor {
	return &MovingAveragePredictor{Window: 7}
}

func (p *MovingAveragePredictor) Predict(history []DailyRecord, forecastDays int) []EfficiencyPoint {
	window := p.Window
	if window <= 0 {
		window = 7
	}

	// 将历史记录转为效率序列
	rates := make([]float64, len(history))
	for i, r := range history {
		if r.Total > 0 {
			rates[i] = math.Round(float64(r.Completed)/float64(r.Total)*10000) / 100
		}
	}

	// 历史标准差，用于置信区间
	stddev := stdDev(rates)

	points := make([]EfficiencyPoint, 0, len(history)+forecastDays)

	// 历史点：actual 填真实值，predicted 填同期移动平均
	for i, r := range history {
		pred := weightedMA(rates, i, window)
		points = append(points, EfficiencyPoint{
			Date:      r.Date,
			Actual:    rates[i],
			Predicted: math.Round(pred*100) / 100,
			Upper:     math.Min(100, math.Round((pred+1.5*stddev)*100)/100),
			Lower:     math.Max(0, math.Round((pred-1.5*stddev)*100)/100),
		})
	}

	// 预测未来 forecastDays 天：actual = 0（无真实值）
	lastPred := weightedMA(rates, len(rates)-1, window)
	today := time.Now()
	for d := 1; d <= forecastDays; d++ {
		date := today.AddDate(0, 0, d).Format("2006-01-02")
		// 轻微衰减：每天预测效率降低 0.2%（模拟不确定性增加）
		pred := math.Max(0, lastPred-float64(d)*0.2)
		halfWidth := 1.5*stddev + float64(d)*0.5 // 预测越远置信区间越宽
		points = append(points, EfficiencyPoint{
			Date:      date,
			Actual:    0,
			Predicted: math.Round(pred*100) / 100,
			Upper:     math.Min(100, math.Round((pred+halfWidth)*100)/100),
			Lower:     math.Max(0, math.Round((pred-halfWidth)*100)/100),
		})
	}

	// 历史数据不足时补全日期（防止前端折线图断裂）
	if len(history) == 0 {
		for d := 29; d >= 0; d-- {
			date := fmt.Sprintf("%s", today.AddDate(0, 0, -d).Format("2006-01-02"))
			points = append(points, EfficiencyPoint{Date: date})
		}
	}

	return points
}

// weightedMA 计算到第 i 个元素为止、最近 window 个值的线性加权平均。
func weightedMA(rates []float64, i, window int) float64 {
	start := i - window + 1
	if start < 0 {
		start = 0
	}
	slice := rates[start : i+1]
	n := len(slice)
	if n == 0 {
		return 75 // 无数据时给默认基准值
	}
	var weightSum, valSum float64
	for j, v := range slice {
		w := float64(j + 1)
		weightSum += w
		valSum += w * v
	}
	return valSum / weightSum
}

func stdDev(vals []float64) float64 {
	if len(vals) == 0 {
		return 5
	}
	var sum float64
	for _, v := range vals {
		sum += v
	}
	mean := sum / float64(len(vals))
	var variance float64
	for _, v := range vals {
		diff := v - mean
		variance += diff * diff
	}
	return math.Sqrt(variance / float64(len(vals)))
}
