import client from './client';
import type { PageData, Robot } from '../types';

export const robotApi = {
  list: (params?: Record<string, unknown>) =>
    client.get<PageData<Robot>>('/robots', { params }),

  getById: (id: string) =>
    client.get<Robot>(`/robots/${id}`),

  create: (data: Partial<Robot>) =>
    client.post<Robot>('/robots', data),

  update: (id: string, data: Partial<Robot>) =>
    client.put<Robot>(`/robots/${id}`, data),

  delete: (id: string) =>
    client.delete<null>(`/robots/${id}`),

  stats: () =>
    client.get<Record<string, number>>('/robots/stats'),

  sendCommand: (id: string, cmd: string, params?: Record<string, unknown>) =>
    client.post<null>(`/robots/${id}/cmd`, { cmd, params }),

  advancedStats: (id: string) =>
    client.get<Record<string, number>>(`/robots/${id}/advanced-stats`),

  applyConfig: (id: string, configId: string) =>
    client.post<null>(`/robots/${id}/config/apply`, { config_id: configId }),
};
