package service

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
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

	faultPredictionRefreshAfter = 5 * time.Minute
)

var (
	faultPatternSeedOnce sync.Once
	faultPatternSeedErr  error
	predictionRefreshes  sync.Map
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

func (s *PHMService) seedFaultPatternsOnce() error {
	faultPatternSeedOnce.Do(func() {
		faultPatternSeedErr = s.SeedFaultPatterns()
	})
	return faultPatternSeedErr
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

// ===== 2.2 故障模式库种子初始化 =====

// SeedFaultPatterns 幂等补齐内置故障模式库，不覆盖已有数据（支持运维人员后续自行扩充维护知识库）。
// 内置模式覆盖电机异常、电池衰减、传感器漂移、通信中断、控制器、传动系统等 50+ 种故障模式。
func (s *PHMService) SeedFaultPatterns() error {
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
		{PatternCode: "DRV_SPEED_FLUCT", Component: model.ComponentDriveMotor, FaultName: "驱动电机转速波动", Category: "motor",
			SymptomDesc: "转速闭环波动增大，行进速度不稳定", SuggestedAction: "检查编码器反馈、电机驱动参数和轮轨摩擦状态",
			RecommendedParts: "电机编码器,驱动器"},
		{PatternCode: "DRV_ENCODER_FAULT", Component: model.ComponentDriveMotor, FaultName: "驱动编码器异常", Category: "motor",
			SymptomDesc: "速度反馈跳变或丢脉冲", SuggestedAction: "检查编码器线束屏蔽与接插件，必要时更换编码器",
			RecommendedParts: "驱动编码器,编码器线束"},
		{PatternCode: "DRV_BEARING_WEAR", Component: model.ComponentDriveMotor, FaultName: "驱动电机轴承磨损", Category: "motor",
			SymptomDesc: "驱动侧振动 RMS 和峰值因子持续升高", SuggestedAction: "安排轴承听诊和振动复测，必要时更换轴承",
			RecommendedParts: "驱动电机轴承"},
		{PatternCode: "DRV_DRIVER_OVERCURRENT", Component: model.ComponentDriveMotor, FaultName: "驱动器过流", Category: "electrical",
			SymptomDesc: "驱动器输出电流短时冲击频繁", SuggestedAction: "检查负载卡滞和驱动器参数，校验电流采样回路",
			RecommendedParts: "电机驱动器"},
		{PatternCode: "DRV_INSULATION_LOW", Component: model.ComponentDriveMotor, FaultName: "驱动电机绝缘降低", Category: "electrical",
			SymptomDesc: "绝缘电阻低于安全阈值，存在漏电风险", SuggestedAction: "停机进行绝缘测试，检查绕组和动力线缆老化情况",
			RecommendedParts: "驱动电机总成,动力线束"},
		{PatternCode: "DRV_COOLING_FAIL", Component: model.ComponentDriveMotor, FaultName: "驱动电机散热失效", Category: "motor",
			SymptomDesc: "温升斜率异常且负载未显著增加", SuggestedAction: "清理散热片和风道，检查散热风扇供电",
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
		{PatternCode: "BRUSH_ENTANGLEMENT", Component: model.ComponentBrushMotor, FaultName: "滚刷缠绕异物", Category: "motor",
			SymptomDesc: "滚刷电机电流上升并伴随周期性振动", SuggestedAction: "停机清理滚刷异物，检查端部密封件",
			RecommendedParts: "滚刷端盖,密封圈"},
		{PatternCode: "BRUSH_MOTOR_OVERHEAT", Component: model.ComponentBrushMotor, FaultName: "滚刷电机过热", Category: "motor",
			SymptomDesc: "清扫电机绕组温度持续偏高", SuggestedAction: "降低清扫负载，检查滚刷阻力和散热状态",
			RecommendedParts: "滚刷电机,散热风扇"},
		{PatternCode: "BRUSH_BELT_SLIP", Component: model.ComponentBrushMotor, FaultName: "滚刷传动皮带打滑", Category: "transmission",
			SymptomDesc: "电机转速正常但清扫效率下降，振动频谱出现低频波动", SuggestedAction: "张紧或更换滚刷传动皮带",
			RecommendedParts: "传动皮带"},
		{PatternCode: "BRUSH_CURRENT_SENSOR_DRIFT", Component: model.ComponentBrushMotor, FaultName: "滚刷电流采样漂移", Category: "sensor",
			SymptomDesc: "电流采样基线缓慢漂移，与负载变化不一致", SuggestedAction: "校准电流传感器并检查采样线束",
			RecommendedParts: "电流传感器"},
		{PatternCode: "BRUSH_GEARBOX_NOISE", Component: model.ComponentBrushMotor, FaultName: "滚刷减速箱异响", Category: "transmission",
			SymptomDesc: "滚刷侧高频振动增加，疑似减速箱啮合异常", SuggestedAction: "检查减速箱润滑和齿面磨损",
			RecommendedParts: "滚刷减速箱,润滑脂"},
		{PatternCode: "BATT_CAPACITY_FADE", Component: model.ComponentBattery, FaultName: "电池容量衰减", Category: "battery",
			SymptomDesc: "同等电量下电压持续偏低，续航下降", SuggestedAction: "安排电池容量检测，评估是否需要更换电池组",
			RecommendedParts: "锂电池组"},
		{PatternCode: "BATT_INTERNAL_RESISTANCE", Component: model.ComponentBattery, FaultName: "电池内阻升高", Category: "battery",
			SymptomDesc: "放电压降幅度明显大于历史基线", SuggestedAction: "检测电池内阻，检查连接端子氧化情况",
			RecommendedParts: "锂电池组,电池连接端子"},
		{PatternCode: "BATT_CELL_IMBALANCE", Component: model.ComponentBattery, FaultName: "电芯一致性异常", Category: "battery",
			SymptomDesc: "充放电末端压差扩大，单体电压离散度升高", SuggestedAction: "执行均衡维护，复测单体压差",
			RecommendedParts: "电池均衡板"},
		{PatternCode: "BATT_THERMAL_RUNAWAY_RISK", Component: model.ComponentBattery, FaultName: "电池热失控风险", Category: "battery",
			SymptomDesc: "电池温升异常且电压波动加剧", SuggestedAction: "立即停止充放电，隔离电池包并进行热成像检查",
			RecommendedParts: "锂电池组,温度传感器"},
		{PatternCode: "BATT_CHARGE_EFF_LOW", Component: model.ComponentBattery, FaultName: "充电效率下降", Category: "battery",
			SymptomDesc: "充电时长增加且满电后续航不足", SuggestedAction: "检查充电器输出、电池管理系统和接插件接触电阻",
			RecommendedParts: "充电器,BMS"},
		{PatternCode: "BATT_BMS_COMM_FAIL", Component: model.ComponentBattery, FaultName: "BMS 通信异常", Category: "comm",
			SymptomDesc: "BMS 数据间歇丢失或电池状态不可读", SuggestedAction: "检查 BMS 通信线束和终端电阻，升级 BMS 固件",
			RecommendedParts: "BMS,通信线束"},
		{PatternCode: "BATT_LOW_TEMP_CHARGE", Component: model.ComponentBattery, FaultName: "低温充电风险", Category: "battery",
			SymptomDesc: "环境或电芯温度低于建议充电温度", SuggestedAction: "启用预热或延迟充电，避免低温析锂",
			RecommendedParts: "电池加热片"},
		{PatternCode: "BATT_CONNECTOR_OXIDATION", Component: model.ComponentBattery, FaultName: "电池连接器氧化", Category: "electrical",
			SymptomDesc: "负载电流变化时端电压抖动明显", SuggestedAction: "清洁或更换电池连接器，复核端子压接质量",
			RecommendedParts: "电池连接端子"},
		{PatternCode: "CTRL_COMM_LOSS", Component: model.ComponentController, FaultName: "通信中断风险", Category: "comm",
			SymptomDesc: "心跳间隔持续超时，信号强度弱", SuggestedAction: "检查天线连接和现场信号覆盖，必要时增设中继设备",
			RecommendedParts: "4G通信模块"},
		{PatternCode: "CTRL_OVERLOAD", Component: model.ComponentController, FaultName: "控制器处理器过载", Category: "controller",
			SymptomDesc: "程序响应延迟增加，运行不稳定", SuggestedAction: "检查后台任务负载，必要时升级固件或更换主控板",
			RecommendedParts: "主控制器"},
		{PatternCode: "CTRL_OVERHEAT", Component: model.ComponentController, FaultName: "控制器散热异常", Category: "controller",
			SymptomDesc: "散热片温度持续偏高，存在降频风险", SuggestedAction: "清理散热片积尘，检查风扇运行状态",
			RecommendedParts: "控制器散热风扇"},
		{PatternCode: "CTRL_MEMORY_LEAK", Component: model.ComponentController, FaultName: "控制程序内存异常", Category: "controller",
			SymptomDesc: "运行时长增加后响应延迟扩大，疑似内存泄漏", SuggestedAction: "采集运行日志，升级控制器固件并观察复现情况",
			RecommendedParts: "主控制器"},
		{PatternCode: "CTRL_RELAY_STICK", Component: model.ComponentController, FaultName: "控制继电器粘连", Category: "electrical",
			SymptomDesc: "控制输出状态与反馈不一致", SuggestedAction: "停机检测继电器触点，必要时更换控制板",
			RecommendedParts: "控制继电器,主控制器"},
		{PatternCode: "CTRL_POWER_RIPPLE", Component: model.ComponentController, FaultName: "控制电源纹波异常", Category: "electrical",
			SymptomDesc: "控制器供电纹波升高导致偶发复位", SuggestedAction: "检测 DC/DC 输出和接地，检查滤波电容状态",
			RecommendedParts: "DC/DC电源模块"},
		{PatternCode: "CTRL_IO_FAILURE", Component: model.ComponentController, FaultName: "控制器 IO 通道异常", Category: "controller",
			SymptomDesc: "输入输出通道响应不稳定或误触发", SuggestedAction: "检查 IO 端子、隔离模块和线束屏蔽",
			RecommendedParts: "IO扩展板"},
		{PatternCode: "CTRL_FIRMWARE_FAULT", Component: model.ComponentController, FaultName: "控制器固件异常", Category: "controller",
			SymptomDesc: "程序运行稳定性下降，日志出现重复异常码", SuggestedAction: "回滚或升级固件，保留异常日志用于分析",
			RecommendedParts: "主控制器"},
		{PatternCode: "CTRL_CLOCK_DRIFT", Component: model.ComponentController, FaultName: "控制器时钟漂移", Category: "controller",
			SymptomDesc: "设备时间与平台时间偏差持续扩大", SuggestedAction: "检查 RTC 电池和 NTP 同步配置",
			RecommendedParts: "RTC电池"},
		{PatternCode: "SENS_GPS_DRIFT", Component: model.ComponentSensor, FaultName: "定位精度漂移", Category: "sensor",
			SymptomDesc: "GPS/RTK 定位精度持续劣化", SuggestedAction: "检查天线安装和遮挡情况，重新校准定位模块",
			RecommendedParts: "GPS/RTK定位模块"},
		{PatternCode: "SENS_DATA_INCONSISTENCY", Component: model.ComponentSensor, FaultName: "传感器数据不一致", Category: "sensor",
			SymptomDesc: "多源定位数据交叉校验偏差增大", SuggestedAction: "校准传感器融合参数，检查传感器连接线路",
			RecommendedParts: "传感器信号线束"},
		{PatternCode: "SENS_IMU_BIAS", Component: model.ComponentSensor, FaultName: "IMU 零偏漂移", Category: "sensor",
			SymptomDesc: "姿态估计长期偏移，静止状态下角速度非零", SuggestedAction: "执行 IMU 静态标定并检查安装松动",
			RecommendedParts: "IMU模块"},
		{PatternCode: "SENS_LIDAR_DIRTY", Component: model.ComponentSensor, FaultName: "激光雷达污染", Category: "sensor",
			SymptomDesc: "点云有效点比例下降，测距噪声增加", SuggestedAction: "清洁雷达视窗，检查防尘罩和加热除雾功能",
			RecommendedParts: "雷达防尘罩"},
		{PatternCode: "SENS_ULTRASONIC_FAIL", Component: model.ComponentSensor, FaultName: "超声传感器失效", Category: "sensor",
			SymptomDesc: "近距离避障回波缺失或距离固定", SuggestedAction: "检查超声探头表面和线束，替换异常探头",
			RecommendedParts: "超声传感器"},
		{PatternCode: "SENS_CAMERA_OCCLUDED", Component: model.ComponentSensor, FaultName: "视觉传感器遮挡", Category: "sensor",
			SymptomDesc: "图像亮度或特征点数量异常下降", SuggestedAction: "清洁镜头，检查防护罩结露和安装角度",
			RecommendedParts: "摄像头防护罩"},
		{PatternCode: "SENS_TEMP_DRIFT", Component: model.ComponentSensor, FaultName: "温度传感器漂移", Category: "sensor",
			SymptomDesc: "温度读数与环境或邻近测点偏差扩大", SuggestedAction: "对比校准温度传感器，检查热耦合位置",
			RecommendedParts: "温度传感器"},
		{PatternCode: "SENS_WATER_INGRESS", Component: model.ComponentSensor, FaultName: "传感器进水风险", Category: "sensor",
			SymptomDesc: "雨后传感器数据跳变或通信错误增多", SuggestedAction: "检查密封圈和接插件防水，烘干后复测",
			RecommendedParts: "防水接插件,密封圈"},
		{PatternCode: "SENS_SYNC_LOSS", Component: model.ComponentSensor, FaultName: "多传感器时间同步异常", Category: "sensor",
			SymptomDesc: "多源传感器时间戳偏差导致融合结果抖动", SuggestedAction: "检查时间同步服务和总线延迟",
			RecommendedParts: "同步模块"},
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
		{PatternCode: "TRANS_CHAIN_SLACK", Component: model.ComponentTransmission, FaultName: "传动链条松弛", Category: "transmission",
			SymptomDesc: "启停时冲击增大，传动效率下降", SuggestedAction: "调整链条张紧度并检查链轮磨损",
			RecommendedParts: "传动链条,链轮"},
		{PatternCode: "TRANS_LUBE_LOW", Component: model.ComponentTransmission, FaultName: "传动润滑不足", Category: "transmission",
			SymptomDesc: "温升和振动同步升高，齿轮啮合噪声增大", SuggestedAction: "补充或更换润滑脂，检查密封泄漏",
			RecommendedParts: "润滑脂,油封"},
		{PatternCode: "TRANS_SHAFT_MISALIGN", Component: model.ComponentTransmission, FaultName: "传动轴对中不良", Category: "transmission",
			SymptomDesc: "低频振动增强且随转速变化明显", SuggestedAction: "重新校准轴系对中，检查安装基座松动",
			RecommendedParts: "联轴器"},
		{PatternCode: "TRANS_COUPLING_WEAR", Component: model.ComponentTransmission, FaultName: "联轴器磨损", Category: "transmission",
			SymptomDesc: "扭矩传递波动，启停冲击明显", SuggestedAction: "检查联轴器间隙和弹性体老化情况",
			RecommendedParts: "联轴器弹性体"},
		{PatternCode: "TRANS_BRAKE_DRAG", Component: model.ComponentTransmission, FaultName: "制动拖滞", Category: "transmission",
			SymptomDesc: "空载电流升高且传动部件温升异常", SuggestedAction: "检查制动器释放间隙和回位弹簧",
			RecommendedParts: "制动器总成"},
		{PatternCode: "TRANS_SEAL_DAMAGE", Component: model.ComponentTransmission, FaultName: "传动密封损坏", Category: "transmission",
			SymptomDesc: "润滑脂泄漏或粉尘进入导致磨损加快", SuggestedAction: "更换密封件并补充润滑脂",
			RecommendedParts: "油封,密封圈"},
		{PatternCode: "TRANS_TRACK_JAM", Component: model.ComponentTransmission, FaultName: "轨道/行走机构卡滞", Category: "transmission",
			SymptomDesc: "行走阻力升高，电流和振动同时异常", SuggestedAction: "清理轨道异物，检查导轮和限位机构",
			RecommendedParts: "导轮,限位轮"},
	}
	for i := range seeds {
		if err := s.patternRepo.CreateIfNotExists(&seeds[i]); err != nil {
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
	if r.WorkStatus == 3 || r.FaultStatus == 1 || r.FaultCode != 0 {
		penalty += 35
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
	case r.WorkStatus == 3 || r.FaultStatus == 1 || r.FaultCode != 0:
		pattern = "DRV_STALL"
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
	if r.WorkStatus == 3 || r.FaultStatus == 1 || r.FaultCode != 0 {
		penalty += 25
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
	case r.WorkStatus == 3 || r.FaultStatus == 1 || r.FaultCode != 0:
		pattern = "BRUSH_LOAD_IMBALANCE"
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
	if r.BatteryLevel > 0 && r.BatteryLevel < 30 {
		penalty += clamp(float64(30-r.BatteryLevel)*1.5+20, 0, 50)
	} else if r.BatteryLevel >= 30 && r.BatteryLevel < 50 {
		penalty += clamp(float64(50-r.BatteryLevel)*0.8, 0, 20)
	}
	if r.CartLowVoltageStatus == 1 || r.ShuttleLowVoltageStatus == 1 {
		penalty += 30
	}
	score := clamp(100-penalty, 0, 100)

	pattern := ""
	switch {
	case deviation > 4:
		pattern = "BATT_INTERNAL_RESISTANCE"
	case r.CartLowVoltageStatus == 1 || r.ShuttleLowVoltageStatus == 1 || r.BatteryLevel < 30:
		pattern = "BATT_CAPACITY_FADE"
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
	if r.OnlineStatus == 0 {
		penalty += 45
	}
	if ctrlTemp > 70 {
		penalty += clamp((ctrlTemp-70)*2, 0, 30)
	}
	score := clamp(100-penalty, 0, 100)

	pattern := ""
	switch {
	case r.OnlineStatus == 0 || gapSeconds > 60:
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
	if r.FaultStatus == 1 && r.FaultCode != 0 {
		penalty += 20
	}
	score := clamp(100-penalty, 0, 100)

	pattern := ""
	if acc > 2 {
		pattern = "SENS_GPS_DRIFT"
	} else if r.FaultStatus == 1 && r.FaultCode != 0 {
		pattern = "SENS_DATA_INCONSISTENCY"
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
	if r.WorkStatus == 3 || r.FaultStatus == 1 || r.FaultCode != 0 {
		penalty += 25
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
		} else if r.WorkStatus == 3 || r.FaultStatus == 1 || r.FaultCode != 0 {
			pattern = "TRANS_TRACK_JAM"
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
		rate := clamp(degradeRate(history), -20, 20)

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
	PredictionID     string  `json:"prediction_id"`
	RobotID          string  `json:"robot_id"`
	RobotName        string  `json:"robot_name"`
	Component        string  `json:"component"`
	FaultType        string  `json:"fault_type"`
	Probability      float64 `json:"probability"`
	RiskLevel        string  `json:"risk_level"`
	Confidence       float64 `json:"confidence"`
	SuggestedAction  string  `json:"suggested_action"`
	RecommendedParts string  `json:"recommended_parts"`
	PredictedTime    string  `json:"predicted_time"`
	PredictedEnd     string  `json:"predicted_end"`
	Status           int8    `json:"status"`
	MaintID          string  `json:"maint_id"`
}

func toFaultPredictionItem(p model.FaultPrediction, robotName string) faultPredictionItem {
	return faultPredictionItem{
		PredictionID:     p.PredictionID,
		RobotID:          p.RobotID,
		RobotName:        robotName,
		Component:        p.Component,
		FaultType:        p.FaultName,
		Probability:      p.Probability,
		RiskLevel:        p.RiskLevel,
		Confidence:       p.Confidence,
		SuggestedAction:  p.SuggestedAction,
		RecommendedParts: p.RecommendedParts,
		PredictedTime:    p.PredictedStart.Format("2006-01-02 15:04:05"),
		PredictedEnd:     p.PredictedEnd.Format("2006-01-02 15:04:05"),
		Status:           p.Status,
		MaintID:          p.MaintID,
	}
}

// PredictFaults 故障预测查询：只返回已持久化的预测结果，不在 HTTP 请求内同步计算。
// robotID/stationID 均为空时返回全部活跃故障预测，前端可用这一路径展示全局预测列表。
// 若结果为空或过期，会异步触发后台刷新，避免电站级查询等待全站机器人预测计算。
func (s *PHMService) PredictFaults(robotID, stationID string) ([]faultPredictionItem, error) {
	nameByID := map[string]string{}
	refreshRobotIDs := make([]string, 0)
	if robotID != "" {
		robot, err := s.robotRepo.GetByID(robotID)
		if err != nil {
			return nil, err
		}
		nameByID[robotID] = robot.RobotName
		refreshRobotIDs = append(refreshRobotIDs, robotID)
	} else if stationID != "" {
		robots, err := s.robotRepo.GetByStationID(stationID)
		if err != nil {
			return nil, err
		}
		refreshRobotIDs = make([]string, 0, len(robots))
		for _, r := range robots {
			nameByID[r.RobotID] = r.RobotName
			refreshRobotIDs = append(refreshRobotIDs, r.RobotID)
		}
	}

	var predictions []model.FaultPrediction
	var err error

	if robotID != "" {
		predictions, err = s.predictionRepo.GetByRobotID(robotID)
	} else if stationID != "" {
		predictions, err = s.predictionRepo.GetByStationID(stationID)
	} else {
		predictions, err = s.predictionRepo.GetAll()
	}
	if err != nil {
		return nil, err
	}

	predictions = dedupeFaultPredictions(predictions)
	if len(nameByID) == 0 {
		nameByID = s.robotNamesForPredictions(predictions)
	}
	if shouldRefreshFaultPredictions(predictions) {
		refreshKey := "all"
		if robotID != "" {
			refreshKey = "robot:" + robotID
		} else if stationID != "" {
			refreshKey = "station:" + stationID
		}
		s.refreshFaultPredictionsAsync(refreshKey, refreshRobotIDs, stationID)
	}

	items := make([]faultPredictionItem, 0, len(predictions))
	for _, p := range predictions {
		items = append(items, toFaultPredictionItem(p, nameByID[p.RobotID]))
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Probability > items[j].Probability })
	return items, nil
}

func (s *PHMService) robotNamesForPredictions(predictions []model.FaultPrediction) map[string]string {
	ids := make([]string, 0, len(predictions))
	seen := make(map[string]struct{}, len(predictions))
	for _, p := range predictions {
		if _, ok := seen[p.RobotID]; ok {
			continue
		}
		seen[p.RobotID] = struct{}{}
		ids = append(ids, p.RobotID)
	}

	robots, err := s.robotRepo.GetByIDs(ids)
	if err != nil {
		return map[string]string{}
	}
	nameByID := make(map[string]string, len(robots))
	for _, r := range robots {
		nameByID[r.RobotID] = r.RobotName
	}
	return nameByID
}

func shouldRefreshFaultPredictions(predictions []model.FaultPrediction) bool {
	if len(predictions) == 0 {
		return true
	}
	latest := predictions[0].UpdateTime
	for _, p := range predictions[1:] {
		if p.UpdateTime.After(latest) {
			latest = p.UpdateTime
		}
	}
	return time.Since(latest) > faultPredictionRefreshAfter
}

func (s *PHMService) refreshFaultPredictionsAsync(key string, robotIDs []string, stationID string) {
	if _, loaded := predictionRefreshes.LoadOrStore(key, struct{}{}); loaded {
		return
	}
	go func() {
		defer predictionRefreshes.Delete(key)
		if err := s.seedFaultPatternsOnce(); err != nil {
			return
		}
		if len(robotIDs) == 0 {
			var robots []model.Robot
			var err error
			if stationID != "" {
				robots, err = s.robotRepo.GetByStationID(stationID)
			} else {
				robots, err = s.robotRepo.GetAll()
			}
			if err != nil {
				return
			}
			robotIDs = make([]string, 0, len(robots))
			for _, r := range robots {
				robotIDs = append(robotIDs, r.RobotID)
			}
		}
		for _, id := range robotIDs {
			_, _ = s.RunPredictionForRobot(id)
		}
	}()
}

func dedupeFaultPredictions(predictions []model.FaultPrediction) []model.FaultPrediction {
	byKey := make(map[string]model.FaultPrediction, len(predictions))
	for _, p := range predictions {
		key := p.RobotID + "|" + p.Component + "|" + p.PatternCode
		existing, ok := byKey[key]
		if !ok {
			byKey[key] = p
			continue
		}
		if shouldPreferPrediction(p, existing) {
			byKey[key] = p
		}
	}
	result := make([]model.FaultPrediction, 0, len(byKey))
	for _, p := range byKey {
		result = append(result, p)
	}
	return result
}

func shouldPreferPrediction(candidate, existing model.FaultPrediction) bool {
	if existing.MaintID == "" && candidate.MaintID != "" {
		return true
	}
	if existing.MaintID != "" && candidate.MaintID == "" {
		return false
	}
	if candidate.Probability != existing.Probability {
		return candidate.Probability > existing.Probability
	}
	return candidate.UpdateTime.After(existing.UpdateTime)
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
