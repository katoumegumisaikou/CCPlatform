import client from './client';
import type { PageData, Station } from '../types';

export const stationApi = {
  list: (params?: Record<string, unknown>) =>
    client.get<PageData<Station>>('/stations', { params }),

  getAll: () =>
    client.get<Station[]>('/stations/all'),

  getById: (id: string) =>
    client.get<Station>(`/stations/${id}`),

  create: (data: Partial<Station>) =>
    client.post<Station>('/stations', data),

  update: (id: string, data: Partial<Station>) =>
    client.put<Station>(`/stations/${id}`, data),

  delete: (id: string) =>
    client.delete<null>(`/stations/${id}`),

  getRobots: (id: string) =>
    client.get<unknown[]>(`/stations/${id}/robots`),
};
