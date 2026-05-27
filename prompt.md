# 任务：重构监控中心地图，支持拖拽缩放、远程控制和数百级机器人模拟

## 目标

重构前端监控中心的地图展示与模拟运行能力：

- 地图主区域要足够大，不能被右侧详情栏挤压。
- GIS 地图和二维场景都要支持拖动、缩放、点击选中机器人。
- 模拟运行要从 3 台机器人扩展到数百台机器人，默认 240 台。
- 远程指令不仅能调用接口，也要在模拟模式下真实改变机器人的行动状态。
- 为性能可以牺牲部分动画效果，优先保证拖拽、缩放、选择、控制、实时状态可用。
- 后端接口保持不变，只改前端代码。

## 审查结论

当前计划方向正确，但有以下问题需要修正：

1. `GisMap` 基于 Leaflet，底图本身已经支持拖拽缩放，真正的问题是机器人使用 `L.marker` DOM 标记，数百台会产生大量 DOM 更新。
2. `GisMap` 目前接收 `selectedRobotId` 但没有实际使用，Canvas 方案必须补齐选中高亮。
3. `Station2DScene` 当前只做固定视口绘制，不支持平移、缩放、框选，点击检测也是每帧生成屏幕坐标后全量排序。
4. 原计划写了 `useWebSocket` 不改，又写了 WebSocket position 合并，存在矛盾。合并逻辑应放在 `Monitor/index.tsx` 的 handler 层，不改 `useWebSocket`。
5. 原计划说远程控制无需改动不准确。真实接口 `robotApi.sendCommand()` 可复用，但模拟模式现在只弹成功提示，没有让机器人停止、启动、回仓或复位。
6. 批量控制如果直接 `Promise.allSettled` 一次发送数百个请求，容易压垮浏览器和后端，应按批次并发。
7. 模拟机器人的二维坐标不能直接用于 GIS 经纬度。GIS 模式下需要把模拟世界坐标映射到电站附近的经纬度。

## 涉及文件

1. `frontend/src/pages/Monitor/index.tsx`
2. `frontend/src/components/GisMap/index.tsx`
3. `frontend/src/components/Station2DScene/index.tsx`

## 不修改的文件

- `frontend/src/api/robots.ts`：继续使用 `robotApi.sendCommand(id, cmd, params)`。
- `frontend/src/api/monitor.ts`：接口保持不变。
- `frontend/src/hooks/useWebSocket.ts`：保持 WebSocket 连接逻辑不变。
- `frontend/src/components/Robot2DScene/index.tsx`：机器人详情缩略图保持不变。
- `frontend/src/types/index.ts`：全局类型保持不变，组件内部新增局部类型即可。
- 后端所有文件：完全不动。

## 具体改动

### 1. `frontend/src/pages/Monitor/index.tsx`

#### 布局调整

- 引入 Ant Design `Drawer`。
- 主布局从 `display: flex` 左右分栏改为单列地图工作区。
- 顶部保留电站选择、模拟运行开关、刷新、地图模式切换。
- 地图容器占满剩余空间，建议 `height: calc(100vh - 180px)`，最小高度 560px。
- 机器人详情、电站概览、机器人列表移入右侧 `Drawer`，宽度 460px。
- 选中机器人时自动打开详情 Drawer。
- 未选中机器人时可通过“机器人列表”按钮打开 Drawer 查看列表。
- 地图上保留一个轻量浮动状态条：总数、在线数、选中数、当前 FPS/渲染模式。

#### 模拟数据扩展

- 新增常量：

```typescript
const MOCK_ROBOT_COUNT = 240;
const MOCK_WORLD_BOUNDS = { minX: 0, maxX: 120, minY: 0, maxY: 90 };
const MOCK_BASE_POINT = { x: 6, y: 8 };
```

- 将当前 `makeMockRealtime(tick)` 从固定 3 台改为基于稳定 seed 生成 240 台：
  - 固定式机器人 100 台，分布在光伏板阵列行列上。
  - 接驳车 40 台，沿主通道和横向道路循环。
  - 全智能机器人 100 台，在多个作业区内巡航。
