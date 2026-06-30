package service

import (
	"ccplatform/internal/model"
	"ccplatform/internal/repository"
	"ccplatform/internal/ws"
	"ccplatform/pkg/util"
	"encoding/json"
	"fmt"
	"math"
	"time"
)

// GeofenceService 电子围栏业务逻辑层。
// 提供围栏 CRUD、围栏告警管理，以及机器人位置与围栏的匹配检测。
type GeofenceService struct {
	fenceRepo      *repository.GeofenceRepo
	alarmRepo      *repository.GeofenceAlarmRepo
	robotRepo      *repository.RobotRepo
	alarmSvc       *AlarmService // 推送告警到告警中心
	hub            *ws.Hub       // WebSocket 实时推送
	cooldownSec    int           // 同一机器人同一围栏的告警冷却时间（秒）
}

// NewGeofenceService 创建 GeofenceService 实例。
func NewGeofenceService(hub *ws.Hub) *GeofenceService {
	return &GeofenceService{
		fenceRepo:   repository.NewGeofenceRepo(),
		alarmRepo:   repository.NewGeofenceAlarmRepo(),
		robotRepo:   repository.NewRobotRepo(),
		alarmSvc:    NewAlarmService(),
		hub:         hub,
		cooldownSec: 300, // 默认 5 分钟冷却
	}
}

// =========================== 围栏 CRUD ===========================

// Create 创建电子围栏。
func (s *GeofenceService) Create(fence *model.Geofence) error {
	fence.GeofenceID = util.GenerateID()
	if err := s.validateFence(fence); err != nil {
		return err
	}
	return s.fenceRepo.Create(fence)
}

// GetByID 根据 ID 查询围栏。
func (s *GeofenceService) GetByID(id string) (*model.Geofence, error) {
	return s.fenceRepo.GetByID(id)
}

// Update 更新电子围栏。
func (s *GeofenceService) Update(fence *model.Geofence) error {
	if err := s.validateFence(fence); err != nil {
		return err
	}
	return s.fenceRepo.Update(fence)
}

// Delete 删除电子围栏。
func (s *GeofenceService) Delete(id string) error {
	return s.fenceRepo.Delete(id)
}

// List 分页查询电子围栏列表。
func (s *GeofenceService) List(page, size int, keyword, scopeType string, fenceType int) ([]model.Geofence, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	return s.fenceRepo.List(page, size, keyword, scopeType, fenceType)
}

// GetAll 查询所有启用围栏。
func (s *GeofenceService) GetAll() ([]model.Geofence, error) {
	return s.fenceRepo.GetAll()
}

// validateFence 校验围栏参数的合法性。
func (s *GeofenceService) validateFence(fence *model.Geofence) error {
	if fence.FenceName == "" {
		return fmt.Errorf("围栏名称不能为空")
	}
	if fence.FenceType < 1 || fence.FenceType > 2 {
		return fmt.Errorf("围栏类型无效: 1=圆形, 2=多边形")
	}
	if fence.ActionType < 1 || fence.ActionType > 2 {
		return fmt.Errorf("动作类型无效: 1=禁止进入, 2=禁止离开")
	}

	if fence.FenceType == 1 {
		// 圆形围栏校验
		if fence.CenterLng == 0 && fence.CenterLat == 0 {
			return fmt.Errorf("圆形围栏需要指定圆心坐标")
		}
		if fence.Radius <= 0 {
			return fmt.Errorf("圆形围栏半径必须大于 0")
		}
	} else if fence.FenceType == 2 {
		// 多边形围栏校验
		if fence.Points == "" {
			return fmt.Errorf("多边形围栏需要指定顶点坐标")
		}
		var pts [][2]float64
		if err := json.Unmarshal([]byte(fence.Points), &pts); err != nil {
			return fmt.Errorf("顶点坐标格式无效: %v", err)
		}
		if len(pts) < 3 {
			return fmt.Errorf("多边形至少需要 3 个顶点")
		}
	}
	return nil
}

// =========================== 围栏告警 CRUD ===========================

// ListAlarms 分页查询围栏告警记录。
func (s *GeofenceService) ListAlarms(page, size int, fenceID, robotID, stationID string, triggerType, handleStatus int) ([]model.GeofenceAlarm, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	return s.alarmRepo.List(page, size, fenceID, robotID, stationID, triggerType, handleStatus)
}

// GetRecentAlarms 查询最近未处理的围栏告警。
func (s *GeofenceService) GetRecentAlarms(limit int) ([]model.GeofenceAlarm, error) {
	return s.alarmRepo.GetRecent(limit)
}

// HandleAlarm 处理围栏告警。
func (s *GeofenceService) HandleAlarm(id, remark string, handleStatus int8) error {
	alarm, err := s.alarmRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("告警不存在: %v", err)
	}
	now := time.Now()
	alarm.HandleStatus = handleStatus
	alarm.HandleTime = &now
	alarm.HandleRemark = remark
	return s.alarmRepo.Update(alarm)
}

// =========================== 围栏检测算法 ===========================

