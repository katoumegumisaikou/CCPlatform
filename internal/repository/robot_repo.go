package repository

import (
	"ccplatform/internal/model"
	"time"

	"gorm.io/gorm"
)

// RobotRepo 机器人数据访问层，封装 robots 表的所有数据库操作。
// 提供标准 CRUD 以及 MQTT 上行数据更新方法（心跳、位置、状态）。
type RobotRepo struct {
	db *gorm.DB
}

// NewRobotRepo 创建 RobotRepo 实例，使用全局 DB 连接。
func NewRobotRepo() *RobotRepo {
	return &RobotRepo{db: DB}
}

// Create 创建新机器人记录。
func (r *RobotRepo) Create(robot *model.Robot) error {
	return r.db.Create(robot).Error
}

// GetByID 根据机器人 ID 查询单条记录。
func (r *RobotRepo) GetByID(id string) (*model.Robot, error) {
	var robot model.Robot
	err := r.db.Where("robot_id = ?", id).First(&robot).Error
	return &robot, err
}

// Update 更新机器人信息（全量更新）。
func (r *RobotRepo) Update(robot *model.Robot) error {
	return r.db.Save(robot).Error
}

// Delete 根据机器人 ID 删除记录。
func (r *RobotRepo) Delete(id string) error {
	return r.db.Where("robot_id = ?", id).Delete(&model.Robot{}).Error
}

// List 分页查询机器人列表，支持按电站、类型、在线状态筛选。
// robotType=0 表示不筛选类型，onlineStatus<0 表示不筛选在线状态。
func (r *RobotRepo) List(page, size int, stationID string, robotType int, onlineStatus int) ([]model.Robot, int64, error) {
	var robots []model.Robot
	var total int64
	query := r.db.Model(&model.Robot{})
	if stationID != "" {
		query = query.Where("station_id = ?", stationID)
	}
	if robotType > 0 {
		query = query.Where("robot_type = ?", robotType)
	}
	if onlineStatus >= 0 {
		query = query.Where("online_status = ?", onlineStatus)
	}
	query.Count(&total)
	err := query.Offset((page - 1) * size).Limit(size).Order("create_time DESC").Find(&robots).Error
	return robots, total, err
}

// GetByStationID 查询指定电站下的所有机器人。
func (r *RobotRepo) GetByStationID(stationID string) ([]model.Robot, error) {
	var robots []model.Robot
	err := r.db.Where("station_id = ?", stationID).Find(&robots).Error
	return robots, err
}

// UpdateHeartbeat 更新机器人心跳数据，由 MQTT 心跳消息处理器调用。
// 同时更新在线状态、电量、工作状态和最后心跳时间。
func (r *RobotRepo) UpdateHeartbeat(id string, onlineStatus int8, batteryLevel int8, workStatus int8) error {
	return r.db.Model(&model.Robot{}).Where("robot_id = ?", id).Updates(map[string]interface{}{
		"online_status":  onlineStatus,
		"battery_level":  batteryLevel,
		"work_status":    workStatus,
		"last_heartbeat": time.Now(),
	}).Error
}

// UpdatePosition 更新机器人位置坐标，由 MQTT 位置消息处理器调用。
// 更新三维坐标 (X/Y/Z) 和朝向角度。
func (r *RobotRepo) UpdatePosition(id string, posX, posY, posZ, heading float64) error {
	return r.db.Model(&model.Robot{}).Where("robot_id = ?", id).Updates(map[string]interface{}{
		"pos_x":   posX,
		"pos_y":   posY,
		"pos_z":   posZ,
		"heading": heading,
	}).Error
}

// UpdateStatus 更新机器人运行状态，由 MQTT 状态消息处理器调用。
// data map 包含 work_status、speed、clean_area、fault_code 等字段。
func (r *RobotRepo) UpdateStatus(id string, data map[string]interface{}) error {
	return r.db.Model(&model.Robot{}).Where("robot_id = ?", id).Updates(data).Error
}

// CountByStation 统计指定电站的机器人总数。
func (r *RobotRepo) CountByStation(stationID string) (int64, error) {
	var count int64
	err := r.db.Model(&model.Robot{}).Where("station_id = ?", stationID).Count(&count).Error
	return count, err
}

// CountOnlineByStation 统计指定电站的在线机器人数量。
func (r *RobotRepo) CountOnlineByStation(stationID string) (int64, error) {
	var count int64
	err := r.db.Model(&model.Robot{}).Where("station_id = ? AND online_status = 1", stationID).Count(&count).Error
	return count, err
}

// CountByStatus 按状态统计机器人数量，返回 total/online/offline/fault 四个维度。
// 用于监控中心仪表盘展示。
func (r *RobotRepo) CountByStatus() (map[string]int64, error) {
	result := make(map[string]int64)
	var total int64
	r.db.Model(&model.Robot{}).Count(&total)
	result["total"] = total

	var online int64
	r.db.Model(&model.Robot{}).Where("online_status = 1").Count(&online)
	result["online"] = online

	var offline int64
	r.db.Model(&model.Robot{}).Where("online_status = 0").Count(&offline)
	result["offline"] = offline

	var fault int64
	r.db.Model(&model.Robot{}).Where("work_status = 3").Count(&fault)
	result["fault"] = fault

	return result, nil
}

// GetOnlineRobots 查询所有在线机器人，用于监控面板和实时数据推送。
func (r *RobotRepo) GetOnlineRobots() ([]model.Robot, error) {
	var robots []model.Robot
	err := r.db.Where("online_status = 1").Find(&robots).Error
	return robots, err
}