- 不要每 tick 重新随机。新增 `makeMockRobotSeed(index)`，机器人 ID、名称、类型、初始轨道、相位、电池基准保持稳定。
- 新增 `mockCommandStateRef` 保存模拟控制状态。由于控制目标可以是多台（`selectedRobotIds`），必须用 Map 存储每台机器人的独立状态：

```typescript
type MockCommandState = {
  mode: 'running' | 'stopped' | 'returning' | 'resetting';
  lastCommand?: string;
  commandAt: number; // tick 时间戳，用于 reset 等定时恢复
  frozenPosition?: { x: number; y: number; heading: number };
};
```

- `mockCommandStateRef` 类型为 `React.MutableRefObject<Map<string, MockCommandState>>`。
- 发送模拟指令时遍历 `selectedRobotIds`，对每台机器人写入或更新对应的 `MockCommandState`：
  - `start`：mode 变为 `running`，删除 frozenPosition，机器人继续按轨迹运动，`work_status = 1`。
  - `stop`：mode 变为 `stopped`，冻结当前坐标存入 frozenPosition，`speed = 0`，`work_status = 0`。
  - `return`：mode 变为 `returning`，机器人从当前位置逐步靠近 `MOCK_BASE_POINT`，到达后 `work_status = 2`。
  - `reset`：mode 变为 `resetting`，记录 commandAt，保持 2 秒 `speed = 0`，之后 mode 恢复为 `running`。
- `makeMockRobot` 生成每台机器人时读取 `mockCommandStateRef.current.get(robotId)`，根据 mode 覆写位置、速度、工作状态字段。
- 模拟 tick 仍保持 1Hz 数据更新即可，不需要高频动画。

#### GIS 模拟坐标映射

- 当前 mock 的 `pos_x/pos_y` 是二维世界坐标，不是经纬度。
- 新增 helper，在传给 `GisMap` 前把 mock 坐标映射到电站经纬度附近：

```typescript
function mockWorldToLatLng(x: number, y: number): { longitude: number; latitude: number } {
  const lngSpan = 0.018;
  const latSpan = 0.012;
  return {
    longitude: MOCK_STATION.longitude + (x / 120 - 0.5) * lngSpan,
    latitude: MOCK_STATION.latitude + (y / 90 - 0.5) * latSpan,
  };
}
```

- `toRobotMarker` 在 `mockEnabled` 时使用映射后的经纬度；真实数据仍保持 `pos_x=longitude`、`pos_y=latitude` 的现有约定。

#### 选择与批量控制

- 新增状态：

```typescript
const [selectedRobotIds, setSelectedRobotIds] = useState<Set<string>>(new Set());
const [drawerOpen, setDrawerOpen] = useState(false);
const [drawerMode, setDrawerMode] = useState<'station' | 'robot' | 'list'>('station');
const [commandRunning, setCommandRunning] = useState(false);
const [commandProgress, setCommandProgress] = useState({ done: 0, total: 0 });
```

- 保留 `selectedRobotId` 作为详情主机器人，兼容现有详情逻辑。
- 新增 `handleRobotSelect(robotId, options)`：
  - 普通点击：只选中该机器人，打开详情 Drawer。
  - Ctrl/Meta 点击：追加或取消选中，不关闭地图。
  - 框选：替换选中集合；按住 Ctrl/Meta 框选则追加。
- 机器人列表不要继续用数百张 `Card`。改为 `Table`，分页 20 或 30 条，支持 rowSelection。
- 远程控制按钮的目标是 `selectedRobotIds`。如果集合为空，则退回当前 `selectedRobotId`。
- 批量发送真实指令时不要一次性发出所有请求，新增批处理 helper：

```typescript
async function sendCommandInBatches(ids: string[], cmd: string, batchSize = 20) {
  const results: PromiseSettledResult<void>[] = [];
  for (let i = 0; i < ids.length; i += batchSize) {
    const batch = ids.slice(i, i + batchSize);
    const settled = await Promise.allSettled(
      batch.map((id) => robotApi.sendCommand(id, cmd).then(() => undefined)),
    );
    results.push(...settled);
    setCommandProgress({ done: results.length, total: ids.length });
  }
  return results;
}
```

- 模拟模式下不要调用接口，改为更新 `mockCommandStateRef`，并立即触发一次 mock 数据刷新。

