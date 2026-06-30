-- 拓达威光伏清扫机器人云控平台 - 全球光伏电站 Mock 数据
-- 共 100 个电站，覆盖全球主要光伏市场

USE ccplatform;

-- ============================================================
-- 中国 (25个): STAT-001 ~ STAT-025
-- ============================================================
INSERT INTO stations (station_id, station_name, station_code, location, longitude, latitude, capacity, panel_area, panel_count, status) VALUES
('STAT-001', '青海塔拉滩光伏电站', 'ST-CN-001', '青海省海南藏族自治州共和县塔拉滩', 100.5823000, 36.1245000, 2000.00, 9000000.00, 3636363, 1),
('STAT-002', '新疆哈密光伏产业园', 'ST-CN-002', '新疆维吾尔自治区哈密市伊州区光伏产业园', 93.5402000, 42.8218000, 1000.00, 4500000.00, 1818181, 1),
('STAT-003', '甘肃敦煌光伏发电基地', 'ST-CN-003', '甘肃省酒泉市敦煌市七里镇', 94.7364000, 40.1245000, 800.00, 3600000.00, 1454545, 1),
('STAT-004', '宁夏石嘴山光伏电站', 'ST-CN-004', '宁夏回族自治区石嘴山市惠农区', 106.4183000, 39.0325000, 640.00, 2880000.00, 1163636, 1),
('STAT-005', '内蒙古鄂尔多斯光伏领跑者基地', 'ST-CN-005', '内蒙古自治区鄂尔多斯市达拉特旗', 109.8567000, 39.6215000, 500.00, 2250000.00, 909090, 1),
('STAT-006', '西藏日喀则光伏电站', 'ST-CN-006', '西藏自治区日喀则市桑珠孜区', 88.9123000, 29.2785000, 300.00, 1350000.00, 545454, 1),
('STAT-007', '青海格尔木东出口光伏电站', 'ST-CN-007', '青海省海西蒙古族藏族自治州格尔木市东出口', 94.9352000, 36.4123000, 600.00, 2700000.00, 1090909, 1),
('STAT-008', '新疆阿克苏光伏电站', 'ST-CN-008', '新疆维吾尔自治区阿克苏地区阿克苏市', 80.3214000, 41.1867000, 400.00, 1800000.00, 727272, 1),
('STAT-009', '甘肃酒泉金塔光伏电站', 'ST-CN-009', '甘肃省酒泉市金塔县红柳洼', 98.5231000, 39.7458000, 350.00, 1575000.00, 636363, 1),
('STAT-010', '河北张家口光伏电站', 'ST-CN-010', '河北省张家口市张北县', 114.9236000, 40.8152000, 200.00, 900000.00, 363636, 1),
('STAT-011', '陕西榆林光伏电站', 'ST-CN-011', '陕西省榆林市榆阳区小纪汗镇', 109.7423000, 38.3145000, 150.00, 675000.00, 272727, 1),
('STAT-012', '山西大同光伏领跑者基地', 'ST-CN-012', '山西省大同市左云县', 113.3152000, 40.0812000, 180.00, 810000.00, 327272, 1),
('STAT-013', '山东东营河口光伏电站', 'ST-CN-013', '山东省东营市河口区', 118.7423000, 37.5245000, 120.00, 540000.00, 218181, 1),
('STAT-014', '江苏盐城滨海光伏电站', 'ST-CN-014', '江苏省盐城市滨海县滨海港经济区', 120.1235000, 33.4125000, 100.00, 450000.00, 181818, 1),
('STAT-015', '安徽淮南水上光伏电站', 'ST-CN-015', '安徽省淮南市潘集区采煤沉陷区', 117.0218000, 32.6324000, 80.00, 360000.00, 145454, 1),
('STAT-016', '江西上饶光伏电站', 'ST-CN-016', '江西省上饶市鄱阳县', 117.8523000, 28.5136000, 60.00, 270000.00, 109090, 1),
('STAT-017', '云南楚雄元谋光伏电站', 'ST-CN-017', '云南省楚雄彝族自治州元谋县', 101.5234000, 25.0835000, 50.00, 225000.00, 90909, 1),
('STAT-018', '贵州毕节威宁光伏电站', 'ST-CN-018', '贵州省毕节市威宁彝族回族苗族自治县', 105.3145000, 27.2832000, 40.00, 180000.00, 72727, 1),
('STAT-019', '四川凉山会东光伏电站', 'ST-CN-019', '四川省凉山彝族自治州会东县', 102.3125000, 27.9124000, 35.00, 157500.00, 63636, 1),
('STAT-020', '海南三亚崖州光伏电站', 'ST-CN-020', '海南省三亚市崖州区', 109.5123000, 18.3125000, 30.00, 135000.00, 54545, 1),
('STAT-021', '广东湛江徐闻光伏电站', 'ST-CN-021', '广东省湛江市徐闻县', 110.4128000, 21.2836000, 25.00, 112500.00, 45454, 1),
('STAT-022', '福建漳州东山光伏电站', 'ST-CN-022', '福建省漳州市东山县', 117.7215000, 24.5128000, 20.00, 90000.00, 36363, 1),
('STAT-023', '浙江嘉兴秀洲光伏电站', 'ST-CN-023', '浙江省嘉兴市秀洲区光伏小镇', 120.8123000, 30.7825000, 15.00, 67500.00, 27272, 1),
('STAT-024', '辽宁朝阳建平光伏电站', 'ST-CN-024', '辽宁省朝阳市建平县', 120.5312000, 41.6134000, 10.00, 45000.00, 18181, 1),
('STAT-025', '吉林白城光伏电站', 'ST-CN-025', '吉林省白城市洮北区', 122.8214000, 45.6128000, 5.00, 22500.00, 9090, 1);

