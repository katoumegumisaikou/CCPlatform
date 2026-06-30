#!/usr/bin/env python3
"""生成 200 个中东地区光伏电站 mock 数据，每站 30-50 台机器人，直接导入 MySQL。"""
import random
import math
import subprocess
import sys
from datetime import datetime, timedelta

random.seed(42)

# =============================================================
# 中东各国分布配置 (国家代码, 名称, 数量, lat范围, lon范围)
# =============================================================
COUNTRIES = [
    ("SA", "Saudi Arabia",       60, (16.0, 31.0), (35.0, 55.0)),
    ("AE", "United Arab Emirates",30, (22.0, 26.0), (51.0, 57.0)),
    ("IR", "Iran",               25, (25.0, 39.0), (44.0, 63.0)),
    ("IQ", "Iraq",               20, (29.0, 37.0), (39.0, 49.0)),
    ("OM", "Oman",               15, (16.0, 26.5), (52.0, 60.0)),
    ("QA", "Qatar",              10, (24.5, 26.5), (50.5, 51.5)),
    ("KW", "Kuwait",             10, (28.5, 30.5), (46.5, 48.5)),
    ("JO", "Jordan",             10, (29.5, 33.5), (35.0, 39.5)),
    ("BH", "Bahrain",             5, (25.8, 26.3), (50.4, 50.7)),
    ("IL", "Israel",              5, (29.5, 33.5), (34.3, 35.9)),
    ("YE", "Yemen",               5, (12.5, 18.5), (43.0, 54.0)),
    ("LB", "Lebanon",             3, (33.0, 34.7), (35.0, 36.5)),
    ("SY", "Syria",               2, (32.5, 37.0), (36.0, 42.5)),
]

CITIES = {
    "SA": ["Riyadh", "Jeddah", "Mecca", "Medina", "Dammam", "Tabuk", "Hail", "Abha",
           "Buraydah", "Jubail", "Yanbu", "Al Kharj", "Al Ahsa", "Taif", "Najran",
           "Sakaka", "Arar", "Rafha", "Hafar Al-Batin", "Dhahran"],
    "AE": ["Abu Dhabi", "Dubai", "Sharjah", "Al Ain", "Ajman", "Ras Al Khaimah",
           "Fujairah", "Umm Al Quwain", "Madinat Zayed", "Al Dhafra"],
    "IR": ["Tehran", "Isfahan", "Shiraz", "Tabriz", "Mashhad", "Yazd", "Kerman",
           "Bandar Abbas", "Ahvaz", "Qom", "Kermanshah", "Zahedan"],
    "IQ": ["Baghdad", "Basra", "Mosul", "Erbil", "Najaf", "Karbala", "Kirkuk",
           "Sulaymaniyah", "Nasiriyah", "Ramadi"],
    "OM": ["Muscat", "Salalah", "Sohar", "Nizwa", "Ibri", "Sur", "Buraimi",
           "Duqm", "Khasab", "Adam"],
    "QA": ["Doha", "Al Rayyan", "Al Wakrah", "Al Khor", "Umm Salal", "Mesaieed",
           "Dukhan", "Al Shamal"],
    "KW": ["Kuwait City", "Al Jahra", "Al Ahmadi", "Hawalli", "Farwaniya",
           "Salmiya", "Fahaheel", "Mubarak Al-Kabeer"],
    "JO": ["Amman", "Zarqa", "Irbid", "Aqaba", "Madaba", "Ma'an", "Tafilah",
           "Karak", "Ajloun", "Jerash"],
    "BH": ["Manama", "Riffa", "Muharraq", "Hamad Town", "Sitra"],
    "IL": ["Tel Aviv", "Jerusalem", "Haifa", "Beersheba", "Eilat"],
    "YE": ["Sanaa", "Aden", "Taiz", "Hodeidah", "Mukalla"],
    "LB": ["Beirut", "Tripoli", "Sidon"],
    "SY": ["Damascus", "Aleppo"],
}

AREA_NAMES = ["Desert", "Industrial", "East", "West", "North", "South", "Central",
              "Al Shams", "Al Noor", "Al Dhafra", "Al Badiyah", "Al Rabigh",
              "Al Hasa", "Az Zawra", "Al Faya", "Al Mafraq", "As Salwa"]
PARK_SUFFIXES = ["Solar Park", "PV Plant", "Solar Farm", "Solar PV Station",
                 "Renewable Energy Park", "Solar Complex", "Photovoltaic Station"]

NOW = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
ROBOT_ID_START = 80001
STATION_ID_START = 10001
station_counter = [STATION_ID_START]
robot_counter = [ROBOT_ID_START]

def make_station_id():
    sid = f"STAT-{station_counter[0]}"
    station_counter[0] += 1
    return sid

def make_station_code(country_code, seq):
    return f"ST-{country_code}-{seq:05d}"

def make_robot_id():
    rid = f"ROB-{robot_counter[0]}"
    robot_counter[0] += 1
    return rid

def random_point(lat_range, lon_range):
    lat = random.uniform(*lat_range)
    lon = random.uniform(*lon_range)
    return round(lat, 7), round(lon, 7)

