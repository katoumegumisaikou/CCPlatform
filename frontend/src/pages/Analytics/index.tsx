import { useState, useEffect, useCallback } from 'react';
import {
  Card, Tabs, Row, Col, Select, DatePicker, Button, Space, Statistic, Spin, Empty, message,
} from 'antd';
import {
  ReloadOutlined, DownloadOutlined,
} from '@ant-design/icons';
import ReactECharts from 'echarts-for-react';
import type { Dayjs } from 'dayjs';
import { analyticsApi } from '../../api/analytics';
import { stationApi } from '../../api/stations';
import type { Station } from '../../types';
import { ROBOT_TYPE_MAP } from '../../utils';

const { RangePicker } = DatePicker;

type TabKey = 'economic' | 'efficiency' | 'runtime' | 'compare' | 'timeseries';

interface SummaryCard {
  title: string;
  value: number;
  prefix?: string;
  suffix?: string;
  precision?: number;
}

export default function AnalyticsPage() {
  const [activeTab, setActiveTab] = useState<TabKey>('economic');
  const [stations, setStations] = useState<Station[]>([]);
  const [stationId, setStationId] = useState<string | undefined>(undefined);
  const [dateRange, setDateRange] = useState<[Dayjs, Dayjs] | null>(null);
  const [robotType, setRobotType] = useState<number | undefined>(undefined);
  const [loading, setLoading] = useState(false);
  const [chartLoading, setChartLoading] = useState(false);

  // economic
  const [economicSummary, setEconomicSummary] = useState<SummaryCard[]>([]);
  const [economicOption, setEconomicOption] = useState<object>({});

  // efficiency
  const [efficiencyLineOption, setEfficiencyLineOption] = useState<object>({});
  const [efficiencyGaugeOption, setEfficiencyGaugeOption] = useState<object>({});
  const [currentEfficiency, setCurrentEfficiency] = useState(0);

  // runtime
  const [runtimeOption, setRuntimeOption] = useState<object>({});

  // compare
  const [compareOption, setCompareOption] = useState<object>({});

  // timeseries
  const [timeseriesOption, setTimeseriesOption] = useState<object>({});

  const [hasData, setHasData] = useState(true);

  // 获取电站列表
  useEffect(() => {
    stationApi.getAll().then((res) => {
      setStations(res || []);
    }).catch(() => {
      message.error('获取电站列表失败');
    });
  }, []);

  const buildQuery = useCallback((): Record<string, unknown> => {
    const params: Record<string, unknown> = {};
    if (stationId) params.station_id = stationId;
    if (robotType !== undefined) params.robot_type = robotType;
    if (dateRange && dateRange[0] && dateRange[1]) {
      params.start_date = dateRange[0].format('YYYY-MM-DD');
      params.end_date = dateRange[1].format('YYYY-MM-DD');
    }
    return params;
  }, [stationId, dateRange, robotType]);

  // 加载当前 tab 数据
  const fetchData = useCallback(async () => {
    const params = buildQuery();
    setChartLoading(true);
    try {
      switch (activeTab) {
        case 'economic': {
          const res = await analyticsApi.economic(params);
          const list = Array.isArray(res) ? res : ((res as Record<string, unknown>)?.list as unknown[] || []);
          if (list.length === 0) {
            setHasData(false);
            break;
          }
          setHasData(true);
          const totalRevenue = list.reduce((s: number, it: any) => s + (it.revenue || 0), 0);
          const totalCost = list.reduce((s: number, it: any) => s + (it.cost || 0), 0);
          const totalProfit = totalRevenue - totalCost;
          setEconomicSummary([
            { title: '总收入（元）', value: totalRevenue, prefix: '¥', precision: 2 },
            { title: '总成本（元）', value: totalCost, prefix: '¥', precision: 2 },
            { title: '总利润（元）', value: totalProfit, prefix: '¥', precision: 2 },
          ]);
          setEconomicOption({
            tooltip: { trigger: 'axis' },
            legend: { data: ['收入', '成本', '利润'] },
            xAxis: {
              type: 'category',
              data: list.map((it: any) => it.date || it.period || ''),
              axisLabel: { rotate: 30 },
            },
            yAxis: { type: 'value', name: '金额（元）' },
            series: [
              { name: '收入', type: 'bar', data: list.map((it: any) => it.revenue || 0), itemStyle: { color: '#52c41a' } },
              { name: '成本', type: 'bar', data: list.map((it: any) => it.cost || 0), itemStyle: { color: '#ff7875' } },
              { name: '利润', type: 'bar', data: list.map((it: any) => (it.revenue || 0) - (it.cost || 0)), itemStyle: { color: '#1890ff' } },
            ],
            grid: { left: '3%', right: '4%', bottom: '15%', containLabel: true },
          });
          break;
        }
        case 'efficiency': {
          const res = await analyticsApi.efficiency(params);
          const list = Array.isArray(res) ? res : ((res as Record<string, unknown>)?.list as unknown[] || []);
          if (list.length === 0) {
            setHasData(false);
            break;
          }
          setHasData(true);
          const efficiencyValues = list.map((it: any) => it.efficiency || it.rate || 0);
          const avgEfficiency = efficiencyValues.length > 0
            ? efficiencyValues.reduce((s: number, v: number) => s + v, 0) / efficiencyValues.length
            : 0;
          setCurrentEfficiency(avgEfficiency);
          setEfficiencyLineOption({
            tooltip: { trigger: 'axis' },
            xAxis: {
              type: 'category',
              data: list.map((it: any) => it.date || it.period || ''),
              axisLabel: { rotate: 30 },
            },
            yAxis: { type: 'value', name: '效率（%）', max: 100 },
            series: [{
              name: '清扫效率',
              type: 'line',
              data: efficiencyValues,
              smooth: true,
              itemStyle: { color: '#1890ff' },
              areaStyle: { color: 'rgba(24,144,255,0.2)' },
            }],
            grid: { left: '3%', right: '4%', bottom: '15%', containLabel: true },
          });
          setEfficiencyGaugeOption({
            series: [{
              type: 'gauge',
              startAngle: 200,
              endAngle: -20,
              min: 0,
              max: 100,
              splitNumber: 10,
              axisLine: { lineStyle: { width: 20, color: [[0.3, '#ff4d4f'], [0.7, '#faad14'], [1, '#52c41a']] } },
              pointer: { itemStyle: { color: 'auto' } },
              axisTick: { distance: -25, length: 6, lineStyle: { width: 2, color: '#999' } },
              splitLine: { distance: -30, length: 20, lineStyle: { width: 3, color: '#999' } },
              axisLabel: { color: '#999', distance: 35, fontSize: 12 },
              detail: { valueAnimation: true, formatter: '{value}%', fontSize: 20, offsetCenter: [0, '60%'] },
              data: [{ value: Math.round(avgEfficiency * 100) / 100, name: '平均效率' }],
            }],
          });
          break;
        }
        case 'runtime': {
          const res = await analyticsApi.reports(params);
          const list = Array.isArray(res) ? res : ((res as Record<string, unknown>)?.list as unknown[] || []);
          if (list.length === 0) {
            setHasData(false);
            break;
          }
          setHasData(true);
          const days = list.map((it: any) => it.date || it.period || '');
          setRuntimeOption({
            tooltip: { trigger: 'axis' },
            legend: { data: ['清扫时长（小时）', '清扫面积（平方米）'] },
            xAxis: { type: 'category', data: days, axisLabel: { rotate: 30 } },
            yAxis: [
              { type: 'value', name: '清扫时长（小时）' },
              { type: 'value', name: '清扫面积（平方米）' },
            ],
            series: [
              { name: '清扫时长（小时）', type: 'bar', data: list.map((it: any) => it.hours || it.cleaning_hours || 0), itemStyle: { color: '#1890ff' } },
              { name: '清扫面积（平方米）', type: 'line', yAxisIndex: 1, data: list.map((it: any) => it.area || it.clean_area || 0), itemStyle: { color: '#52c41a' } },
            ],
            grid: { left: '3%', right: '4%', bottom: '15%', containLabel: true },
          });
          break;
        }
        case 'compare': {
          const res = await analyticsApi.compare(params);
          const list = Array.isArray(res) ? res : ((res as Record<string, unknown>)?.list as unknown[] || []);
          if (list.length === 0) {
            setHasData(false);
            break;
          }
          setHasData(true);
          const groups = Array.from(new Set(list.map((it: any) => it.group || it.station_name || it.name || '')));
          const categories = Array.from(new Set(list.map((it: any) => it.category || it.metric || '')));
          const seriesData = groups.map((group: unknown) => ({
            name: group,
            type: 'bar' as const,
            data: list
              .filter((it: any) => (it.group || it.station_name || it.name || '') === group)
              .map((it: any) => it.value || it.amount || 0),
          }));
          setCompareOption({
            tooltip: { trigger: 'axis' },
            legend: { data: groups as string[] },
            xAxis: { type: 'category', data: categories as string[], axisLabel: { rotate: 30 } },
            yAxis: { type: 'value' },
            series: seriesData,
            grid: { left: '3%', right: '4%', bottom: '15%', containLabel: true },
          });
          break;
        }
        case 'timeseries': {
          const res = await analyticsApi.timeseries(params);
          const list = Array.isArray(res) ? res : ((res as Record<string, unknown>)?.list as unknown[] || []);
          if (list.length === 0) {
            setHasData(false);
            break;
          }
          setHasData(true);
          const xAxisData = list.map((it: any) => it.time || it.date || it.timestamp || '');
          const keys = Object.keys(list[0] || {}).filter((k: string) => k !== 'time' && k !== 'date' && k !== 'timestamp');
          setTimeseriesOption({
            tooltip: { trigger: 'axis' },
            legend: { data: keys },
            xAxis: { type: 'category', data: xAxisData, axisLabel: { rotate: 30 } },
            yAxis: { type: 'value' },
            series: keys.map((k: string) => ({
              name: k,
              type: 'line',
              data: list.map((it: any) => it[k] ?? 0),
              smooth: true,
            })),
            grid: { left: '3%', right: '4%', bottom: '15%', containLabel: true },
          });
          break;
        }
        default:
          break;
      }
    } catch {
      message.error('获取数据失败');
      setHasData(false);
    } finally {
      setChartLoading(false);
    }
  }, [activeTab, buildQuery]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleExport = async () => {
    const params = buildQuery();
    setLoading(true);
    try {
      const res = await analyticsApi.export(params);
      const url = window.URL.createObjectURL(new Blob([res as any]));
      const a = document.createElement('a');
      a.href = url;
      a.download = `analytics_${activeTab}_${Date.now()}.xlsx`;
      document.body.appendChild(a);
      a.click();
      a.remove();
      window.URL.revokeObjectURL(url);
      message.success('导出成功');
    } catch {
      message.error('导出失败');
    } finally {
      setLoading(false);
    }
  };

  const tabItems = [
    {
      key: 'economic',
      label: '经济效益',
      children: (
        <Spin spinning={chartLoading}>
          {!hasData ? <Empty description="暂无经济效益数据" /> : (
            <>
              <Row gutter={16} style={{ marginBottom: 24 }}>
                {economicSummary.map((card) => (
                  <Col span={8} key={card.title}>
                    <Card>
                      <Statistic
                        title={card.title}
                        value={card.value}
                        prefix={card.prefix}
                        suffix={card.suffix}
                        precision={card.precision ?? 0}
                      />
                    </Card>
                  </Col>
                ))}
              </Row>
              <ReactECharts option={economicOption} style={{ height: 400 }} />
            </>
          )}
        </Spin>
      ),
    },
    {
      key: 'efficiency',
      label: '效率分析',
      children: (
        <Spin spinning={chartLoading}>
          {!hasData ? <Empty description="暂无效率分析数据" /> : (
            <Row gutter={16}>
              <Col span={16}>
                <ReactECharts option={efficiencyLineOption} style={{ height: 400 }} />
              </Col>
              <Col span={8}>
                <Card title="当前平均效率" style={{ textAlign: 'center' }}>
                  <ReactECharts option={efficiencyGaugeOption} style={{ height: 280 }} />
                  <Statistic
                    value={currentEfficiency}
                    precision={2}
                    suffix="%"
                    style={{ marginTop: 8 }}
                  />
                </Card>
              </Col>
            </Row>
          )}
        </Spin>
      ),
    },
    {
      key: 'runtime',
      label: '运行统计',
      children: (
        <Spin spinning={chartLoading}>
          {!hasData ? <Empty description="暂无运行统计数据" /> : (
            <ReactECharts option={runtimeOption} style={{ height: 400 }} />
          )}
        </Spin>
      ),
    },
    {
      key: 'compare',
      label: '对比分析',
      children: (
        <Spin spinning={chartLoading}>
          {!hasData ? <Empty description="暂无对比分析数据" /> : (
            <ReactECharts option={compareOption} style={{ height: 400 }} />
          )}
        </Spin>
      ),
    },
    {
      key: 'timeseries',
      label: '时序分析',
      children: (
        <Spin spinning={chartLoading}>
          {!hasData ? <Empty description="暂无时序分析数据" /> : (
            <ReactECharts option={timeseriesOption} style={{ height: 400 }} />
          )}
        </Spin>
      ),
    },
  ];

  return (
    <div>
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={[16, 16]} align="middle">
          <Col xs={24} sm={12} md={6}>
            <Select
              placeholder="选择电站"
              allowClear
              style={{ width: '100%' }}
              value={stationId}
              onChange={(val) => setStationId(val)}
              options={stations.map((s) => ({ label: s.station_name, value: s.station_id }))}
            />
          </Col>
          <Col xs={24} sm={12} md={6}>
            <RangePicker
              style={{ width: '100%' }}
              value={dateRange}
              onChange={(dates) => setDateRange(dates as [Dayjs, Dayjs] | null)}
              placeholder={['开始日期', '结束日期']}
            />
          </Col>
          <Col xs={24} sm={12} md={4}>
            <Select
              placeholder="机器人类型"
              allowClear
              style={{ width: '100%' }}
              value={robotType}
              onChange={(val) => setRobotType(val)}
              options={Object.entries(ROBOT_TYPE_MAP).map(([k, v]) => ({ label: v, value: Number(k) }))}
            />
          </Col>
          <Col xs={24} sm={12} md={4}>
            <Space>
              <Button type="primary" icon={<ReloadOutlined />} onClick={fetchData} loading={chartLoading}>
                查询
              </Button>
              <Button icon={<DownloadOutlined />} onClick={handleExport} loading={loading}>
                导出
              </Button>
            </Space>
          </Col>
        </Row>
      </Card>

      <Card>
        <Tabs
          activeKey={activeTab}
          onChange={(key) => setActiveTab(key as TabKey)}
          items={tabItems}
        />
      </Card>
    </div>
  );
}
