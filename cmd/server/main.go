// Package main 是拓达威光伏清扫机器人云控平台的程序入口。
//
// 启动流程：
//  1. 加载配置文件 (config.yaml)
//  2. 连接 MySQL 数据库并自动迁移表结构
//  3. 初始化默认角色和管理员账号
//  4. 启动 WebSocket Hub（实时推送服务）
//  5. 启动 MQTT Broker（嵌入式，接收机器人上行数据）
//  6. 注册 MQTT 消息处理器（解析心跳/位置/状态/告警）
//  7. 启动 HTTP Server（Gin REST API）
//  8. 等待信号优雅关闭
package main

import (
	"ccplatform/internal/config"
	"ccplatform/internal/mqtt"
	"ccplatform/internal/repository"
	"ccplatform/internal/router"
	"ccplatform/internal/scheduler"
	"ccplatform/internal/service"
	"ccplatform/internal/ws"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 1. 加载配置文件，支持通过命令行参数指定路径
	configPath := "config/config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}
	if err := config.Load(configPath); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. 初始化数据库连接并自动建表
	if err := initDB(); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	// 3. 初始化预定义角色（系统管理员/电站管理员/运维工程师/值班员/普通用户）
	//    和默认管理员账号 (admin / admin123)
	userSvc := service.NewUserService()
	if err := userSvc.InitDefaultRoles(); err != nil {
		log.Printf("Warning: init default roles: %v", err)
	}
	if err := userSvc.InitAdminUser(); err != nil {
		log.Printf("Warning: init admin user: %v", err)
	}

	// 4. 启动 WebSocket Hub，管理所有前端 WebSocket 连接
	//    前端通过 ws://host:port/ws?station_id=xxx 连接，接收实时推送
	hub := ws.NewHub()
	go hub.Run()

	// 5. 启动嵌入式 MQTT Broker（基于 mochi-mqtt），监听机器人上行数据
	if err := mqtt.InitBroker(config.Cfg.MQTT.Port); err != nil {
		log.Fatalf("Failed to init MQTT broker: %v", err)
	}

	// 6. 注册 MQTT Hook，处理上行消息：
	//    - tdw/robot/{id}/heartbeat → 更新在线状态/电量 → WebSocket 推送
	//    - tdw/robot/{id}/position  → 更新坐标位置 → WebSocket 推送
	//    - tdw/robot/{id}/status    → 更新运行状态/环境数据 → WebSocket 推送
	//    - tdw/robot/{id}/alarm     → 写入告警表 → WebSocket 推送
	msgHandler := mqtt.NewMessageHandler(hub)
	if err := msgHandler.RegisterHooks(mqtt.Server); err != nil {
		log.Fatalf("Failed to register MQTT hooks: %v", err)
	}

	// 7. 创建 MQTT Publisher，用于下发控制指令到机器人
	//    Topic: tdw/robot/{id}/cmd 和 tdw/robot/{id}/config
	publisher := mqtt.NewPublisher(mqtt.Server)

	// 8. 启动周期任务调度器，用于定时/周期清扫任务
	//    注入 Publisher 以在创建任务实例时下发 MQTT 指令
	taskScheduler := scheduler.NewTaskScheduler(publisher)
	go taskScheduler.Start()

	// 9. 注册所有 HTTP 路由（REST API）
	r := router.Setup(hub, publisher)

	// 10. 启动 HTTP Server
	addr := fmt.Sprintf(":%d", config.Cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		log.Printf("[HTTP] Server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// 11. 阻塞等待 SIGINT/SIGTERM 信号，优雅关闭所有服务
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	taskScheduler.Stop()
	if mqtt.Server != nil {
		mqtt.Server.Close()
		log.Println("[MQTT] Broker stopped")
	}

	log.Println("Server exited")
}

// initDB 初始化数据库连接并自动迁移表结构，由 repository.InitDB 统一处理。
func initDB() error {
	return repository.InitDB()
}
