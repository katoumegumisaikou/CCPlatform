import { useState, useEffect, useCallback, useMemo } from 'react';
import {
  Card,
  Descriptions,
  Tag,
  Progress,
  Tabs,
  Select,
  Button,
  DatePicker,
  Spin,
  Empty,
  Badge,
  Space,
  Divider,
  Table,
  Typography,
  App,
} from 'antd';
import type { TabsProps } from 'antd';
import {
  ReloadOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  HomeOutlined,
  UndoOutlined,
  EnvironmentOutlined,
  AimOutlined,
} from '@ant-design/icons';
import dayjs, { type Dayjs } from 'dayjs';
import GisMap from '../../components/GisMap';
import { useWebSocket } from '../../hooks/useWebSocket';
import { stationApi } from '../../api/stations';
import { monitorApi } from '../../api/monitor';
import { robotApi } from '../../api/robots';
import type {
  Station,
  StationRealtime,
  RobotRealtime,
  Robot,
  RobotPosition,
  WsMessage,
  WsMessageType,
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
    key: 'start',
    label: '启动清扫',
    icon: <PlayCircleOutlined />,
    color: '#52c41a',
    confirmText: '确认启动清扫任务？',
  },
  {
    key: 'stop',
    label: '停止',
    icon: <PauseCircleOutlined />,
    color: '#ff4d4f',
    confirmText: '确认停止当前任务？',
  },
  {
    key: 'return',
    label: '回仓',
    icon: <HomeOutlined />,
    color: '#faad14',
    confirmText: '确认让机器人返回仓库？',
  },
  {
    key: 'reset',
    label: '复位',
    icon: <UndoOutlined />,
    color: '#1677ff',
    confirmText: '确认复位机器人？',
  },
];

const DEFAULT_CENTER: [number, number] = [39.9, 116.4];

const TRACK_COLUMNS = [
  {
    title: '时间',
    dataIndex: 'timestamp',
    key: 'timestamp',
    width: 160,
    render: (v: string) => dayjs(v).format('MM-DD HH:mm:ss'),
  },
  {
    title: 'X (经度)',
    dataIndex: 'pos_x',
    key: 'pos_x',
    render: (v: number) => v.toFixed(4),
  },
  {
    title: 'Y (纬度)',
    dataIndex: 'pos_y',
    key: 'pos_y',
    render: (v: number) => v.toFixed(4),
  },
  {
    title: '朝向',
    dataIndex: 'heading',
    key: 'heading',
    render: (v: number) => `${v.toFixed(1)}°`,
  },
];

const DATE_PRESETS: { label: string; value: [Dayjs, Dayjs] }[] = [
  { label: '最近1小时', value: [dayjs().subtract(1, 'hour'), dayjs()] },
  { label: '最近6小时', value: [dayjs().subtract(6, 'hour'), dayjs()] },
  { label: '今天', value: [dayjs().startOf('day'), dayjs()] },
  { label: '最近3天', value: [dayjs().subtract(3, 'day').startOf('day'), dayjs()] },
];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** Convert RobotRealtime to the shape GisMap expects for robot markers */
function toRobotMarker(r: RobotRealtime) {
  return {
    robot_id: r.robot_id,
    robot_name: r.robot_name,
    pos_x: r.pos_x,
    pos_y: r.pos_y,
    heading: r.heading,
    online_status: r.online_status,
    work_status: r.work_status,
    battery_level: r.battery_level,
  };
}

/** Convert station to the shape GisMap expects for station markers */
function toStationMarker(s: { station_id: string; station_name: string; longitude: number; latitude: number }) {
  return {
    station_id: s.station_id,
    station_name: s.station_name,
    longitude: s.longitude,
    latitude: s.latitude,
  };
}

