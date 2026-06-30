import { useState, useEffect, useCallback } from 'react';
import {
  Card, Tabs, Row, Col, Select, DatePicker, Button, Space, Table, Tag, Spin, Empty, message, Progress,
} from 'antd';
import {
  ReloadOutlined, WarningOutlined, BulbOutlined, ThunderboltOutlined,
} from '@ant-design/icons';
import ReactECharts from 'echarts-for-react';
import type { Dayjs } from 'dayjs';
import { analyticsApi } from '../../api/analytics';
import { stationApi } from '../../api/stations';
import type { Station } from '../../types';

const { RangePicker } = DatePicker;

type TabKey = 'efficiency' | 'fault' | 'strategy';

interface EfficiencyPredictionItem {
  date: string;
  actual: number;
  predicted: number;
  lower: number;
  upper: number;
}

interface FaultPredictionItem {
  robot_id: string;
  robot_name?: string;
  fault_type: string;
  probability: number;
  suggested_action: string;
  predicted_time?: string;
}

interface StrategyItem {
  id: string;
  title: string;
  description: string;
  priority: 'high' | 'medium' | 'low';
  expected_improvement: string;
}

export default function PredictionsPage() {
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
    const params = buildQuery();
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
  }, [buildQuery]);

  const fetchStrategy = useCallback(async () => {
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
    if (activeTab === 'efficiency') fetchEfficiency();
    else if (activeTab === 'fault') fetchFaults();
    else if (activeTab === 'strategy') fetchStrategy();
  }, [activeTab, fetchEfficiency, fetchFaults, fetchStrategy]);

  const priorityConfig: Record<string, { color: string; icon: React.ReactNode; label: string }> = {
    high: { color: '#ff4d4f', icon: <WarningOutlined />, label: '高优先级' },
    medium: { color: '#faad14', icon: <ThunderboltOutlined />, label: '中优先级' },
    low: { color: '#52c41a', icon: <BulbOutlined />, label: '低优先级' },
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
              rowKey={(record, index) => record.robot_id || String(index)}
              columns={[
                { title: '机器人ID', dataIndex: 'robot_id', key: 'robot_id', width: 150 },
                { title: '机器人名称', dataIndex: 'robot_name', key: 'robot_name', width: 150, render: (v: string) => v || '-' },
                { title: '故障类型', dataIndex: 'fault_type', key: 'fault_type', width: 120 },
                {
                  title: '发生概率', dataIndex: 'probability', key: 'probability', width: 180,
                  render: (v: number) => (
                    <Progress
                      percent={Math.round((v ?? 0) * 100)}
                      size="small"
                      status={v > 0.7 ? 'exception' : v > 0.4 ? 'active' : 'normal'}
                    />
                  ),
                },
                { title: '建议措施', dataIndex: 'suggested_action', key: 'suggested_action' },
                {
                  title: '预计时间', dataIndex: 'predicted_time', key: 'predicted_time', width: 180,
                  render: (v: string) => v || '-',
                },
              ]}
              pagination={false}
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
            <Empty description="暂无预测数据" />
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
  ];

  return (
    <div>
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={[16, 16]} align="middle">
          <Col xs={24} sm={12} md={8}>
            <Select
              placeholder="选择电站"
              allowClear
              style={{ width: '100%' }}
              value={stationId}
              onChange={(val) => setStationId(val)}
              options={stations.map((s) => ({ label: s.station_name, value: s.station_id }))}
            />
          </Col>
          <Col xs={24} sm={12} md={8}>
            <RangePicker
              style={{ width: '100%' }}
              value={dateRange}
              onChange={(dates) => setDateRange(dates as [Dayjs, Dayjs] | null)}
              placeholder={['开始日期', '结束日期']}
            />
          </Col>
          <Col xs={24} sm={12} md={4}>
            <Space>
              <Button
                type="primary"
                icon={<ReloadOutlined />}
                onClick={() => {
                  if (activeTab === 'efficiency') fetchEfficiency();
                  else if (activeTab === 'fault') fetchFaults();
                  else if (activeTab === 'strategy') fetchStrategy();
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
          onChange={(key) => setActiveTab(key as TabKey)}
          items={tabItems}
        />
      </Card>
    </div>
  );
}
