# TDW CCPlatform - 拓达威光伏清扫机器人云控平台

基于 Go 语言开发的光伏清扫机器人云端管理与控制平台后端服务。

## 项目概述

本平台用于统一管理光伏电站中的清扫机器人，支持多电站接入、实时监控、任务调度、告警管理、数据分析和运维管理。机器人通过 MQTT 协议上报数据（心跳、位置、状态、告警），平台将数据持久化到 MySQL 并通过 WebSocket 实时推送给前端。

## 核心功能

| 模块 | 功能 |
|------|------|
| 设备管理 | 电站管理、机器人管理、摄像头管理、设备凭证管理 |
| 监控中心 | Dashboard 概览、实时数据、GIS 地图配置、轨迹回放、环境数据 |
| 任务调度 | 清扫任务 CRUD、智能排班、周期定时调度、进度跟踪 |
| 告警中心 | 告警记录/处理/趋势分析、告警规则配置 |
| 通知管理 | 通知模板管理（短信/邮件/App推送） |
| 数据分析 | 经济效益、效率分析、运行统计、对比报表、时序分析、数据导出 |
| 智能预测 | 效率预测、故障预测、策略优化 |
| 固件管理 | 固件版本管理、OTA 升级任务下发 |
| 维护保养 | 维护任务管理、保养提醒 |
| 权限管理 | 用户/角色 CRUD、JWT 认证、RBAC 权限控制、登录日志 |
| 系统管理 | 系统配置、数据字典、组织架构树、操作审计日志 |
| 备份恢复 | 数据库备份、恢复 |
| 报告模板 | 报告模板管理与报表导出 |

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
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────────────┐  │
│  │ 固件/维护服务 │ │ 通知/审计服务 │ │ 预测/备份/报告服务   │  │
│  └──────────────┘ └──────────────┘ └──────────────────────┘  │
│  ┌──────────────────┐ ┌──────────────────┐                   │
│  │ JWT 认证中间件    │ │ RBAC 权限中间件   │                   │
│  └──────────────────┘ └──────────────────┘                   │
│  ┌──────────────────┐                                        │
│  │ 审计中间件        │                                        │
│  └──────────────────┘                                        │
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
│  │ 25+ 业务表        │  │ 自动迁移 + 种子  │                  │
│  └──────────────────┘  └──────────────────┘                  │
└──────────────────────────────────────────────────────────────┘
```

## 技术栈

| 组件 | 技术选型 | 用途 |
|------|---------|------|
| 语言 | Go 1.21+ | 后端服务 |
| Web 框架 | Gin | HTTP REST API |
| 数据库 | MySQL 8.0 | 业务数据存储 |
| ORM | GORM | 数据库操作与自动迁移 |
| MQTT | mochi-mqtt | 嵌入式 MQTT Broker |
| WebSocket | gorilla/websocket | 实时数据推送 |
| 认证 | JWT (golang-jwt/v5) | 用户认证 |
| 密码 | bcrypt | 密码哈希 |
| 配置 | Viper | 配置文件管理 |
| 定时任务 | robfig/cron v3 | 周期清扫任务调度 |

## 项目结构

```
CCPlatform/
├── cmd/server/main.go              # 程序入口，启动流程:
│                                   #   配置→DB→种子数据→WebSocket→MQTT→Hook
│                                   #   →调度器→Publisher→HTTP Server→优雅关闭
├── config/config.yaml              # 配置文件（数据库、MQTT、JWT、HTTP等）
├── migrations/init.sql             # MySQL 建表和初始化数据脚本
├── Makefile                        # 构建/运行/测试/数据库初始化
├── internal/
│   ├── config/config.go            # 配置结构体定义和加载
│   ├── model/                      # 数据模型（GORM，25张表）
│   │   ├── station.go              # 光伏电站
│   │   ├── robot.go                # 清扫机器人
│   │   ├── robot_config.go         # 机器人配置参数
│   │   ├── robot_position.go       # 机器人位置历史
│   │   ├── task.go                 # 清扫任务
│   │   ├── alarm.go                # 告警记录
│   │   ├── alarm_rule.go           # 告警规则
│   │   ├── user.go                 # 系统用户
│   │   ├── role.go                 # 角色权限
│   │   ├── organization.go         # 组织架构
│   │   ├── firmware.go             # 固件版本
│   │   ├── maintenance.go          # 维护保养
│   │   ├── camera.go               # 摄像头
│   │   ├── dict.go                 # 数据字典
│   │   ├── system_config.go        # 系统配置
│   │   ├── audit_log.go            # 操作审计日志
│   │   ├── login_log.go            # 登录日志
│   │   ├── notify_template.go      # 通知模板
│   │   ├── report_template.go      # 报告模板
│   │   ├── device_credential.go    # 设备凭证
│   │   ├── cleaning_record.go      # 清扫记录
│   │   ├── environment_data.go     # 环境数据
│   │   ├── firmware.go             # 固件信息 (含 upgrade_record)
│   │   └── password_reset.go       # 密码重置
│   ├── repository/                 # 数据访问层（24个Repository）
│   │   ├── db.go                   # 数据库连接初始化
│   │   ├── station_repo.go         # 电站数据操作
│   │   ├── robot_repo.go           # 机器人数据操作
│   │   ├── task_repo.go            # 任务数据操作
│   │   ├── alarm_repo.go           # 告警数据操作
│   │   ├── alarm_rule_repo.go      # 告警规则数据操作
│   │   ├── user_repo.go            # 用户/角色数据操作
│   │   ├── organization_repo.go    # 组织架构数据操作
│   │   ├── firmware_repo.go        # 固件数据操作
│   │   ├── maintenance_repo.go     # 维护保养数据操作
│   │   ├── camera_repo.go          # 摄像头数据操作
│   │   ├── dict_repo.go            # 字典数据操作
│   │   ├── system_config_repo.go   # 系统配置数据操作
│   │   ├── audit_log_repo.go       # 审计日志数据操作
│   │   ├── login_log_repo.go       # 登录日志数据操作
│   │   ├── notify_template_repo.go # 通知模板数据操作
│   │   ├── report_template_repo.go # 报告模板数据操作
│   │   ├── device_credential_repo.go # 设备凭证数据操作
│   │   ├── cleaning_record_repo.go   # 清扫记录数据操作
│   │   ├── environment_data_repo.go  # 环境数据数据操作
│   │   ├── robot_position_repo.go    # 位置历史数据操作
│   │   ├── robot_config_repo.go      # 机器人配置数据操作
│   │   ├── password_reset_repo.go    # 密码重置数据操作
│   │   └── upgrade_record_repo.go    # 升级记录数据操作
│   ├── service/                    # 业务逻辑层（22个Service）
│   │   ├── station_service.go      # 电站业务
│   │   ├── robot_service.go        # 机器人业务
│   │   ├── task_service.go         # 任务业务（CRUD+状态流转+智能排班）
│   │   ├── alarm_service.go        # 告警业务
│   │   ├── alarm_rule_service.go   # 告警规则业务
│   │   ├── user_service.go         # 用户业务（登录+RBAC+验证码+密码重置）
│   │   ├── captcha_service.go      # 图形验证码服务
│   │   ├── monitor_service.go      # 监控中心（Dashboard+实时数据+轨迹+环境）
│   │   ├── analytics_service.go    # 数据分析（经济+效率+报表+对比+时序+导出）
│   │   ├── organization_service.go # 组织架构业务
│   │   ├── firmware_service.go     # 固件管理业务
│   │   ├── maintenance_service.go  # 维护保养业务
│   │   ├── camera_service.go       # 摄像头业务
│   │   ├── dict_service.go         # 数据字典业务
│   │   ├── system_config_service.go# 系统配置业务
│   │   ├── notify_service.go       # 通知模板业务
│   │   ├── prediction_service.go   # 智能预测（效率/故障/策略优化）
│   │   ├── report_template_service.go # 报告模板业务
│   │   ├── backup_service.go       # 数据库备份恢复
│   │   └── robot_config_service.go # 机器人配置业务
│   ├── handler/                    # HTTP 处理器（20个Handler）
│   │   ├── station_handler.go      # 电站 API
│   │   ├── robot_handler.go        # 机器人 API
│   │   ├── task_handler.go         # 任务 API
│   │   ├── alarm_handler.go        # 告警 API
│   │   ├── alarm_rule_handler.go   # 告警规则 API
│   │   ├── user_handler.go         # 用户/角色/验证码 API
│   │   ├── monitor_handler.go      # 监控 API
│   │   ├── analytics_handler.go    # 分析 API
│   │   ├── organization_handler.go # 组织架构 API
│   │   ├── firmware_handler.go     # 固件 API
│   │   ├── maintenance_handler.go  # 维护保养 API
│   │   ├── camera_handler.go       # 摄像头 API
│   │   ├── dict_handler.go         # 字典 API
│   │   ├── system_config_handler.go# 系统配置 API
│   │   ├── audit_handler.go        # 审计日志 API
│   │   ├── notify_handler.go       # 通知模板 API
│   │   ├── prediction_handler.go   # 智能预测 API
│   │   ├── report_template_handler.go # 报告模板 API
│   │   ├── backup_handler.go       # 备份恢复 API
│   │   └── robot_config_handler.go # 机器人配置 API
│   ├── mqtt/                       # MQTT 模块
│   │   ├── broker.go               # 嵌入式 MQTT Broker 启动
│   │   ├── handler.go              # MQTT 消息处理器（上行数据解析）
│   │   └── client.go               # MQTT 发布器（下行指令）
│   ├── ws/hub.go                   # WebSocket 连接管理和广播
│   ├── scheduler/scheduler.go      # 周期任务调度器 (robfig/cron)
│   ├── middleware/                  # Gin 中间件
│   │   ├── auth.go                 # JWT 认证
│   │   ├── rbac.go                 # 角色权限检查
│   │   ├── audit.go                # 操作审计记录
│   │   └── cors.go                 # 跨域支持
│   └── router/                     # 路由注册
│       ├── router.go               # 主路由
│       ├── auth.go                 # 认证子路由
│       ├── station.go              # 电站子路由
│       ├── robot.go                # 机器人和配置子路由
│       ├── task.go                 # 任务子路由
│       └── alarm.go                # 告警和规则子路由
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

