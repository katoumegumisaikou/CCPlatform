package handler

import (
	"ccplatform/internal/mqtt"
	"ccplatform/internal/service"
	"ccplatform/pkg/errcode"
	"ccplatform/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RobotHandler 机器人 HTTP 处理器，处理机器人相关的 RESTful API 请求。
// 包含机器人 CRUD、状态查询和远程控制指令下发。
type RobotHandler struct {
	svc       *service.RobotService
	publisher *mqtt.Publisher // MQTT 下行消息发布器，用于发送控制指令
}

// NewRobotHandler 创建 RobotHandler 实例，注入 MQTT Publisher 用于指令下发。
func NewRobotHandler(publisher *mqtt.Publisher) *RobotHandler {
	return &RobotHandler{
		svc:       service.NewRobotService(),
		publisher: publisher,
	}
}

// List 机器人列表接口，支持分页和多维度筛选。
// GET /api/v1/robots?page=1&size=10&station_id=xxx&robot_type=1&online_status=1
func (h *RobotHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	stationID := c.Query("station_id")
	robotType, _ := strconv.Atoi(c.DefaultQuery("robot_type", "0"))
	onlineStatus, _ := strconv.Atoi(c.DefaultQuery("online_status", "-1"))

	robots, total, err := h.svc.List(page, size, stationID, robotType, onlineStatus)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OKPage(c, robots, total, page, size)
}

// GetByID 机器人详情接口。
// GET /api/v1/robots/:id
func (h *RobotHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	robot, err := h.svc.GetByID(id)
	if err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.OK(c, robot)
}

// GetByStation 查询指定电站下的所有机器人。
// GET /api/v1/stations/:id/robots
func (h *RobotHandler) GetByStation(c *gin.Context) {
	stationID := c.Param("id")
	robots, err := h.svc.GetByStationID(stationID)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, robots)
}

// GetStats 机器人状态统计接口，返回总数/在线/离线/故障数量。
// GET /api/v1/robots/stats
func (h *RobotHandler) GetStats(c *gin.Context) {
	stats, err := h.svc.GetStats()
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, stats)
}

// SendCommandRequest 下发控制指令请求体。
type SendCommandRequest struct {
	Cmd    string                 `json:"cmd" binding:"required"` // 指令名称(start/stop/return/reset)
	Params map[string]interface{} `json:"params"`                 // 指令参数
}

// SendCommand 向机器人下发控制指令，通过 MQTT 发送到 tdw/robot/{id}/cmd。
// POST /api/v1/robots/:id/cmd
func (h *RobotHandler) SendCommand(c *gin.Context) {
	id := c.Param("id")
	var req SendCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	if h.publisher == nil {
		response.ErrorMsg(c, 500, 10011, "MQTT publisher not initialized")
		return
	}
	if err := h.publisher.SendCommand(id, req.Cmd, req.Params); err != nil {
		response.Error(c, errcode.ErrRobotOffline)
		return
	}
	response.OK(c, gin.H{"message": "command sent"})
}
