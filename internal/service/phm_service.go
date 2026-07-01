package service

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"ccplatform/internal/model"
	"ccplatform/internal/repository"
	"ccplatform/pkg/util"
)

// 简化健康评分模型的额定基准常量。
// 说明：项目未接入设备铭牌参数库，此处采用行业经验值作为近似基准，
// 用于计算电流/电压相对偏差，属于简化的规则模型（与本项目现有 PredictionService 风格一致），非精确电气仿真。
const (
	nominalDriveMotorCurrent = 5.0  // 驱动电机额定电流 (A)
	nominalBrushMotorCurrent = 3.0  // 滚刷电机额定电流 (A)
	battNominalMinVoltage    = 42.0 // 电池电量 0% 对应的近似电压 (V)
	battNominalMaxVoltage    = 54.0 // 电池电量 100% 对应的近似电压 (V)
)

// defaultPatternByComponent 各部件在健康度低于绿色档但未命中具体特征模式时使用的兜底故障模式编码。
var defaultPatternByComponent = map[string]string{
	model.ComponentDriveMotor:   "DRV_STALL",
	model.ComponentBrushMotor:   "BRUSH_LOAD_IMBALANCE",
	model.ComponentBattery:      "BATT_CAPACITY_FADE",
	model.ComponentController:   "CTRL_OVERLOAD",
	model.ComponentSensor:       "SENS_DATA_INCONSISTENCY",
	model.ComponentTransmission: "TRANS_GEAR_WEAR",
}

var allComponents = []string{
	model.ComponentDriveMotor,
	model.ComponentBrushMotor,
	model.ComponentBattery,
	model.ComponentController,
	model.ComponentSensor,
	model.ComponentTransmission,
}

// PHMService 故障预测与健康管理（PHM）服务，覆盖需求 2.1-2.4：
// 设备健康评估、故障预测预警、传感器监测数据处理、维护策略优化。
type PHMService struct {
	robotRepo       *repository.RobotRepo
	healthRepo      *repository.ComponentHealthRepo
	historyRepo     *repository.HealthHistoryRepo
	sensorRepo      *repository.SensorDataRepo
	patternRepo     *repository.FaultPatternRepo
	predictionRepo  *repository.FaultPredictionRepo
	maintenanceRepo *repository.MaintenanceRepo
}

