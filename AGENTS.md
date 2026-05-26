# CCPlatform - 拓达威光伏清扫机器人云控平台

## 项目概述

基于 Go 的光伏清扫机器人云控平台后端服务，支持多电站管理、3种机器人类型、MQTT数据采集、6大功能模块。

## 重要：回答前必须阅读需求文档

在回答任何关于项目功能、接口、业务逻辑的问题之前，请先阅读项目根目录下的需求规格说明书：

**文档路径**: `拓达威光伏清扫机器人云控平台开发需求规格说明书.docx`

使用以下命令提取文档内容：
```bash
python3 -c "
import zipfile, xml.etree.ElementTree as ET, sys, re
z = zipfile.ZipFile(sys.argv[1])
xml_content = z.read('word/document.xml')
tree = ET.fromstring(xml_content)
ns = {'w': 'http://schemas.openxmlformats.org/wordprocessingml/2006/main'}
texts = []
for p in tree.iter('{http://schemas.openxmlformats.org/wordprocessingml/2006/main}p'):
    line = ''.join(r.text or '' for r in p.iter('{http://schemas.openxmlformats.org/wordprocessingml/2006/main}t'))
    if line.strip():
        texts.append(line.strip())
print('\n'.join(texts))
" "拓达威光伏清扫机器人云控平台开发需求规格说明书.docx"
```

## 技术栈

- **后端**: Go 1.21+, Gin, MySQL 8.0 + GORM, mochi-mqtt, JWT, gorilla/websocket
- **前端**: React 19 + TypeScript, Vite, Ant Design 6, ECharts, Leaflet, Zustand, React Router 7

## 项目结构

```
cmd/server/main.go              - 后端入口
internal/config/                 - 配置加载
internal/model/                  - 数据模型 (GORM)
internal/repository/             - 数据访问层
internal/service/                - 业务逻辑层
internal/handler/                - HTTP处理器
internal/middleware/              - 中间件 (JWT, RBAC, CORS)
internal/mqtt/                   - MQTT Broker与消息处理
internal/ws/                     - WebSocket实时推送
internal/router/                 - 路由注册
internal/scheduler/              - 定时任务调度
pkg/                             - 公共工具 (响应格式, 错误码, JWT, 哈希)
config/config.yaml               - 配置文件
migrations/init.sql              - 数据库迁移
frontend/                        - 前端代码 (React + TypeScript)
  src/
    api/                         - API 请求层，按模块拆分 (auth/stations/robots/tasks/alarms/monitor/analytics/system)
    components/
      Layout/AppLayout.tsx       - 全局布局 (侧边栏+顶栏+面包屑)
      GisMap/index.tsx           - Leaflet 地图组件
    hooks/useWebSocket.ts        - WebSocket 自定义 Hook
    pages/                       - 页面组件，按功能模块拆分
    store/authStore.ts           - Zustand 认证状态管理
    types/index.ts               - 全局 TypeScript 类型定义
    utils/index.ts               - 常量映射 (状态码→文本/颜色)
    App.tsx                      - 路由配置 + 权限守卫
    main.tsx                     - 入口
```

## 后端按模块读取代码

后端文件按 `{模块}_{层级}` 命名，同一模块的文件分布在 4 个层级目录中：

| 层级 | 目录 | 文件命名 | 示例 |
|------|------|---------|------|
| HTTP 处理器 | `internal/handler/` | `{模块}_handler.go` | `robot_handler.go`, `alarm_handler.go` |
| 业务逻辑 | `internal/service/` | `{模块}_service.go` | `robot_service.go`, `alarm_service.go` |
| 数据访问 | `internal/repository/` | `{模块}_repo.go` | `robot_repo.go`, `alarm_repo.go` |
| 数据模型 | `internal/model/` | `{模块}.go` | `robot.go`, `alarm.go` |

阅读某个后端模块时，按数据流方向依次读取 4 个文件：**Model → Repo → Service → Handler**。

## 前端按模块读取代码

前端代码按功能模块组织，每个模块涉及以下文件：

| 层级 | 目录/文件 | 说明 |
|------|----------|------|
| 类型定义 | `frontend/src/types/index.ts` | 所有模块的 TS 接口定义 |
| API 请求 | `frontend/src/api/{模块}.ts` | 对应后端 handler 的 HTTP 调用 |
| 页面组件 | `frontend/src/pages/{模块}/index.tsx` | 模块主页面 |
| 常量映射 | `frontend/src/utils/index.ts` | 状态码→文本/颜色的映射表 |

阅读某个前端模块时，按数据流方向依次读取：**types → api → utils → pages**。

## 审查代码时的隔离规则

**前后端代码不得在同一轮审查中混读。** 审查时必须明确当前审查的是后端 (Go) 还是前端 (React/TypeScript)，只读取目标端的文件：

- **审查后端**：只读 `internal/`、`pkg/`、`cmd/`、`migrations/` 目录下的代码
- **审查前端**：只读 `frontend/src/` 目录下的代码
- **跨端接口对齐审查**：先完成后端审查并记录接口清单，再独立审查前端是否覆盖了所有接口

## 开发规范

- 使用中文注释和文档
- 遵循分层架构: Handler → Service → Repository → Model
- 统一使用 pkg/response 返回JSON响应
- 统一使用 pkg/errcode 定义错误码

## 提交规范

完成代码修改后，根据变更规模自动执行不同的提交策略：

### 小功能/修复 — 自动commit

适用于：单个文件修改、bug修复、小范围改动（如修改一个struct字段、添加一个方法、修复编译错误等）。

完成后直接 `git add` + `git commit`，commit message 用中文描述本次改动。

### 大功能 — rebase远端后提交

适用于：跨多个文件的改动、新功能模块、架构调整（如本次的ID类型统一、多模块联动修改等）。

完成后执行：
```bash
git add <改动文件>
git commit -m "feat: 功能描述"
git pull origin dev --rebase
git push origin dev
```

### Commit Message 格式

使用中文，格式为：`<类型>: <描述>`

类型包括：
- `feat`: 新功能
- `fix`: 修复bug
- `refactor`: 重构
- `docs`: 文档
- `chore`: 构建/工具变动
