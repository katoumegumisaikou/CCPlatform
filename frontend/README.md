# CCPlatform 前端

拓达威光伏清扫机器人云控平台 Web 管理端。

## 技术栈

- **React 19** + **TypeScript** — UI 组件开发与类型安全
- **Vite** — 开发服务器与生产构建
- **Ant Design 6** — 企业级 UI 组件库
- **ECharts** (echarts-for-react) — 数据可视化图表
- **Leaflet** (react-leaflet) — GIS 地图展示
- **Zustand 5** — 轻量级全局状态管理
- **React Router 7** — 前端路由
- **Axios** — HTTP 请求客户端

## 快速开始

```bash
npm install        # 安装依赖
npm run dev        # 开发模式，默认 http://localhost:5173
npm run build      # 生产构建，输出到 dist/
npm run preview    # 预览生产构建
```

开发环境下 API 请求默认代理到 `http://localhost:8080/api/v1`，配置见 `vite.config.ts`。

## 项目结构

```
src/
├── api/                    # API 请求层，按模块拆分
│   ├── client.ts           # Axios 实例（Token 注入 + 响应解包）
│   ├── auth.ts, stations.ts, robots.ts, tasks.ts
│   ├── alarms.ts, monitor.ts, analytics.ts, system.ts
├── components/
│   ├── Layout/AppLayout.tsx # 全局布局（侧边栏 + 顶栏 + 面包屑）
│   └── GisMap/index.tsx     # Leaflet GIS 地图组件
├── hooks/useWebSocket.ts    # WebSocket 实时推送 Hook
├── pages/                   # 页面组件，按功能模块拆分
│   ├── Login/               # 登录（验证码 + JWT）
│   ├── Dashboard/           # 仪表盘概览
│   ├── Stations/            # 电站管理
│   ├── Robots/              # 机器人管理 + 控制指令
│   ├── Tasks/               # 任务管理 + 智能排班
│   ├── Alarms/              # 告警中心（记录 + 规则）
│   ├── Monitor/             # 实时监控 + 轨迹回放 + GIS
│   ├── Analytics/           # 数据分析（5种报表 + 导出）
│   ├── Predictions/         # 智能预测（效率/故障/策略）
│   ├── System/              # 系统管理（10个Tab）
│   ├── Firmwares/           # 固件管理 + OTA升级
│   ├── Maintenance/         # 维护保养
│   ├── Cameras/             # 摄像头管理
│   └── Profile/             # 个人中心
├── store/authStore.ts       # Zustand 认证状态（token/user/permissions）
├── types/index.ts           # 全局 TypeScript 类型定义
├── utils/index.ts           # 常量映射（状态码 → 文本/颜色）
└── App.tsx                  # 路由配置 + 权限守卫
```

## API 解包约定

Axios 响应拦截器 (`client.ts`) 自动从 `{code: 0, data: X, message: "success"}` 中提取 `X`，所有页面代码直接使用解包后的数据：

| 后端返回 | 拦截器解包后 |
|----------|-------------|
| `{data: {list: [...], total: N}}` | `{list: [...], total: N}` |
| `{data: [...]}` | `[...]` |
| `{data: {...}}` | `{...}` |

页面代码不应再访问 `.data` 属性。
