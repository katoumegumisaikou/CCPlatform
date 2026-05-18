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

-- 固件表
CREATE TABLE IF NOT EXISTS firmwares (
    firmware_id VARCHAR(32) PRIMARY KEY,
    version VARCHAR(50) NOT NULL COMMENT '固件版本号',
    package_path VARCHAR(500) COMMENT '升级包路径',
    compatible_robot_types VARCHAR(200) COMMENT '兼容机器人类型，逗号分隔',
    upgrade_log TEXT COMMENT '升级日志(JSON)',
    status TINYINT DEFAULT 1 COMMENT '状态:0禁用1启用',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_version (version)
) ENGINE=InnoDB COMMENT='固件版本表';

-- 维护记录表
CREATE TABLE IF NOT EXISTS maintenance (
    maint_id VARCHAR(32) PRIMARY KEY,
    robot_id VARCHAR(32) COMMENT '关联机器人ID',
    station_id VARCHAR(32) COMMENT '关联电站ID',
    maint_type VARCHAR(50) COMMENT '维护类型:routine/repair/overhaul',
    handler VARCHAR(50) COMMENT '处理人',
    fault_desc VARCHAR(500) COMMENT '故障描述',
    parts_replaced VARCHAR(500) COMMENT '更换配件',
    maint_time DATETIME COMMENT '维护时间',
    next_maint_time DATETIME COMMENT '下次维护时间',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_robot_id (robot_id),
    INDEX idx_station_id (station_id),
    INDEX idx_maint_time (maint_time)
) ENGINE=InnoDB COMMENT='维护记录表';

-- 告警规则表
CREATE TABLE IF NOT EXISTS alarm_rules (
    rule_id VARCHAR(32) PRIMARY KEY,
    rule_name VARCHAR(100) NOT NULL COMMENT '规则名称',
    alarm_type VARCHAR(50) COMMENT '适用的告警类型',
    threshold_value TEXT COMMENT '阈值配置(JSON)',
    escalation_rule TEXT COMMENT '升级规则(JSON)',
    suppression_rule TEXT COMMENT '抑制规则(JSON)',
    status TINYINT DEFAULT 1 COMMENT '状态:0禁用1启用',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_alarm_type (alarm_type)
) ENGINE=InnoDB COMMENT='告警规则表';

-- 通知模板表
CREATE TABLE IF NOT EXISTS notify_templates (
    tpl_id VARCHAR(32) PRIMARY KEY,
    tpl_name VARCHAR(100) NOT NULL COMMENT '模板名称',
    tpl_type VARCHAR(20) COMMENT '渠道类型:sms/email/app/webhook',
    alarm_level TINYINT COMMENT '适用告警级别，0表示全部',
    channel_config TEXT COMMENT '渠道配置(JSON)',
    content_template TEXT COMMENT '内容模板',
    status TINYINT DEFAULT 1 COMMENT '状态:0禁用1启用',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_tpl_type (tpl_type)
) ENGINE=InnoDB COMMENT='通知模板表';

-- 操作日志表
CREATE TABLE IF NOT EXISTS audit_logs (
    log_id VARCHAR(32) PRIMARY KEY,
    operator_id VARCHAR(32) COMMENT '操作人ID',
    operator_name VARCHAR(50) COMMENT '操作人姓名',
    operation_type VARCHAR(50) COMMENT '操作类型:create/update/delete/login',
    target_resource VARCHAR(200) COMMENT '目标资源路径',
    detail VARCHAR(1000) COMMENT '操作详情',
    ip VARCHAR(50) COMMENT '操作IP',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_operator_id (operator_id),
    INDEX idx_operation_type (operation_type),
    INDEX idx_create_time (create_time)
) ENGINE=InnoDB COMMENT='操作日志表';

-- 登录日志表
CREATE TABLE IF NOT EXISTS login_logs (
    log_id VARCHAR(32) PRIMARY KEY,
    user_id VARCHAR(32) COMMENT '用户ID',
    username VARCHAR(50) COMMENT '用户名',
    login_ip VARCHAR(50) COMMENT '登录IP',
    login_time DATETIME COMMENT '登录时间',
    login_result TINYINT COMMENT '登录结果:1成功0失败',
    fail_reason VARCHAR(200) COMMENT '失败原因',
    INDEX idx_user_id (user_id),
    INDEX idx_login_time (login_time)
) ENGINE=InnoDB COMMENT='登录日志表';

