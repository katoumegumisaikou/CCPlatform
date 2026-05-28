#!/usr/bin/env python3
"""生成 35 个电站、2000 台机器人、50 个任务的 mock 数据 SQL"""

import random
import math
from datetime import datetime, timedelta

random.seed(42)

STATIONS = [
    # 新疆 15 个
    ("新疆哈密光伏电站", 42.8, 93.5),
    ("新疆吐鲁番光伏基地", 42.9, 89.2),
    ("新疆乌鲁木齐光伏电站", 43.8, 87.6),
    ("新疆昌吉光伏电站", 44.0, 87.3),
    ("新疆石河子光伏基地", 44.3, 86.0),
    ("新疆克拉玛依光伏电站", 45.6, 84.9),
    ("新疆伊犁光伏基地", 43.9, 81.3),
    ("新疆巴州光伏电站", 41.7, 86.1),
    ("新疆阿克苏光伏基地", 41.2, 80.3),
    ("新疆喀什光伏电站", 39.5, 76.0),
    ("新疆和田光伏基地", 37.1, 79.9),
    ("新疆塔城光伏电站", 46.7, 82.9),
    ("新疆阿勒泰光伏基地", 47.8, 88.1),
    ("新疆博州光伏电站", 44.9, 82.1),
    ("新疆克州光伏基地", 39.7, 76.2),
    # 甘肃 10 个
    ("甘肃敦煌光伏基地", 40.1, 94.7),
    ("甘肃酒泉光伏电站", 39.7, 98.5),
    ("甘肃嘉峪关光伏基地", 39.8, 98.3),
    ("甘肃张掖光伏电站", 38.9, 100.5),
    ("甘肃金昌光伏基地", 38.5, 102.2),
    ("甘肃武威光伏电站", 37.9, 102.6),
    ("甘肃兰州光伏基地", 36.1, 103.8),
    ("甘肃白银光伏电站", 36.5, 104.2),
    ("甘肃定西光伏基地", 35.6, 104.6),
    ("甘肃庆阳光伏电站", 35.7, 107.6),
    # 青海 10 个
    ("青海格尔木光伏基地", 36.4, 94.9),
    ("青海德令哈光伏电站", 37.4, 97.3),
    ("青海海南州光伏基地", 36.0, 100.6),
    ("青海海西州光伏电站", 37.3, 97.1),
    ("青海西宁光伏基地", 36.6, 101.8),
    ("青海海东光伏电站", 36.5, 102.1),
    ("青海黄南州光伏基地", 35.5, 102.0),
    ("青海果洛州光伏电站", 34.5, 100.2),
    ("青海玉树光伏基地", 33.0, 97.0),
    ("青海海北州光伏电站", 36.9, 100.4),
]

NOW = datetime.now()

def rand_capacity():
    return round(random.uniform(50, 500), 2)

def rand_panel_count():
    return random.randint(5000, 50000)

def rand_battery():
    return random.randint(55, 100)

def rand_work_status():
    return random.choices([1, 0, 2, 3, 4], weights=[60, 15, 10, 10, 5], k=1)[0]

def rand_temp():
    return round(random.uniform(15.0, 45.0), 2)

def rand_humidity():
    return round(random.uniform(10.0, 65.0), 2)

def rand_light():
    return round(random.uniform(30000, 110000), 2)

def rand_wind():
    return round(random.uniform(0.5, 8.5), 2)

def rand_speed():
    return round(random.uniform(0.3, 1.5), 2)

def rand_clean_area():
    return round(random.uniform(100, 5000), 2)


