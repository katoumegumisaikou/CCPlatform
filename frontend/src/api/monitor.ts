import client from './client';
import type { DashboardOverview, RobotPosition, StationRealtime } from '../types';

export const monitorApi = {
  getDashboard: () =>
    client.get<DashboardOverview>('/monitor/overview'),

  getRealtime: (stationId: string) =>
    client.get<StationRealtime>(`/stations/${stationId}/realtime`),

  getTracks: (params: { robot_id: string; start_time: string; end_time: string; limit?: number }) =>
    client.get<RobotPosition[]>('/monitor/tracks', { params }),

  getEnvironment: (params: { robot_id: string; start_time: string; end_time: string }) =>
    client.get<unknown[]>('/monitor/environment', { params }),

  getGisConfig: () =>
    client.get<Record<string, string>>('/monitor/gis-config'),

  updateGisConfig: (data: Record<string, string>) =>
    client.put<null>('/monitor/gis-config', data),
};