def escape_sql(s):
    return s.replace("\\", "\\\\").replace("'", "\\'")

def generate():
    total_stations = sum(c[2] for c in COUNTRIES)
    total_robots = 0
    station_seq = {}
    station_batch = []
    robot_batch = []

    print(f"-- Generating {total_stations} Middle East stations with 30-50 robots each...", file=sys.stderr)

    for country_code, country_name, count, lat_r, lon_r in COUNTRIES:
        station_seq.setdefault(country_code, 0)
        cities = CITIES[country_code]

        for i in range(count):
            seq = station_seq[country_code] + 1
            station_seq[country_code] = seq

            station_id = make_station_id()
            station_code = make_station_code(country_code, seq)
            city = random.choice(cities)
            area = random.choice(AREA_NAMES)
            suffix = random.choice(PARK_SUFFIXES)
            station_name = f"{city} {area} {suffix} {random.randint(1,20)}"
            lat, lon = random_point(lat_r, lon_r)

            # 容量: 1-100 MW, 大部分偏向小型
            capacity = round(random.expovariate(1/20) + 1, 2)
            capacity = max(1, min(100, capacity))
            panel_count = int(capacity * 1000 / 0.55)
            panel_area = round(panel_count * 2.5, 2)
            address = f"{city} {area}, {country_name}"

            station_batch.append(
                f"('{station_id}','{escape_sql(station_name)}','{station_code}',"
                f"'{escape_sql(address)}',{lon},{lat},{capacity},{panel_area},{panel_count},1)"
            )

            # --- 生成该电站的机器人 ---
            robot_count = random.randint(30, 50)
            total_robots += robot_count

            for r in range(robot_count):
                rid = make_robot_id()
                rc = f"RB-{robot_counter[0]-1:07d}"
                rname = f"Robot-{robot_counter[0]-1}"
                rtype = random.choices([1, 2, 3], weights=[40, 35, 25], k=1)[0]
                pos_x = round(random.uniform(0, 200), 3)
                pos_y = round(random.uniform(0, 150), 3)
                pos_z = round(random.uniform(0, 5), 3)
                heading = round(random.uniform(0, 360), 2)
                battery = random.randint(20, 100)
                online = random.choices([0, 1], weights=[8, 92], k=1)[0]
                work = random.choices([0, 1, 2, 3, 4], weights=[30, 45, 10, 10, 5], k=1)[0]
                speed = round(random.uniform(0, 2.0), 2)
                clean_area = round(random.uniform(0, 50), 2)
                fault = 0 if random.random() < 0.85 else random.randint(1, 99)
                temp = round(random.uniform(25, 55), 2)
                humidity = round(random.uniform(5, 40), 2)
                light = round(random.uniform(500, 1200), 2)
                wind = round(random.uniform(0, 15), 2)
                heartbeat = (datetime.now() - timedelta(minutes=random.randint(1, 60))).strftime("%Y-%m-%d %H:%M:%S")

                robot_batch.append(
                    f"('{rid}','{rc}','{rname}',{rtype},'{station_id}',"
                    f"{pos_x},{pos_y},{pos_z},{heading},{battery},{online},{work},"
                    f"{speed},{clean_area},{fault},{temp},{humidity},{light},{wind},"
                    f"'{heartbeat}','{NOW}','{NOW}')"
                )

            # 批量写入，避免内存爆炸
            if len(station_batch) >= 50:
                flush_stations(station_batch)
                station_batch = []
            if len(robot_batch) >= 200:
                flush_robots(robot_batch)
                robot_batch = []

    # 写入剩余
    if station_batch:
        flush_stations(station_batch)
    if robot_batch:
        flush_robots(robot_batch)

    print(f"Done: {total_stations} stations, {total_robots} robots generated.", file=sys.stderr)
    return total_stations, total_robots


def flush_stations(batch):
    if not batch:
        return
    sql = ("INSERT INTO stations (station_id, station_name, station_code, location, "
           "longitude, latitude, capacity, panel_area, panel_count, status) VALUES\n"
           + ",\n".join(batch) + ";\n")
    proc.stdin.write(sql)


def flush_robots(batch):
    if not batch:
        return
    sql = ("INSERT INTO robots (robot_id, robot_code, robot_name, robot_type, station_id, "
           "pos_x, pos_y, pos_z, heading, battery_level, online_status, work_status, "
           "speed, clean_area, fault_code, temperature, humidity, light_intensity, wind_speed, "
           "last_heartbeat, create_time, update_time) VALUES\n"
           + ",\n".join(batch) + ";\n")
    proc.stdin.write(sql)


# =============================================================
if __name__ == "__main__":
    proc = subprocess.Popen(
        ["mysql", "-u", "root", "ccplatform"],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
    )
    try:
        total_stations, total_robots = generate()
    finally:
        proc.stdin.close()
        proc.wait()
        stderr = proc.stderr.read()
        if stderr:
            print(f"MySQL errors:\n{stderr[:2000]}", file=sys.stderr)
        print(f"Inserted: {total_stations} stations, {total_robots} robots", file=sys.stderr)
