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
    gps_longitude DECIMAL(10,7) COMMENT '高精度GPS经度',
    gps_latitude DECIMAL(10,7) COMMENT '高精度GPS纬度',
    gps_altitude DECIMAL(10,3) COMMENT 'GPS高度(m)',
    gps_accuracy DECIMAL(6,3) COMMENT 'GPS定位精度(m)',
    battery_level TINYINT COMMENT '电量百分比',
    battery_voltage DECIMAL(6,2) COMMENT '主电池电压(V)',
    online_status TINYINT DEFAULT 0 COMMENT '在线状态:0离线1在线',
    work_status TINYINT DEFAULT 0 COMMENT '工作状态:0空闲1清扫2充电3故障4维护',
    fault_status TINYINT DEFAULT 0 COMMENT '故障状态:0正常1异常',
    work_period TINYINT DEFAULT 0 COMMENT '当前工作时段:0无1/2/3',
    signal_strength INT DEFAULT 0 COMMENT '综合信号强度',
    signal_4g_strength INT DEFAULT 0 COMMENT '4G信号强度',
    run_duration_seconds BIGINT DEFAULT 0 COMMENT '累计运行时长(秒)',
    speed DECIMAL(5,2) COMMENT '速度(m/s)',
    clean_area DECIMAL(10,2) COMMENT '清扫面积(㎡)',
    fault_code INT DEFAULT 0 COMMENT '故障码',
    cart_battery_voltage DECIMAL(6,2) COMMENT '清扫小车电池电压(V)',
    cart_low_voltage_status TINYINT DEFAULT 0 COMMENT '清扫小车低电压状态:0正常1低压',
    shuttle_battery_voltage DECIMAL(6,2) COMMENT '接驳车电池电压(V)',
    shuttle_low_voltage_status TINYINT DEFAULT 0 COMMENT '接驳车低电压状态:0正常1低压',
    shuttle_motor_current DECIMAL(6,2) COMMENT '接驳车电机电流(A)',
    cart_clean_motor_current DECIMAL(6,2) COMMENT '清扫小车清扫电机电流(A)',
    cart_travel_motor_current DECIMAL(6,2) COMMENT '清扫小车行进电机电流(A)',
    shuttle_power_percent TINYINT DEFAULT 0 COMMENT '接驳车功率选择(%)',
    shuttle_start_position TINYINT DEFAULT 0 COMMENT '接驳车起点位置:0否1是',
    shuttle_end_position TINYINT DEFAULT 0 COMMENT '接驳车终点位置:0否1是',
    shuttle_has_cleaner TINYINT DEFAULT 0 COMMENT '接驳车上是否有清扫车:0否1是',
    temperature DECIMAL(5,2) COMMENT '温度',
    humidity DECIMAL(5,2) COMMENT '湿度',
    light_intensity DECIMAL(10,2) COMMENT '光照强度',
    wind_speed DECIMAL(5,2) COMMENT '风速',
    device_time DATETIME COMMENT '设备当前时间',
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
    scope_type VARCHAR(20) DEFAULT 'global' COMMENT '作用域:global/station/robot',
    scope_id VARCHAR(32) COMMENT 'station_id或robot_id',
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
    schedule_config TEXT COMMENT '时段配置JSON',
    control_config TEXT COMMENT '控制配置JSON',
    threshold_config TEXT COMMENT '阈值配置JSON',
    server_config TEXT COMMENT '服务器配置JSON',
    communication_config TEXT COMMENT '通信参数JSON',
    is_default TINYINT DEFAULT 0 COMMENT '是否为默认模板',
    status TINYINT DEFAULT 1 COMMENT '状态:0禁用1启用',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_robot_type (robot_type)
) ENGINE=InnoDB COMMENT='设备配置模板表';

-- 机器人部件健康指数当前状态表
CREATE TABLE IF NOT EXISTS robot_component_health (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    robot_id VARCHAR(32) COMMENT '机器人ID',
    component VARCHAR(50) COMMENT '部件类型:drive_motor/brush_motor/battery/controller/sensor/transmission',
    health_score DECIMAL(5,2) COMMENT '健康指数0-100',
    health_level VARCHAR(20) COMMENT '健康等级:green/yellow/orange/red',
    degrade_rate DECIMAL(6,3) COMMENT '健康度日均下降速率(分/天)',
    metrics TEXT COMMENT '细分指标快照(JSON)',
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE INDEX idx_robot_component (robot_id, component)
) ENGINE=InnoDB COMMENT='机器人部件健康指数当前状态表';

