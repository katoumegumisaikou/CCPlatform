import client from './client';
import type { LoginRequest, LoginResponse, User } from '../types';

export const authApi = {
  login: (data: LoginRequest) =>
    client.post<LoginResponse>('/auth/login', data),

  me: () =>
    client.get<User>('/auth/me'),

  captcha: () =>
    client.get<{ captcha_id: string; captcha_image: string }>('/auth/captcha'),
};