func NewPHMService() *PHMService {
	return &PHMService{
		robotRepo:       repository.NewRobotRepo(),
		healthRepo:      repository.NewComponentHealthRepo(),
		historyRepo:     repository.NewHealthHistoryRepo(),
		sensorRepo:      repository.NewSensorDataRepo(),
		patternRepo:     repository.NewFaultPatternRepo(),
		predictionRepo:  repository.NewFaultPredictionRepo(),
		maintenanceRepo: repository.NewMaintenanceRepo(),
	}
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func healthLevelOf(score float64) string {
	switch {
	case score >= 80:
		return model.HealthLevelGreen
	case score >= 60:
		return model.HealthLevelYellow
	case score >= 40:
		return model.HealthLevelOrange
	default:
		return model.HealthLevelRed
	}
}

// ===== 2.4 故障模式库种子初始化 =====

// SeedFaultPatterns 幂等初始化内置故障模式库。仅在库为空时写入，不覆盖已有数据（支持运维人员后续自行扩充维护知识库）。
// 内置模式覆盖电机异常、电池衰减、传感器漂移、通信中断、传动系统轴承/齿轮故障等核心分类，可持续扩展至需求所述的50余种。
func (s *PHMService) SeedFaultPatterns() error {
	count, err := s.patternRepo.Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	seeds := []model.FaultPattern{
		{PatternCode: "DRV_STALL", Component: model.ComponentDriveMotor, FaultName: "驱动电机堵转", Category: "motor",
			SymptomDesc: "电机电流显著偏离额定值，转速异常", SuggestedAction: "检查驱动轮及传动机构是否卡滞，测量电机绕组电阻，必要时更换电机",
			RecommendedParts: "驱动电机总成"},
		{PatternCode: "DRV_PHASE_LOSS", Component: model.ComponentDriveMotor, FaultName: "驱动电机缺相", Category: "motor",
			SymptomDesc: "三相电流波形不对称，输出扭矩下降", SuggestedAction: "检查电机接线端子和驱动器输出，紧固或更换接线",
			RecommendedParts: "电机驱动线束"},
		{PatternCode: "DRV_OVERHEAT", Component: model.ComponentDriveMotor, FaultName: "驱动电机过热", Category: "motor",
			SymptomDesc: "绕组温度持续超过阈值", SuggestedAction: "检查散热风道和负载情况，降低连续工作时长，清理积尘",
			RecommendedParts: "电机散热风扇"},
		{PatternCode: "BRUSH_LOAD_IMBALANCE", Component: model.ComponentBrushMotor, FaultName: "滚刷电机负载失衡", Category: "motor",
			SymptomDesc: "清扫电机电流波动明显偏离额定范围", SuggestedAction: "检查滚刷安装是否偏心，清理缠绕异物",
			RecommendedParts: "滚刷总成"},
		{PatternCode: "BRUSH_WEAR", Component: model.ComponentBrushMotor, FaultName: "滚刷刷毛过度磨损", Category: "motor",
			SymptomDesc: "振动有效值持续上升，清扫效率下降", SuggestedAction: "更换滚刷刷毛组件",
			RecommendedParts: "滚刷刷毛"},
		{PatternCode: "BRUSH_BEARING_WEAR", Component: model.ComponentBrushMotor, FaultName: "滚刷电机轴承磨损", Category: "transmission",
			SymptomDesc: "振动频谱出现冲击特征，峰值因子升高", SuggestedAction: "更换滚刷电机轴承",
			RecommendedParts: "滚刷电机轴承"},
		{PatternCode: "BATT_CAPACITY_FADE", Component: model.ComponentBattery, FaultName: "电池容量衰减", Category: "battery",
			SymptomDesc: "同等电量下电压持续偏低，续航下降", SuggestedAction: "安排电池容量检测，评估是否需要更换电池组",
			RecommendedParts: "锂电池组"},
		{PatternCode: "BATT_INTERNAL_RESISTANCE", Component: model.ComponentBattery, FaultName: "电池内阻升高", Category: "battery",
			SymptomDesc: "放电压降幅度明显大于历史基线", SuggestedAction: "检测电池内阻，检查连接端子氧化情况",
			RecommendedParts: "锂电池组,电池连接端子"},
		{PatternCode: "CTRL_COMM_LOSS", Component: model.ComponentController, FaultName: "通信中断风险", Category: "comm",
			SymptomDesc: "心跳间隔持续超时，信号强度弱", SuggestedAction: "检查天线连接和现场信号覆盖，必要时增设中继设备",
			RecommendedParts: "4G通信模块"},
		{PatternCode: "CTRL_OVERLOAD", Component: model.ComponentController, FaultName: "控制器处理器过载", Category: "controller",
			SymptomDesc: "程序响应延迟增加，运行不稳定", SuggestedAction: "检查后台任务负载，必要时升级固件或更换主控板",
			RecommendedParts: "主控制器"},
		{PatternCode: "CTRL_OVERHEAT", Component: model.ComponentController, FaultName: "控制器散热异常", Category: "controller",
			SymptomDesc: "散热片温度持续偏高，存在降频风险", SuggestedAction: "清理散热片积尘，检查风扇运行状态",
			RecommendedParts: "控制器散热风扇"},
		{PatternCode: "SENS_GPS_DRIFT", Component: model.ComponentSensor, FaultName: "定位精度漂移", Category: "sensor",
			SymptomDesc: "GPS/RTK 定位精度持续劣化", SuggestedAction: "检查天线安装和遮挡情况，重新校准定位模块",
			RecommendedParts: "GPS/RTK定位模块"},
		{PatternCode: "SENS_DATA_INCONSISTENCY", Component: model.ComponentSensor, FaultName: "传感器数据不一致", Category: "sensor",
			SymptomDesc: "多源定位数据交叉校验偏差增大", SuggestedAction: "校准传感器融合参数，检查传感器连接线路",
			RecommendedParts: "传感器信号线束"},
		{PatternCode: "TRANS_GEAR_WEAR", Component: model.ComponentTransmission, FaultName: "传动齿轮磨损", Category: "transmission",
			SymptomDesc: "传动效率下降，振动能量缓慢上升", SuggestedAction: "检查齿轮啮合间隙，视磨损程度安排更换",
			RecommendedParts: "传动齿轮组"},
		{PatternCode: "TRANS_BEARING_OUTER", Component: model.ComponentTransmission, FaultName: "轴承外圈故障", Category: "transmission",
			SymptomDesc: "振动频谱出现外圈特征频率成分，峰值因子升高", SuggestedAction: "更换传动轴承，检查轴承座同心度",
			RecommendedParts: "传动轴承"},
		{PatternCode: "TRANS_BEARING_INNER", Component: model.ComponentTransmission, FaultName: "轴承内圈故障", Category: "transmission",
			SymptomDesc: "振动频谱出现内圈特征频率成分", SuggestedAction: "更换传动轴承，检查轴系对中",
			RecommendedParts: "传动轴承"},
		{PatternCode: "TRANS_BEARING_BALL", Component: model.ComponentTransmission, FaultName: "轴承滚珠故障", Category: "transmission",
			SymptomDesc: "振动频谱出现滚珠特征频率成分，高频冲击明显", SuggestedAction: "更换传动轴承，检查润滑状态",
			RecommendedParts: "传动轴承,润滑脂"},
		{PatternCode: "TRANS_INSULATION_AGING", Component: model.ComponentTransmission, FaultName: "绝缘老化漏电风险", Category: "electrical",
			SymptomDesc: "绝缘电阻持续走低，存在漏电风险", SuggestedAction: "安排绝缘测试，检查线缆老化情况，必要时更换线束",
			RecommendedParts: "动力线束"},
	}
	for i := range seeds {
		if err := s.patternRepo.Create(&seeds[i]); err != nil {
			return fmt.Errorf("seed fault pattern %s: %w", seeds[i].PatternCode, err)
		}
	}
	return nil
}

// ===== 2.1 设备健康评估 =====

// componentScore 单个部件的评分结果：健康分数、细分指标快照、命中的故障模式提示（健康时为空）。
type componentScore struct {
	Score       float64
	Metrics     map[string]interface{}
	PatternHint string
}

func driveMotorCurrent(r *model.Robot) float64 {
	if r.RobotType == 2 { // 接驳车
		return r.ShuttleMotorCurrent
	}
	return r.CartTravelMotorCurrent
}

func (s *PHMService) scoreDriveMotor(r *model.Robot) componentScore {
	current := driveMotorCurrent(r)
	penalty := 0.0
	if current > 0 {
		dev := math.Abs(current-nominalDriveMotorCurrent) / nominalDriveMotorCurrent
		penalty += clamp(dev*100*0.6, 0, 40)
	}
	windingTemp := 0.0
	if sensor, err := s.sensorRepo.GetLatest(r.RobotID, model.ComponentDriveMotor); err == nil {
		windingTemp = sensor.WindingTemp
	}
	if windingTemp > 80 {
		penalty += clamp((windingTemp-80)*2, 0, 40)
	}
	score := clamp(100-penalty, 0, 100)

	pattern := ""
	switch {
	case windingTemp > 80:
		pattern = "DRV_OVERHEAT"
	case current == 0 && r.OnlineStatus == 1:
		pattern = "DRV_STALL"
	}
	return componentScore{
		Score: score,
		Metrics: map[string]interface{}{
			"motor_current":   current,
			"nominal_current": nominalDriveMotorCurrent,
			"winding_temp":    windingTemp,
		},
		PatternHint: pattern,
	}
}

func (s *PHMService) scoreBrushMotor(r *model.Robot) componentScore {
	current := r.CartCleanMotorCurrent
	penalty := 0.0
	if current > 0 {
		dev := math.Abs(current-nominalBrushMotorCurrent) / nominalBrushMotorCurrent
		penalty += clamp(dev*100*0.5, 0, 30)
	}
	vibrationRMS := 0.0
	bearingFlag := "none"
	if sensor, err := s.sensorRepo.GetLatest(r.RobotID, model.ComponentBrushMotor); err == nil {
		vibrationRMS = sensor.VibrationRMS
		bearingFlag = sensor.BearingFaultFlag
	}
	penalty += clamp(vibrationRMS*15, 0, 40)
	score := clamp(100-penalty, 0, 100)

	pattern := ""
	switch {
	case bearingFlag != "none":
		pattern = "BRUSH_BEARING_WEAR"
	case vibrationRMS > 3:
		pattern = "BRUSH_WEAR"
	}
	return componentScore{
		Score: score,
		Metrics: map[string]interface{}{
			"motor_current": current,
			"vibration_rms": vibrationRMS,
			"bearing_flag":  bearingFlag,
		},
		PatternHint: pattern,
	}
}

func (s *PHMService) scoreBattery(r *model.Robot) componentScore {
	expected := battNominalMinVoltage + (battNominalMaxVoltage-battNominalMinVoltage)*float64(r.BatteryLevel)/100
	deviation := 0.0
	penalty := 0.0
	if r.BatteryVoltage > 0 {
		deviation = expected - r.BatteryVoltage
		penalty = clamp(deviation*6, 0, 60)
	}
	score := clamp(100-penalty, 0, 100)

	pattern := ""
	switch {
	case deviation > 4:
		pattern = "BATT_INTERNAL_RESISTANCE"
	case score < 60:
		pattern = "BATT_CAPACITY_FADE"
	}
	return componentScore{
		Score: score,
		Metrics: map[string]interface{}{
			"battery_level":     r.BatteryLevel,
			"battery_voltage":   r.BatteryVoltage,
			"expected_voltage":  math.Round(expected*100) / 100,
			"voltage_deviation": math.Round(deviation*100) / 100,
		},
		PatternHint: pattern,
	}
}

func (s *PHMService) scoreController(r *model.Robot) componentScore {
	gapSeconds := 0.0
	if r.LastHeartbeat != nil {
		gapSeconds = time.Since(*r.LastHeartbeat).Seconds()
	}
	ctrlTemp := 0.0
	if sensor, err := s.sensorRepo.GetLatest(r.RobotID, model.ComponentController); err == nil {
		ctrlTemp = sensor.ControllerTemp
	}
	penalty := 0.0
	if r.SignalStrength > 0 && r.SignalStrength < 50 {
		penalty += clamp(float64(50-r.SignalStrength)*0.6, 0, 30)
	}
	if gapSeconds > 30 {
		penalty += clamp((gapSeconds-30)*0.5, 0, 40)
	}
	if ctrlTemp > 70 {
		penalty += clamp((ctrlTemp-70)*2, 0, 30)
	}
	score := clamp(100-penalty, 0, 100)

	pattern := ""
	switch {
	case gapSeconds > 60:
		pattern = "CTRL_COMM_LOSS"
	case ctrlTemp > 70:
		pattern = "CTRL_OVERHEAT"
	}
	return componentScore{
		Score: score,
		Metrics: map[string]interface{}{
			"signal_strength": r.SignalStrength,
			"heartbeat_gap_s": math.Round(gapSeconds*100) / 100,
			"controller_temp": ctrlTemp,
		},
		PatternHint: pattern,
	}
}

func (s *PHMService) scoreSensor(r *model.Robot) componentScore {
	acc := r.GPSAccuracy
	penalty := 0.0
	if acc > 0.5 {
		penalty = clamp((acc-0.5)*40, 0, 60)
	}
	score := clamp(100-penalty, 0, 100)

	pattern := ""
	if acc > 2 {
		pattern = "SENS_GPS_DRIFT"
	}
	return componentScore{
		Score: score,
		Metrics: map[string]interface{}{
			"gps_accuracy": acc,
		},
		PatternHint: pattern,
	}
}

func (s *PHMService) scoreTransmission(r *model.Robot) componentScore {
	rms, crest, insulation := 0.0, 0.0, 0.0
	flag := "none"
	if sensor, err := s.sensorRepo.GetLatest(r.RobotID, model.ComponentTransmission); err == nil {
		rms = sensor.VibrationRMS
		crest = sensor.CrestFactor
		flag = sensor.BearingFaultFlag
		insulation = sensor.InsulationResistMOhm
	}
	penalty := clamp(rms*15, 0, 50)
	if insulation > 0 && insulation < 1 {
		penalty += 30
	}
	score := clamp(100-penalty, 0, 100)

	pattern := ""
	switch flag {
	case "outer_race":
		pattern = "TRANS_BEARING_OUTER"
	case "inner_race":
		pattern = "TRANS_BEARING_INNER"
	case "ball":
		pattern = "TRANS_BEARING_BALL"
	default:
		if insulation > 0 && insulation < 1 {
			pattern = "TRANS_INSULATION_AGING"
		}
	}
	return componentScore{
		Score: score,
		Metrics: map[string]interface{}{
			"vibration_rms":          rms,
			"crest_factor":           crest,
			"bearing_fault_flag":     flag,
			"insulation_resist_mohm": insulation,
		},
		PatternHint: pattern,
	}
}

// scoreAll 对指定机器人计算所有部件的健康评分。
func (s *PHMService) scoreAll(r *model.Robot) map[string]componentScore {
	return map[string]componentScore{
		model.ComponentDriveMotor:   s.scoreDriveMotor(r),
		model.ComponentBrushMotor:   s.scoreBrushMotor(r),
		model.ComponentBattery:      s.scoreBattery(r),
		model.ComponentController:   s.scoreController(r),
		model.ComponentSensor:       s.scoreSensor(r),
		model.ComponentTransmission: s.scoreTransmission(r),
	}
}

// degradeRate 基于历史健康度快照做简单线性回归，估算每日健康度下降速率（分/天）。
// 返回正值表示健康度正在下降，负值/零表示稳定或改善。数据点不足 2 条时返回 0。
func degradeRate(history []model.RobotHealthHistory) float64 {
	n := len(history)
	if n < 2 {
		return 0
	}
	// history 按时间倒序排列，转换为按时间正序的 (天数偏移, 健康分) 序列
	ordered := make([]model.RobotHealthHistory, n)
	for i, h := range history {
		ordered[n-1-i] = h
	}
	t0 := ordered[0].RecordTime
	var sumX, sumY, sumXY, sumX2 float64
	for _, h := range ordered {
		x := h.RecordTime.Sub(t0).Hours() / 24
		y := h.HealthScore
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	denom := float64(n)*sumX2 - sumX*sumX
	if denom == 0 {
		return 0
	}
	slope := (float64(n)*sumXY - sumX*sumY) / denom
	return -slope // slope 为负（health 随时间下降）时，退化速率为正
}

// ComputeHealth 计算并持久化指定机器人所有部件的健康指数：写入当前状态表并追加历史快照。
func (s *PHMService) ComputeHealth(robotID string) ([]model.RobotComponentHealth, error) {
	robot, err := s.robotRepo.GetByID(robotID)
	if err != nil {
		return nil, err
	}

	scores := s.scoreAll(robot)
	now := time.Now()
	result := make([]model.RobotComponentHealth, 0, len(allComponents))

	for _, component := range allComponents {
		cs := scores[component]
		history, _ := s.historyRepo.GetRecent(robotID, component, 30)
		rate := degradeRate(history)

		metricsJSON, _ := json.Marshal(cs.Metrics)
		health := &model.RobotComponentHealth{
			RobotID:     robotID,
			Component:   component,
			HealthScore: math.Round(cs.Score*100) / 100,
			HealthLevel: healthLevelOf(cs.Score),
			DegradeRate: math.Round(rate*1000) / 1000,
			Metrics:     string(metricsJSON),
			UpdateTime:  now,
		}
		if err := s.healthRepo.Upsert(health); err != nil {
			return nil, fmt.Errorf("upsert health %s/%s: %w", robotID, component, err)
		}
		if err := s.historyRepo.Create(&model.RobotHealthHistory{
			RobotID:     robotID,
			Component:   component,
			HealthScore: health.HealthScore,
			RecordTime:  now,
		}); err != nil {
			return nil, fmt.Errorf("append health history %s/%s: %w", robotID, component, err)
		}
		result = append(result, *health)
	}
	return result, nil
}

// GetHealthOverview 返回机器人各部件当前健康状态及综合评分（各部件算术平均）。
func (s *PHMService) GetHealthOverview(robotID string) (map[string]interface{}, error) {
	list, err := s.healthRepo.GetByRobotID(robotID)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		// 尚未计算过健康指数，触发一次即时计算
		if _, err := s.ComputeHealth(robotID); err != nil {
			return nil, err
		}
		list, err = s.healthRepo.GetByRobotID(robotID)
		if err != nil {
			return nil, err
		}
	}

	var total float64
	for _, h := range list {
		total += h.HealthScore
	}
	overall := 0.0
	if len(list) > 0 {
		overall = total / float64(len(list))
	}

	return map[string]interface{}{
		"robot_id":      robotID,
		"overall_score": math.Round(overall*100) / 100,
		"overall_level": healthLevelOf(overall),
		"components":    list,
	}, nil
}

// GetHealthByStation 批量查询电站下所有机器人的健康总览，用于电站维度的健康列表展示。
func (s *PHMService) GetHealthByStation(stationID string) ([]map[string]interface{}, error) {
	robots, err := s.robotRepo.GetByStationID(stationID)
	if err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(robots))
	for _, r := range robots {
		overview, err := s.GetHealthOverview(r.RobotID)
		if err != nil {
			continue
		}
		overview["robot_name"] = r.RobotName
		result = append(result, overview)
	}
	return result, nil
}

