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
  user_id: string;
  username: string;
  real_name: string;
  role_id: string;
  station_ids?: string;
  phone: string;
  email: string;
  status: number;
  role?: Role;
}

export interface Role {
  role_id: string;
  role_name: string;
  description: string;
  permissions: string;
  role_code?: string;
  status?: number;
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
  gps_longitude?: number;
  gps_latitude?: number;
  gps_altitude?: number;
  gps_accuracy?: number;
  speed: number;
  clean_area: number;
  fault_code?: number;
  battery_voltage?: number;
  fault_status?: number;
  work_period?: number;
  signal_strength?: number;
  signal_4g_strength?: number;
  run_duration_seconds?: number;
  cart_battery_voltage?: number;
  cart_low_voltage_status?: number;
  shuttle_battery_voltage?: number;
  shuttle_low_voltage_status?: number;
  shuttle_motor_current?: number;
  cart_clean_motor_current?: number;
  cart_travel_motor_current?: number;
  shuttle_power_percent?: number;
  shuttle_start_position?: number;
  shuttle_end_position?: number;
  shuttle_has_cleaner?: number;
  temperature: number;
  humidity: number;
  light_intensity: number;
  wind_speed: number;
  device_time?: string | null;
  last_heartbeat: string | null;
  create_time: string;
  update_time?: string;
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
  gps_longitude?: number;
  gps_latitude?: number;
  gps_altitude?: number;
  gps_accuracy?: number;
  speed: number;
  clean_area: number;
  battery_voltage?: number;
  fault_status?: number;
  work_period?: number;
  signal_strength?: number;
  signal_4g_strength?: number;
  run_duration_seconds?: number;
  cart_battery_voltage?: number;
  cart_low_voltage_status?: number;
  shuttle_battery_voltage?: number;
  shuttle_low_voltage_status?: number;
  shuttle_motor_current?: number;
  cart_clean_motor_current?: number;
  cart_travel_motor_current?: number;
  shuttle_power_percent?: number;
  shuttle_start_position?: number;
  shuttle_end_position?: number;
  shuttle_has_cleaner?: number;
  temperature: number;
  humidity: number;
  light_intensity: number;
  wind_speed: number;
  device_time?: string | null;
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
  org_id: string;
  org_name: string;
  parent_id: string;
  org_code: string;
  org_type?: string;
  station_ids?: string;
  status: number;
  children?: Organization[];
}

// ===== WebSocket 消息 =====

export type WsMessageType = 'heartbeat' | 'position' | 'status' | 'alarm' | 'clean' | 'geofence_alarm';

export interface WsMessage {
  type: WsMessageType;
  data: Record<string, unknown>;
}

// ===== 数据分析 =====

export interface AnalyticsQuery {
  station_id?: string;
  start_date?: string;
  end_date?: string;
  robot_type?: number;
}

// ===== 电子围栏 =====

export interface Geofence {
  geofence_id: string;
  fence_name: string;
  fence_type: number;
  action_type: number;
  scope_type: string;
  scope_id: string;
  center_lng: number;
  center_lat: number;
  radius: number;
  points: string;
  alarm_level: number;
  description: string;
  color: string;
  status: number;
  create_time: string;
  update_time: string;
}

export interface GeofenceAlarm {
  alarm_id: string;
  fence_id: string;
  fence_name: string;
  robot_id: string;
  robot_name: string;
  station_id: string;
  trigger_type: number;
  position_lng: number;
  position_lat: number;
  alarm_content: string;
  alarm_time: string;
  handle_status: number;
  handle_time: string | null;
  handle_remark: string;
  create_time: string;
}

// ===== 气象联动 =====

export interface WeatherNow {
  location_id: string;
  station_id: string;
  obs_time: string;
  temp: string;
  feels_like: string;
  icon: string;
  text: string;
  wind_dir: string;
  wind_scale: string;
  wind_speed: string;
  humidity: string;
  precip: string;
  pressure: string;
  vis: string;
  cloud: string;
  update_time: string;
}

export interface WeatherDaily {
  location_id: string;
  fx_date: string;
  station_id: string;
  sunrise: string;
  sunset: string;
  temp_max: string;
  temp_min: string;
  icon_day: string;
  text_day: string;
  icon_night: string;
  text_night: string;
  wind_dir_day: string;
  wind_scale_day: string;
  humidity: string;
  precip: string;
  uv_index: string;
  update_time: string;
}

export interface WeatherHourly {
  location_id: string;
  fx_time: string;
  station_id: string;
  temp: string;
  icon: string;
  text: string;
  wind_dir: string;
  wind_scale: string;
  wind_speed: string;
  humidity: string;
  precip: string;
  pressure: string;
  cloud: string;
  update_time: string;
}

export interface WeatherWarning {
  location_id: string;
  warning_id: string;
  station_id: string;
  sender: string;
  pub_time: string;
  title: string;
  start_time: string;
  end_time: string;
  status: string;
  level: string;
  type: string;
  type_name: string;
  text: string;
  update_time: string;
}

export interface WeatherAirQuality {
  location_id: string;
  station_id: string;
  pub_time: string;
  aqi: string;
  level: string;
  category: string;
  primary: string;
  pm10: string;
  pm2p5: string;
  no2: string;
  so2: string;
  co: string;
  o3: string;
  update_time: string;
}

export interface CleaningDecisionRule {
  rule_id: string;
  rule_name: string;
  station_id: string;
  conditions: string;
  dust_rule: string;
  priority: number;
  status: number;
  remark: string;
  create_time: string;
  update_time: string;
}

export interface ExtremeWeatherRule {
  rule_id: string;
  rule_name: string;
  station_id: string;
  weather_type: string;
  trigger_cond: string;
  action: string;
  recovery_cond: string;
  status: number;
  remark: string;
  create_time: string;
  update_time: string;
}

export interface CleaningAdvice {
  suggested_action: string;
  reason: string;
  dust_score: number;
  weather_score: number;
  triggered_warnings: WeatherWarning[];
  recommended_time: string;
}

export interface CleaningScheduleAdjustment {
  id: number;
  station_id: string;
  current_freq: string;
  suggested_freq: string;
  reason: string;
  analysis_data: string;
  status: number;
  create_time: string;
}

// ===== PHM / 智能预测 =====

export type PhmComponent =
  | 'drive_motor'
  | 'brush_motor'
  | 'battery'
  | 'controller'
  | 'sensor'
  | 'transmission';

export interface HealthComponent {
  id: number;
  robot_id: string;
  component: PhmComponent;
  health_score: number;
  health_level: 'green' | 'yellow' | 'orange' | 'red';
  degrade_rate: number;
  metrics: string;
  update_time: string;
  create_time: string;
}

export interface HealthOverview {
  robot_id: string;
  robot_name?: string;
  overall_score: number;
  overall_level: 'green' | 'yellow' | 'orange' | 'red';
  components: HealthComponent[];
}

export interface HealthTrendPoint {
  id: number;
  robot_id: string;
  component: PhmComponent;
  health_score: number;
  record_time: string;
}

export interface HealthTrend {
  robot_id: string;
  component: PhmComponent;
  points: HealthTrendPoint[];
  prev_period_avg?: number;
  curr_period_avg?: number;
  change_rate?: number;
  trend_note?: string;
}

export interface SparePartForecastItem {
  part_name: string;
  expected_qty: number;
  related_fault_count: number;
}

export interface MaintenanceWindowRecommendation {
  robot_id: string;
  recommended_start: string;
  recommended_end: string;
  reason: string;
}