-- 系统配置表
CREATE TABLE IF NOT EXISTS system_configs (
    config_key VARCHAR(100) PRIMARY KEY,
    config_value TEXT COMMENT '配置值',
    category VARCHAR(50) COMMENT '配置分类:notify/ota/backup/video/general',
    description VARCHAR(200) COMMENT '配置描述',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_category (category)
) ENGINE=InnoDB COMMENT='系统配置表';

-- 数据字典表
CREATE TABLE IF NOT EXISTS dicts (
    dict_id VARCHAR(32) PRIMARY KEY,
    dict_type VARCHAR(50) UNIQUE COMMENT '字典类型编码',
    dict_name VARCHAR(100) NOT NULL COMMENT '字典类型名称',
    description VARCHAR(200) COMMENT '描述',
    status TINYINT DEFAULT 1 COMMENT '状态:0禁用1启用',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE INDEX idx_dict_type (dict_type)
) ENGINE=InnoDB COMMENT='数据字典表';

-- 数据字典项表
CREATE TABLE IF NOT EXISTS dict_items (
    item_id VARCHAR(32) PRIMARY KEY,
    dict_type VARCHAR(50) COMMENT '所属字典类型',
    item_key VARCHAR(50) COMMENT '字典项键',
    item_value VARCHAR(200) COMMENT '字典项值',
    sort_order INT DEFAULT 0 COMMENT '排序',
    status TINYINT DEFAULT 1 COMMENT '状态:0禁用1启用',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_dict_type (dict_type)
) ENGINE=InnoDB COMMENT='数据字典项表';

-- 组织架构表
CREATE TABLE IF NOT EXISTS organizations (
    org_id VARCHAR(32) PRIMARY KEY,
    org_name VARCHAR(100) NOT NULL COMMENT '组织名称',
    org_code VARCHAR(50) UNIQUE COMMENT '组织编码',
    org_type VARCHAR(20) COMMENT '组织类型:company/region/station_group',
    parent_id VARCHAR(32) COMMENT '父组织ID',
    station_ids VARCHAR(500) COMMENT '关联电站ID列表',
    sort_order INT DEFAULT 0 COMMENT '排序',
    status TINYINT DEFAULT 1 COMMENT '状态:0禁用1启用',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_parent_id (parent_id),
    UNIQUE INDEX idx_org_code (org_code)
) ENGINE=InnoDB COMMENT='组织架构表';

-- 机器人位置历史表
CREATE TABLE IF NOT EXISTS robot_positions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    robot_id VARCHAR(32) COMMENT '机器人ID',
    pos_x DECIMAL(10,3) COMMENT 'X坐标(m)',
    pos_y DECIMAL(10,3) COMMENT 'Y坐标(m)',
    pos_z DECIMAL(10,3) COMMENT 'Z坐标(m)',
    heading DECIMAL(5,2) COMMENT '朝向角度(度)',
    timestamp DATETIME COMMENT '上报时间',
    INDEX idx_robot_id (robot_id),
    INDEX idx_timestamp (timestamp)
) ENGINE=InnoDB COMMENT='机器人位置历史表';

-- 环境数据历史表
CREATE TABLE IF NOT EXISTS environment_data (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    robot_id VARCHAR(32) COMMENT '机器人ID',
    temperature DECIMAL(5,2) COMMENT '温度(℃)',
    humidity DECIMAL(5,2) COMMENT '湿度(%)',
    light_intensity DECIMAL(10,2) COMMENT '光照强度(lux)',
    wind_speed DECIMAL(5,2) COMMENT '风速(m/s)',
    record_time DATETIME COMMENT '记录时间',
    INDEX idx_robot_id (robot_id),
    INDEX idx_record_time (record_time)
) ENGINE=InnoDB COMMENT='环境数据历史表';

-- 摄像头表
CREATE TABLE IF NOT EXISTS cameras (
    camera_id VARCHAR(32) PRIMARY KEY,
    camera_name VARCHAR(100) NOT NULL COMMENT '摄像头名称',
    station_id VARCHAR(32) COMMENT '所属电站ID',
    stream_url VARCHAR(500) COMMENT '视频流地址',
    camera_type VARCHAR(20) COMMENT '类型:rtsp/hls',
    position VARCHAR(200) COMMENT '安装位置描述',
    status TINYINT DEFAULT 1 COMMENT '状态:0离线1在线',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_station_id (station_id)
) ENGINE=InnoDB COMMENT='摄像头表';

