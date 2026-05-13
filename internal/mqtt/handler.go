package mqtt

import (
	"ccplatform/internal/model"
	"ccplatform/internal/repository"
	"ccplatform/internal/ws"
	"encoding/json"
	"log"
	"log/slog"
	"strings"
	"time"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/storage"
	"github.com/mochi-mqtt/server/v2/packets"
	"github.com/mochi-mqtt/server/v2/system"
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
	RobotID   string  `json:"robot_id"` // 机器人 ID
	Timestamp int64   `json:"timestamp"` // 时间戳
	PosX      float64 `json:"pos_x"`    // X 坐标 (m)
	PosY      float64 `json:"pos_y"`    // Y 坐标 (m)
	PosZ      float64 `json:"pos_z"`    // Z 坐标 (m)
	Heading   float64 `json:"heading"`  // 朝向角度 (度)
}

// StatusPayload 运行状态消息，机器人每10秒上报一次。
// 包含运行参数和环境传感器数据。
type StatusPayload struct {
	RobotID        string  `json:"robot_id"`        // 机器人 ID
	Timestamp      int64   `json:"timestamp"`       // 时间戳
	WorkStatus     int8    `json:"work_status"`     // 工作状态
	Speed          float64 `json:"speed"`           // 速度 (m/s)
	CleanArea      float64 `json:"clean_area"`      // 累计清扫面积 (㎡)
	FaultCode      int     `json:"fault_code"`      // 故障码
	Temperature    float64 `json:"temperature"`     // 温度 (℃)
	Humidity       float64 `json:"humidity"`        // 湿度 (%)
	LightIntensity float64 `json:"light_intensity"` // 光照强度 (lux)
	WindSpeed      float64 `json:"wind_speed"`      // 风速 (m/s)
}

// AlarmPayload 告警消息，机器人故障时实时上报。
type AlarmPayload struct {
	RobotID      string `json:"robot_id"`     // 机器人 ID
	Timestamp    int64  `json:"timestamp"`    // 时间戳
	AlarmLevel   int8   `json:"alarm_level"`  // 告警级别: 1提示 2一般 3严重 4紧急
	AlarmType    string `json:"alarm_type"`   // 告警类型编码
	AlarmContent string `json:"alarm_content"` // 告警描述
}

// MessageHandler 是 MQTT 消息处理器，实现了 mochi-mqtt 的 Hook 接口。
// 通过 OnPublish 拦截所有发布消息，根据 Topic 路径分发到对应的处理函数。
// 处理流程: 解析JSON → 更新数据库 → 通过 WebSocket 推送给前端。
type MessageHandler struct {
	mqtt.HookBase                              // 嵌入 HookBase 提供默认实现
	robotRepo      *repository.RobotRepo       // 机器人数据操作
	alarmRepo      *repository.AlarmRepo       // 告警数据操作
	wsHub          *ws.Hub                     // WebSocket 广播器
}

// NewMessageHandler 创建消息处理器实例。
func NewMessageHandler(wsHub *ws.Hub) *MessageHandler {
	return &MessageHandler{
		robotRepo: repository.NewRobotRepo(),
		alarmRepo: repository.NewAlarmRepo(),
		wsHub:     wsHub,
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
// 消息类型: heartbeat(心跳), position(位置), status(状态), alarm(告警)
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
	}

	return pk, nil
}

// Stub implementations for required interface methods
func (h *MessageHandler) OnStarted()                                    {}
func (h *MessageHandler) OnStopped()                                    {}
func (h *MessageHandler) OnSysInfoTick(*system.Info)                    {}
func (h *MessageHandler) OnSessionEstablish(cl *mqtt.Client, pk packets.Packet)  {}
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
func (h *MessageHandler) OnPacketSent(cl *mqtt.Client, pk packets.Packet, b []byte) {}
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
func (h *MessageHandler) OnUnsubscribed(cl *mqtt.Client, pk packets.Packet)  {}
func (h *MessageHandler) OnPublished(cl *mqtt.Client, pk packets.Packet)     {}
func (h *MessageHandler) OnPublishDropped(cl *mqtt.Client, pk packets.Packet) {}
func (h *MessageHandler) OnRetainMessage(cl *mqtt.Client, pk packets.Packet, r int64) {}
func (h *MessageHandler) OnRetainPublished(cl *mqtt.Client, pk packets.Packet) {}
func (h *MessageHandler) OnQosPublish(cl *mqtt.Client, pk packets.Packet, sent int64, resends int) {}
func (h *MessageHandler) OnQosComplete(cl *mqtt.Client, pk packets.Packet) {}
func (h *MessageHandler) OnQosDropped(cl *mqtt.Client, pk packets.Packet) {}
func (h *MessageHandler) OnPacketIDExhausted(cl *mqtt.Client, pk packets.Packet) {}
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

func (h *MessageHandler) handleHeartbeat(robotID string, payload []byte) {
	var msg HeartbeatPayload
	if err := json.Unmarshal(payload, &msg); err != nil {
		log.Printf("[MQTT] Parse heartbeat error: %v", err)
		return
	}

	if err := h.robotRepo.UpdateHeartbeat(robotID, msg.OnlineStatus, msg.BatteryLevel, msg.WorkStatus); err != nil {
		log.Printf("[MQTT] Update heartbeat error: %v", err)
		return
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

	h.wsHub.BroadcastToAll(ws.Message{
		Type: "position",
		Data: map[string]interface{}{
			"robot_id":  robotID,
			"pos_x":     msg.PosX,
			"pos_y":     msg.PosY,
			"pos_z":     msg.PosZ,
			"heading":   msg.Heading,
			"timestamp": msg.Timestamp,
		},
	})
}

func (h *MessageHandler) handleStatus(robotID string, payload []byte) {
	var msg StatusPayload
	if err := json.Unmarshal(payload, &msg); err != nil {
		log.Printf("[MQTT] Parse status error: %v", err)
		return
	}

	data := map[string]interface{}{
		"work_status":     msg.WorkStatus,
		"speed":           msg.Speed,
		"clean_area":      msg.CleanArea,
		"fault_code":      msg.FaultCode,
		"temperature":     msg.Temperature,
		"humidity":        msg.Humidity,
		"light_intensity": msg.LightIntensity,
		"wind_speed":      msg.WindSpeed,
	}
	if err := h.robotRepo.UpdateStatus(robotID, data); err != nil {
		log.Printf("[MQTT] Update status error: %v", err)
		return
	}

	h.wsHub.BroadcastToAll(ws.Message{
		Type: "status",
		Data: data,
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

	alarm := &model.Alarm{
		AlarmLevel:   msg.AlarmLevel,
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

	h.wsHub.BroadcastToAll(ws.Message{
		Type: "alarm",
		Data: alarm,
	})
}
