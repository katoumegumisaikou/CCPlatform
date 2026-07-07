package handler

import (
	"ccplatform/internal/model"
	"ccplatform/internal/service"
	"ccplatform/pkg/errcode"
	"ccplatform/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CleaningDecisionHandler 清扫决策 HTTP 处理器。
type CleaningDecisionHandler struct {
	svc *service.CleaningDecisionService
}

// NewCleaningDecisionHandler 创建清扫决策 Handler 实例。
func NewCleaningDecisionHandler() *CleaningDecisionHandler {
	return &CleaningDecisionHandler{svc: service.NewCleaningDecisionService()}
}

// ─── 决策规则 ────────────────────────────────────────────────

// ListRules 分页查询清扫决策规则。
// GET /api/v1/cleaning-decisions/rules?station_id=xxx&page=1&size=10
func (h *CleaningDecisionHandler) ListRules(c *gin.Context) {
	stationID := c.Query("station_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	list, total, err := h.svc.ListRules(stationID, page, size)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OKPage(c, list, total, page, size)
}

// CreateRule 创建清扫决策规则。
// POST /api/v1/cleaning-decisions/rules
func (h *CleaningDecisionHandler) CreateRule(c *gin.Context) {
	var rule model.CleaningDecisionRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	if err := h.svc.CreateRule(&rule); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, rule)
}

// GetRule 获取单条决策规则。
// GET /api/v1/cleaning-decisions/rules/:id
func (h *CleaningDecisionHandler) GetRule(c *gin.Context) {
	id := c.Param("id")
	rule, err := h.svc.GetRuleByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, errcode.ErrNotFound)
			return
		}
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, rule)
}

// UpdateRule 更新决策规则。
// PUT /api/v1/cleaning-decisions/rules/:id
func (h *CleaningDecisionHandler) UpdateRule(c *gin.Context) {
	id := c.Param("id")
	var rule model.CleaningDecisionRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	rule.RuleID = id
	if err := h.svc.UpdateRule(&rule); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, nil)
}

// DeleteRule 删除决策规则。
// DELETE /api/v1/cleaning-decisions/rules/:id
func (h *CleaningDecisionHandler) DeleteRule(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteRule(id); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, nil)
}

// ─── 极端天气规则 ────────────────────────────────────────────

func (h *CleaningDecisionHandler) ListExtremeRules(c *gin.Context) {
	stationID := c.Query("station_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	list, total, err := h.svc.ListExtremeRules(stationID, page, size)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OKPage(c, list, total, page, size)
}

func (h *CleaningDecisionHandler) CreateExtremeRule(c *gin.Context) {
	var rule model.ExtremeWeatherRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	if err := h.svc.CreateExtremeRule(&rule); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, rule)
}

func (h *CleaningDecisionHandler) GetExtremeRule(c *gin.Context) {
	id := c.Param("id")
	rule, err := h.svc.GetExtremeRuleByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, errcode.ErrNotFound)
			return
		}
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, rule)
}

func (h *CleaningDecisionHandler) UpdateExtremeRule(c *gin.Context) {
	id := c.Param("id")
	var rule model.ExtremeWeatherRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	rule.RuleID = id
	if err := h.svc.UpdateExtremeRule(&rule); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, nil)
}

func (h *CleaningDecisionHandler) DeleteExtremeRule(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteExtremeRule(id); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, nil)
}

// ─── 清扫建议 ────────────────────────────────────────────────

// GetCleaningAdvice 获取电站的智能清扫建议。
// GET /api/v1/cleaning-decisions/advice?station_id=xxx
func (h *CleaningDecisionHandler) GetCleaningAdvice(c *gin.Context) {
	stationID := c.Query("station_id")
	if stationID == "" {
		response.Error(c, errcode.ErrParam)
		return
	}

	advice, err := h.svc.EvaluateCleaningAdvice(stationID)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, advice)
}

// ─── 动态频次调整 ────────────────────────────────────────────

// GetAdjustments 分页查询频次调整建议记录。
// GET /api/v1/cleaning-decisions/adjustments?station_id=xxx&page=1&size=10
func (h *CleaningDecisionHandler) GetAdjustments(c *gin.Context) {
	stationID := c.Query("station_id")
	if stationID == "" {
		response.Error(c, errcode.ErrParam)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	list, total, err := h.svc.ListAdjustments(stationID, page, size)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OKPage(c, list, total, page, size)
}

// GenerateAdjustment 生成清扫频次调整建议。
// POST /api/v1/cleaning-decisions/adjustments
func (h *CleaningDecisionHandler) GenerateAdjustment(c *gin.Context) {
	var req struct {
		StationID string `json:"station_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}

	advice, err := h.svc.GenerateScheduleAdjustment(req.StationID)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, advice)
}
