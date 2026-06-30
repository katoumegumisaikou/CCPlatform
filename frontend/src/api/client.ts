import axios from 'axios';
import { getToken, removeToken, removeUser } from '../utils';
import { message } from 'antd';

const BASE_URL = import.meta.env.VITE_API_BASE || 'http://localhost:8080/api/v1';

const instance = axios.create({
  baseURL: BASE_URL,
  timeout: 15000,
});

// 请求拦截器：注入 JWT Token
instance.interceptors.request.use((config) => {
  const token = getToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// 响应拦截器：统一错误处理，自动提取 data.data
instance.interceptors.response.use(
  (response) => {
    const { code, message: msg } = response.data;
    if (code !== 0) {
      if (code === 10001 || code === 10002) {
        removeToken();
        removeUser();
        window.location.href = '/login';
        return Promise.reject(new Error(msg || '认证失败'));
      }
      message.error(msg || '请求失败');
      return Promise.reject(new Error(msg));
    }
    return response.data.data;
  },
  (error) => {
    if (error.response?.status === 401) {
      removeToken();
      removeUser();
      window.location.href = '/login';
    }
    message.error(error.message || '网络错误');
    return Promise.reject(error);
  },
);

// 类型安全的请求包装器，匹配拦截器提取 .data.data 后的实际返回值
const client = {
  get: <T>(url: string, config?: Parameters<typeof instance.get>[1]): Promise<T> =>
    instance.get(url, config) as unknown as Promise<T>,
  post: <T>(url: string, data?: unknown, config?: Parameters<typeof instance.post>[2]): Promise<T> =>
    instance.post(url, data, config) as unknown as Promise<T>,
  put: <T>(url: string, data?: unknown, config?: Parameters<typeof instance.put>[2]): Promise<T> =>
    instance.put(url, data, config) as unknown as Promise<T>,
  delete: <T>(url: string, config?: Parameters<typeof instance.delete>[1]): Promise<T> =>
    instance.delete(url, config) as unknown as Promise<T>,
};

export default client;
