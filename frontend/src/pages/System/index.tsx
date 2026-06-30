import { useState, useEffect, useCallback } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import {
  Tabs, Table, Button, Modal, Form, Input, Select, message,
  Popconfirm, Space, Tag, Checkbox, Switch, Empty, Tooltip, Row, Col,
} from 'antd';
import {
  PlusOutlined, EditOutlined, DeleteOutlined, ReloadOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { userApi, roleApi } from '../../api/system';
import { stationApi } from '../../api/stations';
import type { User, Role, Station } from '../../types';

// ===== 常量定义 =====

const PERMISSION_OPTIONS = [
  { label: '电站-创建', value: 'station:create' },
  { label: '电站-编辑', value: 'station:update' },
  { label: '电站-删除', value: 'station:delete' },
  { label: '机器人-创建', value: 'robot:create' },
  { label: '机器人-编辑', value: 'robot:update' },
  { label: '机器人-删除', value: 'robot:delete' },
  { label: '机器人-控制', value: 'robot:control' },
  { label: '机器人-配置', value: 'robot:config' },
  { label: '任务-创建', value: 'task:create' },
  { label: '任务-编辑', value: 'task:update' },
  { label: '任务-删除', value: 'task:delete' },
  { label: '告警-处理', value: 'alarm:handle' },
  { label: '告警-配置', value: 'alarm:config' },
  { label: '用户-创建', value: 'user:create' },
  { label: '用户-编辑', value: 'user:update' },
  { label: '用户-删除', value: 'user:delete' },
  { label: '用户-查看', value: 'user:view' },
  { label: '角色-创建', value: 'role:create' },
  { label: '角色-编辑', value: 'role:update' },
  { label: '角色-删除', value: 'role:delete' },
  { label: '组织-创建', value: 'org:create' },
  { label: '组织-编辑', value: 'org:update' },
  { label: '组织-删除', value: 'org:delete' },
  { label: '固件-创建', value: 'firmware:create' },
  { label: '固件-编辑', value: 'firmware:update' },
  { label: '固件-删除', value: 'firmware:delete' },
  { label: '固件-升级', value: 'firmware:upgrade' },
  { label: '维护-创建', value: 'maintenance:create' },
  { label: '维护-编辑', value: 'maintenance:update' },
  { label: '维护-删除', value: 'maintenance:delete' },
  { label: '摄像头-创建', value: 'camera:create' },
  { label: '摄像头-编辑', value: 'camera:update' },
  { label: '摄像头-删除', value: 'camera:delete' },
  { label: '系统-配置', value: 'system:config' },
  { label: '通知-配置', value: 'notify:config' },
  { label: '监控-查看', value: 'monitor:view' },
];

// ===== 1. 用户管理 =====

function UsersTab() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [roles, setRoles] = useState<Role[]>([]);
  const [stations, setStations] = useState<Station[]>([]);
  const [form] = Form.useForm();

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await userApi.list({ page, size: pageSize });
      setUsers(res.list || []);
      setTotal(res.total || 0);
    } catch {
      setUsers([]);
    } finally {
      setLoading(false);
    }
  }, [page, pageSize]);

  const fetchRolesAndStations = useCallback(async () => {
    try {
      const [rolesRes, stationsRes] = await Promise.all([
        roleApi.list(),
        stationApi.getAll(),
      ]);
      setRoles(rolesRes || []);
      setStations(stationsRes || []);
    } catch { /* ignore */ }
  }, []);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleAdd = () => {
    setEditingUser(null);
    form.resetFields();
    form.setFieldsValue({ status: true, station_ids: [] });
    fetchRolesAndStations();
    setModalOpen(true);
  };

  const handleEdit = (record: User) => {
    setEditingUser(record);
    form.setFieldsValue({
      username: record.username,
      real_name: record.real_name,
      role_id: record.role_id,
      station_ids: record.station_ids ? record.station_ids.split(',').filter(Boolean) : [],
      phone: record.phone,
      email: record.email,
      status: record.status === 1,
    });
    fetchRolesAndStations();
    setModalOpen(true);
  };

  const handleOk = async () => {
    try {
      const values = await form.validateFields();
      const payload = {
        username: values.username,
        real_name: values.real_name,
        role_id: values.role_id,
        station_ids: Array.isArray(values.station_ids) ? values.station_ids.join(',') : '',
        phone: values.phone,
        email: values.email,
        status: values.status ? 1 : 0,
      };
      if (editingUser) {
        await userApi.update(editingUser.user_id, payload);
        message.success('用户更新成功');
      } else {
        await userApi.create({ ...payload, password: values.password });
        message.success('用户创建成功');
      }
      setModalOpen(false);
      fetchData();
    } catch (err: any) {
      if (err?.errorFields) return; // 表单校验失败
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await userApi.delete(id);
      message.success('用户已删除');
      fetchData();
    } catch { /* 拦截器已提示 */ }
  };

  const columns: ColumnsType<User> = [
    { title: '用户名', dataIndex: 'username', key: 'username', width: 120 },
    { title: '姓名', dataIndex: 'real_name', key: 'real_name', width: 100 },
    {
      title: '角色',
      dataIndex: 'role_id',
      key: 'role',
      width: 120,
      render: (_, record) => record.role?.role_name || `角色ID:${record.role_id}`,
    },
    { title: '手机号', dataIndex: 'phone', key: 'phone', width: 130 },
    { title: '邮箱', dataIndex: 'email', key: 'email', width: 180 },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (s: number) => (
        <Tag color={s === 1 ? 'green' : 'red'}>{s === 1 ? '启用' : '禁用'}</Tag>
      ),
    },
    {
      title: '操作',
      key: 'actions',
      width: 150,
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
            编辑
          </Button>
          <Popconfirm
            title="确定删除该用户吗？"
            onConfirm={() => handleDelete(record.user_id)}
            okText="确认"
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
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <Space>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
            添加用户
          </Button>
          <Button icon={<ReloadOutlined />} onClick={fetchData}>刷新</Button>
        </Space>
      </div>
      <Table
        rowKey="user_id"
        columns={columns}
        dataSource={users}
        loading={loading}
        locale={{ emptyText: <Empty description="暂无用户数据" /> }}
        pagination={{
          current: page,
          pageSize,
          total,
          showSizeChanger: true,
          showTotal: (t: number) => `共 ${t} 条`,
          onChange: (p: number, s: number) => { setPage(p); setPageSize(s); },
        }}
      />
      <Modal
        title={editingUser ? '编辑用户' : '添加用户'}
        open={modalOpen}
        onOk={handleOk}
        onCancel={() => setModalOpen(false)}
        destroyOnClose
        width={520}
      >
        <Form form={form} layout="vertical" preserve={false}>
          <Form.Item name="username" label="用户名" rules={[{ required: true, message: '请输入用户名' }]}>
            <Input disabled={!!editingUser} placeholder="请输入用户名" />
          </Form.Item>
          <Form.Item name="real_name" label="姓名" rules={[{ required: true, message: '请输入姓名' }]}>
            <Input placeholder="请输入姓名" />
          </Form.Item>
          {!editingUser && (
            <Form.Item name="password" label="密码" rules={[{ required: true, message: '请输入密码' }]}>
              <Input.Password placeholder="请输入密码" />
            </Form.Item>
          )}
          <Form.Item name="role_id" label="角色" rules={[{ required: true, message: '请选择角色' }]}>
            <Select
              placeholder="请选择角色"
              options={roles.map(r => ({ label: r.role_name, value: r.role_id }))}
            />
          </Form.Item>
          <Form.Item name="station_ids" label="授权电站">
            <Select
              mode="multiple"
              allowClear
              placeholder="请选择可访问电站，不选表示暂不限制"
              options={stations.map(s => ({ label: s.station_name, value: s.station_id }))}
              showSearch
              optionFilterProp="label"
            />
          </Form.Item>
          <Form.Item name="phone" label="手机号">
            <Input placeholder="请输入手机号" />
          </Form.Item>
          <Form.Item name="email" label="邮箱">
            <Input placeholder="请输入邮箱" />
          </Form.Item>
          <Form.Item name="status" label="状态" valuePropName="checked">
            <Switch checkedChildren="启用" unCheckedChildren="禁用" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}