lines = []
lines.append("-- ====== 清理数据 ======")
lines.append("SET FOREIGN_KEY_CHECKS = 0;")
# 清理应用实际使用的表
for t in ["cleaning_records", "robot_positions", "tasks", "robots", "stations",
          "alarms", "alarm_rules", "audit_logs", "login_logs", "maintenance", "upgrade_records",
          "device_credentials", "environment_data",
          "t_alarm", "t_alarm_event", "t_alarm_rule", "t_alarm_suppression",
          "t_clean_area", "t_clean_area_coverage", "t_clean_record",
          "t_clean_task", "t_clean_task_area", "t_clean_task_event",
          "t_clean_task_robot", "t_device_heartbeat", "t_dict_item", "t_dict_type",
          "t_login_log", "t_map_element", "t_mqtt_message_log", "t_operation_log",
          "t_permission", "t_report_snapshot",
          "t_robot", "t_robot_comm_switch_record", "t_robot_command",
          "t_robot_command_ack", "t_robot_config_history", "t_robot_daily_stat",
          "t_robot_position_history", "t_robot_position_latest",
          "t_robot_track_point", "t_role_permission",
          "t_station", "t_station_daily_stat", "t_station_map",
          "t_station_weather", "t_system_config",
          "t_user", "t_user_role", "t_user_station",
          "cameras", "firmwares"]:
    lines.append(f"TRUNCATE TABLE {t};")
lines.append("SET FOREIGN_KEY_CHECKS = 1;")

# ====== 电站 ======
lines.append("\n-- ====== 电站 (35个) ======")
station_ids = []
for i, (name, lat, lng) in enumerate(STATIONS):
    sid = f"ST{i+1:04d}"
    station_ids.append(sid)
    code = f"XJ{i+1:03d}" if "新疆" in name else (f"GS{i+1-15:03d}" if "甘肃" in name else f"QH{i+1-25:03d}")
    if "甘肃" in name:
        code = f"GS{i+1-15:03d}"
    elif "青海" in name:
        code = f"QH{i+1-25:03d}"
    cap = rand_capacity()
    pc = rand_panel_count()
    pa = round(pc * 2.1, 2)
    loc = name.replace("光伏电站", "").replace("光伏基地", "")
    lines.append(
        f"INSERT INTO stations (station_id, station_name, station_code, location, longitude, latitude, "
        f"capacity, panel_area, panel_count, status) VALUES "
        f"('{sid}', '{name}', '{code}', '{loc}', {lng:.6f}, {lat:.6f}, {cap}, {pa}, {pc}, 1);"
    )

# ====== 机器人 ======
lines.append("\n-- ====== 机器人 (2000台: 1200固定式 + 600接驳车 + 200全智能) ======")
robot_ids_by_station = {sid: [] for sid in station_ids}

robot_idx = 0

def gen_robots(count, rtype, prefix, name_prefix):
    global robot_idx
    stmts = []
    for i in range(count):
        rid = f"{prefix}{i+1:04d}"
        rcode = f"RC-{prefix}{i+1:04d}"
        rname = f"{name_prefix}{i+1:04d}"
        sid = station_ids[robot_idx % len(station_ids)]
        robot_ids_by_station[sid].append(rid)
        robot_idx += 1

        battery = rand_battery()
        ws = rand_work_status()
        online = 1 if random.random() < 0.88 else 0
        speed = rand_speed() if ws == 1 else 0
        ca = rand_clean_area()
        temp = rand_temp()
        hum = rand_humidity()
        light = rand_light()
        wind = rand_wind()
        pos_x = round(random.uniform(0, 120), 3)
        pos_y = round(random.uniform(0, 90), 3)
        heading = round(random.uniform(0, 360), 2)
        hb = NOW - timedelta(minutes=random.randint(1, 180))

        stmts.append(
            f"INSERT INTO robots (robot_id, robot_code, robot_name, robot_type, station_id, "
            f"pos_x, pos_y, pos_z, heading, battery_level, online_status, work_status, speed, "
            f"clean_area, fault_code, temperature, humidity, light_intensity, wind_speed, "
            f"last_heartbeat) VALUES "
            f"('{rid}', '{rcode}', '{rname}', {rtype}, '{sid}', "
            f"{pos_x}, {pos_y}, 0.000, {heading}, {battery}, {online}, {ws}, {speed}, "
            f"{ca}, 0, {temp}, {hum}, {light}, {wind}, "
            f"'{hb.strftime('%Y-%m-%d %H:%M:%S')}');"
        )
    return stmts

