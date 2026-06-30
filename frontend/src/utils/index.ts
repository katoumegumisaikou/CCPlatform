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
