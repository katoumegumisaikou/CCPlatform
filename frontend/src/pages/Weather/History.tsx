import { useState, useEffect } from 'react';
import {
  Card, Row, Col, Select, DatePicker, Button, Space, Spin, Empty, Table, message,
} from 'antd';
import { SearchOutlined, HistoryOutlined } from '@ant-design/icons';
import ReactECharts from 'echarts-for-react';
import dayjs from 'dayjs';
import { weatherApi } from '../../api/weather';
import { stationApi } from '../../api/stations';
import type { Station, WeatherDaily, WeatherWarning } from '../../types';

const { RangePicker } = DatePicker;

const WARNING_LEVEL_COLOR: Record<string, string> = {
  'red': '#ff0000', 'orange': '#ff7a00', 'yellow': '#fadb14', 'blue': '#1677ff',
};

type DataSource = 'forecast' | 'warnings';

export default function WeatherHistoryPage() {
  const [stations, setStations] = useState<Station[]>([]);
  const [stationId, setStationId] = useState<string | undefined>(undefined);
  const [loading, setLoading] = useState(false);
  const [dataSource, setDataSource] = useState<DataSource>('forecast');
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(null);

  // 预报历史
  const [dailyData, setDailyData] = useState<WeatherDaily[]>([]);
  // 预警历史
  const [warningList, setWarningList] = useState<WeatherWarning[]>([]);

  useEffect(() => {
    stationApi.getAll().then(setStations).catch(() => message.error('获取电站列表失败'));
  }, []);

  const handleSearch = async () => {
    if (!stationId) return;
    setLoading(true);
    try {
      if (dataSource === 'forecast') {
        const data = await weatherApi.getDailyForecast(stationId, 7);
        setDailyData(data || []);
      } else {
        const params: { station_id: string; page: number; size: number; start?: string; end?: string } = {
          station_id: stationId,
          page: 1,
          size: 50,
        };
        if (dateRange && dateRange[0] && dateRange[1]) {
          params.start = dateRange[0].format('YYYY-MM-DD');
          params.end = dateRange[1].format('YYYY-MM-DD');
        }
        const res = await weatherApi.getWarningHistory(params);
        setWarningList(res.list || []);
      }
    } catch { message.error('查询失败'); }
    setLoading(false);
  };

  // 预报时序图
  const dailyChartOption = {
    tooltip: { trigger: 'axis' },
    legend: { data: ['最高温度', '最低温度', '降水量'] },
    grid: { left: 50, right: 50, top: 30, bottom: 40 },
    xAxis: {
      type: 'category',
      data: dailyData.map(d => d.fx_date),
    },
    yAxis: [
      { type: 'value', name: '温度 (°C)' },
      { type: 'value', name: '降水量 (mm)' },
    ],
    series: [
      {
        name: '最高温度', type: 'line',
        data: dailyData.map(d => parseFloat(d.temp_max) || 0),
        itemStyle: { color: '#ff7a45' },
        smooth: true,
      },
      {
        name: '最低温度', type: 'line',
        data: dailyData.map(d => parseFloat(d.temp_min) || 0),
        itemStyle: { color: '#69b1ff' },
        smooth: true,
      },
      {
        name: '降水量', type: 'bar', yAxisIndex: 1,
        data: dailyData.map(d => parseFloat(d.precip) || 0),
        itemStyle: { color: 'rgba(22,119,255,0.3)' },
        barMaxWidth: 12,
      },
    ],
  };

  return (
    <Spin spinning={loading}>
      <Card style={{ marginBottom: 16 }}>
        <Space wrap>
          <Select
            placeholder="选择电站"
            style={{ width: 240 }}
            value={stationId}
            onChange={(v) => setStationId(v)}
            options={stations.map(s => ({
              value: s.station_id,
              label: `${s.station_name} (${s.station_code})`,
            }))}
            allowClear
            showSearch
            optionFilterProp="label"
          />
          <Select
            value={dataSource}
            onChange={(v) => setDataSource(v)}
            style={{ width: 160 }}
            options={[
              { value: 'forecast', label: '逐天预报趋势' },
              { value: 'warnings', label: '历史预警记录' },
            ]}
          />
          {dataSource === 'warnings' && (
            <RangePicker
              onChange={(dates) => setDateRange(dates as [dayjs.Dayjs, dayjs.Dayjs] | null)}
            />
          )}
          <Button
            type="primary"
            icon={<SearchOutlined />}
            onClick={handleSearch}
            disabled={!stationId}
          >
            查询
          </Button>
        </Space>
      </Card>

      {!stationId ? (
        <Empty description="请先选择电站并查询" style={{ marginTop: 80 }}>
          <HistoryOutlined style={{ fontSize: 48, color: '#ccc' }} />
        </Empty>
      ) : (
        <Row gutter={16}>
          <Col span={dataSource === 'forecast' ? 24 : 24}>
            <Card title={dataSource === 'forecast' ? '逐天预报温度趋势' : '历史预警记录'}>
              {dataSource === 'forecast' ? (
                dailyData.length === 0 ? (
                  <Empty description="暂无预报数据" />
                ) : (
                  <ReactECharts
                    style={{ height: 400 }}
                    option={dailyChartOption}
                    opts={{ renderer: 'svg' }}
                  />
                )
              ) : warningList.length === 0 ? (
                <Empty description="暂无预警记录" />
              ) : (
                <Table
                  dataSource={warningList}
                  rowKey="warning_id"
                  size="small"
                  pagination={{ pageSize: 10, showTotal: (t) => `共 ${t} 条` }}
                  columns={[
                    { title: '级别', dataIndex: 'level', width: 80,
                      render: (v: string) => {
                        const colorMap: Record<string, string> = { 'red': '红色', 'orange': '橙色', 'yellow': '黄色', 'blue': '蓝色' };
                        return <span style={{ color: WARNING_LEVEL_COLOR[v] }}>● {colorMap[v] || v}</span>;
                      } },
                    { title: '标题', dataIndex: 'title', width: 200 },
                    { title: '类型', dataIndex: 'type_name', width: 100 },
                    { title: '发布时间', dataIndex: 'pub_time', width: 150 },
                    { title: '开始~结束', width: 260,
                      render: (_: unknown, r: WeatherWarning) => `${r.start_time} ~ ${r.end_time}` },
                    { title: '描述', dataIndex: 'text', ellipsis: true },
                  ]}
                />
              )}
            </Card>
          </Col>
        </Row>
      )}
    </Spin>
  );
}
