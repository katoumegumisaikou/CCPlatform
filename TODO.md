# TODO

## 监控中心模块 — Redis 缓存层重构

### 当前问题

1. **实时数据走 MySQL 读写**：MQTT handler 收到位置(1Hz)、心跳(5s)、状态(10s) 后直接写 robot 表，GetRealtimeData HTTP 接口再查 robot 表。高频覆盖写入没有持久化意义（历史轨迹已单独写入 robot_positions 表），实时读库浪费数据库资源。
2. **需求文档写了 Redis 但没实现**：需求规格说明书 2.1.3 明确写了缓存数据库 Redis，实际项目完全没有。
3. **WebSocket 全量广播**：MQTT handler 全部使用 BroadcastToAll，Hub 已有 BroadcastToStation 方法但未被调用，所有客户端收到全量数据。

### 改进方案

引入 Redis 作为实时数据中间层，切断 MQTT → MySQL 的直接高频写入链路：

```
机器人 --MQTT--> handler.go
                  ├── 写 Redis 快照 (robot:heartbeat:{id} / robot:realtime:{id} /
                  robot:status:{id})

                  ├── WebSocket 推送到对应电站客户端 (BroadcastToStation)
                  ├── 位置历史 → robot_positions 表 (保留)
                  └── 心跳 → 每 10s 最多写一次 MySQL last_heartbeat

前端首次加载 → HTTP API → 读 Redis 快照 → 返回
                                       └── Redis 不可用 → 降级查 MySQL
后续更新 → WebSocket 增量推送
```

### Redis Key 设计

| Key | 类型 | 字段 | TTL |
|-----|------|------|-----|
| `robot:heartbeat:{id}` | Hash | online_status, battery_level, work_status, last_heartbeat | 30s（过期=离线） |
| `robot:realtime:{id}` | Hash | pos_x, pos_y, pos_z, heading, speed | 15s（过期=位置陈旧） |
| `robot:status:{id}` | Hash | work_status, speed, clean_area, fault_code, temperature, humidity, light_intensity, wind_speed | 30s（过期=状态陈旧） |
| `robot:heartbeat_write:{id}` | Hash | last_write_ts | 60s（心跳写 MySQL 去重） |

### 改造计划

- [ ] 添加 Redis 依赖和配置（go-redis/v9, config/config.yaml, internal/config/config.go）
- [ ] 新增 internal/cache/robot_state.go（Redis 客户端封装，三张 Hash 的存取方法）
- [ ] 改造 MQTT handler：position/status/heartbeat 写 Redis 快照而非 MySQL robot 表，BroadcastToAll 改为 BroadcastToStation
- [ ] 心跳降频写 MySQL：每 10s 最多写一次 last_heartbeat，通过 Redis 标记 key 去重
- [ ] 改造 monitor_service.GetRealtimeData：从 Redis 读快照，Redis 不可用时降级查 MySQL
- [ ] 改造 WebSocket 推送：按电站过滤（BroadcastToStation），不再全量广播

---

## 前端开发规划

### 1. 项目背景

CCPlatform 是光伏清扫机器人云控平台后端，已实现 20+ 功能模块、60+ REST API、MQTT 上行数据处理、WebSocket 实时推送。当前仅有 Go 后端，需要从零构建前端。

### 2. 后端已具备的能力（前端需对接）

**REST API 体系**（13 个路由组，详见 README.md）：
- 认证：JWT 登录、图形验证码、短信验证码、忘记/重置密码
- 电站管理：CRUD + 下属机器人查询 + 实时数据
- 机器人管理：CRUD + 控制指令下发 + 高级统计 + 配置应用
- 任务管理：CRUD + 状态流转（状态机校验）+ 智能调度 + 进度跟踪
- 告警管理：告警列表/处理/统计/趋势 + 告警规则配置
- 监控中心：Dashboard 概览 + 轨迹历史 + 环境数据 + GIS 配置
- 数据分析：经济/效率/运行/对比/时序/导出 6 类报表
- 智能预测：效率预测 + 故障预测 + 策略优化
- 用户/角色/组织：CRUD + 组织架构树
- 系统配置 + 数据字典（类型+字典项）
- 固件管理：版本管理 + OTA 升级下发
- 维护保养：任务管理 + 提醒
- 摄像头管理：录摄像头 + 流地址（RTSP/HLS）
- 通知模板 + 报告模板 + 备份恢复
- 审计日志 + 登录日志

**WebSocket 实时推送**（`/ws?station_id=xxx`）：
- `heartbeat`：在线状态、电量、工作状态（5s 间隔）
- `position`：三维坐标、朝向（1s 间隔）
- `status`：速度、清扫面积、故障码、温湿度/光照/风速（10s 间隔）
- `alarm`：告警实时通知

**RBAC 权限体系**：13 个权限标识，5 种角色（系统管理员/电站管理员/运维工程师/监控值班员/普通用户）

**响应格式**：
- 成功：`{"code":0, "data":..., "message":"ok"}`
- 分页：`{"code":0, "data":{"list":[...], "total":100}, "message":"ok"}`
- 错误：`{"code":10001, "message":"error description"}`

### 3. 核心页面清单

