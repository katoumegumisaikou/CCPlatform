import { useState, useEffect } from 'react';
import {
  Card, Row, Col, Select, DatePicker, Button, Space, Table, Tag, Empty, Spin, message,
} from 'antd';
import { SearchOutlined, WarningOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import { weatherApi } from '../../api/weather';
import { stationApi } from '../../api/stations';
import type { Station, WeatherWarning } from '../../types';

const { RangePicker } = DatePicker;

// 预警级别 -> 颜色
const WARNING_LEVEL_COLOR: Record<string, string> = {
  'red': '#ff0000',
  'orange': '#ff7a00',
  'yellow': '#fadb14',
  'blue': '#1677ff',
  'white': '#999',
};

const WARNING_LEVEL_LABEL: Record<string, string> = {
  'red': '红色预警',
  'orange': '橙色预警',
  'yellow': '黄色预警',
  'blue': '蓝色预警',
  'active': '生效中',
};

export default function WarningsPage() {
  const [stations, setStations] = useState<Station[]>([]);
  const [stationId, setStationId] = useState<string | undefined>(undefined);
  const [loading, setLoading] = useState(false);
  const [activeWarnings, setActiveWarnings] = useState<WeatherWarning[]>([]);
  const [historyWarnings, setHistoryWarnings] = useState<WeatherWarning[]>([]);
  const [historyTotal, setHistoryTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(null);

  useEffect(() => {
    stationApi.getAll().then(setStations).catch(() => message.error('获取电站列表失败'));
  }, []);

  useEffect(() => {
    if (stationId) {
      fetchWarnings();
    }
  }, [stationId]);

  const fetchWarnings = async () => {
    if (!stationId) return;
    setLoading(true);
    try {
      const active = await weatherApi.getActiveWarnings(stationId);
      setActiveWarnings(active || []);
    } catch { setActiveWarnings([]); }
    setLoading(false);
  };

  const handleSearchHistory = async () => {
    if (!stationId) return;
    setLoading(true);
    const params: { station_id: string; page: number; size: number; start?: string; end?: string } = {
      station_id: stationId,
      page,
      size: 10,
    };
    if (dateRange && dateRange[0] && dateRange[1]) {
      params.start = dateRange[0].format('YYYY-MM-DD');
      params.end = dateRange[1].format('YYYY-MM-DD');
    }
    try {
      const res = await weatherApi.getWarningHistory(params);
      setHistoryWarnings(res.list || []);
      setHistoryTotal(res.total || 0);
    } catch { message.error('查询历史预警失败'); }
    setLoading(false);
  };

  const warningColumns = [
    { title: '预警级别', dataIndex: 'level', width: 100,
      render: (level: string) => (
        <Tag color={WARNING_LEVEL_COLOR[level] || '#999'}>
          {WARNING_LEVEL_LABEL[level] || level}
        </Tag>
      ) },
    { title: '预警标题', dataIndex: 'title', width: 200 },
    { title: '预警类型', dataIndex: 'type_name', width: 100 },
    { title: '发布单位', dataIndex: 'sender', width: 120 },
    { title: '发布时间', dataIndex: 'pub_time', width: 150 },
    { title: '开始时间', dataIndex: 'start_time', width: 150 },
    { title: '结束时间', dataIndex: 'end_time', width: 150 },
    { title: '状态', dataIndex: 'status', width: 80,
      render: (s: string) => <Tag>{s}</Tag> },
  ];

  return (
    <Spin spinning={loading}>
      <Card style={{ marginBottom: 16 }}>
        <Space wrap>
          <Select
            placeholder="选择电站"
            style={{ width: 260 }}
            value={stationId}
            onChange={(v) => { setStationId(v); setPage(1); }}
            options={stations.map(s => ({
              value: s.station_id,
              label: `${s.station_name} (${s.station_code})`,
            }))}
            allowClear
            showSearch
            optionFilterProp="label"
          />
        </Space>
      </Card>

      {!stationId ? (
        <Empty description="请先选择电站" style={{ marginTop: 80 }} />
      ) : (
        <Row gutter={16}>
          {/* 左侧：当前生效预警 */}
          <Col span={12}>
            <Card
              title={<span><WarningOutlined style={{ color: '#ff4d4f' }} /> 当前预警</span>}
              extra={<Button size="small" onClick={fetchWarnings}>刷新</Button>}
            >
              {activeWarnings.length === 0 ? (
                <Empty description="当前无生效中的预警" image={Empty.PRESENTED_IMAGE_SIMPLE} />
              ) : (
                activeWarnings.map(w => (
                  <Card
                    key={w.warning_id}
                    size="small"
                    style={{
                      marginBottom: 12,
                      borderLeft: `4px solid ${WARNING_LEVEL_COLOR[w.level] || '#999'}`,
                    }}
                  >
                    <div style={{ marginBottom: 4 }}>
                      <Tag color={WARNING_LEVEL_COLOR[w.level] || '#999'}>
                        {WARNING_LEVEL_LABEL[w.level] || w.level}
                      </Tag>
                      <strong>{w.title}</strong>
                      <span style={{ marginLeft: 8, color: '#888', fontSize: 12 }}>{w.type_name}</span>
                    </div>
                    <div style={{ fontSize: 12, color: '#666', marginBottom: 4 }}>
                      {w.pub_time} · {w.sender}
                    </div>
                    <div style={{ fontSize: 12, color: '#666', marginBottom: 6 }}>
                      生效: {w.start_time} ~ {w.end_time}
                    </div>
                    <div style={{ fontSize: 13 }}>{w.text}</div>
                  </Card>
                ))
              )}
            </Card>
          </Col>

          {/* 右侧：历史预警查询 */}
          <Col span={12}>
            <Card
              title="历史预警查询"
              extra={
                <Button size="small" type="primary" icon={<SearchOutlined />} onClick={handleSearchHistory}>
                  查询
                </Button>
              }
            >
              <Space direction="vertical" style={{ width: '100%', marginBottom: 16 }}>
                <RangePicker
                  onChange={(dates) => setDateRange(dates as [dayjs.Dayjs, dayjs.Dayjs] | null)}
                  style={{ width: '100%' }}
                />
              </Space>
              <Table
                dataSource={historyWarnings}
                rowKey="warning_id"
                size="small"
                columns={warningColumns}
                pagination={{
                  current: page,
                  total: historyTotal,
                  pageSize: 10,
                  onChange: (p) => { setPage(p); handleSearchHistory(); },
                  showTotal: (t) => `共 ${t} 条`,
                }}
                scroll={{ x: 800 }}
              />
            </Card>
          </Col>
        </Row>
      )}
    </Spin>
  );
}