lines.extend(gen_robots(1200, 1, "R1-", "固定式"))
lines.extend(gen_robots(600, 2, "R2-", "接驳车"))
lines.extend(gen_robots(200, 3, "R3-", "全智能"))

# ====== 任务 ======
lines.append("\n-- ====== 任务 (50个) ======")
statuses = [0, 1, 2, 3, 4, 5]
status_weights = [10, 15, 40, 5, 10, 20]
types = [1, 2, 3]
type_weights = [30, 30, 40]

for t in range(50):
    tname = f"清扫任务-{t+1:03d}"
    ttype = random.choices(types, type_weights, k=1)[0]
    ws = random.choices(statuses, status_weights, k=1)[0]
    sid = random.choice(station_ids)
    rid = random.choice(robot_ids_by_station[sid]) if robot_ids_by_station[sid] else None
    rid_val = f"'{rid}'" if rid else "NULL"

    plan_start = NOW - timedelta(days=random.randint(0, 14), hours=random.randint(0, 23))
    plan_end = plan_start + timedelta(hours=random.randint(2, 8))
    actual_start = f"'{plan_start.strftime('%Y-%m-%d %H:%M:%S')}'" if ws >= 1 else "NULL"
    actual_end = f"'{(plan_start + timedelta(hours=random.randint(1, 6))).strftime('%Y-%m-%d %H:%M:%S')}'" if ws in [2, 5] else "NULL"
    ca = round(random.uniform(50, 3000), 2) if ws == 2 else 0

    lines.append(
        f"INSERT INTO tasks (task_name, task_type, robot_id, station_id, area_ids, "
        f"plan_start, plan_end, actual_start, actual_end, clean_area, task_status) VALUES "
        f"('{tname}', {ttype}, {rid_val}, '{sid}', 'area1,area2', "
        f"'{plan_start.strftime('%Y-%m-%d %H:%M:%S')}', "
        f"'{plan_end.strftime('%Y-%m-%d %H:%M:%S')}', "
        f"{actual_start}, {actual_end}, {ca}, {ws});"
    )

# ====== 告警规则 ======
lines.append("\n-- ====== 告警规则 (5条) ======")
alarm_rules = [
    (
        "AR-DEMO-001",
        "电机过流升级规则",
        "MOTOR_OVERLOAD",
        "global",
        "",
        '{"current_amp":{"operator":">","value":18,"unit":"A"}}',
        '[{"count":2,"within_min":15,"to_level":4}]',
        '{"within_sec":300}',
    ),
    (
        "AR-DEMO-002",
        "通讯中断抑制规则",
        "COMMUNICATION_LOST",
        "station",
        "ST0001",
        '{"offline_sec":{"operator":">","value":300,"unit":"s"}}',
        '[{"count":3,"within_min":30,"to_level":3}]',
        '{"within_sec":600}',
    ),
    (
        "AR-DEMO-003",
        "传感器离线规则",
        "SENSOR_OFFLINE",
        "global",
        "",
        '{"missing_position_sec":{"operator":">","value":60,"unit":"s"}}',
        '[{"count":2,"within_min":20,"to_level":3}]',
        '{"within_sec":300}',
    ),
    (
        "AR-DEMO-004",
        "高温保护规则",
        "HIGH_TEMPERATURE",
        "robot",
        "R1-0176",
        '{"temperature":{"operator":">","value":40,"unit":"C"}}',
        '[{"count":2,"within_min":10,"to_level":4}]',
        '{"within_sec":180}',
    ),
    (
        "AR-DEMO-005",
        "OTA升级失败规则",
        "OTA_FAILED",
        "global",
        "",
        '{"retry_count":{"operator":">=","value":3,"unit":"times"}}',
        '[{"count":1,"within_min":60,"to_level":3}]',
        '{"within_sec":900}',
    ),
]

