import { useState } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import {
  Layout, Menu, Button, Avatar, Dropdown, theme, Breadcrumb,
} from 'antd';
import {
  DashboardOutlined, EnvironmentOutlined, RobotOutlined,
  ScheduleOutlined, AlertOutlined, BarChartOutlined,
  SettingOutlined, CloudServerOutlined, ToolOutlined,
  CameraOutlined, UserOutlined, LogoutOutlined, MenuFoldOutlined,
  MenuUnfoldOutlined, LineChartOutlined,
} from '@ant-design/icons';
import { useAuthStore } from '../../store/authStore';

const { Header, Sider, Content } = Layout;

const menuItems = [
  { key: '/dashboard', icon: <DashboardOutlined />, label: '仪表盘' },
  { key: '/monitor', icon: <EnvironmentOutlined />, label: '监控中心' },
  { key: '/stations', icon: <CloudServerOutlined />, label: '电站管理' },
  { key: '/robots', icon: <RobotOutlined />, label: '机器人管理' },
  { key: '/tasks', icon: <ScheduleOutlined />, label: '任务管理' },
  { key: '/alarms', icon: <AlertOutlined />, label: '告警中心' },
  { key: '/analytics', icon: <BarChartOutlined />, label: '数据分析' },
  { key: '/predictions', icon: <LineChartOutlined />, label: '智能预测' },
  {
    key: '/system',
    icon: <SettingOutlined />,
    label: '系统管理',
    children: [
      { key: '/system/users', label: '用户管理' },
      { key: '/system/roles', label: '角色管理' },
      { key: '/system/organizations', label: '组织架构' },
      { key: '/system/configs', label: '系统配置' },
      { key: '/system/dicts', label: '数据字典' },
      { key: '/system/audit-logs', label: '审计日志' },
      { key: '/system/login-logs', label: '登录日志' },
      { key: '/system/notify-templates', label: '通知模板' },
      { key: '/system/report-templates', label: '报告模板' },
      { key: '/system/backups', label: '备份恢复' },
    ],
  },
  { key: '/firmwares', icon: <CloudServerOutlined />, label: '固件管理' },
  { key: '/maintenance', icon: <ToolOutlined />, label: '维护保养' },
  { key: '/cameras', icon: <CameraOutlined />, label: '摄像头管理' },
];

const breadcrumbMap: Record<string, string> = {
  '/dashboard': '仪表盘',
  '/monitor': '监控中心',
  '/stations': '电站管理',
  '/robots': '机器人管理',
  '/tasks': '任务管理',
  '/alarms': '告警中心',
  '/analytics': '数据分析',
  '/predictions': '智能预测',
  '/firmwares': '固件管理',
  '/maintenance': '维护保养',
  '/cameras': '摄像头管理',
  '/system/users': '用户管理',
  '/system/roles': '角色管理',
  '/system/organizations': '组织架构',
  '/system/configs': '系统配置',
  '/system/dicts': '数据字典',
  '/system/audit-logs': '审计日志',
  '/system/login-logs': '登录日志',
  '/system/notify-templates': '通知模板',
  '/system/report-templates': '报告模板',
  '/system/backups': '备份恢复',
};

export default function AppLayout() {
  const [collapsed, setCollapsed] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout } = useAuthStore();
  const { token: { colorBgContainer, borderRadiusLG } } = theme.useToken();

  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(key);
  };

  const userMenuItems = [
    { key: 'profile', icon: <UserOutlined />, label: '个人中心' },
    { type: 'divider' as const },
    { key: 'logout', icon: <LogoutOutlined />, label: '退出登录', danger: true },
  ];

  const handleUserMenuClick = ({ key }: { key: string }) => {
    if (key === 'logout') {
      logout();
      navigate('/login');
    } else if (key === 'profile') {
      navigate('/profile');
    }
  };

  const selectedKeys = [location.pathname];
  const openKeys = ['/system'];

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider trigger={null} collapsible collapsed={collapsed} theme="dark">
        <div style={{
          height: 48, margin: 16, display: 'flex',
          alignItems: 'center', justifyContent: 'center', color: '#fff',
          fontWeight: 'bold', fontSize: collapsed ? 14 : 18,
        }}>
          {collapsed ? 'CC' : 'CCPlatform'}
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={selectedKeys}
          defaultOpenKeys={openKeys}
          items={menuItems}
          onClick={handleMenuClick}
        />
      </Sider>
      <Layout>
        <Header style={{
          padding: '0 24px', background: colorBgContainer,
          display: 'flex', alignItems: 'center', justifyContent: 'space-between',
        }}>
          <Button
            type="text"
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => setCollapsed(!collapsed)}
          />
          <Dropdown menu={{ items: userMenuItems, onClick: handleUserMenuClick }}>
            <div style={{ cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 8 }}>
              <Avatar icon={<UserOutlined />} />
              <span>{user?.real_name || user?.username || '管理员'}</span>
            </div>
          </Dropdown>
        </Header>
        <div style={{ padding: '12px 24px 0' }}>
          <Breadcrumb
            items={[
              { title: '首页' },
              ...(breadcrumbMap[location.pathname]
                ? [{ title: breadcrumbMap[location.pathname] }]
                : []),
            ]}
          />
        </div>
        <Content style={{
          margin: 12, padding: 24, background: colorBgContainer,
          borderRadius: borderRadiusLG, minHeight: 280, overflow: 'auto',
        }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}
