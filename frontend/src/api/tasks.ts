import client from './client';
import type { PageData, Task, TaskProgress } from '../types';

export const taskApi = {
  list: (params?: Record<string, unknown>) =>
    client.get<PageData<Task>>('/tasks', { params }),

  getById: (id: number) =>
    client.get<Task>(`/tasks/${id}`),

  create: (data: Partial<Task>) =>
    client.post<Task>('/tasks', data),

  updateStatus: (id: number, status: number) =>
    client.put<null>(`/tasks/${id}/status`, { status }),

  delete: (id: number) =>
    client.delete<null>(`/tasks/${id}`),

  getProgress: (id: number) =>
    client.get<TaskProgress>(`/tasks/${id}/progress`),

  smartSchedule: (stationId: string) =>
    client.post<Task>('/tasks/smart-schedule', { station_id: stationId }),
};