// GetHealthTrend 返回指定机器人部件的历史健康趋势曲线，并做环比分析（当前半段窗口 vs 上一半段窗口均值对比）。
func (s *PHMService) GetHealthTrend(robotID, component string, days int) (map[string]interface{}, error) {
	if days <= 0 {
		days = 30
	}
	end := time.Now()
	start := end.AddDate(0, 0, -days)
	history, err := s.historyRepo.GetByRange(robotID, component, start, end)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"robot_id":  robotID,
		"component": component,
		"points":    history,
	}
	if len(history) < 2 {
		result["trend_note"] = "历史数据不足，暂无法进行环比分析"
		return result, nil
	}

	mid := start.Add(end.Sub(start) / 2)
	var prevSum, currSum float64
	var prevCount, currCount int
	for _, h := range history {
		if h.RecordTime.Before(mid) {
			prevSum += h.HealthScore
			prevCount++
		} else {
			currSum += h.HealthScore
			currCount++
		}
	}
	if prevCount > 0 && currCount > 0 {
		prevAvg := prevSum / float64(prevCount)
		currAvg := currSum / float64(currCount)
		result["prev_period_avg"] = math.Round(prevAvg*100) / 100
		result["curr_period_avg"] = math.Round(currAvg*100) / 100
		result["change_rate"] = math.Round((currAvg-prevAvg)/prevAvg*10000) / 100 // 百分比，保留2位小数
	} else {
		result["trend_note"] = "历史数据不足，暂无法进行环比分析"
	}
	return result, nil
}