-- ============================================================
-- 印度 (12个): STAT-026 ~ STAT-037
-- ============================================================
INSERT INTO stations (station_id, station_name, station_code, location, longitude, latitude, capacity, panel_area, panel_count, status) VALUES
('STAT-026', 'Bhadla Solar Park', 'ST-IN-001', 'Bhadla, Jodhpur District, Rajasthan, India', 71.9215000, 27.5348000, 2245.00, 10102500.00, 4081818, 1),
('STAT-027', 'Pavagada Solar Park', 'ST-IN-002', 'Pavagada, Tumkur District, Karnataka, India', 77.2845000, 14.2342000, 2050.00, 9225000.00, 3727272, 1),
('STAT-028', 'Kurnool Ultra Mega Solar Park', 'ST-IN-003', 'Kurnool, Andhra Pradesh, India', 78.0532000, 15.8125000, 1000.00, 4500000.00, 1818181, 1),
('STAT-029', 'Rewa Ultra Mega Solar Park', 'ST-IN-004', 'Rewa, Madhya Pradesh, India', 81.3245000, 24.5158000, 750.00, 3375000.00, 1363636, 1),
('STAT-030', 'Charanka Solar Park', 'ST-IN-005', 'Charanka, Patan District, Gujarat, India', 71.2123000, 23.9128000, 600.00, 2700000.00, 1090909, 1),
('STAT-031', 'Kamuthi Solar Power Project', 'ST-IN-006', 'Kamuthi, Ramanathapuram District, Tamil Nadu, India', 78.4125000, 9.4253000, 648.00, 2916000.00, 1178181, 1),
('STAT-032', 'Ananthapuramu Solar Park', 'ST-IN-007', 'Anantapur, Andhra Pradesh, India', 77.6342000, 14.7125000, 500.00, 2250000.00, 909090, 1),
('STAT-033', 'Galiveedu Solar Park', 'ST-IN-008', 'Galiveedu, Annamayya District, Andhra Pradesh, India', 78.5312000, 14.0124000, 400.00, 1800000.00, 727272, 1),
('STAT-034', 'Mandsaur Solar Park', 'ST-IN-009', 'Mandsaur, Madhya Pradesh, India', 75.1245000, 24.0823000, 250.00, 1125000.00, 454545, 1),
('STAT-035', 'Bikaner Solar Park', 'ST-IN-010', 'Bikaner, Rajasthan, India', 73.3124000, 28.0215000, 150.00, 675000.00, 272727, 1),
('STAT-036', 'Dholera Solar Park', 'ST-IN-011', 'Dholera, Ahmedabad District, Gujarat, India', 72.2314000, 22.2128000, 100.00, 450000.00, 181818, 1),
('STAT-037', 'Jodhpur PV Solar Plant', 'ST-IN-012', 'Jodhpur, Rajasthan, India', 73.0245000, 26.3125000, 50.00, 225000.00, 90909, 1);

