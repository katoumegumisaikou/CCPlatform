import { useState, useEffect, Component, type ReactNode } from 'react';
import {
  Card, Tabs, Row, Col, Select, Button, Space, Statistic, Spin, Empty, Table, message, Descriptions, Tag, Result,
} from 'antd';
import {
  ReloadOutlined, CloudOutlined, FlagOutlined, DashboardOutlined,
} from '@ant-design/icons';
import ReactECharts from 'echarts-for-react';
import dayjs from 'dayjs';
import { weatherApi } from '../../api/weather';
import { stationApi } from '../../api/stations';
import type { Station, WeatherNow, WeatherDaily, WeatherHourly, WeatherAirQuality } from '../../types';

// ErrorBoundary 捕获组件渲染异常，防止整个页面白屏
class ErrorBoundary extends Component<{ children: ReactNode }, { hasError: boolean; error: string }> {
  constructor(props: { children: ReactNode }) {
    super(props);
    this.state = { hasError: false, error: '' };
  }
  static getDerivedStateFromError(err: Error) {
    return { hasError: true, error: err.message };
  }
  render() {
    if (this.state.hasError) {
      return <Result status="error" title="页面渲染异常" subTitle={this.state.error} />;
    }
    return this.props.children;
  }
}

type TabKey = 'now' | 'daily' | 'hourly';

const WEATHER_ICON_MAP: Record<string, string> = {
  '100': '☀️', '101': '🌤️', '102': '⛅', '103': '🌥️', '104': '☁️',
  '150': '🌙', '151': '🌙', '152': '🌙', '153': '🌙', '154': '☁️',
  '300': '🌦️', '301': '🌦️', '302': '🌧️', '303': '⛈️', '304': '🌨️',
  '305': '🌧️', '306': '🌧️', '307': '🌧️', '308': '🌧️', '309': '🌧️',
  '310': '🌧️', '311': '🌧️', '312': '🌧️', '313': '🌨️', '314': '🌨️',
  '400': '❄️', '401': '❄️', '402': '❄️', '403': '❄️', '404': '🌧️',
  '405': '🌨️', '406': '🌨️', '407': '🌨️', '500': '🌫️', '501': '🌫️',
  '502': '🌫️', '503': '💨', '504': '💨', '507': '💨', '508': '💨',
};
const getWeatherIcon = (code: string | undefined) => WEATHER_ICON_MAP[code || ''] || '🌡️';

const windLevelText = (scale: string | undefined): string => {
  const s = parseInt(scale || '0', 10);
  if (s <= 2) return '微风'; if (s <= 4) return '和风'; if (s <= 6) return '强风'; if (s <= 8) return '大风'; return '狂风';
};

const aqiColor = (level: string | undefined): string => {
  const l = parseInt(level || '1', 10);
  if (l <= 1) return '#52c41a'; if (l <= 2) return '#faad14'; if (l <= 3) return '#ff7a45'; if (l <= 4) return '#ff4d4f'; return '#8c0000';
};