for rule in alarm_rules:
    rid, name, alarm_type, scope_type, scope_id, threshold, escalation, suppression = rule
    lines.append(
        "INSERT INTO alarm_rules (rule_id, rule_name, alarm_type, scope_type, scope_id, "
        "threshold_value, escalation_rule, suppression_rule, status, create_time, update_time) VALUES "
        f"('{rid}', '{name}', '{alarm_type}', '{scope_type}', '{scope_id}', "
        f"'{threshold}', '{escalation}', '{suppression}', 1, NOW(), NOW()) "
        "ON DUPLICATE KEY UPDATE rule_name = VALUES(rule_name), alarm_type = VALUES(alarm_type), "
        "scope_type = VALUES(scope_type), scope_id = VALUES(scope_id), "
        "threshold_value = VALUES(threshold_value), escalation_rule = VALUES(escalation_rule), "
        "suppression_rule = VALUES(suppression_rule), status = VALUES(status), update_time = NOW();"
    )

# ====== 告警记录 ======
lines.append("\n-- ====== 告警记录 (8条) ======")
alarms = [
    ("ALM-DEMO-001", 4, "MOTOR_OVERLOAD", "R1-0176", "ST0001", "固定式0176电机电流持续过高，已触发过流保护", "MINUTE", 15, 0, "NULL"),
    ("ALM-DEMO-002", 3, "SENSOR_OFFLINE", "R3-0196", "ST0001", "全智能0196定位传感器离线，站内坐标更新中断", "MINUTE", 42, 0, "NULL"),
    ("ALM-DEMO-003", 2, "COMMUNICATION_LOST", "R1-0002", "ST0002", "固定式0002心跳超过5分钟未恢复，请检查通讯链路", "HOUR", 2, 1, "NULL"),
    ("ALM-DEMO-004", 3, "HIGH_TEMPERATURE", "R1-0176", "ST0001", "固定式0176机体温度达到42度，建议暂停清扫降温", "HOUR", 3, 2, "NULL"),
    ("ALM-DEMO-005", 2, "BRUSH_STUCK", "R2-0236", "ST0001", "接驳车0236清扫刷阻力异常，疑似有异物卡滞", "HOUR", 5, 0, "NULL"),
    ("ALM-DEMO-006", 1, "LOW_WATER_LEVEL", "R2-0271", "ST0001", "接驳车0271水箱余量偏低，请计划补水", "HOUR", 8, 3, "NOW() - INTERVAL 7 HOUR"),
    ("ALM-DEMO-007", 3, "OTA_FAILED", "R3-0198", "ST0003", "全智能0198固件升级安装失败，已回滚到上一版本", "DAY", 1, 0, "NULL"),
    ("ALM-DEMO-008", 2, "HIGH_WIND", "R2-0516", "ST0001", "现场风速偏高，建议暂停高处清扫作业", "DAY", 2, 4, "NOW() - INTERVAL 1 DAY"),
]

for alarm in alarms:
    aid, level, alarm_type, robot_id, station_id, content, interval_unit, interval_value, status, handle_time = alarm
    handler_id = "NULL" if status == 0 else "'admin01'"
    lines.append(
        "INSERT INTO alarms (alarm_id, alarm_level, alarm_type, robot_id, station_id, "
        "alarm_content, alarm_time, handle_status, handle_time, handler_id, create_time) VALUES "
        f"('{aid}', {level}, '{alarm_type}', '{robot_id}', '{station_id}', "
        f"'{content}', NOW() - INTERVAL {interval_value} {interval_unit}, {status}, {handle_time}, "
        f"{handler_id}, NOW()) "
        "ON DUPLICATE KEY UPDATE alarm_level = VALUES(alarm_level), alarm_type = VALUES(alarm_type), "
        "robot_id = VALUES(robot_id), station_id = VALUES(station_id), "
        "alarm_content = VALUES(alarm_content), alarm_time = VALUES(alarm_time), "
        "handle_status = VALUES(handle_status), handle_time = VALUES(handle_time), "
        "handler_id = VALUES(handler_id);"
    )

sql = "\n".join(lines)

with open("/home/megumi/gocodehouse/CCPlatform-v1/scripts/mock_data.sql", "w") as f:
    f.write(sql)

print(f"Generated {len(lines)} lines of SQL")
print(f"Stations: {len(station_ids)}")
print(f"Robots: {robot_idx}")
print(f"Tasks: 50")
