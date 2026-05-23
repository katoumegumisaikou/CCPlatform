# 前后端 REST API 数据交互验证

> 仅验证 REST API 请求/响应，不含 MQTT/WebSocket。按模块逐一验证：前端 types → api 调用 → 表单字段名 → 后端 handler request struct 的 json tag 对齐。

## 验证方法

每个模块验证 4 步：
1. `types/index.ts` 类型定义 ↔ 后端 model 字段 (json tag)
2. `api/{模块}.ts` HTTP method + URL + params ↔ 后端 router 注册
3. `pages/{模块}/index.tsx` 表单 name/dataIndex ↔ 后端 request struct json tag
4. 实际调接口，打开浏览器 DevTools Network 面板，检查 request payload 和 response

---

## 0. API 响应解包基线（验证前必读）

```
axios interceptor (client.ts) 行为：
  后端返回 {code:0, data:X, message:"success"}
  → interceptor 返回 X（即 response.data.data）
  → 所有页面代码拿到的是 X，不是完整响应体

三类接口的解包结果：
  分页接口 (list/total/page/size)  → X = {list:[...], total:N}
  数组接口 (/stations/all, /roles)  → X = [...]
  单对象接口 (GET /:id, POST/PUT)   → X = {...}

关键检查：所有页面代码不得再访问 .data 属性（之前 System 页面修过一批）
```

---

## 1. 登录 & 认证

```
文件：Login/index.tsx, api/auth.ts, types/index.ts
后端：user_handler.go LoginRequest/VerifyCaptchaRequest

验证项：
- [ ] 登录 form: username, password 字段名 √（已验证）
- [ ] 验证码: captcha_id, answer 字段名匹配 VerifyCaptchaRequest
- [ ] token 存储 key "token" vs getToken() 读取一致
- [ ] code=10001/HTTP 401 → 清除 token + 跳转 /login
- [ ] GET /auth/me 返回 User 字段 vs types/index.ts User 接口
- [ ] authStore 权限列表来自 role.permissions 字段
```

## 2. 电站管理 — Stations

```
文件：Stations/index.tsx, api/stations.ts, types/index.ts Station
后端：station_handler.go CreateStationRequest/UpdateStationRequest

REST 对齐：
  GET    /stations       → stationApi.list(params)       → 分页列表
  POST   /stations       → stationApi.create(data)       → 创建
  GET    /stations/all   → stationApi.getAll()           → 全部电站(数组)
  GET    /stations/:id   → stationApi.getById(id)        → 单个电站
  PUT    /stations/:id   → stationApi.update(id, data)   → 更新
  DELETE /stations/:id   → stationApi.delete(id)         → 删除
  GET    /stations/:id/robots → stationApi.getRobots(id) → 电站下属机器人

表单字段 vs CreateStationRequest(json tag)：
  station_name √, station_code √, location √, capacity √
  longitude √ (InputNumber 无 stringMode，发 number), latitude √
  panel_area √, panel_count √

⚠ BUG：表格列 dataIndex: 'address' → 应为 'location'（第139行）
  后端返回字段是 location，Station 类型定义也是 location，address 列始终为空

筛选参数：station_name 作为 query param → GET /stations?station_name=xxx
- [ ] 确认后端支持 station_name 过滤
```

## 3. 机器人管理 — Robots

```
文件：Robots/index.tsx, api/robots.ts, types/index.ts Robot
后端：robot_handler.go CreateRobotRequest/UpdateRobotRequest/SendCommandRequest

REST 对齐：
  GET    /robots              → robotApi.list(params)           → 分页
  POST   /robots              → robotApi.create(data)            → 创建
  GET    /robots/stats        → robotApi.stats()                 → 统计
  GET    /robots/:id          → robotApi.getById(id)             → 详情
  PUT    /robots/:id          → robotApi.update(id, data)        → 更新
  DELETE /robots/:id          → robotApi.delete(id)              → 删除
  POST   /robots/:id/cmd      → robotApi.sendCommand(id,cmd)     → 指令
  GET    /robots/:id/advanced-stats → robotApi.advancedStats(id) → 高级统计
  POST   /robots/:id/config/apply   → robotApi.applyConfig(id,configId) → 配置

表单 vs CreateRobotRequest：
  robot_code √, robot_name √, robot_type √ (Select value=Number), station_id √

指令下发 vs SendCommandRequest：
  robotApi.sendCommand(id, cmd, params) → body: { cmd, params } √

筛选参数：station_id, robot_type, online_status → 后端支持
```

## 4. 任务管理 — Tasks

