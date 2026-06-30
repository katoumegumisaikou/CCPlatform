import { useState, useEffect, useCallback } from 'react';
import {
  Card, Table, Button, Modal, Form, Input, Select, DatePicker, Space, Tag, message, Popconfirm, Alert,
} from 'antd';
import {
  PlusOutlined, DeleteOutlined, ReloadOutlined, AlertOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { maintenanceApi } from '../../api/system';
import { stationApi } from '../../api/stations';
import { robotApi } from '../../api/robots';
import type { Maintenance, Station, Robot } from '../../types';

interface MaintenanceWithKey extends Maintenance {
  key: number;
}

export default function MaintenancePage() {
  const [data, setData] = useState<MaintenanceWithKey[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  // filters
  const [filterStationId, setFilterStationId] = useState<string | undefined>(undefined);
  const [filterStatus, setFilterStatus] = useState<number | undefined>(undefined);
  const [stations, setStations] = useState<Station[]>([]);

  // create/edit modal
  const [modalOpen, setModalOpen] = useState(false);
  const [modalTitle, setModalTitle] = useState('新建维护任务');
  const [confirmLoading, setConfirmLoading] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [robots, setRobots] = useState<Robot[]>([]);
  const [form] = Form.useForm();

  // reminders
  const [reminders, setReminders] = useState<Maintenance[]>([]);
  const [remindersLoading, setRemindersLoading] = useState(false);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const params: Record<string, unknown> = { page, size: pageSize };
      if (filterStationId) params.station_id = filterStationId;
      if (filterStatus !== undefined) params.status = filterStatus;

      const res = await maintenanceApi.list(params);
      const list = res?.list || [];
      setData(list.map((it) => ({ ...it, key: it.id })));
      setTotal(res?.total || 0);
    } catch {
      message.error('获取维护任务列表失败');
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, filterStationId, filterStatus]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  useEffect(() => {
    stationApi.getAll().then((res) => {
      setStations(res || []);
    }).catch(() => {
      message.error('获取电站列表失败');
    });
  }, []);

  const fetchRobots = async (stationId?: string) => {
    try {
      const params: Record<string, unknown> = { size: 999 };
      if (stationId) params.station_id = stationId;
      const res = await robotApi.list(params);
      setRobots(res?.list || []);
    } catch {
      message.error('获取机器人列表失败');
    }
  };

  const fetchReminders = async () => {
    setRemindersLoading(true);
    try {
      const res = await maintenanceApi.reminders();
      setReminders(res || []);
    } catch {
      message.error('获取保养提醒失败');
    } finally {
      setRemindersLoading(false);
    }
  };

  useEffect(() => {
    fetchReminders();
  }, []);

  const handleOpenCreate = async () => {
    setEditingId(null);
    setModalTitle('新建维护任务');
    form.resetFields();
    setModalOpen(true);
    await fetchRobots();
  };

  const handleDelete = async (id: number) => {
    try {
      await maintenanceApi.delete(id);
      message.success('删除成功');
      fetchData();
    } catch {
      message.error('删除失败');
    }
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setConfirmLoading(true);
      const payload: Partial<Maintenance> = {
        robot_id: values.robot_id,
        station_id: values.station_id,
        maintenance_type: values.maintenance_type,
        description: values.description || '',
        plan_date: values.plan_date ? values.plan_date.format('YYYY-MM-DD') : '',
      };
      if (editingId) {
        await maintenanceApi.update(editingId, payload);
        message.success('更新成功');
      } else {
        await maintenanceApi.create(payload);
        message.success('创建成功');
      }
      setModalOpen(false);
      fetchData();
      fetchReminders();
    } catch (err: any) {
      if (err?.errorFields) return;
      message.error('操作失败');
    } finally {
      setConfirmLoading(false);
    }
  };

  const statusMap: Record<number, { text: string; color: string }> = {
    0: { text: '待维护', color: 'default' },
    1: { text: '维护中', color: 'processing' },
    2: { text: '已完成', color: 'success' },
    3: { text: '已延期', color: 'warning' },
    4: { text: '已取消', color: 'default' },
  };

  const columns: ColumnsType<MaintenanceWithKey> = [
    { title: '任务ID', dataIndex: 'id', key: 'id', width: 80 },
    { title: '机器人ID', dataIndex: 'robot_id', key: 'robot_id', width: 140, ellipsis: true },
    { title: '所属电站', dataIndex: 'station_id', key: 'station_id', width: 140, ellipsis: true },
    { title: '维护类型', dataIndex: 'maintenance_type', key: 'maintenance_type', width: 120 },
    { title: '描述', dataIndex: 'description', key: 'description', width: 200, ellipsis: true, render: (v: string) => v || '-' },
    { title: '计划日期', dataIndex: 'plan_date', key: 'plan_date', width: 120 },
    {
      title: '完成日期', dataIndex: 'complete_date', key: 'complete_date', width: 120,
      render: (v: string) => v || '-',
    },
    {
      title: '状态', dataIndex: 'status', key: 'status', width: 100,
      render: (v: number) => {
        const s = statusMap[v] || { text: '未知', color: 'default' };
        return <Tag color={s.color}>{s.text}</Tag>;
      },
    },
    {
      title: '操作', key: 'action', width: 100, fixed: 'right',
      render: (_, record) => (
        <Popconfirm
          title="确定删除该维护任务？"
          onConfirm={() => handleDelete(record.id)}
          okText="确定"
          cancelText="取消"
        >
          <Button type="link" danger icon={<DeleteOutlined />} size="small">
            删除
          </Button>
        </Popconfirm>
      ),
    },
  ];

  return (
    <div>
      {/* 保养提醒 */}
      {reminders.length > 0 && (
        <Card style={{ marginBottom: 16 }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 12 }}>
            <Space>
              <AlertOutlined style={{ color: '#faad14', fontSize: 18 }} />
              <span style={{ fontSize: 16, fontWeight: 500 }}>保养提醒</span>
              <Tag color="warning">{reminders.length} 条</Tag>
            </Space>
            <Button size="small" icon={<ReloadOutlined />} onClick={fetchReminders} loading={remindersLoading}>
              刷新
            </Button>
          </div>
          {reminders.map((r) => (
            <Alert
              key={r.id}
              type="warning"
              showIcon
              style={{ marginBottom: 8 }}
              message={
                <span>
                  机器人 <strong>{r.robot_id}</strong> — {r.maintenance_type}
                  <span style={{ marginLeft: 16, color: '#999' }}>
                    计划日期：{r.plan_date}
                  </span>
                  {r.description && (
                    <span style={{ marginLeft: 16, color: '#666' }}>
                      | {r.description}
                    </span>
                  )}
                </span>
              }
            />
          ))}
        </Card>
      )}

      {/* 维护任务 */}
      <Card
        title="维护任务"
        extra={
          <Space>
            <Select
              placeholder="按电站筛选"
              allowClear
              style={{ width: 180 }}
              value={filterStationId}
              onChange={(val) => {
                setFilterStationId(val);
                setPage(1);
              }}
              options={stations.map((s) => ({ label: s.station_name, value: s.station_id }))}
            />
            <Select
              placeholder="按状态筛选"
              allowClear
              style={{ width: 120 }}
              value={filterStatus}
              onChange={(val) => {
                setFilterStatus(val);
                setPage(1);
              }}
              options={Object.entries(statusMap).map(([k, v]) => ({ label: v.text, value: Number(k) }))}
            />
            <Button icon={<ReloadOutlined />} onClick={fetchData}>刷新</Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={handleOpenCreate}>
              新建任务
            </Button>
          </Space>
        }
      >
        <Table
          columns={columns}
          dataSource={data}
          loading={loading}
          scroll={{ x: 1100 }}
          pagination={{
            current: page,
            pageSize,
            total,
            showSizeChanger: true,
            showTotal: (t) => `共 ${t} 条`,
            onChange: (p, ps) => {
              setPage(p);
              setPageSize(ps);
            },
          }}
        />
      </Card>

      {/* 新建/编辑维护任务 Modal */}
      <Modal
        title={modalTitle}
        open={modalOpen}
        onOk={handleSubmit}
        onCancel={() => setModalOpen(false)}
        confirmLoading={confirmLoading}
        okText="保存"
        cancelText="取消"
        destroyOnClose
      >
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item
            name="station_id"
            label="所属电站"
            rules={[{ required: true, message: '请选择电站' }]}
          >
            <Select
              placeholder="选择电站"
              showSearch
              optionFilterProp="label"
              onChange={(val) => {
                form.setFieldValue('robot_id', undefined);
                fetchRobots(val);
              }}
              options={stations.map((s) => ({ label: s.station_name, value: s.station_id }))}
            />
          </Form.Item>
          <Form.Item
            name="robot_id"
            label="机器人"
            rules={[{ required: true, message: '请选择机器人' }]}
          >
            <Select
              placeholder="选择机器人"
              showSearch
              optionFilterProp="label"
              options={robots.map((r) => ({ label: `${r.robot_name || r.robot_id} (${r.robot_id})`, value: r.robot_id }))}
            />
          </Form.Item>
          <Form.Item
            name="maintenance_type"
            label="维护类型"
            rules={[{ required: true, message: '请输入维护类型' }]}
          >
            <Input placeholder="如：定期检查、部件更换、润滑保养" />
          </Form.Item>
          <Form.Item
            name="description"
            label="描述"
          >
            <Input.TextArea rows={3} placeholder="维护内容描述" />
          </Form.Item>
          <Form.Item
            name="plan_date"
            label="计划日期"
            rules={[{ required: true, message: '请选择计划日期' }]}
          >
            <DatePicker style={{ width: '100%' }} placeholder="选择计划日期" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
