package mqtt

import (
	"encoding/json"
	"log"
	"log/slog"
	"strings"
	"time"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/storage"
	"github.com/mochi-mqtt/server/v2/packets"
	"github.com/mochi-mqtt/server/v2/system"

	"ccplatform/internal/model"
	"ccplatform/internal/repository"
	"ccplatform/internal/ws"
	"ccplatform/pkg/util"
)

// ===== 上行消息结构体（机器人→平台） =====

// HeartbeatPayload 心跳消息，机器人每5秒上报一次。
// 用于判断机器人在线状态和电量。
type HeartbeatPayload struct {
	RobotID      string `json:"robot_id"`      // 机器人 ID
	Timestamp    int64  `json:"timestamp"`     // 机器人本地时间戳（秒）
	BatteryLevel int8   `json:"battery_level"` // 电量百分比
	OnlineStatus int8   `json:"online_status"` // 在线状态: 0离线 1在线
	WorkStatus   int8   `json:"work_status"`   // 工作状态: 0空闲 1清扫 2充电 3故障 4维护
}

// PositionPayload 位置消息，机器人每1秒上报一次。
// 包含三维坐标和朝向，用于 GIS 地图实时展示。
type PositionPayload struct {
	RobotID      string   `json:"robot_id"`      // 机器人 ID
	Timestamp    int64    `json:"timestamp"`     // 时间戳
	PosX         float64  `json:"pos_x"`         // X 坐标 (m)
	PosY         float64  `json:"pos_y"`         // Y 坐标 (m)
	PosZ         float64  `json:"pos_z"`         // Z 坐标 (m)
	Heading      float64  `json:"heading"`       // 朝向角度 (度)
	GPSLongitude *float64 `json:"gps_longitude"` // GPS 经度
	GPSLatitude  *float64 `json:"gps_latitude"`  // GPS 纬度
	GPSAltitude  *float64 `json:"gps_altitude"`  // GPS 高度
	GPSAccuracy  *float64 `json:"gps_accuracy"`  // GPS 精度
}

// StatusPayload 运行状态消息，机器人每10秒上报一次。
// 包含运行参数和环境传感器数据。
type StatusPayload struct {
	RobotID                 string   `json:"robot_id"`                   // 机器人 ID
	Timestamp               int64    `json:"timestamp"`                  // 时间戳
	WorkStatus              int8     `json:"work_status"`                // 工作状态
	Speed                   float64  `json:"speed"`                      // 速度 (m/s)
	CleanArea               float64  `json:"clean_area"`                 // 累计清扫面积 (㎡)
	FaultCode               int      `json:"fault_code"`                 // 故障码
	Temperature             float64  `json:"temperature"`                // 温度 (℃)
	Humidity                float64  `json:"humidity"`                   // 湿度 (%)
	LightIntensity          float64  `json:"light_intensity"`            // 光照强度 (lux)
	WindSpeed               float64  `json:"wind_speed"`                 // 风速 (m/s)
	BatteryVoltage          *float64 `json:"battery_voltage"`            // 主电池电压
	FaultStatus             *int8    `json:"fault_status"`               // 故障状态
	WorkPeriod              *int8    `json:"work_period"`                // 工作时段
	SignalStrength          *int     `json:"signal_strength"`            // 综合信号强度
	Signal4GStrength        *int     `json:"signal_4g_strength"`         // 4G 信号强度
	RunDurationSeconds      *int64   `json:"run_duration_seconds"`       // 运行时长
	CartBatteryVoltage      *float64 `json:"cart_battery_voltage"`       // 清扫小车电池电压
	CartLowVoltageStatus    *int8    `json:"cart_low_voltage_status"`    // 清扫小车低压状态
	ShuttleBatteryVoltage   *float64 `json:"shuttle_battery_voltage"`    // 接驳车电池电压
	ShuttleLowVoltageStatus *int8    `json:"shuttle_low_voltage_status"` // 接驳车低压状态
	ShuttleMotorCurrent     *float64 `json:"shuttle_motor_current"`      // 接驳车电机电流
	CartCleanMotorCurrent   *float64 `json:"cart_clean_motor_current"`   // 清扫电机电流
	CartTravelMotorCurrent  *float64 `json:"cart_travel_motor_current"`  // 行进电机电流
	ShuttlePowerPercent     *int8    `json:"shuttle_power_percent"`      // 接驳车功率选择
	ShuttleStartPosition    *int8    `json:"shuttle_start_position"`     // 接驳车起点位置
	ShuttleEndPosition      *int8    `json:"shuttle_end_position"`       // 接驳车终点位置
	ShuttleHasCleaner       *int8    `json:"shuttle_has_cleaner"`        // 接驳车上是否有清扫车
	DeviceTime              *int64   `json:"device_time"`                // 设备当前时间戳
}