// ===== 2.2 故障预测预警 =====

// riskLevelOf 按故障概率划分风险等级：高风险 >70%，中风险 40%-70%，低风险 <40%。
func riskLevelOf(probability float64) string {
	switch {
	case probability > 0.7:
		return model.RiskLevelHigh
	case probability >= 0.4:
		return model.RiskLevelMedium
	default:
		return model.RiskLevelLow
	}
}

// RunPredictionForRobot 计算机器人健康指数并对存在退化风险的部件生成/更新故障预测记录。
// 对高风险预测自动生成维护工单。由 PHM 周期检测器和手动查询接口共同调用，保持预测结果实时性。
func (s *PHMService) RunPredictionForRobot(robotID string) ([]model.FaultPrediction, error) {
	robot, err := s.robotRepo.GetByID(robotID)
	if err != nil {
		return nil, err
	}
	healthList, err := s.ComputeHealth(robotID)
	if err != nil {
		return nil, err
	}
	scores := s.scoreAll(robot)

	var predictions []model.FaultPrediction
	for _, health := range healthList {
		if health.HealthScore >= 80 {
			continue // 绿色健康，无需预测
		}
		cs := scores[health.Component]
		patternCode := cs.PatternHint
		if patternCode == "" {
			patternCode = defaultPatternByComponent[health.Component]
		}
		pattern, err := s.patternRepo.GetByCode(patternCode)
		if err != nil {
			continue // 模式库尚未初始化或编码不存在时跳过
		}

		probability := clamp((100-health.HealthScore)/100, 0, 0.99)
		if health.DegradeRate > 0 {
			probability = clamp(probability+math.Min(0.2, health.DegradeRate*0.05), 0, 0.99)
		}
		riskLevel := riskLevelOf(probability)

		history, _ := s.historyRepo.GetRecent(robotID, health.Component, 30)
		confidence := clamp(0.4+0.05*float64(len(history)), 0.4, 0.95)

		daysToCritical := 5.0 // 默认预测窗口居中值（3-7天）
		if health.DegradeRate > 0 {
			daysToCritical = clamp((health.HealthScore-39)/health.DegradeRate, 3, 7)
		}
		if health.HealthScore < 40 {
			daysToCritical = 0 // 已处于危险状态，建议立即处理
		}
		now := time.Now()
		predictedStart := now.AddDate(0, 0, int(math.Round(daysToCritical)))
		predictedEnd := predictedStart.AddDate(0, 0, 2)

		prediction, err := s.predictionRepo.GetOpenByRobotComponent(robotID, health.Component, patternCode)
		isNew := err != nil
		if isNew {
			prediction = &model.FaultPrediction{
				PredictionID: util.GenerateID(),
				RobotID:      robotID,
				StationID:    robot.StationID,
				Component:    health.Component,
				PatternCode:  patternCode,
				Status:       0,
			}
		}
		prediction.FaultName = pattern.FaultName
		prediction.Probability = math.Round(probability*10000) / 10000
		prediction.RiskLevel = riskLevel
		prediction.Confidence = math.Round(confidence*10000) / 10000
		prediction.PredictedStart = predictedStart
		prediction.PredictedEnd = predictedEnd
		prediction.SuggestedAction = pattern.SuggestedAction
		prediction.RecommendedParts = pattern.RecommendedParts

		if isNew {
			if err := s.predictionRepo.Create(prediction); err != nil {
				continue
			}
		} else {
			if err := s.predictionRepo.Update(prediction); err != nil {
				continue
			}
		}

		// 高风险且尚未生成工单时，自动生成预测性维护工单
		if prediction.RiskLevel == model.RiskLevelHigh && prediction.MaintID == "" {
			if _, err := s.generateWorkorderForPrediction(prediction, "系统自动"); err != nil {
				// 工单生成失败不影响预测结果返回，仅记录到 predictions 里维持原状
			}
		}

		predictions = append(predictions, *prediction)
	}
	return predictions, nil
}