```
文件：Tasks/index.tsx, api/tasks.ts, types/index.ts Task
后端：task_handler.go CreateTaskRequest/UpdateTaskStatusRequest/SmartScheduleRequest

REST 对齐：
  GET    /tasks              → taskApi.list(params)          → 分页
  POST   /tasks              → taskApi.create(data)          → 创建
  GET    /tasks/:id          → taskApi.getById(id)           → 详情
  PUT    /tasks/:id/status   → taskApi.updateStatus(id,status) → 状态更新
  DELETE /tasks/:id          → taskApi.delete(id)            → 删除
  GET    /tasks/:id/progress → taskApi.getProgress(id)       → 进度
  POST   /tasks/smart-schedule → taskApi.smartSchedule(stationId) → 智能排程

表单 vs CreateTaskRequest：
  task_name √, task_type √, robot_id √, area_ids √
  plan_start √ (dayjs → "YYYY-MM-DD HH:mm:ss" 字符串), plan_end √
  cron_expr √
  task_type=2 → plan_start/plan_end 必填 √
  task_type=3 → cron_expr 必填 √

状态更新 vs UpdateTaskStatusRequest(body: {status})：
  taskApi.updateStatus(id, status) → 发送 {status} √

智能排程 vs SmartScheduleRequest(body: {station_id}) √

状态机逻辑对照：待执行0→启动1, 执行中1→暂停3/完成2, 已暂停3→恢复1/完成2
- [ ] 前端按钮展示规则与后端 TaskService 状态机一致
```

## 5. 仪表盘 — Dashboard

```
文件：Dashboard/index.tsx, api/monitor.ts monitorApi.getDashboard()
后端：GET /monitor/overview → Monitor.GetDashboard

返回字段对照（DashboardOverview）：
  station_count √, today_clean_area √, running_tasks √, unhandled_alarms √
  robot_stats: 后端返回 {"online":0,"offline":0,"fault":0,"total":0} (字符串 key)
    - 统计卡片取值 data?.robot_stats?.['1'] ?? data?.robot_stats?.online
      数字 key '1'=undefined → fallback 'online' 能拿到值，但所有状态都映射成"在线"
    - 实际应统一用字符串 key 或改取 total 字段
  alarm_stats: 后端返回 {"total":0} 或其他 key

- [ ] robot_stats 统计卡片：当前只显示"在线"数量正确吗？
- [ ] 清扫面积趋势图是 mock 数据，后续替换为 GET /analytics/timeseries

---

## 已验证 & 已修复

| # | 文件 | 行 | 问题 | 状态 |
|---|------|----|------|------|
| 1 | Stations/index.tsx | 139 | `dataIndex: 'address'` vs 后端字段 `location` 不匹配 | ✅ 已修复 |
| 2 | Analytics/index.tsx | 84,119,170,196,224 | `(res as any)?.data ?? res` API 已解包但再次 .data | ✅ 已修复 — 直接使用 res |
| 3 | Stations (上一轮) | form | longitude/latitude stringMode, address→location, 缺panel_area/panel_count | ✅ 已修复 |
| 4 | System (上一轮) | 163,567,1259,1359 | roles/orgs .data 访问, Tree onSelect 类型, 审计/登录日志 .data | ✅ 已修复 |
| 5 | useWebSocket (上一轮) | 10 | useRef 泛型缺初始值 | ✅ 已修复 |
| 6 | Monitor (上一轮) | Divider | orientation="start" 非 antd v6 合法 prop | ✅ 已修复 |
```

## 6. 告警中心 — Alarms

```
文件：Alarms/index.tsx, api/alarms.ts, types/index.ts Alarm/AlarmRule
后端：alarm_handler.go HandleAlarmRequest，alarm_rule_handler.go (绑 model.AlarmRule)

告警记录：
  GET    /alarms         → alarmApi.list(params)           → 分页
  GET    /alarms/:id     → alarmApi.getById(id)            → 详情
  PUT    /alarms/:id     → alarmApi.handle(id,{handle_status,handle_remark}) → 处理
  GET    /alarms/stats   → alarmApi.stats()                → 统计(Record<string,number>)
  GET    /alarms/recent  → alarmApi.recent()               → 最近告警(Alarm[])
  GET    /alarms/trends  → alarmApi.trends(params)         → 趋势

告警规则：
  GET    /alarm-rules     → alarmApi.rules.list(params)    → 分页
  POST   /alarm-rules     → alarmApi.rules.create(data)    → 创建
  PUT    /alarm-rules/:id → alarmApi.rules.update(id,data) → 更新
  DELETE /alarm-rules/:id → alarmApi.rules.delete(id)      → 删除

处理弹窗 vs HandleAlarmRequest：
  handle_status √ (必填), handle_remark √ (可选)
  - [ ] 确认后端 HandleAlarmRequest 是否包含 handle_remark 字段（当前 struct 只有 handle_status）

告警规则表单 vs model.AlarmRule(json tag)：
  rule_name √, alarm_type √, scope_type √, scope_id √
  suppression_rule √ (JSON 字符串), escalation_rule √ (JSON 字符串)
  - [ ] 确认创建/编辑时 status 字段是否需要前端传值

筛选参数：alarm_level, handle_status, robot_id, station_id
  - [ ] 确认后端 GET /alarms 支持这些 query params

表格 dataIndex：alarm_id, alarm_level, alarm_type, robot_id, alarm_content, alarm_time, handle_status
  - [ ] 确认与后端 Alarm model json tag 一致

常量值验证：
  ALARM_LEVEL_MAP: 1=提示, 2=一般, 3=严重, 4=紧急
  ALARM_HANDLE_MAP: 0=未处理, 1=已确认, 2=处理中, 3=已完成, 4=已忽略
  - [ ] 与后端枚举定义一致
```