// AlarmPayload 告警消息，机器人故障时实时上报。
type AlarmPayload struct {
	RobotID      string `json:"robot_id"`      // 机器人 ID
	Timestamp    int64  `json:"timestamp"`     // 时间戳
	AlarmLevel   int8   `json:"alarm_level"`   // 告警级别: 1提示 2一般 3严重 4紧急
	AlarmType    string `json:"alarm_type"`    // 告警类型编码
	AlarmContent string `json:"alarm_content"` // 告警描述
}

// CleanPayload 清扫数据消息，机器人清扫时实时上报。
type CleanPayload struct {
	RobotID   string  `json:"robot_id"`   // 机器人 ID
	TaskID    uint    `json:"task_id"`    // 关联任务 ID
	Timestamp int64   `json:"timestamp"`  // 时间戳
	PosX      float64 `json:"pos_x"`      // 当前位置X
	PosY      float64 `json:"pos_y"`      // 当前位置Y
	CleanArea float64 `json:"clean_area"` // 累计清扫面积(㎡)
}

// OTAPayload OTA 升级状态消息，机器人上报升级进度。
type OTAPayload struct {
	RobotID    string `json:"robot_id"`    // 机器人 ID
	Timestamp  int64  `json:"timestamp"`   // 时间戳
	FirmwareID string `json:"firmware_id"` // 固件 ID
	Status     int8   `json:"status"`      // 升级状态: 1下载中 2安装中 3成功 4失败
	ErrorMsg   string `json:"error_msg"`   // 错误信息
}

// SensorPayload PHM 传感器消息，机器人周期上报关键部件的振动/温度/电气参数，用于故障预测与健康管理。
// VibrationSamples 为可选的原始振动加速度采样序列，若设备侧已完成频谱分析，可直接上报 VibrationRMS/VibrationPeak。
type SensorPayload struct {
	RobotID              string    `json:"robot_id"`               // 机器人 ID
	Timestamp            int64     `json:"timestamp"`              // 时间戳
	Component            string    `json:"component"`              // 部件类型，见 model.Component* 常量
	VibrationSamples     []float64 `json:"vibration_samples"`      // 原始振动加速度采样序列（可选，配合 SampleRateHz 做频谱分析）
	SampleRateHz         float64   `json:"sample_rate_hz"`         // 采样率 (Hz)
	VibrationRMS         *float64  `json:"vibration_rms"`          // 振动有效值，设备侧已计算时可直接上报
	VibrationPeak        *float64  `json:"vibration_peak"`         // 振动峰值，设备侧已计算时可直接上报
	WindingTemp          *float64  `json:"winding_temp"`           // 电机绕组温度 (℃)
	ControllerTemp       *float64  `json:"controller_temp"`        // 控制器散热片温度 (℃)
	InsulationResistMOhm *float64  `json:"insulation_resist_mohm"` // 绝缘电阻 (MΩ)
}

// MessageHandler 是 MQTT 消息处理器，实现了 mochi-mqtt 的 Hook 接口。
// 通过 OnPublish 拦截所有发布消息，根据 Topic 路径分发到对应的处理函数。
// 处理流程: 解析JSON → 更新数据库 → 通过 WebSocket 推送给前端。
//
// TODO: MQTT 客户端认证增强 — 替换 AllowHook 为基于 DeviceCredential 的证书/Token 认证 (需求 5.2.1)
type MessageHandler struct {
	mqtt.HookBase                                  // 嵌入 HookBase 提供默认实现
	robotRepo      *repository.RobotRepo           // 机器人数据操作
	alarmRepo      *repository.AlarmRepo           // 告警数据操作
	posHistoryRepo *repository.RobotPositionRepo   // 位置历史
	envRepo        *repository.EnvironmentDataRepo // 环境数据
	cleanRepo      *repository.CleaningRecordRepo  // 清扫记录
	taskRepo       *repository.TaskRepo            // 任务数据操作（clean 面积更新、work_status 联动）
	sensorRepo     *repository.SensorDataRepo      // PHM 传感器数据（振动/温度/电气参数）
	influx         *repository.InfluxWriter
	wsHub          *ws.Hub // WebSocket 广播器
}

// NewMessageHandler 创建消息处理器实例。
func NewMessageHandler(wsHub *ws.Hub, influx *repository.InfluxWriter) *MessageHandler {
	return &MessageHandler{
		robotRepo:      repository.NewRobotRepo(),
		alarmRepo:      repository.NewAlarmRepo(),
		posHistoryRepo: repository.NewRobotPositionRepo(),
		envRepo:        repository.NewEnvironmentDataRepo(),
		cleanRepo:      repository.NewCleaningRecordRepo(),
		taskRepo:       repository.NewTaskRepo(),
		sensorRepo:     repository.NewSensorDataRepo(),
		influx:         influx,
		wsHub:          wsHub,
	}
}

// RegisterHooks 将本处理器注册为 MQTT Broker 的 Hook。
func (h *MessageHandler) RegisterHooks(server *mqtt.Server) error {
	return server.AddHook(h, nil)
}

// ID 返回 Hook 唯一标识。
func (h *MessageHandler) ID() string {
	return "tdw-message-handler"
}