// faultPredictionItem 兼容前端既有 FaultPredictionItem 展示结构的响应形状。
type faultPredictionItem struct {
	RobotID         string  `json:"robot_id"`
	RobotName       string  `json:"robot_name"`
	Component       string  `json:"component"`
	FaultType       string  `json:"fault_type"`
	Probability     float64 `json:"probability"`
	RiskLevel       string  `json:"risk_level"`
	Confidence      float64 `json:"confidence"`
	SuggestedAction string  `json:"suggested_action"`
	PredictedTime   string  `json:"predicted_time"`
}

func toFaultPredictionItem(p model.FaultPrediction, robotName string) faultPredictionItem {
	return faultPredictionItem{
		RobotID:         p.RobotID,
		RobotName:       robotName,
		Component:       p.Component,
		FaultType:       p.FaultName,
		Probability:     p.Probability,
		RiskLevel:       p.RiskLevel,
		Confidence:      p.Confidence,
		SuggestedAction: p.SuggestedAction,
		PredictedTime:   p.PredictedStart.Format("2006-01-02 15:04:05"),
	}
}

// PredictFaults 故障预测查询：优先按 robotID 查询单台机器人，否则按 stationID 批量查询电站下所有机器人。
// 每次查询都会重新运行一次预测计算，保证结果基于最新健康数据；返回结构兼容前端既有展示字段。
func (s *PHMService) PredictFaults(robotID, stationID string) ([]faultPredictionItem, error) {
	var robotIDs []string
	nameByID := map[string]string{}

	if robotID != "" {
		robot, err := s.robotRepo.GetByID(robotID)
		if err != nil {
			return nil, err
		}
		robotIDs = []string{robotID}
		nameByID[robotID] = robot.RobotName
	} else if stationID != "" {
		robots, err := s.robotRepo.GetByStationID(stationID)
		if err != nil {
			return nil, err
		}
		for _, r := range robots {
			robotIDs = append(robotIDs, r.RobotID)
			nameByID[r.RobotID] = r.RobotName
		}
	} else {
		return nil, fmt.Errorf("robot_id 或 station_id 必须提供其中之一")
	}

	items := make([]faultPredictionItem, 0)
	for _, id := range robotIDs {
		predictions, err := s.RunPredictionForRobot(id)
		if err != nil {
			continue
		}
		for _, p := range predictions {
			items = append(items, toFaultPredictionItem(p, nameByID[id]))
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Probability > items[j].Probability })
	return items, nil
}

