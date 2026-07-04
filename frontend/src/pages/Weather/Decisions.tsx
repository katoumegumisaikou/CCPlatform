import { useState, useEffect } from 'react';
import {
  Card, Tabs, Row, Col, Select, Button, Space, Table, Tag, Descriptions,
  Progress, Spin, Empty, Modal, Form, Input, InputNumber, message,
} from 'antd';
import { PlusOutlined, ThunderboltOutlined } from '@ant-design/icons';
import { cleaningDecisionApi } from '../../api/cleaningDecision';
import { stationApi } from '../../api/stations';
import type { Station, CleaningDecisionRule, ExtremeWeatherRule, CleaningAdvice, CleaningScheduleAdjustment } from '../../types';

type TabKey = 'advice' | 'rules' | 'extreme' | 'adjustments';

const ACTION_TAG_COLOR: Record<string, string> = {
  '建议清扫': 'green',
  '暂缓清扫': 'orange',
  '暂停清扫': 'red',
  '暂无数据': 'default',
};

export default function DecisionsPage() {
  const [activeTab, setActiveTab] = useState<TabKey>('advice');
  const [stations, setStations] = useState<Station[]>([]);
  const [stationId, setStationId] = useState<string | undefined>(undefined);
  const [loading, setLoading] = useState(false);

  // 建议
  const [advice, setAdvice] = useState<CleaningAdvice | null>(null);

  // 规则
  const [rules, setRules] = useState<CleaningDecisionRule[]>([]);
  const [rulesTotal, setRulesTotal] = useState(0);
  const [rulesPage, setRulesPage] = useState(1);
  const [ruleModalOpen, setRuleModalOpen] = useState(false);
  const [editingRule, setEditingRule] = useState<CleaningDecisionRule | null>(null);
  const [ruleForm] = Form.useForm();

  // 极端天气规则
  const [extremeRules, setExtremeRules] = useState<ExtremeWeatherRule[]>([]);
  const [extremeTotal, setExtremeTotal] = useState(0);
  const [extremePage, setExtremePage] = useState(1);
  const [extremeModalOpen, setExtremeModalOpen] = useState(false);
  const [editingExtreme, setEditingExtreme] = useState<ExtremeWeatherRule | null>(null);
  const [extremeForm] = Form.useForm();

  // 频次调整
  const [adjustments, setAdjustments] = useState<CleaningScheduleAdjustment[]>([]);
  const [adjTotal, setAdjTotal] = useState(0);
  const [adjPage, setAdjPage] = useState(1);

  useEffect(() => {
    stationApi.getAll().then(setStations).catch(() => message.error('获取电站列表失败'));
  }, []);

  useEffect(() => {
    if (!stationId) return;
    if (activeTab === 'advice') fetchAdvice();
    else if (activeTab === 'rules') fetchRules();
    else if (activeTab === 'extreme') fetchExtremeRules();
    else if (activeTab === 'adjustments') fetchAdjustments();
  }, [stationId, activeTab]);

  const fetchAdvice = async () => {
    if (!stationId) return;
    setLoading(true);
    try {
      const result = await cleaningDecisionApi.getAdvice(stationId);
      setAdvice(result);
    } catch { setAdvice(null); }
    setLoading(false);
  };

  const fetchRules = async (page = 1) => {
    if (!stationId) return;
    setLoading(true);
    try {
      const res = await cleaningDecisionApi.getRules({ station_id: stationId, page, size: 10 });
      setRules(res.list || []);
      setRulesTotal(res.total || 0);
      setRulesPage(page);
    } catch { message.error('获取规则失败'); }
    setLoading(false);
  };

  const fetchExtremeRules = async (page = 1) => {
    if (!stationId) return;
    setLoading(true);
    try {
      const res = await cleaningDecisionApi.getExtremeRules({ station_id: stationId, page, size: 10 });
      setExtremeRules(res.list || []);
      setExtremeTotal(res.total || 0);
      setExtremePage(page);
    } catch { message.error('获取极端天气规则失败'); }
    setLoading(false);
  };

  const fetchAdjustments = async (page = 1) => {
    if (!stationId) return;
    setLoading(true);
    try {
      const res = await cleaningDecisionApi.getAdjustments({ station_id: stationId, page, size: 10 });
      setAdjustments(res.list || []);
      setAdjTotal(res.total || 0);
      setAdjPage(page);
    } catch { setAdjustments([]); }
    setLoading(false);
  };

  const handleSaveRule = async () => {
    const values = await ruleForm.validateFields();
    const data = {
      rule_name: values.rule_name,
      station_id: values.global ? '' : stationId!,
      conditions: JSON.stringify({
        max_wind_scale: values.max_wind_scale,
        min_temp: values.min_temp,
        max_temp: values.max_temp,
        max_humidity: values.max_humidity,
        no_rain: values.no_rain,
        min_vis_km: values.min_vis_km,
      }),
      dust_rule: JSON.stringify({
        days_since_last_clean: values.dust_days || 3,
        dust_threshold: values.dust_threshold || 500,
      }),
      priority: values.priority || 1,
      status: 1,
      remark: values.remark || '',
    };
    try {
      if (editingRule) {
        await cleaningDecisionApi.updateRule(editingRule.rule_id, data);
      } else {
        await cleaningDecisionApi.createRule(data);
      }
      message.success(editingRule ? '规则已更新' : '规则已创建');
      setRuleModalOpen(false);
      fetchRules(rulesPage);
    } catch { message.error('保存失败'); }
  };

  const handleSaveExtremeRule = async () => {
    const values = await extremeForm.validateFields();
    const data = {
      rule_name: values.rule_name,
      station_id: values.global ? '' : stationId!,
      weather_type: values.weather_type,
      trigger_cond: JSON.stringify({
        wind_scale_gt: values.wind_scale_gt,
        precip_gt: values.precip_gt,
        warning_level: values.warning_level,
      }),
      action: values.action,
      recovery_cond: JSON.stringify({ wind_scale_lt: 4, precip_lt: 1 }),
      status: 1,
      remark: values.remark || '',
    };
    try {
      if (editingExtreme) {
        await cleaningDecisionApi.updateExtremeRule(editingExtreme.rule_id, data);
      } else {
        await cleaningDecisionApi.createExtremeRule(data);
      }
      message.success(editingExtreme ? '规则已更新' : '规则已创建');
      setExtremeModalOpen(false);
      fetchExtremeRules(extremePage);
    } catch { message.error('保存失败'); }
  };

  const handleGenAdjustment = async () => {
    if (!stationId) return;
    try {
      await cleaningDecisionApi.generateAdjustment(stationId);
      message.success('频次调整建议已生成');
      fetchAdjustments(1);
    } catch { message.error('生成失败'); }
  };

  // ---- Tab 1: 清扫建议 ----
  const renderAdviceTab = () => (
    <Spin spinning={loading}>
      {!advice ? (
        <Empty description="点击获取建议" />
      ) : (
        <Row gutter={16}>
          <Col span={16}>
            <Card title="决策结果" size="small">
              <Descriptions column={1} size="small">
                <Descriptions.Item label="建议动作">
                  <Tag color={ACTION_TAG_COLOR[advice.suggested_action] || 'default'} style={{ fontSize: 16, padding: '4px 12px' }}>
                    {advice.suggested_action}
                  </Tag>
                </Descriptions.Item>
                <Descriptions.Item label="决策理由">{advice.reason}</Descriptions.Item>
                <Descriptions.Item label="推荐时间">{advice.recommended_time}</Descriptions.Item>
              </Descriptions>
            </Card>
          </Col>
          <Col span={8}>
            <Card title="评分" size="small">
              <div style={{ marginBottom: 16 }}>
                <div style={{ marginBottom: 4 }}>灰尘累积评分</div>
                <Progress percent={Math.round(advice.dust_score * 100)} size="small"
                  status={advice.dust_score > 0.7 ? 'exception' : advice.dust_score > 0.4 ? 'normal' : 'success'} />
              </div>
              <div>
                <div style={{ marginBottom: 4 }}>气象适宜度</div>
                <Progress percent={Math.round(advice.weather_score * 100)} size="small"
                  status={advice.weather_score > 0.7 ? 'success' : advice.weather_score > 0.4 ? 'normal' : 'exception'} />
              </div>
            </Card>
          </Col>
          {advice.triggered_warnings.length > 0 && (
            <Col span={24}>
              <Card title="⚠️ 当前天气预警" size="small" style={{ marginTop: 16, borderLeft: '4px solid #ff4d4f' }}>
                {advice.triggered_warnings.map(w => (
                  <Tag color="red" key={w.warning_id} style={{ marginBottom: 4 }}>
                    [{w.level}] {w.title}
                  </Tag>
                ))}
              </Card>
            </Col>
          )}
        </Row>
      )}
    </Spin>
  );

  // ---- Tab 2: 决策规则 ----
  const renderRulesTab = () => (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditingRule(null); ruleForm.resetFields(); setRuleModalOpen(true); }}>
          新建规则
        </Button>
      </Space>
      <Table
        dataSource={rules}
        rowKey="rule_id"
        size="small"
        pagination={{ current: rulesPage, total: rulesTotal, pageSize: 10, onChange: fetchRules }}
        columns={[
          { title: '规则名称', dataIndex: 'rule_name', width: 120 },
          { title: '电站', dataIndex: 'station_id', width: 80, render: (v: string) => v || '全局' },
          { title: '条件', dataIndex: 'conditions', width: 200, ellipsis: true },
          { title: '灰尘规则', dataIndex: 'dust_rule', width: 150, ellipsis: true },
          { title: '优先级', dataIndex: 'priority', width: 60 },
          { title: '状态', dataIndex: 'status', width: 60,
            render: (v: number) => <Tag color={v === 1 ? 'green' : 'red'}>{v === 1 ? '启用' : '停用'}</Tag> },
          { title: '操作', width: 100,
            render: (_: unknown, r: CleaningDecisionRule) => (
              <Space size="small">
                <a onClick={() => { setEditingRule(r); ruleForm.setFieldsValue({
                  rule_name: r.rule_name,
                  max_wind_scale: JSON.parse(r.conditions || '{}').max_wind_scale,
                  min_temp: JSON.parse(r.conditions || '{}').min_temp,
                  max_temp: JSON.parse(r.conditions || '{}').max_temp,
                  max_humidity: JSON.parse(r.conditions || '{}').max_humidity,
                  priority: r.priority,
                  remark: r.remark,
                  dust_days: JSON.parse(r.dust_rule || '{}').days_since_last_clean,
                  dust_threshold: JSON.parse(r.dust_rule || '{}').dust_threshold,
                }); setRuleModalOpen(true); }}>
                  编辑
                </a>
                <a onClick={() => { cleaningDecisionApi.deleteRule(r.rule_id).then(() => fetchRules(rulesPage)); }} style={{ color: '#ff4d4f' }}>
                  删除
                </a>
              </Space>
            ),
          },
        ]}
      />

      {/* 规则编辑弹窗 */}
      <Modal title={editingRule ? '编辑决策规则' : '新建决策规则'} open={ruleModalOpen}
        onOk={handleSaveRule} onCancel={() => setRuleModalOpen(false)} width={560}>
        <Form form={ruleForm} layout="vertical" size="small">
          <Form.Item name="rule_name" label="规则名称" rules={[{ required: true }]}>
            <Input placeholder="如：标准清扫条件" />
          </Form.Item>
          <Form.Item name="global" label="全局规则" valuePropName="checked">
            <span style={{ fontSize: 12, color: '#888' }}>勾选后适用于所有电站</span>
          </Form.Item>
          <Row gutter={16}>
            <Col span={8}><Form.Item name="max_wind_scale" label="最大风力(级)"><InputNumber min={1} max={12} style={{ width: '100%' }} /></Form.Item></Col>
            <Col span={8}><Form.Item name="min_temp" label="最低温度(°C)"><InputNumber min={-50} max={50} style={{ width: '100%' }} /></Form.Item></Col>
            <Col span={8}><Form.Item name="max_temp" label="最高温度(°C)"><InputNumber min={-50} max={60} style={{ width: '100%' }} /></Form.Item></Col>
          </Row>
          <Row gutter={16}>
            <Col span={8}><Form.Item name="max_humidity" label="最大湿度(%)"><InputNumber min={0} max={100} style={{ width: '100%' }} /></Form.Item></Col>
            <Col span={8}><Form.Item name="min_vis_km" label="最低能见度(km)"><InputNumber min={0} max={50} style={{ width: '100%' }} /></Form.Item></Col>
            <Col span={8}><Form.Item name="no_rain" label="要求无降雨" valuePropName="checked"><span style={{ fontSize: 12 }}>勾选</span></Form.Item></Col>
          </Row>
          <Row gutter={16}>
            <Col span={8}><Form.Item name="dust_days" label="灰尘累积天数"><InputNumber min={1} max={30} style={{ width: '100%' }} /></Form.Item></Col>
            <Col span={8}><Form.Item name="dust_threshold" label="灰尘阈值"><InputNumber min={100} max={2000} style={{ width: '100%' }} /></Form.Item></Col>
            <Col span={8}><Form.Item name="priority" label="优先级"><InputNumber min={1} max={5} style={{ width: '100%' }} /></Form.Item></Col>
          </Row>
          <Form.Item name="remark" label="备注"><Input.TextArea rows={2} /></Form.Item>
        </Form>
      </Modal>
    </div>
  );

  // ---- Tab 3: 极端天气规则 ----
  const renderExtremeTab = () => (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button type="primary" icon={<PlusOutlined />}
          onClick={() => { setEditingExtreme(null); extremeForm.resetFields(); setExtremeModalOpen(true); }}>
          新建规则
        </Button>
      </Space>
      <Table
        dataSource={extremeRules}
        rowKey="rule_id"
        size="small"
        pagination={{ current: extremePage, total: extremeTotal, pageSize: 10, onChange: fetchExtremeRules }}
        columns={[
          { title: '规则名称', dataIndex: 'rule_name', width: 120 },
          { title: '电站', dataIndex: 'station_id', width: 80, render: (v: string) => v || '全局' },
          { title: '天气类型', dataIndex: 'weather_type', width: 80 },
          { title: '触发条件', dataIndex: 'trigger_cond', width: 150, ellipsis: true },
          { title: '动作', dataIndex: 'action', width: 100,
            render: (v: string) => <Tag color={v === '暂停清扫' ? 'red' : v === '返回充电' ? 'orange' : 'blue'}>{v}</Tag> },
          { title: '状态', dataIndex: 'status', width: 60,
            render: (v: number) => <Tag color={v === 1 ? 'green' : 'red'}>{v === 1 ? '启用' : '停用'}</Tag> },
          { title: '操作', width: 100,
            render: (_: unknown, r: ExtremeWeatherRule) => (
              <Space size="small">
                <a onClick={() => { setEditingExtreme(r); extremeForm.setFieldsValue({
                  rule_name: r.rule_name,
                  weather_type: r.weather_type,
                  wind_scale_gt: JSON.parse(r.trigger_cond || '{}').wind_scale_gt,
                  precip_gt: JSON.parse(r.trigger_cond || '{}').precip_gt,
                  action: r.action,
                  remark: r.remark,
                }); setExtremeModalOpen(true); }}>
                  编辑
                </a>
                <a onClick={() => { cleaningDecisionApi.deleteExtremeRule(r.rule_id).then(() => fetchExtremeRules(extremePage)); }} style={{ color: '#ff4d4f' }}>
                  删除
                </a>
              </Space>
            ),
          },
        ]}
      />

      <Modal title={editingExtreme ? '编辑极端天气规则' : '新建极端天气规则'} open={extremeModalOpen}
        onOk={handleSaveExtremeRule} onCancel={() => setExtremeModalOpen(false)} width={480}>
        <Form form={extremeForm} layout="vertical" size="small">
          <Form.Item name="rule_name" label="规则名称" rules={[{ required: true }]}>
            <Input placeholder="如：大风紧急避险" />
          </Form.Item>
          <Form.Item name="weather_type" label="天气类型" rules={[{ required: true }]}>
            <Select options={[
              { value: '大风', label: '大风' }, { value: '暴雨', label: '暴雨' },
              { value: '暴雪', label: '暴雪' }, { value: '高温', label: '高温' },
              { value: '沙尘暴', label: '沙尘暴' }, { value: '雷电', label: '雷电' },
            ]} />
          </Form.Item>
          <Row gutter={16}>
            <Col span={12}><Form.Item name="wind_scale_gt" label="风力超过(级)触发"><InputNumber min={1} max={12} style={{ width: '100%' }} /></Form.Item></Col>
            <Col span={12}><Form.Item name="precip_gt" label="降雨超过(mm)触发"><InputNumber min={1} max={200} style={{ width: '100%' }} /></Form.Item></Col>
          </Row>
          <Form.Item name="warning_level" label="预警级别触发">
            <Select options={[
              { value: 'red', label: '红色预警' }, { value: 'orange', label: '橙色预警' },
              { value: 'yellow', label: '黄色预警' }, { value: 'blue', label: '蓝色预警' },
            ]} />
          </Form.Item>
          <Form.Item name="action" label="响应动作" rules={[{ required: true }]}>
            <Select options={[
              { value: '暂停清扫', label: '暂停清扫' },
              { value: '返回充电', label: '返回充电' },
              { value: '降低速度', label: '降低速度' },
            ]} />
          </Form.Item>
          <Form.Item name="remark" label="备注"><Input.TextArea rows={2} /></Form.Item>
        </Form>
      </Modal>
    </div>
  );

  // ---- Tab 4: 频次调整 ----
  const renderAdjustmentsTab = () => (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button type="primary" icon={<ThunderboltOutlined />} onClick={handleGenAdjustment}>
          生成频次分析
        </Button>
      </Space>
      <Table
        dataSource={adjustments}
        rowKey="id"
        size="small"
        pagination={{ current: adjPage, total: adjTotal, pageSize: 10, onChange: fetchAdjustments }}
        columns={[
          { title: '当前频次', dataIndex: 'current_freq', width: 100 },
          { title: '建议频次', dataIndex: 'suggested_freq', width: 100,
            render: (v: string) => <Tag color="blue">{v}</Tag> },
          { title: '调整原因', dataIndex: 'reason', width: 300, ellipsis: true },
          { title: '状态', dataIndex: 'status', width: 80,
            render: (v: number) => {
              const map: Record<number, { text: string; color: string }> = {
                1: { text: '建议', color: 'blue' },
                2: { text: '已采纳', color: 'green' },
                3: { text: '已忽略', color: 'default' },
              };
              return <Tag color={map[v]?.color}>{map[v]?.text || v}</Tag>;
            } },
          { title: '时间', dataIndex: 'create_time', width: 160 },
        ]}
      />
    </div>
  );

  const tabItems = [
    { key: 'advice' as TabKey, label: '清扫建议' },
    { key: 'rules' as TabKey, label: '决策规则' },
    { key: 'extreme' as TabKey, label: '极端天气规则' },
    { key: 'adjustments' as TabKey, label: '频次调整' },
  ];

  return (
    <div>
      <Card style={{ marginBottom: 16 }}>
        <Space>
          <Select
            placeholder="选择电站"
            style={{ width: 260 }}
            value={stationId}
            onChange={setStationId}
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
        <Tabs
          activeKey={activeTab}
          onChange={(k) => setActiveTab(k as TabKey)}
          items={tabItems.map(tab => ({
            key: tab.key,
            label: tab.label,
            children: tab.key === 'advice' ? renderAdviceTab()
              : tab.key === 'rules' ? renderRulesTab()
              : tab.key === 'extreme' ? renderExtremeTab()
              : renderAdjustmentsTab(),
          }))}
        />
      )}
    </div>
  );
}
