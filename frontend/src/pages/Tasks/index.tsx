import { useState, useEffect } from 'react';
import {
  Table, Button, Modal, Form, Input, Select, Popconfirm, message, Space, Tag,
  Card, Row, Col, Typography, Radio, DatePicker, Progress, Tooltip,
} from 'antd';
import {
  PlusOutlined, ReloadOutlined, SearchOutlined, DeleteOutlined,
  PlayCircleOutlined, PauseCircleOutlined, CheckCircleOutlined, CloseCircleOutlined,
  ThunderboltOutlined, DashboardOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import { taskApi } from '../../api/tasks';
import { robotApi } from '../../api/robots';
import { stationApi } from '../../api/stations';
import type { Task, TaskProgress, Robot, Station } from '../../types';
import { TASK_STATUS_MAP, TASK_TYPE_MAP } from '../../utils';

const { Title, Text } = Typography;

interface FilterParams {
  station_id?: string;
  robot_id?: string;
  task_status?: number;
}

interface TaskProgressCache {
  [taskId: number]: { progress: TaskProgress; visible: boolean };
}

export default function TasksPage() {
  const [data, setData] = useState<Task[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [modalVisible, setModalVisible] = useState(false);
  const [form] = Form.useForm();
  const [filters, setFilters] = useState<FilterParams>({});
  const [stations, setStations] = useState<Station[]>([]);
  const [robots, setRobots] = useState<Robot[]>([]);
  const [taskType, setTaskType] = useState<number>(1);
  const [progressMap, setProgressMap] = useState<TaskProgressCache>({});
  const [progressLoading, setProgressLoading] = useState<Record<number, boolean>>({});
  const [smartScheduleLoading, setSmartScheduleLoading] = useState(false);
  const [smartStationId, setSmartStationId] = useState<string | undefined>(undefined);

  const fetchStations = async () => {
    try {
      const res = await stationApi.getAll();
      setStations(res);
    } catch {
      // error handled by interceptor
    }
  };

  const fetchRobots = async (stationId?: string) => {
    try {
      const params: Record<string, unknown> = { size: 999 };
      if (stationId) params.station_id = stationId;
      const res = await robotApi.list(params);
      setRobots(res.list);
    } catch {
      // error handled by interceptor
    }
  };

  useEffect(() => {
    fetchStations();
    fetchRobots();
  }, []);

  const fetchData = async () => {
    setLoading(true);
    try {
      const params: Record<string, unknown> = {
        page,
        size: pageSize,
      };
      if (filters.station_id) params.station_id = filters.station_id;
      if (filters.robot_id) params.robot_id = filters.robot_id;
      if (filters.task_status !== undefined) params.task_status = filters.task_status;
      const res = await taskApi.list(params);
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
    form.resetFields();
    setTaskType(1);
    form.setFieldsValue({ task_type: 1 });
    setModalVisible(true);
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      // 格式化日期
      const payload: Record<string, unknown> = { ...values };
      if (values.plan_start) {
        payload.plan_start = dayjs(values.plan_start).format('YYYY-MM-DD HH:mm:ss');
      }
      if (values.plan_end) {
        payload.plan_end = dayjs(values.plan_end).format('YYYY-MM-DD HH:mm:ss');
      }
      // area_ids 前端输入是逗号分隔的字符串，保持原样发送
      await taskApi.create(payload as Partial<Task>);
      message.success('创建任务成功');
      setModalVisible(false);
      form.resetFields();
      fetchData();
    } catch (err) {
      if (err && typeof err === 'object' && 'errorFields' in err) {
        return;
      }
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await taskApi.delete(id);
      message.success('删除成功');
      fetchData();
    } catch {
      // error handled by interceptor
    }
  };

  const handleUpdateStatus = async (taskId: number, status: number, statusLabel: string) => {
    try {
      await taskApi.updateStatus(taskId, status);
      message.success(`任务已${statusLabel}`);
      fetchData();
    } catch {
      // error handled by interceptor
    }
  };

  const handleShowProgress = async (taskId: number) => {
    setProgressLoading((prev) => ({ ...prev, [taskId]: true }));
    try {
      const res = await taskApi.getProgress(taskId);
      setProgressMap((prev) => ({
        ...prev,
        [taskId]: { progress: res, visible: true },
      }));
    } catch {
      // error handled by interceptor
    } finally {
      setProgressLoading((prev) => ({ ...prev, [taskId]: false }));
    }
  };

  const handleSmartSchedule = async () => {
    if (!smartStationId) {
      message.warning('请先选择电站');
      return;
    }
    setSmartScheduleLoading(true);
    try {
      await taskApi.smartSchedule(smartStationId);
      message.success('智能排程已完成');
      fetchData();
    } catch {
      // error handled by interceptor
    } finally {
      setSmartScheduleLoading(false);
    }
  };

  // 获取状态操作按钮配置
  const getStatusActions = (task: Task) => {
    const { task_status } = task;
    const actions: { status: number; label: string; icon: React.ReactNode; color?: string }[] = [];

    if (task_status === 0) {
      // 待执行 -> 启动
      actions.push({ status: 1, label: '启动', icon: <PlayCircleOutlined />, color: 'blue' });
    }
    if (task_status === 1) {
      // 执行中 -> 暂停, 完成
      actions.push({ status: 3, label: '暂停', icon: <PauseCircleOutlined />, color: 'orange' });
      actions.push({ status: 2, label: '完成', icon: <CheckCircleOutlined />, color: 'green' });
    }
    if (task_status === 3) {
      // 已暂停 -> 恢复, 完成
      actions.push({ status: 1, label: '恢复', icon: <PlayCircleOutlined />, color: 'blue' });
      actions.push({ status: 2, label: '完成', icon: <CheckCircleOutlined />, color: 'green' });
    }
    if (task_status === 0 || task_status === 1 || task_status === 3) {
      // 未完成状态 -> 可取消
      actions.push({ status: 4, label: '取消', icon: <CloseCircleOutlined />, color: 'default' });
    }

    return actions;
  };

  const columns: ColumnsType<Task> = [
    {
      title: '任务名称',
      dataIndex: 'task_name',
      key: 'task_name',
      width: 180,
      ellipsis: true,
    },
    {
      title: '类型',
      dataIndex: 'task_type',
      key: 'task_type',
      width: 80,
      align: 'center',
      render: (type: number) => (
        <Tag>{TASK_TYPE_MAP[type] || type}</Tag>
      ),
    },
    {
      title: '机器人',
      dataIndex: 'robot_id',
      key: 'robot_id',
      width: 140,
      ellipsis: true,
    },
    {
      title: '任务状态',
      dataIndex: 'task_status',
      key: 'task_status',
      width: 100,
      align: 'center',
      render: (status: number) => {
        const cfg = TASK_STATUS_MAP[status] || { text: '未知', color: 'default' };
        return <Tag color={cfg.color}>{cfg.text}</Tag>;
      },
    },
    {
      title: '清扫面积',
      dataIndex: 'clean_area',
      key: 'clean_area',
      width: 120,
      align: 'right',
      render: (val: number) => (val ? `${val.toFixed(2)} m²` : '-'),
    },
    {
      title: '计划开始',
      dataIndex: 'plan_start',
      key: 'plan_start',
      width: 160,
      render: (val: string | null) => val || '-',
    },
    {
      title: '计划结束',
      dataIndex: 'plan_end',
      key: 'plan_end',
      width: 160,
      render: (val: string | null) => val || '-',
    },
    {
      title: '创建时间',
      dataIndex: 'create_time',
      key: 'create_time',
      width: 170,
    },
    {
      title: '操作',
      key: 'actions',
      width: 320,
      align: 'center',
      fixed: 'right',
      render: (_val: unknown, record: Task) => {
        const statusActions = getStatusActions(record);
        const progressInfo = progressMap[record.task_id];
        const isProgressLoading = progressLoading[record.task_id];

        return (
          <div>
            <Space size="small" wrap>
              <Button
                type="link"
                size="small"
                icon={<DashboardOutlined />}
                loading={isProgressLoading}
                onClick={() => handleShowProgress(record.task_id)}
              >
                进度
              </Button>
              {statusActions.map((act) => (
                <Popconfirm
                  key={act.status}
                  title={`确定将任务状态变更为「${act.label}」吗？`}
                  onConfirm={() => handleUpdateStatus(record.task_id, act.status, act.label)}
                  okText="确定"
                  cancelText="取消"
                >
                  <Button
                    type="link"
                    size="small"
                    icon={act.icon}
                    style={{ color: act.color }}
                  >
                    {act.label}
                  </Button>
                </Popconfirm>
              ))}
              <Popconfirm
                title="确定删除该任务吗？"
                description="删除后不可恢复"
                onConfirm={() => handleDelete(record.task_id)}
                okText="确定"
                cancelText="取消"
              >
                <Button type="link" size="small" danger icon={<DeleteOutlined />}>
                  删除
                </Button>
              </Popconfirm>
            </Space>
            {/* 内联进度条 */}
            {progressInfo?.visible && (
              <div style={{ marginTop: 8, padding: '8px 12px', background: '#fafafa', borderRadius: 4 }}>
                <Row gutter={16} align="middle">
                  <Col flex="auto">
                    <Progress
                      percent={Math.round(progressInfo.progress.progress * 100) / 100}
                      status={progressInfo.progress.task_status === 5 ? 'exception' : 'active'}
                      format={(p) => `${p}%`}
                    />
                  </Col>
                  <Col>
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      剩余 {progressInfo.progress.estimated_remaining_seconds > 0
                        ? `${Math.round(progressInfo.progress.estimated_remaining_seconds)}秒`
                        : '-'}
                    </Text>
                  </Col>
                </Row>
              </div>
            )}
          </div>
        );
      },
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <Title level={4} style={{ marginBottom: 16 }}>任务管理</Title>

      {/* 筛选栏 */}
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={[16, 16]} align="middle">
          <Col>
            <Select
              placeholder="选择电站"
              value={filters.station_id}
              onChange={(val) => {
                setFilters((prev) => ({ ...prev, station_id: val, robot_id: undefined }));
                fetchRobots(val);
              }}
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
              placeholder="选择机器人"
              value={filters.robot_id}
              onChange={(val) => setFilters((prev) => ({ ...prev, robot_id: val }))}
              allowClear
              showSearch
              optionFilterProp="label"
              style={{ width: 180 }}
              options={robots.map((r) => ({
                label: `${r.robot_name} (${r.robot_code})`,
                value: r.robot_id,
              }))}
            />
          </Col>
          <Col>
            <Select
              placeholder="任务状态"
              value={filters.task_status}
              onChange={(val) => setFilters((prev) => ({ ...prev, task_status: val }))}
              allowClear
              style={{ width: 140 }}
              options={Object.entries(TASK_STATUS_MAP).map(([k, v]) => ({
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
                新建任务
              </Button>
              <Button icon={<ReloadOutlined />} onClick={fetchData}>
                刷新
              </Button>
            </Space>
          </Col>
          <Col>
            <Space>
              <Select
                placeholder="选择电站"
                value={smartStationId}
                onChange={setSmartStationId}
                allowClear
                style={{ width: 180 }}
                options={stations.map((s) => ({
                  label: s.station_name,
                  value: s.station_id,
                }))}
              />
              <Tooltip title="根据选定电站自动生成最优排程任务">
                <Button
                  icon={<ThunderboltOutlined />}
                  onClick={handleSmartSchedule}
                  loading={smartScheduleLoading}
                >
                  智能排程
                </Button>
              </Tooltip>
            </Space>
          </Col>
        </Row>

        <Table<Task>
          rowKey="task_id"
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
            emptyText: '暂无任务数据',
          }}
          scroll={{ x: 1400 }}
        />
      </Card>

      {/* 创建弹窗 */}
      <Modal
        title="新建任务"
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => {
          setModalVisible(false);
          form.resetFields();
        }}
        okText="保存"
        cancelText="取消"
        destroyOnClose
        width={640}
      >
        <Form
          form={form}
          layout="vertical"
          style={{ marginTop: 16 }}
          initialValues={{ task_type: 1 }}
        >
          <Form.Item
            name="task_name"
            label="任务名称"
            rules={[{ required: true, message: '请输入任务名称' }]}
          >
            <Input placeholder="请输入任务名称" />
          </Form.Item>
          <Form.Item
            name="task_type"
            label="任务类型"
            rules={[{ required: true, message: '请选择任务类型' }]}
          >
            <Radio.Group
              onChange={(e) => setTaskType(e.target.value)}
              optionType="button"
              buttonStyle="solid"
            >
              <Radio.Button value={1}>即时</Radio.Button>
              <Radio.Button value={2}>定时</Radio.Button>
              <Radio.Button value={3}>周期</Radio.Button>
            </Radio.Group>
          </Form.Item>
          <Form.Item
            name="robot_id"
            label="机器人"
            rules={[{ required: true, message: '请选择机器人' }]}
          >
            <Select
              placeholder="请选择机器人"
              showSearch
              optionFilterProp="label"
              options={robots.map((r) => ({
                label: `${r.robot_name} (${r.robot_code})`,
                value: r.robot_id,
              }))}
            />
          </Form.Item>
          <Form.Item
            name="area_ids"
            label="清扫区域ID（逗号分隔）"
          >
            <Input placeholder="请输入清扫区域ID，多个用逗号分隔" />
          </Form.Item>
          {/* 定时任务：需要计划开始/结束 */}
          {taskType === 2 && (
            <>
              <Form.Item
                name="plan_start"
                label="计划开始时间"
                rules={[{ required: true, message: '请选择计划开始时间' }]}
              >
                <DatePicker
                  showTime
                  format="YYYY-MM-DD HH:mm:ss"
                  style={{ width: '100%' }}
                  placeholder="请选择计划开始时间"
                />
              </Form.Item>
              <Form.Item
                name="plan_end"
                label="计划结束时间"
                rules={[{ required: true, message: '请选择计划结束时间' }]}
              >
                <DatePicker
                  showTime
                  format="YYYY-MM-DD HH:mm:ss"
                  style={{ width: '100%' }}
                  placeholder="请选择计划结束时间"
                />
              </Form.Item>
            </>
          )}
          {/* 周期任务：需要cron表达式 */}
          {taskType === 3 && (
            <Form.Item
              name="cron_expr"
              label="Cron 表达式"
              rules={[{ required: true, message: '请输入 Cron 表达式' }]}
              extra="例如：0 6 * * * 表示每天早上6点执行"
            >
              <Input placeholder="请输入 Cron 表达式，如 0 6 * * *" />
            </Form.Item>
          )}
        </Form>
      </Modal>
    </div>
  );
}
