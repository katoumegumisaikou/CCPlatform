import { useState, useEffect, useCallback, useMemo, useRef } from 'react';
import {
  Card, Descriptions, Tag, Progress, Tabs, Select, Button, DatePicker,
  Spin, Empty, Badge, Space, Divider, Table, Typography, App, Slider,
  Segmented, Switch, Drawer,
} from 'antd';
import type { TabsProps } from 'antd';
import {
  ReloadOutlined, PlayCircleOutlined, PauseCircleOutlined,
  HomeOutlined, UndoOutlined, EnvironmentOutlined,
  UnorderedListOutlined,
} from '@ant-design/icons';
import dayjs, { type Dayjs } from 'dayjs';
import GisMap from '../../components/GisMap';
import Robot2DScene from '../../components/Robot2DScene';
import Station2DScene from '../../components/Station2DScene';
import { useWebSocket } from '../../hooks/useWebSocket';
import { stationApi } from '../../api/stations';
import { monitorApi } from '../../api/monitor';
import { robotApi } from '../../api/robots';
import type {
  Station, StationRealtime, RobotRealtime, Robot, RobotPosition,
  WsMessage, WsMessageType,
} from '../../types';
import { ROBOT_TYPE_MAP, ONLINE_STATUS_MAP, WORK_STATUS_MAP } from '../../utils';

const { Text } = Typography;
const { RangePicker } = DatePicker;

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

interface CommandDef {
  key: string;
  label: string;
  icon: React.ReactNode;
  color: string;
  confirmText: string;
}

const ROBOT_COMMANDS: CommandDef[] = [
  {
    key: 'start', label: '启动清扫',
    icon: <PlayCircleOutlined />, color: '#52c41a',
    confirmText: '确认启动清扫任务？',
  },
  {
    key: 'stop', label: '停止',
    icon: <PauseCircleOutlined />, color: '#ff4d4f',
    confirmText: '确认停止当前任务？',
  },
  {
    key: 'return', label: '回仓',
    icon: <HomeOutlined />, color: '#faad14',
    confirmText: '确认让机器人返回仓库？',
  },
  {
    key: 'reset', label: '复位',
    icon: <UndoOutlined />, color: '#1677ff',
    confirmText: '确认复位机器人？',
  },
];

const DEFAULT_CENTER: [number, number] = [39.9, 116.4];

const TRACK_COLUMNS = [
  { title: '时间', dataIndex: 'timestamp', key: 'timestamp', width: 160,
    render: (v: string) => dayjs(v).format('MM-DD HH:mm:ss') },
  { title: 'X (经度)', dataIndex: 'pos_x', key: 'pos_x', render: (v: number) => v.toFixed(4) },
  { title: 'Y (纬度)', dataIndex: 'pos_y', key: 'pos_y', render: (v: number) => v.toFixed(4) },
  { title: '朝向', dataIndex: 'heading', key: 'heading', render: (v: number) => `${v.toFixed(1)}°` },
];

const DATE_PRESETS: { label: string; value: [Dayjs, Dayjs] }[] = [
  { label: '最近1小时', value: [dayjs().subtract(1, 'hour'), dayjs()] },
  { label: '最近6小时', value: [dayjs().subtract(6, 'hour'), dayjs()] },
  { label: '今天', value: [dayjs().startOf('day'), dayjs()] },
  { label: '最近3天', value: [dayjs().subtract(3, 'day').startOf('day'), dayjs()] },
];

const MOCK_STATION: Station = {
  station_id: 'mock-station-001', station_name: 'Mock光伏示范电站',
  station_code: 'MOCK-PV-001', longitude: 116.397428, latitude: 39.90923,
  capacity: 48.6, location: '模拟光伏园区', panel_area: 128000,
  panel_count: 3200, org_id: 1, status: 1, create_time: new Date().toISOString(),
};

const MOCK_ROBOT_COUNT = 30;
const MOCK_WORLD_BOUNDS = { minX: 0, maxX: 120, minY: 0, maxY: 67 };
const MOCK_BASE_POINT = { x: 6, y: 8 };

type MockCommandMode = 'running' | 'stopped' | 'returning' | 'resetting';

