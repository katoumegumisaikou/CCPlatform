-- 拓达威光伏清扫机器人云控平台 - 数据库初始化脚本

CREATE DATABASE IF NOT EXISTS ccplatform CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE ccplatform;

-- 电站表
CREATE TABLE IF NOT EXISTS stations (
    station_id VARCHAR(32) PRIMARY KEY,
    station_name VARCHAR(100) NOT NULL COMMENT '电站名称',
    station_code VARCHAR(50) UNIQUE COMMENT '电站编码',
    location VARCHAR(200) COMMENT '电站地址',
    longitude DECIMAL(10,7) COMMENT '经度',
    latitude DECIMAL(10,7) COMMENT '纬度',
    capacity DECIMAL(10,2) COMMENT '装机容量(MW)',
    panel_area DECIMAL(12,2) COMMENT '光伏板总面积(㎡)',
    panel_count INT COMMENT '光伏板数量',
    status TINYINT DEFAULT 1 COMMENT '状态:0停用1启用',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_station_code (station_code),
    INDEX idx_status (status)
) ENGINE=InnoDB COMMENT='光伏电站表';

-- 机器人表
CREATE TABLE IF NOT EXISTS robots (
    robot_id VARCHAR(32) PRIMARY KEY,
    robot_code VARCHAR(50) UNIQUE COMMENT '机器人编码',
    robot_name VARCHAR(100) COMMENT '机器人名称',
    robot_type TINYINT NOT NULL COMMENT '类型:1固定式2接驳车3全智能',
    station_id VARCHAR(32) COMMENT '所属电站ID',
    pos_x DECIMAL(10,3) COMMENT 'X坐标(m)',
    pos_y DECIMAL(10,3) COMMENT 'Y坐标(m)',
    pos_z DECIMAL(10,3) COMMENT 'Z坐标(m)',
    heading DECIMAL(5,2) COMMENT '朝向角度(度)',
    battery_level TINYINT COMMENT '电量百分比',
    online_status TINYINT DEFAULT 0 COMMENT '在线状态:0离线1在线',
    work_status TINYINT DEFAULT 0 COMMENT '工作状态:0空闲1清扫2充电3故障4维护',
    speed DECIMAL(5,2) COMMENT '速度(m/s)',
    clean_area DECIMAL(10,2) COMMENT '清扫面积(㎡)',
    fault_code INT DEFAULT 0 COMMENT '故障码',
    temperature DECIMAL(5,2) COMMENT '温度',
    humidity DECIMAL(5,2) COMMENT '湿度',
    light_intensity DECIMAL(10,2) COMMENT '光照强度',
    wind_speed DECIMAL(5,2) COMMENT '风速',
    last_heartbeat DATETIME COMMENT '最后心跳时间',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_station_id (station_id),
    INDEX idx_robot_type (robot_type),
    INDEX idx_online_status (online_status)
) ENGINE=InnoDB COMMENT='机器人表';

-- 清扫任务表
CREATE TABLE IF NOT EXISTS tasks (
    task_id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    task_name VARCHAR(100) COMMENT '任务名称',
    task_type TINYINT NOT NULL COMMENT '任务类型:1即时2定时3周期',
    robot_id VARCHAR(32) COMMENT '执行机器人ID',
    station_id VARCHAR(32) COMMENT '所属电站ID',
    area_ids VARCHAR(500) COMMENT '清扫区域ID列表',
    plan_start DATETIME COMMENT '计划开始时间',
    plan_end DATETIME COMMENT '计划结束时间',
    actual_start DATETIME COMMENT '实际开始时间',
    actual_end DATETIME COMMENT '实际结束时间',
    clean_area DECIMAL(10,2) COMMENT '清扫面积(㎡)',
    task_status TINYINT DEFAULT 0 COMMENT '状态:0待执行1执行中2完成3暂停4取消5失败',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_station_id (station_id),
    INDEX idx_robot_id (robot_id),
    INDEX idx_task_status (task_status)
) ENGINE=InnoDB COMMENT='清扫任务表';

-- 告警表
CREATE TABLE IF NOT EXISTS alarms (
    alarm_id VARCHAR(32) PRIMARY KEY,
    alarm_level TINYINT NOT NULL COMMENT '告警级别:1提示2一般3严重4紧急',
    alarm_type VARCHAR(50) COMMENT '告警类型',
    robot_id VARCHAR(32) COMMENT '关联机器人ID',
    station_id VARCHAR(32) COMMENT '关联电站ID',
    alarm_content VARCHAR(500) COMMENT '告警内容',
    alarm_time DATETIME COMMENT '告警产生时间',
    handle_status TINYINT DEFAULT 0 COMMENT '处理状态:0未处理1已确认2已处理3已忽略',
    handle_time DATETIME COMMENT '处理时间',
    handler_id VARCHAR(32) COMMENT '处理人ID',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_station_id (station_id),
    INDEX idx_robot_id (robot_id),
    INDEX idx_alarm_level (alarm_level),
    INDEX idx_handle_status (handle_status),
    INDEX idx_alarm_time (alarm_time)
) ENGINE=InnoDB COMMENT='告警表';

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    user_id VARCHAR(32) PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL COMMENT '用户名',
    password_hash VARCHAR(100) NOT NULL COMMENT '密码哈希',
    real_name VARCHAR(50) COMMENT '真实姓名',
    phone VARCHAR(20) COMMENT '手机号',
    email VARCHAR(100) COMMENT '邮箱',
    role_id VARCHAR(32) COMMENT '角色ID',
    station_ids VARCHAR(500) COMMENT '授权电站ID列表',
    status TINYINT DEFAULT 1 COMMENT '状态:0禁用1启用',
    last_login DATETIME COMMENT '最后登录时间',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_role_id (role_id)
) ENGINE=InnoDB COMMENT='用户表';

-- 角色表
CREATE TABLE IF NOT EXISTS roles (
    role_id VARCHAR(32) PRIMARY KEY,
    role_name VARCHAR(50) UNIQUE NOT NULL COMMENT '角色名称',
    description VARCHAR(200) COMMENT '角色描述',
    permissions TEXT COMMENT '权限列表(JSON)',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB COMMENT='角色表';

-- 插入默认角色
INSERT INTO roles (role_id, role_name, description, permissions) VALUES
('sys_admin', '系统管理员', '拥有系统全部权限', '["*"]'),
('station_mgr', '电站管理员', '负责电站管理', '["station:*","robot:*","task:*","alarm:*","monitor:view","analytics:view"]'),
('engineer', '运维工程师', '设备维护和故障处理', '["robot:view","robot:control","task:view","alarm:view","alarm:handle","monitor:view"]'),
('operator', '监控值班员', '日常监控和告警处理', '["monitor:view","alarm:view","alarm:handle","task:view"]'),
('viewer', '普通用户', '仅查看权限', '["monitor:view","analytics:view"]')
ON DUPLICATE KEY UPDATE role_name = VALUES(role_name);

-- 插入默认管理员(密码: admin123)
INSERT INTO users (user_id, username, password_hash, real_name, role_id, status) VALUES
('admin01', 'admin', '$2b$10$iZB4HLEY0uaS4bOmB0bsVe.Qz./gS6Jm8t0csVzMMLNkJDIROoCsK', '系统管理员', 'sys_admin', 1)
ON DUPLICATE KEY UPDATE username = VALUES(username);