## 7. 系统管理 — System（10 个 Tab）

```
文件：System/index.tsx, api/system.ts (13个 API 对象)
后端：user/role/organization/system_config/dict/audit/notify/report_template/backup handler

用户管理：
  GET /users, POST /users, PUT /users/:id, DELETE /users/:id
  创建表单 vs CreateUserRequest: username, password, real_name, phone, email, role_id, station_ids
  编辑表单 vs UpdateUserRequest: real_name, phone, email, role_id, station_ids, status
  - [ ] user_id 类型：后端 string("admin01")，前端路径参数类型一致
  - [ ] 确认 station_ids 前端如何输入（input? select?）

角色管理：
  GET /roles, POST /roles, PUT /roles/:id, DELETE /roles/:id
  返回数组(非分页)，System 页面是否还有 .data/.list 遗留访问
  表单 vs CreateRoleRequest: role_name, description, permissions
  - [ ] permissions 序列化格式：前端多选 → JSON.stringify → 后端 JSON 字符串

组织管理：
  GET /organizations, GET /organizations/tree, GET /organizations/:id
  POST /organizations, PUT /organizations/:id, DELETE /organizations/:id
  表单 vs model.Organization(json tag)

系统配置：
  GET /system/configs, PUT /system/configs/:key (body:{config_value})
  - [ ] config_key 类型：后端 string，前端 URL 参数一致

数据字典：
  GET /dicts, POST /dicts, PUT /dicts/:id, DELETE /dicts/:id
  GET /dicts/:type/items, POST /dicts/:type/items
  PUT /dicts/items/:id, DELETE /dicts/items/:id

审计日志 / 登录日志：
  GET /audit-logs, GET /login-logs
  - [ ] 返回字段与表格 dataIndex 对齐（System/index.tsx 之前修过 .data 访问）

通知模板 / 报告模板 / 备份恢复：
  各 4 个 CRUD 端点，确认表单字段与 model 结构体 json tag 对齐
```

## 8. 监控中心 — Monitor（仅 REST 部分）

```
文件：Monitor/index.tsx, api/monitor.ts
后端：monitor_handler.go, station_handler.go (GetRealtimeData)

REST 接口：
  GET /monitor/overview      → monitorApi.getDashboard()       → DashboardOverview
  GET /stations/:id/realtime → monitorApi.getRealtime(stationId) → StationRealtime
  GET /monitor/tracks        → monitorApi.getTracks({robot_id,start_time,end_time,limit}) → RobotPosition[]
  GET /monitor/environment   → monitorApi.getEnvironment({robot_id,start_time,end_time})
  GET /monitor/gis-config    → monitorApi.getGisConfig()       → Record<string,string>
  PUT /monitor/gis-config    → monitorApi.updateGisConfig(data) → GISConfigRequest

轨迹回放字段 RobotPosition vs 后端返回：
  robot_id, pos_x, pos_y, pos_z, heading, timestamp

电站实时数据 StationRealtime/RobotRealtime vs 后端返回：
  station_id, station_name, longitude, latitude
  robots[].robot_id, robot_name, robot_type, online_status, work_status
  robots[].battery_level, pos_x, pos_y, pos_z, heading, speed, clean_area
  robots[].temperature, humidity, light_intensity, wind_speed
```

## 9. 数据分析 — Analytics

```
文件：Analytics/index.tsx, api/analytics.ts
后端：analytics_handler.go, prediction_handler.go (全 GET，query params)

接口对照：
  GET /analytics/economic                  → analyticsApi.economic(params)
  GET /analytics/efficiency                → analyticsApi.efficiency(params)
  GET /analytics/reports                   → analyticsApi.reports(params)
  GET /analytics/compare                   → analyticsApi.compare(params)
  GET /analytics/timeseries                → analyticsApi.timeseries(params)
  GET /analytics/export                    → analyticsApi.export(params) [blob]
  GET /analytics/predictions/efficiency    → analyticsApi.predictions.efficiency(params)
  GET /analytics/predictions/fault         → analyticsApi.predictions.fault(params)
  GET /analytics/strategy/optimize         → analyticsApi.predictions.strategy(params)

Query params: station_id, start_date, end_date, robot_type √

⚠ 注意：Analytics 页面中多处 (res as any)?.data ?? res
  这表明返回数据格式不确定。需确认后端各接口实际返回的 JSON 结构。
  例如 GET /analytics/economic 返回的是 {data:{...}} 还是直接数组？
  用浏览器 DevTools Network 面板实际调一次确认。

导出功能：responseType: blob，用 window.URL.createObjectURL 下载
  - [ ] 确认后端 GET /analytics/export 返回的是文件流而非 JSON
```

