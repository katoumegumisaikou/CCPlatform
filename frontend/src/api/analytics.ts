import client from './client';

export const analyticsApi = {
  economic: (params?: Record<string, unknown>) =>
    client.get<unknown>('/analytics/economic', { params }),

  efficiency: (params?: Record<string, unknown>) =>
    client.get<unknown>('/analytics/efficiency', { params }),

  reports: (params?: Record<string, unknown>) =>
    client.get<unknown>('/analytics/reports', { params }),

  compare: (params?: Record<string, unknown>) =>
    client.get<unknown>('/analytics/compare', { params }),

  timeseries: (params?: Record<string, unknown>) =>
    client.get<unknown>('/analytics/timeseries', { params }),

  export: (params?: Record<string, unknown>) =>
    client.get('/analytics/export', { params, responseType: 'blob' }),

  predictions: {
    efficiency: (params?: Record<string, unknown>) =>
      client.get<unknown>('/analytics/predictions/efficiency', { params }),
    fault: (params?: Record<string, unknown>) =>
      client.get<unknown>('/analytics/predictions/fault', { params }),
    strategy: (params?: Record<string, unknown>) =>
      client.get<unknown>('/analytics/strategy/optimize', { params }),
    spareParts: (params?: Record<string, unknown>) =>
      client.get<unknown>('/analytics/predictions/spare-parts', { params }),
    maintenanceWindow: (params?: Record<string, unknown>) =>
      client.get<unknown>('/analytics/predictions/maintenance-window', { params }),
    generateWorkorder: (predictionId: string) =>
      client.post<unknown>(`/analytics/predictions/${predictionId}/workorder`),
  },

  health: {
    overview: (params?: Record<string, unknown>) =>
      client.get<unknown>('/analytics/health', { params }),
    byStation: (params?: Record<string, unknown>) =>
      client.get<unknown>('/analytics/health/station', { params }),
    trend: (params?: Record<string, unknown>) =>
      client.get<unknown>('/analytics/health/trend', { params }),
  },
};
