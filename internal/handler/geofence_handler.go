package handler

import (
	"ccplatform/internal/model"
	"ccplatform/internal/service"
	"ccplatform/internal/ws"
	"ccplatform/pkg/errcode"
	"ccplatform/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GeofenceHandler 电子围栏 HTTP 处理器。
type GeofenceHandler struct {
	svc *service.GeofenceService
}

// NewGeofenceHandler 创建 GeofenceHandler 实例。
func NewGeofenceHandler(hub *ws.Hub) *GeofenceHandler {
	return &GeofenceHandler{svc: service.NewGeofenceService(hub)}
}

// =========================== 围栏管理 ===========================

type CreateGeofenceRequest struct {
	FenceName   string  `json:"fence_name" binding:"required"`   // 围栏名称
	FenceType   int8    `json:"fence_type" binding:"required"`   // 围栏形状: 1=圆形 2=多边形
	ActionType  int8    `json:"action_type" binding:"required"`  // 动作类型: 1=禁止进入 2=禁止离开
	ScopeType   string  `json:"scope_type"`                      // 作用域: global/station/robot
	ScopeID     string  `json:"scope_id"`                        // 作用域 ID
	CenterLng   float64 `json:"center_lng"`                      // 圆心经度（圆形）
	CenterLat   float64 `json:"center_lat"`                      // 圆心纬度（圆形）
	Radius      float64 `json:"radius"`                          // 半径（圆形，单位米）
	Points      string  `json:"points"`                          // 多边形顶点坐标 JSON
	AlarmLevel  int8    `json:"alarm_level"`                     // 告警级别
	Description string  `json:"description"`                     // 描述
	Color       string  `json:"color"`                           // 地图显示颜色
	Status      int8    `json:"status"`                          // 状态: 0禁用 1启用
}

// List 电子围栏列表（分页）。
// GET /api/v1/geofences?page=1&size=10&keyword=xxx&scope_type=global&fence_type=1
func (h *GeofenceHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	keyword := c.Query("keyword")
	scopeType := c.Query("scope_type")
	fenceType, _ := strconv.Atoi(c.DefaultQuery("fence_type", "0"))

	fences, total, err := h.svc.List(page, size, keyword, scopeType, fenceType)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OKPage(c, fences, total, page, size)
}

// GetAll 查询所有启用围栏（无分页）。
// GET /api/v1/geofences/all
func (h *GeofenceHandler) GetAll(c *gin.Context) {
	fences, err := h.svc.GetAll()
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, fences)
}

// GetByID 围栏详情。
// GET /api/v1/geofences/:id
func (h *GeofenceHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	fence, err := h.svc.GetByID(id)
	if err != nil {
		response.Error(c, errcode.ErrGeofenceNotFound)
		return
	}
	response.OK(c, fence)
}

// Create 创建电子围栏。
// POST /api/v1/geofences
func (h *GeofenceHandler) Create(c *gin.Context) {
	var req CreateGeofenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	fence := &model.Geofence{
		FenceName:   req.FenceName,
		FenceType:   req.FenceType,
		ActionType:  req.ActionType,
		ScopeType:   req.ScopeType,
		ScopeID:     req.ScopeID,
		CenterLng:   req.CenterLng,
		CenterLat:   req.CenterLat,
		Radius:      req.Radius,
		Points:      req.Points,
		AlarmLevel:  req.AlarmLevel,
		Description: req.Description,
		Color:       req.Color,
		Status:      req.Status,
	}
	if fence.AlarmLevel == 0 {
		fence.AlarmLevel = 2 // 默认一般告警
	}
	if fence.Status == 0 {
		fence.Status = 1 // 默认启用
	}
	if fence.Color == "" {
		fence.Color = "#ff4d4f" // 默认红色
	}
	if fence.ScopeType == "" {
		fence.ScopeType = "global"
	}
	if err := h.svc.Create(fence); err != nil {
		response.Error(c, errcode.ErrGeofenceConflict)
		return
	}
	response.OK(c, fence)
}