// PointInCircle 判断点是否在圆形区域内。
// 使用 Haversine 公式计算球面距离。
func (s *GeofenceService) PointInCircle(lng, lat, centerLng, centerLat, radiusMeters float64) bool {
	const R = 6371000 // 地球半径（米）
	dLat := (lat - centerLat) * math.Pi / 180
	dLng := (lng - centerLng) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(centerLat*math.Pi/180)*math.Cos(lat*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := R * c
	return distance <= radiusMeters
}

// PointInPolygon 判断点是否在多边形内部（射线法 Ray Casting Algorithm）。
// 从目标点向右发射水平射线，计算与多边形边的交点数量：
//
//	奇数 → 点在多边形内部；偶数 → 点在多边形外部。
func (s *GeofenceService) PointInPolygon(lng, lat float64, points [][2]float64) bool {
	n := len(points)
	if n < 3 {
		return false
	}
	inside := false
	j := n - 1
	for i := 0; i < n; i++ {
		// 检查射线是否穿过边 (points[i], points[j])
		yi, yj := points[i][1], points[j][1]
		xi, xj := points[i][0], points[j][0]

		// 判断点的纬度是否在边的纬度范围内
		if (yi > lat) != (yj > lat) {
			// 计算射线与边的交点的经度
			intersectLng := xi + (lat-yi)*(xj-xi)/(yj-yi)
			if lng < intersectLng {
				inside = !inside
			}
		}
		j = i
	}
	return inside
}

// ParsePolygonPoints 解析多边形顶点坐标 JSON 字符串。
func (s *GeofenceService) ParsePolygonPoints(pointsJSON string) ([][2]float64, error) {
	var pts [][2]float64
	if err := json.Unmarshal([]byte(pointsJSON), &pts); err != nil {
		return nil, fmt.Errorf("解析顶点坐标失败: %v", err)
	}
	return pts, nil
}

// CheckRobotPosition 检测机器人位置是否触发电子围栏规则。
// 返回触发的告警列表（可能同时触发多个围栏）。
// 该方法设计为由 MQTT 位置消息处理器或定时任务调用。
func (s *GeofenceService) CheckRobotPosition(robotID string, lng, lat float64) ([]*model.GeofenceAlarm, error) {
	robot, err := s.robotRepo.GetByID(robotID)
	if err != nil {
		return nil, fmt.Errorf("获取机器人信息失败: %v", err)
	}

	// 查询可能关联到该机器人的围栏
	// 优先级：机器人 > 电站 > 全局
	var fences []model.Geofence

	// 查询机器人级别的围栏
	robotFences, _ := s.fenceRepo.GetByScope("robot", robotID)
	fences = append(fences, robotFences...)

	// 查询电站级别的围栏
	stationFences, _ := s.fenceRepo.GetByScope("station", robot.StationID)
	fences = append(fences, stationFences...)

	// 查询全局围栏
	globalFences, _ := s.fenceRepo.GetGlobalFences()
	fences = append(fences, globalFences...)

	if len(fences) == 0 {
		return nil, nil
	}

	now := time.Now()
	var triggeredAlarms []*model.GeofenceAlarm

	for _, fence := range fences {
		var inside bool
		if fence.FenceType == 1 {
			// 圆形围栏
			inside = s.PointInCircle(lng, lat, fence.CenterLng, fence.CenterLat, fence.Radius)
		} else {
			// 多边形围栏
			pts, err := s.ParsePolygonPoints(fence.Points)
			if err != nil {
				continue
			}
			inside = s.PointInPolygon(lng, lat, pts)
		}

		shouldAlarm := false
		var triggerType int8

		if fence.ActionType == 1 {
			// 禁止进入：在围栏内部 → 触发告警
			shouldAlarm = inside
			triggerType = 1
		} else if fence.ActionType == 2 {
			// 禁止离开：在围栏外部 → 触发告警
			shouldAlarm = !inside
			triggerType = 2
		}

		if !shouldAlarm {
			continue
		}

		// 检查冷却期，避免重复告警
		hasRecent, err := s.alarmRepo.HasRecentAlarm(fence.GeofenceID, robotID, triggerType, s.cooldownSec)
		if err != nil || hasRecent {
			continue
		}

		// 构建告警内容
		actionText := "进入禁区"
		if triggerType == 2 {
			actionText = "离开工作区"
		}

		alarmContent := fmt.Sprintf("机器人 %s %s「%s」，位置(%.6f, %.6f)",
			robot.RobotName, actionText, fence.FenceName, lng, lat)

		alarm := &model.GeofenceAlarm{
			AlarmID:      util.GenerateID(),
			FenceID:      fence.GeofenceID,
			FenceName:    fence.FenceName,
			RobotID:      robotID,
			RobotName:    robot.RobotName,
			StationID:    robot.StationID,
			TriggerType:  triggerType,
			PositionLng:  lng,
			PositionLat:  lat,
			AlarmContent: alarmContent,
			AlarmTime:    now,
			HandleStatus: 0,
		}

		// 保存到围栏告警表
		if err := s.alarmRepo.Create(alarm); err != nil {
			continue
		}

		// 同时推送到告警中心
		alarmType := "geofence_forbid_entry"
		if triggerType == 2 {
			alarmType = "geofence_leave_area"
		}
		_ = s.alarmSvc.CreateFromGeofence(robotID, robot.StationID, alarmType, fence.AlarmLevel, alarmContent)

		// WebSocket 推送
		s.pushGeofenceAlarm(alarm)

		triggeredAlarms = append(triggeredAlarms, alarm)
	}

	return triggeredAlarms, nil
}

// pushGeofenceAlarm 通过 WebSocket 实时推送围栏告警。
func (s *GeofenceService) pushGeofenceAlarm(alarm *model.GeofenceAlarm) {
	if s.hub == nil {
		return
	}
	s.hub.BroadcastToAll(ws.Message{
		Type: "geofence_alarm",
		Data: map[string]interface{}{
			"alarm_id":      alarm.AlarmID,
			"fence_id":      alarm.FenceID,
			"fence_name":    alarm.FenceName,
			"robot_id":      alarm.RobotID,
			"robot_name":    alarm.RobotName,
			"trigger_type":  alarm.TriggerType,
			"position_lng":  alarm.PositionLng,
			"position_lat":  alarm.PositionLat,
			"alarm_content": alarm.AlarmContent,
			"alarm_time":    alarm.AlarmTime.Format(time.RFC3339),
		},
	})
}