interface MockCommandState {
  mode: MockCommandMode;
  lastCommand?: string;
  commandAt: number;
  frozenPosition?: { x: number; y: number; heading: number };
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function toRobotMarker(r: RobotRealtime) {
  return {
    robot_id: r.robot_id, robot_name: r.robot_name, robot_type: r.robot_type,
    pos_x: r.pos_x, pos_y: r.pos_y, heading: r.heading,
    online_status: r.online_status, work_status: r.work_status,
    battery_level: r.battery_level,
  };
}

function toStationMarker(s: { station_id: string; station_name: string; longitude: number; latitude: number }) {
  return { station_id: s.station_id, station_name: s.station_name, longitude: s.longitude, latitude: s.latitude };
}

function pick<T extends string | number>(realtimeVal: T | undefined | null, detailVal: T | undefined | null, fallback: T): T {
  if (realtimeVal != null) return realtimeVal;
  if (detailVal != null) return detailVal;
  return fallback;
}

function mockWorldToLatLng(x: number, y: number): { longitude: number; latitude: number } {
  const lngSpan = 0.018;
  const latSpan = 0.012;
  return {
    longitude: MOCK_STATION.longitude + (x / MOCK_WORLD_BOUNDS.maxX - 0.5) * lngSpan,
    latitude: MOCK_STATION.latitude + (y / MOCK_WORLD_BOUNDS.maxY - 0.5) * latSpan,
  };
}

function clamp(value: number, min: number, max: number): number {
  return Math.max(min, Math.min(max, value));
}

function mockTrackSeed(robotId: string) {
  return robotId.split('').reduce((sum, char) => sum + char.charCodeAt(0), 0) % 80;
}

// ---------------------------------------------------------------------------
// Mock robot generation (stable seeds)
// ---------------------------------------------------------------------------

interface MockRobotSeed {
  robotId: string;
  robotName: string;
  robotType: number;
  initX: number;
  initY: number;
  phase: number;
  batteryBase: number;
}

function makeMockRobotSeed(index: number): MockRobotSeed {
  // 面板行参数（对齐 Station2DScene PANEL_ORIGIN_Y=5, PANEL_GAP_Y=13.6, PANEL_CELL_H=11）
  const panelRowCenters = [0, 1, 2, 3, 4].map(r => 5 + r * 13.6 + 11 / 2); // Y中心: 10.5, 24.1, 37.7, 51.3, 64.9
  // 列间通道 X 坐标（面板列宽17, 列间隙约0.3, 通道在列之间）
  const panelColGaps = [9 + 17 * 0, 9 + 17 * 1, 9 + 17 * 2, 9 + 17 * 3, 9 + 17 * 4, 9 + 17 * 5, 9 + 17 * 6];

  // 10 固定式 (index 0-9): 每2台一组分布在5行面板上，沿面板行X轴清扫
  if (index < 10) {
    const i = index;
    const row = i % 5; // 5行面板，每行2台
    const col = Math.floor(i / 5); // 0或1
    const minX = panelColGaps[0] + 3;
    const maxX = panelColGaps[panelColGaps.length - 1] - 3;
    return {
      robotId: `mock-fixed-${String(i + 1).padStart(3, '0')}`,
      robotName: `固定式${String(i + 1).padStart(3, '0')}`,
      robotType: 1,
      initX: col === 0 ? minX + (maxX - minX) * 0.25 : minX + (maxX - minX) * 0.75,
      initY: panelRowCenters[row],
      phase: i * 0.35,
      batteryBase: 82 + (i % 18),
    };
  }
  // 10 接驳车 (index 10-19): 在面板列间通道沿Y轴穿行
  if (index < 20) {
    const i = index - 10;
    // 使用列间隙作为通道（跳过首尾，用中间5个通道）
    const chXs = [panelColGaps[1], panelColGaps[2], panelColGaps[3], panelColGaps[4], panelColGaps[5]];
    // 每2台共享一个通道，分布在面板区顶部和底部
    const chIdx = Math.floor(i / 2) % chXs.length;
    return {
      robotId: `mock-shuttle-${String(i + 1).padStart(3, '0')}`,
      robotName: `接驳车${String(i + 1).padStart(3, '0')}`,
      robotType: 2,
      initX: chXs[chIdx] - 1,
      initY: i % 2 === 0 ? 7 : 62, // 顶部和底部
      phase: i * 0.45,
      batteryBase: 78 + (i % 22),
    };
  }
  // 10 全智能 (index 20-29): 全区域自由巡航
  const i = index - 20;
  return {
    robotId: `mock-smart-${String(i + 1).padStart(3, '0')}`,
    robotName: `全智能${String(i + 1).padStart(3, '0')}`,
    robotType: 3,
    initX: 25 + (i % 5) * 16,
    initY: 12 + Math.floor(i / 5) * 15,
    phase: i * 0.62,
    batteryBase: 72 + (i % 28),
  };
}

function makeMockRobot(
  tick: number,
  seed: MockRobotSeed,
  cmdState?: MockCommandState,
): RobotRealtime {
  const t = tick / 10;
  const batteryLevel = Math.max(12, Math.round(seed.batteryBase - ((tick + seed.phase * 100) % 280) * 0.16));
  const cleanAreaBase = 860 + tick * (seed.robotType === 2 ? 4.8 : seed.robotType === 1 ? 3.2 : 2.6);

  let posX: number;
  let posY: number;
  let heading: number;
  let speed: number;
  let workStatus = 1;

  if (cmdState?.mode === 'stopped' && cmdState.frozenPosition) {
    return {
      robot_id: seed.robotId, robot_name: seed.robotName, robot_type: seed.robotType,
      online_status: 1, work_status: 0, battery_level: batteryLevel,
      pos_x: cmdState.frozenPosition.x, pos_y: cmdState.frozenPosition.y, pos_z: 0.3,
      heading: cmdState.frozenPosition.heading, speed: 0,
      clean_area: cleanAreaBase, temperature: 30 + Math.sin(t) * 2,
      humidity: 46 + Math.cos(t) * 8, light_intensity: 76000 + Math.sin(t * 0.4) * 8000,
      wind_speed: 2.2 + Math.cos(t * 0.7) * 0.5,
    };
  }

  if (cmdState?.mode === 'resetting') {
    workStatus = 0;
    speed = 0;
    heading = 0;
    posX = seed.initX;
    posY = seed.initY;
  } else if (cmdState?.mode === 'returning') {
    const dx = MOCK_BASE_POINT.x - seed.initX;
    const dy = MOCK_BASE_POINT.y - seed.initY;
    const progress = Math.min(1, (tick - cmdState.commandAt) / 40);
    posX = seed.initX + dx * progress;
    posY = seed.initY + dy * progress;
    heading = (Math.atan2(dy, dx) * 180) / Math.PI;
    speed = 0.85;
    workStatus = progress >= 1 ? 2 : 1;
  } else if (seed.robotType === 1) {
    // 固定式：严格沿面板行 X 轴往复清扫，Y 锁定在面板行中心
    const wave = Math.sin(t * 0.72 + seed.phase);
    heading = Math.cos(t * 0.72 + seed.phase) >= 0 ? 90 : 270;
    posX = clamp(seed.initX + wave * 20, 12, 108);
    posY = seed.initY; // 锁定面板行，不偏离
    speed = 0.75;
  } else if (seed.robotType === 2) {
    // 接驳车：Y 轴上下穿行于面板列间通道，X 锁定在通道
    const wave = Math.sin(t * 0.55 + seed.phase);
    heading = Math.cos(t * 0.55 + seed.phase) >= 0 ? 180 : 0;
    posX = seed.initX;
    posY = clamp(seed.initY + wave * 26, 7, 62);
    speed = 1.05;
  } else {
    // 全智能：Lissajous 自由巡航，以整个面板区为边界
    const ax = 35;
    const ay = 28;
    const xPhase = t * 0.42 + seed.phase;
    const yPhase = t * 0.31 + seed.phase * 1.3;
    posX = clamp(seed.initX + Math.sin(xPhase) * ax + Math.sin(t * 0.17 + seed.phase) * 6, 9, 111);
    posY = clamp(seed.initY + Math.sin(yPhase) * ay, 5, 65);
    // 用 t+0.1 近似计算朝向
    const nextX = clamp(seed.initX + Math.sin((t + 0.1) * 0.42 + seed.phase) * ax + Math.sin((t + 0.1) * 0.17 + seed.phase) * 6, 9, 111);
    const nextY = clamp(seed.initY + Math.sin((t + 0.1) * 0.31 + seed.phase * 1.3) * ay, 5, 65);
    heading = (Math.atan2(nextX - posX, nextY - posY) * 180) / Math.PI;
    speed = 0.68 + Math.sin(t * 0.21) * 0.22;
    workStatus = tick % 70 > 60 ? 2 : 1;
  }

  return {
    robot_id: seed.robotId, robot_name: seed.robotName, robot_type: seed.robotType,
    online_status: 1, work_status: workStatus, battery_level: batteryLevel,
    pos_x: posX, pos_y: posY, pos_z: 0.3, heading, speed,
    clean_area: cleanAreaBase, temperature: 30 + Math.sin(t + seed.phase) * 2,
    humidity: 46 + Math.cos(t + seed.phase) * 8,
    light_intensity: 76000 + Math.sin(t * 0.4 + seed.phase) * 8000,
    wind_speed: 2.2 + Math.cos(t * 0.7 + seed.phase) * 0.5,
  };
}

function makeMockRealtime(tick: number, cmdStates: Map<string, MockCommandState>): StationRealtime {
  const robots: RobotRealtime[] = [];
  for (let i = 0; i < MOCK_ROBOT_COUNT; i++) {
    const seed = makeMockRobotSeed(i);
    const cs = cmdStates.get(seed.robotId);
    robots.push(makeMockRobot(tick, seed, cs));
  }
  return {
    station_id: MOCK_STATION.station_id, station_name: MOCK_STATION.station_name,
    longitude: MOCK_STATION.longitude, latitude: MOCK_STATION.latitude,
    robots,
  };
}

function makeRobotDetailFromRealtime(robot: RobotRealtime): Robot {
  return {
    ...robot, robot_code: robot.robot_id.toUpperCase(),
    station_id: MOCK_STATION.station_id, station_name: MOCK_STATION.station_name,
    last_heartbeat: new Date().toISOString(), create_time: MOCK_STATION.create_time,
  };
}

function getMockSeedIndexByRobotId(robotId: string): number {
  const match = robotId.match(/mock-(fixed|shuttle|smart)-(\d+)/);
  if (!match) return 0;
  const num = parseInt(match[2], 10) - 1;
  if (match[1] === 'fixed') return num;
  if (match[1] === 'shuttle') return 10 + num;
  return 20 + num;
}

function makeMockTracks(robot: RobotRealtime | undefined, cmdStates: Map<string, MockCommandState>): RobotPosition[] {
  const target = robot ?? makeMockRobot(0, makeMockRobotSeed(0));
  const seedIndex = getMockSeedIndexByRobotId(target.robot_id);
  const seed = makeMockRobotSeed(seedIndex);
  const startTick = Math.max(0, mockTrackSeed(target.robot_id));
  return Array.from({ length: 90 }, (_, index) => {
    const point = makeMockRobot(startTick + index, seed, cmdStates.get(target.robot_id));
    return {
      robot_id: point.robot_id, pos_x: point.pos_x, pos_y: point.pos_y,
      pos_z: point.pos_z, heading: point.heading,
      timestamp: new Date(Date.now() - (90 - index) * 1000).toISOString(),
    };
  });
}

// ---------------------------------------------------------------------------
// Batch command helper
// ---------------------------------------------------------------------------

async function sendCommandInBatches(
  ids: string[], cmd: string, batchSize: number,
  onProgress: (done: number, total: number) => void,
): Promise<PromiseSettledResult<void>[]> {
  const results: PromiseSettledResult<void>[] = [];
  for (let i = 0; i < ids.length; i += batchSize) {
    const batch = ids.slice(i, i + batchSize);
    const settled = await Promise.allSettled(
      batch.map((id) => robotApi.sendCommand(id, cmd).then(() => undefined)),
    );
    results.push(...settled);
    onProgress(results.length, ids.length);
  }
  return results;
}

// ===== Robot list table columns =====

const ROBOT_LIST_COLUMNS = [
  { title: '名称', dataIndex: 'robot_name', key: 'robot_name', width: 140, ellipsis: true },
  { title: '类型', dataIndex: 'robot_type', key: 'robot_type', width: 80,
    render: (v: number) => <Tag>{ROBOT_TYPE_MAP[v] ?? '未知'}</Tag> },
  { title: '在线', dataIndex: 'online_status', key: 'online_status', width: 60,
    render: (v: number) => {
      const o = ONLINE_STATUS_MAP[v] ?? ONLINE_STATUS_MAP[0];
      return <Badge status={o.color === 'green' ? 'success' : 'default'} text="" />;
    } },
  { title: '工作状态', dataIndex: 'work_status', key: 'work_status', width: 90,
    render: (v: number) => {
      const w = WORK_STATUS_MAP[v] ?? WORK_STATUS_MAP[0];
      return <Tag color={w.color}>{w.text}</Tag>;
    } },
  { title: '电量', dataIndex: 'battery_level', key: 'battery_level', width: 70,
    render: (v: number) => <Progress percent={v} size="small" style={{ maxWidth: 60 }} format={(p) => `${p}%`} /> },
];

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export default function Monitor() {
  const { modal, message } = App.useApp();

  // ---- 电站列表 ----
  const [stations, setStations] = useState<Station[]>([]);
  const [stationsLoading, setStationsLoading] = useState(false);
  const [stationsError, setStationsError] = useState<string | null>(null);

  // ---- 选中电站 ----
  const [selectedStationId, setSelectedStationId] = useState<string | undefined>(undefined);
  const [mapMode, setMapMode] = useState<'gis' | 'scene'>('gis');
  const [mockEnabled, setMockEnabled] = useState(false);
  const mockTickRef = useRef(0);
  const mockCommandStateRef = useRef<Map<string, MockCommandState>>(new Map());
  const lastRealStationIdRef = useRef<string | undefined>(undefined);

  // ---- 实时数据 ----
  const [stationRealtime, setStationRealtime] = useState<StationRealtime | null>(null);
  const [realtimeLoading, setRealtimeLoading] = useState(false);
  const [realtimeError, setRealtimeError] = useState<string | null>(null);
  const [robotMarkers, setRobotMarkers] = useState<Map<string, RobotRealtime>>(new Map());

  // ---- 选择 ----
  const [selectedRobotId, setSelectedRobotId] = useState<string | undefined>(undefined);
  const [selectedRobotIds, setSelectedRobotIds] = useState<Set<string>>(new Set());
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [drawerMode, setDrawerMode] = useState<'station' | 'robot' | 'list'>('station');
  const [commandRunning, setCommandRunning] = useState(false);
  const [commandProgress, setCommandProgress] = useState({ done: 0, total: 0 });

  // ---- 机器人详情 ----
  const [robotDetail, setRobotDetail] = useState<Robot | null>(null);
  const [robotLoading, setRobotLoading] = useState(false);

  // ---- 轨迹 ----
  const [tracks, setTracks] = useState<RobotPosition[]>([]);
  const [tracksLoading, setTracksLoading] = useState(false);
  const [trackDateRange, setTrackDateRange] = useState<[Dayjs, Dayjs] | null>(null);
  const [activeRobotTab, setActiveRobotTab] = useState('basic');
  const [trackPlaybackIndex, setTrackPlaybackIndex] = useState(0);
  const [trackPlaybackRunning, setTrackPlaybackRunning] = useState(false);

  // ---- 环境 ----
  const [envHistory, setEnvHistory] = useState<unknown[]>([]);
  const [envLoading, setEnvLoading] = useState(false);

  // ---- WebSocket position batching ----
  const pendingPositionsRef = useRef<Map<string, Record<string, unknown>>>(new Map());
  const flushPositionTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // ---- Robot list table pagination ----
  const [robotListPage, setRobotListPage] = useState(1);
  const [robotListPageSize, setRobotListPageSize] = useState(20);

  const resetRobotContext = useCallback(() => {
    setSelectedRobotId(undefined);
    setSelectedRobotIds(new Set());
    setRobotDetail(null);
    setTracks([]);
    setTrackPlaybackIndex(0);
    setTrackPlaybackRunning(false);
    setActiveRobotTab('basic');
    setTrackDateRange(null);
    setEnvHistory([]);
    setRobotListPage(1);
    setDrawerOpen(false);
  }, []);

  const removeMockStation = useCallback(() => {
    setStations((prev) => prev.filter((s) => s.station_id !== MOCK_STATION.station_id));
    mockCommandStateRef.current.clear();
  }, []);

  // =====================================================================
  // 数据获取
  // =====================================================================

  const fetchStations = useCallback(async () => {
    setStationsLoading(true);
    setStationsError(null);
    try {
      const res = await stationApi.getAll();
      setStations(res ?? []);
    } catch (err: unknown) {
      setStationsError(err instanceof Error ? err.message : '获取电站列表失败');
    } finally {
      setStationsLoading(false);
    }
  }, []);

  useEffect(() => { fetchStations(); }, [fetchStations]);

  const fetchStationRealtime = useCallback(async (stationId: string) => {
    if (stationId === MOCK_STATION.station_id) return;
    setRealtimeLoading(true);
    setRealtimeError(null);
    try {
      const res = await monitorApi.getRealtime(stationId);
      setStationRealtime(res);
      const next = new Map<string, RobotRealtime>();
      if (res?.robots) {
        for (const r of res.robots) next.set(r.robot_id, r);
      }
      setRobotMarkers(next);
    } catch (err: unknown) {
      setRealtimeError(err instanceof Error ? err.message : '获取实时数据失败');
      setStationRealtime(null);
      setRobotMarkers(new Map());
    } finally {
      setRealtimeLoading(false);
    }
  }, []);

  const fetchRobotDetail = useCallback(async (robotId: string) => {
    setRobotLoading(true);
    try {
      const res = await robotApi.getById(robotId);
      setRobotDetail(res ?? null);
    } catch {
      message.error('获取机器人详情失败');
      setRobotDetail(null);
    } finally {
      setRobotLoading(false);
    }
  }, [message]);

  // =====================================================================
  // 事件处理
  // =====================================================================

  const handleStationSelect = useCallback((value: string) => {
    if (value === MOCK_STATION.station_id) {
      if (selectedStationId && selectedStationId !== MOCK_STATION.station_id) {
        lastRealStationIdRef.current = selectedStationId;
      }
      setMockEnabled(true);
      return;
    }

    lastRealStationIdRef.current = value;
    setMockEnabled(false);
    removeMockStation();
    setMapMode('gis');
    setSelectedStationId(value);
    setStationRealtime(null);
    setRobotMarkers(new Map());
    resetRobotContext();
    fetchStationRealtime(value);
  }, [fetchStationRealtime, removeMockStation, resetRobotContext, selectedStationId]);

  const handleClearStation = useCallback(() => {
    lastRealStationIdRef.current = undefined;
    setMockEnabled(false);
    removeMockStation();
    setMapMode('gis');
    setSelectedStationId(undefined);
    setStationRealtime(null);
    setRobotMarkers(new Map());
    resetRobotContext();
    setRealtimeError(null);
    setRealtimeLoading(false);
  }, [removeMockStation, resetRobotContext]);

  const handleMockSwitchChange = useCallback((checked: boolean) => {
    if (checked) {
      if (selectedStationId && selectedStationId !== MOCK_STATION.station_id) {
        lastRealStationIdRef.current = selectedStationId;
      }
      setMockEnabled(true);
      return;
    }

    const stationId = lastRealStationIdRef.current;
    if (stationId) {
      handleStationSelect(stationId);
    } else {
      handleClearStation();
    }
  }, [handleClearStation, handleStationSelect, selectedStationId]);

  const handleMapModeChange = useCallback((value: string | number) => {
    const nextMode = value as 'gis' | 'scene';
    if (nextMode === 'gis' && mockEnabled) {
      const stationId = lastRealStationIdRef.current;
      if (stationId) {
        handleStationSelect(stationId);
      } else {
        handleClearStation();
      }
      return;
    }
    setMapMode(nextMode);
  }, [handleClearStation, handleStationSelect, mockEnabled]);

  // ---- 模拟运行 effect ----
  useEffect(() => {
    if (!mockEnabled) return undefined;

    setStations((prev) =>
      prev.some((s) => s.station_id === MOCK_STATION.station_id) ? prev : [MOCK_STATION, ...prev]);
    setSelectedStationId(MOCK_STATION.station_id);
    setMapMode('scene');
    setRealtimeError(null);
    setRealtimeLoading(false);
    setTracks([]);
    setTrackPlaybackIndex(0);
    setTrackPlaybackRunning(false);
    setActiveRobotTab('basic');
    setTrackDateRange([dayjs().subtract(5, 'minute'), dayjs()]);

    const updateMockData = () => {
      const data = makeMockRealtime(mockTickRef.current, mockCommandStateRef.current);
      mockTickRef.current += 1;

      const robotArr = data.robots;
      const hasSelected = selectedRobotId && robotArr.some((r) => r.robot_id === selectedRobotId);
      const sid = hasSelected ? selectedRobotId : robotArr[0]?.robot_id;
      const sel = robotArr.find((r) => r.robot_id === sid) ?? robotArr[0];

      setStationRealtime(data);
      setRobotMarkers(new Map(robotArr.map((r) => [r.robot_id, r])));
      if (sid) setSelectedRobotId(sid);
      if (sel) setRobotDetail(makeRobotDetailFromRealtime(sel));

      // 清理 returning 已到达的机器人
      const cs = mockCommandStateRef.current;
      cs.forEach((state, robotId) => {
        if (state.mode === 'returning') {
          const robot = robotArr.find((r) => r.robot_id === robotId);
          if (robot) {
            const dx = MOCK_BASE_POINT.x - robot.pos_x;
            const dy = MOCK_BASE_POINT.y - robot.pos_y;
            if (Math.sqrt(dx * dx + dy * dy) < 0.5) {
              state.mode = 'running';
              state.lastCommand = undefined;
            }
          }
        }
        if (state.mode === 'resetting' && mockTickRef.current - state.commandAt > 20) {
          state.mode = 'running';
          state.lastCommand = undefined;
        }
      });
    };

    updateMockData();
    const timer = window.setInterval(updateMockData, 1000);
    return () => window.clearInterval(timer);
  }, [mockEnabled, selectedRobotId]);

  // ---- 机器人选择 ----
  const handleRobotSelect = useCallback(
    (robotId: string, options?: { additive?: boolean }) => {
      const additive = options?.additive ?? false;
      setSelectedRobotId(robotId);
      setTrackPlaybackIndex(0);
      setTrackPlaybackRunning(false);
      setActiveRobotTab('basic');
      setDrawerMode('robot');
      setDrawerOpen(true);

      if (additive) {
        setSelectedRobotIds((prev) => {
          const next = new Set(prev);
          if (next.has(robotId)) next.delete(robotId);
          else next.add(robotId);
          return next;
        });
      } else {
        setSelectedRobotIds(new Set([robotId]));
      }

      if (mockEnabled) {
        const robot = robotMarkers.get(robotId);
        if (robot) setRobotDetail(makeRobotDetailFromRealtime(robot));
        return;
      }
      fetchRobotDetail(robotId);
    },
    [fetchRobotDetail, mockEnabled, robotMarkers],
  );

  const handleSelectionChange = useCallback(
    (robotIds: string[], options?: { additive?: boolean }) => {
      if (options?.additive) {
        setSelectedRobotIds((prev) => {
          const next = new Set(prev);
          for (const id of robotIds) next.add(id);
          return next;
        });
      } else {
        setSelectedRobotIds(new Set(robotIds));
      }
    },
    [],
  );

  // ---- 远程指令 ----
  const handleCommand = useCallback(
    (cmd: CommandDef) => {
      const targetIds = selectedRobotIds.size > 0
        ? Array.from(selectedRobotIds)
        : selectedRobotId ? [selectedRobotId] : [];

      if (targetIds.length === 0) return;

      const plural = targetIds.length > 1 ? `(${targetIds.length}台)` : '';
      modal.confirm({
        title: '确认操作',
        content: `${cmd.confirmText} ${plural}`,
        okText: '确认', cancelText: '取消',
        okButtonProps: { danger: cmd.key === 'stop' },
        onOk: async () => {
          if (mockEnabled) {
            const cs = mockCommandStateRef.current;
            const tick = mockTickRef.current;
            for (const id of targetIds) {
              const robot = robotMarkers.get(id);
              const state: MockCommandState = {
                mode: cmd.key === 'start' ? 'running'
                  : cmd.key === 'stop' ? 'stopped'
                  : cmd.key === 'return' ? 'returning'
                  : 'resetting',
                lastCommand: cmd.key,
                commandAt: tick,
              };
              if (cmd.key === 'stop') {
                state.frozenPosition = robot
                  ? { x: robot.pos_x, y: robot.pos_y, heading: robot.heading }
                  : undefined;
              } else if (cmd.key === 'start') {
                state.frozenPosition = undefined;
              }
              cs.set(id, state);
            }
            message.success(`${cmd.label}模拟指令已执行 ${plural}`);
            return;
          }

          setCommandRunning(true);
          setCommandProgress({ done: 0, total: targetIds.length });
          try {
            const results = await sendCommandInBatches(
              targetIds, cmd.key, 20,
              (done, total) => setCommandProgress({ done, total }),
            );
            const ok = results.filter((r) => r.status === 'fulfilled').length;
            message.success(`${cmd.label}: ${ok}/${targetIds.length} 成功`);
          } catch {
            message.error(`${cmd.label}指令发送失败`);
          } finally {
            setCommandRunning(false);
          }
        },
      });
    },
    [selectedRobotId, selectedRobotIds, modal, message, mockEnabled, robotMarkers],
  );

  // ---- 轨迹查询 ----
  const handleFetchTracks = useCallback(async () => {
    if (!selectedRobotId || !trackDateRange) return;
    setTracksLoading(true);
    if (mockEnabled) {
      const nextTracks = makeMockTracks(robotMarkers.get(selectedRobotId), mockCommandStateRef.current);
      setTracks(nextTracks);
      setTrackPlaybackIndex(0);
      setTrackPlaybackRunning(true);
      setActiveRobotTab('tracks');
      setTracksLoading(false);
      return;
    }
    try {
      const [start, end] = trackDateRange;
      const res = await monitorApi.getTracks({
        robot_id: selectedRobotId,
        start_time: start.toISOString(), end_time: end.toISOString(), limit: 500,
      });
      const nextTracks = res ?? [];
      setTracks(nextTracks);
      setTrackPlaybackIndex(0);
      setTrackPlaybackRunning(nextTracks.length > 1);
      setActiveRobotTab('tracks');
    } catch {
      message.error('获取轨迹数据失败');
      setTracks([]);
      setTrackPlaybackIndex(0);
      setTrackPlaybackRunning(false);
    } finally {
      setTracksLoading(false);
    }
  }, [selectedRobotId, trackDateRange, message, mockEnabled, robotMarkers]);

  useEffect(() => {
    if (tracks.length <= 1 || !trackPlaybackRunning || activeRobotTab !== 'tracks') return undefined;
    const timer = window.setInterval(() => {
      setTrackPlaybackIndex((prev) => {
        if (prev >= tracks.length - 1) { setTrackPlaybackRunning(false); return prev; }
        return prev + 1;
      });
    }, 450);
    return () => window.clearInterval(timer);
  }, [tracks, trackPlaybackRunning, activeRobotTab]);

  // ---- 环境查询 ----
  const handleFetchEnv = useCallback(async () => {
    if (!selectedRobotId || !trackDateRange) return;
    setEnvLoading(true);
    if (mockEnabled) {
      const robot = robotMarkers.get(selectedRobotId);
      const nextEnv = Array.from({ length: 12 }, (_, index) => ({
        timestamp: new Date(Date.now() - (12 - index) * 60000).toISOString(),
        temperature: (robot?.temperature ?? 30) + Math.sin(index) * 1.2,
        humidity: (robot?.humidity ?? 48) + Math.cos(index) * 2.5,
      }));
      setEnvHistory(nextEnv);
      setEnvLoading(false);
      return;
    }
    try {
      const [start, end] = trackDateRange;
      const res = await monitorApi.getEnvironment({
        robot_id: selectedRobotId,
        start_time: start.toISOString(), end_time: end.toISOString(),
      });
      setEnvHistory(res ?? []);
    } catch {
      message.error('获取环境数据失败');
      setEnvHistory([]);
    } finally {
      setEnvLoading(false);
    }
  }, [selectedRobotId, trackDateRange, message, mockEnabled, robotMarkers]);

  // =====================================================================
  // WebSocket realtime (with position batching)
  // =====================================================================

  const wsHandlers: Partial<Record<WsMessageType, (msg: WsMessage) => void>> = useMemo(() => ({
    position(msg) {
      const d = msg.data as Record<string, unknown>;
      if (!d?.robot_id) return;
      pendingPositionsRef.current.set(String(d.robot_id), d);
      if (flushPositionTimerRef.current == null) {
        flushPositionTimerRef.current = setTimeout(() => {
          flushPositionTimerRef.current = null;
          const pending = pendingPositionsRef.current;
          if (pending.size === 0) return;
          pendingPositionsRef.current = new Map();
          setRobotMarkers((prev) => {
            const next = new Map(prev);
            pending.forEach((d2, rid) => {
              const cur = next.get(rid);
              if (!cur) return;
              next.set(rid, {
                ...cur,
                pos_x: (d2.pos_x as number) ?? cur.pos_x,
                pos_y: (d2.pos_y as number) ?? cur.pos_y,
                pos_z: (d2.pos_z as number) ?? cur.pos_z,
                heading: (d2.heading as number) ?? cur.heading,
                speed: (d2.speed as number) ?? cur.speed,
              });
            });
            return next;
          });
          if (selectedRobotId && pending.has(selectedRobotId)) {
            const d3 = pending.get(selectedRobotId)!;
            setRobotDetail((prev) => {
              if (!prev) return prev;
              return {
                ...prev,
                pos_x: (d3.pos_x as number) ?? prev.pos_x,
                pos_y: (d3.pos_y as number) ?? prev.pos_y,
                pos_z: (d3.pos_z as number) ?? prev.pos_z,
                heading: (d3.heading as number) ?? prev.heading,
                speed: (d3.speed as number) ?? prev.speed,
              };
            });
          }
        }, 100);
      }
    },

    heartbeat(msg) {
      const d = msg.data as Record<string, unknown>;
      if (!d?.robot_id) return;
      const rid = String(d.robot_id);
      setRobotMarkers((prev) => {
        const next = new Map(prev);
        const cur = next.get(rid);
        if (!cur) return prev;
        next.set(rid, { ...cur, online_status: (d.online_status as number) ?? cur.online_status });
        return next;
      });
      if (rid === selectedRobotId) {
        setRobotDetail((prev) => {
          if (!prev) return prev;
          return { ...prev, online_status: (d.online_status as number) ?? prev.online_status };
        });
      }
    },

    status(msg) {
      const d = msg.data as Record<string, unknown>;
      if (!d?.robot_id) return;
      const rid = String(d.robot_id);
      setRobotMarkers((prev) => {
        const next = new Map(prev);
        const cur = next.get(rid);
        if (!cur) return prev;
        next.set(rid, { ...cur, work_status: (d.work_status as number) ?? cur.work_status });
        return next;
      });
      if (rid === selectedRobotId) {
        setRobotDetail((prev) => {
          if (!prev) return prev;
          return { ...prev, work_status: (d.work_status as number) ?? prev.work_status };
        });
      }
    },

    clean(msg) {
      const d = msg.data as Record<string, unknown>;
      if (!d?.robot_id) return;
      const rid = String(d.robot_id);
      setRobotMarkers((prev) => {
        const next = new Map(prev);
        const cur = next.get(rid);
        if (!cur) return prev;
        next.set(rid, { ...cur, clean_area: (d.clean_area as number) ?? cur.clean_area });
        return next;
      });
      if (rid === selectedRobotId) {
        setRobotDetail((prev) => {
          if (!prev) return prev;
          return { ...prev, clean_area: (d.clean_area as number) ?? prev.clean_area };
        });
      }
    },

    alarm(msg) {
      const d = msg.data as Record<string, unknown>;
      message.warning(`[告警] ${d?.robot_id ?? ''}: ${d?.alarm_content ?? JSON.stringify(d)}`);
    },
  }), [message, selectedRobotId]);

  useWebSocket(selectedStationId, wsHandlers);

  // Cleanup flush timer on unmount
  useEffect(() => {
    return () => {
      if (flushPositionTimerRef.current != null) {
        clearTimeout(flushPositionTimerRef.current);
      }
    };
  }, []);

  // =====================================================================
  // 派生数据
  // =====================================================================

  const stationOptions = useMemo(() => {
    const list = mockEnabled && !stations.some((s) => s.station_id === MOCK_STATION.station_id)
      ? [MOCK_STATION, ...stations] : stations;
    return list.map((s) => ({ value: s.station_id, label: s.station_name }));
  }, [stations, mockEnabled]);

  const robotMarkerList = useMemo(
    () => Array.from(robotMarkers.values()).map((r) => {
      if (mockEnabled) {
        const ll = mockWorldToLatLng(r.pos_x, r.pos_y);
        return { ...toRobotMarker(r), pos_x: ll.longitude, pos_y: ll.latitude };
      }
      return toRobotMarker(r);
    }),
    [robotMarkers, mockEnabled],
  );

  const stationMarkerList = useMemo(() => {
    if (stationRealtime) return [toStationMarker(stationRealtime)];
    return [];
  }, [stationRealtime]);

  const mapCenter = useMemo((): [number, number] => {
    if (stationRealtime) return [stationRealtime.latitude, stationRealtime.longitude];
    return DEFAULT_CENTER;
  }, [stationRealtime]);

  const currentRobotRealtime = useMemo(() => {
    if (!selectedRobotId) return null;
    return robotMarkers.get(selectedRobotId) ?? null;
  }, [robotMarkers, selectedRobotId]);

  const stationRobotList = useMemo(() => Array.from(robotMarkers.values()), [robotMarkers]);

  const replayRobot = useMemo(() => {
    const detail = robotDetail;
    const rt = currentRobotRealtime;
    if (tracks.length === 0) return null;
    const currentTrack = tracks[Math.min(trackPlaybackIndex, tracks.length - 1)];
    if (!currentTrack) return null;
    return {
      robot_id: pick(rt?.robot_id, detail?.robot_id, ''),
      robot_name: pick(rt?.robot_name, detail?.robot_name, ''),
      robot_type: pick(rt?.robot_type, detail?.robot_type, 0),
      online_status: pick(rt?.online_status, detail?.online_status, 0),
      work_status: pick(rt?.work_status, detail?.work_status, 1),
      battery_level: pick(rt?.battery_level, detail?.battery_level, 0),
      pos_x: currentTrack.pos_x, pos_y: currentTrack.pos_y,
      heading: currentTrack.heading,
      speed: pick(rt?.speed, detail?.speed, 0),
      clean_area: pick(rt?.clean_area, detail?.clean_area, 0),
    };
  }, [tracks, trackPlaybackIndex, currentRobotRealtime, robotDetail]);

  const trackTrailPoints = useMemo(
    () => tracks.slice(0, Math.min(trackPlaybackIndex + 1, tracks.length)),
    [tracks, trackPlaybackIndex],
  );

  const onlineCount = stationRobotList.filter((r) => r.online_status === 1).length;

  // =====================================================================
  // 渲染：Drawer 内容
  // =====================================================================

  const renderDrawerContent = () => {
    if (drawerMode === 'list') return renderRobotList();
    if (drawerMode === 'robot' && selectedRobotId) return renderRobotDetail();
    return renderStationInfo();
  };

  const drawerTitle = () => {
    if (drawerMode === 'list') return '机器人列表';
    if (drawerMode === 'robot' && selectedRobotId) {
      const rt = currentRobotRealtime;
      const detail = robotDetail;
      return rt?.robot_name || detail?.robot_name || '机器人详情';
    }
    return '电站概览';
  };

  /** 电站概览 */
  const renderStationInfo = () => {
    const station = stationRealtime;
    return (
      <>
        {realtimeError ? (
          <Empty description={realtimeError}>
            <Button type="primary" onClick={() => selectedStationId && fetchStationRealtime(selectedStationId)}>
              重试
            </Button>
          </Empty>
        ) : station ? (
          <>
            <Descriptions column={1} size="small" bordered>
              <Descriptions.Item label="电站名称">{station.station_name}</Descriptions.Item>
              <Descriptions.Item label="电站ID">{station.station_id}</Descriptions.Item>
              <Descriptions.Item label="经度">{station.longitude.toFixed(6)}</Descriptions.Item>
              <Descriptions.Item label="纬度">{station.latitude.toFixed(6)}</Descriptions.Item>
              <Descriptions.Item label="在线机器人">{onlineCount} / {stationRobotList.length}</Descriptions.Item>
            </Descriptions>
            <Divider />
            <Button
              type="primary"
              icon={<UnorderedListOutlined />}
              onClick={() => setDrawerMode('list')}
              block
            >
              查看机器人列表 ({stationRobotList.length})
            </Button>
          </>
        ) : (
          <Empty description="暂无电站数据" />
        )}
      </>
    );
  };

  /** 机器人列表（Table） */
  const renderRobotList = () => {
    const dataSource = stationRobotList.map((r, idx) => ({ ...r, key: r.robot_id, _idx: idx }));
    const rowSelection = {
      selectedRowKeys: Array.from(selectedRobotIds),
      onChange: (keys: React.Key[]) => setSelectedRobotIds(new Set(keys as string[])),
    };
    return (
      <>
        <Space style={{ marginBottom: 12 }}>
          <Button size="small" onClick={() => setDrawerMode('station')}>
            返回电站概览
          </Button>
          {selectedRobotIds.size > 0 && (
            <Text type="secondary" style={{ fontSize: 12 }}>
              已选 {selectedRobotIds.size} 台
            </Text>
          )}
        </Space>
        <Table
          rowKey="robot_id"
          columns={ROBOT_LIST_COLUMNS}
          dataSource={dataSource}
          size="small"
          rowSelection={rowSelection}
          pagination={{
            current: robotListPage, pageSize: robotListPageSize, total: stationRobotList.length,
            showSizeChanger: true, pageSizeOptions: ['10', '20', '30', '50'],
            showTotal: (t) => `共 ${t} 台`,
            onChange: (p, s) => { setRobotListPage(p); setRobotListPageSize(s); },
          }}
          scroll={{ y: 'calc(100vh - 260px)' }}
          onRow={(record) => ({
            onClick: () => handleRobotSelect(record.robot_id),
            style: { cursor: 'pointer' },
          })}
          locale={{ emptyText: <Empty description="暂无机器人数据" /> }}
        />
      </>
    );
  };

  /** 机器人详情 */
  const renderRobotDetail = () => {
    const detail = robotDetail;
    const rt = currentRobotRealtime;
    if (!detail && !rt) return <Empty description="未找到机器人数据" />;
    if (robotLoading) return <div style={{ textAlign: 'center', padding: 60 }}><Spin /></div>;

    const r = {
      robot_id: pick(rt?.robot_id, detail?.robot_id, ''),
      robot_name: pick(rt?.robot_name, detail?.robot_name, ''),
      robot_type: pick(rt?.robot_type, detail?.robot_type, 0),
      online_status: pick(rt?.online_status, detail?.online_status, 0),
      work_status: pick(rt?.work_status, detail?.work_status, 0),
      battery_level: pick(rt?.battery_level, detail?.battery_level, 0),
      pos_x: pick(rt?.pos_x, detail?.pos_x, 0),
      pos_y: pick(rt?.pos_y, detail?.pos_y, 0),
      pos_z: pick(rt?.pos_z, detail?.pos_z, 0),
      heading: pick(rt?.heading, detail?.heading, 0),
      speed: pick(rt?.speed, detail?.speed, 0),
      clean_area: pick(rt?.clean_area, detail?.clean_area, 0),
      temperature: pick(rt?.temperature, detail?.temperature, 0),
      humidity: pick(rt?.humidity, detail?.humidity, 0),
      light_intensity: pick(rt?.light_intensity, detail?.light_intensity, 0),
      wind_speed: pick(rt?.wind_speed, detail?.wind_speed, 0),
    };

    const online = ONLINE_STATUS_MAP[r.online_status] ?? ONLINE_STATUS_MAP[0];
    const work = WORK_STATUS_MAP[r.work_status] ?? WORK_STATUS_MAP[0];
    const robotTypeName = ROBOT_TYPE_MAP[r.robot_type] ?? '未知';
    const batteryStatus = r.battery_level < 20 ? 'exception' : r.battery_level < 50 ? 'normal' : 'success';
    const isTrackView = activeRobotTab === 'tracks' && replayRobot;
    const sceneRobot = isTrackView ? replayRobot : r;
    const sceneFooterLabel = isTrackView
      ? `轨迹回放 ${trackPlaybackIndex + 1}/${tracks.length} · ${dayjs(trackTrailPoints[trackTrailPoints.length - 1]?.timestamp).format('MM-DD HH:mm:ss')}`
      : `实时监控 · 朝向 ${r.heading.toFixed(1)}°`;

    const tabItems: TabsProps['items'] = [
      {
        key: 'basic', label: '基本信息',
        children: (
          <Descriptions column={1} size="small" bordered>
            <Descriptions.Item label="机器人ID"><Text copyable style={{ fontSize: 12 }}>{r.robot_id}</Text></Descriptions.Item>
            <Descriptions.Item label="机器人名称">{r.robot_name}</Descriptions.Item>
            <Descriptions.Item label="机器人类型"><Tag>{robotTypeName}</Tag></Descriptions.Item>
            <Descriptions.Item label="在线状态">
              <Badge status={online.color === 'green' ? 'success' : 'default'} text={online.text} />
            </Descriptions.Item>
            <Descriptions.Item label="工作状态"><Tag color={work.color}>{work.text}</Tag></Descriptions.Item>
            <Descriptions.Item label="电池电量">
              <Progress percent={r.battery_level} size="small" status={batteryStatus}
                format={(p) => `${p}%`} style={{ maxWidth: 200 }} />
            </Descriptions.Item>
          </Descriptions>
        ),
      },
      {
        key: 'realtime', label: '实时数据',
        children: (
          <Descriptions column={1} size="small" bordered>
            <Descriptions.Item label="X坐标">{r.pos_x.toFixed(4)}</Descriptions.Item>
            <Descriptions.Item label="Y坐标">{r.pos_y.toFixed(4)}</Descriptions.Item>
            <Descriptions.Item label="Z坐标">{r.pos_z.toFixed(4)}</Descriptions.Item>
            <Descriptions.Item label="朝向">{r.heading.toFixed(1)}°</Descriptions.Item>
            <Descriptions.Item label="速度"><Text strong>{r.speed.toFixed(2)}</Text><Text type="secondary"> m/s</Text></Descriptions.Item>
            <Descriptions.Item label="清扫面积">
              <Text strong style={{ color: '#1677ff', fontSize: 15 }}>{r.clean_area.toFixed(2)}</Text>
              <Text type="secondary"> m²</Text>
            </Descriptions.Item>
          </Descriptions>
        ),
      },
      {
        key: 'env', label: '环境参数',
        children: (
          <div>
            <Descriptions column={1} size="small" bordered>
              <Descriptions.Item label="温度">
                <Text strong style={{ color: r.temperature > 40 ? '#ff4d4f' : r.temperature < 0 ? '#1677ff' : '#333' }}>
                  {r.temperature.toFixed(1)}
                </Text>
                <Text type="secondary"> °C</Text>
              </Descriptions.Item>
              <Descriptions.Item label="湿度">
                <Progress percent={Math.min(r.humidity, 100)} size="small" strokeColor="#1890ff"
                  format={(p) => `${p?.toFixed(0)}%`} style={{ maxWidth: 200 }} />
              </Descriptions.Item>
              <Descriptions.Item label="光照强度">
                <Text strong>{r.light_intensity.toFixed(0)}</Text><Text type="secondary"> lux</Text>
              </Descriptions.Item>
              <Descriptions.Item label="风速">
                <Text strong>{r.wind_speed.toFixed(1)}</Text><Text type="secondary"> m/s</Text>
              </Descriptions.Item>
            </Descriptions>
            <Divider style={{ fontSize: 12, marginTop: 16 }}>环境历史</Divider>
            <Space direction="vertical" style={{ width: '100%' }}>
              <Space>
                <RangePicker showTime size="small"
                  value={trackDateRange as [Dayjs, Dayjs] | null}
                  onChange={(dates) => setTrackDateRange(dates as [Dayjs, Dayjs] | null)}
                  presets={DATE_PRESETS} style={{ width: 220 }} />
                <Button size="small" onClick={handleFetchEnv} loading={envLoading} disabled={!trackDateRange}>
                  查询
                </Button>
              </Space>
              {envHistory.length > 0 ? (
                <Table dataSource={envHistory as Record<string, unknown>[]}
                  rowKey={(_, idx) => `env-${idx}`} size="small"
                  pagination={{ pageSize: 5, size: 'small' }} scroll={{ y: 150 }}
                  columns={[
                    { title: '时间', dataIndex: 'timestamp', render: (v: unknown) => dayjs(v as string).format('HH:mm:ss') },
                    { title: '温度', dataIndex: 'temperature', render: (v: unknown) => `${(v as number)?.toFixed(1) ?? '-'}°C` },
                    { title: '湿度', dataIndex: 'humidity', render: (v: unknown) => `${(v as number)?.toFixed(1) ?? '-'}%` },
                  ]} />
              ) : (
                <Empty description="暂无环境历史数据" image={Empty.PRESENTED_IMAGE_SIMPLE} style={{ marginTop: 8 }} />
              )}
            </Space>
          </div>
        ),
      },
      {
        key: 'tracks', label: '轨迹回放',
        children: (
          <Space direction="vertical" style={{ width: '100%' }} size="middle">
            {replayRobot ? (
              <Robot2DScene robot={replayRobot} peers={stationRobotList}
                trailPoints={trackTrailPoints} footerLabel={sceneFooterLabel} />
            ) : null}
            <Space>
              <RangePicker showTime size="small"
                value={trackDateRange as [Dayjs, Dayjs] | null}
                onChange={(dates) => setTrackDateRange(dates as [Dayjs, Dayjs] | null)}
                presets={DATE_PRESETS} style={{ width: 220 }} />
              <Button type="primary" size="small" onClick={handleFetchTracks}
                loading={tracksLoading} disabled={!trackDateRange}>
                查询轨迹
              </Button>
              <Button size="small" onClick={() => setTrackPlaybackRunning((prev) => !prev)}
                disabled={tracks.length <= 1}>
                {trackPlaybackRunning ? '暂停回放' : '播放回放'}
              </Button>
              <Button size="small" onClick={() => { setTrackPlaybackIndex(0); setTrackPlaybackRunning(false); }}
                disabled={tracks.length === 0}>
                回到起点
              </Button>
            </Space>
            {tracks.length > 0 ? (
              <Slider min={0} max={tracks.length - 1} step={1} value={trackPlaybackIndex}
                onChange={(value) => { setTrackPlaybackRunning(false); setTrackPlaybackIndex(value); }}
                tooltip={{ formatter: (value) => {
                  const point = tracks[value ?? 0];
                  return point ? dayjs(point.timestamp).format('MM-DD HH:mm:ss') : '';
                }}} />
            ) : null}
            {tracks.length > 0 ? (
              <Table dataSource={tracks} columns={TRACK_COLUMNS}
                rowKey={(_, idx) => `trk-${idx}`} size="small"
                pagination={{ pageSize: 8, size: 'small', showSizeChanger: false }}
                scroll={{ y: 240 }} />
            ) : (
              <Empty description="请选择时间范围后点击「查询轨迹」" image={Empty.PRESENTED_IMAGE_SIMPLE} />
            )}
          </Space>
        ),
      },
    ];

    return (
      <>
        <Space style={{ marginBottom: 12 }}>
          <Button size="small" onClick={() => setDrawerMode('station')}>返回电站概览</Button>
          <Button size="small" onClick={() => setDrawerMode('list')}>机器人列表</Button>
        </Space>
        <div style={{ marginBottom: 12 }}>
          <Robot2DScene robot={sceneRobot} peers={stationRobotList}
            trailPoints={isTrackView ? trackTrailPoints : undefined}
            footerLabel={sceneFooterLabel}
            style={{ marginBottom: 12 }} />
        </div>
        <Tabs items={tabItems} activeKey={activeRobotTab}
          onChange={(key) => { setActiveRobotTab(key); if (key !== 'tracks') setTrackPlaybackRunning(false); }}
          size="small" />
        <Divider style={{ margin: '8px 0 12px' }} />
        <div>
          <Text type="secondary" style={{ display: 'block', marginBottom: 8, fontSize: 12 }}>
            远程控制 {selectedRobotIds.size > 0 ? `(${selectedRobotIds.size}台)` : ''}
          </Text>
          {commandRunning && (
            <Progress percent={Math.round((commandProgress.done / commandProgress.total) * 100)}
              size="small" style={{ marginBottom: 8 }}
              format={() => `${commandProgress.done}/${commandProgress.total}`} />
          )}
          <Space wrap size={[8, 8]}>
            {ROBOT_COMMANDS.map((cmd) => (
              <Button key={cmd.key} size="small" icon={cmd.icon}
                style={{ borderColor: cmd.color, color: cmd.color }}
                onClick={() => handleCommand(cmd)}
                loading={commandRunning}>
                {cmd.label}
              </Button>
            ))}
          </Space>
        </div>
      </>
    );
  };

  // =====================================================================
  // 主布局
  // =====================================================================

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%', padding: 16, boxSizing: 'border-box' }}>
      {/* 工具栏 */}
      <Card size="small" style={{ marginBottom: 12, flexShrink: 0 }}>
        <Space wrap>
          <Text strong style={{ whiteSpace: 'nowrap' }}>选择电站：</Text>
          <Select showSearch allowClear placeholder="请选择电站"
            style={{ width: 280 }} value={selectedStationId}
            onChange={handleStationSelect} onClear={handleClearStation}
            loading={stationsLoading}
            filterOption={(input, option) =>
              (option?.label as string)?.toLowerCase().includes(input.toLowerCase())}
            options={stationOptions}
            notFoundContent={stationsLoading ? <Spin size="small" />
              : <Empty description="暂无电站" image={Empty.PRESENTED_IMAGE_SIMPLE} />} />

          <Space size={6}>
            <Text type="secondary" style={{ fontSize: 12 }}>模拟运行</Text>
            <Switch size="small" checked={mockEnabled}
              onChange={handleMockSwitchChange} />
          </Space>

          {selectedStationId && (
            <Button icon={<ReloadOutlined />} onClick={() => fetchStationRealtime(selectedStationId)}
              loading={realtimeLoading} disabled={mockEnabled}>刷新</Button>
          )}

          {selectedStationId && (
            <Segmented size="small" value={mapMode}
              onChange={handleMapModeChange}
              options={[{ label: 'GIS地图', value: 'gis' }, { label: '二维场景', value: 'scene' }]} />
          )}

          {stationsError && <Text type="danger" style={{ fontSize: 12 }}>{stationsError}</Text>}

          {/* 状态概览 */}
          {stationRealtime && (
            <>
              <Divider type="vertical" />
              <Text type="secondary" style={{ fontSize: 12 }}>
                在线 {onlineCount} / 共 {stationRobotList.length} 台
              </Text>
            </>
          )}

          {/* Drawer 切换按钮（未选中机器人时） */}
          {selectedStationId && (
            <>
              <Divider type="vertical" />
              <Button size="small" icon={<UnorderedListOutlined />}
                onClick={() => { setDrawerMode('list'); setDrawerOpen(true); }}>
                机器人列表
              </Button>
              <Button size="small" icon={<EnvironmentOutlined />}
                onClick={() => { setDrawerMode('station'); setDrawerOpen(true); }}>
                电站概览
              </Button>
            </>
          )}
        </Space>
      </Card>

      {/* 地图容器 - 全宽 */}
      <div style={{
        flex: 1, borderRadius: 8, overflow: 'hidden', position: 'relative',
        minHeight: 560, border: '1px solid #f0f0f0',
        height: 'calc(100vh - 120px)',
      }}>
        {/* 加载指示器 */}
        {realtimeLoading && (
          <div style={{
            position: 'absolute', top: '50%', left: '50%', transform: 'translate(-50%, -50%)',
            zIndex: 10, background: 'rgba(255,255,255,0.85)', padding: '24px 32px',
            borderRadius: 8, boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
          }}>
            <Spin tip="加载地图数据..." />
          </div>
        )}

        {/* 错误状态 */}
        {!realtimeLoading && realtimeError && (
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%', background: '#fafafa' }}>
            <Empty description={realtimeError}>
              <Button type="primary" onClick={() => selectedStationId && fetchStationRealtime(selectedStationId)}>
                重试
              </Button>
            </Empty>
          </div>
        )}

        {/* GIS 地图 */}
        {!realtimeError && (!selectedStationId || mapMode === 'gis') && (
          <GisMap
            stations={stationMarkerList}
            robots={robotMarkerList}
            center={mapCenter}
            zoom={selectedStationId ? 16 : 5}
            selectedRobotId={selectedRobotId}
            selectedRobotIds={selectedRobotIds}
            onRobotSelect={handleRobotSelect}
            style={{ width: '100%', height: '100%' }}
          />
        )}

        {/* 2D 场景 */}
        {!realtimeError && selectedStationId && mapMode === 'scene' && (
          <Station2DScene
            stationName={stationRealtime?.station_name}
            robots={stationRobotList}
            selectedRobotId={selectedRobotId}
            selectedRobotIds={selectedRobotIds}
            onRobotSelect={handleRobotSelect}
            onSelectionChange={handleSelectionChange}
          />
        )}

        {/* 浮动状态条 */}
        {stationRealtime && (
          <div style={{
            position: 'absolute', bottom: 12, left: 12, zIndex: 10,
            background: 'rgba(0,0,0,0.65)', borderRadius: 6, padding: '4px 12px',
            color: '#fff', fontSize: 12, fontFamily: 'monospace',
            display: 'flex', gap: 16,
          }}>
            <span>总数 {stationRobotList.length}</span>
            <span>在线 {onlineCount}</span>
            {selectedRobotIds.size > 0 && <span>选中 {selectedRobotIds.size}</span>}
            <span>{mapMode === 'scene' ? '二维场景' : 'GIS地图'}</span>
          </div>
        )}
      </div>

      {/* Drawer */}
      <Drawer
        title={drawerTitle()}
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        width={460}
        destroyOnClose={false}
      >
        {renderDrawerContent()}
      </Drawer>
    </div>
  );
}