// Provides 声明本 Hook 关注的事件类型：消息发布、连接、断开。
func (h *MessageHandler) Provides(b byte) bool {
	return b == mqtt.OnPublish || b == mqtt.OnConnect || b == mqtt.OnDisconnect
}

// Init Hook 初始化（无特殊配置）。
func (h *MessageHandler) Init(config any) error {
	return nil
}

// Stop Hook 停止（无特殊清理）。
func (h *MessageHandler) Stop() error {
	return nil
}

// SetOpts 设置 Hook 选项（由 Broker 调用）。
func (h *MessageHandler) SetOpts(l *slog.Logger, o *mqtt.HookOptions) {}

// OnConnect 客户端连接事件，记录日志。
func (h *MessageHandler) OnConnect(cl *mqtt.Client, pk packets.Packet) error {
	log.Printf("[MQTT] Client connected: %s", cl.ID)
	return nil
}

// OnDisconnect 客户端断开事件，记录日志。
func (h *MessageHandler) OnDisconnect(cl *mqtt.Client, err error, expire bool) {
	log.Printf("[MQTT] Client disconnected: %s", cl.ID)
}

// OnPublish 核心方法：拦截所有发布消息，解析 Topic 并分发处理。
// Topic 格式: tdw/robot/{robotID}/{消息类型}
// 消息类型: heartbeat(心跳), position(位置), status(状态), alarm(告警), clean(清扫), ota(升级), sensor(PHM传感器)
func (h *MessageHandler) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	topic := pk.TopicName
	payload := pk.Payload

	// 解析 Topic 路径，格式必须是 tdw/robot/{id}/{type}
	parts := strings.Split(topic, "/")
	if len(parts) != 4 || parts[0] != "tdw" || parts[1] != "robot" {
		return pk, nil // 非机器人消息，忽略
	}

	robotID := parts[2]
	msgType := parts[3]

	// 根据消息类型分发到对应处理函数
	switch msgType {
	case "heartbeat":
		h.handleHeartbeat(robotID, payload)
	case "position":
		h.handlePosition(robotID, payload)
	case "status":
		h.handleStatus(robotID, payload)
	case "alarm":
		h.handleAlarm(robotID, payload)
	case "clean":
		h.handleClean(robotID, payload)
	case "ota":
		h.handleOTA(robotID, payload)
	case "sensor":
		h.handleSensor(robotID, payload)
	}

	return pk, nil
}

// Stub implementations for required interface methods
func (h *MessageHandler) OnStarted()                                              {}
func (h *MessageHandler) OnStopped()                                              {}
func (h *MessageHandler) OnSysInfoTick(*system.Info)                              {}
func (h *MessageHandler) OnSessionEstablish(cl *mqtt.Client, pk packets.Packet)   {}
func (h *MessageHandler) OnSessionEstablished(cl *mqtt.Client, pk packets.Packet) {}
func (h *MessageHandler) OnAuthPacket(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	return pk, nil
}
func (h *MessageHandler) OnPacketRead(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	return pk, nil
}
func (h *MessageHandler) OnPacketEncode(cl *mqtt.Client, pk packets.Packet) packets.Packet {
	return pk
}
func (h *MessageHandler) OnPacketSent(cl *mqtt.Client, pk packets.Packet, b []byte)       {}
func (h *MessageHandler) OnPacketProcessed(cl *mqtt.Client, pk packets.Packet, err error) {}
func (h *MessageHandler) OnSubscribe(cl *mqtt.Client, pk packets.Packet) packets.Packet {
	return pk
}
func (h *MessageHandler) OnSubscribed(cl *mqtt.Client, pk packets.Packet, reasonCodes []byte) {}
func (h *MessageHandler) OnSelectSubscribers(subs *mqtt.Subscribers, pk packets.Packet) *mqtt.Subscribers {
	return subs
}
func (h *MessageHandler) OnUnsubscribe(cl *mqtt.Client, pk packets.Packet) packets.Packet {
	return pk
}
func (h *MessageHandler) OnUnsubscribed(cl *mqtt.Client, pk packets.Packet)                        {}
func (h *MessageHandler) OnPublished(cl *mqtt.Client, pk packets.Packet)                           {}
func (h *MessageHandler) OnPublishDropped(cl *mqtt.Client, pk packets.Packet)                      {}
func (h *MessageHandler) OnRetainMessage(cl *mqtt.Client, pk packets.Packet, r int64)              {}
func (h *MessageHandler) OnRetainPublished(cl *mqtt.Client, pk packets.Packet)                     {}
func (h *MessageHandler) OnQosPublish(cl *mqtt.Client, pk packets.Packet, sent int64, resends int) {}
func (h *MessageHandler) OnQosComplete(cl *mqtt.Client, pk packets.Packet)                         {}
func (h *MessageHandler) OnQosDropped(cl *mqtt.Client, pk packets.Packet)                          {}
func (h *MessageHandler) OnPacketIDExhausted(cl *mqtt.Client, pk packets.Packet)                   {}
func (h *MessageHandler) OnWill(cl *mqtt.Client, will mqtt.Will) (mqtt.Will, error) {
	return will, nil
}
func (h *MessageHandler) OnWillSent(cl *mqtt.Client, pk packets.Packet) {}
func (h *MessageHandler) OnClientExpired(cl *mqtt.Client)               {}
func (h *MessageHandler) OnRetainedExpired(filter string)               {}
func (h *MessageHandler) OnConnectAuthenticate(cl *mqtt.Client, pk packets.Packet) bool {
	return true
}
func (h *MessageHandler) OnACLCheck(cl *mqtt.Client, topic string, write bool) bool {
	return true
}
func (h *MessageHandler) StoredClients() ([]storage.Client, error) {
	return nil, nil
}
func (h *MessageHandler) StoredSubscriptions() ([]storage.Subscription, error) {
	return nil, nil
}
func (h *MessageHandler) StoredInflightMessages() ([]storage.Message, error) {
	return nil, nil
}
func (h *MessageHandler) StoredRetainedMessages() ([]storage.Message, error) {
	return nil, nil
}
func (h *MessageHandler) StoredSysInfo() (storage.SystemInfo, error) {
	return storage.SystemInfo{}, nil
}