export default function WeatherPage() {
  const [activeTab, setActiveTab] = useState<TabKey>('now');
  const [stations, setStations] = useState<Station[]>([]);
  const [stationId, setStationId] = useState<string | undefined>(undefined);
  const [loading, setLoading] = useState(false);
  const [lastRefresh, setLastRefresh] = useState('');

  const [nowWeather, setNowWeather] = useState<WeatherNow | null>(null);
  const [airQuality, setAirQuality] = useState<WeatherAirQuality | null>(null);
  const [dailyForecast, setDailyForecast] = useState<WeatherDaily[]>([]);
  const [hourlyForecast, setHourlyForecast] = useState<WeatherHourly[]>([]);

  useEffect(() => {
    stationApi.getAll().then(setStations).catch(() => {});
  }, []);

  const fetchWeatherData = async (sid: string) => {
    setLoading(true);
    try { const now = await weatherApi.getNowWeather(sid); setNowWeather(now); } catch { setNowWeather(null); }
    try { const daily = await weatherApi.getDailyForecast(sid, 7); setDailyForecast(daily || []); } catch { setDailyForecast([]); }
    try { const hourly = await weatherApi.getHourlyForecast(sid, 24); setHourlyForecast(hourly || []); } catch { setHourlyForecast([]); }
    try { const air = await weatherApi.getAirQuality(sid); setAirQuality(air); } catch { setAirQuality(null); }
    setLoading(false);
  };

  useEffect(() => {
    if (stationId) { fetchWeatherData(stationId); }
  }, [stationId]);

  const handleRefresh = async () => {
    if (!stationId) return;
    setLoading(true);
    try {
      await weatherApi.refresh(stationId);
      message.success('天气数据刷新成功');
      setLastRefresh(dayjs().format('HH:mm:ss'));
      await fetchWeatherData(stationId);
    } catch {
      message.error('天气数据刷新失败，请检查 API 配置');
      setLoading(false);
    }
  };

  // ---- 实时天气 Tab ----
  const renderNowTab = () => (
    <Spin spinning={loading}>
      {!nowWeather ? (
        <Empty description="暂无实时天气数据，请点击「刷新数据」按钮获取" />
      ) : (
        <>
          <Row gutter={16} style={{ marginBottom: 16 }}>
            <Col span={6}>
              <Card>
                <Statistic title={`${nowWeather.text || '-'} ${getWeatherIcon(nowWeather.icon)}`} value={nowWeather.temp || '--'} suffix="°C" />
                <div style={{ fontSize: 12, color: '#888', marginTop: 4 }}>体感 {nowWeather.feels_like || '--'}°C</div>
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic title={nowWeather.wind_dir ? `${nowWeather.wind_dir} · ${windLevelText(nowWeather.wind_scale)}` : '风力'} value={nowWeather.wind_speed || '--'} suffix="km/h" prefix="🌬️" />
                <div style={{ fontSize: 12, color: '#888', marginTop: 4 }}>{nowWeather.wind_scale || '-'} 级</div>
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic title="相对湿度" value={nowWeather.humidity || '--'} suffix="%" prefix="💧" />
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic title="能见度" value={nowWeather.vis || '--'} suffix="km" prefix="👁️" />
              </Card>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={12}>
              <Card title="气象详情" size="small">
                <Descriptions column={2} size="small">
                  <Descriptions.Item label="大气压强">{nowWeather.pressure || '-'} hPa</Descriptions.Item>
                  <Descriptions.Item label="降水量">{nowWeather.precip || '0'} mm</Descriptions.Item>
                  <Descriptions.Item label="云量">{nowWeather.cloud || '-'}%</Descriptions.Item>
                  <Descriptions.Item label="观测时间">{nowWeather.obs_time || '-'}</Descriptions.Item>
                </Descriptions>
              </Card>
            </Col>
            <Col span={12}>
              <Card title="空气质量" size="small">
                {airQuality ? (
                  <Row gutter={16}>
                    <Col span={12}>
                      <ReactECharts key={`aqi-${stationId}`} style={{ height: 160 }} opts={{ renderer: 'svg' }}
                        option={{ series: [{ type: 'gauge', startAngle: 210, endAngle: -30, min: 0, max: 500, progress: { show: true, width: 12 }, axisLine: { lineStyle: { width: 12 } }, axisTick: { show: false }, splitLine: { show: false }, axisLabel: { show: false }, detail: { formatter: '{value}', fontSize: 24, offsetCenter: [0, '60%'] }, data: [{ value: parseInt(airQuality.aqi || '0', 10) || 0, name: 'AQI' }] }] }}
                      />
                    </Col>
                    <Col span={12}>
                      <Descriptions column={1} size="small">
                        <Descriptions.Item label="等级"><Tag color={aqiColor(airQuality.level)}>{airQuality.category || '-'}</Tag></Descriptions.Item>
                        <Descriptions.Item label="首要污染物">{airQuality.primary || '无'}</Descriptions.Item>
                        <Descriptions.Item label="PM2.5">{airQuality.pm2p5 || '-'} μg/m³</Descriptions.Item>
                        <Descriptions.Item label="PM10">{airQuality.pm10 || '-'} μg/m³</Descriptions.Item>
                      </Descriptions>
                    </Col>
                  </Row>
                ) : <Empty description="暂无空气质量数据（免费套餐可能不支持）" />}
              </Card>
            </Col>
          </Row>
        </>
      )}
    </Spin>
  );

  // ---- 7天预报 Tab ----
  const renderDailyTab = () => (
    <Spin spinning={loading}>
      {dailyForecast.length === 0 ? (
        <Empty description="暂无预报数据" />
      ) : (
        <>
          <Card style={{ marginBottom: 16 }}>
            <ReactECharts key={`daily-${stationId}`} style={{ height: 350 }} opts={{ renderer: 'svg' }}
              option={{
                tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
                legend: { data: ['最高温度', '最低温度', '降水量'] },
                grid: { left: 50, right: 50, top: 30, bottom: 40 },
                xAxis: { type: 'category', data: dailyForecast.map(d => (d.fx_date || '').slice(5) + ' ' + (d.text_day || '')), axisLabel: { rotate: 30, fontSize: 11 } },
                yAxis: [{ type: 'value', name: '温度 (°C)' }, { type: 'value', name: '降水量 (mm)' }],
                series: [
                  { name: '最高温度', type: 'bar', data: dailyForecast.map(d => parseFloat(d.temp_max) || 0), itemStyle: { color: '#ff7a45' }, barMaxWidth: 16 },
                  { name: '最低温度', type: 'bar', data: dailyForecast.map(d => parseFloat(d.temp_min) || 0), itemStyle: { color: '#69b1ff' }, barMaxWidth: 16 },
                  { name: '降水量', type: 'line', yAxisIndex: 1, data: dailyForecast.map(d => parseFloat(d.precip) || 0), lineStyle: { color: '#1677ff', width: 2 }, itemStyle: { color: '#1677ff' }, symbol: 'circle' },
                ],
              }}
            />
          </Card>
          <Table dataSource={dailyForecast} rowKey="fx_date" size="small" pagination={false}
            columns={[
              { title: '日期', dataIndex: 'fx_date', width: 100 },
              { title: '白天', width: 90, render: (_: unknown, r: WeatherDaily) => `${getWeatherIcon(r.icon_day)} ${r.text_day || ''}` },
              { title: '夜间', width: 90, render: (_: unknown, r: WeatherDaily) => `${getWeatherIcon(r.icon_night)} ${r.text_night || ''}` },
              { title: '最高/最低', width: 100, render: (_: unknown, r: WeatherDaily) => <span><span style={{ color: '#ff4d4f' }}>{r.temp_max || '-'}°</span> / <span style={{ color: '#1677ff' }}>{r.temp_min || '-'}°</span></span> },
              { title: '风力风向', width: 100, render: (_: unknown, r: WeatherDaily) => `${r.wind_dir_day || '-'} ${r.wind_scale_day || ''}级` },
              { title: '湿度', dataIndex: 'humidity', width: 70, render: (v: string) => v ? `${v}%` : '-' },
              { title: '降水量', dataIndex: 'precip', width: 80, render: (v: string) => v ? `${v}mm` : '0' },
              { title: '紫外线', dataIndex: 'uv_index', width: 70 },
            ]}
          />
        </>
      )}
    </Spin>
  );

  // ---- 逐小时预报 Tab ----
  const renderHourlyTab = () => (
    <Spin spinning={loading}>
      {hourlyForecast.length === 0 ? (
        <Empty description="暂无逐小时预报数据" />
      ) : (
        <>
          <Card style={{ marginBottom: 16 }}>
            <ReactECharts key={`hourly-${stationId}`} style={{ height: 300 }} opts={{ renderer: 'svg' }}
              option={{
                tooltip: { trigger: 'axis' },
                legend: { data: ['温度', '湿度'] },
                grid: { left: 50, right: 50, top: 30, bottom: 60 },
                xAxis: { type: 'category', data: hourlyForecast.map(h => dayjs(h.fx_time).format('HH:mm')), axisLabel: { rotate: 45, fontSize: 10 } },
                yAxis: [{ type: 'value', name: '温度 (°C)' }, { type: 'value', name: '湿度 (%)' }],
                series: [
                  { name: '温度', type: 'line', data: hourlyForecast.map(h => parseFloat(h.temp) || 0), smooth: true, lineStyle: { color: '#ff7a45', width: 2 } },
                  { name: '湿度', type: 'line', yAxisIndex: 1, data: hourlyForecast.map(h => parseFloat(h.humidity) || 0), smooth: true, lineStyle: { color: '#69b1ff', width: 2 } },
                ],
              }}
            />
          </Card>
          <div style={{ overflowX: 'auto', whiteSpace: 'nowrap', paddingBottom: 8 }}>
            {hourlyForecast.map(h => (
              <Card key={h.fx_time} size="small"
                style={{ display: 'inline-block', width: 110, marginRight: 10, textAlign: 'center' }}
                styles={{ body: { padding: 8 } }}>
                <div style={{ fontSize: 12, color: '#888' }}>{dayjs(h.fx_time).format('HH:mm')}</div>
                <div style={{ fontSize: 24 }}>{getWeatherIcon(h.icon)}</div>
                <div style={{ fontSize: 11 }}>{h.text || '-'}</div>
                <div style={{ fontSize: 16, fontWeight: 600 }}>{h.temp || '-'}°C</div>
              </Card>
            ))}
          </div>
        </>
      )}
    </Spin>
  );

  return (
    <ErrorBoundary>
      <Card style={{ marginBottom: 16 }}>
        <Row align="middle" justify="space-between">
          <Col>
            <Space>
              <Select placeholder="选择电站" style={{ width: 260 }} value={stationId}
                onChange={(v) => { setStationId(v); setNowWeather(null); setDailyForecast([]); setHourlyForecast([]); setAirQuality(null); }}
                options={stations.map(s => ({ value: s.station_id, label: `${s.station_name} (${s.station_code})` }))}
                allowClear showSearch optionFilterProp="label" />
              <Button type="primary" icon={<ReloadOutlined spin={loading} />} onClick={handleRefresh}
                loading={loading} disabled={!stationId}>刷新数据</Button>
            </Space>
          </Col>
          <Col>{lastRefresh && <span style={{ color: '#888', fontSize: 12 }}>上次刷新: {lastRefresh}</span>}</Col>
        </Row>
      </Card>
      {!stationId ? (
        <Empty description="请先选择电站" style={{ marginTop: 80 }} />
      ) : (
        <Tabs activeKey={activeTab} onChange={(k) => setActiveTab(k as TabKey)}
          items={[
            { key: 'now', label: <span><DashboardOutlined /> 实时天气</span>, children: renderNowTab() },
            { key: 'daily', label: <span><FlagOutlined /> 7天预报</span>, children: renderDailyTab() },
            { key: 'hourly', label: <span><CloudOutlined /> 逐小时预报</span>, children: renderHourlyTab() },
          ]}
        />
      )}
    </ErrorBoundary>
  );
}