## 10. 固件 / 维护 / 摄像头

```
固件管理 — Firmwares/index.tsx, api/system.ts firmwareApi:
  GET /firmwares, POST /firmwares (FormData multipart)
  PUT /firmwares/:id, DELETE /firmwares/:id
  POST /firmwares/:id/upgrade → body: {robot_ids:string[]} vs UpgradeRobotRequest
  - [ ] POST /firmwares 是 FormData 上传，确认文件字段名与后端期望一致

维护保养 — Maintenance/index.tsx, api/system.ts maintenanceApi:
  GET /maintenance, POST /maintenance, PUT /maintenance/:id, DELETE /maintenance/:id
  GET /maintenance/reminders
  表单 vs model.Maintenance(json tag)

摄像头 — Cameras/index.tsx, api/system.ts cameraApi:
  GET /cameras, POST /cameras, PUT /cameras/:id, DELETE /cameras/:id
  表单 vs model.Camera(json tag)
```

## 11. 机器人配置 — RobotConfig

```
api/robots.ts robotApi.applyConfig:
  GET /robot-configs, POST /robot-configs, PUT /robot-configs/:id, DELETE /robot-configs/:id
  POST /robots/:id/config/apply → body: {config_id} vs ApplyConfigRequest √
  表单 vs model.RobotConfig(json tag)
```

---

## 跨切面验证

### A. 响应解包一致性（防回归）

```
分页接口 (stations/robots/tasks/users/alarms/audit-logs/login-logs/alarm-rules)：
  后端: {code:0, data:{list:[...], total:N}}
  拦截器解包后: {list:[...], total:N}
  前端取值: res.list, res.total

数组接口 (stations/all, roles, organizations)：
  后端: {code:0, data:[...]}
  拦截器解包后: [...]
  前端: 直接当数组用 — 确认无 .data 或 .list 访问

单个对象接口 (GET /:id, POST/PUT 返回值)：
  后端: {code:0, data:{...}}
  拦截器解包后: {...}
  前端: 直接当对象用

⚠ 重点检查 System/index.tsx（2038 行，之前修过一批 .data 访问，可能还有遗漏）
```

### B. Go 类型 → JS 类型 边界

```
Go int/int8/int64 → JSON number → JS number
  Select/Option value 必须用 Number(k)，否则 "1" !== 1 导致匹配失败

Go float64 → JSON number → JS number
  InputNumber 必须不设 stringMode，否则发 "116.4" 字符串 → Go ShouldBindJSON 失败 → 400

Go string → JSON string → JS string
  station_id/robot_id 等 UUID/短ID，URL 拼接正常

Go time.Time → JSON string (ISO 8601) → JS string
  前端直接显示 √
  前端 dayjs format → 后端 time.Time 解析 "YYYY-MM-DD HH:mm:ss" √

Go *time.Time → JSON string | null → JS 需处理 null
  表格 render: val || '-' 或 val ? xx : '-'
```

### C. 错误处理路径

```
所有 API 调用 catch 块的 3 种分支：
  1. 网络错误 → interceptor 统一弹 message.error
  2. 业务错误(code≠0) → interceptor 统一弹 message.error
  3. 表单验证错误(errorFields) → 判断后 return，不弹 toast

- [ ] 确认无页面自己再弹一次 message.error 导致重复提示
- [ ] 确认 loading 状态在所有分支下正确翻转 (try-finally √)
```

---

## 已确认 Bug

| # | 文件 | 行 | 问题 | 修复 |
|---|------|----|------|------|
| 1 | Stations/index.tsx | 139 | `dataIndex: 'address'` → 后端返回 `location` | 改为 `dataIndex: 'location'` |
| 2 | Analytics/index.tsx | 84,119,170,196,224 | `(res as any)?.data ?? res` — API 已解包了但代码再次 .data | 统一改为直接使用 res，确认后端返回格式后去掉 `.data` |

---

## 验证顺序

```
P0: 登录 → 电站 CRUD → 仪表盘 (核心链路)
P1: 机器人 CRUD + 任务 CRUD + 告警 CRUD (业务核心)
P2: System 10 个 Tab + 监控中心 REST + 数据分析
P3: 固件 + 维护 + 摄像头 + 响应解包一致性检查

每完成一项标记 ✅
```