func addIfPresent[T any](target map[string]interface{}, key string, value *T) {
	if value != nil {
		target[key] = *value
	}
}

func (h *MessageHandler) handleHeartbeat(robotID string, payload []byte) {
	var msg HeartbeatPayload
	if err := json.Unmarshal(payload, &msg); err != nil {
		log.Printf("[MQTT] Parse heartbeat error: %v", err)
		return
	}

	// 获取旧 work_status 用于变更检测（机器人首次心跳时跳过）
	var oldWorkStatus int8 = -1
	if robot, err := h.robotRepo.GetByID(robotID); err == nil {
		oldWorkStatus = robot.WorkStatus
	}

	if err := h.robotRepo.UpdateHeartbeat(robotID, msg.OnlineStatus, msg.BatteryLevel, msg.WorkStatus); err != nil {
		log.Printf("[MQTT] Update heartbeat error: %v", err)
		return
	}
	h.writeInflux("robot_heartbeat", robotID, map[string]interface{}{
		"battery_level": msg.BatteryLevel,
		"online_status": msg.OnlineStatus,
		"work_status":   msg.WorkStatus,
	}, time.Unix(msg.Timestamp, 0))

	// work_status 变更时触发任务状态联动
	if oldWorkStatus >= 0 {
		h.onWorkStatusChange(robotID, oldWorkStatus, msg.WorkStatus)
	}

	h.wsHub.BroadcastToAll(ws.Message{
		Type: "heartbeat",
		Data: map[string]interface{}{
			"robot_id":       robotID,
			"online_status":  msg.OnlineStatus,
			"battery_level":  msg.BatteryLevel,
			"work_status":    msg.WorkStatus,
			"last_heartbeat": time.Now(),
		},
	})
}

func (h *MessageHandler) handlePosition(robotID string, payload []byte) {
	var msg PositionPayload
	if err := json.Unmarshal(payload, &msg); err != nil {
		log.Printf("[MQTT] Parse position error: %v", err)
		return
	}

	if err := h.robotRepo.UpdatePosition(robotID, msg.PosX, msg.PosY, msg.PosZ, msg.Heading); err != nil {
		log.Printf("[MQTT] Update position error: %v", err)
		return
	}
	positionExtra := map[string]interface{}{}
	addIfPresent(positionExtra, "gps_longitude", msg.GPSLongitude)
	addIfPresent(positionExtra, "gps_latitude", msg.GPSLatitude)
	addIfPresent(positionExtra, "gps_altitude", msg.GPSAltitude)
	addIfPresent(positionExtra, "gps_accuracy", msg.GPSAccuracy)
	if len(positionExtra) > 0 {
		if err := h.robotRepo.UpdateStatus(robotID, positionExtra); err != nil {
			log.Printf("[MQTT] Update GPS fields error: %v", err)
		}
	}

	// 写入位置历史表，用于轨迹回放
	h.posHistoryRepo.Create(&model.RobotPosition{
		RobotID:   robotID,
		PosX:      msg.PosX,
		PosY:      msg.PosY,
		PosZ:      msg.PosZ,
		Heading:   msg.Heading,
		Timestamp: time.Unix(msg.Timestamp, 0),
	})
	positionInfluxData := map[string]interface{}{
		"pos_x":   msg.PosX,
		"pos_y":   msg.PosY,
		"pos_z":   msg.PosZ,
		"heading": msg.Heading,
	}
	addIfPresent(positionInfluxData, "gps_longitude", msg.GPSLongitude)
	addIfPresent(positionInfluxData, "gps_latitude", msg.GPSLatitude)
	addIfPresent(positionInfluxData, "gps_altitude", msg.GPSAltitude)
	addIfPresent(positionInfluxData, "gps_accuracy", msg.GPSAccuracy)
	h.writeInflux("robot_position", robotID, positionInfluxData, time.Unix(msg.Timestamp, 0))

	positionData := map[string]interface{}{
		"robot_id":  robotID,
		"pos_x":     msg.PosX,
		"pos_y":     msg.PosY,
		"pos_z":     msg.PosZ,
		"heading":   msg.Heading,
		"timestamp": msg.Timestamp,
	}
	addIfPresent(positionData, "gps_longitude", msg.GPSLongitude)
	addIfPresent(positionData, "gps_latitude", msg.GPSLatitude)
	addIfPresent(positionData, "gps_altitude", msg.GPSAltitude)
	addIfPresent(positionData, "gps_accuracy", msg.GPSAccuracy)

	h.wsHub.BroadcastToAll(ws.Message{
		Type: "position",
		Data: positionData,
	})
}

