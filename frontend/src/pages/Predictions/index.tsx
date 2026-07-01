import { useState, useEffect, useCallback } from 'react';
import {
  App, Card, Tabs, Row, Col, Select, DatePicker, Button, Space, Table, Tag, Spin, Empty, Progress, Statistic, Flex,
} from 'antd';
import {
  ReloadOutlined, WarningOutlined, BulbOutlined, ThunderboltOutlined, HeartOutlined,
} from '@ant-design/icons';
import ReactECharts from 'echarts-for-react';
import type { Dayjs } from 'dayjs';
import { analyticsApi } from '../../api/analytics';
import { stationApi } from '../../api/stations';
import { robotApi } from '../../api/robots';
import type {
  Station, Robot, HealthOverview, HealthTrend, SparePartForecastItem, MaintenanceWindowRecommendation, PhmComponent,
} from '../../types';

const { RangePicker } = DatePicker;

type TabKey = 'efficiency' | 'fault' | 'strategy' | 'health';

const COMPONENT_LABELS: Record<string, string> = {
  drive_motor: '驱动电机',
  brush_motor: '滚刷电机',
  battery: '电池组',
  controller: '控制器',
  sensor: '传感器',
  transmission: '传动系统',
};

const HEALTH_LEVEL_COLORS: Record<string, string> = {
  green: '#52c41a',
  yellow: '#fadb14',
  orange: '#fa8c16',
  red: '#ff4d4f',
};

const HEALTH_LEVEL_LABELS: Record<string, string> = {
  green: '健康',
  yellow: '亚健康',
  orange: '异常',
  red: '危险',
};

interface EfficiencyPredictionItem {
  date: string;
  actual: number;
  predicted: number;
  lower: number;
  upper: number;
}

interface FaultPredictionItem {
  prediction_id: string;
  robot_id: string;
  robot_name?: string;
  component: string;
  fault_type: string;
  probability: number;
  risk_level: 'high' | 'medium' | 'low';
  confidence: number;
  suggested_action: string;
  recommended_parts?: string;
  predicted_time?: string;
  predicted_end?: string;
  status: number;
  maint_id?: string;
}

interface StrategyItem {
  id: string;
  title: string;
  description: string;
  priority: 'high' | 'medium' | 'low';
  expected_improvement: string;
}

