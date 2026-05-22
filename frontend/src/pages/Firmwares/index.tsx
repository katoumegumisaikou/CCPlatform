import { useState, useEffect, useCallback } from 'react';
import {
  Card, Table, Button, Modal, Form, Input, Select, Upload, Space, Tag, message, Popconfirm, Checkbox, Row, Col, Empty,
} from 'antd';
import {
  PlusOutlined, DeleteOutlined, ReloadOutlined, UploadOutlined, CloudUploadOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import type { UploadFile } from 'antd/es/upload';
import { firmwareApi } from '../../api/system';
import { stationApi } from '../../api/stations';
import { robotApi } from '../../api/robots';
import type { Firmware, Station, Robot } from '../../types';
import { ROBOT_TYPE_MAP } from '../../utils';

interface FirmwareWithKey extends Firmware {
  key: string;
}

export default function FirmwaresPage() {
  const [data, setData] = useState<FirmwareWithKey[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  // upload / edit modal
  const [modalOpen, setModalOpen] = useState(false);
  const [modalTitle, setModalTitle] = useState('上传固件');
  const [confirmLoading, setConfirmLoading] = useState(false);
  const [fileList, setFileList] = useState<UploadFile[]>([]);
  const [form] = Form.useForm();

  // upgrade modal
  const [upgradeModalOpen, setUpgradeModalOpen] = useState(false);
  const [upgradeFirmwareId, setUpgradeFirmwareId] = useState<string>('');
  const [upgradeConfirmLoading, setUpgradeConfirmLoading] = useState(false);
  const [robots, setRobots] = useState<Robot[]>([]);
  const [robotsLoading, setRobotsLoading] = useState(false);
  const [selectedRobotIds, setSelectedRobotIds] = useState<string[]>([]);
  const [upgradeStationFilter, setUpgradeStationFilter] = useState<string | undefined>(undefined);
  const [stations, setStations] = useState<Station[]>([]);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await firmwareApi.list({ page, size: pageSize });
      const list = res?.list || [];
      setData(list.map((it) => ({ ...it, key: it.firmware_id })));
      setTotal(res?.total || 0);
    } catch {
      message.error('获取固件列表失败');
    } finally {
      setLoading(false);
    }
  }, [page, pageSize]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  useEffect(() => {
    stationApi.getAll().then((res) => {
      setStations(res || []);
    }).catch(() => {
      message.error('获取电站列表失败');
    });
  }, []);

  const handleOpenCreate = () => {
    setModalTitle('上传固件');
    form.resetFields();
    setFileList([]);
    setModalOpen(true);
  };

  const handleDelete = async (id: string) => {
    try {
      await firmwareApi.delete(id);
      message.success('删除成功');
      fetchData();
    } catch {
      message.error('删除失败');
    }
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setConfirmLoading(true);
      const fd = new FormData();
      if (fileList.length > 0 && fileList[0].originFileObj) {
        fd.append('file', fileList[0].originFileObj);
      }
      fd.append('version', values.version);
      fd.append('robot_type', String(values.robot_type));
      if (values.description) fd.append('description', values.description);

      await firmwareApi.create(fd);
      message.success('上传成功');
      setModalOpen(false);
      fetchData();
    } catch (err: any) {
      if (err?.errorFields) return; // form validation
      message.error('上传失败');
    } finally {
      setConfirmLoading(false);
    }
  };

  const handleOpenUpgrade = async (firmwareId: string) => {
    setUpgradeFirmwareId(firmwareId);
    setSelectedRobotIds([]);
    setUpgradeStationFilter(undefined);
    setRobotsLoading(true);
    setUpgradeModalOpen(true);
    try {
      const params: Record<string, unknown> = { size: 999 };
      const res = await robotApi.list(params);
      setRobots(res?.list || []);
    } catch {
      message.error('获取机器人列表失败');
    } finally {
      setRobotsLoading(false);
    }
  };

  const handleUpgradeSubmit = async () => {
    if (selectedRobotIds.length === 0) {
      message.warning('请至少选择一个机器人');
      return;
    }
    setUpgradeConfirmLoading(true);
    try {
      await firmwareApi.upgrade(upgradeFirmwareId, selectedRobotIds);
      message.success('固件升级指令已下发');
      setUpgradeModalOpen(false);
    } catch {
      message.error('下发升级失败');
    } finally {
      setUpgradeConfirmLoading(false);
    }
  };

  const fileSizeFormatter = (bytes: number): string => {
    if (!bytes || bytes === 0) return '0 B';
    const units = ['B', 'KB', 'MB', 'GB'];
    let i = 0;
    let size = bytes;
    while (size >= 1024 && i < units.length - 1) {
      size /= 1024;
      i++;
    }
    return `${size.toFixed(1)} ${units[i]}`;
  };

  const columns: ColumnsType<FirmwareWithKey> = [
    { title: '固件ID', dataIndex: 'firmware_id', key: 'firmware_id', width: 160, ellipsis: true },
    { title: '版本号', dataIndex: 'version', key: 'version', width: 120 },
    { title: '文件名', dataIndex: 'file_name', key: 'file_name', width: 200, ellipsis: true },
    {
      title: '文件大小', dataIndex: 'file_size', key: 'file_size', width: 100,
      render: (v: number) => fileSizeFormatter(v),
    },
    {
      title: '适用机器人类型', dataIndex: 'robot_type', key: 'robot_type', width: 140,
      render: (v: number) => <Tag>{ROBOT_TYPE_MAP[v] || '未知'}</Tag>,
    },
    { title: '描述', dataIndex: 'description', key: 'description', width: 200, ellipsis: true, render: (v: string) => v || '-' },
    {
      title: '状态', dataIndex: 'status', key: 'status', width: 100,
      render: (v: number) => (v === 1 ? <Tag color="green">启用</Tag> : <Tag color="default">停用</Tag>),
    },
    { title: '创建时间', dataIndex: 'create_time', key: 'create_time', width: 180 },
    {
      title: '操作', key: 'action', width: 180, fixed: 'right',
      render: (_, record) => (
        <Space size="small">
          <Popconfirm
            title="确定删除该固件？"
            onConfirm={() => handleDelete(record.firmware_id)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="link" danger icon={<DeleteOutlined />} size="small">
              删除
            </Button>
          </Popconfirm>
          <Button
            type="link"
            icon={<CloudUploadOutlined />}
            size="small"
            onClick={() => handleOpenUpgrade(record.firmware_id)}
          >
            下发升级
          </Button>
        </Space>
      ),
    },
  ];

  // Filter robots by station for upgrade modal
  const filteredRobots = upgradeStationFilter
    ? robots.filter((r) => r.station_id === upgradeStationFilter)
    : robots;

  return (
    <div>
      <Card
        title="固件管理"
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} onClick={fetchData}>刷新</Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={handleOpenCreate}>
              上传固件
            </Button>
          </Space>
        }
      >
        <Table
          columns={columns}
          dataSource={data}
          loading={loading}
          scroll={{ x: 1300 }}
          pagination={{
            current: page,
            pageSize,
            total,
            showSizeChanger: true,
            showTotal: (t) => `共 ${t} 条`,
            onChange: (p, ps) => {
              setPage(p);
              setPageSize(ps);
            },
          }}
        />
      </Card>

      {/* 上传/编辑固件 Modal */}
      <Modal
        title={modalTitle}
        open={modalOpen}
        onOk={handleSubmit}
        onCancel={() => setModalOpen(false)}
        confirmLoading={confirmLoading}
        okText="上传"
        cancelText="取消"
        destroyOnClose
      >
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item
            name="version"
            label="版本号"
            rules={[{ required: true, message: '请输入版本号' }]}
          >
            <Input placeholder="如 v1.2.3" />
          </Form.Item>
          <Form.Item
            name="robot_type"
            label="适用机器人类型"
            rules={[{ required: true, message: '请选择机器人类型' }]}
          >
            <Select
              placeholder="选择机器人类型"
              options={Object.entries(ROBOT_TYPE_MAP).map(([k, v]) => ({ label: v, value: Number(k) }))}
            />
          </Form.Item>
          <Form.Item
            name="description"
            label="描述"
          >
            <Input.TextArea rows={3} placeholder="固件更新说明" />
          </Form.Item>
          <Form.Item
            label="固件文件"
            required
          >
            <Upload
              fileList={fileList}
              beforeUpload={(file) => {
                setFileList([{
                  uid: '-1',
                  name: file.name,
                  status: 'done',
                  originFileObj: file as any,
                } as UploadFile]);
                return false;
              }}
              onRemove={() => setFileList([])}
              maxCount={1}
            >
              <Button icon={<UploadOutlined />}>选择文件</Button>
            </Upload>
          </Form.Item>
        </Form>
      </Modal>

      {/* 下发升级 Modal */}
      <Modal
        title="下发固件升级"
        open={upgradeModalOpen}
        onOk={handleUpgradeSubmit}
        onCancel={() => setUpgradeModalOpen(false)}
        confirmLoading={upgradeConfirmLoading}
        okText="确认下发"
        cancelText="取消"
        destroyOnClose
        width={700}
      >
        <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
          <Col span={12}>
            <Select
              placeholder="按电站筛选"
              allowClear
              style={{ width: '100%' }}
              value={upgradeStationFilter}
              onChange={(val) => setUpgradeStationFilter(val)}
              options={stations.map((s) => ({ label: s.station_name, value: s.station_id }))}
            />
          </Col>
          <Col span={12}>
            <Space>
              <Button size="small" onClick={() => setSelectedRobotIds(filteredRobots.map((r) => r.robot_id))}>
                全选
              </Button>
              <Button size="small" onClick={() => setSelectedRobotIds([])}>
                取消全选
              </Button>
            </Space>
          </Col>
        </Row>
        {robotsLoading ? (
          <Empty description="加载机器人列表中..." />
        ) : filteredRobots.length === 0 ? (
          <Empty description="暂无可升级的机器人" />
        ) : (
          <Checkbox.Group
            value={selectedRobotIds}
            onChange={(vals) => setSelectedRobotIds(vals as string[])}
            style={{ width: '100%', maxHeight: 350, overflowY: 'auto' }}
          >
            <Row gutter={[8, 8]}>
              {filteredRobots.map((r) => (
                <Col span={12} key={r.robot_id}>
                  <Checkbox value={r.robot_id}>
                    {r.robot_name || r.robot_id} ({ROBOT_TYPE_MAP[r.robot_type] || '未知'})
                  </Checkbox>
                </Col>
              ))}
            </Row>
          </Checkbox.Group>
        )}
      </Modal>
    </div>
  );
}