func (h *MessageHandler) handleStatus(robotID string, payload []byte) {
	var msg StatusPayload
	if err := json.Unmarshal(payload, &msg); err != nil {
		log.Printf("[MQTT] Parse status error: %v", err)
		return
	}

	// 获取旧 work_status 用于变更检测
	var oldWorkStatus int8 = -1
	if robot, err := h.robotRepo.GetByID(robotID); err == nil {
		oldWorkStatus = robot.WorkStatus
	}

	updateData := map[string]interface{}{
		"work_status":     msg.WorkStatus,
		"speed":           msg.Speed,
		"clean_area":      msg.CleanArea,
		"fault_code":      msg.FaultCode,
		"temperature":     msg.Temperature,
		"humidity":        msg.Humidity,
		"light_intensity": msg.LightIntensity,
		"wind_speed":      msg.WindSpeed,
	}
	addIfPresent(updateData, "battery_voltage", msg.BatteryVoltage)
	addIfPresent(updateData, "fault_status", msg.FaultStatus)
	addIfPresent(updateData, "work_period", msg.WorkPeriod)
	addIfPresent(updateData, "signal_strength", msg.SignalStrength)
	addIfPresent(updateData, "signal_4g_strength", msg.Signal4GStrength)
	addIfPresent(updateData, "run_duration_seconds", msg.RunDurationSeconds)
	addIfPresent(updateData, "cart_battery_voltage", msg.CartBatteryVoltage)
	addIfPresent(updateData, "cart_low_voltage_status", msg.CartLowVoltageStatus)
	addIfPresent(updateData, "shuttle_battery_voltage", msg.ShuttleBatteryVoltage)
	addIfPresent(updateData, "shuttle_low_voltage_status", msg.ShuttleLowVoltageStatus)
	addIfPresent(updateData, "shuttle_motor_current", msg.ShuttleMotorCurrent)
	addIfPresent(updateData, "cart_clean_motor_current", msg.CartCleanMotorCurrent)
	addIfPresent(updateData, "cart_travel_motor_current", msg.CartTravelMotorCurrent)
	addIfPresent(updateData, "shuttle_power_percent", msg.ShuttlePowerPercent)
	addIfPresent(updateData, "shuttle_start_position", msg.ShuttleStartPosition)
	addIfPresent(updateData, "shuttle_end_position", msg.ShuttleEndPosition)
	addIfPresent(updateData, "shuttle_has_cleaner", msg.ShuttleHasCleaner)
	if msg.DeviceTime != nil {
		updateData["device_time"] = time.Unix(*msg.DeviceTime, 0)
	}
	if err := h.robotRepo.UpdateStatus(robotID, updateData); err != nil {
		log.Printf("[MQTT] Update status error: %v", err)
		return
	}

	// work_status 变更时触发任务状态联动
	if oldWorkStatus >= 0 {
		h.onWorkStatusChange(robotID, oldWorkStatus, msg.WorkStatus)
	}

	// 写入环境数据历史表
	h.envRepo.Create(&model.EnvironmentData{
		RobotID:        robotID,
		Temperature:    msg.Temperature,
		Humidity:       msg.Humidity,
		LightIntensity: msg.LightIntensity,
		WindSpeed:      msg.WindSpeed,
		RecordTime:     time.Unix(msg.Timestamp, 0),
	})
	h.writeInflux("robot_status", robotID, updateData, time.Unix(msg.Timestamp, 0))

	statusData := map[string]interface{}{
		"robot_id":  robotID,
		"timestamp": msg.Timestamp,
	}
	for key, value := range updateData {
		statusData[key] = value
	}
	h.wsHub.BroadcastToAll(ws.Message{
		Type: "status",
		Data: statusData,
	})
}