#### WebSocket position 合并

- 不修改 `useWebSocket`。
- 在 `Monitor/index.tsx` 内新增：

```typescript
const pendingPositionsRef = useRef<Map<string, Record<string, unknown>>>(new Map());
const flushPositionTimerRef = useRef<number | null>(null);
```

- `position` handler 只把同一机器人的最新 position 放入 `pendingPositionsRef`。
- 每 100ms flush 一次，单次 `setRobotMarkers` 合并所有 position，减少数百机器人高频上报造成的 React 更新次数。
- heartbeat/status/clean 可以保留现状；如果后续数据量变大，再同样合并。

### 2. `frontend/src/components/Station2DScene/index.tsx`

#### Props 调整

新增和替换为以下能力：

```typescript
interface Station2DSceneProps {
  stationName?: string;
  robots: SceneRobot[];
  selectedRobotId?: string;
  selectedRobotIds?: Set<string>;
  onRobotSelect?: (robotId: string, options?: { additive?: boolean }) => void;
  onSelectionChange?: (robotIds: string[], options?: { additive?: boolean }) => void;
  style?: CSSProperties;
}
```

- 保留 `onRobotClick` 兼容可以不做，推荐统一迁移到 `onRobotSelect`。
- `selectedRobotIds` 用于批量高亮；`selectedRobotId` 用于主详情高亮。

#### 坐标与视口

- 不再每帧用当前机器人 min/max 归一化，否则地图会随机器人运动抖动。
- 使用固定世界坐标范围 `0..120`、`0..90`。
- 新增 refs：

```typescript
const viewportRef = useRef({ zoom: 1, panX: 0, panY: 0 });
const pointerStateRef = useRef<...>();
const visibleRobotsRef = useRef<Array<{ robotId: string; sx: number; sy: number }>>([]);
```

- 实现工具函数：

```typescript
function worldToScreen(point, viewport, canvasSize): Point
function screenToWorld(point, viewport, canvasSize): Point
function clampZoom(nextZoom: number): number
function getVisibleWorldRect(viewport, canvasSize): Rect
```

- 缩放范围建议 `0.4x ~ 6x`。
- 滚轮缩放以鼠标位置为锚点，缩放后鼠标下的世界坐标保持不变。
- 拖拽平移使用 Pointer Events：`onPointerDown`、`onPointerMove`、`onPointerUp`、`onPointerCancel`。
- 移动端至少支持单指拖动；双指缩放可作为增强项，不作为首轮强制项。

#### 渲染性能

- 当机器人数量超过 100 时，目标渲染 15fps，不追求 60fps。
- `requestAnimationFrame` 中用时间戳节流：

```typescript
const frameInterval = robots.length > 100 ? 1000 / 15 : 1000 / 30;
if (time - lastRenderTimeRef.current < frameInterval) {
  frameId = requestAnimationFrame(render);
  return;
}
```

- 每帧只绘制视口内机器人。先用 `getVisibleWorldRect` 筛选，再绘制。
- 点击检测只在 `visibleRobotsRef` 中找最近点，数百级 O(可见数量) 足够，不必第一版强上空间索引。
- LOD 规则：
  - `zoom < 0.75`：机器人只画状态圆点和选中环，不画名称。
  - `zoom >= 0.75 && robots.length <= 120`：画简化车体和少量名称。
  - `zoom >= 1.5`：画朝向箭头、电量短条、主选中名称。
- 背景光伏板阵列也按世界坐标绘制，并根据可见范围裁剪行列。
- HUD 显示：电站名、在线数、总数、当前缩放、可见机器人数量。

#### 交互

- 单击机器人：调用 `onRobotSelect(robotId, { additive: event.ctrlKey || event.metaKey })`。
- 空白处拖拽：平移地图。
- Shift + 拖拽：画选择框，松手后调用 `onSelectionChange(ids, { additive: event.ctrlKey || event.metaKey })`。
- 双击空白处：视图复位到 `zoom=1, panX=0, panY=0`。

### 3. `frontend/src/components/GisMap/index.tsx`

#### Props 调整