-- ============================================================
-- 美国 (12个): STAT-038 ~ STAT-049
-- ============================================================
INSERT INTO stations (station_id, station_name, station_code, location, longitude, latitude, capacity, panel_area, panel_count, status) VALUES
('STAT-038', 'Solar Star I & II', 'ST-US-001', 'Rosamond, Kern County, California, USA', -118.4123000, 34.8234000, 579.00, 2605500.00, 1052727, 1),
('STAT-039', 'Topaz Solar Farm', 'ST-US-002', 'San Luis Obispo County, California, USA', -120.1234000, 35.4123000, 550.00, 2475000.00, 1000000, 1),
('STAT-040', 'Ivanpah Solar Electric Generating System', 'ST-US-003', 'Ivanpah Valley, San Bernardino County, California, USA', -115.5234000, 35.6128000, 392.00, 1764000.00, 712727, 1),
('STAT-041', 'Agua Caliente Solar Project', 'ST-US-004', 'Yuma County, Arizona, USA', -112.9124000, 32.9235000, 290.00, 1305000.00, 527272, 1),
('STAT-042', 'Desert Sunlight Solar Farm', 'ST-US-005', 'Desert Center, Riverside County, California, USA', -115.4123000, 33.8345000, 550.00, 2475000.00, 1000000, 1),
('STAT-043', 'Copper Mountain Solar Facility', 'ST-US-006', 'Boulder City, Clark County, Nevada, USA', -114.9238000, 35.6123000, 802.00, 3609000.00, 1458181, 1),
('STAT-044', 'Mount Signal Solar', 'ST-US-007', 'Calexico, Imperial County, California, USA', -115.6123000, 32.7125000, 460.00, 2070000.00, 836363, 1),
('STAT-045', 'Lone Valley Solar Park', 'ST-US-008', 'Kern County, California, USA', -118.1423000, 34.8125000, 300.00, 1350000.00, 545454, 1),
('STAT-046', 'Rosamond Solar Project', 'ST-US-009', 'Rosamond, Kern County, California, USA', -118.4123000, 34.9125000, 350.00, 1575000.00, 636363, 1),
('STAT-047', 'Roadrunner Solar Project', 'ST-US-010', 'El Paso County, Texas, USA', -106.4123000, 31.9124000, 497.00, 2236500.00, 903636, 1),
('STAT-048', 'Permian Energy Center', 'ST-US-011', 'Andrews County, Texas, USA', -102.5234000, 31.5123000, 420.00, 1890000.00, 763636, 1),
('STAT-049', 'Samson Solar Energy Center', 'ST-US-012', 'Lamar County, Texas, USA', -95.5124000, 33.5234000, 250.00, 1125000.00, 454545, 1);