// ===== 2.2(3) 预测性维护建议：自动生成维护工单 =====

// generateWorkorderForPrediction 根据故障预测自动创建维护工单，并回写预测记录的关联工单号和处理状态。
func (s *PHMService) generateWorkorderForPrediction(p *model.FaultPrediction, handler string) (*model.Maintenance, error) {
	maint := &model.Maintenance{
		MaintID:       util.GenerateID(),
		RobotID:       p.RobotID,
		StationID:     p.StationID,
		MaintType:     "predictive",
		Handler:       handler,
		FaultDesc:     fmt.Sprintf("[预测性维护] %s（部件: %s，故障概率 %.0f%%）: %s", p.FaultName, p.Component, p.Probability*100, p.SuggestedAction),
		PartsReplaced: p.RecommendedParts,
		MaintTime:     nil,
		NextMaintTime: &p.PredictedStart,
	}
	if err := s.maintenanceRepo.Create(maint); err != nil {
		return nil, err
	}
	p.Status = 1
	p.MaintID = maint.MaintID
	if err := s.predictionRepo.Update(p); err != nil {
		return nil, err
	}
	return maint, nil
}

// GenerateWorkorder 供接口手动触发：为指定故障预测生成维护工单（若尚未生成）。
func (s *PHMService) GenerateWorkorder(predictionID, handler string) (*model.Maintenance, error) {
	p, err := s.predictionRepo.GetByID(predictionID)
	if err != nil {
		return nil, err
	}
	if p.MaintID != "" {
		return s.maintenanceRepo.GetByID(p.MaintID)
	}
	if handler == "" {
		handler = "人工"
	}
	return s.generateWorkorderForPrediction(p, handler)
}

