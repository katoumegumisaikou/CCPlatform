import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, Form, Input, Button, message, Spin } from 'antd';
import { UserOutlined, LockOutlined, SafetyCertificateOutlined } from '@ant-design/icons';
import { useAuthStore } from '../../store/authStore';
import { authApi } from '../../api/auth';

interface CaptchaData {
  captcha_id: string;
  captcha_image: string;
}

const LoginPage: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [captcha, setCaptcha] = useState<CaptchaData | null>(null);
  const [captchaLoading, setCaptchaLoading] = useState(false);
  const navigate = useNavigate();
  const login = useAuthStore((s) => s.login);
  const token = useAuthStore((s) => s.token);

  // 已登录则直接跳转
  useEffect(() => {
    if (token) {
      navigate('/dashboard', { replace: true });
    }
  }, [token, navigate]);

  const fetchCaptcha = async () => {
    setCaptchaLoading(true);
    try {
      const res = await authApi.captcha();
      setCaptcha(res);
    } catch {
      // 验证码为可选功能，加载失败不影响登录
    } finally {
      setCaptchaLoading(false);
    }
  };

  const onFinish = async (values: {
    username: string;
    password: string;
    captcha_code?: string;
  }) => {
    setLoading(true);
    try {
      await login({
        username: values.username,
        password: values.password,
        captcha_id: captcha?.captcha_id,
        captcha_code: values.captcha_code,
      });
      message.success('登录成功');
      navigate('/dashboard', { replace: true });
    } catch {
      message.error('登录失败，请检查用户名和密码');
      // 登录失败后刷新验证码
      fetchCaptcha();
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      }}
    >
      <Card
        title={
          <div style={{ textAlign: 'center', fontSize: 22, fontWeight: 600 }}>
            拓达威云控平台
          </div>
        }
        style={{
          width: 420,
          boxShadow: '0 8px 24px rgba(0,0,0,0.15)',
        }}
        styles={{ body: { padding: '32px 40px' } }}
      >
        <Form
          name="login"
          size="large"
          onFinish={onFinish}
          autoComplete="off"
          initialValues={{ username: '', password: '' }}
        >
          <Form.Item
            name="username"
            rules={[{ required: true, message: '请输入用户名' }]}
          >
            <Input prefix={<UserOutlined />} placeholder="用户名" />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[{ required: true, message: '请输入密码' }]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="密码" />
          </Form.Item>

          {captcha && (
            <>
              <Form.Item name="captcha_code">
                <Input
                  prefix={<SafetyCertificateOutlined />}
                  placeholder="验证码"
                />
              </Form.Item>
              <Form.Item>
                <Spin spinning={captchaLoading}>
                  <img
                    src={captcha.captcha_image}
                    alt="验证码"
                    onClick={fetchCaptcha}
                    style={{
                      cursor: 'pointer',
                      height: 40,
                      border: '1px solid #d9d9d9',
                      borderRadius: 4,
                    }}
                    title="点击刷新验证码"
                  />
                </Spin>
              </Form.Item>
            </>
          )}

          <Form.Item style={{ marginBottom: 0 }}>
            <Button type="primary" htmlType="submit" loading={loading} block>
              登 录
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default LoginPage;
