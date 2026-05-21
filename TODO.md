# TODO: 周期清扫计划 — 按日/周/月配置定时任务

## 背景

需求规格说明书 4.2.2 明确要求"设置周期性清扫计划，支持按日/周/月配置"。
模型已定义 `CronExpr` 字段，调度器已引入 robfig/cron，但端到端链路未串通。

当前断点：
1. `CreateTaskRequest` 没有 `cron_expr` 字段，前端选不了周期频率
2. `TaskService` 没持有 `TaskScheduler`，调不了 `ScheduleTask`
3. `Create` 对 `task_type=3`（周期任务）跟即时任务一样处理——直接下发 MQTT `task` 指令，不该

## 目标

创建周期任务时：写入 DB → 注册到 cron 调度器；调度器到点自动创建新任务实例并下发给机器人。

## 具体改动

### 1. CreateTaskRequest 新增 cron_expr 字段

**文件**: `internal/handler/task_handler.go` → `CreateTaskRequest` 结构体

新增字段：
```go
CronExpr string `json:"cron_expr"` // 周期任务 cron 表达式（task_type=3 时必填）
```

### 2. TaskHandler.Create 透传 CronExpr

**文件**: `internal/handler/task_handler.go` → `Create` 方法

在构建 `model.Task` 时补上：
```go
CronExpr: req.CronExpr,
```

### 3. TaskService 注入 TaskScheduler

**文件**: `internal/service/task_service.go`

- `TaskService` 结构体新增 `scheduler *scheduler.TaskScheduler` 字段
- `NewTaskService` 签名改为 `NewTaskService(publisher *mqtt.Publisher, sch *scheduler.TaskScheduler) *TaskService`
- 初始化时注入 scheduler

### 4. TaskService.Create 处理周期任务

**文件**: `internal/service/task_service.go` → `Create` 方法

当前逻辑：创建任务 → 下发 MQTT `task` 指令 → 即时任务立即标为执行中。

修改为：
```go
// 周期任务：写入 DB 后注册到 cron 调度器，不下发 MQTT
if task.TaskType == 3 {
    if err := s.scheduler.ScheduleTask(task); err != nil {
        return fmt.Errorf("schedule periodic task: %w", err)
    }
    return nil
}
// 非周期任务：按原逻辑下发 MQTT
```

### 5. 三种任务类型的 MQTT 行为

- **即时任务（task_type=1）**：创建后立即下发 MQTT `task` 指令，并直接标为执行中。
- **定时任务（task_type=2）**：有具体 `plan_start`。
  - 若 `plan_start` 为 nil 或已过期 → 视为即时任务，立即下发
  - 若 `plan_start` 在未来 → 生成一次性 cron 表达式注册到调度器，到点创建实例并下发。
    调度器回调执行后自动移除该 cron，避免次年重复触发。
- **周期任务（task_type=3）**：使用前端传入的 `cron_expr` 注册到调度器，不自动移除（按日/周/月重复执行）。

### 6. TaskService.Create — 定时任务分支

**文件**: `internal/service/task_service.go` → `Create` 方法

在周期任务分支之后、MQTT 下发之前，新增定时任务分支：

```go
// 定时任务：plan_start 在未来则注册为一次性调度
if task.TaskType == 2 && task.PlanStart != nil && task.PlanStart.After(time.Now()) {
    cronExpr := fmt.Sprintf("0 %d %d %d %d *",
        task.PlanStart.Minute(), task.PlanStart.Hour(),
        task.PlanStart.Day(), task.PlanStart.Month())
    task.CronExpr = cronExpr
    _ = s.repo.Update(task) // 回写生成的 cron_expr 到 DB
    if err := s.scheduler.ScheduleTask(task); err != nil {
        return fmt.Errorf("schedule one-shot task: %w", err)
    }
    return nil
}
```

### 7. TaskScheduler 新增 ScheduleOneShot 方法

**文件**: `internal/scheduler/scheduler.go`

新增方法：与 `ScheduleTask` 相同，但回调执行后自动调用 `RemoveTask` 移除自身，
避免 `0 30 8 1 6 *` 这类表达式每年重复触发。

```go
func (s *TaskScheduler) ScheduleOneShot(task *model.Task) error {
    s.cron.Remove(cron.EntryID(task.TaskID))
    
    _, err := s.cron.AddFunc(task.CronExpr, func() {
        // 同 ScheduleTask：创建新实例 + MQTT 下发（此处省略重复代码）
        // ...
        // 执行后移除，防止次年重复触发
        s.cron.Remove(cron.EntryID(task.TaskID))
    })
    return err
}
```

或者直接改现有 `ScheduleTask` 的行为：task_type=2 时回调内自移除，task_type=3 时不移除。

## 附注

上述 1-4、8-10 已实现（周期任务链路已串通）。剩余 6-7（定时任务 type=2）待实现，涉及：

| 文件 | 待改动 |
|------|--------|
| `internal/service/task_service.go` | Create 新增定时任务分支（plan_start 在未来 → 生成 cron 表达式 → 注册调度器） |
| `internal/scheduler/scheduler.go` | 新增 `ScheduleOneShot` 方法（执行后自移除，避免次年重复触发） |
