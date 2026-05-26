# CCPlatform REST API 文档

> 基准路径：`/api/v1` | 响应格式：`{code: 0, message: "success", data: ...}` | 认证方式：JWT Bearer Token

## 目录

- [1. 通用规范](#1-通用规范)
- [2. 公开接口（无需 Token）](#2-公开接口无需-token)
- [3. 认证信息](#3-认证信息)
- [4. 电站管理](#4-电站管理)
- [5. 机器人管理](#5-机器人管理)
- [6. 任务管理](#6-任务管理)
- [7. 告警中心](#7-告警中心)
- [8. 监控中心](#8-监控中心)
- [9. 数据分析](#9-数据分析)
- [10. 用户与权限](#10-用户与权限)
- [11. 系统管理](#11-系统管理)
- [12. 组织管理](#12-组织管理)
- [13. 设备管理](#13-设备管理)
- [14. 报告与备份](#14-报告与备份)
- [15. 错误码](#15-错误码)

## 1. 通用规范

### 1.1 响应格式

所有接口统一返回：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

| 场景 | data 内容 |
|------|----------|
| 分页列表 | `{list: [...], total: 123, page: 1, size: 10}` |
| 数组数据 | `[...]` |
| 单个对象 | `{...}` |
| 操作成功 | `{message: "ok"}` 或空 |

### 1.2 认证

除公开接口外，所有请求需携带 JWT Token：

```
Authorization: Bearer <token>
```

### 1.3 权限

部分接口标注了所需权限码（RBAC），由后端中间件校验。当前预定义 5 种角色：

| 角色 | 权限范围 |
|------|---------|
| 系统管理员 | 全部权限 |
| 电站管理员 | station:\*, robot:\*, task:\*, alarm:\*, monitor:\* |
| 运维工程师 | robot:\*, task:\*, monitor:\*, firmware:\* |
| 值班员 | alarm:view, monitor:view, task:view |
| 普通用户 | 仅查看个人数据 |

---

## 2. 公开接口（无需 Token）

### 2.1 登录

**POST** `/api/v1/auth/login`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | ✅ | 用户名 |
| password | string | ✅ | 密码 |

```
POST /api/v1/auth/login
{"username": "admin", "password": "admin123"}

→ {code: 0, data: {token: "eyJ...", user: {...}}}
```

### 2.2 获取验证码

**GET** `/api/v1/auth/captcha`

→ `{code: 0, data: {captcha_id: "xxx", image: "base64..."}}`

### 2.3 校验验证码

**POST** `/api/v1/auth/verify-captcha`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| captcha_id | string | ✅ | 验证码 ID |
| answer | string | ✅ | 验证码答案 |

### 2.4 发送短信验证码

**POST** `/api/v1/auth/sms-code`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| phone | string | ✅ | 手机号 |

### 2.5 校验短信验证码

**POST** `/api/v1/auth/verify-sms`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| phone | string | ✅ | 手机号 |
| code | string | ✅ | 短信验证码 |

### 2.6 忘记密码

**POST** `/api/v1/auth/forgot-password`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | ✅ | 用户名 |

### 2.7 重置密码

**POST** `/api/v1/auth/reset-password`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| token | string | ✅ | 重置令牌 |
| new_password | string | ✅ | 新密码 |

### 2.8 健康检查

**GET** `/health`

→ `{status: "ok"}`

---

## 3. 认证信息

### 3.1 获取当前用户

**GET** `/api/v1/auth/me` 🔒 JWT

→ 返回当前登录用户的完整信息（User 对象 + 角色 + 权限）

---

## 4. 电站管理

### 4.1 电站列表（分页）

**GET** `/api/v1/stations` 🔒 JWT

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码 |
| size | int | 10 | 每页条数 |
| keyword | string | — | 搜索关键词（匹配名称/编码） |

→ `{list: [Station], total: N, page: 1, size: 10}`

### 4.2 获取全部电站

**GET** `/api/v1/stations/all` 🔒 JWT

→ `[Station, ...]` 返回数组，无分页

### 4.3 电站详情

**GET** `/api/v1/stations/:id` 🔒 JWT

→ `{...Station}`

### 4.4 创建电站

**POST** `/api/v1/stations` 🔒 JWT + `station:create`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| station_name | string | ✅ | 电站名称 |
| station_code | string | ✅ | 电站编码 |
| location | string | | 电站地址 |
| longitude | float64 | | 经度 |
| latitude | float64 | | 纬度 |
| capacity | float64 | | 装机容量 (MW) |
| panel_area | float64 | | 组件面积 (㎡) |
| panel_count | int | | 组件数量 |

### 4.5 更新电站

**PUT** `/api/v1/stations/:id` 🔒 JWT + `station:update`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| station_name | string | | 电站名称 |
| location | string | | 电站地址 |
| longitude | float64 | | 经度 |
| latitude | float64 | | 纬度 |
| capacity | float64 | | 装机容量 (MW) |
| panel_area | float64 | | 组件面积 (㎡) |
| panel_count | int | | 组件数量 |
| status | int8 | | 状态 (0禁用 1启用) |

### 4.6 删除电站

**DELETE** `/api/v1/stations/:id` 🔒 JWT + `station:delete`

### 4.7 电站下属机器人

**GET** `/api/v1/stations/:id/robots` 🔒 JWT

→ `[Robot, ...]`

### 4.8 电站实时数据

**GET** `/api/v1/stations/:id/realtime` 🔒 JWT

→ `{station_id, station_name, longitude, latitude, robots: [RobotRealtime, ...]}`

RobotRealtime 字段：`robot_id, robot_name, robot_type, online_status, work_status, battery_level, pos_x, pos_y, pos_z, heading, speed, clean_area, temperature, humidity, light_intensity, wind_speed`

---

## 5. 机器人管理

### 5.1 机器人列表（分页）

**GET** `/api/v1/robots` 🔒 JWT

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码 |
| size | int | 10 | 每页条数 |
| station_id | string | — | 电站 ID |
| robot_type | int | 0 | 机器人类型 (0=全部) |
| online_status | int | -1 | 在线状态 (-1=全部, 0=离线, 1=在线) |

### 5.2 机器人统计

**GET** `/api/v1/robots/stats` 🔒 JWT

→ `{total, online, offline, fault}`

### 5.3 机器人详情

**GET** `/api/v1/robots/:id` 🔒 JWT

### 5.4 创建机器人

**POST** `/api/v1/robots` 🔒 JWT + `robot:create`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| robot_code | string | ✅ | 机器人编码 |
| robot_name | string | ✅ | 机器人名称 |
| robot_type | int8 | ✅ | 机器人类型 |
| station_id | string | ✅ | 所属电站 ID |

### 5.5 更新机器人

**PUT** `/api/v1/robots/:id` 🔒 JWT + `robot:update`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| robot_name | string | | 机器人名称 |
| robot_type | int8 | | 机器人类型 |
| station_id | string | | 电站 ID |
| status | int8 | | 状态 |

### 5.6 删除机器人

**DELETE** `/api/v1/robots/:id` 🔒 JWT + `robot:delete`

### 5.7 下发控制指令

**POST** `/api/v1/robots/:id/cmd` 🔒 JWT + `robot:control`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| cmd | string | ✅ | 指令名称 (start/stop/return/reset) |
| params | object | | 指令参数 (map) |

### 5.8 高级统计

**GET** `/api/v1/robots/:id/advanced-stats` 🔒 JWT

### 5.9 应用配置

**POST** `/api/v1/robots/:id/config/apply` 🔒 JWT + `robot:config`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| config_id | string | ✅ | 配置模板 ID |

---

## 6. 任务管理

### 6.1 任务列表（分页）

**GET** `/api/v1/tasks` 🔒 JWT

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码 |
| size | int | 10 | 每页条数 |
| station_id | string | — | 电站 ID |
| robot_id | string | — | 机器人 ID |
| task_status | int | -1 | 任务状态 (-1=全部) |

任务状态枚举：`0=待执行, 1=执行中, 2=已完成, 3=已暂停, 4=已取消, 5=失败`

### 6.2 任务详情

**GET** `/api/v1/tasks/:id` 🔒 JWT

### 6.3 创建任务

**POST** `/api/v1/tasks` 🔒 JWT + `task:create`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| task_name | string | ✅ | 任务名称 |
| task_type | int8 | ✅ | 1=即时, 2=定时, 3=周期 |
| robot_id | string | ✅ | 执行机器人 ID |
| area_ids | string | | 清扫区域 ID 列表 |
| plan_start | string | | 计划开始时间 (task_type=2 必填，格式 `YYYY-MM-DD HH:mm:ss`) |
| plan_end | string | | 计划结束时间 (task_type=2 必填) |
| cron_expr | string | | cron 表达式 (task_type=3 必填，如 `0 0 8 * * *`) |

### 6.4 更新任务状态

**PUT** `/api/v1/tasks/:id/status` 🔒 JWT + `task:update`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| status | int8 | ✅ | 目标状态 (1=执行, 2=完成, 3=暂停, 4=取消) |

状态变更会通过 MQTT 下发指令通知机器人。

### 6.5 删除任务

**DELETE** `/api/v1/tasks/:id` 🔒 JWT + `task:delete`

### 6.6 任务进度

**GET** `/api/v1/tasks/:id/progress` 🔒 JWT

### 6.7 智能排程

**POST** `/api/v1/tasks/smart-schedule` 🔒 JWT + `task:create`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| station_id | string | ✅ | 电站 ID |

---

## 7. 告警中心

### 7.1 告警记录

#### 7.1.1 告警列表（分页）

**GET** `/api/v1/alarms` 🔒 JWT

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码 |
| size | int | 10 | 每页条数 |
| station_id | string | — | 电站 ID |
| robot_id | string | — | 机器人 ID |
| alarm_level | int | 0 | 告警级别 (0=全部, 1=提示, 2=一般, 3=严重, 4=紧急) |
| handle_status | int | -1 | 处理状态 (-1=全部, 0=未处理, 1=已确认, 2=处理中, 3=已完成, 4=已忽略) |

#### 7.1.2 告警统计

**GET** `/api/v1/alarms/stats` 🔒 JWT

→ `{total: 100, unhandled: 5, ...}` (Record<string, number>)

#### 7.1.3 最近告警

**GET** `/api/v1/alarms/recent` 🔒 JWT

→ `[Alarm, ...]`

#### 7.1.4 告警详情

**GET** `/api/v1/alarms/:id` 🔒 JWT

#### 7.1.5 处理告警

**PUT** `/api/v1/alarms/:id` 🔒 JWT + `alarm:handle`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| handle_status | int8 | ✅ | 处理状态 (1=已确认, 2=已处理, 3=已忽略) |

#### 7.1.6 告警趋势

**GET** `/api/v1/alarms/trends` 🔒 JWT

### 7.2 告警规则

#### 7.2.1 规则列表（分页）

**GET** `/api/v1/alarm-rules` 🔒 JWT

#### 7.2.2 创建规则

**POST** `/api/v1/alarm-rules` 🔒 JWT + `alarm:config`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| rule_name | string | ✅ | 规则名称 |
| alarm_type | string | | 适用告警类型 |
| scope_type | string | | 作用域 (global/station/robot) |
| scope_id | string | | 作用域 ID |
| threshold_value | string | | 阈值配置 (JSON) |
| suppression_rule | string | | 抑制规则 (JSON, e.g. `{within_sec: 300}`) |
| escalation_rule | string | | 升级规则 (JSON, e.g. `[{count: 3, within_min: 10, to_level: 4}]`) |
| status | int8 | | 状态 (0禁用 1启用) |

#### 7.2.3 更新规则

**PUT** `/api/v1/alarm-rules/:id` 🔒 JWT + `alarm:config`

参数同创建。

#### 7.2.4 删除规则

**DELETE** `/api/v1/alarm-rules/:id` 🔒 JWT + `alarm:config`

### 7.3 通知模板

#### 7.3.1 模板列表（分页）

**GET** `/api/v1/notify/templates` 🔒 JWT

#### 7.3.2 创建模板

**POST** `/api/v1/notify/templates` 🔒 JWT + `notify:config`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| tpl_name | string | ✅ | 模板名称 |
| tpl_type | string | | 渠道类型 (sms/email/app/webhook) |
| alarm_level | int8 | | 适用告警级别 (0=全部) |
| channel_config | string | | 渠道配置 (JSON) |
| content_template | string | | 内容模板，支持变量替换 |
| status | int8 | | 状态 (0禁用 1启用) |

#### 7.3.3 更新模板

**PUT** `/api/v1/notify/templates/:id` 🔒 JWT + `notify:config`

#### 7.3.4 删除模板

**DELETE** `/api/v1/notify/templates/:id` 🔒 JWT + `notify:config`

---

## 8. 监控中心

### 8.1 仪表盘总览

**GET** `/api/v1/monitor/overview` 🔒 JWT

→ `{station_count, today_clean_area, running_tasks, unhandled_alarms, robot_stats: {...}, alarm_stats: {...}}`

### 8.2 轨迹回放

**GET** `/api/v1/monitor/tracks` 🔒 JWT + `monitor:view`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| robot_id | string | ✅ | 机器人 ID |
| start_time | string | | 开始时间 |
| end_time | string | | 结束时间 |
| limit | int | | 返回条数限制 |

→ `[RobotPosition, ...]` `{robot_id, pos_x, pos_y, pos_z, heading, timestamp}`

### 8.3 环境历史

**GET** `/api/v1/monitor/environment` 🔒 JWT + `monitor:view`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| robot_id | string | ✅ | 机器人 ID |
| start_time | string | | 开始时间 |
| end_time | string | | 结束时间 |

→ `[EnvironmentData, ...]` `{temperature, humidity, light_intensity, wind_speed, ...}`

### 8.4 GIS 配置

**GET** `/api/v1/monitor/gis-config` 🔒 JWT

→ `Record<string, string>`

**PUT** `/api/v1/monitor/gis-config` 🔒 JWT + `system:config`

Body: `Record<string, string>` (key-value 对)

---

## 9. 数据分析

全部为 GET 请求，🔒 JWT。

### 9.1 经济效益报表

**GET** `/api/v1/analytics/economic`

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| station_id | string | — | 电站 ID |
| start_date | string | 30天前 | 开始日期 (YYYY-MM-DD) |
| end_date | string | 今天 | 结束日期 (YYYY-MM-DD) |

### 9.2 清扫效率报表

**GET** `/api/v1/analytics/efficiency`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| station_id | string | | 电站 ID |

### 9.3 运行报表

**GET** `/api/v1/analytics/reports`

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| station_id | string | — | 电站 ID |
| start_date | string | 30天前 | 开始日期 |
| end_date | string | 今天 | 结束日期 |

### 9.4 多维对比

**GET** `/api/v1/analytics/compare`

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| dimension | string | "station" | 对比维度 (station/robot/date) |
| ids | string | — | 对比目标 ID 列表 |

### 9.5 时序数据

**GET** `/api/v1/analytics/timeseries`

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| station_id | string | — | 电站 ID |
| granularity | string | "day" | 粒度 (hour/day/month) |
| start_time | string | — | 开始时间 |
| end_time | string | — | 结束时间 |

### 9.6 导出报表

**GET** `/api/v1/analytics/export`

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| station_id | string | — | 电站 ID |
| start_date | string | 30天前 | 开始日期 |
| end_date | string | 今天 | 结束日期 |

→ 返回文件流 (blob)，前端通过 `window.URL.createObjectURL` 下载

### 9.7 效率预测

**GET** `/api/v1/analytics/predictions/efficiency` 🔒 JWT

### 9.8 故障预测

**GET** `/api/v1/analytics/predictions/fault` 🔒 JWT

### 9.9 策略优化

**GET** `/api/v1/analytics/strategy/optimize` 🔒 JWT

---

## 10. 用户与权限

### 10.1 用户管理

#### 10.1.1 用户列表（分页）

**GET** `/api/v1/users` 🔒 JWT + `user:view`

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码 |
| size | int | 10 | 每页条数 |
| keyword | string | — | 搜索关键词 |

#### 10.1.2 创建用户

**POST** `/api/v1/users` 🔒 JWT + `user:create`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | ✅ | 用户名 |
| password | string | ✅ | 密码 |
| real_name | string | | 真实姓名 |
| phone | string | | 手机号 |
| email | string | | 邮箱 |
| role_id | string | | 角色 ID |
| station_ids | string | | 授权电站 ID 列表 |

#### 10.1.3 更新用户

**PUT** `/api/v1/users/:id` 🔒 JWT + `user:update`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| real_name | string | | 真实姓名 |
| phone | string | | 手机号 |
| email | string | | 邮箱 |
| role_id | string | | 角色 ID |
| station_ids | string | | 授权电站 ID 列表 |
| status | int8 | | 状态 (0禁用 1启用) |

#### 10.1.4 删除用户

**DELETE** `/api/v1/users/:id` 🔒 JWT + `user:delete`

### 10.2 角色管理

#### 10.2.1 角色列表

**GET** `/api/v1/roles` 🔒 JWT

→ `[Role, ...]` 返回数组，无分页

#### 10.2.2 创建角色

**POST** `/api/v1/roles` 🔒 JWT + `role:create`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| role_name | string | ✅ | 角色名称 |
| description | string | | 角色描述 |
| permissions | string | | 权限列表 (JSON 数组字符串) |

#### 10.2.3 更新角色

**PUT** `/api/v1/roles/:id` 🔒 JWT + `role:update`

#### 10.2.4 删除角色

**DELETE** `/api/v1/roles/:id` 🔒 JWT + `role:delete`

### 10.3 登录日志

**GET** `/api/v1/login-logs` 🔒 JWT

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码 |
| size | int | 10 | 每页条数 |
| user_id | string | — | 用户 ID |
| start_time | string | — | 开始时间 |
| end_time | string | — | 结束时间 |

---

## 11. 系统管理

### 11.1 系统配置

#### 11.1.1 全部配置

**GET** `/api/v1/system/configs` 🔒 JWT

#### 11.1.2 单个配置

**GET** `/api/v1/system/configs/:key` 🔒 JWT

#### 11.1.3 设置配置

**PUT** `/api/v1/system/configs/:key` 🔒 JWT + `system:config`

Body 为 SystemConfig 对象：`{config_key, config_value, category, description}`

#### 11.1.4 删除配置

**DELETE** `/api/v1/system/configs/:key` 🔒 JWT + `system:config`

### 11.2 数据字典

#### 11.2.1 字典类型管理

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| GET | `/api/v1/dicts` | JWT | 字典类型列表 |
| POST | `/api/v1/dicts` | `system:config` | 创建字典类型 |
| PUT | `/api/v1/dicts/:id` | `system:config` | 更新字典类型 |
| DELETE | `/api/v1/dicts/:id` | `system:config` | 删除字典类型 |

创建字典类型字段：`{dict_type, dict_name, description, status}`

#### 11.2.2 字典项管理

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| GET | `/api/v1/dicts/:type/items` | JWT | 某字典类型的全部项 |
| POST | `/api/v1/dicts/:type/items` | `system:config` | 新增字典项 |
| PUT | `/api/v1/dicts/items/:id` | `system:config` | 更新字典项 |
| DELETE | `/api/v1/dicts/items/:id` | `system:config` | 删除字典项 |

创建字典项字段：`{dict_type, item_key, item_value, sort_order, status}`

### 11.3 审计日志

**GET** `/api/v1/audit-logs` 🔒 JWT

分页查询，记录所有敏感操作的用户、时间、IP、操作内容。

---

## 12. 组织管理

### 12.1 组织列表

**GET** `/api/v1/organizations` 🔒 JWT

### 12.2 组织树

**GET** `/api/v1/organizations/tree` 🔒 JWT

→ 返回树形结构数据

### 12.3 组织详情

**GET** `/api/v1/organizations/:id` 🔒 JWT

### 12.4 创建组织

**POST** `/api/v1/organizations` 🔒 JWT + `org:create`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| org_name | string | ✅ | 组织名称 |
| org_code | string | | 组织编码 |
| org_type | string | | 类型 (company/region/station_group) |
| parent_id | string | | 父组织 ID (空=根节点) |
| station_ids | string | | 关联电站 ID 列表 |
| sort_order | int | | 排序 |
| status | int8 | | 状态 (0禁用 1启用) |

### 12.5 更新组织

**PUT** `/api/v1/organizations/:id` 🔒 JWT + `org:update`

### 12.6 删除组织

**DELETE** `/api/v1/organizations/:id` 🔒 JWT + `org:delete`

---

## 13. 设备管理

### 13.1 固件管理

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| GET | `/api/v1/firmwares` | JWT | 固件列表 |
| POST | `/api/v1/firmwares` | `firmware:create` | 上传固件 (multipart/form-data) |
| PUT | `/api/v1/firmwares/:id` | `firmware:update` | 更新固件信息 |
| DELETE | `/api/v1/firmwares/:id` | `firmware:delete` | 删除固件 |

固件字段：`{version, package_path, compatible_robot_types, upgrade_log, status}`

#### 13.1.1 固件升级

**POST** `/api/v1/firmwares/:id/upgrade` 🔒 JWT + `firmware:upgrade`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| robot_ids | string[] | ✅ | 目标机器人 ID 列表 |

### 13.2 维护保养

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| GET | `/api/v1/maintenance` | JWT | 维护记录列表 |
| POST | `/api/v1/maintenance` | `maintenance:create` | 创建维护记录 |
| PUT | `/api/v1/maintenance/:id` | `maintenance:update` | 更新维护记录 |
| DELETE | `/api/v1/maintenance/:id` | `maintenance:delete` | 删除维护记录 |
| GET | `/api/v1/maintenance/reminders` | JWT | 维护提醒列表 |

维护记录字段：`{robot_id, station_id, maint_type, handler, fault_desc, parts_replaced, maint_time, next_maint_time}`

### 13.3 摄像头管理

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| GET | `/api/v1/cameras` | JWT | 摄像头列表 |
| POST | `/api/v1/cameras` | `camera:create` | 添加摄像头 |
| PUT | `/api/v1/cameras/:id` | `camera:update` | 更新摄像头 |
| DELETE | `/api/v1/cameras/:id` | `camera:delete` | 删除摄像头 |

摄像头字段：`{camera_name, station_id, stream_url, camera_type(rtsp/hls), position, status}`

### 13.4 机器人配置模板

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| GET | `/api/v1/robot-configs` | JWT | 配置模板列表 |
| POST | `/api/v1/robot-configs` | `robot:config` | 创建配置模板 |
| PUT | `/api/v1/robot-configs/:id` | `robot:config` | 更新配置模板 |
| DELETE | `/api/v1/robot-configs/:id` | `robot:config` | 删除配置模板 |

配置模板字段：`{config_name, robot_type(0=通用), config_data(JSON), is_default, status}`

---

## 14. 报告与备份

### 14.1 报告模板

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| GET | `/api/v1/report-templates` | JWT | 模板列表 |
| POST | `/api/v1/report-templates` | `system:config` | 创建模板 |
| PUT | `/api/v1/report-templates/:id` | `system:config` | 更新模板 |
| DELETE | `/api/v1/report-templates/:id` | `system:config` | 删除模板 |

报告模板字段：`{tpl_name, tpl_type(efficiency/economic/custom), dimensions(JSON), metrics(JSON), status}`

### 14.2 备份管理

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| GET | `/api/v1/backups` | `system:config` | 备份列表 |
| POST | `/api/v1/backups` | `system:config` | 创建备份 |
| DELETE | `/api/v1/backups/:filename` | `system:config` | 删除备份 |
| POST | `/api/v1/backups/:filename/restore` | `system:config` | 恢复备份 |

---

## 15. 错误码

| code | HTTP Status | 说明 |
|------|-------------|------|
| 0 | 200 | 成功 |
| 401 | 401 | Token 无效或已过期 |
| 403 | 403 | 无权限 |
| 404 | 404 | 资源不存在 |
| 10001 | 400 | 参数校验失败 |
| 10002 | 401 | 用户名或密码错误 |
| 10003 | 409 | 资源已存在（唯一键冲突） |
| 20001 | 500 | 服务器内部错误 |

### 15.1 WebSocket

**GET** `ws://host:port/ws?station_id=xxx`

通过 WebSocket 连接接收实时推送，消息格式：

```json
{"type": "heartbeat|position|status|alarm", "data": {...}}
```

## 附录：完整接口清单

| # | 方法 | 路径 | 权限 | 分页 |
|---|------|------|------|------|
| 1 | GET | `/health` | 公开 | — |
| 2 | POST | `/api/v1/auth/login` | 公开 | — |
| 3 | GET | `/api/v1/auth/captcha` | 公开 | — |
| 4 | POST | `/api/v1/auth/verify-captcha` | 公开 | — |
| 5 | POST | `/api/v1/auth/sms-code` | 公开 | — |
| 6 | POST | `/api/v1/auth/verify-sms` | 公开 | — |
| 7 | POST | `/api/v1/auth/forgot-password` | 公开 | — |
| 8 | POST | `/api/v1/auth/reset-password` | 公开 | — |
| 9 | GET | `/api/v1/auth/me` | JWT | — |
| 10 | GET | `/api/v1/stations` | JWT | ✅ |
| 11 | POST | `/api/v1/stations` | station:create | — |
| 12 | GET | `/api/v1/stations/all` | JWT | — |
| 13 | GET | `/api/v1/stations/:id` | JWT | — |
| 14 | PUT | `/api/v1/stations/:id` | station:update | — |
| 15 | DELETE | `/api/v1/stations/:id` | station:delete | — |
| 16 | GET | `/api/v1/stations/:id/robots` | JWT | — |
| 17 | GET | `/api/v1/stations/:id/realtime` | JWT | — |
| 18 | GET | `/api/v1/robots` | JWT | ✅ |
| 19 | POST | `/api/v1/robots` | robot:create | — |
| 20 | GET | `/api/v1/robots/stats` | JWT | — |
| 21 | GET | `/api/v1/robots/:id` | JWT | — |
| 22 | PUT | `/api/v1/robots/:id` | robot:update | — |
| 23 | DELETE | `/api/v1/robots/:id` | robot:delete | — |
| 24 | POST | `/api/v1/robots/:id/cmd` | robot:control | — |
| 25 | GET | `/api/v1/robots/:id/advanced-stats` | JWT | — |
| 26 | POST | `/api/v1/robots/:id/config/apply` | robot:config | — |
| 27 | GET | `/api/v1/robot-configs` | JWT | — |
| 28 | POST | `/api/v1/robot-configs` | robot:config | — |
| 29 | PUT | `/api/v1/robot-configs/:id` | robot:config | — |
| 30 | DELETE | `/api/v1/robot-configs/:id` | robot:config | — |
| 31 | GET | `/api/v1/tasks` | JWT | ✅ |
| 32 | POST | `/api/v1/tasks` | task:create | — |
| 33 | GET | `/api/v1/tasks/:id` | JWT | — |
| 34 | PUT | `/api/v1/tasks/:id/status` | task:update | — |
| 35 | DELETE | `/api/v1/tasks/:id` | task:delete | — |
| 36 | GET | `/api/v1/tasks/:id/progress` | JWT | — |
| 37 | POST | `/api/v1/tasks/smart-schedule` | task:create | — |
| 38 | GET | `/api/v1/alarms` | JWT | ✅ |
| 39 | GET | `/api/v1/alarms/stats` | JWT | — |
| 40 | GET | `/api/v1/alarms/recent` | JWT | — |
| 41 | GET | `/api/v1/alarms/:id` | JWT | — |
| 42 | PUT | `/api/v1/alarms/:id` | alarm:handle | — |
| 43 | GET | `/api/v1/alarms/trends` | JWT | — |
| 44 | GET | `/api/v1/alarm-rules` | JWT | ✅ |
| 45 | POST | `/api/v1/alarm-rules` | alarm:config | — |
| 46 | PUT | `/api/v1/alarm-rules/:id` | alarm:config | — |
| 47 | DELETE | `/api/v1/alarm-rules/:id` | alarm:config | — |
| 48 | GET | `/api/v1/notify/templates` | JWT | ✅ |
| 49 | POST | `/api/v1/notify/templates` | notify:config | — |
| 50 | PUT | `/api/v1/notify/templates/:id` | notify:config | — |
| 51 | DELETE | `/api/v1/notify/templates/:id` | notify:config | — |
| 52 | GET | `/api/v1/monitor/overview` | JWT | — |
| 53 | GET | `/api/v1/monitor/tracks` | monitor:view | — |
| 54 | GET | `/api/v1/monitor/environment` | monitor:view | — |
| 55 | GET | `/api/v1/monitor/gis-config` | JWT | — |
| 56 | PUT | `/api/v1/monitor/gis-config` | system:config | — |
| 57 | GET | `/api/v1/analytics/economic` | JWT | — |
| 58 | GET | `/api/v1/analytics/efficiency` | JWT | — |
| 59 | GET | `/api/v1/analytics/reports` | JWT | — |
| 60 | GET | `/api/v1/analytics/compare` | JWT | — |
| 61 | GET | `/api/v1/analytics/export` | JWT | — |
| 62 | GET | `/api/v1/analytics/timeseries` | JWT | — |
| 63 | GET | `/api/v1/analytics/predictions/efficiency` | JWT | — |
| 64 | GET | `/api/v1/analytics/predictions/fault` | JWT | — |
| 65 | GET | `/api/v1/analytics/strategy/optimize` | JWT | — |
| 66 | GET | `/api/v1/users` | user:view | ✅ |
| 67 | POST | `/api/v1/users` | user:create | — |
| 68 | PUT | `/api/v1/users/:id` | user:update | — |
| 69 | DELETE | `/api/v1/users/:id` | user:delete | — |
| 70 | GET | `/api/v1/roles` | JWT | — |
| 71 | POST | `/api/v1/roles` | role:create | — |
| 72 | PUT | `/api/v1/roles/:id` | role:update | — |
| 73 | DELETE | `/api/v1/roles/:id` | role:delete | — |
| 74 | GET | `/api/v1/login-logs` | JWT | ✅ |
| 75 | GET | `/api/v1/system/configs` | JWT | — |
| 76 | GET | `/api/v1/system/configs/:key` | JWT | — |
| 77 | PUT | `/api/v1/system/configs/:key` | system:config | — |
| 78 | DELETE | `/api/v1/system/configs/:key` | system:config | — |
| 79 | GET | `/api/v1/dicts` | JWT | — |
| 80 | POST | `/api/v1/dicts` | system:config | — |
| 81 | PUT | `/api/v1/dicts/:id` | system:config | — |
| 82 | DELETE | `/api/v1/dicts/:id` | system:config | — |
| 83 | GET | `/api/v1/dicts/:type/items` | JWT | — |
| 84 | POST | `/api/v1/dicts/:type/items` | system:config | — |
| 85 | PUT | `/api/v1/dicts/items/:id` | system:config | — |
| 86 | DELETE | `/api/v1/dicts/items/:id` | system:config | — |
| 87 | GET | `/api/v1/audit-logs` | JWT | ✅ |
| 88 | GET | `/api/v1/organizations` | JWT | — |
| 89 | GET | `/api/v1/organizations/tree` | JWT | — |
| 90 | GET | `/api/v1/organizations/:id` | JWT | — |
| 91 | POST | `/api/v1/organizations` | org:create | — |
| 92 | PUT | `/api/v1/organizations/:id` | org:update | — |
| 93 | DELETE | `/api/v1/organizations/:id` | org:delete | — |
| 94 | GET | `/api/v1/firmwares` | JWT | — |
| 95 | POST | `/api/v1/firmwares` | firmware:create | — |
| 96 | PUT | `/api/v1/firmwares/:id` | firmware:update | — |
| 97 | DELETE | `/api/v1/firmwares/:id` | firmware:delete | — |
| 98 | POST | `/api/v1/firmwares/:id/upgrade` | firmware:upgrade | — |
| 99 | GET | `/api/v1/maintenance` | JWT | — |
| 100 | POST | `/api/v1/maintenance` | maintenance:create | — |
| 101 | PUT | `/api/v1/maintenance/:id` | maintenance:update | — |
| 102 | DELETE | `/api/v1/maintenance/:id` | maintenance:delete | — |
| 103 | GET | `/api/v1/maintenance/reminders` | JWT | — |
| 104 | GET | `/api/v1/cameras` | JWT | — |
| 105 | POST | `/api/v1/cameras` | camera:create | — |
| 106 | PUT | `/api/v1/cameras/:id` | camera:update | — |
| 107 | DELETE | `/api/v1/cameras/:id` | camera:delete | — |
| 108 | GET | `/api/v1/report-templates` | JWT | — |
| 109 | POST | `/api/v1/report-templates` | system:config | — |
| 110 | PUT | `/api/v1/report-templates/:id` | system:config | — |
| 111 | DELETE | `/api/v1/report-templates/:id` | system:config | — |
| 112 | GET | `/api/v1/backups` | system:config | — |
| 113 | POST | `/api/v1/backups` | system:config | — |
| 114 | DELETE | `/api/v1/backups/:filename` | system:config | — |
| 115 | POST | `/api/v1/backups/:filename/restore` | system:config | — |

> 共 **115 个端点**：8 个公开 + 107 个需 JWT（其中 52 个额外需要 RBAC 权限）