| 页面 | 路由 | 关键要求 |
|------|------|---------|
| 登录页 | `/login` | JWT 登录、图形验证码、记住密码 |
| 首页仪表盘 | `/dashboard` | 电站/机器人/任务/告警统计卡片、清扫面积趋势图、机器人状态分布饼图、实时在线率/出仓率 |
| **监控中心** | `/monitor` | **核心页面**：左 GIS 地图 + 右信息面板。地图展示电站位置、光伏板阵列、机器人实时位置（WebSocket 驱动）；支持卫星/矢量切换、3D/2D 切换、缩放测距；点击机器人显示详情卡片（基本信息+实时位置+运行数据+环境参数+远程控制按钮）；支持轨迹回放 |
| 电站管理 | `/stations` | 表格+分页+搜索、新建/编辑弹窗、电站下机器人列表 |
| 机器人管理 | `/robots` | 表格（按电站/类型/在线状态筛选）、详情页（基本信息+实时数据+控制面板）、指令下发（start/stop/return/reset） |
| 任务管理 | `/tasks` | 表格（按电站/机器人/状态筛选）、创建任务表单（选机器人+定时/手动/周期）、智能调度一键创建、进度条展示 |
| 告警中心 | `/alarms` | 告警列表（按级别/状态筛选）、处理操作（确认/处理/忽略）、告警规则配置（作用域+抑制+升级）、趋势图表 |
| 数据分析 | `/analytics` | 经济效益/效率分析/运行统计/对比分析/时序分析 5 个子页，图表+表格+导出按钮 |
| 智能预测 | `/predictions` | 效率预测曲线、故障预测列表、策略优化建议 |
| 固件管理 | `/firmwares` | 固件列表+上传、OTA 升级任务下发、升级进度跟踪 |
| 维护保养 | `/maintenance` | 维护任务表格、保养提醒列表 |
| 摄像头管理 | `/cameras` | 摄像头列表、添加/编辑（含流地址 RTSP/HLS 输入） |
| 系统管理 | `/system` | 用户管理、角色管理、组织架构树、系统配置、数据字典、审计日志、登录日志、通知模板、报告模板、备份恢复 |
| 个人中心 | `/profile` | 修改密码、个人信息 |

### 4. 关键技术难点

1. **GIS 地图**：需要集成地图 SDK（Leaflet/高德/百度/Mapbox），实现：
   - 电站标记点 + 光伏板阵列多边形覆盖物
   - 机器人实时位置图标（WebSocket 驱动，1Hz 刷新）
   - 地图图层切换（卫星图/矢量图）、3D 视角（Cesium.js 或 Mapbox GL JS）
   - 轨迹回放（播放历史位置点，支持倍速和拖拽）

2. **实时数据**：WebSocket 连接管理 + 心跳检测 + 断线重连 + 按 station_id 过滤订阅

3. **视频监控**：浏览器无法直接播放 RTSP，需要：
   - 方案 A：后端转码 RTSP → WebRTC/HLS/FLV（需新增后端视频流代理服务）
   - 方案 B：摄像头本身支持 HLS/WebRTC 输出
   - 前端集成播放器（flv.js / video.js / Jessibuca）

4. **大数据量表格**：机器人位置历史、审计日志等表可能有海量数据，需要虚拟滚动

5. **权限控制**：前端路由守卫 + 按钮级权限（基于 RBAC 权限标识）

6. **任务下发流程缺失**：当前任务创建只写 MySQL，未通过 MQTT 下发给机器人。前端需知晓此限制，创建任务后需额外调用 `POST /robots/:id/cmd` 发 start 指令

### 5. 请你规划以下内容

请基于以上信息，为 CCPlatform 前端项目输出一份完整的编码规划，包含：

1. **技术选型**：框架（React/Vue/Angular？）、UI 组件库、GIS 地图库、图表库、视频播放方案、状态管理、构建工具，并说明选型理由
2. **项目目录结构**：pages/components/hooks/services/store/utils 等如何组织
3. **路由设计**：完整路由表（路径+权限守卫+懒加载）
4. **组件树设计**：全局布局（侧边栏+顶栏+内容区）、核心页面（监控中心为主）的组件拆分
5. **数据流设计**：API 请求封装（axios/fetch 拦截器+JWT 注入）、WebSocket 连接管理（单例+自动重连+消息分发）、状态管理（全局状态 vs 页面状态）
6. **开发阶段划分**：建议按什么顺序开发（如 Phase1 登录+布局+仪表盘 → Phase2 监控中心 → Phase3 业务 CRUD → Phase4 数据分析 → Phase5 系统管理 → Phase6 高级功能）
7. **关键技术难点的实现方案**（GIS 地图、视频监控、实时数据推送、大数据量表等）
8. **与后端的对接约定**：API base URL 配置、请求/响应类型定义、错误码处理、WebSocket 消息类型定义

注意：
- 本平台用于光伏电站运维管理，用户以 PC Web 管理端为主，可兼顾大屏展示
- 监控中心是核心差异化页面，需重点设计
- 摄像头 RTSP 流播放是已知技术难点，请给出可行方案（含后端需配合的部分）
- 建议在项目根目录新建 `frontend/` 目录，与现有 Go 后端代码隔离