func (h *MessageHandler) handleAlarm(robotID string, payload []byte) {
	var msg AlarmPayload
	if err := json.Unmarshal(payload, &msg); err != nil {
		log.Printf("[MQTT] Parse alarm error: %v", err)
		return
	}

	robot, err := h.robotRepo.GetByID(robotID)
	if err != nil {
		log.Printf("[MQTT] Get robot for alarm error: %v", err)
		return
	}

	// 规则评估：抑制检查 + 升级检查
	evalResult := h.evaluateAlarmRules(msg.AlarmType, robotID, robot.StationID, msg.AlarmLevel)
	if evalResult != nil && !evalResult.ShouldCreate {
		return
	}
	finalLevel := msg.AlarmLevel
	if evalResult != nil && evalResult.Escalated {
		finalLevel = evalResult.FinalLevel
	}

	alarm := &model.Alarm{
		AlarmID:      util.GenerateID(),
		AlarmLevel:   finalLevel,
		AlarmType:    msg.AlarmType,
		RobotID:      robotID,
		StationID:    robot.StationID,
		AlarmContent: msg.AlarmContent,
		AlarmTime:    time.Unix(msg.Timestamp, 0),
	}

	if err := h.alarmRepo.Create(alarm); err != nil {
		log.Printf("[MQTT] Create alarm error: %v", err)
		return
	}
	h.writeInflux("robot_alarm", robotID, map[string]interface{}{
		"alarm_level":   finalLevel,
		"alarm_type":    msg.AlarmType,
		"alarm_content": msg.AlarmContent,
	}, time.Unix(msg.Timestamp, 0))

	// 通知分发：查找匹配的通知模板
	notifyRepo := repository.NewNotifyTemplateRepo()
	templates, _ := notifyRepo.GetActive()
	for _, tpl := range templates {
		if tpl.AlarmLevel == 0 || tpl.AlarmLevel == msg.AlarmLevel {
			log.Printf("[MQTT] Notify template matched: %s via %s for alarm %s", tpl.TplName, tpl.TplType, alarm.AlarmID)
		}
	}

	h.wsHub.BroadcastToAll(ws.Message{
		Type: "alarm",
		Data: alarm,
	})
}

func (h *MessageHandler) handleClean(robotID string, payload []byte) {
	var msg CleanPayload
	if err := json.Unmarshal(payload, &msg); err != nil {
		log.Printf("[MQTT] Parse clean error: %v", err)
		return
	}

	robot, err := h.robotRepo.GetByID(robotID)
	if err != nil {
		log.Printf("[MQTT] Get robot for clean error: %v", err)
		return
	}

	h.cleanRepo.Create(&model.CleaningRecord{
		TaskID:     msg.TaskID,
		RobotID:    robotID,
		StationID:  robot.StationID,
		PosX:       msg.PosX,
		PosY:       msg.PosY,
		CleanArea:  msg.CleanArea,
		RecordTime: time.Unix(msg.Timestamp, 0),
	})
	h.writeInflux("robot_clean", robotID, map[string]interface{}{
		"task_id":    msg.TaskID,
		"pos_x":      msg.PosX,
		"pos_y":      msg.PosY,
		"clean_area": msg.CleanArea,
	}, time.Unix(msg.Timestamp, 0))

	// 实时更新任务的累计清扫面积
	if msg.TaskID > 0 {
		if err := h.taskRepo.UpdateCleanArea(msg.TaskID, msg.CleanArea); err != nil {
			log.Printf("[MQTT] Update task clean_area error (task=%d): %v", msg.TaskID, err)
		}
	}

	h.wsHub.BroadcastToAll(ws.Message{
		Type: "clean",
		Data: map[string]interface{}{
			"robot_id":   robotID,
			"task_id":    msg.TaskID,
			"clean_area": msg.CleanArea,
			"timestamp":  msg.Timestamp,
		},
	})
}

func (h *MessageHandler) handleOTA(robotID string, payload []byte) {
	var msg OTAPayload
	if err := json.Unmarshal(payload, &msg); err != nil {
		log.Printf("[MQTT] Parse OTA error: %v", err)
		return
	}

	upgradeRepo := repository.NewUpgradeRecordRepo()
	records, _ := upgradeRepo.GetByRobotID(robotID)
	var latest *model.UpgradeRecord
	for i := range records {
		if records[i].Status == 0 || records[i].Status == 1 || records[i].Status == 2 {
			latest = &records[i]
			break
		}
	}
	if latest != nil {
		latest.Status = msg.Status
		latest.ErrorMsg = msg.ErrorMsg
		now := time.Now()
		if msg.Status == 3 || msg.Status == 4 {
			latest.CompleteTime = &now
		}
		upgradeRepo.Update(latest)
	}

	log.Printf("[MQTT] OTA status for robot %s: status=%d", robotID, msg.Status)
	h.writeInflux("robot_ota", robotID, map[string]interface{}{
		"firmware_id": msg.FirmwareID,
		"status":      msg.Status,
		"error_msg":   msg.ErrorMsg,
	}, time.Unix(msg.Timestamp, 0))
}

