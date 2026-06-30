package service

import (
	"ccplatform/internal/repository"
	"log"
	"time"
)

// GeofenceChecker 电子围栏检测器，周期检测机器人位置是否触发围栏规则。
// 每隔 interval 扫描所有在线机器人的最新 GPS 位置，匹配活跃围栏。
type GeofenceChecker struct {
	geoService *GeofenceService
	robotRepo  *repository.RobotRepo
	interval   time.Duration
	stopCh     chan struct{}
}

// NewGeofenceChecker 创建 GeofenceChecker 实例。
func NewGeofenceChecker(geoService *GeofenceService, interval time.Duration) *GeofenceChecker {
	return &GeofenceChecker{
		geoService: geoService,
		robotRepo:  repository.NewRobotRepo(),
		interval:   interval,
		stopCh:     make(chan struct{}),
	}
}

// Start 启动周期检测循环。
func (c *GeofenceChecker) Start() {
	log.Printf("[GeofenceChecker] Started with interval %v", c.interval)
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	// 启动时立即执行一次检测
	c.checkAll()

	for {
		select {
		case <-ticker.C:
			c.checkAll()
		case <-c.stopCh:
			log.Println("[GeofenceChecker] Stopped")
			return
		}
	}
}

// Stop 停止检测器。
func (c *GeofenceChecker) Stop() {
	close(c.stopCh)
}

// checkAll 检测所有在线机器人是否触发电子围栏规则。
func (c *GeofenceChecker) checkAll() {
	robots, err := c.robotRepo.GetOnlineRobots()
	if err != nil {
		log.Printf("[GeofenceChecker] Get online robots error: %v", err)
		return
	}

	for _, robot := range robots {
		// 只对具有有效 GPS 坐标的机器人进行围栏检测
		if robot.GPSLongitude == 0 && robot.GPSLatitude == 0 {
			// 如果没有 GPS 坐标，尝试使用场景坐标近似
			continue
		}

		alarms, err := c.geoService.CheckRobotPosition(robot.RobotID, robot.GPSLongitude, robot.GPSLatitude)
		if err != nil {
			log.Printf("[GeofenceChecker] Check robot %s error: %v", robot.RobotID, err)
			continue
		}
		if len(alarms) > 0 {
			log.Printf("[GeofenceChecker] Robot %s triggered %d geofence alarm(s)", robot.RobotID, len(alarms))
		}
	}
}
