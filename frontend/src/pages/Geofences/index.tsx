import { useState, useEffect, useCallback, useRef } from 'react';
import {
  Card, Table, Button, Space, Tag, Modal, Form, Input, InputNumber,
  Select, Popconfirm, Typography, Tabs, Descriptions, Empty,
  Badge, Divider, Tooltip, Row, Col, App,
} from 'antd';
import type { TabsProps } from 'antd';
import {
  PlusOutlined, EditOutlined, DeleteOutlined, EyeOutlined,
  EnvironmentOutlined, AimOutlined, AlertOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import L from 'leaflet';
import 'leaflet/dist/leaflet.css';
import { geofenceApi } from '../../api/geofence';
import type { Geofence, GeofenceAlarm } from '../../types';

const { Text } = Typography;
const { TextArea } = Input;

const FENCE_TYPE_MAP: Record<number, string> = { 1: '圆形围栏', 2: '多边形围栏' };
const ACTION_TYPE_MAP: Record<number, string> = { 1: '禁止进入', 2: '禁止离开' };
const SCOPE_TYPE_MAP: Record<string, string> = { global: '全局', station: '电站', robot: '机器人' };
const TRIGGER_TYPE_MAP: Record<number, string> = { 1: '进入禁区', 2: '离开工作区' };

const ALARM_LEVEL_MAP: Record<number, { text: string; color: string }> = {
  1: { text: '提示', color: 'blue' },
  2: { text: '一般', color: 'orange' },
  3: { text: '严重', color: 'red' },
  4: { text: '紧急', color: '#ff0000' },
};

const ALARM_HANDLE_MAP: Record<number, { text: string; color: string }> = {
  0: { text: '未处理', color: 'error' },
  1: { text: '已确认', color: 'warning' },
  2: { text: '处理中', color: 'processing' },
  3: { text: '已完成', color: 'success' },
  4: { text: '已忽略', color: 'default' },
};

// ---- Leaflet 绘制辅助 ----

function createLeafletMap(container: HTMLDivElement, center: [number, number], zoom = 15): L.Map {
  const map = L.map(container, {
    center,
    zoom,
    zoomControl: true,
  });

  // 高德地图卫星图（国内可访问）
  L.tileLayer('https://webrd01.is.autonavi.com/appmaptile?lang=zh_cn&size=1&scale=1&style=8&x={x}&y={y}&z={z}', {
    attribution: '&copy; AutoNavi',
    maxZoom: 18,
    subdomains: ['webrd01', 'webrd02', 'webrd03', 'webrd04'],
  }).addTo(map);

  // 如果高德也不显示，备用切换到 OpenStreetMap
  // L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {...});

  return map;
}

// 颜色选项
const COLOR_OPTIONS = [
  { label: '红色', value: '#ff4d4f' },
  { label: '橙色', value: '#fa8c16' },
  { label: '金色', value: '#faad14' },
  { label: '绿色', value: '#52c41a' },
  { label: '蓝色', value: '#1677ff' },
  { label: '紫色', value: '#722ed1' },
];

export default function Geofences() {
  const { message } = App.useApp();

  // ---- 数据状态 ----
  const [fences, setFences] = useState<Geofence[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [keyword, setKeyword] = useState('');

  // ---- 围栏告警 ----
  const [alarms, setAlarms] = useState<GeofenceAlarm[]>([]);
  const [alarmsLoading, setAlarmsLoading] = useState(false);
  const [alarmsTotal, setAlarmsTotal] = useState(0);
  const [alarmPage, setAlarmPage] = useState(1);

  // ---- 弹窗 ----
  const [modalOpen, setModalOpen] = useState(false);
  const [modalTitle, setModalTitle] = useState('创建电子围栏');
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form] = Form.useForm();

  // ---- 详情弹窗 ----
  const [detailOpen, setDetailOpen] = useState(false);
  const [detailFence, setDetailFence] = useState<Geofence | null>(null);

  // ---- 地图绘制 ----
  const mapRef = useRef<L.Map | null>(null);
  const mapContainerRef = useRef<HTMLDivElement>(null);
  const drawnLayerRef = useRef<L.Circle | L.Polygon | null>(null);
  const [mapVisible, setMapVisible] = useState(false);

  // =====================================================================
  // 数据加载
  // =====================================================================

  const fetchFences = useCallback(async (p?: number, ps?: number) => {
    setLoading(true);
    try {
      const res = await geofenceApi.list({
        page: p ?? page, size: ps ?? pageSize, keyword: keyword || undefined,
      });
      setFences(res?.list ?? []);
      setTotal(res?.total ?? 0);
    } catch {
      message.error('获取围栏列表失败');
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, keyword, message]);

  const fetchAlarms = useCallback(async (p?: number) => {
    setAlarmsLoading(true);
    try {
      const res = await geofenceApi.listAlarms({ page: p ?? alarmPage, size: 10 });
      setAlarms(res?.list ?? []);
      setAlarmsTotal(res?.total ?? 0);
    } catch {
      message.error('获取围栏告警失败');
    } finally {
      setAlarmsLoading(false);
    }
  }, [alarmPage, message]);

  useEffect(() => { fetchFences(); }, [fetchFences]);
  useEffect(() => { fetchAlarms(); }, [fetchAlarms]);

  // =====================================================================
  // 地图绘制功能
  // =====================================================================

  const initMap = useCallback((center?: [number, number]): L.Map | undefined => {
    const container = mapContainerRef.current;
    if (!container) return undefined;

    // 清理旧地图
    if (mapRef.current) {
      mapRef.current.remove();
      mapRef.current = null;
    }

    const map = createLeafletMap(container, center ?? [39.9, 116.4], 16);
    mapRef.current = map;

    // 修复在弹窗中显示时的大小问题
    setTimeout(() => { map.invalidateSize(); }, 200);

    // 添加绘制控件
    const fenceType = form.getFieldValue('fence_type');

    if (fenceType === 1) {
      // 圆形 - 点击设置圆心，半径用输入框
      map.on('click', (e: L.LeafletMouseEvent) => {
        const { lat, lng } = e.latlng;
        form.setFieldsValue({ center_lat: lat, center_lng: lng });
        drawCircleOnMap(lat, lng, form.getFieldValue('radius') || 50);
      });
    } else {
      // 多边形 - 点击添加顶点
      const vertices: [number, number][] = [];
      map.on('click', (e: L.LeafletMouseEvent) => {
        vertices.push([e.latlng.lat, e.latlng.lng]);
        drawPolygonOnMap(vertices);
        form.setFieldsValue({
          points: JSON.stringify(vertices.map((v) => [v[1], v[0]])),
        });
      });
    }
    return map;
  }, [form]);

  const drawCircleOnMap = (lat: number, lng: number, radius: number) => {
    if (!mapRef.current) return;
    if (drawnLayerRef.current) {
      drawnLayerRef.current.remove();
    }
    const color = form.getFieldValue('color') || '#ff4d4f';
    drawnLayerRef.current = L.circle([lat, lng], {
      radius,
      color,
      fillColor: color,
      fillOpacity: 0.15,
      weight: 2,
    }).addTo(mapRef.current);
    mapRef.current.fitBounds(drawnLayerRef.current.getBounds().pad(0.2));
  };

  const drawPolygonOnMap = (vertices: [number, number][]) => {
    if (!mapRef.current) return;
    if (drawnLayerRef.current) {
      drawnLayerRef.current.remove();
    }
    if (vertices.length < 2) return;
    const color = form.getFieldValue('color') || '#ff4d4f';
    drawnLayerRef.current = L.polygon(vertices, {
      color,
      fillColor: color,
      fillOpacity: 0.15,
      weight: 2,
    }).addTo(mapRef.current);
    if (vertices.length >= 3) {
      mapRef.current.fitBounds(drawnLayerRef.current.getBounds().pad(0.2));
    }
  };

  const handleMapDraw = useCallback(() => {
    setMapVisible(true);
    // 地图初始化交给 useEffect(mapVisible) 处理
  }, []);

  // 当 mapVisible 变成 true 时，等待 DOM 渲染完成后初始化地图
  useEffect(() => {
    if (!mapVisible || !modalOpen) return;

    // 用 requestAnimationFrame 确保 DOM 已渲染
    const raf = requestAnimationFrame(() => {
      setTimeout(() => {
        const container = mapContainerRef.current;
        if (!container) return;

        // 清理旧地图
        if (mapRef.current) {
          mapRef.current.remove();
          mapRef.current = null;
        }

        const center: [number, number] = [
          form.getFieldValue('center_lat') || 39.9,
          form.getFieldValue('center_lng') || 116.4,
        ];
        const map = initMap(center);
        if (!map) return;

        // 恢复已绘制的围栏
        const fenceType = form.getFieldValue('fence_type');
        const color = form.getFieldValue('color') || '#ff4d4f';
        if (fenceType === 1 && form.getFieldValue('center_lat')) {
          const lat = form.getFieldValue('center_lat');
          const lng = form.getFieldValue('center_lng');
          const radius = form.getFieldValue('radius') || 50;
          drawnLayerRef.current = L.circle([lat, lng], {
            radius, color, fillColor: color, fillOpacity: 0.15, weight: 2,
          }).addTo(map);
          map.setView([lat, lng], 16);
        } else if (fenceType === 2 && form.getFieldValue('points')) {
          try {
            const pts: [number, number][] = JSON.parse(form.getFieldValue('points'));
            if (pts.length >= 3) {
              const latlngs = pts.map((p) => [p[1], p[0]] as [number, number]);
              drawnLayerRef.current = L.polygon(latlngs, {
                color, fillColor: color, fillOpacity: 0.15, weight: 2,
              }).addTo(map);
              map.fitBounds(drawnLayerRef.current.getBounds().pad(0.2));
            }
          } catch { /* ignore */ }
        }
      }, 100);
    });

    return () => cancelAnimationFrame(raf);
  }, [mapVisible, modalOpen, form, initMap]);

  // =====================================================================
  // CRUD 操作
  // =====================================================================

  const handleCreate = () => {
    setEditingId(null);
    setModalTitle('创建电子围栏');
    form.resetFields();
    form.setFieldsValue({ fence_type: 1, action_type: 1, alarm_level: 2, status: 1, color: '#ff4d4f', scope_type: 'global' });
    setMapVisible(false);
    setModalOpen(true);
  };

  const handleEdit = (fence: Geofence) => {
    setEditingId(fence.geofence_id);
    setModalTitle('编辑电子围栏');
    form.setFieldsValue(fence);
    setMapVisible(false);
    setModalOpen(true);
  };

  const handleDetail = (fence: Geofence) => {
    setDetailFence(fence);
    setDetailOpen(true);
  };

  const handleDelete = async (id: string) => {
    try {
      await geofenceApi.delete(id);
      message.success('围栏已删除');
      fetchFences();
    } catch {
      message.error('删除失败');
    }
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      if (editingId) {
        await geofenceApi.update(editingId, values);
        message.success('围栏已更新');
      } else {
        await geofenceApi.create(values);
        message.success('围栏已创建');
      }
      setModalOpen(false);
      fetchFences();
    } catch (err: unknown) {
      if (err && typeof err === 'object' && 'errorFields' in err) return; // 表单校验错误
      message.error('操作失败');
    }
  };

  const handleFenceTypeChange = (value: number) => {
    if (value === 2) {
      form.setFieldsValue({ center_lng: undefined, center_lat: undefined, radius: undefined });
    } else {
      form.setFieldsValue({ points: undefined });
    }
    // 清理已绘制图形
    if (drawnLayerRef.current) {
      drawnLayerRef.current.remove();
      drawnLayerRef.current = null;
    }
    if (mapRef.current) {
      mapRef.current.off('click');
    }
  };

  const handleRadiusChange = (value: number | null) => {
    const lat = form.getFieldValue('center_lat');
    const lng = form.getFieldValue('center_lng');
    if (lat && lng && value && value > 0) {
      drawCircleOnMap(lat, lng, value);
    }
  };

  // =====================================================================
  // 围栏表格列定义
  // =====================================================================

  const FENCE_COLUMNS = [
    { title: '围栏名称', dataIndex: 'fence_name', key: 'fence_name', width: 160, ellipsis: true },
    {
      title: '类型', dataIndex: 'fence_type', key: 'fence_type', width: 100,
      render: (v: number) => <Tag>{FENCE_TYPE_MAP[v] ?? '未知'}</Tag>,
    },
    {
      title: '动作', dataIndex: 'action_type', key: 'action_type', width: 100,
      render: (v: number) => (
        <Tag color={v === 1 ? 'red' : 'blue'}>{ACTION_TYPE_MAP[v] ?? '未知'}</Tag>
      ),
    },
    {
      title: '作用域', dataIndex: 'scope_type', key: 'scope_type', width: 80,
      render: (v: string) => <Tag>{SCOPE_TYPE_MAP[v] ?? v}</Tag>,
    },
    {
      title: '状态', dataIndex: 'status', key: 'status', width: 70,
      render: (v: number) => <Badge status={v === 1 ? 'success' : 'default'} text={v === 1 ? '启用' : '禁用'} />,
    },
    {
      title: '告警级别', dataIndex: 'alarm_level', key: 'alarm_level', width: 80,
      render: (v: number) => {
        const level = ALARM_LEVEL_MAP[v] ?? { text: '未知', color: 'default' };
        return <Tag color={level.color}>{level.text}</Tag>;
      },
    },
    {
      title: '描述', dataIndex: 'description', key: 'description', width: 200,
      ellipsis: true, render: (v: string) => v || '-',
    },
    {
      title: '创建时间', dataIndex: 'create_time', key: 'create_time', width: 170,
      render: (v: string) => dayjs(v).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '操作', key: 'action', width: 200, fixed: 'right' as const,
      render: (_: unknown, record: Geofence) => (
        <Space size="small">
          <Tooltip title="查看详情"><Button size="small" icon={<EyeOutlined />} onClick={() => handleDetail(record)} /></Tooltip>
          <Tooltip title="编辑"><Button size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)} /></Tooltip>
          <Popconfirm title="确认删除该围栏？" onConfirm={() => handleDelete(record.geofence_id)}>
            <Tooltip title="删除"><Button size="small" danger icon={<DeleteOutlined />} /></Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  // =====================================================================
  // 告警表格列定义
  // =====================================================================

  const ALARM_COLUMNS = [
    { title: '告警内容', dataIndex: 'alarm_content', key: 'alarm_content', width: 300, ellipsis: true },
    { title: '围栏名称', dataIndex: 'fence_name', key: 'fence_name', width: 120, ellipsis: true },
    { title: '机器人', dataIndex: 'robot_name', key: 'robot_name', width: 120 },
    {
      title: '触发类型', dataIndex: 'trigger_type', key: 'trigger_type', width: 100,
      render: (v: number) => <Tag color={v === 1 ? 'red' : 'blue'}>{TRIGGER_TYPE_MAP[v] ?? '未知'}</Tag>,
    },
    {
      title: '处理状态', dataIndex: 'handle_status', key: 'handle_status', width: 90,
      render: (v: number) => {
        const s = ALARM_HANDLE_MAP[v] ?? { text: '未知', color: 'default' };
        return <Badge status={s.color as 'error' | 'warning' | 'processing' | 'success' | 'default'} text={s.text} />;
      },
    },
    {
      title: '触发时间', dataIndex: 'alarm_time', key: 'alarm_time', width: 160,
      render: (v: string) => dayjs(v).format('YYYY-MM-DD HH:mm:ss'),
    },
  ];

  // =====================================================================
  // 详情弹窗
  // =====================================================================

  const renderDetail = () => {
    if (!detailFence) return <Empty description="未选择围栏" />;
    const f = detailFence;
    const fenceColor = f.color || '#ff4d4f';

    return (
      <div>
        <Descriptions column={1} size="small" bordered>
          <Descriptions.Item label="围栏名称"><Text strong>{f.fence_name}</Text></Descriptions.Item>
          <Descriptions.Item label="围栏类型">{FENCE_TYPE_MAP[f.fence_type]}</Descriptions.Item>
          <Descriptions.Item label="动作类型">
            <Tag color={f.action_type === 1 ? 'red' : 'blue'}>{ACTION_TYPE_MAP[f.action_type]}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="作用域">{SCOPE_TYPE_MAP[f.scope_type] ?? f.scope_type}</Descriptions.Item>
          {f.scope_id && <Descriptions.Item label="作用域ID">{f.scope_id}</Descriptions.Item>}
          <Descriptions.Item label="告警级别">
            <Tag color={ALARM_LEVEL_MAP[f.alarm_level]?.color}>{ALARM_LEVEL_MAP[f.alarm_level]?.text}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="状态">
            <Badge status={f.status === 1 ? 'success' : 'default'} text={f.status === 1 ? '启用' : '禁用'} />
          </Descriptions.Item>
          <Descriptions.Item label="地图颜色">
            <div style={{ width: 20, height: 20, background: fenceColor, borderRadius: 4, display: 'inline-block', verticalAlign: 'middle' }} />
          </Descriptions.Item>
          <Descriptions.Item label="描述">{f.description || '-'}</Descriptions.Item>
          <Descriptions.Item label="创建时间">{dayjs(f.create_time).format('YYYY-MM-DD HH:mm:ss')}</Descriptions.Item>
        </Descriptions>

        {f.fence_type === 1 && (
          <Card size="small" title="圆形围栏参数" style={{ marginTop: 16 }}>
            <Descriptions column={2} size="small">
              <Descriptions.Item label="圆心经度">{f.center_lng?.toFixed(6)}</Descriptions.Item>
              <Descriptions.Item label="圆心纬度">{f.center_lat?.toFixed(6)}</Descriptions.Item>
              <Descriptions.Item label="半径">{f.radius} 米</Descriptions.Item>
            </Descriptions>
          </Card>
        )}

        {f.fence_type === 2 && f.points && (
          <Card size="small" title="多边形顶点" style={{ marginTop: 16 }}>
            <Text code style={{ fontSize: 11, whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
              {f.points}
            </Text>
          </Card>
        )}

        <Divider />
        <div ref={mapContainerRef} style={{ width: '100%', height: 300, borderRadius: 6, border: '1px solid #f0f0f0' }} />
      </div>
    );
  };

  // 当详情弹窗打开时，渲染围栏到小地图
  useEffect(() => {
    if (!detailOpen || !detailFence || !mapContainerRef.current) return;

    // 清理旧地图
    if (mapRef.current) {
      mapRef.current.remove();
      mapRef.current = null;
    }
    if (drawnLayerRef.current) {
      drawnLayerRef.current = null;
    }

    const f = detailFence;
    const color = f.color || '#ff4d4f';

    if (f.fence_type === 1 && f.center_lat && f.center_lng) {
      const map = createLeafletMap(mapContainerRef.current, [f.center_lat, f.center_lng], 16);
      mapRef.current = map;
      L.circle([f.center_lat, f.center_lng], {
        radius: f.radius || 50, color, fillColor: color, fillOpacity: 0.15, weight: 2,
      }).addTo(map);
    } else if (f.fence_type === 2 && f.points) {
      try {
        const pts: [number, number][] = JSON.parse(f.points);
        if (pts.length >= 3) {
          const latlngs = pts.map((p) => [p[1], p[0]] as [number, number]);
          const map = createLeafletMap(mapContainerRef.current, latlngs[0], 16);
          mapRef.current = map;
          L.polygon(latlngs, {
            color, fillColor: color, fillOpacity: 0.15, weight: 2,
          }).addTo(map);
          map.fitBounds(L.latLngBounds(latlngs).pad(0.2));
        }
      } catch { /* ignore */ }
    }

    return () => {
      if (mapRef.current) {
        mapRef.current.remove();
        mapRef.current = null;
      }
    };
  }, [detailOpen, detailFence]);

  // 清理创建/编辑弹窗中的地图
  useEffect(() => {
    if (!modalOpen) {
      if (mapRef.current) {
        mapRef.current.remove();
        mapRef.current = null;
      }
      if (drawnLayerRef.current) {
        drawnLayerRef.current = null;
      }
    }
  }, [modalOpen]);

  // =====================================================================
  // 标签页配置
  // =====================================================================

  const tabItems: TabsProps['items'] = [
    {
      key: 'fences',
      label: <span><EnvironmentOutlined /> 围栏管理</span>,
      children: (
        <div>
          <Space style={{ marginBottom: 16 }}>
            <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>创建围栏</Button>
            <Input.Search
              placeholder="搜索围栏名称"
              allowClear
              style={{ width: 240 }}
              onSearch={(value) => { setKeyword(value); setPage(1); }}
            />
          </Space>
          <Table
            rowKey="geofence_id"
            columns={FENCE_COLUMNS}
            dataSource={fences}
            loading={loading}
            scroll={{ x: 1100 }}
            pagination={{
              current: page, pageSize, total,
              showSizeChanger: true, pageSizeOptions: ['10', '20', '50'],
              showTotal: (t) => `共 ${t} 条`,
              onChange: (p, ps) => { setPage(p); setPageSize(ps); },
            }}
            locale={{ emptyText: <Empty description="暂无电子围栏" image={Empty.PRESENTED_IMAGE_SIMPLE} /> }}
          />
        </div>
      ),
    },
    {
      key: 'alarms',
      label: <span><AlertOutlined /> 围栏告警 ({alarmsTotal})</span>,
      children: (
        <Table
          rowKey="alarm_id"
          columns={ALARM_COLUMNS}
          dataSource={alarms}
          loading={alarmsLoading}
          scroll={{ x: 900 }}
          pagination={{
            current: alarmPage, pageSize: 10, total: alarmsTotal,
            showTotal: (t) => `共 ${t} 条`,
            onChange: (p) => setAlarmPage(p),
          }}
          locale={{ emptyText: <Empty description="暂无围栏告警" image={Empty.PRESENTED_IMAGE_SIMPLE} /> }}
        />
      ),
    },
  ];

  // =====================================================================
  // 主布局
  // =====================================================================

  return (
    <div>
      <Tabs items={tabItems} />

      {/* 创建/编辑弹窗 */}
      <Modal
        title={modalTitle}
        open={modalOpen}
        onOk={handleSubmit}
        onCancel={() => setModalOpen(false)}
        width={720}
        okText="保存"
        cancelText="取消"
        destroyOnClose
      >
        <Form form={form} layout="vertical" size="small">
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="fence_name" label="围栏名称" rules={[{ required: true, message: '请输入围栏名称' }]}>
                <Input placeholder="例如：东区作业区" />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item name="fence_type" label="围栏类型" rules={[{ required: true }]}>
                <Select onChange={handleFenceTypeChange}
                  options={[
                    { value: 1, label: '圆形围栏' },
                    { value: 2, label: '多边形围栏' },
                  ]} />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item name="action_type" label="动作类型" rules={[{ required: true }]}>
                <Select options={[
                  { value: 1, label: '禁止进入' },
                  { value: 2, label: '禁止离开' },
                ]} />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={6}>
              <Form.Item name="scope_type" label="作用域">
                <Select options={[
                  { value: 'global', label: '全局' },
                  { value: 'station', label: '电站' },
                  { value: 'robot', label: '机器人' },
                ]} />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item name="scope_id" label="作用域ID">
                <Input placeholder="留空=全局" />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item name="alarm_level" label="告警级别">
                <Select options={[
                  { value: 1, label: '提示' },
                  { value: 2, label: '一般' },
                  { value: 3, label: '严重' },
                  { value: 4, label: '紧急' },
                ]} />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item name="status" label="状态">
                <Select options={[
                  { value: 1, label: '启用' },
                  { value: 0, label: '禁用' },
                ]} />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item name="description" label="描述">
            <TextArea rows={2} placeholder="围栏用途说明" />
          </Form.Item>

          <Form.Item name="color" label="地图颜色">
            <Select>
              {COLOR_OPTIONS.map((c) => (
                <Select.Option key={c.value} value={c.value}>
                  <Space>
                    <div style={{ width: 16, height: 16, background: c.value, borderRadius: 3, display: 'inline-block' }} />
                    {c.label}
                  </Space>
                </Select.Option>
              ))}
            </Select>
          </Form.Item>

          {/* 圆形围栏参数 */}
          <Form.Item noStyle shouldUpdate={(prev, curr) => prev.fence_type !== curr.fence_type}>
            {({ getFieldValue }) => getFieldValue('fence_type') === 1 ? (
              <Row gutter={16}>
                <Col span={8}>
                  <Form.Item name="center_lng" label="圆心经度">
                    <InputNumber style={{ width: '100%' }} placeholder="例如: 116.3974" step={0.0001} />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name="center_lat" label="圆心纬度">
                    <InputNumber style={{ width: '100%' }} placeholder="例如: 39.9092" step={0.0001} />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name="radius" label="半径（米）">
                    <InputNumber style={{ width: '100%' }} min={1} max={10000} placeholder="50" onChange={handleRadiusChange} />
                  </Form.Item>
                </Col>
              </Row>
            ) : (
              <Form.Item name="points" label="多边形顶点坐标 (JSON)">
                <TextArea rows={3} placeholder='[[116.397,39.909],[116.398,39.909],[116.398,39.910],[116.397,39.910]]' />
              </Form.Item>
            )}
          </Form.Item>

          <Divider />
          <Space>
            <Button icon={<AimOutlined />} onClick={handleMapDraw}>
              {mapVisible ? '刷新地图' : '打开地图绘制'}
            </Button>
            {mapVisible && (
              <Text type="secondary" style={{ fontSize: 12 }}>
                {form.getFieldValue('fence_type') === 1
                  ? '点击地图设置圆心位置'
                  : '依次点击地图添加多边形顶点'}
              </Text>
            )}
          </Space>
          {mapVisible && (
            <div
              ref={mapContainerRef}
              style={{
                width: '100%', height: 400, marginTop: 8,
                borderRadius: 6, border: '1px solid #d9d9d9',
                position: 'relative', zIndex: 1050,
              }}
            />
          )}
        </Form>
      </Modal>

      {/* 详情弹窗 */}
      <Modal
        title={`围栏详情 - ${detailFence?.fence_name ?? ''}`}
        open={detailOpen}
        onCancel={() => setDetailOpen(false)}
        footer={<Button onClick={() => setDetailOpen(false)}>关闭</Button>}
        width={640}
        destroyOnClose
      >
        {renderDetail()}
      </Modal>
    </div>
  );
}