type UpdateGeofenceRequest struct {
	FenceName   string  `json:"fence_name"`
	FenceType   int8    `json:"fence_type"`
	ActionType  int8    `json:"action_type"`
	ScopeType   string  `json:"scope_type"`
	ScopeID     string  `json:"scope_id"`
	CenterLng   float64 `json:"center_lng"`
	CenterLat   float64 `json:"center_lat"`
	Radius      float64 `json:"radius"`
	Points      string  `json:"points"`
	AlarmLevel  int8    `json:"alarm_level"`
	Description string  `json:"description"`
	Color       string  `json:"color"`
	Status      *int8   `json:"status"`
}

// Update 更新电子围栏。
// PUT /api/v1/geofences/:id
func (h *GeofenceHandler) Update(c *gin.Context) {
	id := c.Param("id")
	fence, err := h.svc.GetByID(id)
	if err != nil {
		response.Error(c, errcode.ErrGeofenceNotFound)
		return
	}
	var req UpdateGeofenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	if req.FenceName != "" {
		fence.FenceName = req.FenceName
	}
	if req.FenceType > 0 {
		fence.FenceType = req.FenceType
	}
	if req.ActionType > 0 {
		fence.ActionType = req.ActionType
	}
	if req.ScopeType != "" {
		fence.ScopeType = req.ScopeType
	}
	if req.ScopeID != "" {
		fence.ScopeID = req.ScopeID
	}
	if req.CenterLng != 0 {
		fence.CenterLng = req.CenterLng
	}
	if req.CenterLat != 0 {
		fence.CenterLat = req.CenterLat
	}
	if req.Radius != 0 {
		fence.Radius = req.Radius
	}
	if req.Points != "" {
		fence.Points = req.Points
	}
	if req.AlarmLevel > 0 {
		fence.AlarmLevel = req.AlarmLevel
	}
	if req.Description != "" {
		fence.Description = req.Description
	}
	if req.Color != "" {
		fence.Color = req.Color
	}
	if req.Status != nil {
		fence.Status = *req.Status
	}
	if err := h.svc.Update(fence); err != nil {
		response.Error(c, errcode.ErrGeofenceConflict)
		return
	}
	response.OK(c, fence)
}

// Delete 删除电子围栏。
// DELETE /api/v1/geofences/:id
func (h *GeofenceHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(id); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, nil)
}

// =========================== 围栏告警管理 ===========================

// ListAlarms 围栏告警记录列表（分页）。
// GET /api/v1/geofence-alarms?page=1&size=10&fence_id=xxx&robot_id=xxx&trigger_type=1&handle_status=0
func (h *GeofenceHandler) ListAlarms(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	fenceID := c.Query("fence_id")
	robotID := c.Query("robot_id")
	stationID := c.Query("station_id")
	triggerType, _ := strconv.Atoi(c.DefaultQuery("trigger_type", "0"))
	handleStatus, _ := strconv.Atoi(c.DefaultQuery("handle_status", "-1"))

	alarms, total, err := h.svc.ListAlarms(page, size, fenceID, robotID, stationID, triggerType, handleStatus)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OKPage(c, alarms, total, page, size)
}

// GetRecentAlarms 最近未处理的围栏告警。
// GET /api/v1/geofence-alarms/recent?limit=10
func (h *GeofenceHandler) GetRecentAlarms(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	alarms, err := h.svc.GetRecentAlarms(limit)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, alarms)
}

type HandleGeofenceAlarmRequest struct {
	HandleStatus int8   `json:"handle_status" binding:"required"` // 处理状态
	Remark       string `json:"remark"`                            // 处理备注
}

// HandleAlarm 处理围栏告警。
// PUT /api/v1/geofence-alarms/:id
func (h *GeofenceHandler) HandleAlarm(c *gin.Context) {
	id := c.Param("id")
	var req HandleGeofenceAlarmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	if err := h.svc.HandleAlarm(id, req.Remark, req.HandleStatus); err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.OK(c, nil)
}