-- ============================================================
-- 中东 (10个): STAT-050 ~ STAT-059
-- ============================================================
INSERT INTO stations (station_id, station_name, station_code, location, longitude, latitude, capacity, panel_area, panel_count, status) VALUES
('STAT-050', 'Mohammed bin Rashid Al Maktoum Solar Park', 'ST-AE-001', 'Seih Al-Dahal, Dubai, United Arab Emirates', 55.3245000, 24.8234000, 1627.00, 7321500.00, 2958181, 1),
('STAT-051', 'Noor Abu Dhabi Solar Plant', 'ST-AE-002', 'Sweihan, Abu Dhabi, United Arab Emirates', 54.6234000, 24.3125000, 1177.00, 5296500.00, 2140000, 1),
('STAT-052', 'Al Dhafra Solar PV', 'ST-AE-003', 'Al Dhafra, Abu Dhabi, United Arab Emirates', 54.5123000, 24.1825000, 2000.00, 9000000.00, 3636363, 1),
('STAT-053', 'Sudair Solar PV Plant', 'ST-SA-001', 'Sudair Industrial City, Riyadh Province, Saudi Arabia', 47.0124000, 24.5234000, 1500.00, 6750000.00, 2727272, 1),
('STAT-054', 'Sakaka Solar PV Plant', 'ST-SA-002', 'Sakaka, Al Jouf Province, Saudi Arabia', 40.2134000, 29.9128000, 300.00, 1350000.00, 545454, 1),
('STAT-055', 'Shuaiba Solar PV Project', 'ST-SA-003', 'Shuaiba, Makkah Province, Saudi Arabia', 39.1245000, 21.7125000, 600.00, 2700000.00, 1090909, 1),
('STAT-056', 'Ibri II Solar PV Plant', 'ST-OM-001', 'Ibri, Ad Dhahirah Governorate, Oman', 56.5123000, 23.2345000, 500.00, 2250000.00, 909090, 1),
('STAT-057', 'Al Kharsaah Solar PV Plant', 'ST-QA-001', 'Al Kharsaah, Al Rayyan, Qatar', 51.5234000, 25.3125000, 800.00, 3600000.00, 1454545, 1),
('STAT-058', 'Ashalim Solar Thermal Power Station', 'ST-IL-001', 'Ashalim, Southern District, Israel', 34.7125000, 30.9234000, 250.00, 1125000.00, 454545, 1),
('STAT-059', 'Shagaya Renewable Energy Park', 'ST-KW-001', 'Shagaya, Al Jahra Governorate, Kuwait', 47.5124000, 29.3125000, 50.00, 225000.00, 90909, 1);

-- ============================================================
-- 欧洲 (10个): STAT-060 ~ STAT-069
-- ============================================================
INSERT INTO stations (station_id, station_name, station_code, location, longitude, latitude, capacity, panel_area, panel_count, status) VALUES
('STAT-060', 'Núñez de Balboa Solar Plant', 'ST-ES-001', 'Usagre, Badajoz, Extremadura, Spain', -6.5123000, 38.5234000, 500.00, 2250000.00, 909090, 1),
('STAT-061', 'Francisco Pizarro Solar Plant', 'ST-ES-002', 'Torrecillas de la Tiesa, Cáceres, Extremadura, Spain', -5.8124000, 39.2345000, 590.00, 2655000.00, 1072727, 1),
('STAT-062', 'Cestas Solar Park', 'ST-FR-001', 'Cestas, Gironde, Nouvelle-Aquitaine, France', -0.8123000, 44.7234000, 300.00, 1350000.00, 545454, 1),
('STAT-063', 'Templin Solar Park', 'ST-DE-001', 'Templin, Brandenburg, Germany', 13.5124000, 53.1234000, 128.00, 576000.00, 232727, 1),
('STAT-064', 'Montalto di Castro Solar Plant', 'ST-IT-001', 'Montalto di Castro, Lazio, Italy', 14.5132000, 41.5234000, 103.00, 463500.00, 187272, 1),
('STAT-065', 'Kozani Solar Park', 'ST-GR-001', 'Kozani, Western Macedonia, Greece', 21.8123000, 40.3125000, 204.00, 918000.00, 370909, 1),
('STAT-066', 'Witznitz Solar Park', 'ST-DE-002', 'Witznitz, Saxony, Germany', 12.5134000, 51.2345000, 650.00, 2925000.00, 1181818, 1),
('STAT-067', 'Horta Solar Park', 'ST-PT-001', 'Horta, Faro, Algarve, Portugal', -7.5123000, 37.2345000, 80.00, 360000.00, 145454, 1),
('STAT-068', 'Talayuela Solar Plant', 'ST-ES-003', 'Talayuela, Cáceres, Extremadura, Spain', -5.6123000, 39.9125000, 300.00, 1350000.00, 545454, 1),
('STAT-069', 'Solara4 Solar Plant', 'ST-PT-002', 'Ourique, Beja, Alentejo, Portugal', -8.0124000, 37.8123000, 220.00, 990000.00, 400000, 1);

