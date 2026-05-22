import { useState, useEffect } from 'react';
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
} from '../../utils';

const { Title } = Typography;

interface FilterParams {
  station_id?: string;
  robot_type?: number;
  online_status?: number;
}

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
  const COMMANDS: { key: string; label: string; icon: React.ReactNode; cmd: string }[] = [
    { key: 'start', label: '启动', icon: <PlayCircleOutlined />, cmd: 'start' },
    { key: 'pause', label: '停止', icon: <PauseCircleOutlined />, cmd: 'pause' },
    { key: 'home', label: '回仓', icon: <HomeOutlined />, cmd: 'home' },
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

  const columns: ColumnsType<Robot> = [
    {
      title: '机器人编码',
      dataIndex: 'robot_code',
      key: 'robot_code',
      width: 140,
    },
    {
      title: '名称',
      dataIndex: 'robot_name',
      key: 'robot_name',
      width: 150,
      ellipsis: true,
    },
    {
      title: '类型',
      dataIndex: 'robot_type',
      key: 'robot_type',
      width: 100,
      align: 'center',
      render: (type: number) => ROBOT_TYPE_MAP[type] || type,
    },
    {
      title: '所属电站',
      dataIndex: 'station_name',
      key: 'station_name',
      width: 160,
      ellipsis: true,
    },
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
    {
      title: '电量',
      dataIndex: 'battery_level',
      key: 'battery_level',
      width: 120,
      align: 'center',
      render: (level: number) => (
        <Progress
          percent={level}
          size="small"
          status={level <= 20 ? 'exception' : level <= 50 ? 'normal' : 'success'}
          format={(p) => `${p}%`}
        />
      ),
    },
    {
      title: '最后心跳',
      dataIndex: 'last_heartbeat',
      key: 'last_heartbeat',
      width: 170,
    },
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
          scroll={{ x: 1300 }}
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
        width={720}
      >
        {detailRobot && (
          <Descriptions bordered column={2} size="small" style={{ marginTop: 8 }}>
            <Descriptions.Item label="机器人编码">{detailRobot.robot_code}</Descriptions.Item>
            <Descriptions.Item label="机器人名称">{detailRobot.robot_name}</Descriptions.Item>
            <Descriptions.Item label="类型">
              <Tag>{ROBOT_TYPE_MAP[detailRobot.robot_type] || detailRobot.robot_type}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="所属电站">{detailRobot.station_name || detailRobot.station_id}</Descriptions.Item>
            <Descriptions.Item label="在线状态">
              <Tag color={ONLINE_STATUS_MAP[detailRobot.online_status]?.color}>
                {ONLINE_STATUS_MAP[detailRobot.online_status]?.text || '未知'}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="工作状态">
              <Tag color={WORK_STATUS_MAP[detailRobot.work_status]?.color}>
                {WORK_STATUS_MAP[detailRobot.work_status]?.text || '未知'}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="电量">{detailRobot.battery_level}%</Descriptions.Item>
            <Descriptions.Item label="清扫面积">{detailRobot.clean_area?.toFixed(2)} m²</Descriptions.Item>
            <Descriptions.Item label="速度">{detailRobot.speed} m/s</Descriptions.Item>
            <Descriptions.Item label="朝向">{detailRobot.heading}°</Descriptions.Item>
            <Descriptions.Item label="位置">
              ({detailRobot.pos_x}, {detailRobot.pos_y}, {detailRobot.pos_z})
            </Descriptions.Item>
            <Descriptions.Item label="温度">{detailRobot.temperature}°C</Descriptions.Item>
            <Descriptions.Item label="湿度">{detailRobot.humidity}%</Descriptions.Item>
            <Descriptions.Item label="光照强度">{detailRobot.light_intensity} lux</Descriptions.Item>
            <Descriptions.Item label="风速">{detailRobot.wind_speed} m/s</Descriptions.Item>
            <Descriptions.Item label="最后心跳">{detailRobot.last_heartbeat}</Descriptions.Item>
            <Descriptions.Item label="创建时间">{detailRobot.create_time}</Descriptions.Item>
          </Descriptions>
        )}
      </Modal>
    </div>
  );
}