// ===== 2. 角色管理 =====

function RolesTab() {
  const [roles, setRoles] = useState<Role[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingRole, setEditingRole] = useState<Role | null>(null);
  const [form] = Form.useForm();

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await roleApi.list();
      setRoles(res || []);
    } catch {
      setRoles([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleAdd = () => {
    setEditingRole(null);
    form.resetFields();
    form.setFieldsValue({ status: true, permissions: [] });
    setModalOpen(true);
  };

  const handleEdit = (record: Role) => {
    setEditingRole(record);
    form.setFieldsValue({
      role_name: record.role_name,
      role_code: record.role_code,
      permissions: record.permissions ? record.permissions.split(',') : [],
      status: record.status === 1,
    });
    setModalOpen(true);
  };

  const handleOk = async () => {
    try {
      const values = await form.validateFields();
      const payload = {
        role_name: values.role_name,
        role_code: values.role_code,
        permissions: (values.permissions as string[]).join(','),
        status: values.status ? 1 : 0,
      };
      if (editingRole) {
        await roleApi.update(editingRole.role_id, payload);
        message.success('角色更新成功');
      } else {
        await roleApi.create(payload);
        message.success('角色创建成功');
      }
      setModalOpen(false);
      fetchData();
    } catch (err: any) {
      if (err?.errorFields) return;
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await roleApi.delete(id);
      message.success('角色已删除');
      fetchData();
    } catch { /* ignore */ }
  };

  const columns: ColumnsType<Role> = [
    { title: '角色名', dataIndex: 'role_name', key: 'role_name', width: 130 },
    { title: '角色编码', dataIndex: 'role_code', key: 'role_code', width: 130 },
    {
      title: '权限列表',
      dataIndex: 'permissions',
      key: 'permissions',
      ellipsis: true,
      render: (perms: string) => {
        if (!perms) return <Tag>无权限</Tag>;
        const list = perms.split(',').slice(0, 5);
        const more = perms.split(',').length > 5;
        return (
          <Tooltip title={perms.split(',').join('、')}>
            <Space size={[2, 2]} wrap>
              {list.map(p => <Tag key={p} color="blue">{p}</Tag>)}
              {more && <Tag>+{perms.split(',').length - 5}</Tag>}
            </Space>
          </Tooltip>
        );
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (s: number) => (
        <Tag color={s === 1 ? 'green' : 'red'}>{s === 1 ? '启用' : '禁用'}</Tag>
      ),
    },
    {
      title: '操作',
      key: 'actions',
      width: 150,
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
            编辑
          </Button>
          <Popconfirm
            title="确定删除该角色吗？"
            onConfirm={() => handleDelete(record.role_id)}
            okText="确认"
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
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <Space>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
            添加角色
          </Button>
          <Button icon={<ReloadOutlined />} onClick={fetchData}>刷新</Button>
        </Space>
      </div>
      <Table
        rowKey="role_id"
        columns={columns}
        dataSource={roles}
        loading={loading}
        locale={{ emptyText: <Empty description="暂无角色数据" /> }}
        pagination={false}
      />
      <Modal
        title={editingRole ? '编辑角色' : '添加角色'}
        open={modalOpen}
        onOk={handleOk}
        onCancel={() => setModalOpen(false)}
        destroyOnClose
        width={680}
      >
        <Form form={form} layout="vertical" preserve={false}>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="role_name" label="角色名" rules={[{ required: true, message: '请输入角色名' }]}>
                <Input placeholder="请输入角色名" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="role_code" label="角色编码" rules={[{ required: true, message: '请输入角色编码' }]}>
                <Input placeholder="请输入角色编码" />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item
            name="permissions"
            label="权限列表"
            rules={[{ required: true, message: '请选择至少一项权限' }]}
          >
            <Checkbox.Group options={PERMISSION_OPTIONS} style={{ display: 'flex', flexWrap: 'wrap', gap: '8px 16px' }} />
          </Form.Item>
          <Form.Item name="status" label="状态" valuePropName="checked">
            <Switch checkedChildren="启用" unCheckedChildren="禁用" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}

// ===== 主页面组件 =====

const VALID_SYSTEM_PATHS = ['users', 'roles'];

export default function SystemPage() {
  const location = useLocation();
  const navigate = useNavigate();

  const pathParts = location.pathname.split('/');
  const rawKey = pathParts[2] || 'users';
  const activeKey = VALID_SYSTEM_PATHS.includes(rawKey) ? rawKey : 'users';

  // 历史/手输旧路径自动重定向
  useEffect(() => {
    if (pathParts[2] && !VALID_SYSTEM_PATHS.includes(pathParts[2])) {
      navigate('/system/users', { replace: true });
    }
  }, [location.pathname]);

  const handleTabChange = (key: string) => {
    navigate(`/system/${key}`);
  };

  const tabItems = [
    { key: 'users', label: '用户管理' },
    { key: 'roles', label: '角色管理' },
  ];

  const renderTabContent = () => {
    switch (activeKey) {
      case 'users': return <UsersTab />;
      case 'roles': return <RolesTab />;
      default: return <UsersTab />;
    }
  };

  return (
    <div>
      <Tabs activeKey={activeKey} onChange={handleTabChange} items={tabItems} />
      <div style={{ marginTop: 8 }}>
        {renderTabContent()}
      </div>
    </div>
  );
}