export default function PredictionsPage() {
  const { message } = App.useApp();
  const [activeTab, setActiveTab] = useState<TabKey>('efficiency');
  const [stations, setStations] = useState<Station[]>([]);
  const [stationId, setStationId] = useState<string | undefined>(undefined);
  const [dateRange, setDateRange] = useState<[Dayjs, Dayjs] | null>(null);
  const [loading, setLoading] = useState(false);

  // 效率预测
  const [efficiencyOption, setEfficiencyOption] = useState<object>({});
  const [hasEfficiencyData, setHasEfficiencyData] = useState(true);

  // 故障预测
  const [faultList, setFaultList] = useState<FaultPredictionItem[]>([]);
  const [faultLoading, setFaultLoading] = useState(false);

  // 策略优化
  const [strategyList, setStrategyList] = useState<StrategyItem[]>([]);
  const [strategyLoading, setStrategyLoading] = useState(false);

  // 设备健康
  const [robots, setRobots] = useState<Robot[]>([]);
  const [robotId, setRobotId] = useState<string | undefined>(undefined);
  const [healthComponent, setHealthComponent] = useState<PhmComponent>('drive_motor');
  const [healthOverview, setHealthOverview] = useState<HealthOverview | null>(null);
  const [healthTrend, setHealthTrend] = useState<HealthTrend | null>(null);
  const [healthTrendOption, setHealthTrendOption] = useState<object>({});
  const [spareParts, setSpareParts] = useState<SparePartForecastItem[]>([]);
  const [maintWindow, setMaintWindow] = useState<MaintenanceWindowRecommendation | null>(null);
  const [healthLoading, setHealthLoading] = useState(false);

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
    if (dateRange && dateRange[0] && dateRange[1]) {
      params.start_date = dateRange[0].format('YYYY-MM-DD');
      params.end_date = dateRange[1].format('YYYY-MM-DD');
    }
    return params;
  }, [stationId, dateRange]);

  const fetchEfficiency = useCallback(async () => {
    if (!stationId) return;
    const params = buildQuery();
    setLoading(true);
    try {
      const res = await analyticsApi.predictions.efficiency(params);
      const list: EfficiencyPredictionItem[] = Array.isArray(res) ? res : ((res as Record<string, unknown>)?.list as EfficiencyPredictionItem[] || []);
      if (list.length === 0) {
        setHasEfficiencyData(false);
        return;
      }
      setHasEfficiencyData(true);
      const dates = list.map((it) => it.date);
      setEfficiencyOption({
        tooltip: { trigger: 'axis' },
        legend: { data: ['实际效率', '预测效率', '置信区间上限', '置信区间下限'] },
        xAxis: { type: 'category', data: dates, axisLabel: { rotate: 30 } },
        yAxis: { type: 'value', name: '效率（%）', max: 100 },
        series: [
          {
            name: '实际效率', type: 'line',
            data: list.map((it) => it.actual),
            itemStyle: { color: '#1890ff' },
            lineStyle: { width: 2 },
          },
          {
            name: '预测效率', type: 'line',
            data: list.map((it) => it.predicted),
            itemStyle: { color: '#faad14' },
            lineStyle: { type: 'dashed', width: 2 },
          },
          {
            name: '置信区间上限', type: 'line',
            data: list.map((it) => it.upper),
            lineStyle: { opacity: 0 },
            itemStyle: { opacity: 0 },
            areaStyle: { color: 'rgba(250,173,20,0.1)' },
            showSymbol: false,
          },
          {
            name: '置信区间下限', type: 'line',
            data: list.map((it) => it.lower),
            lineStyle: { opacity: 0 },
            itemStyle: { opacity: 0 },
            areaStyle: { color: 'rgba(250,173,20,0.15)' },
            showSymbol: false,
          },
        ],
        grid: { left: '3%', right: '4%', bottom: '15%', containLabel: true },
      });
    } catch {
      message.error('获取效率预测失败');
      setHasEfficiencyData(false);
    } finally {
      setLoading(false);
    }
  }, [buildQuery]);

  const fetchFaults = useCallback(async () => {
    const params: Record<string, unknown> = {};
    if (robotId) params.robot_id = robotId;
    else if (stationId) params.station_id = stationId;
    setFaultLoading(true);
    try {
      const res = await analyticsApi.predictions.fault(params);
      const list: FaultPredictionItem[] = Array.isArray(res) ? res : ((res as Record<string, unknown>)?.list as FaultPredictionItem[] || []);
      setFaultList(list);
    } catch {
      message.error('获取故障预测失败');
    } finally {
      setFaultLoading(false);
    }
  }, [stationId, robotId]);

  const handleGenerateWorkorder = useCallback(async (predictionId: string) => {
    try {
      await analyticsApi.predictions.generateWorkorder(predictionId);
      message.success('维护工单已生成');
      fetchFaults();
    } catch {
      message.error('生成维护工单失败');
    }
  }, [fetchFaults]);

  const fetchStrategy = useCallback(async () => {
    if (!stationId) {
      setStrategyList([]);
      return;
    }
    const params = buildQuery();
    setStrategyLoading(true);
    try {
      const res = await analyticsApi.predictions.strategy(params);
      const list: StrategyItem[] = Array.isArray(res) ? res : ((res as Record<string, unknown>)?.list as StrategyItem[] || []);
      setStrategyList(list);
    } catch {
      message.error('获取策略优化失败');
    } finally {
      setStrategyLoading(false);
    }
  }, [buildQuery]);

  useEffect(() => {
    const params: Record<string, unknown> = { size: 999 };
    if (stationId) params.station_id = stationId;
    robotApi.list(params).then((res) => {
      const list = res?.list || [];
      setRobots(list);
      if (activeTab === 'health' && list.length > 0 && !list.find((r) => r.robot_id === robotId)) {
        setRobotId(list[0].robot_id);
      } else if (list.length === 0 || (robotId && !list.find((r) => r.robot_id === robotId))) {
        setRobotId(undefined);
      }
    }).catch(() => {
      message.error('获取机器人列表失败');
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [stationId]);

  const fetchHealth = useCallback(async () => {
    if (!robotId) {
      setHealthOverview(null);
      return;
    }
    setHealthLoading(true);
    try {
      const [overview, trend, parts, window_] = await Promise.all([
        analyticsApi.health.overview({ robot_id: robotId }) as Promise<HealthOverview>,
        analyticsApi.health.trend({ robot_id: robotId, component: healthComponent, days: 30 }) as Promise<HealthTrend>,
        analyticsApi.predictions.spareParts(stationId ? { station_id: stationId } : {}) as Promise<SparePartForecastItem[]>,
        analyticsApi.predictions.maintenanceWindow({ robot_id: robotId }) as Promise<MaintenanceWindowRecommendation>,
      ]);
      setHealthOverview(overview);
      setHealthTrend(trend);
      setSpareParts(Array.isArray(parts) ? parts : []);
      setMaintWindow(window_);

      const points = trend?.points || [];
      setHealthTrendOption({
        tooltip: { trigger: 'axis' },
        xAxis: {
          type: 'category',
          data: points.map((p) => p.record_time),
          axisLabel: { rotate: 30, formatter: (v: string) => v.slice(5, 16) },
        },
        yAxis: { type: 'value', name: '健康指数', min: 0, max: 100 },
        series: [
          {
            name: '健康指数', type: 'line',
            data: points.map((p) => p.health_score),
            itemStyle: { color: '#1890ff' },
            areaStyle: { color: 'rgba(24,144,255,0.1)' },
            smooth: true,
          },
        ],
        grid: { left: '3%', right: '4%', bottom: '15%', containLabel: true },
      });
    } catch {
      message.error('获取设备健康数据失败');
    } finally {
      setHealthLoading(false);
    }
  }, [robotId, healthComponent, stationId]);

  useEffect(() => {
    if (activeTab === 'efficiency') fetchEfficiency();
    else if (activeTab === 'fault') fetchFaults();
    else if (activeTab === 'strategy') fetchStrategy();
    else if (activeTab === 'health') fetchHealth();
  }, [activeTab, fetchEfficiency, fetchFaults, fetchStrategy, fetchHealth]);

  const priorityConfig: Record<string, { color: string; icon: React.ReactNode; label: string }> = {
    high: { color: '#ff4d4f', icon: <WarningOutlined />, label: '高优先级' },
    medium: { color: '#faad14', icon: <ThunderboltOutlined />, label: '中优先级' },
    low: { color: '#52c41a', icon: <BulbOutlined />, label: '低优先级' },
  };

  const riskConfig: Record<string, { color: string; label: string }> = {
    high: { color: 'red', label: '高风险' },
    medium: { color: 'orange', label: '中风险' },
    low: { color: 'green', label: '低风险' },
  };

  const tabItems = [
    {
      key: 'efficiency',
      label: '效率预测',
      children: (
        <Spin spinning={loading}>
          {!hasEfficiencyData ? (
            <Empty description="暂无预测数据" />
          ) : (
            <ReactECharts option={efficiencyOption} style={{ height: 400 }} />
          )}
        </Spin>
      ),
    },
    {
      key: 'fault',
      label: '故障预测',
      children: (
        <Spin spinning={faultLoading}>
          {faultList.length === 0 ? (
            <Empty description="暂无预测数据" />
          ) : (
            <Table
              dataSource={faultList}
              rowKey={(record) => record.prediction_id || `${record.robot_id}-${record.component}-${record.fault_type}-${record.predicted_time}`}
              columns={[
                { title: '机器人ID', dataIndex: 'robot_id', key: 'robot_id', width: 150 },
                { title: '机器人名称', dataIndex: 'robot_name', key: 'robot_name', width: 150, render: (v: string) => v || '-' },
                {
                  title: '部件',
                  dataIndex: 'component',
                  key: 'component',
                  width: 110,
                  render: (v: string) => COMPONENT_LABELS[v] || v,
                },
                { title: '故障类型', dataIndex: 'fault_type', key: 'fault_type', width: 120 },
                {
                  title: '发生概率', dataIndex: 'probability', key: 'probability', width: 180,
                  render: (v: number) => {
                    const percent = Math.round((v ?? 0) * 100);
                    return (
                      <Space size={8} style={{ width: '100%' }}>
                        <Progress
                          percent={percent}
                          size="small"
                          status={v > 0.7 ? 'exception' : v > 0.4 ? 'active' : 'normal'}
                          showInfo={false}
                          style={{ minWidth: 92, flex: 1 }}
                        />
                        <span style={{ width: 42, textAlign: 'right' }}>{percent}%</span>
                      </Space>
                    );
                  },
                },
                {
                  title: '风险等级',
                  dataIndex: 'risk_level',
                  key: 'risk_level',
                  width: 100,
                  render: (v: string) => {
                    const cfg = riskConfig[v] || riskConfig.low;
                    return <Tag color={cfg.color}>{cfg.label}</Tag>;
                  },
                },
                {
                  title: '置信度',
                  dataIndex: 'confidence',
                  key: 'confidence',
                  width: 90,
                  render: (v: number) => `${Math.round((v ?? 0) * 100)}%`,
                },
                { title: '建议措施', dataIndex: 'suggested_action', key: 'suggested_action' },
                {
                  title: '推荐备件',
                  dataIndex: 'recommended_parts',
                  key: 'recommended_parts',
                  width: 140,
                  render: (v: string) => v || '-',
                },
                {
                  title: '预测窗口',
                  key: 'predicted_window',
                  width: 220,
                  render: (_, record) => (
                    <span>{record.predicted_time || '-'} ~ {record.predicted_end || '-'}</span>
                  ),
                },
                {
                  title: '工单',
                  key: 'workorder',
                  width: 120,
                  render: (_, record) => (
                    record.maint_id ? (
                      <Tag color="blue">已生成</Tag>
                    ) : (
                      <Button
                        size="small"
                        type="link"
                        disabled={!record.prediction_id}
                        onClick={() => handleGenerateWorkorder(record.prediction_id)}
                      >
                        生成工单
                      </Button>
                    )
                  ),
                },
              ]}
              scroll={{ x: 1600 }}
              pagination={{ pageSize: 20, showSizeChanger: true }}
            />
          )}
        </Spin>
      ),
    },
    {
      key: 'strategy',
      label: '策略优化',
      children: (
        <Spin spinning={strategyLoading}>
          {strategyList.length === 0 ? (
            <Empty description={stationId ? '暂无预测数据' : '请选择电站'} />
          ) : (
            <Row gutter={[16, 16]}>
              {strategyList.map((item) => {
                const config = priorityConfig[item.priority] || priorityConfig.medium;
                return (
                  <Col xs={24} sm={12} md={8} key={item.id}>
                    <Card
                      hoverable
                      title={
                        <Space>
                          <span style={{ color: config.color }}>{config.icon}</span>
                          <span>{item.title}</span>
                        </Space>
                      }
                      extra={<Tag color={config.color}>{config.label}</Tag>}
                    >
                      <p>{item.description}</p>
                      <p style={{ color: '#52c41a', fontWeight: 500 }}>
                        预期提升：{item.expected_improvement}
                      </p>
                    </Card>
                  </Col>
                );
              })}
            </Row>
          )}
        </Spin>
      ),
    },
    {
      key: 'health',
      label: '设备健康',
      children: (
        <Spin spinning={healthLoading}>
          {!robotId || !healthOverview ? (
            <Empty description="请选择机器人" />
          ) : (
            <Row gutter={[16, 16]}>
              <Col span={24}>
                <Card size="small">
                  <Space size="large" align="center">
                    <Statistic
                      title="综合健康指数"
                      value={healthOverview.overall_score}
                      precision={1}
                      styles={{ content: { color: HEALTH_LEVEL_COLORS[healthOverview.overall_level] } }}
                      prefix={<HeartOutlined />}
                    />
                    <Tag color={HEALTH_LEVEL_COLORS[healthOverview.overall_level]}>
                      {HEALTH_LEVEL_LABELS[healthOverview.overall_level] || healthOverview.overall_level}
                    </Tag>
                  </Space>
                </Card>
              </Col>
              <Col span={24}>
                <Row gutter={[16, 16]}>
                  {healthOverview.components.map((comp) => (
                    <Col xs={24} sm={12} md={8} lg={4} key={comp.component}>
                      <Card
                        size="small"
                        hoverable
                        style={{
                          cursor: 'pointer',
                          borderColor: healthComponent === comp.component ? HEALTH_LEVEL_COLORS[comp.health_level] : undefined,
                        }}
                        onClick={() => setHealthComponent(comp.component)}
                      >
                        <div style={{ fontWeight: 500, marginBottom: 4 }}>
                          {COMPONENT_LABELS[comp.component] || comp.component}
                        </div>
                        <Progress
                          type="dashboard"
                          percent={Math.round(comp.health_score)}
                          size={80}
                          strokeColor={HEALTH_LEVEL_COLORS[comp.health_level]}
                        />
                        <div style={{ color: HEALTH_LEVEL_COLORS[comp.health_level] }}>
                          {HEALTH_LEVEL_LABELS[comp.health_level] || comp.health_level}
                        </div>
                      </Card>
                    </Col>
                  ))}
                </Row>
              </Col>
              <Col xs={24} lg={14}>
                <Card size="small" title={`${COMPONENT_LABELS[healthComponent] || healthComponent} 健康趋势（近30天）`}>
                  {!healthTrend || healthTrend.points.length === 0 ? (
                    <Empty description="暂无历史数据" />
                  ) : (
                    <>
                      <ReactECharts option={healthTrendOption} style={{ height: 280 }} />
                      {healthTrend.change_rate !== undefined && (
                        <div style={{ marginTop: 8, color: healthTrend.change_rate < 0 ? '#ff4d4f' : '#52c41a' }}>
                          环比变化：{healthTrend.change_rate}%
                        </div>
                      )}
                      {healthTrend.trend_note && <div style={{ color: '#999' }}>{healthTrend.trend_note}</div>}
                    </>
                  )}
                </Card>
              </Col>
              <Col xs={24} lg={10}>
                <Card size="small" title="维护建议" style={{ marginBottom: 16 }}>
                  {maintWindow ? (
                    <>
                      <p>推荐窗口：{maintWindow.recommended_start} ~ {maintWindow.recommended_end}</p>
                      <p style={{ color: '#999' }}>{maintWindow.reason}</p>
                    </>
                  ) : (
                    <Empty description="暂无维护建议" />
                  )}
                </Card>
                <Card size="small" title="备件需求预测">
                  {spareParts.length === 0 ? (
                    <Empty description="暂无备件需求" />
                  ) : (
                    <Flex vertical gap={8}>
                      {spareParts.map((item) => (
                        <Flex
                          key={item.part_name}
                          align="center"
                          justify="space-between"
                          gap={12}
                        >
                          <span>{item.part_name}</span>
                          <Tag color="blue" style={{ marginInlineEnd: 0 }}>
                            预计需求 {item.expected_qty}
                          </Tag>
                        </Flex>
                      ))}
                    </Flex>
                  )}
                </Card>
              </Col>
            </Row>
          )}
        </Spin>
      ),
    },
  ];

  return (
    <div>
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={[16, 16]} align="middle">
          <Col xs={24} sm={12} md={8}>
            <Select
              placeholder={activeTab === 'fault' ? '选择电站（可选）' : '选择电站'}
              allowClear
              style={{ width: '100%' }}
              value={stationId}
              onChange={(val) => setStationId(val)}
              options={stations.map((s) => ({ label: s.station_name, value: s.station_id }))}
            />
          </Col>
          {activeTab === 'health' || activeTab === 'fault' ? (
            <Col xs={24} sm={12} md={8}>
              <Select
                placeholder={activeTab === 'fault' ? '选择机器人（可选）' : '选择机器人'}
                allowClear={activeTab === 'fault'}
                style={{ width: '100%' }}
                value={robotId}
                onChange={(val) => setRobotId(val)}
                options={robots.map((r) => ({ label: r.robot_name, value: r.robot_id }))}
              />
            </Col>
          ) : (
            <Col xs={24} sm={12} md={8}>
              <RangePicker
                style={{ width: '100%' }}
                value={dateRange}
                onChange={(dates) => setDateRange(dates as [Dayjs, Dayjs] | null)}
                placeholder={['开始日期', '结束日期']}
              />
            </Col>
          )}
          <Col xs={24} sm={12} md={4}>
            <Space>
              <Button
                type="primary"
                icon={<ReloadOutlined />}
                onClick={() => {
                  if (activeTab === 'efficiency') fetchEfficiency();
                  else if (activeTab === 'fault') fetchFaults();
                  else if (activeTab === 'strategy') fetchStrategy();
                  else if (activeTab === 'health') fetchHealth();
                }}
              >
                查询
              </Button>
            </Space>
          </Col>
        </Row>
      </Card>

      <Card>
        <Tabs
          activeKey={activeTab}
          onChange={(key) => {
            const nextTab = key as TabKey;
            setActiveTab(nextTab);
            if (nextTab === 'health' && !robotId && robots.length > 0) {
              setRobotId(robots[0].robot_id);
            }
          }}
          items={tabItems}
        />
      </Card>
    </div>
  );
}