-- ============================================================
-- 非洲 (8个): STAT-070 ~ STAT-077
-- ============================================================
INSERT INTO stations (station_id, station_name, station_code, location, longitude, latitude, capacity, panel_area, panel_count, status) VALUES
('STAT-070', 'Noor Ouarzazate Solar Complex', 'ST-MA-001', 'Ouarzazate, Drâa-Tafilalet, Morocco', -6.9123000, 31.0234000, 580.00, 2610000.00, 1054545, 1),
('STAT-071', 'Benban Solar Park', 'ST-EG-001', 'Benban, Aswan Governorate, Egypt', 32.9124000, 24.5234000, 1650.00, 7425000.00, 3000000, 1),
('STAT-072', 'De Aar Solar Power Plant', 'ST-ZA-001', 'De Aar, Northern Cape, South Africa', 24.0125000, -30.7234000, 175.00, 787500.00, 318181, 1),
('STAT-073', 'Jasper Solar Power Plant', 'ST-ZA-002', 'Postmasburg, Northern Cape, South Africa', 22.5124000, -28.5128000, 96.00, 432000.00, 174545, 1),
('STAT-074', 'Garissa Solar Power Plant', 'ST-KE-001', 'Garissa, Garissa County, Kenya', 39.6234000, -0.5124000, 55.00, 247500.00, 100000, 1),
('STAT-075', 'Scaling Solar Plant', 'ST-ZM-001', 'Lusaka South Multi-Facility Economic Zone, Zambia', 28.3125000, -15.4234000, 54.00, 243000.00, 98181, 1),
('STAT-076', 'Bokhol Solar Plant', 'ST-SN-001', 'Bokhol, Saint-Louis Region, Senegal', -15.5124000, 16.5234000, 30.00, 135000.00, 54545, 1),
('STAT-077', 'Zagtouli Solar Power Station', 'ST-BF-001', 'Zagtouli, Ouagadougou, Burkina Faso', -1.5124000, 12.4234000, 33.00, 148500.00, 60000, 1);

-- ============================================================
-- 南美及墨西哥 (8个): STAT-078 ~ STAT-085
-- ============================================================
INSERT INTO stations (station_id, station_name, station_code, location, longitude, latitude, capacity, panel_area, panel_count, status) VALUES
('STAT-078', 'Villanueva Solar Park', 'ST-MX-001', 'Villanueva, Coahuila, Mexico', -103.0123000, 25.6234000, 828.00, 3726000.00, 1505454, 1),
('STAT-079', 'Cerro Dominador Solar Plant', 'ST-CL-001', 'María Elena, Antofagasta Region, Chile', -69.8123000, -24.2345000, 250.00, 1125000.00, 454545, 1),
('STAT-080', 'Luz del Norte Solar Plant', 'ST-CL-002', 'Copiapó, Atacama Region, Chile', -70.0124000, -27.0125000, 141.00, 634500.00, 256363, 1),
('STAT-081', 'São Gonçalo Solar Park', 'ST-BR-001', 'São Gonçalo do Gurguéia, Piauí, Brazil', -43.0125000, -9.0123000, 608.00, 2736000.00, 1105454, 1),
('STAT-082', 'Janaúba Solar Complex', 'ST-BR-002', 'Janaúba, Minas Gerais, Brazil', -43.3124000, -15.8234000, 500.00, 2250000.00, 909090, 1),
('STAT-083', 'Cauchari Solar Plant', 'ST-AR-001', 'Cauchari, Jujuy Province, Argentina', -66.5124000, -23.5345000, 300.00, 1350000.00, 545454, 1),
('STAT-084', 'San Juan Solar Park', 'ST-AR-002', 'San Juan, San Juan Province, Argentina', -68.5234000, -31.5124000, 100.00, 450000.00, 181818, 1),
('STAT-085', 'Potrero Solar Park', 'ST-MX-002', 'Lagos de Moreno, Jalisco, Mexico', -101.0124000, 21.5345000, 150.00, 675000.00, 272727, 1);

