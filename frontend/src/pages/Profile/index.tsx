import { useState } from 'react';
import { Card, Descriptions, Form, Input, Button, message } from 'antd';
import {
  UserOutlined,
  LockOutlined,
  PhoneOutlined,
  MailOutlined,
} from '@ant-design/icons';
import { useAuthStore } from '../../store/authStore';

const ProfilePage: React.FC = () => {
  const user = useAuthStore((s) => s.user);
  const [changingPassword, setChangingPassword] = useState(false);
  const [form] = Form.useForm();

  const handleChangePassword = async (_values: {
    old_password: string;
    new_password: string;
    confirm_password: string;
  }) => {
    // TODO: 对接后端修改密码接口
    message.info('密码修改功能开发中，敬请期待');
    setChangingPassword(false);
    form.resetFields();
  };

  if (!user) {
    return (
      <div style={{ padding: 24, textAlign: 'center', color: '#999' }}>
        无法加载用户信息
      </div>
    );
  }

  return (
    <div style={{ padding: 24, maxWidth: 800, margin: '0 auto' }}>
      {/* 个人信息 */}
      <Card title="个人信息" style={{ marginBottom: 24 }}>
        <Descriptions column={2} bordered>
          <Descriptions.Item label="用户名">
            <UserOutlined style={{ marginRight: 8 }} />
            {user.username}
          </Descriptions.Item>
          <Descriptions.Item label="真实姓名">
            {user.real_name || '-'}
          </Descriptions.Item>
          <Descriptions.Item label="角色">
            {user.role?.role_name ?? '-'}
          </Descriptions.Item>
          <Descriptions.Item label="组织ID">
            {user.org_id ?? '-'}
          </Descriptions.Item>
          <Descriptions.Item label="手机号">
            <PhoneOutlined style={{ marginRight: 8 }} />
            {user.phone || '-'}
          </Descriptions.Item>
          <Descriptions.Item label="邮箱">
            <MailOutlined style={{ marginRight: 8 }} />
            {user.email || '-'}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      {/* 修改密码 */}
      <Card title="修改密码">
        {!changingPassword ? (
          <Button type="primary" onClick={() => setChangingPassword(true)}>
            修改密码
          </Button>
        ) : (
          <Form
            form={form}
            layout="vertical"
            style={{ maxWidth: 400 }}
            onFinish={handleChangePassword}
          >
            <Form.Item
              name="old_password"
              label="原密码"
              rules={[{ required: true, message: '请输入原密码' }]}
            >
              <Input.Password
                prefix={<LockOutlined />}
                placeholder="请输入原密码"
              />
            </Form.Item>

            <Form.Item
              name="new_password"
              label="新密码"
              rules={[
                { required: true, message: '请输入新密码' },
                { min: 6, message: '密码长度不能少于6位' },
              ]}
            >
              <Input.Password
                prefix={<LockOutlined />}
                placeholder="请输入新密码"
              />
            </Form.Item>

            <Form.Item
              name="confirm_password"
              label="确认新密码"
              dependencies={['new_password']}
              rules={[
                { required: true, message: '请确认新密码' },
                ({ getFieldValue }) => ({
                  validator(_, value) {
                    if (!value || getFieldValue('new_password') === value) {
                      return Promise.resolve();
                    }
                    return Promise.reject(new Error('两次输入的密码不一致'));
                  },
                }),
              ]}
            >
              <Input.Password
                prefix={<LockOutlined />}
                placeholder="请再次输入新密码"
              />
            </Form.Item>

            <Form.Item>
              <Button type="primary" htmlType="submit" style={{ marginRight: 12 }}>
                确认修改
              </Button>
              <Button
                onClick={() => {
                  setChangingPassword(false);
                  form.resetFields();
                }}
              >
                取消
              </Button>
            </Form.Item>
          </Form>
        )}
      </Card>
    </div>
  );
};

export default ProfilePage;
