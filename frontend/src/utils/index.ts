// 机器人类型映射
export const ROBOT_TYPE_MAP: Record<number, string> = {
  1: '固定式',
  2: '接驳车',
  3: '全智能',
};

// 在线状态映射
export const ONLINE_STATUS_MAP: Record<number, { text: string; color: string }> = {
  0: { text: '离线', color: 'default' },
  1: { text: '在线', color: 'green' },
};

// 工作状态映射
export const WORK_STATUS_MAP: Record<number, { text: string; color: string }> = {
  0: { text: '空闲', color: 'default' },
  1: { text: '清扫中', color: 'blue' },
  2: { text: '充电中', color: 'orange' },
  3: { text: '故障', color: 'red' },
  4: { text: '维护', color: 'purple' },
};

// 任务状态映射
export const TASK_STATUS_MAP: Record<number, { text: string; color: string }> = {
  0: { text: '待执行', color: 'default' },
  1: { text: '执行中', color: 'processing' },
  2: { text: '已完成', color: 'success' },
  3: { text: '已暂停', color: 'warning' },
  4: { text: '已取消', color: 'default' },
  5: { text: '失败', color: 'error' },
};

// 任务类型映射
export const TASK_TYPE_MAP: Record<number, string> = {
  1: '即时',
  2: '定时',
  3: '周期',
};

// 告警级别映射
export const ALARM_LEVEL_MAP: Record<number, { text: string; color: string }> = {
  1: { text: '提示', color: 'blue' },
  2: { text: '一般', color: 'orange' },
  3: { text: '严重', color: 'red' },
  4: { text: '紧急', color: '#ff0000' },
};

// 告警处理状态
export const ALARM_HANDLE_MAP: Record<number, { text: string; color: string }> = {
  0: { text: '未处理', color: 'error' },
  1: { text: '已确认', color: 'warning' },
  2: { text: '处理中', color: 'processing' },
  3: { text: '已完成', color: 'success' },
  4: { text: '已忽略', color: 'default' },
};

// 获取 Token
export const getToken = (): string | null => localStorage.getItem('token');

// 设置 Token
export const setToken = (token: string): void => localStorage.setItem('token', token);

// 移除 Token
export const removeToken = (): void => localStorage.removeItem('token');

// 获取用户信息
export const getUser = () => {
  const raw = localStorage.getItem('user');
  if (!raw) return null;
  try {
    return JSON.parse(raw);
  } catch {
    return null;
  }
};

// 设置用户信息
export const setUser = (user: unknown): void => localStorage.setItem('user', JSON.stringify(user));

// 移除用户信息
export const removeUser = (): void => localStorage.removeItem('user');

// ===== 机器人类型字段可见性 =====

// 通用字段 — 所有机器人类型都显示
const ROBOT_COMMON_FIELDS = [
  'robot_id', 'robot_code', 'robot_name', 'robot_type', 'station_id', 'station_name',
  'online_status', 'work_status', 'battery_level', 'battery_voltage', 'fault_status', 'fault_code',
  'pos_x', 'pos_y', 'pos_z', 'heading',
  'gps_longitude', 'gps_latitude', 'gps_altitude', 'gps_accuracy',
  'speed', 'clean_area',
  'work_period', 'signal_strength', 'signal_4g_strength', 'run_duration_seconds',
  'temperature', 'humidity', 'light_intensity', 'wind_speed',
  'device_time', 'last_heartbeat', 'create_time', 'update_time',
];

// 接驳车特有字段 (type=2)
const ROBOT_SHUTTLE_FIELDS = [
  'shuttle_battery_voltage', 'shuttle_low_voltage_status', 'shuttle_motor_current',
  'shuttle_power_percent', 'shuttle_start_position', 'shuttle_end_position', 'shuttle_has_cleaner',
];

// 清扫小车特有字段 (type=2 接驳车的清扫小车, type=3 全智能自带清扫小车)
const ROBOT_CART_FIELDS = [
  'cart_battery_voltage', 'cart_low_voltage_status',
  'cart_clean_motor_current', 'cart_travel_motor_current',
];

/**
 * 每种机器人类型在表格中应显示的字段列表。
 * type=1 固定式: 仅通用字段
 * type=2 接驳车:   通用 + 接驳车 + 清扫小车
 * type=3 全智能:   通用 + 清扫小车
 */
export const ROBOT_TABLE_FIELDS: Record<number, string[]> = {
  1: [...ROBOT_COMMON_FIELDS],
  2: [...ROBOT_COMMON_FIELDS, ...ROBOT_SHUTTLE_FIELDS, ...ROBOT_CART_FIELDS],
  3: [...ROBOT_COMMON_FIELDS, ...ROBOT_CART_FIELDS],
};

/** 当未选择类型筛选时，默认只显示通用字段 */
export const ROBOT_DEFAULT_TABLE_FIELDS: string[] = [...ROBOT_COMMON_FIELDS];

/** 详情抽屉 section key 可见性 */
export const ROBOT_DETAIL_SECTIONS: Record<number, string[]> = {
  1: ['basic', 'status', 'position', 'environment'],
  2: ['basic', 'status', 'position', 'environment', 'shuttle_cart'],
  3: ['basic', 'status', 'position', 'environment', 'shuttle_cart'],
};
