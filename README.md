# TDW CCPlatform - 拓达威光伏清扫机器人云控平台

基于 Go 语言开发的光伏清扫机器人云端管理与控制平台后端服务。

## 项目概述

本平台用于统一管理光伏电站中的清扫机器人，支持多电站接入、实时监控、任务调度、告警管理和数据分析。机器人通过 MQTT 协议上报数据（心跳、位置、状态、告警），平台将数据持久化到 MySQL 并通过 WebSocket 实时推送给前端。

## 架构设计

```
┌──────────────────────────────────────────────────────────────┐
│                        展示层 (Frontend)                      │
│         Web 管理端 / 移动 APP / 大屏展示 (WebSocket)          │
└──────────────────────────┬───────────────────────────────────┘
                           │ HTTP REST API / WebSocket
┌──────────────────────────▼───────────────────────────────────┐
│                       服务层 (Service)                        │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────────┐ │
│  │ 电站服务  │ │ 机器人服务│ │ 任务服务  │ │ 告警/监控/分析服务│ │
│  └──────────┘ └──────────┘ └──────────┘ └──────────────────┘ │
│  ┌──────────────────┐ ┌──────────────────┐                   │
│  │ JWT 认证中间件    │ │ RBAC 权限中间件   │                   │
│  └──────────────────┘ └──────────────────┘                   │
└──────────────────────────┬───────────────────────────────────┘
                           │
┌──────────────────────────▼───────────────────────────────────┐
│                       接入层 (Access)                         │
│  ┌──────────────────┐  ┌──────────────────┐                  │
│  │ MQTT Broker       │  │ WebSocket Hub    │                  │
│  │ (mochi-mqtt)      │  │ (gorilla/ws)     │                  │
│  │ 端口: 1883        │  │ 端口: 8080/ws    │                  │
│  └────────┬─────────┘  └──────────────────┘                  │
└───────────┼──────────────────────────────────────────────────┘
            │ MQTT 消息
┌───────────▼──────────────────────────────────────────────────┐
│                       数据层 (Storage)                        │
│  ┌──────────────────┐  ┌──────────────────┐                  │
│  │ MySQL 8.0         │  │ GORM ORM         │                  │
│  │ stations/robots   │  │ 自动迁移          │                  │
│  │ tasks/alarms      │  │                  │                  │
│  │ users/roles       │  │                  │                  │
│  └──────────────────┘  └──────────────────┘                  │
└──────────────────────────────────────────────────────────────┘
```

## 技术栈

| 组件 | 技术选型 | 用途 |
|------|---------|------|
| 语言 | Go 1.21+ | 后端服务 |
| Web 框架 | Gin | HTTP REST API |
| 数据库 | MySQL 8.0 | 业务数据存储 |
| ORM | GORM | 数据库操作 |
| MQTT | mochi-mqtt | 嵌入式 MQTT Broker |
| WebSocket | gorilla/websocket | 实时数据推送 |
| 认证 | JWT (golang-jwt) | 用户认证 |
| 密码 | bcrypt | 密码哈希 |
| 配置 | Viper | 配置文件管理 |

## 项目结构

