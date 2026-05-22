import { useState, useEffect, useCallback } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import {
  Tabs, Table, Button, Modal, Form, Input, Select, message,
  Popconfirm, Space, Tag, Card, Tree, Checkbox, DatePicker,
  InputNumber, Switch, Empty, Spin, Tooltip, Row, Col,
} from 'antd';
import {
  PlusOutlined, EditOutlined, DeleteOutlined, ReloadOutlined,
  SearchOutlined, ExclamationCircleOutlined,
  RollbackOutlined, FolderOpenOutlined,
} from '@ant-design/icons';
import type { DataNode } from 'antd/es/tree';
import type { ColumnsType } from 'antd/es/table';
import {
  userApi, roleApi, orgApi, configApi, dictApi,
  notifyApi, reportApi, backupApi, auditApi, loginLogApi,
} from '../../api/system';
import client from '../../api/client';
import type {
  User, Role, Organization, SystemConfig, DictType, DictItem,
  NotifyTemplate, ReportTemplate,
} from '../../types';
import type { Dayjs } from 'dayjs';

const { TextArea } = Input;
const { RangePicker } = DatePicker;

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

const TPL_TYPE_OPTIONS = [
  { label: '短信', value: 'sms' },
  { label: '邮件', value: 'email' },
  { label: '应用内通知', value: 'app' },
];

const ALARM_LEVEL_OPTIONS = [
  { label: '提示', value: 1 },
  { label: '一般', value: 2 },
  { label: '严重', value: 3 },
  { label: '紧急', value: 4 },
];

const STATUS_OPTIONS = [
  { label: '启用', value: 1 },
  { label: '禁用', value: 0 },
];

