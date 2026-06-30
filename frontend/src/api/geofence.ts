import client from './client';
import type { PageData, Geofence, GeofenceAlarm } from '../types';

export const geofenceApi = {
  // ===== 围栏管理 =====
  list: (params?: Record<string, unknown>) =>
    client.get<PageData<Geofence>>('/geofences', { params }),

  getAll: () =>
    client.get<Geofence[]>('/geofences/all'),

  getById: (id: string) =>
    client.get<Geofence>(`/geofences/${id}`),

  create: (data: Partial<Geofence>) =>
    client.post<Geofence>('/geofences', data),

  update: (id: string, data: Partial<Geofence>) =>
    client.put<Geofence>(`/geofences/${id}`, data),

  delete: (id: string) =>
    client.delete<null>(`/geofences/${id}`),

  // ===== 围栏告警 =====
  listAlarms: (params?: Record<string, unknown>) =>
    client.get<PageData<GeofenceAlarm>>('/geofence-alarms', { params }),

  getRecentAlarms: (limit?: number) =>
    client.get<GeofenceAlarm[]>('/geofence-alarms/recent', { params: { limit } }),

  handleAlarm: (id: string, data: { handle_status: number; remark?: string }) =>
    client.put<null>(`/geofence-alarms/${id}`, data),
};
