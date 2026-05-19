package service

import (
	"ccplatform/internal/model"
	"ccplatform/internal/repository"
)

// MonitorService 监控中心业务逻辑层，聚合多维度数据供仪表盘和实时监控使用。
type MonitorService struct {
	stationRepo *repository.StationRepo
	robotRepo   *repository.RobotRepo
	taskRepo    *repository.TaskRepo
	alarmRepo   *repository.AlarmRepo
}

// NewMonitorService 创建 MonitorService 实例，注入所有需要的 Repo。
func NewMonitorService() *MonitorService {
	return &MonitorService{
		stationRepo: repository.NewStationRepo(),
		robotRepo:   repository.NewRobotRepo(),
		taskRepo:    repository.NewTaskRepo(),
		alarmRepo:   repository.NewAlarmRepo(),
	}
}

// DashboardOverview 监控仪表盘概览数据，包含电站、机器人、任务、告警的汇总统计。
type DashboardOverview struct {
	StationCount    int64            `json:"station_count"`     // 电站总数
	RobotStats      map[string]int64 `json:"robot_stats"`       // 机器人统计(总数/在线/离线/故障)
	TodayCleanArea  float64          `json:"today_clean_area"`  // 今日清扫面积(㎡)
	RunningTasks    int64            `json:"running_tasks"`     // 执行中任务数
	UnhandledAlarms int64            `json:"unhandled_alarms"`  // 未处理告警数
	AlarmStats      map[string]int64 `json:"alarm_stats"`       // 告警级别统计(info/warning/error/critical)
}

// GetDashboard 获取监控仪表盘概览数据，聚合电站、机器人、告警等多个维度。
func (s *MonitorService) GetDashboard() (*DashboardOverview, error) {
	stations, err := s.stationRepo.GetAll()
	if err != nil {
		return nil, err
	}

	robotStats, err := s.robotRepo.CountByStatus()
	if err != nil {
		return nil, err
	}

	alarmStats, err := s.alarmRepo.CountByLevel()
	if err != nil {
		return nil, err
	}

	unhandled, err := s.alarmRepo.CountUnhandled()
	if err != nil {
		return nil, err
	}

	todayCleanArea, err := s.taskRepo.SumTodayCleanArea()
	if err != nil {
		return nil, err
	}

	runningTasks, err := s.taskRepo.CountRunning()
	if err != nil {
		return nil, err
	}

	return &DashboardOverview{
		StationCount:    int64(len(stations)),
		RobotStats:      robotStats,
		TodayCleanArea:  todayCleanArea,
		RunningTasks:    runningTasks,
		UnhandledAlarms: unhandled,
		AlarmStats:      alarmStats,
	}, nil
}

// StationRealtime 电站实时数据，包含电站信息和下属所有机器人的实时状态。
type StationRealtime struct {
	StationID   string          `json:"station_id"`   // 电站 ID
	StationName string          `json:"station_name"` // 电站名称
	Longitude   float64         `json:"longitude"`    // 经度
	Latitude    float64         `json:"latitude"`     // 纬度
	Robots      []RobotRealtime `json:"robots"`       // 机器人实时数据列表
}

// RobotRealtime 机器人实时数据，用于 GIS 地图和监控面板展示。
type RobotRealtime struct {
	RobotID        string  `json:"robot_id"`        // 机器人 ID
	RobotName      string  `json:"robot_name"`      // 机器人名称
	RobotType      int8    `json:"robot_type"`      // 机器人类型
	OnlineStatus   int8    `json:"online_status"`   // 在线状态
	WorkStatus     int8    `json:"work_status"`     // 工作状态
	BatteryLevel   int8    `json:"battery_level"`   // 电量百分比
	PosX           float64 `json:"pos_x"`           // X 坐标
	PosY           float64 `json:"pos_y"`           // Y 坐标
	PosZ           float64 `json:"pos_z"`           // Z 坐标
	Heading        float64 `json:"heading"`         // 朝向角度
	Speed          float64 `json:"speed"`           // 速度(m/s)
	CleanArea      float64 `json:"clean_area"`      // 累计清扫面积(㎡)
	Temperature    float64 `json:"temperature"`     // 环境温度(℃)
	Humidity       float64 `json:"humidity"`        // 环境湿度(%)
	LightIntensity float64 `json:"light_intensity"` // 光照强度(lux)
	WindSpeed      float64 `json:"wind_speed"`      // 风速(m/s)
}

// GetRealtimeData 获取指定电站的实时监控数据，包含电站下所有机器人的位置和状态。
func (s *MonitorService) GetRealtimeData(stationID string) (*StationRealtime, error) {
	station, err := s.stationRepo.GetByID(stationID)
	if err != nil {
		return nil, err
	}

	robots, err := s.robotRepo.GetByStationID(stationID)
	if err != nil {
		return nil, err
	}

	var robotRealtimes []RobotRealtime
	for _, r := range robots {
		robotRealtimes = append(robotRealtimes, RobotRealtime{
			RobotID:        r.RobotID,
			RobotName:      r.RobotName,
			RobotType:      r.RobotType,
			OnlineStatus:   r.OnlineStatus,
			WorkStatus:     r.WorkStatus,
			BatteryLevel:   r.BatteryLevel,
			PosX:           r.PosX,
			PosY:           r.PosY,
			PosZ:           r.PosZ,
			Heading:        r.Heading,
			Speed:          r.Speed,
			CleanArea:      r.CleanArea,
			Temperature:    r.Temperature,
			Humidity:       r.Humidity,
			LightIntensity: r.LightIntensity,
			WindSpeed:      r.WindSpeed,
		})
	}

	return &StationRealtime{
		StationID:   station.StationID,
		StationName: station.StationName,
		Longitude:   station.Longitude,
		Latitude:    station.Latitude,
		Robots:      robotRealtimes,
	}, nil
}

// GetPositionHistory 查询机器人历史轨迹，支持机器人 ID 和时间段筛选。
func (s *MonitorService) GetPositionHistory(robotID, startTime, endTime string, limit int) ([]model.RobotPosition, error) {
	return s.robotRepo.GetPositionHistory(robotID, startTime, endTime, limit)
}
