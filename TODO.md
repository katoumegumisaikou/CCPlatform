# 权限模块修复计划（基于 08-permission Review）

> Review 结论：RBAC 中间件使用硬编码权限，数据库 `roles.permissions` 字段形同虚设，权限粒度过粗（模块级而非操作级），导致普通用户和监控值班员越权。

## 问题列表

| # | 严重度 | 问题 | 影响角色 |
|---|--------|------|---------|
| 1 | P0 | `checkPermission` 使用硬编码 map，不读 DB `roles.permissions` | 全部 |
| 2 | P0 | 权限粒度只有模块级（`monitor`），无操作级（`monitor:view` vs `monitor:control`） | viewer, operator |
| 3 | P0 | viewer 有 `monitor` 全权限，需求仅允许"监控中心-查看" | viewer |
| 4 | P0 | operator 有 `alarm` 全权限（含 config），需求仅允许"告警管理-处理" | operator |
| 5 | P1 | 数据隔离未实现 — `user.StationIDs` 未被任何中间件/Service 使用 | station_mgr |

## 涉及文件

| 文件 | 改动 |
|------|------|
| `internal/middleware/rbac.go` | 核心：`checkPermission` 改为读 DB `roles.permissions` JSON，支持操作级匹配 |
| `internal/service/user_service.go` | `InitDefaultRoles()` 电站管理员 `monitor:view` → `monitor:*` |

## 不改的文件

| 文件 | 原因 |
|------|------|
| `internal/handler/` | 各 handler 不感知 RBAC 内部实现 |
| `internal/router/*.go` | `RBAC("xxx:action")` 参数不变，只是匹配逻辑变 |
| `internal/model/role.go` | 预定义角色常量不变 |

---

## 1. rbac.go — 重构 RBAC 中间件

### 现状

```go
// rbac.go:42-73 — 硬编码权限，不读 DB，只匹配模块名
func checkPermission(roleID, permission string) bool {
    permissions := map[string][]string{
        model.RoleStationMgr: {"station", "robot", "task", "alarm", "monitor", "analytics"},
        model.RoleEngineer:   {"robot", "task", "alarm", "monitor"},
        model.RoleOperator:   {"monitor", "alarm", "task"},
        model.RoleViewer:     {"monitor", "analytics"},
    }
    // module = permission 中 ':' 前的部分
    // 匹配任一 allowed module 即通过
}
```

问题：
- DB 中 `roles.permissions` = `["station:*","robot:*","monitor:view"]` 从未被读取
- `"monitor"` 匹配了所有 monitor 操作，无法区分 view/control

### 目标

```go
// 改为：从 DB 读 role.Permissions JSON，支持 "module:action" 精确匹配
//
// 匹配规则：
//   perm == "*"                              → 通过（sys_admin）
//   perm == "module"                         → 通过（该模块所有操作）
//   perm == "module:*"                       → 通过（该模块所有操作）
//   perm == "module:action"                  → required == module:action 时通过
//   perm == "module:control" && required == "module:view" → control 隐含 view
//
// roleRepo 通过 NewRBAC() 在 main.go 初始化时注入（单例缓存，避免每次请求查 DB）
```

### 具体改动

1. **新增 `NewRBAC(roleRepo)` 构造函数**，启动时注入 RoleRepo
2. **RBAC 中间件**改为闭包捕获 roleRepo
3. **`checkPermission`** 改为：
   - 从 roleRepo 读 `role.Permissions` JSON
   - 解析 JSON 数组
   - 按上述规则逐条匹配
4. **router.go** 改为 `rbacMiddleware := middleware.NewRBAC(roleRepo)` 后传入

### 权限匹配伪代码

```
func matchPermission(permList []string, required string) bool:
  requiredMod, requiredAct = split(required, ":")
  for each perm in permList:
    if perm == "*" → true
    mod, act = split(perm, ":")
    if mod != requiredMod → continue
    if act == "" || act == "*" → true
    if act == requiredAct → true
    if act == "control" && requiredAct == "view" → true  // control 隐含 view
  return false
```

---

## 2. InitDefaultRoles() — 修正电站管理员权限

### 现状 (user_service.go:161)

```go
{RoleID: model.RoleStationMgr, Permissions: `["station:*","robot:*","task:*","alarm:*","monitor:view","analytics:view"]`},
```

### 目标

```go
{RoleID: model.RoleStationMgr, Permissions: `["station:*","robot:*","task:*","alarm:*","monitor:*","analytics:*"]`},
```

需求 6.2 中电站管理员拥有"监控中心-控制"和"数据分析"全部权限，当前 `monitor:view` 过窄。

---

## 3. router.go — RBAC 中间件改为依赖注入

### 现状

```go
// 各处直接 middleware.RBAC("xxx:action") 调用
```

### 目标

```go
// cmd/server/main.go 初始化：
rbacMiddleware := middleware.NewRBAC(repository.NewRoleRepo())

// router.go Setup 接受 rbacMiddleware，各路由改调用
```

---

## 各角色权限对齐验证（修复后）

| 需求 | sys_admin | station_mgr | engineer | operator | viewer |
|------|:---------:|:-----------:|:--------:|:--------:|:------:|
| 监控中心-查看 | `*` | `monitor:*` | `monitor:view` | `monitor:view` | `monitor:view` |
| 监控中心-控制 | `*` | `monitor:*` | `monitor:view` | `monitor:view` | ❌ |
| 任务管理 | `*` | `task:*` | `task:view` | `task:view` | ❌ |
| 设备管理-配置 | `*` | `robot:*` | `robot:*` | ❌ | ❌ |
| 告警管理-处理 | `*` | `alarm:*` | `alarm:handle` | `alarm:handle` | ❌ |
| 数据分析 | `*` | `analytics:*` | ❌ | ❌ | `analytics:view` |
| 系统管理 | `*` | ❌ | ❌ | ❌ | ❌ |

---

## 执行步骤

1. 修改 `rbac.go` — 新增 `NewRBAC(roleRepo)`，重构 `checkPermission` 为操作级匹配
2. 修改 `router.go` — 改为注入 rbacMiddleware
3. 修改 `main.go` — 初始化 rbacMiddleware
4. 修改 `user_service.go` — `InitDefaultRoles()` 修正 station_mgr 的 permissions
5. 编译验证 + curl 测试各角色关键接口

## 不在此次范围

- 数据隔离（StationIDs 过滤）— 后续单独计划
- 密码复杂度 / 登录失败限制 / Token 刷新 — 后续单独计划
