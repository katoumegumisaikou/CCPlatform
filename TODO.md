# TODO: 机器人上行数据 → 任务状态联动

## 背景

当前 MQTT 上行数据处理（handler.go）只更新 robots 表和写入历史记录表，
**完全没有联动 tasks 表**。这导致：

1. 机器人在清扫过程中上报的 `clean` 消息带有 `task_id` 和 `clean_area`，
   但任务的 `clean_area` 字段始终为 0，任务进度无法实时更新
2. 机器人 `work_status` 从 1(清扫) 变为 0(空闲) 时，平台不知道任务已完成，
   任务永远停留在"执行中"状态，除非工作人员手动调用 UpdateStatus API

## 目标

让机器人的实时上行数据驱动任务状态变化，实现闭环。

## 具体改动

### 1. TaskRepo 新增方法

**文件**: `internal/repository/task_repo.go`

新增 `UpdateCleanArea(taskID uint, area float64) error`：
- 更新指定任务的 `clean_area` 字段
- 实现：`db.Model(&Task{}).Where("task_id = ?", taskID).Update("clean_area", area)`

### 2. MessageHandler 注入 taskRepo

**文件**: `internal/mqtt/handler.go`

- 在 `MessageHandler` 结构体中新增 `taskRepo *repository.TaskRepo` 字段
- 在 `NewMessageHandler` 中初始化 `taskRepo: repository.NewTaskRepo()`

### 3. handleClean 更新任务清扫面积

**文件**: `internal/mqtt/handler.go` → `handleClean` 方法

在写入 `CleaningRecord` 之后，新增逻辑：
- 如果 `msg.TaskID > 0`，调用 `h.taskRepo.UpdateCleanArea(msg.TaskID, msg.CleanArea)`
- 失败不阻塞主流程，仅 log 警告

### 4. handleHeartbeat 检测 work_status 变更 → 自动完成/暂停/失败任务

**文件**: `internal/mqtt/handler.go` → `handleHeartbeat` 方法

旧状态的获取方式：
1. 先调 `h.robotRepo.GetByID(robotID)` 获取当前机器人记录，记下 `oldWorkStatus`
2. 再调 `h.robotRepo.UpdateHeartbeat(robotID, ...)` 写新值
3. 如果 `oldWorkStatus != newWorkStatus`，触发任务联动

每条心跳多一次 DB 查询（GetByID），仅在状态变化时才额外查任务表，开销可接受。

状态变更 → 任务联动对照表：

| 旧 work_status | 新 work_status | 动作 |
|---------------|---------------|------|
| 1 (清扫) | 0 (空闲) | 查找活跃任务(status=1)，标记为已完成(2)，设置 actual_end |
| 1 (清扫) | 2 (充电) | 查找活跃任务(status=1)，标记为已暂停(3)。机器人低电量回充，充完可能继续 |
| 1 (清扫) | 3 (故障) | 查找活跃任务(status=1)，标记为失败(5) |
| 2 (充电) | 1 (清扫) | 查找暂停的任务(status=3)，恢复为执行中(1)。机器人充满电继续扫 |
| 3 (故障) | 0 (空闲) | 不自动处理（故障恢复需人工确认，通过 API 手动操作） |

查找活跃任务使用已有的 `taskRepo.GetActiveByRobot(robotID)` 方法。
但该方法目前只查 status=1 和 0，需扩展为同时查 status=3（已暂停），
以便"充电→清扫"时能找到暂停的任务恢复。详见第 5 节。

### 5. GetActiveByRobot 扩展支持查询暂停任务

**文件**: `internal/repository/task_repo.go` → `GetActiveByRobot` 方法

当前实现只查 `task_status = 1` 和 `task_status = 0`。
为支持"充电→清扫"时恢复暂停的任务，需增加对 `task_status = 3`（已暂停）的查询。

修改后的优先级：
1. 先查执行中 (status=1) — 只有一个机器人同时只执行一个任务
2. 再查已暂停 (status=3) — 可能因低电量回充而暂停
3. 最后查待执行 (status=0) — 作为兜底

### 6. 提取公共函数 onWorkStatusChange

**文件**: `internal/mqtt/handler.go`

`handleHeartbeat` 和 `handleStatus` 都携带 `work_status`，都会触发变更检测。
提取一个公共函数避免重复代码：

```go
func (h *MessageHandler) onWorkStatusChange(robotID string, oldStatus, newStatus int8) {
    // 状态未变化，跳过
    if oldStatus == newStatus {
        return
    }
    switch {
    case oldStatus == 1 && newStatus == 0: // 清扫→空闲: 完成
        h.completeActiveTask(robotID)
    case oldStatus == 1 && newStatus == 2: // 清扫→充电: 暂停
        h.pauseActiveTask(robotID)
    case oldStatus == 1 && newStatus == 3: // 清扫→故障: 失败
        h.failActiveTask(robotID)
    case oldStatus == 2 && newStatus == 1: // 充电→清扫: 恢复
        h.resumePausedTask(robotID)
    }
}
```

**注意**: `handleStatus` 和 `handleHeartbeat` 都会触发此函数。因为 heartbeat(5s) 频率高于 status(10s)，
大多数变更由 heartbeat 先捕获。status 消息到达时 `oldStatus` 已被 heartbeat 更新，`oldStatus == newStatus`，
条件不满足，直接跳过，不会重复处理。

### 7. handleStatus 同步检测逻辑

**文件**: `internal/mqtt/handler.go` → `handleStatus` 方法

与 handleHeartbeat 相同模式：
1. 先 `GetByID(robotID)` 获取旧 work_status
2. 再 `UpdateStatus(robotID, data)` 写新值
3. 从 `data` 中提取新的 work_status，调用 `onWorkStatusChange`

## 影响范围

- `internal/repository/task_repo.go` — 新增 `UpdateCleanArea` 方法，扩展 `GetActiveByRobot` 支持 status=3
- `internal/mqtt/handler.go` — 结构体 +1 字段(taskRepo)，新增 `onWorkStatusChange` 及 4 个辅助方法，
  handleClean/heartbeat/handleStatus 各加一段逻辑
- 不涉及 handler/service/router 层，不改变 HTTP 接口