-- 清扫记录表
CREATE TABLE IF NOT EXISTS cleaning_records (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    task_id BIGINT UNSIGNED COMMENT '关联任务ID',
    robot_id VARCHAR(32) COMMENT '机器人ID',
    station_id VARCHAR(32) COMMENT '电站ID',
    pos_x DECIMAL(10,3) COMMENT '当前位置X(m)',
    pos_y DECIMAL(10,3) COMMENT '当前位置Y(m)',
    clean_area DECIMAL(10,2) COMMENT '累计清扫面积(㎡)',
    record_time DATETIME COMMENT '记录时间',
    INDEX idx_robot_id (robot_id),
    INDEX idx_station_id (station_id),
    INDEX idx_record_time (record_time),
    INDEX idx_task_id (task_id)
) ENGINE=InnoDB COMMENT='清扫记录表';

-- 报表模板表
CREATE TABLE IF NOT EXISTS report_templates (
    tpl_id VARCHAR(32) PRIMARY KEY,
    tpl_name VARCHAR(100) NOT NULL COMMENT '模板名称',
    tpl_type VARCHAR(20) COMMENT '模板类型:efficiency/economic/custom',
    dimensions TEXT COMMENT '数据维度(JSON)',
    metrics TEXT COMMENT '指标配置(JSON)',
    status TINYINT DEFAULT 1 COMMENT '状态:0禁用1启用',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB COMMENT='报表模板表';

-- 设备认证凭据表
CREATE TABLE IF NOT EXISTS device_credentials (
    credential_id VARCHAR(32) PRIMARY KEY,
    device_id VARCHAR(32) UNIQUE COMMENT '设备ID(机器人/摄像头等)',
    device_type VARCHAR(20) COMMENT '设备类型:robot/camera/sensor',
    username VARCHAR(50) COMMENT 'MQTT用户名',
    password_hash VARCHAR(100) COMMENT '密码哈希',
    token VARCHAR(200) COMMENT '认证Token',
    status TINYINT DEFAULT 1 COMMENT '状态:0禁用1启用',
    last_connected DATETIME COMMENT '最后连接时间',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE INDEX idx_device_id (device_id)
) ENGINE=InnoDB COMMENT='设备认证凭据表';

-- OTA升级记录表
CREATE TABLE IF NOT EXISTS upgrade_records (
    record_id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    firmware_id VARCHAR(32) COMMENT '固件ID',
    robot_id VARCHAR(32) COMMENT '机器人ID',
    from_version VARCHAR(50) COMMENT '升级前版本',
    to_version VARCHAR(50) COMMENT '升级后版本',
    status TINYINT DEFAULT 0 COMMENT '状态:0下发1下载中2安装中3成功4失败',
    error_msg VARCHAR(500) COMMENT '错误信息',
    dispatch_time DATETIME COMMENT '下发时间',
    complete_time DATETIME COMMENT '完成时间',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_firmware_id (firmware_id),
    INDEX idx_robot_id (robot_id)
) ENGINE=InnoDB COMMENT='OTA升级记录表';

-- 密码重置表
CREATE TABLE IF NOT EXISTS password_resets (
    reset_id VARCHAR(32) PRIMARY KEY,
    user_id VARCHAR(32) COMMENT '用户ID',
    token VARCHAR(200) UNIQUE COMMENT '重置Token',
    expires_at DATETIME COMMENT '过期时间',
    used TINYINT DEFAULT 0 COMMENT '是否已使用',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE INDEX idx_token (token),
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB COMMENT='密码重置表';

-- 设备配置表
CREATE TABLE IF NOT EXISTS robot_configs (
    config_id VARCHAR(32) PRIMARY KEY,
    config_name VARCHAR(100) NOT NULL COMMENT '配置名称',
    robot_type TINYINT COMMENT '适用机器人类型，0表示通用',
    config_data TEXT COMMENT '配置参数(JSON)',
    is_default TINYINT DEFAULT 0 COMMENT '是否为默认模板',
    status TINYINT DEFAULT 1 COMMENT '状态:0禁用1启用',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_robot_type (robot_type)
) ENGINE=InnoDB COMMENT='设备配置模板表';
