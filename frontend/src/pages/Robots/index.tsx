import { useState, useEffect, useMemo, type ReactNode } from 'react';
import {
  Table, Button, Modal, Form, Input, Select, Popconfirm, message, Space, Tag,
  Card, Row, Col, Typography, Progress, Descriptions,
} from 'antd';
import {
  PlusOutlined, EditOutlined, DeleteOutlined, ReloadOutlined, SearchOutlined,
  PlayCircleOutlined, PauseCircleOutlined, HomeOutlined, UndoOutlined, InfoCircleOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { robotApi } from '../../api/robots';
import { stationApi } from '../../api/stations';
import type { Robot, Station } from '../../types';
import {
  ROBOT_TYPE_MAP, ONLINE_STATUS_MAP, WORK_STATUS_MAP,
  ROBOT_TABLE_FIELDS, ROBOT_DEFAULT_TABLE_FIELDS, ROBOT_DETAIL_SECTIONS,
} from '../../utils';

const { Title, Text } = Typography;

const displayNumber = (value: number | null | undefined, suffix = '', digits?: number) => {
  if (value === undefined || value === null) return '-';
  return `${digits === undefined ? value : value.toFixed(digits)}${suffix}`;
};

const displayBool = (value: number | null | undefined, trueText = '是', falseText = '否') => {
  if (value === undefined || value === null) return '-';
  return value === 1 ? trueText : falseText;
};

const displayText = (value: string | null | undefined) => value || '-';

const displayDuration = (seconds: number | null | undefined) => {
  if (seconds === undefined || seconds === null) return '-';
  const safeSeconds = Math.max(0, seconds);
  const hours = Math.floor(safeSeconds / 3600);
  const minutes = Math.floor((safeSeconds % 3600) / 60);
  return hours > 0 ? `${hours}小时${minutes}分` : `${minutes}分`;
};

interface FilterParams {
  station_id?: string;
  robot_type?: number;
  online_status?: number;
}

type RobotColumn = ColumnsType<Robot>[number];

type DetailField = {
  label: string;
  value: ReactNode;
};

const textColumn = (
  title: string,
  dataIndex: keyof Robot,
  width = 120,
  ellipsis = false,
): RobotColumn => ({
  title,
  dataIndex,
  key: String(dataIndex),
  width,
  ellipsis,
  render: (value: string | null | undefined) => displayText(value),
});

const numberColumn = (
  title: string,
  dataIndex: keyof Robot,
  width = 110,
  suffix = '',
  digits?: number,
): RobotColumn => ({
  title,
  dataIndex,
  key: String(dataIndex),
  width,
  align: 'right',
  render: (value: number | null | undefined) => displayNumber(value, suffix, digits),
});

const boolColumn = (
  title: string,
  dataIndex: keyof Robot,
  width = 120,
  trueText = '是',
  falseText = '否',
): RobotColumn => ({
  title,
  dataIndex,
  key: String(dataIndex),
  width,
  align: 'center',
  render: (value: number | null | undefined) => displayBool(value, trueText, falseText),
});

const getRobotDetailSections = (robot: Robot): { key: string; title: string; fields: DetailField[] }[] => {
  const type = robot.robot_type;
  const visibleSections = ROBOT_DETAIL_SECTIONS[type] || ROBOT_DETAIL_SECTIONS[1];

  const allSections: { key: string; title: string; fields: DetailField[] }[] = [
    {
      key: 'basic',
      title: '基础信息',
      fields: [
        { label: '机器人ID', value: <Text copyable>{robot.robot_id}</Text> },
        { label: '机器人编码', value: robot.robot_code },
        { label: '机器人名称', value: robot.robot_name },
        { label: '类型', value: <Tag>{ROBOT_TYPE_MAP[type] || type}</Tag> },
        { label: '所属电站ID', value: robot.station_id },
        { label: '所属电站名称', value: displayText(robot.station_name) },
        { label: '创建时间', value: displayText(robot.create_time) },
        { label: '更新时间', value: displayText(robot.update_time) },
      ],
    },
    {
      key: 'status',
      title: '状态与电气',
      fields: [
        {
          label: '在线状态',
          value: (
            <Tag color={ONLINE_STATUS_MAP[robot.online_status]?.color}>
              {ONLINE_STATUS_MAP[robot.online_status]?.text || '未知'}
            </Tag>
          ),
        },
        {
          label: '工作状态',
          value: (
            <Tag color={WORK_STATUS_MAP[robot.work_status]?.color}>
              {WORK_STATUS_MAP[robot.work_status]?.text || '未知'}
            </Tag>
          ),
        },
        { label: '电量', value: displayNumber(robot.battery_level, '%') },
        { label: '电池电压', value: displayNumber(robot.battery_voltage, ' V', 1) },
        { label: '故障状态', value: displayBool(robot.fault_status, '异常', '正常') },
        { label: '故障码', value: displayNumber(robot.fault_code) },
        { label: '工作时段', value: displayNumber(robot.work_period) },
        { label: '综合信号强度', value: displayNumber(robot.signal_strength) },
        { label: '4G信号强度', value: displayNumber(robot.signal_4g_strength) },
        { label: '累计运行时长', value: displayDuration(robot.run_duration_seconds) },
        { label: '设备时间', value: displayText(robot.device_time) },
        { label: '最后心跳', value: displayText(robot.last_heartbeat) },
      ],
    },
    {
      key: 'position',
      title: '位置与GPS',
      fields: [
        { label: 'X坐标', value: displayNumber(robot.pos_x, ' m', 3) },
        { label: 'Y坐标', value: displayNumber(robot.pos_y, ' m', 3) },
        { label: 'Z坐标', value: displayNumber(robot.pos_z, ' m', 3) },
        { label: '朝向', value: displayNumber(robot.heading, '°', 2) },
        { label: 'GPS经度', value: displayNumber(robot.gps_longitude, '', 7) },
        { label: 'GPS纬度', value: displayNumber(robot.gps_latitude, '', 7) },
        { label: 'GPS海拔', value: displayNumber(robot.gps_altitude, ' m', 3) },
        { label: 'GPS定位精度', value: displayNumber(robot.gps_accuracy, ' m', 3) },
      ],
    },
    {
      key: 'environment',
      title: '作业与环境',
      fields: [
        { label: '速度', value: displayNumber(robot.speed, ' m/s', 2) },
        { label: '累计清扫面积', value: displayNumber(robot.clean_area, ' m²', 2) },
        { label: '温度', value: displayNumber(robot.temperature, '°C', 1) },
        { label: '湿度', value: displayNumber(robot.humidity, '%', 1) },
        { label: '光照强度', value: displayNumber(robot.light_intensity, ' lux', 2) },
        { label: '风速', value: displayNumber(robot.wind_speed, ' m/s', 2) },
      ],
    },
  ];

  // 接驳车与清扫小车 — 仅 type=2,3 显示
  if (visibleSections.includes('shuttle_cart')) {
    const cartFields: DetailField[] = [
      { label: '清扫小车电池电压', value: displayNumber(robot.cart_battery_voltage, ' V', 1) },
      { label: '清扫小车低电压状态', value: displayBool(robot.cart_low_voltage_status, '低压', '正常') },
      { label: '清扫小车清扫电机电流', value: displayNumber(robot.cart_clean_motor_current, ' A', 1) },
      { label: '清扫小车行进电机电流', value: displayNumber(robot.cart_travel_motor_current, ' A', 1) },
    ];

    // 接驳车特有字段仅 type=2 显示
    if (type === 2) {
      cartFields.push(
        { label: '接驳车电池电压', value: displayNumber(robot.shuttle_battery_voltage, ' V', 1) },
        { label: '接驳车低电压状态', value: displayBool(robot.shuttle_low_voltage_status, '低压', '正常') },
        { label: '接驳车电机电流', value: displayNumber(robot.shuttle_motor_current, ' A', 1) },
        { label: '接驳车功率选择', value: displayNumber(robot.shuttle_power_percent, '%') },
        { label: '接驳车起点位置', value: displayBool(robot.shuttle_start_position) },
        { label: '接驳车终点位置', value: displayBool(robot.shuttle_end_position) },
        { label: '接驳车上是否有清扫车', value: displayBool(robot.shuttle_has_cleaner) },
      );
    }

    allSections.push({
      key: 'shuttle_cart',
      title: type === 2 ? '接驳车与清扫小车' : '清扫小车',
      fields: cartFields,
    });
  }

  return allSections;
};

export default function RobotsPage() {
  const [data, setData] = useState<Robot[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingRecord, setEditingRecord] = useState<Robot | null>(null);
  const [detailVisible, setDetailVisible] = useState(false);
  const [detailRobot, setDetailRobot] = useState<Robot | null>(null);
  const [form] = Form.useForm();
  const [filters, setFilters] = useState<FilterParams>({});
  const [stations, setStations] = useState<Station[]>([]);

  // 加载电站列表用于筛选和创建
  const fetchStations = async () => {
    try {
      const res = await stationApi.getAll();
      setStations(res);
    } catch {
      // error handled by interceptor
    }
  };

  useEffect(() => {
    fetchStations();
  }, []);

  const fetchData = async () => {
    setLoading(true);
    try {
      const params: Record<string, unknown> = {
        page,
        size: pageSize,
      };
      if (filters.station_id) params.station_id = filters.station_id;
      if (filters.robot_type !== undefined) params.robot_type = filters.robot_type;
      if (filters.online_status !== undefined) params.online_status = filters.online_status;
      const res = await robotApi.list(params);
      setData(res.list);
      setTotal(res.total);
    } catch {
      // error handled by interceptor
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [page, pageSize]);

  const handleSearch = () => {
    setPage(1);
    fetchData();
  };

  const handleReset = () => {
    setFilters({});
    setPage(1);
  };

  const handleCreate = () => {
    setEditingRecord(null);
    form.resetFields();
    setModalVisible(true);
  };

  const handleEdit = (record: Robot) => {
    setEditingRecord(record);
    form.setFieldsValue({
      robot_code: record.robot_code,
      robot_name: record.robot_name,
      robot_type: record.robot_type,
      station_id: record.station_id,
    });
    setModalVisible(true);
  };

  const handleDelete = async (id: string) => {
    try {
      await robotApi.delete(id);
      message.success('删除成功');
      fetchData();
    } catch {
      // error handled by interceptor
    }
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      if (editingRecord) {
        await robotApi.update(editingRecord.robot_id, values);
        message.success('更新成功');
      } else {
        await robotApi.create(values);
        message.success('创建成功');
      }
      setModalVisible(false);
      form.resetFields();
      fetchData();
    } catch (err) {
      if (err && typeof err === 'object' && 'errorFields' in err) {
        return;
      }
    }
  };

  // 控制命令映射
  const COMMANDS: { key: string; label: string; icon: ReactNode; cmd: string }[] = [
    { key: 'start', label: '启动', icon: <PlayCircleOutlined />, cmd: 'start' },
    { key: 'stop', label: '停止', icon: <PauseCircleOutlined />, cmd: 'stop' },
    { key: 'return', label: '回仓', icon: <HomeOutlined />, cmd: 'return' },
    { key: 'reset', label: '复位', icon: <UndoOutlined />, cmd: 'reset' },
  ];

  const handleCommand = async (robotId: string, cmd: string, label: string) => {
    try {
      await robotApi.sendCommand(robotId, cmd);
      message.success(`${label}指令已发送`);
      fetchData();
    } catch {
      // error handled by interceptor
    }
  };

  const handleShowDetail = (record: Robot) => {
    setDetailRobot(record);
    setDetailVisible(true);
  };

  const allColumns: ColumnsType<Robot> = [
    textColumn('机器人ID', 'robot_id', 160, true),
    textColumn('机器人编码', 'robot_code', 140, true),
    textColumn('名称', 'robot_name', 150, true),
    {
      title: '类型',
      dataIndex: 'robot_type',
      key: 'robot_type',
      width: 100,
      align: 'center',
      render: (type: number) => ROBOT_TYPE_MAP[type] || type,
    },
    textColumn('所属电站ID', 'station_id', 140, true),
    textColumn('所属电站名称', 'station_name', 160, true),
    numberColumn('X坐标(m)', 'pos_x', 110, '', 3),
    numberColumn('Y坐标(m)', 'pos_y', 110, '', 3),
    numberColumn('Z坐标(m)', 'pos_z', 110, '', 3),
    numberColumn('朝向(°)', 'heading', 100, '', 2),
    numberColumn('GPS经度', 'gps_longitude', 120, '', 7),
    numberColumn('GPS纬度', 'gps_latitude', 120, '', 7),
    numberColumn('GPS海拔(m)', 'gps_altitude', 120, '', 3),
    numberColumn('GPS精度(m)', 'gps_accuracy', 120, '', 3),
    {
      title: '电量',
      dataIndex: 'battery_level',
      key: 'battery_level',
      width: 120,
      align: 'center',
      render: (level: number | null | undefined) => {
        const percent = level ?? 0;
        return (
          <Progress
            percent={percent}
            size="small"
            status={percent <= 20 ? 'exception' : percent <= 50 ? 'normal' : 'success'}
            format={(p) => `${p}%`}
          />
        );
      },
    },
    numberColumn('电池电压(V)', 'battery_voltage', 120, '', 1),
    {
      title: '在线状态',
      dataIndex: 'online_status',
      key: 'online_status',
      width: 100,
      align: 'center',
      render: (status: number) => {
        const cfg = ONLINE_STATUS_MAP[status] || { text: '未知', color: 'default' };
        return <Tag color={cfg.color}>{cfg.text}</Tag>;
      },
    },
    {
      title: '工作状态',
      dataIndex: 'work_status',
      key: 'work_status',
      width: 100,
      align: 'center',
      render: (status: number) => {
        const cfg = WORK_STATUS_MAP[status] || { text: '未知', color: 'default' };
        return <Tag color={cfg.color}>{cfg.text}</Tag>;
      },
    },
    boolColumn('故障状态', 'fault_status', 110, '异常', '正常'),
    numberColumn('工作时段', 'work_period', 100),
    numberColumn('综合信号', 'signal_strength', 110),
    numberColumn('4G信号', 'signal_4g_strength', 110),
    {
      title: '运行时长',
      dataIndex: 'run_duration_seconds',
      key: 'run_duration_seconds',
      width: 120,
      align: 'right',
      render: (seconds: number | null | undefined) => displayDuration(seconds),
    },
    numberColumn('速度(m/s)', 'speed', 110, '', 2),
    numberColumn('清扫面积(m²)', 'clean_area', 130, '', 2),
    numberColumn('故障码', 'fault_code', 100),
    numberColumn('小车电压(V)', 'cart_battery_voltage', 120, '', 1),
    boolColumn('小车低压', 'cart_low_voltage_status', 110, '低压', '正常'),
    numberColumn('接驳车电压(V)', 'shuttle_battery_voltage', 130, '', 1),
    boolColumn('接驳车低压', 'shuttle_low_voltage_status', 120, '低压', '正常'),
    numberColumn('接驳车电流(A)', 'shuttle_motor_current', 130, '', 1),
    numberColumn('清扫电机电流(A)', 'cart_clean_motor_current', 140, '', 1),
    numberColumn('行进电机电流(A)', 'cart_travel_motor_current', 140, '', 1),
    numberColumn('接驳车功率(%)', 'shuttle_power_percent', 130),
    boolColumn('接驳起点', 'shuttle_start_position', 110),
    boolColumn('接驳终点', 'shuttle_end_position', 110),
    boolColumn('车上有清扫车', 'shuttle_has_cleaner', 130),
    numberColumn('温度(°C)', 'temperature', 100, '', 1),
    numberColumn('湿度(%)', 'humidity', 100, '', 1),
    numberColumn('光照(lux)', 'light_intensity', 110, '', 2),
    numberColumn('风速(m/s)', 'wind_speed', 110, '', 2),
    textColumn('设备时间', 'device_time', 170, true),
    textColumn('最后心跳', 'last_heartbeat', 170, true),
    textColumn('创建时间', 'create_time', 170, true),
    textColumn('更新时间', 'update_time', 170, true),
    {
      title: '操作',
      key: 'actions',
      width: 300,
      align: 'center',
      fixed: 'right',
      render: (_val: unknown, record: Robot) => (
        <Space size="small" wrap>
          <Button
            type="link"
            size="small"
            icon={<InfoCircleOutlined />}
            onClick={() => handleShowDetail(record)}
          >
            详情
          </Button>
          {COMMANDS.map((c) => (
            <Popconfirm
              key={c.key}
              title={`确定向该机器人发送「${c.label}」指令吗？`}
              onConfirm={() => handleCommand(record.robot_id, c.cmd, c.label)}
              okText="确定"
              cancelText="取消"
            >
              <Button type="link" size="small" icon={c.icon}>
                {c.label}
              </Button>
            </Popconfirm>
          ))}
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定删除该机器人吗？"
            description="删除后不可恢复"
            onConfirm={() => handleDelete(record.robot_id)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const visibleFields = filters.robot_type
    ? (ROBOT_TABLE_FIELDS[filters.robot_type] || ROBOT_DEFAULT_TABLE_FIELDS)
    : ROBOT_DEFAULT_TABLE_FIELDS;

  const columns = useMemo(
    () =>
      allColumns.filter(
        (col) => col.key === 'actions' || visibleFields.includes(col.key as string),
      ),
    [filters.robot_type],
  );

  return (
    <div style={{ padding: 24 }}>
      <Title level={4} style={{ marginBottom: 16 }}>机器人管理</Title>

      {/* 筛选栏 */}
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={[16, 16]} align="middle">
          <Col>
            <Select
              placeholder="选择电站"
              value={filters.station_id}
              onChange={(val) => setFilters((prev) => ({ ...prev, station_id: val }))}
              allowClear
              style={{ width: 180 }}
              options={stations.map((s) => ({
                label: s.station_name,
                value: s.station_id,
              }))}
            />
          </Col>
          <Col>
            <Select
              placeholder="选择类型"
              value={filters.robot_type}
              onChange={(val) => setFilters((prev) => ({ ...prev, robot_type: val }))}
              allowClear
              style={{ width: 140 }}
              options={Object.entries(ROBOT_TYPE_MAP).map(([k, v]) => ({
                label: v,
                value: Number(k),
              }))}
            />
          </Col>
          <Col>
            <Select
              placeholder="在线状态"
              value={filters.online_status}
              onChange={(val) => setFilters((prev) => ({ ...prev, online_status: val }))}
              allowClear
              style={{ width: 140 }}
              options={Object.entries(ONLINE_STATUS_MAP).map(([k, v]) => ({
                label: v.text,
                value: Number(k),
              }))}
            />
          </Col>
          <Col>
            <Space>
              <Button type="primary" icon={<SearchOutlined />} onClick={handleSearch}>
                搜索
              </Button>
              <Button onClick={handleReset}>重置</Button>
            </Space>
          </Col>
        </Row>
      </Card>

      {/* 操作栏 + 表格 */}
      <Card>
        <Row justify="space-between" align="middle" style={{ marginBottom: 16 }}>
          <Col>
            <Space>
              <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
                新建机器人
              </Button>
              <Button icon={<ReloadOutlined />} onClick={fetchData}>
                刷新
              </Button>
            </Space>
          </Col>
        </Row>

        <Table<Robot>
          rowKey="robot_id"
          columns={columns}
          dataSource={data}
          loading={loading}
          pagination={{
            current: page,
            pageSize,
            total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (t) => `共 ${t} 条`,
            onChange: (p, ps) => {
              setPage(p);
              setPageSize(ps);
            },
          }}
          locale={{
            emptyText: '暂无机器人数据',
          }}
          scroll={{ x: Math.max(1200, columns.length * 140) }}
        />
      </Card>

      {/* 创建/编辑弹窗 */}
      <Modal
        title={editingRecord ? '编辑机器人' : '新建机器人'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => {
          setModalVisible(false);
          form.resetFields();
        }}
        okText="保存"
        cancelText="取消"
        destroyOnClose
        width={520}
      >
        <Form
          form={form}
          layout="vertical"
          style={{ marginTop: 16 }}
        >
          <Form.Item
            name="robot_code"
            label="机器人编码"
            rules={[{ required: true, message: '请输入机器人编码' }]}
          >
            <Input placeholder="请输入机器人编码" disabled={!!editingRecord} />
          </Form.Item>
          <Form.Item
            name="robot_name"
            label="机器人名称"
            rules={[{ required: true, message: '请输入机器人名称' }]}
          >
            <Input placeholder="请输入机器人名称" />
          </Form.Item>
          <Form.Item
            name="robot_type"
            label="机器人类型"
            rules={[{ required: true, message: '请选择机器人类型' }]}
          >
            <Select
              placeholder="请选择机器人类型"
              options={Object.entries(ROBOT_TYPE_MAP).map(([k, v]) => ({
                label: v,
                value: Number(k),
              }))}
            />
          </Form.Item>
          <Form.Item
            name="station_id"
            label="所属电站"
            rules={[{ required: true, message: '请选择所属电站' }]}
          >
            <Select
              placeholder="请选择所属电站"
              showSearch
              optionFilterProp="label"
              options={stations.map((s) => ({
                label: `${s.station_name} (${s.station_code})`,
                value: s.station_id,
              }))}
            />
          </Form.Item>
        </Form>
      </Modal>

      {/* 机器人详情抽屉 */}
      <Modal
        title={`机器人详情 - ${detailRobot?.robot_name || ''}`}
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={[
          <Button key="close" onClick={() => setDetailVisible(false)}>
            关闭
          </Button>,
        ]}
        width={860}
      >
        {detailRobot && (
          <Space direction="vertical" size={16} style={{ width: '100%', marginTop: 8 }}>
            {getRobotDetailSections(detailRobot).map((section) => (
              <Descriptions
                key={section.title}
                title={section.title}
                bordered
                column={2}
                size="small"
              >
                {section.fields.map((field) => (
                  <Descriptions.Item key={`${section.title}-${field.label}`} label={field.label}>
                    {field.value}
                  </Descriptions.Item>
                ))}
              </Descriptions>
            ))}
          </Space>
        )}
      </Modal>
    </div>
  );
}