// ===== 2.4 维护策略优化：备件需求预测 & 维护窗口推荐 =====

// sparePartForecastItem 备件需求预测条目。
type sparePartForecastItem struct {
	PartName     string `json:"part_name"`
	ExpectedQty  int    `json:"expected_qty"`
	RelatedFault int    `json:"related_fault_count"`
}

// ForecastSpareParts 汇总当前未处理(open)的故障预测中推荐更换的备件，按需求数量降序排列。
// 每条预测按 1 单位备件需求计算（简化估算，未接入库存/BOM 系统），可选按电站过滤。
func (s *PHMService) ForecastSpareParts(stationID string) ([]sparePartForecastItem, error) {
	predictions, err := s.predictionRepo.GetOpenByStatus(stationID)
	if err != nil {
		return nil, err
	}
	counts := map[string]int{}
	for _, p := range predictions {
		if p.RecommendedParts == "" {
			continue
		}
		for _, part := range strings.Split(p.RecommendedParts, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			counts[part]++
		}
	}
	items := make([]sparePartForecastItem, 0, len(counts))
	for part, count := range counts {
		items = append(items, sparePartForecastItem{PartName: part, ExpectedQty: count, RelatedFault: count})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ExpectedQty > items[j].ExpectedQty })
	return items, nil
}

