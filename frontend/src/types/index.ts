// ===== 通用类型 =====

export interface ApiResponse<T = unknown> {
  code: number;
  data: T;
  message: string;
}

export interface PageData<T> {
  list: T[];
  total: number;
}

export interface PageResponse<T> {
  code: number;
  data: PageData<T>;
  message: string;
}

export interface Pagination {
  page: number;
  size: number;
}

// ===== 用户 & 认证 =====

export interface User {
  user_id: number;
  username: string;
  real_name: string;
  role_id: number;
  org_id: number;
  phone: string;
  email: string;
  status: number;
  role?: Role;
}

export interface Role {
  role_id: number;
  role_name: string;
  role_code: string;
  permissions: string;
  status: number;
}

export interface LoginRequest {
  username: string;
  password: string;
  captcha_id?: string;
  captcha_code?: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}

// ===== 电站 =====

export interface Station {
  station_id: string;
  station_name: string;
  station_code: string;
  longitude: number;
  latitude: number;
  capacity: number;
  location: string;
  panel_area: number;
  panel_count: number;
  org_id: number;
  status: number;
  create_time: string;
}

// ===== 机器人 =====

export interface Robot {
  robot_id: string;
  robot_code: string;
  robot_name: string;
  robot_type: number;
  station_id: string;
  station_name?: string;
  online_status: number;
  work_status: number;
  battery_level: number;
  pos_x: number;
  pos_y: number;
  pos_z: number;
  heading: number;
  speed: number;
  clean_area: number;
  temperature: number;
  humidity: number;
  light_intensity: number;
  wind_speed: number;
  last_heartbeat: string;
  create_time: string;
}

// ===== 任务 =====

export interface Task {
  task_id: number;
  task_name: string;
  task_type: number;
  robot_id: string;
  station_id: string;
  area_ids: string;
  plan_start: string | null;
  plan_end: string | null;
  actual_start: string | null;
  actual_end: string | null;
  clean_area: number;
  task_status: number;
  cron_expr: string;
  create_time: string;
}

export interface TaskProgress {
  task_id: number;
  task_name: string;
  task_status: number;
  clean_area: number;
  progress: number;
  estimated_remaining_seconds: number;
}

// ===== 告警 =====

export interface Alarm {
  alarm_id: string;
  alarm_level: number;
  alarm_type: string;
  robot_id: string;
  station_id: string;
  alarm_content: string;
  alarm_time: string;
  handle_status: number;
  handler: string;
  handle_time: string | null;
  handle_remark: string;
}

export interface AlarmRule {
  rule_id: number;
  rule_name: string;
  alarm_type: string;
  scope_type: string;
  scope_id: string;
  suppression_rule: string;
  escalation_rule: string;
  status: number;
}

// ===== 监控 =====

export interface DashboardOverview {
  station_count: number;
  robot_stats: Record<string, number>;
  today_clean_area: number;
  running_tasks: number;
  unhandled_alarms: number;
  alarm_stats: Record<string, number>;
}

export interface StationRealtime {
  station_id: string;
  station_name: string;
  longitude: number;
  latitude: number;
  robots: RobotRealtime[];
}

export interface RobotRealtime {
  robot_id: string;
  robot_name: string;
  robot_type: number;
  online_status: number;
  work_status: number;
  battery_level: number;
  pos_x: number;
  pos_y: number;
  pos_z: number;
  heading: number;
  speed: number;
  clean_area: number;
  temperature: number;
  humidity: number;
  light_intensity: number;
  wind_speed: number;
}

export interface RobotPosition {
  robot_id: string;
  pos_x: number;
  pos_y: number;
  pos_z: number;
  heading: number;
  timestamp: string;
}

// ===== 摄像头 =====

export interface Camera {
  camera_id: string;
  camera_name: string;
  station_id: string;
  stream_url: string;
  camera_type: string;
  position: string;
  status: number;
}

// ===== 固件 =====

export interface Firmware {
  firmware_id: string;
  version: string;
  file_name: string;
  file_size: number;
  robot_type: number;
  description: string;
  status: number;
  create_time: string;
}

// ===== 维护 =====

export interface Maintenance {
  id: number;
  robot_id: string;
  station_id: string;
  maintenance_type: string;
  description: string;
  plan_date: string;
  complete_date: string | null;
  status: number;
}

// ===== 组织 =====

export interface Organization {
  org_id: number;
  org_name: string;
  parent_id: number;
  org_code: string;
  status: number;
  children?: Organization[];
}

// ===== 系统配置 =====

export interface SystemConfig {
  config_key: string;
  config_value: string;
  category: string;
  description: string;
}

// ===== 字典 =====

export interface DictType {
  id: number;
  dict_type: string;
  dict_name: string;
  status: number;
}

export interface DictItem {
  id: number;
  dict_type: string;
  dict_key: string;
  dict_value: string;
  sort_order: number;
  status: number;
}

// ===== WebSocket 消息 =====

export type WsMessageType = 'heartbeat' | 'position' | 'status' | 'alarm' | 'clean';

export interface WsMessage {
  type: WsMessageType;
  data: Record<string, unknown>;
}

// ===== 通知模板 =====

export interface NotifyTemplate {
  tpl_id: string;
  tpl_name: string;
  tpl_type: string;
  alarm_level: number;
  content: string;
  status: number;
}

// ===== 报告模板 =====

export interface ReportTemplate {
  tpl_id: string;
  tpl_name: string;
  tpl_type: string;
  dimensions: string;
  metrics: string;
  status: number;
}

// ===== 数据分析 =====

export interface AnalyticsQuery {
  station_id?: string;
  start_date?: string;
  end_date?: string;
  robot_type?: number;
}