```
CCPlatform/
├── cmd/server/main.go              # 程序入口，启动顺序：配置→DB→MQTT→HTTP
├── config/config.yaml              # 配置文件（数据库、MQTT、JWT等）
├── migrations/init.sql             # MySQL 建表和初始化数据脚本
├── internal/
│   ├── config/config.go            # 配置结构体定义和加载
│   ├── model/                      # 数据模型（GORM，对应数据库表）
│   │   ├── station.go              # 光伏电站
│   │   ├── robot.go                # 清扫机器人
│   │   ├── task.go                 # 清扫任务
│   │   ├── alarm.go                # 告警记录
│   │   ├── user.go                 # 系统用户
│   │   └── role.go                 # 角色权限
│   ├── repository/                 # 数据访问层（CRUD操作）
│   │   ├── db.go                   # 数据库连接初始化
│   │   ├── station_repo.go         # 电站数据操作
│   │   ├── robot_repo.go           # 机器人数据操作
│   │   ├── task_repo.go            # 任务数据操作
│   │   ├── alarm_repo.go           # 告警数据操作
│   │   └── user_repo.go            # 用户/角色数据操作
│   ├── service/                    # 业务逻辑层
│   │   ├── station_service.go      # 电站业务（CRUD）
│   │   ├── robot_service.go        # 机器人业务（CRUD+控制）
│   │   ├── task_service.go         # 任务业务（CRUD+状态流转）
│   │   ├── alarm_service.go        # 告警业务（CRUD+处理）
│   │   ├── user_service.go         # 用户业务（登录+RBAC）
│   │   ├── monitor_service.go      # 监控中心（Dashboard+实时数据）
│   │   └── analytics_service.go    # 数据分析（经济效益+效率报表）
│   ├── handler/                    # HTTP 处理器（Gin Handler）
│   │   ├── station_handler.go      # 电站 API
│   │   ├── robot_handler.go        # 机器人 API
│   │   ├── task_handler.go         # 任务 API
│   │   ├── alarm_handler.go        # 告警 API
│   │   ├── user_handler.go         # 用户/角色 API
│   │   ├── monitor_handler.go      # 监控 API
│   │   └── analytics_handler.go    # 分析 API
│   ├── mqtt/                       # MQTT 模块
│   │   ├── broker.go               # 嵌入式 MQTT Broker 启动
│   │   ├── handler.go              # MQTT 消息处理器（上行数据解析）
│   │   └── client.go               # MQTT 发布器（下行指令）
│   ├── ws/hub.go                   # WebSocket 连接管理和广播
│   ├── middleware/                  # Gin 中间件
│   │   ├── auth.go                 # JWT 认证
│   │   ├── rbac.go                 # 角色权限检查
│   │   └── cors.go                 # 跨域支持
│   └── router/router.go           # 路由注册
└── pkg/                            # 公共工具包
    ├── errcode/errcode.go          # 统一错误码定义
    ├── response/response.go        # 统一 JSON 响应格式
    └── util/
        ├── jwt.go                  # JWT 生成和解析
        └── hash.go                 # bcrypt 密码哈希
```

## 快速开始

### 前置条件

- Go 1.21+
- MySQL 8.0+

### 1. 初始化数据库

```bash
# 创建数据库和表，插入默认角色和管理员账号
mysql -u root -p < migrations/init.sql
```

### 2. 修改配置

编辑 `config/config.yaml`：

```yaml
database:
  host: 127.0.0.1
  port: 3306
  user: root
  password: your_password    # 修改为实际密码
  dbname: ccplatform

mqtt:
  port: 1883                 # MQTT Broker 端口

jwt:
  secret: your-secret-key    # 修改为随机密钥
```

### 3. 编译运行

```bash
# 方式一：使用 Makefile
make run

# 方式二：直接运行
go run ./cmd/server

# 方式三：编译后运行
go build -o server ./cmd/server
./server
```

### 4. 默认账号

| 用户名 | 密码 | 角色 |
|--------|------|------|
| admin | admin123 | 系统管理员 |

## API 接口

### 认证

```
POST   /api/v1/auth/login           # 登录获取 Token
GET    /api/v1/auth/me               # 获取当前用户信息
```

### 电站管理

```
GET    /api/v1/stations              # 电站列表（分页）
POST   /api/v1/stations              # 创建电站
GET    /api/v1/stations/all          # 所有启用电站
GET    /api/v1/stations/:id          # 电站详情
PUT    /api/v1/stations/:id          # 更新电站
DELETE /api/v1/stations/:id          # 删除电站
GET    /api/v1/stations/:id/robots   # 电站下的机器人
GET    /api/v1/stations/:id/realtime # 电站实时监控数据
```

### 机器人管理

```
GET    /api/v1/robots                # 机器人列表（支持按电站/类型/在线状态筛选）
GET    /api/v1/robots/stats          # 机器人统计（总数/在线/离线/故障）
GET    /api/v1/robots/:id            # 机器人详情
POST   /api/v1/robots/:id/cmd        # 下发控制指令（MQTT）
```

### 任务管理

```
GET    /api/v1/tasks                 # 任务列表（支持按电站/机器人/状态筛选）
POST   /api/v1/tasks                 # 创建清扫任务
GET    /api/v1/tasks/:id             # 任务详情
PUT    /api/v1/tasks/:id/status      # 变更任务状态（启动/暂停/完成/取消）
DELETE /api/v1/tasks/:id             # 删除任务
```

### 告警管理

```
GET    /api/v1/alarms                # 告警列表（支持按级别/状态筛选）
GET    /api/v1/alarms/stats          # 告警统计（按级别分组）
GET    /api/v1/alarms/recent         # 最近告警
GET    /api/v1/alarms/:id            # 告警详情
PUT    /api/v1/alarms/:id            # 处理告警（确认/处理/忽略）
```

