package handler

import (
	"ccplatform/internal/model"
	"ccplatform/internal/service"
	"ccplatform/pkg/errcode"
	"ccplatform/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// StationHandler 电站 HTTP 处理器，处理电站相关的 RESTful API 请求。
type StationHandler struct {
	svc *service.StationService
}

// NewStationHandler 创建 StationHandler 实例。
func NewStationHandler() *StationHandler {
	return &StationHandler{svc: service.NewStationService()}
}

// List 电站列表接口，支持分页和关键词搜索。
// GET /api/v1/stations?page=1&size=10&keyword=xxx
func (h *StationHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	keyword := c.Query("keyword")

	stations, total, err := h.svc.List(page, size, keyword)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OKPage(c, stations, total, page, size)
}

// GetAll 查询所有启用状态的电站，用于下拉选择框。
// GET /api/v1/stations/all
func (h *StationHandler) GetAll(c *gin.Context) {
	stations, err := h.svc.GetAll()
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, stations)
}

// GetByID 电站详情接口。
// GET /api/v1/stations/:id
func (h *StationHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	station, err := h.svc.GetByID(id)
	if err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.OK(c, station)
}

// CreateStationRequest 创建电站请求体。
type CreateStationRequest struct {
	StationName string  `json:"station_name" binding:"required"` // 电站名称（必填）
	StationCode string  `json:"station_code" binding:"required"` // 电站编码（必填）
	Location    string  `json:"location"`                        // 电站地址
	Longitude   float64 `json:"longitude"`                       // 经度
	Latitude    float64 `json:"latitude"`                        // 纬度
	Capacity    float64 `json:"capacity"`                        // 装机容量(MW)
	PanelArea   float64 `json:"panel_area"`                      // 组件面积(㎡)
	PanelCount  int     `json:"panel_count"`                     // 组件数量
}

// Create 创建电站接口。
// POST /api/v1/stations
func (h *StationHandler) Create(c *gin.Context) {
	var req CreateStationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	station := &model.Station{
		StationName: req.StationName,
		StationCode: req.StationCode,
		Location:    req.Location,
		Longitude:   req.Longitude,
		Latitude:    req.Latitude,
		Capacity:    req.Capacity,
		PanelArea:   req.PanelArea,
		PanelCount:  req.PanelCount,
		Status:      1,
	}
	if err := h.svc.Create(station); err != nil {
		response.Error(c, errcode.ErrDuplicate)
		return
	}
	response.OK(c, station)
}

// UpdateStationRequest 更新电站请求体，所有字段可选。
type UpdateStationRequest struct {
	StationName string  `json:"station_name"` // 电站名称
	Location    string  `json:"location"`     // 电站地址
	Longitude   float64 `json:"longitude"`    // 经度
	Latitude    float64 `json:"latitude"`     // 纬度
	Capacity    float64 `json:"capacity"`     // 装机容量(MW)
	PanelArea   float64 `json:"panel_area"`   // 组件面积(㎡)
	PanelCount  int     `json:"panel_count"`  // 组件数量
	Status      *int8   `json:"status"`       // 状态(0禁用 1启用)
}

// Update 更新电站接口，采用部分更新策略（仅更新非零值字段）。
// PUT /api/v1/stations/:id
func (h *StationHandler) Update(c *gin.Context) {
	id := c.Param("id")
	station, err := h.svc.GetByID(id)
	if err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	var req UpdateStationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	if req.StationName != "" {
		station.StationName = req.StationName
	}
	if req.Location != "" {
		station.Location = req.Location
	}
	if req.Longitude != 0 {
		station.Longitude = req.Longitude
	}
	if req.Latitude != 0 {
		station.Latitude = req.Latitude
	}
	if req.Capacity != 0 {
		station.Capacity = req.Capacity
	}
	if req.PanelArea != 0 {
		station.PanelArea = req.PanelArea
	}
	if req.PanelCount != 0 {
		station.PanelCount = req.PanelCount
	}
	if req.Status != nil {
		station.Status = *req.Status
	}
	if err := h.svc.Update(station); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, station)
}

// Delete 删除电站接口。
// DELETE /api/v1/stations/:id
func (h *StationHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(id); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, nil)
}
