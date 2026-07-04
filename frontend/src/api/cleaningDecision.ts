import client from './client';
import type { CleaningDecisionRule, ExtremeWeatherRule, CleaningAdvice, CleaningScheduleAdjustment } from '../types';

export const cleaningDecisionApi = {
  // ===== 决策规则 =====
  getRules: (params: { station_id?: string; page?: number; size?: number }) =>
    client.get<{ list: CleaningDecisionRule[]; total: number }>('/cleaning-decisions/rules', { params }),

  createRule: (data: Partial<CleaningDecisionRule>) =>
    client.post<CleaningDecisionRule>('/cleaning-decisions/rules', data),

  getRule: (id: string) =>
    client.get<CleaningDecisionRule>(`/cleaning-decisions/rules/${id}`),

  updateRule: (id: string, data: Partial<CleaningDecisionRule>) =>
    client.put<null>(`/cleaning-decisions/rules/${id}`, data),

  deleteRule: (id: string) =>
    client.delete<null>(`/cleaning-decisions/rules/${id}`),

  // ===== 极端天气规则 =====
  getExtremeRules: (params: { station_id?: string; page?: number; size?: number }) =>
    client.get<{ list: ExtremeWeatherRule[]; total: number }>('/cleaning-decisions/extreme-rules', { params }),

  createExtremeRule: (data: Partial<ExtremeWeatherRule>) =>
    client.post<ExtremeWeatherRule>('/cleaning-decisions/extreme-rules', data),

  updateExtremeRule: (id: string, data: Partial<ExtremeWeatherRule>) =>
    client.put<null>(`/cleaning-decisions/extreme-rules/${id}`, data),

  deleteExtremeRule: (id: string) =>
    client.delete<null>(`/cleaning-decisions/extreme-rules/${id}`),

  // ===== 清扫建议 =====
  getAdvice: (stationId: string) =>
    client.get<CleaningAdvice>('/cleaning-decisions/advice', { params: { station_id: stationId } }),

  // ===== 频次调整 =====
  getAdjustments: (params: { station_id: string; page?: number; size?: number }) =>
    client.get<{ list: CleaningScheduleAdjustment[]; total: number }>('/cleaning-decisions/adjustments', { params }),

  generateAdjustment: (stationId: string) =>
    client.post<CleaningScheduleAdjustment>('/cleaning-decisions/adjustments', { station_id: stationId }),
};