// handleSensor 处理 PHM 传感器消息：计算振动 RMS/峰值/峰值因子/主频，识别轴承故障特征，
// 写入 robot_sensor_data 表供 PHMService 做健康评分与故障预测，并推送 InfluxDB + WebSocket。
func (h *MessageHandler) handleSensor(robotID string, payload []byte) {
	var msg SensorPayload
	if err := json.Unmarshal(payload, &msg); err != nil {
		log.Printf("[MQTT] Parse sensor error: %v", err)
		return
	}
	if msg.Component == "" {
		log.Printf("[MQTT] Sensor message missing component, robot=%s", robotID)
		return
	}

	rms := 0.0
	peak := 0.0
	dominantFreq := 0.0
	if len(msg.VibrationSamples) > 0 {
		rms = util.ComputeRMS(msg.VibrationSamples)
		peak = util.ComputePeak(msg.VibrationSamples)
		dominantFreq = util.DominantFrequency(msg.VibrationSamples, msg.SampleRateHz)
	}
	if msg.VibrationRMS != nil {
		rms = *msg.VibrationRMS
	}
	if msg.VibrationPeak != nil {
		peak = *msg.VibrationPeak
	}
	crestFactor := util.CrestFactor(peak, rms)
	bearingFlag := util.ClassifyBearingFault(dominantFreq, crestFactor)

	data := &model.RobotSensorData{
		RobotID:          robotID,
		Component:        msg.Component,
		VibrationRMS:     rms,
		VibrationPeak:    peak,
		CrestFactor:      crestFactor,
		DominantFreqHz:   dominantFreq,
		BearingFaultFlag: bearingFlag,
		RecordTime:       time.Unix(msg.Timestamp, 0),
	}
	if msg.WindingTemp != nil {
		data.WindingTemp = *msg.WindingTemp
	}
	if msg.ControllerTemp != nil {
		data.ControllerTemp = *msg.ControllerTemp
	}
	if msg.InsulationResistMOhm != nil {
		data.InsulationResistMOhm = *msg.InsulationResistMOhm
	}

	if err := h.sensorRepo.Create(data); err != nil {
		log.Printf("[MQTT] Create sensor data error: %v", err)
		return
	}

	influxFields := map[string]interface{}{
		"component":          msg.Component,
		"vibration_rms":      rms,
		"vibration_peak":     peak,
		"crest_factor":       crestFactor,
		"dominant_freq_hz":   dominantFreq,
		"bearing_fault_flag": bearingFlag,
	}
	if msg.WindingTemp != nil {
		influxFields["winding_temp"] = *msg.WindingTemp
	}
	if msg.ControllerTemp != nil {
		influxFields["controller_temp"] = *msg.ControllerTemp
	}
	if msg.InsulationResistMOhm != nil {
		influxFields["insulation_resist_mohm"] = *msg.InsulationResistMOhm
	}
	h.writeInflux("robot_sensor", robotID, influxFields, time.Unix(msg.Timestamp, 0))

	h.wsHub.BroadcastToAll(ws.Message{
		Type: "sensor",
		Data: map[string]interface{}{
			"robot_id":           robotID,
			"component":          msg.Component,
			"vibration_rms":      rms,
			"vibration_peak":     peak,
			"crest_factor":       crestFactor,
			"dominant_freq_hz":   dominantFreq,
			"bearing_fault_flag": bearingFlag,
			"timestamp":          msg.Timestamp,
		},
	})
}

func (h *MessageHandler) writeInflux(measurement, robotID string, fields map[string]interface{}, ts time.Time) {
	if err := h.influx.WritePoint(measurement, map[string]string{"robot_id": robotID}, fields, ts); err != nil {
		log.Printf("[InfluxDB] Write %s error: %v", measurement, err)
	}
}

// ===== 告警规则评估（内置于 handler，避免 mqtt ↔ service 循环引用）=====

// ===== 机器人 work_status 变更 → 任务状态联动 =====

// onWorkStatusChange 根据机器人 work_status 的变化触发对应任务状态流转。
// handleHeartbeat 和 handleStatus 共用此函数，状态未变化时直接返回。
func (h *MessageHandler) onWorkStatusChange(robotID string, oldStatus, newStatus int8) {
	if oldStatus == newStatus {
		return
	}
	log.Printf("[MQTT] Robot %s work_status changed: %d → %d", robotID, oldStatus, newStatus)
	switch {
	case oldStatus == 1 && newStatus == 0: // 清扫→空闲: 完成任务
		h.completeActiveTask(robotID)
	case oldStatus == 1 && newStatus == 2: // 清扫→充电: 暂停任务
		h.pauseActiveTask(robotID)
	case oldStatus == 1 && newStatus == 3: // 清扫→故障: 失败任务
		h.failActiveTask(robotID)
	case oldStatus == 2 && newStatus == 1: // 充电→清扫: 恢复暂停的任务
		h.resumePausedTask(robotID)
	}
}

func (h *MessageHandler) completeActiveTask(robotID string) {
	task, err := h.taskRepo.GetActiveByRobot(robotID)
	if err != nil {
		log.Printf("[MQTT] completeActiveTask: no active task for robot %s: %v", robotID, err)
		return
	}
	now := time.Now()
	task.ActualEnd = &now
	task.TaskStatus = 2
	if err := h.taskRepo.Update(task); err != nil {
		log.Printf("[MQTT] completeActiveTask: update task %d error: %v", task.TaskID, err)
		return
	}
	log.Printf("[MQTT] Task %d completed by robot %s work_status change", task.TaskID, robotID)
}

