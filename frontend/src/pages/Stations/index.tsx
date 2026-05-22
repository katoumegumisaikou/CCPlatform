import { useState, useEffect } from 'react';
import {
  Table, Button, Modal, Form, Input, InputNumber, Popconfirm, message, Space, Tag,
  Card, Row, Col, Typography,
} from 'antd';
import {
  PlusOutlined, EditOutlined, DeleteOutlined, ReloadOutlined, SearchOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { stationApi } from '../../api/stations';
import type { Station } from '../../types';

const { Title } = Typography;

interface FilterParams {
  station_name?: string;
}

export default function StationsPage() {
  const [data, setData] = useState<Station[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingRecord, setEditingRecord] = useState<Station | null>(null);
  const [form] = Form.useForm();
  const [filters, setFilters] = useState<FilterParams>({});

  const fetchData = async () => {
    setLoading(true);
    try {
      const params: Record<string, unknown> = {
        page,
        size: pageSize,
      };
      if (filters.station_name) params.station_name = filters.station_name;
      const res = await stationApi.list(params);
      setData(res.list);
      setTotal(res.total);
    } catch {
      // error handled by interceptor
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [page, pageSize]);

  const handleSearch = () => {
    setPage(1);
    fetchData();
  };

  const handleReset = () => {
    setFilters({});
    setPage(1);
  };

  const handleCreate = () => {
    setEditingRecord(null);
    form.resetFields();
    setModalVisible(true);
  };

  const handleEdit = (record: Station) => {
    setEditingRecord(record);
    form.setFieldsValue({
      station_code: record.station_code,
      station_name: record.station_name,
      capacity: record.capacity,
      longitude: record.longitude,
      latitude: record.latitude,
      location: record.location,
      panel_area: record.panel_area,
      panel_count: record.panel_count,
    });
    setModalVisible(true);
  };

  const handleDelete = async (id: string) => {
    try {
      await stationApi.delete(id);
      message.success('删除成功');
      fetchData();
    } catch {
      // error handled by interceptor
    }
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      if (editingRecord) {
        await stationApi.update(editingRecord.station_id, values);
        message.success('更新成功');
      } else {
        await stationApi.create(values);
        message.success('创建成功');
      }
      setModalVisible(false);
      form.resetFields();
      fetchData();
    } catch (err) {
      if (err && typeof err === 'object' && 'errorFields' in err) {
        // form validation error, do nothing
        return;
      }
      // API error handled by interceptor
    }
  };

  const columns: ColumnsType<Station> = [
    {
      title: '电站编码',
      dataIndex: 'station_code',
      key: 'station_code',
      width: 140,
    },
    {
      title: '电站名称',
      dataIndex: 'station_name',
      key: 'station_name',
      width: 200,
      ellipsis: true,
    },
    {
      title: '装机容量(kW)',
      dataIndex: 'capacity',
      key: 'capacity',
      width: 130,
      align: 'right',
      render: (val: number) => val?.toLocaleString() ?? '-',
    },
    {
      title: '地址',
      dataIndex: 'address',
      key: 'address',
      ellipsis: true,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
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
      width: 180,
      align: 'center',
      render: (_val: unknown, record: Station) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定删除该电站吗？"
            description="删除后不可恢复"
            onConfirm={() => handleDelete(record.station_id)}
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
      <Title level={4} style={{ marginBottom: 16 }}>电站管理</Title>

      {/* 筛选栏 */}
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={[16, 16]} align="middle">
          <Col>
            <Input
              placeholder="搜索电站名称"
              value={filters.station_name || ''}
              onChange={(e) => setFilters((prev) => ({ ...prev, station_name: e.target.value }))}
              onPressEnter={handleSearch}
              style={{ width: 220 }}
              allowClear
            />
          </Col>
          <Col>
            <Space>
              <Button type="primary" icon={<SearchOutlined />} onClick={handleSearch}>
                搜索
              </Button>
              <Button onClick={handleReset}>重置</Button>
            </Space>
          </Col>
        </Row>
      </Card>

      {/* 操作栏 + 表格 */}
      <Card>
        <Row justify="space-between" align="middle" style={{ marginBottom: 16 }}>
          <Col>
            <Space>
              <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
                新建电站
              </Button>
              <Button icon={<ReloadOutlined />} onClick={fetchData}>
                刷新
              </Button>
            </Space>
          </Col>
        </Row>

        <Table<Station>
          rowKey="station_id"
          columns={columns}
          dataSource={data}
          loading={loading}
          pagination={{
            current: page,
            pageSize,
            total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (t) => `共 ${t} 条`,
            onChange: (p, ps) => {
              setPage(p);
              setPageSize(ps);
            },
          }}
          locale={{
            emptyText: '暂无电站数据',
          }}
          scroll={{ x: 900 }}
        />
      </Card>

      {/* 创建/编辑弹窗 */}
      <Modal
        title={editingRecord ? '编辑电站' : '新建电站'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => {
          setModalVisible(false);
          form.resetFields();
        }}
        okText="保存"
        cancelText="取消"
        destroyOnClose
        width={560}
      >
        <Form
          form={form}
          layout="vertical"
          style={{ marginTop: 16 }}
        >
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="station_code"
                label="电站编码"
                rules={[{ required: true, message: '请输入电站编码' }]}
              >
                <Input placeholder="请输入电站编码" disabled={!!editingRecord} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="station_name"
                label="电站名称"
                rules={[{ required: true, message: '请输入电站名称' }]}
              >
                <Input placeholder="请输入电站名称" />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="capacity"
                label="装机容量(kW)"
                rules={[{ required: true, message: '请输入装机容量' }]}
              >
                <InputNumber
                  placeholder="请输入装机容量"
                  min={0}
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="longitude"
                label="经度"
                rules={[{ required: true, message: '请输入经度' }]}
              >
                <InputNumber
                  placeholder="请输入经度"
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="latitude"
                label="纬度"
                rules={[{ required: true, message: '请输入纬度' }]}
              >
                <InputNumber
                  placeholder="请输入纬度"
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="panel_area" label="组件面积(㎡)">
                <InputNumber placeholder="组件面积" min={0} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="panel_count" label="组件数量">
                <InputNumber placeholder="组件数量" min={0} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item
            name="location"
            label="地址"
            rules={[{ required: true, message: '请输入地址' }]}
          >
            <Input placeholder="请输入详细地址" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
