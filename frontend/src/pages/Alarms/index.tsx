import { useState, useEffect } from 'react';
import {
  Table, Button, Modal, Form, Input, Select, Radio, Popconfirm, message, Space, Tag,
  Card, Row, Col, Typography, Tabs, Descriptions, Statistic,
} from 'antd';
import {
  PlusOutlined, EditOutlined, DeleteOutlined, ReloadOutlined, SearchOutlined,
  ExclamationCircleOutlined, CheckCircleOutlined, SyncOutlined, EyeOutlined,
  StopOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { alarmApi } from '../../api/alarms';
import { stationApi } from '../../api/stations';
import { robotApi } from '../../api/robots';
import type { Alarm, AlarmRule, Station, Robot } from '../../types';
import { ALARM_LEVEL_MAP, ALARM_HANDLE_MAP } from '../../utils';

const { Title } = Typography;
const { TextArea } = Input;

interface AlarmFilterParams {
  alarm_level?: number;
  handle_status?: number;
  robot_id?: string;
  station_id?: string;
}

interface AlarmStats {
  total: number;
  byLevel: Record<string, number>;
}

export default function AlarmsPage() {
  const [activeTab, setActiveTab] = useState<string>('records');

  // ========== 告警记录 state ==========
  const [alarmData, setAlarmData] = useState<Alarm[]>([]);
  const [alarmLoading, setAlarmLoading] = useState(false);
  const [alarmTotal, setAlarmTotal] = useState(0);
  const [alarmPage, setAlarmPage] = useState(1);
  const [alarmPageSize, setAlarmPageSize] = useState(10);
  const [alarmFilters, setAlarmFilters] = useState<AlarmFilterParams>({});
  const [alarmStats, setAlarmStats] = useState<AlarmStats>({ total: 0, byLevel: {} });
  const [handleModalVisible, setHandleModalVisible] = useState(false);
  const [handlingAlarm, setHandlingAlarm] = useState<Alarm | null>(null);
  const [handleForm] = Form.useForm();
  const [detailVisible, setDetailVisible] = useState(false);
  const [detailAlarm, setDetailAlarm] = useState<Alarm | null>(null);
  const [stations, setStations] = useState<Station[]>([]);
  const [robots, setRobots] = useState<Robot[]>([]);

  // ========== 告警规则 state ==========
  const [ruleData, setRuleData] = useState<AlarmRule[]>([]);
  const [ruleLoading, setRuleLoading] = useState(false);
  const [ruleTotal, setRuleTotal] = useState(0);
  const [rulePage, setRulePage] = useState(1);
  const [rulePageSize, setRulePageSize] = useState(10);
  const [ruleModalVisible, setRuleModalVisible] = useState(false);
  const [editingRule, setEditingRule] = useState<AlarmRule | null>(null);
  const [ruleForm] = Form.useForm();
  const [ruleScopeType, setRuleScopeType] = useState<string>('global');

  // ========== 基础数据加载 ==========
  const fetchStations = async () => {
    try {
      const res = await stationApi.getAll();
      setStations(res);
    } catch {
      // error handled by interceptor
    }
  };

  const fetchRobots = async () => {
    try {
      const res = await robotApi.list({ size: 999 });
      setRobots(res.list);
    } catch {
      // error handled by interceptor
    }
  };

  useEffect(() => {
    fetchStations();
    fetchRobots();
  }, []);

  // ========== 告警统计 ==========
  const fetchAlarmStats = async () => {
    try {
      const res = await alarmApi.stats();
      const byLevel: Record<string, number> = {};
      let total = 0;
      for (const [key, val] of Object.entries(res)) {
        byLevel[key] = val;
        total += val;
      }
      setAlarmStats({ total, byLevel });
    } catch {
      // error handled by interceptor
    }
  };

  // ========== 告警记录 ==========
  const fetchAlarmData = async () => {
    setAlarmLoading(true);
    try {
      const params: Record<string, unknown> = {
        page: alarmPage,
        size: alarmPageSize,
      };
      if (alarmFilters.alarm_level !== undefined) params.alarm_level = alarmFilters.alarm_level;
      if (alarmFilters.handle_status !== undefined) params.handle_status = alarmFilters.handle_status;
      if (alarmFilters.robot_id) params.robot_id = alarmFilters.robot_id;
      if (alarmFilters.station_id) params.station_id = alarmFilters.station_id;
      const res = await alarmApi.list(params);
      setAlarmData(res.list);
      setAlarmTotal(res.total);
    } catch {
      // error handled by interceptor
    } finally {
      setAlarmLoading(false);
    }
  };

  useEffect(() => {
    fetchAlarmData();
  }, [alarmPage, alarmPageSize]);

  useEffect(() => {
    if (activeTab === 'records') {
      fetchAlarmData();
      fetchAlarmStats();
    }
  }, [activeTab]);

  const handleAlarmSearch = () => {
    setAlarmPage(1);
    fetchAlarmData();
  };

  const handleAlarmReset = () => {
    setAlarmFilters({});
    setAlarmPage(1);
  };

  const handleShowDetail = async (record: Alarm) => {
    try {
      const res = await alarmApi.getById(record.alarm_id);
      setDetailAlarm(res);
      setDetailVisible(true);
    } catch {
      // try fallback with list data
      setDetailAlarm(record);
      setDetailVisible(true);
    }
  };

  const handleOpenHandle = (record: Alarm) => {
    setHandlingAlarm(record);
    handleForm.resetFields();
    handleForm.setFieldsValue({
      handle_status: record.handle_status === 0 ? 1 : record.handle_status,
    });
    setHandleModalVisible(true);
  };

  const handleSubmitHandle = async () => {
    if (!handlingAlarm) return;
    try {
      const values = await handleForm.validateFields();
      await alarmApi.handle(handlingAlarm.alarm_id, {
        handle_status: values.handle_status,
        handle_remark: values.handle_remark,
      });
      message.success('处理成功');
      setHandleModalVisible(false);
      handleForm.resetFields();
      fetchAlarmData();
      fetchAlarmStats();
    } catch (err) {
      if (err && typeof err === 'object' && 'errorFields' in err) {
        return;
      }
    }
  };

  const handleQuickStatusChange = async (record: Alarm, handleStatus: number) => {
    try {
      await alarmApi.handle(record.alarm_id, { handle_status: handleStatus });
      const statusText = ALARM_HANDLE_MAP[handleStatus]?.text || '目标状态';
      setAlarmData(prev => prev.map(item => (
        item.alarm_id === record.alarm_id
          ? { ...item, handle_status: handleStatus }
          : item
      )));
      setDetailAlarm(prev => (
        prev?.alarm_id === record.alarm_id
          ? { ...prev, handle_status: handleStatus }
          : prev
      ));
      message.success(`告警状态已更新为${statusText}`);
      fetchAlarmData();
      fetchAlarmStats();
    } catch {
      // error handled by interceptor
    }
  };

  const alarmColumns: ColumnsType<Alarm> = [
    {
      title: '告警ID',
      dataIndex: 'alarm_id',
      key: 'alarm_id',
      width: 100,
      ellipsis: true,
    },
    {
      title: '级别',
      dataIndex: 'alarm_level',
      key: 'alarm_level',
      width: 80,
      align: 'center',
      render: (level: number) => {
        const cfg = ALARM_LEVEL_MAP[level] || { text: '未知', color: 'default' };
        return <Tag color={cfg.color}>{cfg.text}</Tag>;
      },
    },
    {
      title: '告警类型',
      dataIndex: 'alarm_type',
      key: 'alarm_type',
      width: 120,
    },
    {
      title: '机器人',
      dataIndex: 'robot_id',
      key: 'robot_id',
      width: 140,
      ellipsis: true,
    },
    {
      title: '告警内容',
      dataIndex: 'alarm_content',
      key: 'alarm_content',
      ellipsis: true,
    },
    {
      title: '告警时间',
      dataIndex: 'alarm_time',
      key: 'alarm_time',
      width: 170,
    },
    {
      title: '处理状态',
      dataIndex: 'handle_status',
      key: 'handle_status',
      width: 100,
      align: 'center',
      render: (status: number) => {
        const cfg = ALARM_HANDLE_MAP[status] || { text: '未知', color: 'default' };
        return <Tag color={cfg.color}>{cfg.text}</Tag>;
      },
    },
    {
      title: '操作',
      key: 'actions',
      width: 240,
      align: 'center',
      render: (_val: unknown, record: Alarm) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => handleShowDetail(record)}
          >
            详情
          </Button>
          {record.handle_status !== 3 && record.handle_status !== 4 && (
            <>
              <Button
                type="link"
                size="small"
                icon={<CheckCircleOutlined />}
                onClick={() => handleQuickStatusChange(record, 1)}
              >
                确认
              </Button>
              <Button
                type="link"
                size="small"
                icon={<SyncOutlined />}
                onClick={() => handleQuickStatusChange(record, 2)}
              >
                处理中
              </Button>
              <Button
                type="link"
                size="small"
                style={{ color: 'green' }}
                icon={<CheckCircleOutlined />}
                onClick={() => handleQuickStatusChange(record, 3)}
              >
                完成
              </Button>
              <Button
                type="link"
                size="small"
                style={{ color: '#999' }}
                icon={<StopOutlined />}
                onClick={() => handleQuickStatusChange(record, 4)}
              >
                忽略
              </Button>
            </>
          )}
        </Space>
      ),
    },
  ];

  // ========== 告警规则 ==========
  const fetchRuleData = async () => {
    setRuleLoading(true);
    try {
      const params: Record<string, unknown> = {
        page: rulePage,
        size: rulePageSize,
      };
      const res = await alarmApi.rules.list(params);
      setRuleData(res.list);
      setRuleTotal(res.total);
    } catch {
      // error handled by interceptor
    } finally {
      setRuleLoading(false);
    }
  };

  useEffect(() => {
    if (activeTab === 'rules') {
      fetchRuleData();
    }
  }, [activeTab, rulePage, rulePageSize]);

  const handleRuleCreate = () => {
    setEditingRule(null);
    setRuleScopeType('global');
    ruleForm.resetFields();
    ruleForm.setFieldsValue({ scope_type: 'global' });
    setRuleModalVisible(true);
  };

  const handleRuleEdit = (record: AlarmRule) => {
    setEditingRule(record);
    setRuleScopeType(record.scope_type || 'global');
    ruleForm.setFieldsValue({
      rule_name: record.rule_name,
      alarm_type: record.alarm_type,
      scope_type: record.scope_type || 'global',
      scope_id: record.scope_id,
      suppression_rule: record.suppression_rule,
      escalation_rule: record.escalation_rule,
    });
    setRuleModalVisible(true);
  };

  const handleRuleDelete = async (id: number) => {
    try {
      await alarmApi.rules.delete(id);
      message.success('删除成功');
      fetchRuleData();
    } catch {
      // error handled by interceptor
    }
  };

  const handleRuleSubmit = async () => {
    try {
      const values = await ruleForm.validateFields();
      if (editingRule) {
        await alarmApi.rules.update(editingRule.rule_id, values as Partial<AlarmRule>);
        message.success('更新成功');
      } else {
        await alarmApi.rules.create(values as Partial<AlarmRule>);
        message.success('创建成功');
      }
      setRuleModalVisible(false);
      ruleForm.resetFields();
      fetchRuleData();
    } catch (err) {
      if (err && typeof err === 'object' && 'errorFields' in err) {
        return;
      }
    }
  };

  const ruleColumns: ColumnsType<AlarmRule> = [
    {
      title: '规则名称',
      dataIndex: 'rule_name',
      key: 'rule_name',
      width: 180,
      ellipsis: true,
    },
    {
      title: '告警类型',
      dataIndex: 'alarm_type',
      key: 'alarm_type',
      width: 120,
    },
    {
      title: '作用域类型',
      dataIndex: 'scope_type',
      key: 'scope_type',
      width: 120,
      align: 'center',
      render: (type: string) => {
        const map: Record<string, { text: string; color: string }> = {
          global: { text: '全局', color: 'blue' },
          station: { text: '电站', color: 'green' },
          robot: { text: '机器人', color: 'orange' },
        };
        const cfg = map[type] || { text: type, color: 'default' };
        return <Tag color={cfg.color}>{cfg.text}</Tag>;
      },
    },
    {
      title: '作用域ID',
      dataIndex: 'scope_id',
      key: 'scope_id',
      width: 140,
      render: (val: string) => val || '-',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      align: 'center',
      render: (status: number) => (
        <Tag color={status === 1 ? 'green' : 'default'}>
          {status === 1 ? '启用' : '禁用'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'actions',
      width: 160,
      align: 'center',
      render: (_val: unknown, record: AlarmRule) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleRuleEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定删除该规则吗？"
            onConfirm={() => handleRuleDelete(record.rule_id)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <Title level={4} style={{ marginBottom: 16 }}>告警中心</Title>

      <Tabs activeKey={activeTab} onChange={setActiveTab}>
        {/* ==================== 告警记录 Tab ==================== */}
        <Tabs.TabPane tab="告警记录" key="records">
          {/* 统计卡片 */}
          <Row gutter={16} style={{ marginBottom: 16 }}>
            <Col span={4}>
              <Card size="small">
                <Statistic
                  title="告警总数"
                  value={alarmStats.total}
                  prefix={<ExclamationCircleOutlined />}
                />
              </Card>
            </Col>
            {Object.entries(ALARM_LEVEL_MAP).map(([level, cfg]) => (
              <Col span={4} key={level}>
                <Card size="small">
                  <Statistic
                    title={`${cfg.text}告警`}
                    value={alarmStats.byLevel[level] || 0}
                    valueStyle={{ color: cfg.color }}
                  />
                </Card>
              </Col>
            ))}
          </Row>

          {/* 筛选栏 */}
          <Card style={{ marginBottom: 16 }}>
            <Row gutter={[16, 16]} align="middle">
              <Col>
                <Select
                  placeholder="告警级别"
                  value={alarmFilters.alarm_level}
                  onChange={(val) => setAlarmFilters((prev) => ({ ...prev, alarm_level: val }))}
                  allowClear
                  style={{ width: 130 }}
                  options={Object.entries(ALARM_LEVEL_MAP).map(([k, v]) => ({
                    label: v.text,
                    value: Number(k),
                  }))}
                />
              </Col>
              <Col>
                <Select
                  placeholder="处理状态"
                  value={alarmFilters.handle_status}
                  onChange={(val) => setAlarmFilters((prev) => ({ ...prev, handle_status: val }))}
                  allowClear
                  style={{ width: 130 }}
                  options={Object.entries(ALARM_HANDLE_MAP).map(([k, v]) => ({
                    label: v.text,
                    value: Number(k),
                  }))}
                />
              </Col>
              <Col>
                <Select
                  placeholder="选择电站"
                  value={alarmFilters.station_id}
                  onChange={(val) => setAlarmFilters((prev) => ({ ...prev, station_id: val }))}
                  allowClear
                  style={{ width: 180 }}
                  options={stations.map((s) => ({
                    label: s.station_name,
                    value: s.station_id,
                  }))}
                />
              </Col>
              <Col>
                <Select
                  placeholder="选择机器人"
                  value={alarmFilters.robot_id}
                  onChange={(val) => setAlarmFilters((prev) => ({ ...prev, robot_id: val }))}
                  allowClear
                  showSearch
                  optionFilterProp="label"
                  style={{ width: 180 }}
                  options={robots.map((r) => ({
                    label: `${r.robot_name} (${r.robot_code})`,
                    value: r.robot_id,
                  }))}
                />
              </Col>
              <Col>
                <Space>
                  <Button type="primary" icon={<SearchOutlined />} onClick={handleAlarmSearch}>
                    搜索
                  </Button>
                  <Button onClick={handleAlarmReset}>重置</Button>
                </Space>
              </Col>
            </Row>
          </Card>

          {/* 表格 */}
          <Card>
            <Row style={{ marginBottom: 16 }}>
              <Col>
                <Button icon={<ReloadOutlined />} onClick={fetchAlarmData}>
                  刷新
                </Button>
              </Col>
            </Row>
            <Table<Alarm>
              rowKey="alarm_id"
              columns={alarmColumns}
              dataSource={alarmData}
              loading={alarmLoading}
              pagination={{
                current: alarmPage,
                pageSize: alarmPageSize,
                total: alarmTotal,
                showSizeChanger: true,
                showQuickJumper: true,
                showTotal: (t) => `共 ${t} 条`,
                onChange: (p, ps) => {
                  setAlarmPage(p);
                  setAlarmPageSize(ps);
                },
              }}
              locale={{
                emptyText: '暂无告警数据',
              }}
              scroll={{ x: 1200 }}
            />
          </Card>

          {/* 告警详情弹窗 */}
          <Modal
            title="告警详情"
            open={detailVisible}
            onCancel={() => setDetailVisible(false)}
            footer={[
              <Button key="close" onClick={() => setDetailVisible(false)}>关闭</Button>,
              detailAlarm && detailAlarm.handle_status !== 3 && detailAlarm.handle_status !== 4 && (
                <Button
                  key="handle"
                  type="primary"
                  onClick={() => {
                    setDetailVisible(false);
                    handleOpenHandle(detailAlarm);
                  }}
                >
                  处理
                </Button>
              ),
            ]}
            width={640}
          >
            {detailAlarm && (
              <Descriptions bordered column={2} size="small" style={{ marginTop: 8 }}>
                <Descriptions.Item label="告警ID">{detailAlarm.alarm_id}</Descriptions.Item>
                <Descriptions.Item label="告警类型">{detailAlarm.alarm_type}</Descriptions.Item>
                <Descriptions.Item label="告警级别">
                  <Tag color={ALARM_LEVEL_MAP[detailAlarm.alarm_level]?.color}>
                    {ALARM_LEVEL_MAP[detailAlarm.alarm_level]?.text || '未知'}
                  </Tag>
                </Descriptions.Item>
                <Descriptions.Item label="处理状态">
                  <Tag color={ALARM_HANDLE_MAP[detailAlarm.handle_status]?.color}>
                    {ALARM_HANDLE_MAP[detailAlarm.handle_status]?.text || '未知'}
                  </Tag>
                </Descriptions.Item>
                <Descriptions.Item label="机器人">{detailAlarm.robot_id}</Descriptions.Item>
                <Descriptions.Item label="电站">{detailAlarm.station_id}</Descriptions.Item>
                <Descriptions.Item label="告警时间">{detailAlarm.alarm_time}</Descriptions.Item>
                <Descriptions.Item label="处理人">{detailAlarm.handler || '-'}</Descriptions.Item>
                <Descriptions.Item label="处理时间" span={2}>
                  {detailAlarm.handle_time || '-'}
                </Descriptions.Item>
                <Descriptions.Item label="告警内容" span={2}>
                  {detailAlarm.alarm_content}
                </Descriptions.Item>
                <Descriptions.Item label="处理备注" span={2}>
                  {detailAlarm.handle_remark || '-'}
                </Descriptions.Item>
              </Descriptions>
            )}
          </Modal>

          {/* 处理弹窗 */}
          <Modal
            title="处理告警"
            open={handleModalVisible}
            onOk={handleSubmitHandle}
            onCancel={() => {
              setHandleModalVisible(false);
              handleForm.resetFields();
            }}
            okText="保存"
            cancelText="取消"
            destroyOnClose
            width={500}
          >
            <Form
              form={handleForm}
              layout="vertical"
              style={{ marginTop: 16 }}
            >
              <Form.Item
                name="handle_status"
                label="处理状态"
                rules={[{ required: true, message: '请选择处理状态' }]}
              >
                <Select
                  placeholder="请选择处理状态"
                  options={Object.entries(ALARM_HANDLE_MAP).map(([k, v]) => ({
                    label: v.text,
                    value: Number(k),
                  }))}
                />
              </Form.Item>
              <Form.Item
                name="handle_remark"
                label="处理备注"
              >
                <TextArea
                  rows={4}
                  placeholder="请输入处理备注信息（可选）"
                  maxLength={500}
                  showCount
                />
              </Form.Item>
            </Form>
          </Modal>
        </Tabs.TabPane>

        {/* ==================== 告警规则 Tab ==================== */}
        <Tabs.TabPane tab="告警规则" key="rules">
          <Card>
            <Row style={{ marginBottom: 16 }}>
              <Col>
                <Space>
                  <Button type="primary" icon={<PlusOutlined />} onClick={handleRuleCreate}>
                    新建规则
                  </Button>
                  <Button icon={<ReloadOutlined />} onClick={fetchRuleData}>
                    刷新
                  </Button>
                </Space>
              </Col>
            </Row>

            <Table<AlarmRule>
              rowKey="rule_id"
              columns={ruleColumns}
              dataSource={ruleData}
              loading={ruleLoading}
              pagination={{
                current: rulePage,
                pageSize: rulePageSize,
                total: ruleTotal,
                showSizeChanger: true,
                showQuickJumper: true,
                showTotal: (t) => `共 ${t} 条`,
                onChange: (p, ps) => {
                  setRulePage(p);
                  setRulePageSize(ps);
                },
              }}
              locale={{
                emptyText: '暂无告警规则',
              }}
              scroll={{ x: 800 }}
            />
          </Card>

          {/* 规则创建/编辑弹窗 */}
          <Modal
            title={editingRule ? '编辑告警规则' : '新建告警规则'}
            open={ruleModalVisible}
            onOk={handleRuleSubmit}
            onCancel={() => {
              setRuleModalVisible(false);
              ruleForm.resetFields();
            }}
            okText="保存"
            cancelText="取消"
            destroyOnClose
            width={600}
          >
            <Form
              form={ruleForm}
              layout="vertical"
              style={{ marginTop: 16 }}
              initialValues={{ scope_type: 'global' }}
            >
              <Form.Item
                name="rule_name"
                label="规则名称"
                rules={[{ required: true, message: '请输入规则名称' }]}
              >
                <Input placeholder="请输入规则名称" />
              </Form.Item>
              <Form.Item
                name="alarm_type"
                label="告警类型"
                rules={[{ required: true, message: '请输入告警类型' }]}
              >
                <Input placeholder="例如：故障告警、离线告警、低电量告警" />
              </Form.Item>
              <Form.Item
                name="scope_type"
                label="作用域类型"
                rules={[{ required: true, message: '请选择作用域类型' }]}
              >
                <Radio.Group
                  onChange={(e) => setRuleScopeType(e.target.value)}
                  optionType="button"
                  buttonStyle="solid"
                >
                  <Radio.Button value="global">全局</Radio.Button>
                  <Radio.Button value="station">电站</Radio.Button>
                  <Radio.Button value="robot">机器人</Radio.Button>
                </Radio.Group>
              </Form.Item>
              {ruleScopeType !== 'global' && (
                <Form.Item
                  name="scope_id"
                  label={ruleScopeType === 'station' ? '电站ID' : '机器人ID'}
                  rules={[{ required: true, message: '请输入作用域ID' }]}
                >
                  <Input
                    placeholder={
                      ruleScopeType === 'station'
                        ? '请输入电站ID'
                        : '请输入机器人ID'
                    }
                  />
                </Form.Item>
              )}
              <Form.Item
                name="suppression_rule"
                label="抑制规则（JSON）"
                extra={'例如：{"max_count": 5, "window_minutes": 10}'}
              >
                <TextArea
                  rows={3}
                  placeholder='请输入抑制规则 JSON，例如 {"max_count": 5, "window_minutes": 10}'
                />
              </Form.Item>
              <Form.Item
                name="escalation_rule"
                label="升级规则（JSON）"
                extra={'例如：{"timeout_minutes": 30, "escalate_to_level": 3}'}
              >
                <TextArea
                  rows={3}
                  placeholder='请输入升级规则 JSON，例如 {"timeout_minutes": 30, "escalate_to_level": 3}'
                />
              </Form.Item>
            </Form>
          </Modal>
        </Tabs.TabPane>
      </Tabs>
    </div>
  );
}
