package service

import (
	"ccplatform/internal/repository"
	"fmt"
	"strconv"
	"time"
)

// AnalyticsService 数据分析业务逻辑层，计算经济技术指标和生成运行报表。
// 对应需求文档中的"经济技术指标"和"数据统计分析"模块。
type AnalyticsService struct {
	stationRepo  *repository.StationRepo
	robotRepo    *repository.RobotRepo
	taskRepo     *repository.TaskRepo
	cleanRepo    *repository.CleaningRecordRepo
}

// NewAnalyticsService 创建 AnalyticsService 实例。
func NewAnalyticsService() *AnalyticsService {
	return &AnalyticsService{
		stationRepo: repository.NewStationRepo(),
		robotRepo:   repository.NewRobotRepo(),
		taskRepo:    repository.NewTaskRepo(),
		cleanRepo:   repository.NewCleaningRecordRepo(),
	}
}

// EconomicReport 经济效益报表，包含电费增收、人工替代率、投资回收期等指标。
type EconomicReport struct {
	StationID        string  `json:"station_id"`         // 电站 ID
	StationName      string  `json:"station_name"`       // 电站名称
	TotalCleanArea   float64 `json:"total_clean_area"`   // 累计清扫面积(㎡)
	ElectricRevenue  float64 `json:"electric_revenue"`   // 电费增收额(元)
	LaborReplaceRate float64 `json:"labor_replace_rate"` // 人工替代率(%)
	UnitCost         float64 `json:"unit_cost"`          // 单位成本(元/㎡)
	InvestPayback    float64 `json:"invest_payback"`     // 投资回收期(年)
	Coverage         float64 `json:"coverage"`           // 清扫覆盖率(%)
}

// EfficiencyReport 效率分析报表，包含清扫效率、覆盖率、设备利用率等指标。
type EfficiencyReport struct {
	StationID       string  `json:"station_id"`        // 电站 ID
	StationName     string  `json:"station_name"`      // 电站名称
	CleanEfficiency float64 `json:"clean_efficiency"`  // 清扫效率(㎡/h)
	CoverageRate    float64 `json:"coverage_rate"`     // 覆盖率(%)
	RepeatRate      float64 `json:"repeat_rate"`       // 重复清扫率(%)
	OnlineRate      float64 `json:"online_rate"`       // 设备在线率(%)
	UtilizationRate float64 `json:"utilization_rate"`  // 设备利用率(%)
}

// GetEconomicReport 生成指定电站的经济效益报表。
// 计算模型: 电费增收 = 清扫面积 × 单价, 投资回收期 = 设备投资 / 年净收益。
// 注: 当前使用简化模型，实际应接入电价数据和历史发电数据。
func (s *AnalyticsService) GetEconomicReport(stationID string, startDate, endDate time.Time) (*EconomicReport, error) {
	station, err := s.stationRepo.GetByID(stationID)
	if err != nil {
		return nil, err
	}

	totalCleanArea, err := s.taskRepo.SumCleanAreaByStation(stationID)
	if err != nil {
		return nil, err
	}

	// 简化计算模型 - 实际应接入电价数据和历史发电数据
	electricRevenue := totalCleanArea * 0.5   // 假设: 每㎡清扫增收0.5元电费
	laborReplaceRate := 100.0                  // 假设: 全部由机器人清扫
	unitCost := 0.1                            // 假设: 单位成本0.1元/㎡
	annualNetIncome := electricRevenue - totalCleanArea*unitCost
	investPayback := 0.0
	if annualNetIncome > 0 {
		investPayback = 1000000 / annualNetIncome // 假设: 设备投资100万
	}

	coverage := 0.0
	if station.PanelArea > 0 {
		coverage = totalCleanArea / station.PanelArea * 100
	}

	return &EconomicReport{
		StationID:        station.StationID,
		StationName:      station.StationName,
		TotalCleanArea:   totalCleanArea,
		ElectricRevenue:  electricRevenue,
		LaborReplaceRate: laborReplaceRate,
		UnitCost:         unitCost,
		InvestPayback:    investPayback,
		Coverage:         coverage,
	}, nil
}