func (h *MessageHandler) pauseActiveTask(robotID string) {
	task, err := h.taskRepo.GetActiveByRobot(robotID)
	if err != nil {
		log.Printf("[MQTT] pauseActiveTask: no active task for robot %s: %v", robotID, err)
		return
	}
	if err := h.taskRepo.UpdateStatus(task.TaskID, 3); err != nil {
		log.Printf("[MQTT] pauseActiveTask: update task %d error: %v", task.TaskID, err)
		return
	}
	log.Printf("[MQTT] Task %d paused by robot %s work_status change (low battery)", task.TaskID, robotID)
}

func (h *MessageHandler) failActiveTask(robotID string) {
	task, err := h.taskRepo.GetActiveByRobot(robotID)
	if err != nil {
		log.Printf("[MQTT] failActiveTask: no active task for robot %s: %v", robotID, err)
		return
	}
	if err := h.taskRepo.UpdateStatus(task.TaskID, 5); err != nil {
		log.Printf("[MQTT] failActiveTask: update task %d error: %v", task.TaskID, err)
		return
	}
	log.Printf("[MQTT] Task %d failed by robot %s work_status change (fault)", task.TaskID, robotID)
}

func (h *MessageHandler) resumePausedTask(robotID string) {
	task, err := h.taskRepo.GetActiveByRobot(robotID)
	if err != nil {
		log.Printf("[MQTT] resumePausedTask: no paused task for robot %s: %v", robotID, err)
		return
	}
	if task.TaskStatus != 3 {
		return // 只恢复暂停的任务
	}
	if err := h.taskRepo.UpdateStatus(task.TaskID, 1); err != nil {
		log.Printf("[MQTT] resumePausedTask: update task %d error: %v", task.TaskID, err)
		return
	}
	log.Printf("[MQTT] Task %d resumed by robot %s work_status change (charging→cleaning)", task.TaskID, robotID)
}

// ===== 告警规则评估（内置于 handler，避免 mqtt ↔ service 循环引用）=====

// escalationEntry 升级规则条目：在 within_min 分钟内发生 count 次同类型告警，则将级别升级到 to_level。
type escalationEntry struct {
	Count     int  `json:"count"`
	WithinMin int  `json:"within_min"`
	ToLevel   int8 `json:"to_level"`
}

type suppressionEntry struct {
	WithinSec int `json:"within_sec"`
}

type alarmRuleEvalResult struct {
	ShouldCreate bool
	Escalated    bool
	FinalLevel   int8
}

func (h *MessageHandler) evaluateAlarmRules(alarmType, robotID, stationID string, originalLevel int8) *alarmRuleEvalResult {
	result := &alarmRuleEvalResult{
		ShouldCreate: true,
		FinalLevel:   originalLevel,
	}

	rules, err := repository.NewAlarmRuleRepo().GetByAlarmType(alarmType)
	if err != nil || len(rules) == 0 {
		return result
	}

	rule := pickMatchingRule(rules, robotID, stationID)
	if rule == nil {
		return result
	}

	if rule.SuppressionRule != "" {
		var suppr suppressionEntry
		if err := json.Unmarshal([]byte(rule.SuppressionRule), &suppr); err == nil && suppr.WithinSec > 0 {
			if latest, err := h.alarmRepo.GetLatestAlarmTime(robotID, alarmType); err == nil && time.Since(*latest) < time.Duration(suppr.WithinSec)*time.Second {
				log.Printf("[AlarmRule] Suppressed: %s robot=%s (window=%ds)", alarmType, robotID, suppr.WithinSec)
				result.ShouldCreate = false
				return result
			}
		}
	}

	if rule.EscalationRule != "" {
		var entries []escalationEntry
		if err := json.Unmarshal([]byte(rule.EscalationRule), &entries); err == nil {
			for _, entry := range entries {
				if entry.Count <= 0 || entry.WithinMin <= 0 || entry.ToLevel <= result.FinalLevel {
					continue
				}
				count, err := h.alarmRepo.CountRecentByType(robotID, alarmType, entry.WithinMin)
				if err != nil {
					continue
				}
				if int(count)+1 >= entry.Count {
					result.Escalated = true
					result.FinalLevel = entry.ToLevel
					log.Printf("[AlarmRule] Escalated: %s robot=%s level %d→%d (count=%d, threshold=%d, window=%dmin)",
						alarmType, robotID, originalLevel, result.FinalLevel, count+1, entry.Count, entry.WithinMin)
				}
			}
		}
	}

	return result
}

// pickMatchingRule 按 robot > station > global 优先级选取最匹配的规则。
func pickMatchingRule(rules []model.AlarmRule, robotID, stationID string) *model.AlarmRule {
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