/** Pick the best available value: realtime first, then detail, then default */
function pick<T extends string | number>(realtimeVal: T | undefined | null, detailVal: T | undefined | null, fallback: T): T {
  if (realtimeVal != null) return realtimeVal;
  if (detailVal != null) return detailVal;
  return fallback;
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export default function Monitor() {
  const { modal, message } = App.useApp();

  // ---- 电站列表 ----
  const [stations, setStations] = useState<Station[]>([]);
  const [stationsLoading, setStationsLoading] = useState(false);
  const [stationsError, setStationsError] = useState<string | null>(null);

  // ---- 当前选中的电站 ----
  const [selectedStationId, setSelectedStationId] = useState<string | undefined>(undefined);

  // ---- 电站实时数据 ----
  const [stationRealtime, setStationRealtime] = useState<StationRealtime | null>(null);
  const [realtimeLoading, setRealtimeLoading] = useState(false);
  const [realtimeError, setRealtimeError] = useState<string | null>(null);

  // ---- 机器人在线标记(本地Map，用于WebSocket增量更新，避免全量重取) ----
  const [robotMarkers, setRobotMarkers] = useState<Map<string, RobotRealtime>>(new Map());

  // ---- 选中的机器人 ----
  const [selectedRobotId, setSelectedRobotId] = useState<string | undefined>(undefined);

  // ---- 机器人详情 ----
  const [robotDetail, setRobotDetail] = useState<Robot | null>(null);
  const [robotLoading, setRobotLoading] = useState(false);

  // ---- 轨迹回放 ----
  const [tracks, setTracks] = useState<RobotPosition[]>([]);
  const [tracksLoading, setTracksLoading] = useState(false);
  const [trackDateRange, setTrackDateRange] = useState<[Dayjs, Dayjs] | null>(null);

  // ---- 环境历史数据 ----
  const [envHistory, setEnvHistory] = useState<unknown[]>([]);
  const [envLoading, setEnvLoading] = useState(false);

  // =====================================================================
  // 数据获取
  // =====================================================================

  /** 获取全部电站列表 */
  const fetchStations = useCallback(async () => {
    setStationsLoading(true);
    setStationsError(null);
    try {
      const res = await stationApi.getAll();
      setStations(res ?? []);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : '获取电站列表失败';
      setStationsError(msg);
    } finally {
      setStationsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchStations();
  }, [fetchStations]);

  /** 获取指定电站的实时数据(含机器人列表) */
  const fetchStationRealtime = useCallback(async (stationId: string) => {
    setRealtimeLoading(true);
    setRealtimeError(null);
    try {
      const res = await monitorApi.getRealtime(stationId);
      const data = res;
      setStationRealtime(data);

      // 用 API 返回的机器人数据初始化本地标记 Map
      const next = new Map<string, RobotRealtime>();
      if (data?.robots) {
        for (const r of data.robots) {
          next.set(r.robot_id, r);
        }
      }
      setRobotMarkers(next);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : '获取实时数据失败';
      setRealtimeError(msg);
      setStationRealtime(null);
      setRobotMarkers(new Map());
    } finally {
      setRealtimeLoading(false);
    }
  }, []);

  /** 获取单个机器人详情(完整字段) */
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

  /** 选择电站 */
  const handleStationSelect = useCallback(
    (value: string) => {
      setSelectedStationId(value);
      setSelectedRobotId(undefined);
      setRobotDetail(null);
      setTracks([]);
      setTrackDateRange(null);
      fetchStationRealtime(value);
    },
    [fetchStationRealtime],
  );

  /** 清除电站选择 - 全部重置 */
  const handleClearStation = useCallback(() => {
    setSelectedStationId(undefined);
    setStationRealtime(null);
    setRobotMarkers(new Map());
    setSelectedRobotId(undefined);
    setRobotDetail(null);
    setTracks([]);
    setTrackDateRange(null);
    setRealtimeError(null);
  }, []);

  /** 点击地图上的机器人标记 */
  const handleRobotClick = useCallback(
    (robotId: string) => {
      setSelectedRobotId(robotId);
      fetchRobotDetail(robotId);
    },
    [fetchRobotDetail],
  );

  /** 向机器人发送远程指令 */
  const handleCommand = useCallback(
    (cmd: CommandDef) => {
      if (!selectedRobotId) return;
      modal.confirm({
        title: '确认操作',
        content: cmd.confirmText,
        okText: '确认',
        cancelText: '取消',
        okButtonProps: { danger: cmd.key === 'stop' },
        onOk: async () => {
          try {
            await robotApi.sendCommand(selectedRobotId, cmd.key);
            message.success(`${cmd.label}指令已发送`);
          } catch {
            message.error(`${cmd.label}指令发送失败`);
          }
        },
      });
    },
    [selectedRobotId, modal, message],
  );

  /** 查询轨迹 */
  const handleFetchTracks = useCallback(async () => {
    if (!selectedRobotId || !trackDateRange) return;
    setTracksLoading(true);
    try {
      const [start, end] = trackDateRange;
      const res = await monitorApi.getTracks({
        robot_id: selectedRobotId,
        start_time: start.toISOString(),
        end_time: end.toISOString(),
        limit: 500,
      });
      setTracks(res ?? []);
    } catch {
      message.error('获取轨迹数据失败');
      setTracks([]);
    } finally {
      setTracksLoading(false);
    }
  }, [selectedRobotId, trackDateRange, message]);

  /** 查询环境历史 */
  const handleFetchEnv = useCallback(async () => {
    if (!selectedRobotId || !trackDateRange) return;
    setEnvLoading(true);
    try {
      const [start, end] = trackDateRange;
      const res = await monitorApi.getEnvironment({
        robot_id: selectedRobotId,
        start_time: start.toISOString(),
        end_time: end.toISOString(),
      });
      setEnvHistory(res ?? []);
    } catch {
      message.error('获取环境数据失败');
      setEnvHistory([]);
    } finally {
      setEnvLoading(false);
    }
  }, [selectedRobotId, trackDateRange, message]);

  // =====================================================================
  // WebSocket 实时通讯
  // =====================================================================

  const wsHandlers: Partial<Record<WsMessageType, (msg: WsMessage) => void>> = {
    position(msg) {
      const d = msg.data as Record<string, unknown>;
      if (!d?.robot_id) return;
      const rid = String(d.robot_id);
      setRobotMarkers((prev) => {
        const next = new Map(prev);
        const cur = next.get(rid);
        if (!cur) return prev; // 不认识的机器人，忽略
        next.set(rid, {
          ...cur,
          pos_x: (d.pos_x as number) ?? cur.pos_x,
          pos_y: (d.pos_y as number) ?? cur.pos_y,
          pos_z: (d.pos_z as number) ?? cur.pos_z,
          heading: (d.heading as number) ?? cur.heading,
          speed: (d.speed as number) ?? cur.speed,
        });
        return next;
      });

      // 如果当前选中的机器人收到位置更新，同步 robotDetail 的位置字段
      if (rid === selectedRobotId) {
        setRobotDetail((prev) => {
          if (!prev) return prev;
          return {
            ...prev,
            pos_x: (d.pos_x as number) ?? prev.pos_x,
            pos_y: (d.pos_y as number) ?? prev.pos_y,
            pos_z: (d.pos_z as number) ?? prev.pos_z,
            heading: (d.heading as number) ?? prev.heading,
            speed: (d.speed as number) ?? prev.speed,
          };
        });
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
  };

  // 当 selectedStationId 变化时，useWebSocket 内部会重连
  useWebSocket(selectedStationId, wsHandlers);

  // =====================================================================
  // 派生数据
  // =====================================================================

  const stationOptions = useMemo(
    () => stations.map((s) => ({ value: s.station_id, label: s.station_name })),
    [stations],
  );

  const robotMarkerList = useMemo(
    () => Array.from(robotMarkers.values()).map(toRobotMarker),
    [robotMarkers],
  );

  const stationMarkerList = useMemo(() => {
    if (stationRealtime) {
      return [toStationMarker(stationRealtime)];
    }
    // 未选择电站时，地图上不显示电站标记（避免显示全局电站干扰视觉）
    return [];
  }, [stationRealtime]);

  const mapCenter = useMemo((): [number, number] => {
    if (stationRealtime) {
      return [stationRealtime.latitude, stationRealtime.longitude];
    }
    return DEFAULT_CENTER;
  }, [stationRealtime]);

  // 当前选中机器人的实时数据（来自 WebSocket 增量更新）
  const currentRobotRealtime = useMemo(() => {
    if (!selectedRobotId) return null;
    return robotMarkers.get(selectedRobotId) ?? null;
  }, [robotMarkers, selectedRobotId]);

  // 当前电站的机器人列表（来自实时数据 Map）
  const stationRobotList = useMemo(
    () => Array.from(robotMarkers.values()),
    [robotMarkers],
  );

  // =====================================================================
  // 渲染：右侧信息面板
  // =====================================================================

  const renderInfoPanel = () => {
    // 尚未选择电站
    if (!selectedStationId) {
      return (
        <Card>
          <Empty description="请先选择电站" />
        </Card>
      );
    }

    // 已选择电站但未选择机器人 → 展示电站概览 + 机器人列表
    if (!selectedRobotId) {
      return renderStationInfo();
    }

    // 正在加载机器人详情
    if (robotLoading) {
      return (
        <Card>
          <div style={{ display: 'flex', justifyContent: 'center', padding: 60 }}>
            <Spin tip="加载机器人详情..." />
          </div>
        </Card>
      );
    }

    // 已选择机器人 → 展示机器人详情
    return renderRobotDetail();
  };

  /** 电站概览 + 机器人列表 */
  const renderStationInfo = () => {
    const station = stationRealtime;

    return (
      <Card
        title={
          <Space>
            <EnvironmentOutlined />
            <span>电站概览</span>
          </Space>
        }
        loading={realtimeLoading}
        style={{ height: '100%' }}
      >
        {realtimeError ? (
          <Empty description={realtimeError}>
            <Button type="primary" onClick={() => fetchStationRealtime(selectedStationId!)}>
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
              <Descriptions.Item label="在线机器人">
                {stationRobotList.filter((r) => r.online_status === 1).length} / {stationRobotList.length}
              </Descriptions.Item>
            </Descriptions>

            <Divider style={{ fontSize: 13 }}>机器人列表</Divider>

            {stationRobotList.length === 0 ? (
              <Empty description="该电站暂无机器人" image={Empty.PRESENTED_IMAGE_SIMPLE} />
            ) : (
              <div style={{ maxHeight: 360, overflowY: 'auto', paddingRight: 4 }}>
                {stationRobotList.map((r) => {
                  const online = ONLINE_STATUS_MAP[r.online_status] ?? ONLINE_STATUS_MAP[0];
                  const work = WORK_STATUS_MAP[r.work_status] ?? WORK_STATUS_MAP[0];
                  return (
                    <Card
                      key={r.robot_id}
                      size="small"
                      hoverable
                      style={{ marginBottom: 8 }}
                      onClick={() => handleRobotClick(r.robot_id)}
                    >
                      <Space direction="vertical" style={{ width: '100%' }} size={4}>
                        <Space>
                          <Text strong>{r.robot_name}</Text>
                          <Badge
                            status={online.color === 'green' ? 'success' : 'default'}
                            text={online.text}
                          />
                        </Space>
                        <Space size={8}>
                          <Tag color={work.color}>{work.text}</Tag>
                          <Text type="secondary">{ROBOT_TYPE_MAP[r.robot_type] ?? '未知'}</Text>
                          <Text type="secondary" style={{ fontSize: 12 }}>
                            电量 {r.battery_level}%
                          </Text>
                        </Space>
                      </Space>
                    </Card>
                  );
                })}
              </div>
            )}
          </>
        ) : null}
      </Card>
    );
  };

  /** 机器人详情 - 含四个标签页 + 远程控制按钮 */
  const renderRobotDetail = () => {
    const detail = robotDetail;
    const rt = currentRobotRealtime;

    if (!detail && !rt) {
      return (
        <Card>
          <Empty description="未找到机器人数据" />
        </Card>
      );
    }

    // 合并 realtime (WebSocket) 优先，detail (API) 兜底
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

    // 电池 Progress 状态
    const batteryStatus = r.battery_level < 20 ? 'exception' : r.battery_level < 50 ? 'normal' : 'success';

    const tabItems: TabsProps['items'] = [
      // ===== Tab 1: 基本信息 =====
      {
        key: 'basic',
        label: '基本信息',
        children: (
          <Descriptions column={1} size="small" bordered>
            <Descriptions.Item label="机器人ID">
              <Text copyable style={{ fontSize: 12 }}>{r.robot_id}</Text>
            </Descriptions.Item>
            <Descriptions.Item label="机器人名称">{r.robot_name}</Descriptions.Item>
            <Descriptions.Item label="机器人类型">
              <Tag>{robotTypeName}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="在线状态">
              <Badge
                status={online.color === 'green' ? 'success' : 'default'}
                text={online.text}
              />
            </Descriptions.Item>
            <Descriptions.Item label="工作状态">
              <Tag color={work.color}>{work.text}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="电池电量">
              <Progress
                percent={r.battery_level}
                size="small"
                status={batteryStatus}
                format={(p) => `${p}%`}
                style={{ maxWidth: 200 }}
              />
            </Descriptions.Item>
          </Descriptions>
        ),
      },

      // ===== Tab 2: 实时数据 =====
      {
        key: 'realtime',
        label: '实时数据',
        children: (
          <Descriptions column={1} size="small" bordered>
            <Descriptions.Item label="X坐标">{r.pos_x.toFixed(4)}</Descriptions.Item>
            <Descriptions.Item label="Y坐标">{r.pos_y.toFixed(4)}</Descriptions.Item>
            <Descriptions.Item label="Z坐标">{r.pos_z.toFixed(4)}</Descriptions.Item>
            <Descriptions.Item label="朝向">{r.heading.toFixed(1)}°</Descriptions.Item>
            <Descriptions.Item label="速度">
              <Text strong>{r.speed.toFixed(2)}</Text>
              <Text type="secondary"> m/s</Text>
            </Descriptions.Item>
            <Descriptions.Item label="清扫面积">
              <Text strong style={{ color: '#1677ff', fontSize: 15 }}>
                {r.clean_area.toFixed(2)}
              </Text>
              <Text type="secondary"> m²</Text>
            </Descriptions.Item>
          </Descriptions>
        ),
      },

      // ===== Tab 3: 环境参数 =====
      {
        key: 'env',
        label: '环境参数',
        children: (
          <div>
            <Descriptions column={1} size="small" bordered>
              <Descriptions.Item label="温度">
                <Text
                  strong
                  style={{
                    color: r.temperature > 40 ? '#ff4d4f' : r.temperature < 0 ? '#1677ff' : '#333',
                  }}
                >
                  {r.temperature.toFixed(1)}
                </Text>
                <Text type="secondary"> °C</Text>
              </Descriptions.Item>
              <Descriptions.Item label="湿度">
                <Progress
                  percent={Math.min(r.humidity, 100)}
                  size="small"
                  strokeColor="#1890ff"
                  format={(p) => `${p?.toFixed(0)}%`}
                  style={{ maxWidth: 200 }}
                />
              </Descriptions.Item>
              <Descriptions.Item label="光照强度">
                <Text strong>{r.light_intensity.toFixed(0)}</Text>
                <Text type="secondary"> lux</Text>
              </Descriptions.Item>
              <Descriptions.Item label="风速">
                <Text strong>{r.wind_speed.toFixed(1)}</Text>
                <Text type="secondary"> m/s</Text>
              </Descriptions.Item>
            </Descriptions>

            <Divider style={{ fontSize: 12, marginTop: 16 }}>环境历史</Divider>

            <Space direction="vertical" style={{ width: '100%' }}>
              <Space>
                <RangePicker
                  showTime
                  size="small"
                  value={trackDateRange as [Dayjs, Dayjs] | null}
                  onChange={(dates) => setTrackDateRange(dates as [Dayjs, Dayjs] | null)}
                  presets={DATE_PRESETS}
                  style={{ width: 220 }}
                />
                <Button size="small" onClick={handleFetchEnv} loading={envLoading} disabled={!trackDateRange}>
                  查询
                </Button>
              </Space>

              {envHistory.length > 0 ? (
                <Table
                  dataSource={envHistory as Record<string, unknown>[]}
                  rowKey={(record, idx) => `${(record as Record<string, unknown>).timestamp ?? 'env'}-${idx}`}
                  size="small"
                  pagination={{ pageSize: 5, size: 'small' }}
                  scroll={{ y: 150 }}
                  columns={[
                    {
                      title: '时间',
                      dataIndex: 'timestamp',
                      render: (v: unknown) => dayjs(v as string).format('HH:mm:ss'),
                    },
                    {
                      title: '温度',
                      dataIndex: 'temperature',
                      render: (v: unknown) => `${(v as number)?.toFixed(1) ?? '-'}°C`,
                    },
                    {
                      title: '湿度',
                      dataIndex: 'humidity',
                      render: (v: unknown) => `${(v as number)?.toFixed(1) ?? '-'}%`,
                    },
                  ]}
                />
              ) : (
                <Empty
                  description="暂无环境历史数据"
                  image={Empty.PRESENTED_IMAGE_SIMPLE}
                  style={{ marginTop: 8 }}
                />
              )}
            </Space>
          </div>
        ),
      },

      // ===== Tab 4: 轨迹回放 =====
      {
        key: 'tracks',
        label: '轨迹回放',
        children: (
          <Space direction="vertical" style={{ width: '100%' }} size="middle">
            <Space>
              <RangePicker
                showTime
                size="small"
                value={trackDateRange as [Dayjs, Dayjs] | null}
                onChange={(dates) => setTrackDateRange(dates as [Dayjs, Dayjs] | null)}
                presets={DATE_PRESETS}
                style={{ width: 220 }}
              />
              <Button
                type="primary"
                size="small"
                onClick={handleFetchTracks}
                loading={tracksLoading}
                disabled={!trackDateRange}
              >
                查询轨迹
              </Button>
            </Space>

            {tracks.length > 0 ? (
              <Table
                dataSource={tracks}
                columns={TRACK_COLUMNS}
                rowKey={(record, idx) => `${record.timestamp}-${idx}`}
                size="small"
                pagination={{ pageSize: 8, size: 'small', showSizeChanger: false }}
                scroll={{ y: 240 }}
              />
            ) : (
              <Empty
                description="请选择时间范围后点击「查询轨迹」"
                image={Empty.PRESENTED_IMAGE_SIMPLE}
              />
            )}
          </Space>
        ),
      },
    ];

    return (
      <Card
        title={
          <Space>
            <AimOutlined />
            <Text strong style={{ fontSize: 15 }}>
              {r.robot_name}
            </Text>
            <Badge
              status={online.color === 'green' ? 'success' : 'default'}
              text={online.text}
            />
          </Space>
        }
        extra={
          <Button
            type="text"
            size="small"
            icon={<ReloadOutlined />}
            onClick={() => fetchRobotDetail(selectedRobotId!)}
            loading={robotLoading}
          />
        }
      >
        <Tabs
          items={tabItems}
          defaultActiveKey="basic"
          size="small"
        />

        <Divider style={{ margin: '8px 0 12px' }} />

        {/* 远程控制 */}
        <div>
          <Text type="secondary" style={{ display: 'block', marginBottom: 8, fontSize: 12 }}>
            远程控制
          </Text>
          <Space wrap size={[8, 8]}>
            {ROBOT_COMMANDS.map((cmd) => (
              <Button
                key={cmd.key}
                size="small"
                icon={cmd.icon}
                style={{ borderColor: cmd.color, color: cmd.color }}
                onClick={() => handleCommand(cmd)}
              >
                {cmd.label}
              </Button>
            ))}
          </Space>
        </div>
      </Card>
    );
  };

  // =====================================================================
  // 主布局
  // =====================================================================

  return (
    <div style={{ display: 'flex', height: '100%', gap: 16, padding: 16, boxSizing: 'border-box' }}>
      {/* ========== 左侧：地图区域 ========== */}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0 }}>
        {/* 工具栏 */}
        <Card size="small" style={{ marginBottom: 12, flexShrink: 0 }}>
          <Space wrap>
            <Text strong style={{ whiteSpace: 'nowrap' }}>
              选择电站：
            </Text>
            <Select
              showSearch
              allowClear
              placeholder="请选择电站"
              style={{ width: 280 }}
              value={selectedStationId}
              onChange={handleStationSelect}
              onClear={handleClearStation}
              loading={stationsLoading}
              filterOption={(input, option) =>
                (option?.label as string)?.toLowerCase().includes(input.toLowerCase())
              }
              options={stationOptions}
              notFoundContent={
                stationsLoading ? (
                  <Spin size="small" />
                ) : (
                  <Empty description="暂无电站" image={Empty.PRESENTED_IMAGE_SIMPLE} />
                )
              }
            />

            {selectedStationId && (
              <Button
                icon={<ReloadOutlined />}
                onClick={() => fetchStationRealtime(selectedStationId)}
                loading={realtimeLoading}
              >
                刷新
              </Button>
            )}

            {stationsError && (
              <Text type="danger" style={{ fontSize: 12 }}>
                {stationsError}
              </Text>
            )}

            {/* 状态概览小标签 */}
            {stationRealtime && (
              <>
                <Divider type="vertical" />
                <Text type="secondary" style={{ fontSize: 12 }}>
                  在线 {stationRobotList.filter((r) => r.online_status === 1).length} / 共{' '}
                  {stationRobotList.length} 台
                </Text>
              </>
            )}
          </Space>
        </Card>

        {/* 地图容器 */}
        <div
          style={{
            flex: 1,
            borderRadius: 8,
            overflow: 'hidden',
            position: 'relative',
            minHeight: 400,
            border: '1px solid #f0f0f0',
          }}
        >
          {/* 加载指示器 */}
          {realtimeLoading && (
            <div
              style={{
                position: 'absolute',
                top: '50%',
                left: '50%',
                transform: 'translate(-50%, -50%)',
                zIndex: 10,
                background: 'rgba(255,255,255,0.85)',
                padding: '24px 32px',
                borderRadius: 8,
                boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
              }}
            >
              <Spin tip="加载地图数据..." />
            </div>
          )}

          {/* 错误状态 */}
          {!realtimeLoading && realtimeError && (
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                height: '100%',
                background: '#fafafa',
              }}
            >
              <Empty description={realtimeError}>
                <Button
                  type="primary"
                  onClick={() => selectedStationId && fetchStationRealtime(selectedStationId)}
                >
                  重试
                </Button>
              </Empty>
            </div>
          )}

          {/* 地图 */}
          {!realtimeError && (
            <GisMap
              stations={stationMarkerList}
              robots={robotMarkerList}
              center={mapCenter}
              zoom={selectedStationId ? 15 : 5}
              selectedRobotId={selectedRobotId}
              onRobotClick={handleRobotClick}
              style={{ width: '100%', height: '100%' }}
            />
          )}
        </div>
      </div>

      {/* ========== 右侧：信息面板 ========== */}
      <div
        style={{
          width: 380,
          flexShrink: 0,
          overflowY: 'auto',
          maxHeight: 'calc(100vh - 112px)',
        }}
      >
        {renderInfoPanel()}
      </div>
    </div>
  );
}
