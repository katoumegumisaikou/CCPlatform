import { useState, useEffect } from 'react';
import { Card, Row, Col, Statistic, Spin } from 'antd';
import {
  BankOutlined,
  RobotOutlined,
  PlayCircleOutlined,
  WarningOutlined,
  ThunderboltOutlined,
  AlertOutlined,
} from '@ant-design/icons';
import ReactECharts from 'echarts-for-react';
import { monitorApi } from '../../api/monitor';
import type { DashboardOverview } from '../../types';

/** 将 API 返回的 robot_stats key 转为中文标签 */
function statusLabel(key: string): string {
  const map: Record<string, string> = {
    '0': '离线',
    '1': '在线',
    '2': '清扫中',
    '3': '充电中',
    '4': '故障',
    '5': '维护',
    total: '总计',
    online: '在线',
    offline: '离线',
    cleaning: '清扫中',
    charging: '充电中',
    fault: '故障',
    idle: '空闲',
    maintenance: '维护',
  };
  return map[key] ?? key;
}

/** 将 API 返回的 alarm_stats key 转为中文标签 */
function alarmLevelLabel(key: string): string {
  const map: Record<string, string> = {
    '1': '提示',
    '2': '一般',
    '3': '严重',
    '4': '紧急',
    prompt: '提示',
    normal: '一般',
    serious: '严重',
    urgent: '紧急',
  };
  return map[key] ?? key;
}

/** 生成近 N 天日期标签 (MM-DD) */
function buildDateLabels(days: number): string[] {
  const today = new Date();
  return Array.from({ length: days }, (_, i) => {
    const d = new Date(today);
    d.setDate(d.getDate() - (days - 1 - i));
    return d.toISOString().slice(5, 10);
  });
}

const DashboardPage: React.FC = () => {
  const [loading, setLoading] = useState(true);
  const [data, setData] = useState<DashboardOverview | null>(null);

  useEffect(() => {
    fetchDashboard();
  }, []);

  const fetchDashboard = async () => {
    setLoading(true);
    try {
      const res = await monitorApi.getDashboard();
      setData(res);
    } catch {
      // 错误已由 axios 拦截器统一处理
    } finally {
      setLoading(false);
    }
  };

  // 在线机器人数量：支持数字键 "1" 或字符串键 "online"
  const robotOnlineCount =
    data?.robot_stats?.['1'] ?? data?.robot_stats?.online ?? 0;

  // ---- ECharts 配置 ----

  // 饼图：机器人状态分布
  const robotStatusPieOption = {
    title: {
      text: '机器人状态分布',
      left: 'center',
      textStyle: { fontSize: 14 },
    },
    tooltip: { trigger: 'item' as const },
    legend: {
      bottom: 0,
      type: 'scroll' as const,
    },
    series: [
      {
        type: 'pie' as const,
        radius: ['40%', '65%'],
        center: ['50%', '50%'],
        avoidLabelOverlap: true,
        itemStyle: {
          borderRadius: 4,
          borderColor: '#fff',
          borderWidth: 2,
        },
        label: { show: true, formatter: '{b}: {c}' },
        emphasis: {
          label: { fontSize: 16, fontWeight: 'bold' },
        },
        data: data?.robot_stats
          ? Object.entries(data.robot_stats)
              .filter(([key]) => key !== 'total')
              .map(([key, value]) => ({ name: statusLabel(key), value }))
          : [
              { name: '在线', value: 12 },
              { name: '离线', value: 5 },
              { name: '清扫中', value: 8 },
              { name: '充电中', value: 3 },
              { name: '故障', value: 1 },
            ],
      },
    ],
  };

  // 折线图：近7日清扫面积趋势（模拟数据）
  const mockDays = 7;
  const mockDateLabels = buildDateLabels(mockDays);
  const mockTrendValues = mockDateLabels.map(() =>
    Math.round(Math.random() * 500 + 200),
  );

  const cleanTrendOption = {
    title: {
      text: '近7日清扫面积趋势',
      left: 'center',
      textStyle: { fontSize: 14 },
    },
    tooltip: {
      trigger: 'axis' as const,
      formatter: '{b}<br/>清扫面积: {c} m²',
    },
    xAxis: {
      type: 'category' as const,
      data: mockDateLabels,
      axisLabel: { rotate: 30 },
    },
    yAxis: {
      type: 'value' as const,
      name: '平方米',
    },
    grid: { left: '10%', right: '10%', bottom: '15%', top: '20%' },
    series: [
      {
        type: 'line' as const,
        name: '清扫面积',
        data: mockTrendValues,
        smooth: true,
        lineStyle: { width: 3 },
        areaStyle: { opacity: 0.2 },
        symbol: 'circle' as const,
        symbolSize: 6,
      },
    ],
  };

  return (
    <Spin spinning={loading}>
      <div style={{ padding: 24 }}>
        {/* 第一行：统计卡片 */}
        <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="电站总数"
                value={data?.station_count ?? 0}
                prefix={<BankOutlined />}
                valueStyle={{ color: '#1677ff' }}
              />
            </Card>
          </Col>

          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="在线机器人"
                value={robotOnlineCount}
                prefix={<RobotOutlined />}
                valueStyle={{ color: '#52c41a' }}
              />
            </Card>
          </Col>

          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="运行中任务"
                value={data?.running_tasks ?? 0}
                prefix={<PlayCircleOutlined />}
                valueStyle={{ color: '#faad14' }}
              />
            </Card>
          </Col>

          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="未处理告警"
                value={data?.unhandled_alarms ?? 0}
                prefix={<WarningOutlined />}
                valueStyle={{
                  color: data?.unhandled_alarms ? '#ff4d4f' : '#52c41a',
                }}
              />
            </Card>
          </Col>
        </Row>

        {/* 第二行：今日清扫面积 + 告警统计 */}
        <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
          <Col xs={24} lg={12}>
            <Card loading={loading}>
              <Statistic
                title="今日清扫面积"
                value={data?.today_clean_area ?? 0}
                suffix="m²"
                prefix={<ThunderboltOutlined />}
                valueStyle={{ color: '#1677ff', fontSize: 32 }}
              />
            </Card>
          </Col>

          <Col xs={24} lg={12}>
            <Card title="告警统计" loading={loading}>
              {data?.alarm_stats &&
              Object.keys(data.alarm_stats).length > 0 ? (
                Object.entries(data.alarm_stats).map(([level, count]) => (
                  <div
                    key={level}
                    style={{
                      display: 'flex',
                      justifyContent: 'space-between',
                      padding: '10px 0',
                      borderBottom: '1px solid #f0f0f0',
                    }}
                  >
                    <span>
                      <AlertOutlined style={{ marginRight: 8 }} />
                      {alarmLevelLabel(level)}
                    </span>
                    <span style={{ fontWeight: 600 }}>{count as number}</span>
                  </div>
                ))
              ) : (
                <div
                  style={{
                    color: '#999',
                    textAlign: 'center',
                    padding: 32,
                  }}
                >
                  暂无数据
                </div>
              )}
            </Card>
          </Col>
        </Row>

        {/* 第三行：ECharts 图表 */}
        <Row gutter={[16, 16]}>
          <Col xs={24} lg={12}>
            <Card loading={loading}>
              <ReactECharts
                option={robotStatusPieOption}
                style={{ height: 320 }}
              />
            </Card>
          </Col>

          <Col xs={24} lg={12}>
            <Card loading={loading}>
              <ReactECharts
                option={cleanTrendOption}
                style={{ height: 320 }}
              />
            </Card>
          </Col>
        </Row>
      </div>
    </Spin>
  );
};

export default DashboardPage;
