# 任务：精简前端系统管理模块，仅保留用户管理和角色管理

## 目标
移除前端"系统管理"页面中除**用户管理**和**角色管理**之外的所有子模块，后端 API 保持不变。

## 涉及文件
1. `frontend/src/pages/System/index.tsx` — 系统管理主页面
2. `frontend/src/components/Layout/AppLayout.tsx` — 侧边栏菜单 + 面包屑映射
3. `frontend/src/api/system.ts` — 仅清理移除 tab 专用的 API 导出，不能影响其他页面
4. `frontend/src/types/index.ts` — 仅清理移除 tab 专用的类型定义，不能影响其他页面

## 具体改动

### 1. `frontend/src/pages/System/index.tsx`

**删除以下组件/代码块：**
- `OrganizationsTab` 组件（组织架构，约 230 行）
- `ConfigsTab` 组件（系统配置，约 150 行）
- `DictsTab` 组件（数据字典，约 330 行）
- `AuditLogsTab` 组件（审计日志，约 100 行）
- `LoginLogsTab` 组件（登录日志，约 100 行）
- `NotifyTemplatesTab` 组件（通知模板，约 180 行）
- `ReportTemplatesTab` 组件（报告模板，约 220 行）
- `BackupsTab` 组件（备份恢复，约 150 行）
- 所有仅被上述组件使用的 import：`Tree`、`DatePicker`、`InputNumber`、`Spin`、`SearchOutlined`、`ExclamationCircleOutlined`、`RollbackOutlined`、`FolderOpenOutlined`
- 注意不要删除仍被 `RolesTab` 使用的 `Checkbox`、`Tooltip`、`Row`、`Col`
- 所有仅被上述组件使用的 API import；`UsersTab` 仍需要 `orgApi.list()` 加载所属组织下拉，所以 `System/index.tsx` 中至少保留 `userApi`、`roleApi`、`orgApi`
- 所有仅被上述组件使用的类型 import
- 所有仅被上述组件使用的常量/工具函数/本地类型（如 `AuditLogEntry`、`LoginLogEntry`、`BackupEntry`、`TPL_TYPE_OPTIONS`、`ALARM_LEVEL_OPTIONS`、`formatFileSize`、`formatDateTime` 等，确认无其他引用后删除）
- 注意不要删除仍被 `RolesTab` 使用的 `PERMISSION_OPTIONS`
- `client` import（如果仅用于 `ConfigsTab`）

**保留的代码：**
- `UsersTab` 组件
- `RolesTab` 组件
- `SystemPage` 主组件（但需修改 tabItems 和 renderTabContent）

**保留的依赖：**
- `UsersTab` 仍依赖 `Organization` 类型和 `orgApi.list()`，因为用户创建/编辑表单需要选择 `org_id`
- `RolesTab` 仍依赖 `PERMISSION_OPTIONS`、`Checkbox`、`Tooltip`、`Row`、`Col`

**修改主组件 `SystemPage`：**
- `tabItems` 只保留两项：
  - `{ key: 'users', label: '用户管理' }`
  - `{ key: 'roles', label: '角色管理' }`
- `renderTabContent` 的 switch 只保留 `users` 和 `roles` 两个 case，default 也指向 `users`
- 处理历史/手输旧路径：如果访问 `/system/organizations`、`/system/configs` 等已移除路径，应自动跳转或归一化到 `/system/users`，避免 `Tabs activeKey` 指向不存在的 tab

### 2. `frontend/src/components/Layout/AppLayout.tsx`

**侧边栏菜单 `menuItems` 中**，`/system` 的 children 只保留：
```
{ key: '/system/users', label: '用户管理' },
{ key: '/system/roles', label: '角色管理' },
```

**面包屑 `breadcrumbMap` 中**，只保留：
```
'/system/users': '用户管理',
'/system/roles': '角色管理',
```

### 3. `frontend/src/api/system.ts`

检查并删除仅被移除 tab 使用的 API 导出：
- 可以删除：`configApi`、`dictApi`、`notifyApi`、`reportApi`、`backupApi`、`auditApi`、`loginLogApi`
- 必须保留：`userApi`、`roleApi`
- 必须保留：`orgApi`，因为 `UsersTab` 创建/编辑用户时仍需要组织下拉
- 必须保留：`cameraApi`、`firmwareApi`、`maintenanceApi`，因为它们分别被摄像头、固件、维护页面使用

删除前用 `rg` 检查引用，不能为了精简系统管理而破坏其他页面。

### 4. `frontend/src/types/index.ts`

检查并删除仅被移除 tab 使用的 TypeScript 类型定义：
- 可以删除：`SystemConfig`、`DictType`、`DictItem`、`NotifyTemplate`、`ReportTemplate`
- 必须保留：`Organization`，因为 `UsersTab` 仍使用组织下拉
- 必须保留：`Camera`、`Firmware`、`Maintenance`，因为其他页面仍使用这些类型

删除前用 `rg` 检查引用，不能删除被 System 之外或保留 tab 使用的类型。

## 约束
- 后端接口完全不动，只改前端代码
- 删除代码时确保 TypeScript 编译通过（无未使用的 import、无类型错误）
- 清理 import 时仔细检查，不要误删被保留代码引用的内容
- 完成后运行 `cd frontend && npx tsc -b` 确认 TypeScript 编译无错误
- 如时间允许，再运行 `cd frontend && npm run build` 验证 Vite 构建
