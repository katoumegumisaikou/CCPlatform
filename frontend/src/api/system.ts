import client from './client';
import type {
  PageData, User, Role, Organization, SystemConfig, DictType, DictItem,
  NotifyTemplate, ReportTemplate, Camera, Firmware, Maintenance,
} from '../types';

export const userApi = {
  list: (params?: Record<string, unknown>) =>
    client.get<PageData<User>>('/users', { params }),
  create: (data: Partial<User>) =>
    client.post<User>('/users', data),
  update: (id: number, data: Partial<User>) =>
    client.put<User>(`/users/${id}`, data),
  delete: (id: number) =>
    client.delete<null>(`/users/${id}`),
};

export const roleApi = {
  list: () =>
    client.get<Role[]>('/roles'),
  create: (data: Partial<Role>) =>
    client.post<Role>('/roles', data),
  update: (id: number, data: Partial<Role>) =>
    client.put<Role>(`/roles/${id}`, data),
  delete: (id: number) =>
    client.delete<null>(`/roles/${id}`),
};

export const orgApi = {
  list: () =>
    client.get<Organization[]>('/organizations'),
  tree: () =>
    client.get<Organization[]>('/organizations/tree'),
  getById: (id: number) =>
    client.get<Organization>(`/organizations/${id}`),
  create: (data: Partial<Organization>) =>
    client.post<Organization>('/organizations', data),
  update: (id: number, data: Partial<Organization>) =>
    client.put<Organization>(`/organizations/${id}`, data),
  delete: (id: number) =>
    client.delete<null>(`/organizations/${id}`),
};

export const configApi = {
  list: () =>
    client.get<SystemConfig[]>('/system/configs'),
  get: (key: string) =>
    client.get<SystemConfig>(`/system/configs/${key}`),
  set: (key: string, value: string) =>
    client.put<null>(`/system/configs/${key}`, { config_value: value }),
  delete: (key: string) =>
    client.delete<null>(`/system/configs/${key}`),
};

export const dictApi = {
  types: {
    list: () => client.get<DictType[]>('/dicts'),
    create: (data: Partial<DictType>) => client.post<DictType>('/dicts', data),
    update: (id: number, data: Partial<DictType>) => client.put<DictType>(`/dicts/${id}`, data),
    delete: (id: number) => client.delete<null>(`/dicts/${id}`),
  },
  items: {
    list: (type: string) => client.get<DictItem[]>(`/dicts/${type}/items`),
    create: (type: string, data: Partial<DictItem>) =>
      client.post<DictItem>(`/dicts/${type}/items`, data),
    update: (id: number, data: Partial<DictItem>) => client.put<DictItem>(`/dicts/items/${id}`, data),
    delete: (id: number) => client.delete<null>(`/dicts/items/${id}`),
  },
};

export const notifyApi = {
  list: () =>
    client.get<NotifyTemplate[]>('/notify/templates'),
  create: (data: Partial<NotifyTemplate>) =>
    client.post<NotifyTemplate>('/notify/templates', data),
  update: (id: string, data: Partial<NotifyTemplate>) =>
    client.put<NotifyTemplate>(`/notify/templates/${id}`, data),
  delete: (id: string) =>
    client.delete<null>(`/notify/templates/${id}`),
};

export const reportApi = {
  list: () =>
    client.get<ReportTemplate[]>('/report-templates'),
  create: (data: Partial<ReportTemplate>) =>
    client.post<ReportTemplate>('/report-templates', data),
  update: (id: string, data: Partial<ReportTemplate>) =>
    client.put<ReportTemplate>(`/report-templates/${id}`, data),
  delete: (id: string) =>
    client.delete<null>(`/report-templates/${id}`),
};

export const cameraApi = {
  list: (params?: Record<string, unknown>) =>
    client.get<PageData<Camera>>('/cameras', { params }),
  create: (data: Partial<Camera>) =>
    client.post<Camera>('/cameras', data),
  update: (id: string, data: Partial<Camera>) =>
    client.put<Camera>(`/cameras/${id}`, data),
  delete: (id: string) =>
    client.delete<null>(`/cameras/${id}`),
};

export const firmwareApi = {
  list: (params?: Record<string, unknown>) =>
    client.get<PageData<Firmware>>('/firmwares', { params }),
  create: (data: FormData) =>
    client.post<Firmware>('/firmwares', data, {
      headers: { 'Content-Type': 'multipart/form-data' },
    }),
  update: (id: string, data: Partial<Firmware>) =>
    client.put<Firmware>(`/firmwares/${id}`, data),
  delete: (id: string) =>
    client.delete<null>(`/firmwares/${id}`),
  upgrade: (id: string, robotIds: string[]) =>
    client.post<null>(`/firmwares/${id}/upgrade`, { robot_ids: robotIds }),
};

export const maintenanceApi = {
  list: (params?: Record<string, unknown>) =>
    client.get<PageData<Maintenance>>('/maintenance', { params }),
  create: (data: Partial<Maintenance>) =>
    client.post<Maintenance>('/maintenance', data),
  update: (id: number, data: Partial<Maintenance>) =>
    client.put<Maintenance>(`/maintenance/${id}`, data),
  delete: (id: number) =>
    client.delete<null>(`/maintenance/${id}`),
  reminders: () =>
    client.get<Maintenance[]>('/maintenance/reminders'),
};

export const auditApi = {
  list: (params?: Record<string, unknown>) =>
    client.get<PageData<unknown>>('/audit-logs', { params }),
};

export const loginLogApi = {
  list: (params?: Record<string, unknown>) =>
    client.get<PageData<unknown>>('/login-logs', { params }),
};

export const backupApi = {
  list: () =>
    client.get<Array<{ filename: string; size: number; create_time: string }>>('/backups'),
  create: () =>
    client.post<null>('/backups'),
  delete: (filename: string) =>
    client.delete<null>(`/backups/${filename}`),
  restore: (filename: string) =>
    client.post<null>(`/backups/${filename}/restore`),
};
