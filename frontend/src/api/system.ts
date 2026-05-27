import client from './client';
import type {
  PageData, User, Role, Organization, Camera, Firmware, Maintenance,
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

