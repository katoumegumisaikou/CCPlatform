package service

import (
	"ccplatform/internal/repository"
	"log"
	"time"
)

// PHMChecker 故障预测与健康管理周期检测器，周期性地为所有在线机器人计算健康指数并运行故障预测。
// 实现需求 2.1/2.2 所述"实时健康监测"与"故障发生前3-7天预警"，是从被动维修向预测性维护转变的核心调度组件。
type PHMChecker struct {
	phmService *PHMService
	robotRepo  *repository.RobotRepo
	interval   time.Duration
	stopCh     chan struct{}
}

// NewPHMChecker 创建 PHMChecker 实例。
func NewPHMChecker(phmService *PHMService, interval time.Duration) *PHMChecker {
	return &PHMChecker{
		phmService: phmService,
		robotRepo:  repository.NewRobotRepo(),
		interval:   interval,
		stopCh:     make(chan struct{}),
	}
}

// Start 启动周期检测循环：启动时先做一次故障模式库种子初始化，随后周期计算健康指数与故障预测。
func (c *PHMChecker) Start() {
	log.Printf("[PHMChecker] Started with interval %v", c.interval)
	if err := c.phmService.SeedFaultPatterns(); err != nil {
		log.Printf("[PHMChecker] Seed fault patterns error: %v", err)
	}

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	// 启动时立即执行一次检测
	c.checkAll()

	for {
		select {
		case <-ticker.C:
			c.checkAll()
		case <-c.stopCh:
			log.Println("[PHMChecker] Stopped")
			return
		}
	}
}

// Stop 停止检测器。
func (c *PHMChecker) Stop() {
	close(c.stopCh)
}

// checkAll 为所有在线机器人计算健康指数并运行故障预测。
func (c *PHMChecker) checkAll() {
	robots, err := c.robotRepo.GetOnlineRobots()
	if err != nil {
		log.Printf("[PHMChecker] Get online robots error: %v", err)
		return
	}

	checkedCount := 0
	riskRobotCount := 0
	riskComponentCount := 0
	errorCount := 0
	errorSamples := make([]string, 0, 3)

	for _, robot := range robots {
		predictions, err := c.phmService.RunPredictionForRobot(robot.RobotID)
		if err != nil {
			errorCount++
			if len(errorSamples) < 3 {
				errorSamples = append(errorSamples, robot.RobotID+": "+err.Error())
			}
			continue
		}
		checkedCount++
		if len(predictions) > 0 {
			riskRobotCount++
			riskComponentCount += len(predictions)
		}
	}

	if errorCount > 0 {
		log.Printf("[PHMChecker] Checked %d/%d online robots, risk robots=%d, risk components=%d, errors=%d, samples=%v",
			checkedCount, len(robots), riskRobotCount, riskComponentCount, errorCount, errorSamples)
		return
	}
	log.Printf("[PHMChecker] Checked %d online robots, risk robots=%d, risk components=%d",
		checkedCount, riskRobotCount, riskComponentCount)
}
