import client from './client';
import type { Alarm, AlarmRule, PageData } from '../types';

export const alarmApi = {
  list: (params?: Record<string, unknown>) =>
    client.get<PageData<Alarm>>('/alarms', { params }),

  getById: (id: string) =>
    client.get<Alarm>(`/alarms/${id}`),

  handle: (id: string, data: { handle_status: number; handle_remark?: string }) =>
    client.put<null>(`/alarms/${id}`, data),

  stats: () =>
    client.get<Record<string, number>>('/alarms/stats'),

  recent: () =>
    client.get<Alarm[]>('/alarms/recent'),

  trends: (params?: Record<string, unknown>) =>
    client.get<unknown[]>('/alarms/trends', { params }),

  rules: {
    list: (params?: Record<string, unknown>) =>
      client.get<PageData<AlarmRule>>('/alarm-rules', { params }),
    create: (data: Partial<AlarmRule>) =>
      client.post<AlarmRule>('/alarm-rules', data),
    update: (id: number, data: Partial<AlarmRule>) =>
      client.put<AlarmRule>(`/alarm-rules/${id}`, data),
    delete: (id: number) =>
      client.delete<null>(`/alarm-rules/${id}`),
  },
};