// RecommendMaintenanceWindow 推荐机器人下一次维护的最佳时间窗口（简化启发式）。
// 优先选择日出前低发电影响时段，避开日间清扫作业窗口。
// TODO: 结合气象预测和发电计划进一步优化窗口推荐 (需求 2.2.3 / 3.2)，待气象联动模块（三、气象联动功能）完成后接入。
func (s *PHMService) RecommendMaintenanceWindow(robotID string) (map[string]interface{}, error) {
	predictions, err := s.predictionRepo.GetByRobotID(robotID)
	if err != nil {
		return nil, err
	}
	var earliest *model.FaultPrediction
	for i := range predictions {
		if predictions[i].Status != 0 {
			continue
		}
		if earliest == nil || predictions[i].PredictedStart.Before(earliest.PredictedStart) {
			earliest = &predictions[i]
		}
	}

	baseDate := time.Now().AddDate(0, 0, 1)
	reason := "默认建议次日凌晨低发电影响时段"
	if earliest != nil {
		baseDate = earliest.PredictedStart
		reason = fmt.Sprintf("基于故障预测窗口（%s，风险等级: %s）推荐维护时段", earliest.FaultName, earliest.RiskLevel)
	}
	windowStart := time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(), 6, 0, 0, 0, baseDate.Location())
	windowEnd := time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(), 7, 0, 0, 0, baseDate.Location())

	return map[string]interface{}{
		"robot_id":          robotID,
		"recommended_start": windowStart.Format("2006-01-02 15:04:05"),
		"recommended_end":   windowEnd.Format("2006-01-02 15:04:05"),
		"reason":            reason,
	}, nil
}