# 或使用 Makefile
make db-create
make db-init
```

### 2. 修改配置

编辑 `config/config.yaml`：

```yaml
server:
  port: 8080                    # HTTP 服务端口

database:
  host: 127.0.0.1
  port: 3306
  user: root
  password: your_password       # 修改为实际密码
  dbname: ccplatform

mqtt:
  port: 1883                    # MQTT Broker 端口

jwt:
  secret: your-secret-key       # 修改为随机密钥
```

### 3. 编译运行

```bash
# 开发模式（直接运行）
make dev

# 构建并运行
make run

# 编译
make build

# 运行测试
make test

# 整理依赖
make tidy
```

### 4. 启动流程

程序启动后依次执行：
1. 加载配置文件 (config/config.yaml)
2. 连接 MySQL 并自动迁移 25 张业务表
3. 初始化 5 种默认角色和 admin 管理员账号
4. 启动 WebSocket Hub（实时推送）
5. 启动嵌入式 MQTT Broker（端口 1883）
6. 注册 MQTT 消息处理器（心跳/位置/状态/告警）
7. 启动周期任务调度器（定时清扫）
8. 注册 HTTP 路由并启动服务（端口 8080）

### 5. 默认账号

| 用户名 | 密码 | 角色 |
|--------|------|------|
| admin | admin123 | 系统管理员 |

## API 接口

### 认证 (`/api/v1/auth`) - 公开接口

```
POST   /auth/login               # 登录获取 Token
GET    /auth/captcha              # 获取图形验证码
POST   /auth/verify-captcha       # 校验图形验证码
POST   /auth/sms-code             # 发送短信验证码
POST   /auth/verify-sms           # 校验短信验证码
POST   /auth/forgot-password      # 忘记密码
POST   /auth/reset-password       # 重置密码
```

### 认证 (`/api/v1/auth`) - 需要 Token

```
GET    /auth/me                   # 获取当前用户信息
```

### 电站管理 (`/api/v1/stations`)

```
GET    /stations                  # 电站列表（分页）
POST   /stations                  # 创建电站 [RBAC: station:create]
GET    /stations/all              # 所有启用电站
GET    /stations/:id              # 电站详情
PUT    /stations/:id              # 更新电站 [RBAC: station:update]
DELETE /stations/:id              # 删除电站 [RBAC: station:delete]
GET    /stations/:id/robots       # 电站下的机器人
GET    /stations/:id/realtime     # 电站实时监控数据
```

### 机器人管理 (`/api/v1/robots`)

```
GET    /robots                    # 机器人列表（支持按电站/类型/在线状态筛选）
POST   /robots                    # 创建机器人 [RBAC: robot:create]
GET    /robots/stats              # 机器人统计总数
PUT    /robots/:id                # 更新机器人 [RBAC: robot:update]
DELETE /robots/:id                # 删除机器人 [RBAC: robot:delete]
GET    /robots/:id                # 机器人详情
POST   /robots/:id/cmd            # 下发控制指令 [RBAC: robot:control]
GET    /robots/:id/advanced-stats # 机器人高级统计
POST   /robots/:id/config/apply   # 应用机器人配置 [RBAC: robot:config]
```

### 任务管理 (`/api/v1/tasks`)

```
GET    /tasks                     # 任务列表（支持按电站/机器人/状态筛选）
POST   /tasks                     # 创建清扫任务 [RBAC: task:create]
POST   /tasks/smart-schedule      # 智能排班 [RBAC: task:create]
GET    /tasks/:id                 # 任务详情
GET    /tasks/:id/progress        # 任务进度跟踪
PUT    /tasks/:id/status          # 变更任务状态 [RBAC: task:update]
DELETE /tasks/:id                 # 删除任务 [RBAC: task:delete]
```

### 告警管理 (`/api/v1/alarms`)

```
GET    /alarms                    # 告警列表（支持按级别/状态筛选）
GET    /alarms/stats              # 告警统计（按级别分组）
GET    /alarms/recent             # 最近告警
GET    /alarms/trends             # 告警趋势分析
GET    /alarms/:id                # 告警详情
PUT    /alarms/:id                # 处理告警（确认/处理/忽略） [RBAC: alarm:handle]
```

### 监控中心 (`/api/v1/monitor`)

```
GET    /monitor/overview          # Dashboard 概览数据
GET    /monitor/tracks            # 机器人轨迹历史 [RBAC: monitor:view]
GET    /monitor/environment       # 环境数据历史 [RBAC: monitor:view]
GET    /monitor/gis-config        # GIS 地图配置
PUT    /monitor/gis-config        # 更新 GIS 配置 [RBAC: system:config]
```

### 数据分析 (`/api/v1/analytics`)

```
GET    /analytics/economic        # 经济效益报表
GET    /analytics/efficiency      # 效率分析报表
GET    /analytics/reports         # 运行统计报表
GET    /analytics/compare         # 对比分析报表
GET    /analytics/export          # 导出报表数据
GET    /analytics/timeseries      # 时序分析数据
```

### 智能预测 (`/api/v1/analytics/predictions`)

```
GET    /analytics/predictions/efficiency   # 效率预测
GET    /analytics/predictions/fault        # 故障预测
GET    /analytics/strategy/optimize        # 策略优化
```

### 用户管理 (`/api/v1/users`)

```
GET    /users                     # 用户列表 [RBAC: user:view]
POST   /users                     # 创建用户 [RBAC: user:create]
PUT    /users/:id                 # 更新用户 [RBAC: user:update]
DELETE /users/:id                 # 删除用户 [RBAC: user:delete]
```

### 角色管理 (`/api/v1/roles`)

```
GET    /roles                     # 角色列表
POST   /roles                     # 创建角色 [RBAC: role:create]
PUT    /roles/:id                 # 更新角色 [RBAC: role:update]
DELETE /roles/:id                 # 删除角色 [RBAC: role:delete]
```

### 组织架构 (`/api/v1/organizations`)

```
GET    /organizations             # 组织列表
GET    /organizations/tree        # 组织架构树
GET    /organizations/:id         # 组织详情
POST   /organizations             # 创建组织 [RBAC: org:create]
PUT    /organizations/:id         # 更新组织 [RBAC: org:update]
DELETE /organizations/:id         # 删除组织 [RBAC: org:delete]
```

### 系统配置 (`/api/v1/system/configs`)

```
GET    /system/configs            # 配置列表
GET    /system/configs/:key       # 获取配置项
PUT    /system/configs/:key       # 设置配置项 [RBAC: system:config]
DELETE /system/configs/:key       # 删除配置项 [RBAC: system:config]
```

### 数据字典 (`/api/v1/dicts`)

```
GET    /dicts                     # 字典类型列表
POST   /dicts                     # 创建字典类型 [RBAC: system:config]
PUT    /dicts/:id                 # 更新字典类型 [RBAC: system:config]
DELETE /dicts/:id                 # 删除字典类型 [RBAC: system:config]
GET    /dicts/:type/items         # 字典项列表
POST   /dicts/:type/items         # 创建字典项 [RBAC: system:config]
PUT    /dicts/items/:id           # 更新字典项 [RBAC: system:config]
DELETE /dicts/items/:id           # 删除字典项 [RBAC: system:config]
```

### 固件管理 (`/api/v1/firmwares`)

```
GET    /firmwares                 # 固件列表
POST   /firmwares                 # 上传固件 [RBAC: firmware:create]
PUT    /firmwares/:id             # 更新固件信息 [RBAC: firmware:update]
DELETE /firmwares/:id             # 删除固件 [RBAC: firmware:delete]
POST   /firmwares/:id/upgrade     # 下发升级任务 [RBAC: firmware:upgrade]
```

### 维护保养 (`/api/v1/maintenance`)

```
GET    /maintenance               # 维护任务列表
POST   /maintenance               # 创建维护任务 [RBAC: maintenance:create]
PUT    /maintenance/:id           # 更新维护任务 [RBAC: maintenance:update]
DELETE /maintenance/:id           # 删除维护任务 [RBAC: maintenance:delete]
GET    /maintenance/reminders     # 保养提醒
```

### 告警规则 (`/api/v1/alarm-rules`)

```
GET    /alarm-rules               # 告警规则列表
POST   /alarm-rules               # 创建告警规则 [RBAC: alarm:config]
PUT    /alarm-rules/:id           # 更新告警规则 [RBAC: alarm:config]
DELETE /alarm-rules/:id           # 删除告警规则 [RBAC: alarm:config]
```

### 机器人配置 (`/api/v1/robot-configs`)

```
GET    /robot-configs             # 机器人配置列表
POST   /robot-configs             # 创建配置模板 [RBAC: robot:config]
PUT    /robot-configs/:id         # 更新配置模板 [RBAC: robot:config]
DELETE /robot-configs/:id         # 删除配置模板 [RBAC: robot:config]
```

### 通知模板 (`/api/v1/notify/templates`)

```
GET    /notify/templates          # 通知模板列表
POST   /notify/templates          # 创建通知模板 [RBAC: notify:config]
PUT    /notify/templates/:id      # 更新通知模板 [RBAC: notify:config]
DELETE /notify/templates/:id      # 删除通知模板 [RBAC: notify:config]
```

### 摄像头 (`/api/v1/cameras`)

```
GET    /cameras                   # 摄像头列表
POST   /cameras                   # 添加摄像头 [RBAC: camera:create]
PUT    /cameras/:id               # 更新摄像头 [RBAC: camera:update]
DELETE /cameras/:id               # 删除摄像头 [RBAC: camera:delete]
```

### 报告模板 (`/api/v1/report-templates`)

```
GET    /report-templates          # 报告模板列表
POST   /report-templates          # 创建报告模板 [RBAC: system:config]
PUT    /report-templates/:id      # 更新报告模板 [RBAC: system:config]
DELETE /report-templates/:id      # 删除报告模板 [RBAC: system:config]
```

### 备份恢复 (`/api/v1/backups`)

```
GET    /backups                   # 备份列表 [RBAC: system:config]
POST   /backups                   # 创建备份 [RBAC: system:config]
DELETE /backups/:filename         # 删除备份 [RBAC: system:config]
POST   /backups/:filename/restore # 恢复备份 [RBAC: system:config]
```

### 审计与日志

```
GET    /api/v1/audit-logs         # 操作审计日志列表
GET    /api/v1/login-logs         # 登录日志列表
```

### WebSocket

```
GET    /ws?station_id=xxx         # 实时数据推送（可选 station_id 过滤）
```

推送消息类型：`heartbeat`（心跳）、`position`（位置）、`status`（状态）、`alarm`（告警）

## MQTT 协议

### 上行 Topic（机器人→平台）

| Topic | 说明 | 周期 |
|-------|------|------|
| `tdw/robot/{id}/heartbeat` | 心跳（电量/在线/工作状态） | 5s |
| `tdw/robot/{id}/position` | 位置坐标（X/Y/Z/朝向） | 1s |
| `tdw/robot/{id}/status` | 运行状态（速度/清扫面积/环境数据） | 10s |
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

| 权限 | 系统管理员 | 电站管理员 | 运维工程师 | 监控值班员 | 普通用户 |
|------|:---------:|:---------:|:---------:|:---------:|:-------:|
| 监控看板 | Y | Y | Y | Y | Y |
| 机器人控制 | Y | Y | Y | Y | - |
| 任务管理 | Y | Y | Y | Y | - |
| 告警处理 | Y | Y | Y | Y | - |
| 设备配置 | Y | Y | - | - | - |
| 数据分析 | Y | Y | - | - | Y |
| 固件升级 | Y | Y | Y | - | - |
| 维护管理 | Y | Y | Y | - | - |
| 系统管理 | Y | - | - | - | - |

内置 RBAC 权限标识：`station:*`, `robot:*`, `task:*`, `alarm:*`, `user:*`, `role:*`, `org:*`, `firmware:*`, `maintenance:*`, `camera:*`, `system:config`, `notify:config`, `monitor:view`

## 数据模型

### 核心实体关系

```
Station (电站) 1──N Robot (机器人)
    │                │
    │                ├──1──N Task (清扫任务)
    │                ├──1──N Alarm (告警)
    │                ├──1──N CleaningRecord (清扫记录)
    │                ├──1──N RobotPosition (位置历史)
    │                └──1──N RobotConfig (配置参数)
    │
    ├──N── Camera (摄像头)
    ├──N── Maintenance (维护任务)
    ├──1──N EnvironmentData (环境数据)
    │
    User (用户) ──1── Role (角色)
    │
    Organization (组织架构) 1──N Station (电站)
```

### 机器人类型

| 类型 | 适用场景 | 清扫效率 | 通信方式 |
|------|---------|---------|---------|
| 固定式 (1) | 大型集中式电站 | 2000㎡/h | 4G/5G + LoRaWAN |
| 接驳车 (2) | 多阵列大型电站 | 3000㎡/h | 4G/5G + WiFi6 |
| 全智能 (3) | 分布式/复杂地形 | 1500㎡/h | 4G/5G + 星闪 |

## 开发指南

### 添加新的功能模块

1. 在 `internal/model/` 中定义 GORM 数据模型
2. 在 `main.go` 的 `AutoMigrate` 中添加模型引用
3. 在 `internal/repository/` 中添加数据访问层
4. 在 `internal/service/` 中实现业务逻辑
5. 在 `internal/handler/` 中编写 HTTP 处理器
6. 在 `internal/router/` 中注册路由

### 添加新的 MQTT Topic 处理

1. 在 `internal/mqtt/handler.go` 中定义消息结构体
2. 在 `OnPublish` 方法的 switch 中添加新 case
3. 实现对应的 handle 方法（解析→存库→WebSocket 推送）

## License

MIT License