### 监控中心

```
GET    /api/v1/monitor/overview      # Dashboard 概览数据
```

### 数据分析

```
GET    /api/v1/analytics/economic    # 经济效益报表
GET    /api/v1/analytics/efficiency  # 效率分析报表
GET    /api/v1/analytics/reports     # 运行统计报表
```

### 用户管理

```
GET    /api/v1/users                 # 用户列表
POST   /api/v1/users                 # 创建用户
PUT    /api/v1/users/:id             # 更新用户
DELETE /api/v1/users/:id             # 删除用户
GET    /api/v1/roles                 # 角色列表
POST   /api/v1/roles                 # 创建角色
```

### WebSocket

```
GET    /ws?station_id=xxx            # 实时数据推送（可选 station_id 过滤）
```

推送消息类型：`heartbeat`（心跳）、`position`（位置）、`status`（状态）、`alarm`（告警）

## MQTT 协议

### 上行 Topic（机器人→平台）

| Topic | 说明 | 采集周期 |
|-------|------|---------|
| `tdw/robot/{id}/heartbeat` | 心跳（电量/在线/工作状态） | 5秒 |
| `tdw/robot/{id}/position` | 位置坐标（X/Y/Z/朝向） | 1秒 |
| `tdw/robot/{id}/status` | 运行状态（速度/清扫面积/环境数据） | 10秒 |
| `tdw/robot/{id}/alarm` | 告警信息 | 实时 |

### 下行 Topic（平台→机器人）

| Topic | 说明 |
|-------|------|
| `tdw/robot/{id}/cmd` | 控制指令（启动/停止/回仓/复位） |
| `tdw/robot/{id}/config` | 配置参数下发 |

### 消息格式示例

```json
// 心跳
{"robot_id":"R001","timestamp":1711900000,"battery_level":85,"online_status":1,"work_status":1}

// 位置
{"robot_id":"R001","timestamp":1711900000,"pos_x":123.456,"pos_y":78.901,"pos_z":2.5,"heading":90.0}

// 状态
{"robot_id":"R001","timestamp":1711900000,"work_status":2,"speed":1.5,"clean_area":500.0,"fault_code":0,"temperature":25.5,"humidity":60.0,"light_intensity":80000,"wind_speed":3.2}

// 告警
{"robot_id":"R001","timestamp":1711900000,"alarm_level":3,"alarm_type":"MOTOR_FAULT","alarm_content":"电机过流保护"}
```

## 角色权限

| 角色 | 监控 | 控制 | 任务 | 设备配置 | 告警处理 | 分析 | 系统管理 |
|------|------|------|------|---------|---------|------|---------|
| 系统管理员 | Y | Y | Y | Y | Y | Y | Y |
| 电站管理员 | Y | Y | Y | Y | Y | Y | - |
| 运维工程师 | Y | Y | Y | - | Y | - | - |
| 监控值班员 | Y | Y | Y | - | Y | - | - |
| 普通用户 | Y | - | - | - | - | Y | - |

## 数据模型

### 核心实体关系

```
Station (电站) 1──N Robot (机器人)
    │                │
    │                ├──1──N Task (清扫任务)
    │                │
    │                └──1──N Alarm (告警)
    │
    └──N── User (用户) ──1── Role (角色)
```

### 机器人类型

| 类型 | 适用场景 | 清扫效率 | 通信方式 |
|------|---------|---------|---------|
| 固定式 (1) | 大型集中式电站 | 2000㎡/h | 4G/5G + LoRaWAN |
| 接驳车 (2) | 多阵列大型电站 | 3000㎡/h | 4G/5G + WiFi6 |
| 全智能 (3) | 分布式/复杂地形 | 1500㎡/h | 4G/5G + 星闪 |

## 开发指南

### 添加新的 API 接口

1. 在 `internal/model/` 中定义数据模型
2. 在 `internal/repository/` 中添加数据操作方法
3. 在 `internal/service/` 中实现业务逻辑
4. 在 `internal/handler/` 中编写 HTTP 处理器
5. 在 `internal/router/router.go` 中注册路由

### 添加新的 MQTT Topic 处理

1. 在 `internal/mqtt/handler.go` 中定义消息结构体
2. 在 `OnPublish` 方法的 switch 中添加新的 case
3. 实现对应的 handle 方法（解析→存库→WebSocket推送）

## License

MIT License