```typescript
interface GisMapProps {
  stations: StationMarker[];
  robots: RobotMarker[];
  center?: [number, number];
  zoom?: number;
  selectedRobotId?: string;
  selectedRobotIds?: Set<string>;
  onRobotSelect?: (robotId: string, options?: { additive?: boolean }) => void;
  style?: React.CSSProperties;
}
```

- 将 `onRobotClick` 迁移为 `onRobotSelect`。如需兼容旧调用，可内部适配。
- `selectedRobotId` 和 `selectedRobotIds` 都要参与 Canvas 高亮。

#### Canvas 叠加层

- 保留 Leaflet map、OSM tile layer、电站 DOM marker。
- 删除机器人 `L.marker` 逻辑和 `robotMarkersRef`。
- 新增一个机器人 Canvas overlay：
  - 在 `overlayPane` 中创建 absolute canvas。
  - 地图 `move`、`zoom`、`resize`、`moveend`、`zoomend` 时重设 canvas 尺寸和位置。
  - 使用 `map.latLngToContainerPoint([robot.pos_y, robot.pos_x])` 转屏幕坐标。
  - 用单次 Canvas 绘制批量机器人圆点、朝向箭头、选中环。
- 不要给每台机器人绑定 popup。点击后交给 Monitor 打开 Drawer。
- Canvas 点击检测使用屏幕坐标阈值 8 到 12px，遍历当前可见机器人即可。
- `robots` 更新时只触发一次 redraw，不再逐个 `setLatLng`。
- 使用 `requestAnimationFrame` 合并重绘：

```typescript
function scheduleDraw() {
  if (drawFrameRef.current != null) return;
  drawFrameRef.current = requestAnimationFrame(() => {
    drawFrameRef.current = null;
    drawRobots();
  });
}
```

#### GIS LOD

- zoom 小于 13：只画 3px 圆点，不画箭头。
- zoom 13 到 16：画 5px 圆点和短箭头。
- zoom 大于 16：画选中名称和电量短条。
- 离线机器人统一灰色，故障机器人红色优先。

## 实施顺序

1. 先改 `Monitor/index.tsx`：布局 Drawer、240 mock、模拟命令状态、批量选择和批量指令。
2. 再改 `Station2DScene/index.tsx`：固定世界坐标、拖拽缩放、视口裁剪、LOD、框选。
3. 再改 `GisMap/index.tsx`：机器人 DOM marker 改 Canvas overlay，补齐选中高亮和点击选择。
4. 最后加 WebSocket position 合并，避免高频位置更新造成 React 连续重渲染。

## 验收标准

- 打开监控中心，开启模拟运行后默认显示 240 台机器人。
- 二维场景可拖动、滚轮缩放、点击选择机器人、Shift 框选机器人、双击复位视图。
- GIS 地图可拖动、缩放，机器人使用 Canvas 绘制，不再创建数百个 `L.marker`。
- 选中机器人后 Drawer 展示机器人详情和远程控制按钮。
- 模拟模式下发送 `start`、`stop`、`return`、`reset` 后，机器人行动状态有可见变化。
- 真实模式下远程控制仍调用 `robotApi.sendCommand()`，后端接口不变。
- 数百机器人场景下地图操作不卡死，二维场景目标帧率不低于 15fps。
- 机器人列表支持分页和多选，不用数百张卡片一次性铺满。

## 验证命令

```bash
cd frontend && npx tsc -b
cd frontend && npm run build
```

## 手工验证

1. 进入监控中心，打开模拟运行。
2. 切换到二维场景，确认 240 台机器人出现，滚轮缩放和拖拽平移正常。
3. 点击单台机器人，确认 Drawer 打开且详情更新。
4. Ctrl/Meta 点击多台机器人或 Shift 框选，确认选中数变化。
5. 对选中机器人发送停止、启动、回仓、复位，确认模拟行动状态变化。
6. 切换到 GIS 地图，确认底图拖拽缩放正常，机器人高亮和点击选择正常。
7. 执行构建命令，确认 TypeScript 和 Vite 构建通过。

## 约束

- 后端接口完全不动，只改前端代码。
- 不引入新的 npm 依赖。
- 不追求复杂动画，优先保证数百台机器人下的交互和控制稳定。
- 首轮实现不强制双指缩放和复杂空间索引；单指拖拽、滚轮缩放、视口裁剪必须完成。