-- ============================================================
-- 亚太其他 (10个): STAT-086 ~ STAT-095
-- ============================================================
INSERT INTO stations (station_id, station_name, station_code, location, longitude, latitude, capacity, panel_area, panel_count, status) VALUES
('STAT-086', 'Setouchi Kirei Solar Power Plant', 'ST-JP-001', 'Setouchi, Okayama Prefecture, Japan', 134.2124000, 34.7234000, 235.00, 1057500.00, 427272, 1),
('STAT-087', 'Kagoshima Nanatsujima Solar Plant', 'ST-JP-002', 'Kagoshima, Kagoshima Prefecture, Japan', 130.6234000, 31.5124000, 70.00, 315000.00, 127272, 1),
('STAT-088', 'Nyngan Solar Plant', 'ST-AU-001', 'Nyngan, New South Wales, Australia', 147.2123000, -31.6234000, 102.00, 459000.00, 185454, 1),
('STAT-089', 'Broken Hill Solar Plant', 'ST-AU-002', 'Broken Hill, New South Wales, Australia', 141.5124000, -31.9234000, 53.00, 238500.00, 96363, 1),
('STAT-090', 'Sunraysia Solar Farm', 'ST-AU-003', 'Balranald, New South Wales, Australia', 142.5124000, -34.5234000, 200.00, 900000.00, 363636, 1),
('STAT-091', 'Darlington Point Solar Farm', 'ST-AU-004', 'Darlington Point, New South Wales, Australia', 146.0123000, -34.6234000, 333.00, 1498500.00, 605454, 1),
('STAT-092', 'Dau Tieng Solar Power Complex', 'ST-VN-001', 'Dau Tieng District, Tay Ninh Province, Vietnam', 106.2124000, 11.3125000, 420.00, 1890000.00, 763636, 1),
('STAT-093', 'Sirindhorn Dam Floating Solar Farm', 'ST-TH-001', 'Sirindhorn, Ubon Ratchathani Province, Thailand', 105.4234000, 15.2345000, 45.00, 202500.00, 81818, 1),
('STAT-094', 'Cadiz Solar Power Plant', 'ST-PH-001', 'Cadiz City, Negros Occidental, Philippines', 123.0123000, 10.9234000, 132.00, 594000.00, 240000, 1),
('STAT-095', 'Minami Soma Solar Park', 'ST-JP-003', 'Minamisōma, Fukushima Prefecture, Japan', 140.9124000, 37.6234000, 59.00, 265500.00, 107272, 1);

-- ============================================================
-- 其他地区 (5个): STAT-096 ~ STAT-100
-- ============================================================
INSERT INTO stations (station_id, station_name, station_code, location, longitude, latitude, capacity, panel_area, panel_count, status) VALUES
('STAT-096', 'Sarnia Solar Farm', 'ST-CA-001', 'Sarnia, Ontario, Canada', -82.3124000, 42.9234000, 80.00, 360000.00, 145454, 1),
('STAT-097', 'Burnoye Solar Power Plant', 'ST-KZ-001', 'Burnoye, Jambyl Region, Kazakhstan', 71.4234000, 42.7123000, 100.00, 450000.00, 181818, 1),
('STAT-098', 'Nur Navoi Solar Park', 'ST-UZ-001', 'Navoiy, Navoiy Region, Uzbekistan', 66.9124000, 39.6234000, 100.00, 450000.00, 181818, 1),
('STAT-099', 'Sainshand Solar Park', 'ST-MN-001', 'Sainshand, Dornogovi Province, Mongolia', 110.1234000, 44.9124000, 30.00, 135000.00, 54545, 1),
('STAT-100', 'Rivne Solar Plant', 'ST-UA-001', 'Rivne, Rivne Oblast, Ukraine', 26.3124000, 50.6234000, 50.00, 225000.00, 90909, 1);