// ===== 工具函数 =====

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`;
}

function formatDateTime(dateStr: string): string {
  if (!dateStr) return '-';
  const d = new Date(dateStr);
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
}

// ===== 审计日志本地类型 =====

interface AuditLogEntry {
  id: number;
  operator: string;
  action: string;
  target: string;
  detail: string;
  ip_address: string;
  created_at: string;
}

interface LoginLogEntry {
  id: number;
  username: string;
  ip_address: string;
  login_time: string;
  result: string;
}

interface BackupEntry {
  filename: string;
  size: number;
  create_time: string;
}

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
  const [orgs, setOrgs] = useState<Organization[]>([]);
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

  const fetchRolesAndOrgs = useCallback(async () => {
    try {
      const [rolesRes, orgsRes] = await Promise.all([
        roleApi.list(),
        orgApi.list(),
      ]);
      setRoles(rolesRes || []);
      setOrgs(orgsRes || []);
    } catch { /* ignore */ }
  }, []);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleAdd = () => {
    setEditingUser(null);
    form.resetFields();
    form.setFieldsValue({ status: 1 });
    fetchRolesAndOrgs();
    setModalOpen(true);
  };

  const handleEdit = (record: User) => {
    setEditingUser(record);
    form.setFieldsValue({
      username: record.username,
      real_name: record.real_name,
      role_id: record.role_id,
      org_id: record.org_id,
      phone: record.phone,
      email: record.email,
      status: record.status,
    });
    fetchRolesAndOrgs();
    setModalOpen(true);
  };

  const handleOk = async () => {
    try {
      const values = await form.validateFields();
      if (editingUser) {
        await userApi.update(editingUser.user_id, values);
        message.success('用户更新成功');
      } else {
        await userApi.create(values);
        message.success('用户创建成功');
      }
      setModalOpen(false);
      fetchData();
    } catch (err: any) {
      if (err?.errorFields) return; // 表单校验失败
    }
  };

  const handleDelete = async (id: number) => {
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
          <Form.Item name="org_id" label="所属组织" rules={[{ required: true, message: '请选择组织' }]}>
            <Select
              placeholder="请选择组织"
              options={orgs.map(o => ({ label: o.org_name, value: o.org_id }))}
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

  const handleDelete = async (id: number) => {
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

// ===== 3. 组织架构 =====

function OrganizationsTab() {
  const [treeData, setTreeData] = useState<DataNode[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedOrg, setSelectedOrg] = useState<Organization | null>(null);
  const [selectedKeys, setSelectedKeys] = useState<(string | number)[]>([]);
  const [modalOpen, setModalOpen] = useState(false);
  const [editForm] = Form.useForm();
  const [modalForm] = Form.useForm();

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await orgApi.tree();
      const orgs = res || [];
      setTreeData(convertToTreeData(orgs));
    } catch {
      setTreeData([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { fetchData(); }, [fetchData]);

  const convertToTreeData = (orgs: Organization[]): DataNode[] => {
    return orgs.map(org => ({
      key: org.org_id,
      title: org.org_name,
      children: org.children ? convertToTreeData(org.children) : undefined,
      data: org,
    })) as DataNode[];
  };

  const findOrgInTree = (nodes: DataNode[], id: number): Organization | null => {
    for (const node of nodes) {
      const org = (node as any).data as Organization;
      if (org && org.org_id === id) return org;
      if (node.children) {
        const found = findOrgInTree(node.children, id);
        if (found) return found;
      }
    }
    return null;
  };

  const handleSelect = (keys: React.Key[]) => {
    setSelectedKeys(keys as (string | number)[]);
    if (keys.length > 0) {
      const org = findOrgInTree(treeData, Number(keys[0]));
      setSelectedOrg(org);
      if (org) {
        editForm.setFieldsValue({
          org_name: org.org_name,
          org_code: org.org_code,
          parent_id: org.parent_id,
          status: org.status === 1,
        });
      }
    } else {
      setSelectedOrg(null);
      editForm.resetFields();
    }
  };

  const handleAddChild = () => {
    if (!selectedOrg) {
      message.warning('请先选择父组织');
      return;
    }
    modalForm.resetFields();
    setModalOpen(true);
  };

  const handleModalOk = async () => {
    try {
      const values = await modalForm.validateFields();
      await orgApi.create({
        org_name: values.org_name,
        org_code: values.org_code,
        parent_id: selectedOrg!.org_id,
      });
      message.success('子组织创建成功');
      setModalOpen(false);
      fetchData();
      setSelectedOrg(null);
      setSelectedKeys([]);
      editForm.resetFields();
    } catch (err: any) {
      if (err?.errorFields) return;
    }
  };

  const handleEditSave = async () => {
    if (!selectedOrg) return;
    try {
      const values = await editForm.validateFields();
      await orgApi.update(selectedOrg.org_id, {
        org_name: values.org_name,
        org_code: values.org_code,
        parent_id: selectedOrg.parent_id,
        status: values.status ? 1 : 0,
      });
      message.success('组织更新成功');
      fetchData();
    } catch (err: any) {
      if (err?.errorFields) return;
    }
  };

  const handleDelete = async () => {
    if (!selectedOrg) return;
    if (selectedOrg.children && selectedOrg.children.length > 0) {
      message.warning('该组织下存在子组织，无法删除');
      return;
    }
    try {
      await orgApi.delete(selectedOrg.org_id);
      message.success('组织已删除');
      setSelectedOrg(null);
      setSelectedKeys([]);
      editForm.resetFields();
      fetchData();
    } catch { /* ignore */ }
  };

  return (
    <div>
      {loading ? (
        <div style={{ textAlign: 'center', padding: 80 }}><Spin size="large" tip="加载组织架构..." /></div>
      ) : treeData.length === 0 ? (
        <Empty description="暂无组织数据" style={{ padding: 80 }} />
      ) : (
        <Row gutter={24}>
          <Col span={10}>
            <Card
              title="组织架构树"
              extra={
                <Space>
                  <Button type="primary" size="small" icon={<PlusOutlined />} onClick={handleAddChild}>
                    添加子组织
                  </Button>
                  <Button size="small" icon={<ReloadOutlined />} onClick={fetchData} />
                </Space>
              }
              style={{ height: 520, overflow: 'auto' }}
            >
              <Tree
                showLine={{ showLeafIcon: false }}
                defaultExpandAll
                selectedKeys={selectedKeys}
                onSelect={handleSelect}
                treeData={treeData}
                style={{ marginTop: 8 }}
              />
            </Card>
          </Col>
          <Col span={14}>
            <Card
              title={selectedOrg ? `组织详情 - ${selectedOrg.org_name}` : '组织详情'}
              extra={
                selectedOrg ? (
                  <Popconfirm
                    title={
                      selectedOrg.children && selectedOrg.children.length > 0
                        ? '该组织存在子组织，无法删除'
                        : '确定删除该组织吗？'
                    }
                    onConfirm={handleDelete}
                    okText="确认"
                    cancelText="取消"
                    disabled={!!(selectedOrg.children && selectedOrg.children.length > 0)}
                  >
                    <Button danger size="small" icon={<DeleteOutlined />}>
                      删除
                    </Button>
                  </Popconfirm>
                ) : null
              }
              style={{ height: 520 }}
            >
              {selectedOrg ? (
                <Form form={editForm} layout="vertical">
                  <Form.Item name="org_name" label="组织名称" rules={[{ required: true, message: '请输入组织名称' }]}>
                    <Input placeholder="请输入组织名称" />
                  </Form.Item>
                  <Form.Item name="org_code" label="组织编码" rules={[{ required: true, message: '请输入组织编码' }]}>
                    <Input placeholder="请输入组织编码" />
                  </Form.Item>
                  <Form.Item label="父组织ID">
                    <Input value={selectedOrg.parent_id || '无（根组织）'} disabled />
                  </Form.Item>
                  <Form.Item name="status" label="状态" valuePropName="checked">
                    <Switch checkedChildren="启用" unCheckedChildren="禁用" />
                  </Form.Item>
                  <Form.Item>
                    <Button type="primary" onClick={handleEditSave}>保存修改</Button>
                  </Form.Item>
                </Form>
              ) : (
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: 400 }}>
                  <Empty description="请在左侧选择组织" />
                </div>
              )}
            </Card>
          </Col>
        </Row>
      )}
      <Modal
        title="添加子组织"
        open={modalOpen}
        onOk={handleModalOk}
        onCancel={() => setModalOpen(false)}
        destroyOnClose
      >
        <Form form={modalForm} layout="vertical" preserve={false}>
          <Form.Item label="父组织">
            <Input value={selectedOrg?.org_name || ''} disabled />
          </Form.Item>
          <Form.Item name="org_name" label="组织名称" rules={[{ required: true, message: '请输入组织名称' }]}>
            <Input placeholder="请输入组织名称" />
          </Form.Item>
          <Form.Item name="org_code" label="组织编码" rules={[{ required: true, message: '请输入组织编码' }]}>
            <Input placeholder="请输入组织编码" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}

// ===== 4. 系统配置 =====

function ConfigsTab() {
  const [configs, setConfigs] = useState<SystemConfig[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingConfig, setEditingConfig] = useState<SystemConfig | null>(null);
  const [form] = Form.useForm();

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await configApi.list();
      setConfigs(res || []);
    } catch {
      setConfigs([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleAdd = () => {
    setEditingConfig(null);
    form.resetFields();
    setModalOpen(true);
  };

  const handleEdit = (record: SystemConfig) => {
    setEditingConfig(record);
    form.setFieldsValue({
      config_key: record.config_key,
      config_value: record.config_value,
      category: record.category,
      description: record.description,
    });
    setModalOpen(true);
  };

  const handleOk = async () => {
    try {
      const values = await form.validateFields();
      if (editingConfig) {
        // 编辑：更新配置值和元数据
        await configApi.set(editingConfig.config_key, values.config_value);
        // 尝试更新 category 和 description（如果后端支持）
        try {
          await client.put(`/system/configs/${editingConfig.config_key}`, {
            config_value: values.config_value,
            category: values.category,
            description: values.description,
          });
        } catch { /* 忽略额外字段不支持的情况 */ }
        message.success('配置更新成功');
      } else {
        // 新增：创建新配置项
        await client.post('/system/configs', {
          config_key: values.config_key,
          config_value: values.config_value,
          category: values.category || '',
          description: values.description || '',
        });
        message.success('配置创建成功');
      }
      setModalOpen(false);
      fetchData();
    } catch (err: any) {
      if (err?.errorFields) return;
    }
  };

  const handleDelete = async (key: string) => {
    try {
      await configApi.delete(key);
      message.success('配置已删除');
      fetchData();
    } catch { /* ignore */ }
  };

  const columns: ColumnsType<SystemConfig> = [
    { title: '配置键', dataIndex: 'config_key', key: 'config_key', width: 180, ellipsis: true },
    { title: '配置值', dataIndex: 'config_value', key: 'config_value', width: 200, ellipsis: true },
    {
      title: '分类',
      dataIndex: 'category',
      key: 'category',
      width: 100,
      render: (c: string) => c ? <Tag>{c}</Tag> : <Tag color="default">未分类</Tag>,
    },
    { title: '描述', dataIndex: 'description', key: 'description', ellipsis: true },
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
            title="确定删除该配置吗？"
            onConfirm={() => handleDelete(record.config_key)}
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
            添加配置
          </Button>
          <Button icon={<ReloadOutlined />} onClick={fetchData}>刷新</Button>
        </Space>
      </div>
      <Table
        rowKey="config_key"
        columns={columns}
        dataSource={configs}
        loading={loading}
        locale={{ emptyText: <Empty description="暂无配置数据" /> }}
        pagination={false}
      />
      <Modal
        title={editingConfig ? '编辑配置' : '添加配置'}
        open={modalOpen}
        onOk={handleOk}
        onCancel={() => setModalOpen(false)}
        destroyOnClose
        width={520}
      >
        <Form form={form} layout="vertical" preserve={false}>
          <Form.Item name="config_key" label="配置键" rules={[{ required: true, message: '请输入配置键' }]}>
            <Input disabled={!!editingConfig} placeholder="请输入配置键（如 app.max_retry）" />
          </Form.Item>
          <Form.Item name="config_value" label="配置值" rules={[{ required: true, message: '请输入配置值' }]}>
            <Input.TextArea rows={3} placeholder="请输入配置值" />
          </Form.Item>
          <Form.Item name="category" label="分类">
            <Input placeholder="如：系统、通知、安全" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={2} placeholder="请输入配置说明" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}

// ===== 5. 数据字典 =====

function DictsTab() {
  const [dictTypes, setDictTypes] = useState<DictType[]>([]);
  const [dictItems, setDictItems] = useState<DictItem[]>([]);
  const [typesLoading, setTypesLoading] = useState(false);
  const [itemsLoading, setItemsLoading] = useState(false);
  const [selectedType, setSelectedType] = useState<DictType | null>(null);
  // 字典类型弹窗
  const [typeModalOpen, setTypeModalOpen] = useState(false);
  const [editingType, setEditingType] = useState<DictType | null>(null);
  const [typeForm] = Form.useForm();
  // 字典项弹窗
  const [itemModalOpen, setItemModalOpen] = useState(false);
  const [editingItem, setEditingItem] = useState<DictItem | null>(null);
  const [itemForm] = Form.useForm();

  const fetchTypes = useCallback(async () => {
    setTypesLoading(true);
    try {
      const res = await dictApi.types.list();
      setDictTypes(res || []);
    } catch {
      setDictTypes([]);
    } finally {
      setTypesLoading(false);
    }
  }, []);

  const fetchItems = useCallback(async (dictType: string) => {
    setItemsLoading(true);
    try {
      const res = await dictApi.items.list(dictType);
      setDictItems(res || []);
    } catch {
      setDictItems([]);
    } finally {
      setItemsLoading(false);
    }
  }, []);

  useEffect(() => { fetchTypes(); }, [fetchTypes]);

  useEffect(() => {
    if (selectedType) {
      fetchItems(selectedType.dict_type);
    } else {
      setDictItems([]);
    }
  }, [selectedType, fetchItems]);

  // ===== 字典类型 CRUD =====

  const handleAddType = () => {
    setEditingType(null);
    typeForm.resetFields();
    typeForm.setFieldsValue({ status: 1 });
    setTypeModalOpen(true);
  };

  const handleEditType = (record: DictType) => {
    setEditingType(record);
    typeForm.setFieldsValue({
      dict_type: record.dict_type,
      dict_name: record.dict_name,
      status: record.status,
    });
    setTypeModalOpen(true);
  };

  const handleTypeOk = async () => {
    try {
      const values = await typeForm.validateFields();
      if (editingType) {
        await dictApi.types.update(editingType.id, values);
        message.success('字典类型更新成功');
      } else {
        await dictApi.types.create(values);
        message.success('字典类型创建成功');
      }
      setTypeModalOpen(false);
      fetchTypes();
    } catch (err: any) {
      if (err?.errorFields) return;
    }
  };

  const handleDeleteType = async (id: number) => {
    try {
      await dictApi.types.delete(id);
      message.success('字典类型已删除');
      if (selectedType?.id === id) {
        setSelectedType(null);
      }
      fetchTypes();
    } catch { /* ignore */ }
  };

  // ===== 字典项 CRUD =====

  const handleAddItem = () => {
    if (!selectedType) {
      message.warning('请先在左侧选择字典类型');
      return;
    }
    setEditingItem(null);
    itemForm.resetFields();
    itemForm.setFieldsValue({ status: 1, sort_order: 0 });
    setItemModalOpen(true);
  };

  const handleEditItem = (record: DictItem) => {
    setEditingItem(record);
    itemForm.setFieldsValue({
      dict_key: record.dict_key,
      dict_value: record.dict_value,
      sort_order: record.sort_order,
      status: record.status,
    });
    setItemModalOpen(true);
  };

  const handleItemOk = async () => {
    if (!selectedType) return;
    try {
      const values = await itemForm.validateFields();
      if (editingItem) {
        await dictApi.items.update(editingItem.id, values);
        message.success('字典项更新成功');
      } else {
        await dictApi.items.create(selectedType.dict_type, values);
        message.success('字典项创建成功');
      }
      setItemModalOpen(false);
      fetchItems(selectedType.dict_type);
    } catch (err: any) {
      if (err?.errorFields) return;
    }
  };

  const handleDeleteItem = async (id: number) => {
    try {
      await dictApi.items.delete(id);
      message.success('字典项已删除');
      if (selectedType) fetchItems(selectedType.dict_type);
    } catch { /* ignore */ }
  };

  // ===== 列定义 =====

  const typeColumns: ColumnsType<DictType> = [
    { title: '字典编码', dataIndex: 'dict_type', key: 'dict_type', width: 140 },
    { title: '字典名称', dataIndex: 'dict_name', key: 'dict_name' },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 70,
      render: (s: number) => <Tag color={s === 1 ? 'green' : 'red'}>{s === 1 ? '启用' : '禁用'}</Tag>,
    },
    {
      title: '操作',
      key: 'actions',
      width: 120,
      render: (_, record) => (
        <Space size={0}>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={(e) => { e.stopPropagation(); handleEditType(record); }} />
          <Popconfirm
            title="确定删除该字典类型吗？这将同时删除所有相关字典项。"
            onConfirm={() => handleDeleteType(record.id)}
            okText="确认"
            cancelText="取消"
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />} onClick={(e) => e.stopPropagation()} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const itemColumns: ColumnsType<DictItem> = [
    { title: '字典键', dataIndex: 'dict_key', key: 'dict_key', width: 140 },
    { title: '字典值', dataIndex: 'dict_value', key: 'dict_value' },
    { title: '排序', dataIndex: 'sort_order', key: 'sort_order', width: 70 },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 70,
      render: (s: number) => <Tag color={s === 1 ? 'green' : 'red'}>{s === 1 ? '启用' : '禁用'}</Tag>,
    },
    {
      title: '操作',
      key: 'actions',
      width: 120,
      render: (_, record) => (
        <Space size={0}>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEditItem(record)} />
          <Popconfirm
            title="确定删除该字典项吗？"
            onConfirm={() => handleDeleteItem(record.id)}
            okText="确认"
            cancelText="取消"
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <Row gutter={24}>
      {/* 左：字典类型 */}
      <Col span={11}>
        <Card
          title="字典类型"
          extra={
            <Space>
              <Button type="primary" size="small" icon={<PlusOutlined />} onClick={handleAddType}>添加</Button>
              <Button size="small" icon={<ReloadOutlined />} onClick={fetchTypes} />
            </Space>
          }
        >
          <Table
            rowKey="id"
            columns={typeColumns}
            dataSource={dictTypes}
            loading={typesLoading}
            size="small"
            locale={{ emptyText: <Empty description="暂无字典类型" /> }}
            pagination={false}
            scroll={{ y: 420 }}
            onRow={(record) => ({
              onClick: () => setSelectedType(record),
              style: {
                cursor: 'pointer',
                background: selectedType?.id === record.id ? '#e6f4ff' : undefined,
              },
            })}
          />
        </Card>
      </Col>
      {/* 右：字典项 */}
      <Col span={13}>
        <Card
          title={selectedType ? `字典项 - ${selectedType.dict_name}` : '字典项'}
          extra={
            <Space>
              <Button type="primary" size="small" icon={<PlusOutlined />} onClick={handleAddItem} disabled={!selectedType}>
                添加
              </Button>
              <Button
                size="small"
                icon={<ReloadOutlined />}
                disabled={!selectedType}
                onClick={() => selectedType && fetchItems(selectedType.dict_type)}
              />
            </Space>
          }
        >
          {selectedType ? (
            <Table
              rowKey="id"
              columns={itemColumns}
              dataSource={dictItems}
              loading={itemsLoading}
              size="small"
              locale={{ emptyText: <Empty description="暂无字典项" /> }}
              pagination={false}
              scroll={{ y: 420 }}
            />
          ) : (
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: 420 }}>
              <Empty description="请先在左侧选择字典类型" />
            </div>
          )}
        </Card>
      </Col>

      {/* 字典类型弹窗 */}
      <Modal
        title={editingType ? '编辑字典类型' : '添加字典类型'}
        open={typeModalOpen}
        onOk={handleTypeOk}
        onCancel={() => setTypeModalOpen(false)}
        destroyOnClose
      >
        <Form form={typeForm} layout="vertical" preserve={false}>
          <Form.Item name="dict_type" label="字典编码" rules={[{ required: true, message: '请输入字典编码' }]}>
            <Input disabled={!!editingType} placeholder="请输入字典编码（如 robot_type）" />
          </Form.Item>
          <Form.Item name="dict_name" label="字典名称" rules={[{ required: true, message: '请输入字典名称' }]}>
            <Input placeholder="请输入字典名称" />
          </Form.Item>
          <Form.Item name="status" label="状态">
            <Select options={STATUS_OPTIONS} />
          </Form.Item>
        </Form>
      </Modal>

      {/* 字典项弹窗 */}
      <Modal
        title={editingItem ? '编辑字典项' : '添加字典项'}
        open={itemModalOpen}
        onOk={handleItemOk}
        onCancel={() => setItemModalOpen(false)}
        destroyOnClose
      >
        <Form form={itemForm} layout="vertical" preserve={false}>
          <Form.Item name="dict_key" label="字典键" rules={[{ required: true, message: '请输入字典键' }]}>
            <Input placeholder="请输入字典键（如 1）" />
          </Form.Item>
          <Form.Item name="dict_value" label="字典值" rules={[{ required: true, message: '请输入字典值' }]}>
            <Input placeholder="请输入字典值（如 固定式）" />
          </Form.Item>
          <Form.Item name="sort_order" label="排序">
            <InputNumber min={0} style={{ width: '100%' }} placeholder="排序号" />
          </Form.Item>
          <Form.Item name="status" label="状态">
            <Select options={STATUS_OPTIONS} />
          </Form.Item>
        </Form>
      </Modal>
    </Row>
  );
}

// ===== 6. 审计日志 =====

function AuditLogsTab() {
  const [logs, setLogs] = useState<AuditLogEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [dateRange, setDateRange] = useState<[Dayjs | null, Dayjs | null] | null>(null);
  const [actionFilter, setActionFilter] = useState<string | undefined>(undefined);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const params: Record<string, unknown> = { page, size: pageSize };
      if (dateRange && dateRange[0] && dateRange[1]) {
        params.start_time = dateRange[0].format('YYYY-MM-DD');
        params.end_time = dateRange[1].format('YYYY-MM-DD');
      }
      if (actionFilter) params.action = actionFilter;
      const res = await auditApi.list(params);
      const data = res as { list: AuditLogEntry[]; total: number };
      setLogs(data.list || []);
      setTotal(data.total || 0);
    } catch {
      setLogs([]);
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, dateRange, actionFilter]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleSearch = () => {
    setPage(1);
    fetchData();
  };

  const columns: ColumnsType<AuditLogEntry> = [
    { title: '操作人', dataIndex: 'operator', key: 'operator', width: 100 },
    {
      title: '操作类型',
      dataIndex: 'action',
      key: 'action',
      width: 100,
      render: (a: string) => <Tag color="blue">{a}</Tag>,
    },
    { title: '操作对象', dataIndex: 'target', key: 'target', width: 150, ellipsis: true },
    { title: '操作详情', dataIndex: 'detail', key: 'detail', ellipsis: true },
    { title: 'IP地址', dataIndex: 'ip_address', key: 'ip_address', width: 140 },
    {
      title: '操作时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (t: string) => formatDateTime(t),
    },
  ];

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Space wrap>
          <RangePicker
            value={dateRange as any}
            onChange={(dates) => setDateRange(dates as [Dayjs | null, Dayjs | null] | null)}
            placeholder={['开始日期', '结束日期']}
          />
          <Input
            placeholder="操作类型"
            value={actionFilter}
            onChange={(e) => setActionFilter(e.target.value)}
            style={{ width: 150 }}
            allowClear
          />
          <Button type="primary" icon={<SearchOutlined />} onClick={handleSearch}>
            搜索
          </Button>
          <Button icon={<ReloadOutlined />} onClick={fetchData}>刷新</Button>
        </Space>
      </div>
      <Table
        rowKey="id"
        columns={columns}
        dataSource={logs}
        loading={loading}
        locale={{ emptyText: <Empty description="暂无审计日志" /> }}
        pagination={{
          current: page,
          pageSize,
          total,
          showSizeChanger: true,
          showTotal: (t: number) => `共 ${t} 条`,
          onChange: (p: number, s: number) => { setPage(p); setPageSize(s); },
        }}
      />
    </div>
  );
}

// ===== 7. 登录日志 =====

function LoginLogsTab() {
  const [logs, setLogs] = useState<LoginLogEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [dateRange, setDateRange] = useState<[Dayjs | null, Dayjs | null] | null>(null);
  const [usernameFilter, setUsernameFilter] = useState('');

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const params: Record<string, unknown> = { page, size: pageSize };
      if (dateRange && dateRange[0] && dateRange[1]) {
        params.start_time = dateRange[0].format('YYYY-MM-DD');
        params.end_time = dateRange[1].format('YYYY-MM-DD');
      }
      if (usernameFilter) params.username = usernameFilter;
      const res = await loginLogApi.list(params);
      const data = res as { list: LoginLogEntry[]; total: number };
      setLogs(data.list || []);
      setTotal(data.total || 0);
    } catch {
      setLogs([]);
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, dateRange, usernameFilter]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleSearch = () => {
    setPage(1);
    fetchData();
  };

  const columns: ColumnsType<LoginLogEntry> = [
    { title: '用户名', dataIndex: 'username', key: 'username', width: 130 },
    { title: '登录IP', dataIndex: 'ip_address', key: 'ip_address', width: 150 },
    {
      title: '登录时间',
      dataIndex: 'login_time',
      key: 'login_time',
      width: 180,
      render: (t: string) => formatDateTime(t),
    },
    {
      title: '登录结果',
      dataIndex: 'result',
      key: 'result',
      width: 100,
      render: (r: string) => {
        const isSuccess = r === 'success' || r === '成功';
        return <Tag color={isSuccess ? 'green' : 'red'}>{r}</Tag>;
      },
    },
  ];

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Space wrap>
          <RangePicker
            value={dateRange as any}
            onChange={(dates) => setDateRange(dates as [Dayjs | null, Dayjs | null] | null)}
            placeholder={['开始日期', '结束日期']}
          />
          <Input
            placeholder="用户名"
            value={usernameFilter}
            onChange={(e) => setUsernameFilter(e.target.value)}
            style={{ width: 150 }}
            allowClear
          />
          <Button type="primary" icon={<SearchOutlined />} onClick={handleSearch}>
            搜索
          </Button>
          <Button icon={<ReloadOutlined />} onClick={fetchData}>刷新</Button>
        </Space>
      </div>
      <Table
        rowKey="id"
        columns={columns}
        dataSource={logs}
        loading={loading}
        locale={{ emptyText: <Empty description="暂无登录日志" /> }}
        pagination={{
          current: page,
          pageSize,
          total,
          showSizeChanger: true,
          showTotal: (t: number) => `共 ${t} 条`,
          onChange: (p: number, s: number) => { setPage(p); setPageSize(s); },
        }}
      />
    </div>
  );
}

// ===== 8. 通知模板 =====

function NotifyTemplatesTab() {
  const [templates, setTemplates] = useState<NotifyTemplate[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingTpl, setEditingTpl] = useState<NotifyTemplate | null>(null);
  const [form] = Form.useForm();

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await notifyApi.list();
      setTemplates(res || []);
    } catch {
      setTemplates([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleAdd = () => {
    setEditingTpl(null);
    form.resetFields();
    form.setFieldsValue({ status: 1 });
    setModalOpen(true);
  };

  const handleEdit = (record: NotifyTemplate) => {
    setEditingTpl(record);
    form.setFieldsValue({
      tpl_name: record.tpl_name,
      tpl_type: record.tpl_type,
      alarm_level: record.alarm_level,
      content: record.content,
      status: record.status,
    });
    setModalOpen(true);
  };

  const handleOk = async () => {
    try {
      const values = await form.validateFields();
      if (editingTpl) {
        await notifyApi.update(editingTpl.tpl_id, values);
        message.success('模板更新成功');
      } else {
        await notifyApi.create(values);
        message.success('模板创建成功');
      }
      setModalOpen(false);
      fetchData();
    } catch (err: any) {
      if (err?.errorFields) return;
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await notifyApi.delete(id);
      message.success('模板已删除');
      fetchData();
    } catch { /* ignore */ }
  };

  const columns: ColumnsType<NotifyTemplate> = [
    { title: '模板名', dataIndex: 'tpl_name', key: 'tpl_name', width: 150 },
    {
      title: '类型',
      dataIndex: 'tpl_type',
      key: 'tpl_type',
      width: 110,
      render: (t: string) => {
        const map: Record<string, { text: string; color: string }> = {
          sms: { text: '短信', color: 'blue' },
          email: { text: '邮件', color: 'green' },
          app: { text: '应用通知', color: 'orange' },
        };
        return <Tag color={map[t]?.color || 'default'}>{map[t]?.text || t}</Tag>;
      },
    },
    {
      title: '告警级别',
      dataIndex: 'alarm_level',
      key: 'alarm_level',
      width: 90,
      render: (l: number) => {
        const map: Record<number, { text: string; color: string }> = {
          1: { text: '提示', color: 'blue' },
          2: { text: '一般', color: 'orange' },
          3: { text: '严重', color: 'red' },
          4: { text: '紧急', color: '#ff0000' },
        };
        return <Tag color={map[l]?.color}>{map[l]?.text || l}</Tag>;
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
            title="确定删除该模板吗？"
            onConfirm={() => handleDelete(record.tpl_id)}
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
            添加模板
          </Button>
          <Button icon={<ReloadOutlined />} onClick={fetchData}>刷新</Button>
        </Space>
      </div>
      <Table
        rowKey="tpl_id"
        columns={columns}
        dataSource={templates}
        loading={loading}
        locale={{ emptyText: <Empty description="暂无通知模板" /> }}
        pagination={false}
      />
      <Modal
        title={editingTpl ? '编辑通知模板' : '添加通知模板'}
        open={modalOpen}
        onOk={handleOk}
        onCancel={() => setModalOpen(false)}
        destroyOnClose
        width={560}
      >
        <Form form={form} layout="vertical" preserve={false}>
          <Form.Item name="tpl_name" label="模板名" rules={[{ required: true, message: '请输入模板名' }]}>
            <Input placeholder="请输入模板名" />
          </Form.Item>
          <Form.Item name="tpl_type" label="通知类型" rules={[{ required: true, message: '请选择通知类型' }]}>
            <Select options={TPL_TYPE_OPTIONS} placeholder="请选择通知类型" />
          </Form.Item>
          <Form.Item name="alarm_level" label="告警级别">
            <Select options={ALARM_LEVEL_OPTIONS} placeholder="请选择告警级别" allowClear />
          </Form.Item>
          <Form.Item name="content" label="模板内容" rules={[{ required: true, message: '请输入模板内容' }]}>
            <TextArea rows={5} placeholder="请输入模板内容，支持变量占位符" />
          </Form.Item>
          <Form.Item name="status" label="状态">
            <Select options={STATUS_OPTIONS} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}

// ===== 9. 报告模板 =====

function ReportTemplatesTab() {
  const [templates, setTemplates] = useState<ReportTemplate[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingTpl, setEditingTpl] = useState<ReportTemplate | null>(null);
  const [form] = Form.useForm();

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await reportApi.list();
      setTemplates(res || []);
    } catch {
      setTemplates([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleAdd = () => {
    setEditingTpl(null);
    form.resetFields();
    form.setFieldsValue({ status: 1 });
    setModalOpen(true);
  };

  const handleEdit = (record: ReportTemplate) => {
    setEditingTpl(record);
    form.setFieldsValue({
      tpl_name: record.tpl_name,
      tpl_type: record.tpl_type,
      dimensions: record.dimensions,
      metrics: record.metrics,
      status: record.status,
    });
    setModalOpen(true);
  };

  const handleOk = async () => {
    try {
      const values = await form.validateFields();
      if (editingTpl) {
        await reportApi.update(editingTpl.tpl_id, values);
        message.success('模板更新成功');
      } else {
        await reportApi.create(values);
        message.success('模板创建成功');
      }
      setModalOpen(false);
      fetchData();
    } catch (err: any) {
      if (err?.errorFields) return;
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await reportApi.delete(id);
      message.success('模板已删除');
      fetchData();
    } catch { /* ignore */ }
  };

  const formatJsonPreview = (jsonStr: string): string => {
    if (!jsonStr) return '-';
    try {
      const obj = JSON.parse(jsonStr);
      return JSON.stringify(obj, null, 2);
    } catch {
      return jsonStr;
    }
  };

  const columns: ColumnsType<ReportTemplate> = [
    { title: '模板名', dataIndex: 'tpl_name', key: 'tpl_name', width: 150 },
    {
      title: '类型',
      dataIndex: 'tpl_type',
      key: 'tpl_type',
      width: 100,
      render: (t: string) => <Tag color="purple">{t}</Tag>,
    },
    {
      title: '维度',
      dataIndex: 'dimensions',
      key: 'dimensions',
      width: 200,
      ellipsis: true,
      render: (d: string) => (
        <Tooltip title={<pre style={{ margin: 0, fontSize: 12 }}>{formatJsonPreview(d)}</pre>}>
          <span>{d ? d.substring(0, 40) + (d.length > 40 ? '...' : '') : '-'}</span>
        </Tooltip>
      ),
    },
    {
      title: '指标',
      dataIndex: 'metrics',
      key: 'metrics',
      width: 200,
      ellipsis: true,
      render: (m: string) => (
        <Tooltip title={<pre style={{ margin: 0, fontSize: 12 }}>{formatJsonPreview(m)}</pre>}>
          <span>{m ? m.substring(0, 40) + (m.length > 40 ? '...' : '') : '-'}</span>
        </Tooltip>
      ),
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
            title="确定删除该模板吗？"
            onConfirm={() => handleDelete(record.tpl_id)}
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
            添加模板
          </Button>
          <Button icon={<ReloadOutlined />} onClick={fetchData}>刷新</Button>
        </Space>
      </div>
      <Table
        rowKey="tpl_id"
        columns={columns}
        dataSource={templates}
        loading={loading}
        locale={{ emptyText: <Empty description="暂无报告模板" /> }}
        pagination={false}
      />
      <Modal
        title={editingTpl ? '编辑报告模板' : '添加报告模板'}
        open={modalOpen}
        onOk={handleOk}
        onCancel={() => setModalOpen(false)}
        destroyOnClose
        width={560}
      >
        <Form form={form} layout="vertical" preserve={false}>
          <Form.Item name="tpl_name" label="模板名" rules={[{ required: true, message: '请输入模板名' }]}>
            <Input placeholder="请输入模板名" />
          </Form.Item>
          <Form.Item name="tpl_type" label="类型" rules={[{ required: true, message: '请输入模板类型' }]}>
            <Input placeholder="如：日报、周报、月报" />
          </Form.Item>
          <Form.Item
            name="dimensions"
            label="维度（JSON格式）"
            rules={[
              { required: true, message: '请输入维度' },
              {
                validator: (_, value) => {
                  if (!value) return Promise.resolve();
                  try { JSON.parse(value); return Promise.resolve(); }
                  catch { return Promise.reject('请输入有效的 JSON 格式'); }
                },
              },
            ]}
          >
            <TextArea rows={3} placeholder='如：["电站","机器人","日期"]' />
          </Form.Item>
          <Form.Item
            name="metrics"
            label="指标（JSON格式）"
            rules={[
              { required: true, message: '请输入指标' },
              {
                validator: (_, value) => {
                  if (!value) return Promise.resolve();
                  try { JSON.parse(value); return Promise.resolve(); }
                  catch { return Promise.reject('请输入有效的 JSON 格式'); }
                },
              },
            ]}
          >
            <TextArea rows={3} placeholder='如：["清扫面积","工作时长","告警次数"]' />
          </Form.Item>
          <Form.Item name="status" label="状态">
            <Select options={STATUS_OPTIONS} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}

// ===== 10. 备份恢复 =====

function BackupsTab() {
  const [backups, setBackups] = useState<BackupEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [creating, setCreating] = useState(false);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await backupApi.list();
      setBackups(res || []);
    } catch {
      setBackups([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleCreate = async () => {
    setCreating(true);
    try {
      await backupApi.create();
      message.success('备份创建成功');
      fetchData();
    } catch { /* ignore */ }
    finally { setCreating(false); }
  };

  const handleRestore = (filename: string) => {
    Modal.confirm({
      title: '确认恢复',
      icon: <ExclamationCircleOutlined />,
      content: '恢复备份将覆盖当前所有数据，确定要恢复吗？',
      okText: '确认',
      cancelText: '取消',
      onOk: () => {
        Modal.confirm({
          title: '最终确认',
          icon: <ExclamationCircleOutlined />,
          content: '此操作不可逆！恢复后当前数据将被永久替换为备份数据。确定继续吗？',
          okText: '确认恢复',
          okType: 'danger',
          cancelText: '取消',
          onOk: async () => {
            try {
              await backupApi.restore(filename);
              message.success('数据恢复成功');
              fetchData();
            } catch { /* ignore */ }
          },
        });
      },
    });
  };

  const handleDelete = async (filename: string) => {
    try {
      await backupApi.delete(filename);
      message.success('备份已删除');
      fetchData();
    } catch { /* ignore */ }
  };

  const columns: ColumnsType<BackupEntry> = [
    {
      title: '文件名',
      dataIndex: 'filename',
      key: 'filename',
      ellipsis: true,
      render: (name: string) => (
        <Space>
          <FolderOpenOutlined />
          <span>{name}</span>
        </Space>
      ),
    },
    {
      title: '大小',
      dataIndex: 'size',
      key: 'size',
      width: 100,
      render: (s: number) => formatFileSize(s),
    },
    {
      title: '创建时间',
      dataIndex: 'create_time',
      key: 'create_time',
      width: 180,
      render: (t: string) => formatDateTime(t),
    },
    {
      title: '操作',
      key: 'actions',
      width: 180,
      render: (_, record) => (
        <Space>
          <Popconfirm
            title="确定恢复到此备份吗？"
            description="恢复操作将覆盖当前数据。"
            onConfirm={() => handleRestore(record.filename)}
            okText="确认"
            cancelText="取消"
          >
            <Button type="link" size="small" icon={<RollbackOutlined />}>
              恢复
            </Button>
          </Popconfirm>
          <Popconfirm
            title="确定删除该备份吗？"
            onConfirm={() => handleDelete(record.filename)}
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
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={handleCreate}
            loading={creating}
          >
            创建备份
          </Button>
          <Button icon={<ReloadOutlined />} onClick={fetchData}>刷新</Button>
        </Space>
      </div>
      <Table
        rowKey="filename"
        columns={columns}
        dataSource={backups}
        loading={loading}
        locale={{ emptyText: <Empty description="暂无备份数据" /> }}
        pagination={false}
      />
    </div>
  );
}

// ===== 主页面组件 =====

export default function SystemPage() {
  const location = useLocation();
  const navigate = useNavigate();

  const pathParts = location.pathname.split('/');
  const activeKey = pathParts[2] || 'users';

  const handleTabChange = (key: string) => {
    navigate(`/system/${key}`);
  };

  const tabItems = [
    { key: 'users', label: '用户管理' },
    { key: 'roles', label: '角色管理' },
    { key: 'organizations', label: '组织架构' },
    { key: 'configs', label: '系统配置' },
    { key: 'dicts', label: '数据字典' },
    { key: 'audit-logs', label: '审计日志' },
    { key: 'login-logs', label: '登录日志' },
    { key: 'notify-templates', label: '通知模板' },
    { key: 'report-templates', label: '报告模板' },
    { key: 'backups', label: '备份恢复' },
  ];

  const renderTabContent = () => {
    switch (activeKey) {
      case 'users': return <UsersTab />;
      case 'roles': return <RolesTab />;
      case 'organizations': return <OrganizationsTab />;
      case 'configs': return <ConfigsTab />;
      case 'dicts': return <DictsTab />;
      case 'audit-logs': return <AuditLogsTab />;
      case 'login-logs': return <LoginLogsTab />;
      case 'notify-templates': return <NotifyTemplatesTab />;
      case 'report-templates': return <ReportTemplatesTab />;
      case 'backups': return <BackupsTab />;
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