-- 机器人部件健康指数历史快照表
CREATE TABLE IF NOT EXISTS robot_health_history (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    robot_id VARCHAR(32) COMMENT '机器人ID',
    component VARCHAR(50) COMMENT '部件类型',
    health_score DECIMAL(5,2) COMMENT '健康指数0-100',
    record_time DATETIME COMMENT '记录时间',
    INDEX idx_health_hist_robot_comp (robot_id, component),
    INDEX idx_record_time (record_time)
) ENGINE=InnoDB COMMENT='机器人部件健康指数历史快照表';

-- PHM传感器数据表(振动/温度/电气参数)
CREATE TABLE IF NOT EXISTS robot_sensor_data (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    robot_id VARCHAR(32) COMMENT '机器人ID',
    component VARCHAR(50) COMMENT '采集部件',
    vibration_rms DECIMAL(8,4) COMMENT '振动有效值(mm/s)',
    vibration_peak DECIMAL(8,4) COMMENT '振动峰值(mm/s)',
    crest_factor DECIMAL(6,3) COMMENT '峰值因子',
    dominant_freq_hz DECIMAL(8,2) COMMENT '振动信号主频(Hz)',
    bearing_fault_flag VARCHAR(20) COMMENT '轴承故障特征:none/outer_race/inner_race/ball',
    winding_temp DECIMAL(5,2) COMMENT '电机绕组温度(℃)',
    controller_temp DECIMAL(5,2) COMMENT '控制器散热片温度(℃)',
    insulation_resist_mohm DECIMAL(10,2) COMMENT '绝缘电阻(MΩ)',
    record_time DATETIME COMMENT '记录时间',
    INDEX idx_robot_id (robot_id),
    INDEX idx_component (component),
    INDEX idx_record_time (record_time)
) ENGINE=InnoDB COMMENT='PHM传感器数据表';

-- 故障模式库表
CREATE TABLE IF NOT EXISTS fault_patterns (
    pattern_code VARCHAR(50) PRIMARY KEY COMMENT '故障模式编码',
    component VARCHAR(50) COMMENT '所属部件',
    fault_name VARCHAR(100) COMMENT '故障名称',
    category VARCHAR(50) COMMENT '故障类别:motor/battery/sensor/comm/transmission/electrical',
    symptom_desc VARCHAR(500) COMMENT '典型症状描述',
    suggested_action VARCHAR(500) COMMENT '建议处置方案',
    recommended_parts VARCHAR(200) COMMENT '推荐备件(逗号分隔)',
    INDEX idx_component (component)
) ENGINE=InnoDB COMMENT='故障模式库表';

-- 故障预测结果表
CREATE TABLE IF NOT EXISTS fault_predictions (
    prediction_id VARCHAR(32) PRIMARY KEY,
    robot_id VARCHAR(32) COMMENT '机器人ID',
    station_id VARCHAR(32) COMMENT '电站ID',
    component VARCHAR(50) COMMENT '部件类型',
    pattern_code VARCHAR(50) COMMENT '故障模式编码',
    fault_name VARCHAR(100) COMMENT '故障名称',
    probability DECIMAL(5,4) COMMENT '故障概率0-1',
    risk_level VARCHAR(20) COMMENT '风险等级:high/medium/low',
    confidence DECIMAL(5,4) COMMENT '预测置信度0-1',
    predicted_start DATETIME COMMENT '预测故障窗口开始时间',
    predicted_end DATETIME COMMENT '预测故障窗口结束时间',
    suggested_action VARCHAR(500) COMMENT '建议处置方案',
    recommended_parts VARCHAR(200) COMMENT '推荐备件',
    status TINYINT DEFAULT 0 COMMENT '状态:0待处理1已生成工单2已忽略',
    maint_id VARCHAR(32) COMMENT '关联维护工单ID',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_robot_id (robot_id),
    INDEX idx_station_id (station_id),
    INDEX idx_status (status)
) ENGINE=InnoDB COMMENT='故障预测结果表';
