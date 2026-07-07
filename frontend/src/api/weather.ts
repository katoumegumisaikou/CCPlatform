import client from './client';
import type { WeatherNow, WeatherDaily, WeatherHourly, WeatherWarning, WeatherAirQuality } from '../types';

export const weatherApi = {
  // 实时天气
  getNowWeather: (stationId: string) =>
    client.get<WeatherNow>('/weather/now', { params: { station_id: stationId } }),

  // 刷新天气数据
  refresh: (stationId: string) =>
    client.post<null>('/weather/refresh', { station_id: stationId }),

  // 7天逐天预报
  getDailyForecast: (stationId: string, days?: number) =>
    client.get<WeatherDaily[]>('/weather/forecast/daily', {
      params: { station_id: stationId, days },
    }),

  // 逐小时预报
  getHourlyForecast: (stationId: string, hours?: number) =>
    client.get<WeatherHourly[]>('/weather/forecast/hourly', {
      params: { station_id: stationId, hours },
    }),

  // 当前生效预警
  getActiveWarnings: (stationId: string) =>
    client.get<WeatherWarning[]>('/weather/warnings', {
      params: { station_id: stationId },
    }),

  // 历史预警（分页）
  getWarningHistory: (params: {
    station_id: string;
    start?: string;
    end?: string;
    page?: number;
    size?: number;
  }) =>
    client.get<{ list: WeatherWarning[]; total: number }>('/weather/warnings/history', {
      params,
    }),

  // 空气质量
  getAirQuality: (stationId: string) =>
    client.get<WeatherAirQuality>('/weather/air-quality', {
      params: { station_id: stationId },
    }),
};
