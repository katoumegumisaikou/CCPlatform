import { useState, useEffect, useCallback } from 'react';
import {
  Card, Table, Button, Modal, Form, Input, Select, Radio, Space, Tag, message, Popconfirm, Alert,
} from 'antd';
import {
  PlusOutlined, DeleteOutlined, ReloadOutlined, EyeOutlined, EditOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { cameraApi } from '../../api/system';
import { stationApi } from '../../api/stations';
import type { Camera, Station } from '../../types';

interface CameraWithKey extends Camera {
  key: string;
}

export default function CamerasPage() {
  const [data, setData] = useState<CameraWithKey[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  // create/edit modal
  const [modalOpen, setModalOpen] = useState(false);
  const [modalTitle, setModalTitle] = useState('添加摄像头');
  const [confirmLoading, setConfirmLoading] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form] = Form.useForm();
  const [stations, setStations] = useState<Station[]>([]);

  // preview modal
  const [previewModalOpen, setPreviewModalOpen] = useState(false);
  const [previewCamera, setPreviewCamera] = useState<Camera | null>(null);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await cameraApi.list({ page, size: pageSize });
      const list = res?.list || [];
      setData(list.map((it) => ({ ...it, key: it.camera_id })));
      setTotal(res?.total || 0);
    } catch {
      message.error('获取摄像头列表失败');
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
    setEditingId(null);
    setModalTitle('添加摄像头');
    form.resetFields();
    setModalOpen(true);
  };

  const handleOpenEdit = (record: Camera) => {
    setEditingId(record.camera_id);
    setModalTitle('编辑摄像头');
    form.setFieldsValue({
      camera_name: record.camera_name,
      station_id: record.station_id,
      stream_url: record.stream_url,
      camera_type: record.camera_type,
      position: record.position,
    });
    setModalOpen(true);
  };

  const handleDelete = async (id: string) => {
    try {
      await cameraApi.delete(id);
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
      const payload: Partial<Camera> = {
        camera_name: values.camera_name,
        station_id: values.station_id,
        stream_url: values.stream_url,
        camera_type: values.camera_type,
        position: values.position || '',
      };
      if (editingId) {
        await cameraApi.update(editingId, payload);
        message.success('更新成功');
      } else {
        await cameraApi.create(payload);
        message.success('创建成功');
      }
      setModalOpen(false);
      fetchData();
    } catch (err: any) {
      if (err?.errorFields) return;
      message.error('操作失败');
    } finally {
      setConfirmLoading(false);
    }
  };

  const handlePreview = (record: Camera) => {
    setPreviewCamera(record);
    setPreviewModalOpen(true);
  };

  const getStationName = (stationId: string): string => {
    const station = stations.find((s) => s.station_id === stationId);
    return station?.station_name || stationId;
  };

  const columns: ColumnsType<CameraWithKey> = [
    { title: '摄像头ID', dataIndex: 'camera_id', key: 'camera_id', width: 160, ellipsis: true },
    { title: '名称', dataIndex: 'camera_name', key: 'camera_name', width: 150, ellipsis: true },
    {
      title: '所属电站', dataIndex: 'station_id', key: 'station_id', width: 150, ellipsis: true,
      render: (v: string) => getStationName(v),
    },
    {
      title: '流类型', dataIndex: 'camera_type', key: 'camera_type', width: 100,
      render: (v: string) => {
        const isHls = v === 'hls' || v === 'HLS';
        return <Tag color={isHls ? 'green' : 'orange'}>{isHls ? 'HLS' : 'RTSP'}</Tag>;
      },
    },
    { title: '安装位置', dataIndex: 'position', key: 'position', width: 150, ellipsis: true, render: (v: string) => v || '-' },
    {
      title: '状态', dataIndex: 'status', key: 'status', width: 100,
      render: (v: number) => (v === 1 ? <Tag color="green">在线</Tag> : <Tag color="default">离线</Tag>),
    },
    {
      title: '操作', key: 'action', width: 200, fixed: 'right',
      render: (_, record) => (
        <Space size="small">
          <Button
            type="link"
            icon={<EyeOutlined />}
            size="small"
            onClick={() => handlePreview(record)}
          >
            预览
          </Button>
          <Button
            type="link"
            icon={<EditOutlined />}
            size="small"
            onClick={() => handleOpenEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定删除该摄像头？"
            onConfirm={() => handleDelete(record.camera_id)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="link" danger icon={<DeleteOutlined />} size="small">
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const isHlsStream = (camera: Camera | null): boolean => {
    if (!camera) return false;
    const t = camera.camera_type?.toLowerCase();
    return t === 'hls';
  };

  return (
    <div>
      <Alert
        type="warning"
        showIcon
        style={{ marginBottom: 16 }}
        message="注意：RTSP 流无法在浏览器中直接播放，需通过后端转码为 WebRTC/HLS 格式，或使用支持 RTSP 的播放器插件"
        closable
      />

      <Card
        title="摄像头管理"
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} onClick={fetchData}>刷新</Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={handleOpenCreate}>
              添加摄像头
            </Button>
          </Space>
        }
      >
        <Table
          columns={columns}
          dataSource={data}
          loading={loading}
          scroll={{ x: 1100 }}
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

      {/* 添加/编辑摄像头 Modal */}
      <Modal
        title={modalTitle}
        open={modalOpen}
        onOk={handleSubmit}
        onCancel={() => setModalOpen(false)}
        confirmLoading={confirmLoading}
        okText="保存"
        cancelText="取消"
        destroyOnClose
      >
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item
            name="camera_name"
            label="摄像头名称"
            rules={[{ required: true, message: '请输入摄像头名称' }]}
          >
            <Input placeholder="如：1号清洗区摄像头" />
          </Form.Item>
          <Form.Item
            name="station_id"
            label="所属电站"
            rules={[{ required: true, message: '请选择电站' }]}
          >
            <Select
              placeholder="选择电站"
              showSearch
              optionFilterProp="label"
              options={stations.map((s) => ({ label: s.station_name, value: s.station_id }))}
            />
          </Form.Item>
          <Form.Item
            name="stream_url"
            label="流地址"
            rules={[{ required: true, message: '请输入流地址' }]}
          >
            <Input placeholder="如：rtsp://192.168.1.100:554/stream 或 https://example.com/stream.m3u8" />
          </Form.Item>
          <Form.Item
            name="camera_type"
            label="流类型"
            rules={[{ required: true, message: '请选择流类型' }]}
          >
            <Radio.Group>
              <Radio value="rtsp">RTSP</Radio>
              <Radio value="hls">HLS</Radio>
            </Radio.Group>
          </Form.Item>
          <Form.Item
            name="position"
            label="安装位置"
          >
            <Input placeholder="如：电站东侧、清洗区上方" />
          </Form.Item>
        </Form>
      </Modal>

      {/* 预览 Modal */}
      <Modal
        title={`预览 - ${previewCamera?.camera_name || ''}`}
        open={previewModalOpen}
        onCancel={() => setPreviewModalOpen(false)}
        footer={null}
        width={800}
        destroyOnClose
      >
        {previewCamera && isHlsStream(previewCamera) ? (
          <div style={{ width: '100%', height: 450, background: '#000', borderRadius: 4, overflow: 'hidden' }}>
            <iframe
              src={previewCamera.stream_url}
              title="camera-preview"
              style={{ width: '100%', height: '100%', border: 'none' }}
              allowFullScreen
            />
          </div>
        ) : (
          <div style={{ textAlign: 'center', padding: 48 }}>
            {previewCamera ? (
              <>
                <Alert
                  type="warning"
                  showIcon
                  style={{ marginBottom: 16 }}
                  message="RTSP 流无法在浏览器中直接播放"
                  description="请使用支持 RTSP 的播放器（如 VLC），或联系管理员配置后端转码服务"
                />
                <div style={{ marginTop: 16 }}>
                  <p style={{ color: '#666', marginBottom: 8 }}>流地址：</p>
                  <Input.TextArea
                    value={previewCamera.stream_url}
                    readOnly
                    rows={2}
                    style={{ textAlign: 'left' }}
                  />
                </div>
                <Button
                  type="primary"
                  style={{ marginTop: 16 }}
                  onClick={() => window.open(previewCamera.stream_url, '_blank')}
                >
                  尝试在新窗口打开
                </Button>
              </>
            ) : null}
          </div>
        )}
      </Modal>
    </div>
  );
}