// GetEfficiencyReport 生成指定电站的效率分析报表。
// 包含清扫效率、覆盖率、设备在线率等维度。
func (s *AnalyticsService) GetEfficiencyReport(stationID string) (*EfficiencyReport, error) {
	station, err := s.stationRepo.GetByID(stationID)
	if err != nil {
		return nil, err
	}

	robots, err := s.robotRepo.GetByStationID(stationID)
	if err != nil {
		return nil, err
	}

	// 统计在线率
	totalRobots := len(robots)
	onlineCount := 0
	for _, r := range robots {
		if r.OnlineStatus == 1 {
			onlineCount++
		}
	}
	onlineRate := 0.0
	if totalRobots > 0 {
		onlineRate = float64(onlineCount) / float64(totalRobots) * 100
	}

	cleanEfficiency := 2000.0 // 假设清扫效率 2000㎡/h (固定式)

	return &EfficiencyReport{
		StationID:       station.StationID,
		StationName:     station.StationName,
		CleanEfficiency: cleanEfficiency,
		CoverageRate:    95.0,  // TODO: 从实际清扫轨迹计算
		RepeatRate:      5.0,   // TODO: 从实际清扫轨迹计算
		OnlineRate:      onlineRate,
		UtilizationRate: 80.0,  // TODO: 从任务执行时长计算
	}, nil
}

// GetRunReport 生成指定电站的运行报表，汇总清扫面积、任务数、机器人状态等。
func (s *AnalyticsService) GetRunReport(stationID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	station, err := s.stationRepo.GetByID(stationID)
	if err != nil {
		return nil, err
	}

	totalCleanArea, _ := s.taskRepo.SumCleanAreaByStation(stationID)
	taskCount, _ := s.taskRepo.CountByStation(stationID)
	robotCount, _ := s.robotRepo.CountByStation(stationID)
	onlineCount, _ := s.robotRepo.CountOnlineByStation(stationID)

	return map[string]interface{}{
		"station_id":       station.StationID,
		"station_name":     station.StationName,
		"total_clean_area": totalCleanArea,
		"task_count":       taskCount,
		"robot_count":      robotCount,
		"online_count":     onlineCount,
		"period_start":     startDate,
		"period_end":       endDate,
	}, nil
}

// Compare 多维度对比分析，支持按电站、机器人、时段对比。
func (s *AnalyticsService) Compare(dimension string, ids []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	result["dimension"] = dimension
	var items []map[string]interface{}
	for _, id := range ids {
		item := map[string]interface{}{"id": id}
		cleanArea, _ := s.taskRepo.SumCleanAreaByStation(id)
		taskCount, _ := s.taskRepo.CountByStation(id)
		robotCount, _ := s.robotRepo.CountByStation(id)
		onlineCount, _ := s.robotRepo.CountOnlineByStation(id)
		item["clean_area"] = cleanArea
		item["task_count"] = taskCount
		item["robot_count"] = robotCount
		item["online_count"] = onlineCount
		items = append(items, item)
	}
	result["items"] = items
	return result, nil
}

// ExportReport 导出运行报表为 CSV 二维数组。
func (s *AnalyticsService) ExportReport(stationID string, startDate, endDate time.Time) ([][]string, error) {
	report, err := s.GetRunReport(stationID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	var rows [][]string
	rows = append(rows, []string{"指标", "值"})
	for k, v := range report {
		rows = append(rows, []string{k, fmt.Sprintf("%v", v)})
	}
	return rows, nil
}

// GetTimeSeries 按时间粒度聚合清扫面积时序数据。
func (s *AnalyticsService) GetTimeSeries(stationID, granularity, startTime, endTime string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	records, err := s.cleanRepo.GetByStation(stationID, startTime, endTime)
	if err != nil {
		return results, nil
	}
	buckets := make(map[string]float64)
	for _, rec := range records {
		var key string
		switch granularity {
		case "hour":
			key = rec.RecordTime.Format("2006-01-02 15:00")
		case "week":
			_, week := rec.RecordTime.ISOWeek()
			key = fmt.Sprintf("%d-W%02d", rec.RecordTime.Year(), week)
		case "month":
			key = rec.RecordTime.Format("2006-01")
		case "year":
			key = rec.RecordTime.Format("2006")
		default:
			key = rec.RecordTime.Format("2006-01-02")
		}
		buckets[key] += rec.CleanArea
	}
	for k, v := range buckets {
		results = append(results, map[string]interface{}{"period": k, "value": strconv.FormatFloat(v, 'f', 2, 64)})
	}
	return results, nil
}
