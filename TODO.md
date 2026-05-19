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
