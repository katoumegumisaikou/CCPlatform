package model

import "time"

// WeatherNowCache 实时天气缓存，按 location_id 关联到电站区域。
// 主键为 location_id，每次从和风天气刷新时 UPSERT 覆盖写入。
type WeatherNowCache struct {
	LocationID string    `gorm:"primaryKey;type:varchar(20)" json:"location_id"`
	StationID  string    `gorm:"type:varchar(32);index" json:"station_id"` // 关联的光伏电站 ID
	ObsTime    string    `gorm:"type:varchar(25)" json:"obs_time"`         // 数据观测时间（ISO8601）
	Temp       string    `gorm:"type:varchar(10)" json:"temp"`             // 温度（℃）
	FeelsLike  string    `gorm:"type:varchar(10)" json:"feels_like"`       // 体感温度（℃）
	Icon       string    `gorm:"type:varchar(10)" json:"icon"`             // 天气图标代码
	Text       string    `gorm:"type:varchar(20)" json:"text"`             // 天气状况文字描述
	WindDir    string    `gorm:"type:varchar(20)" json:"wind_dir"`         // 风向文字
	WindScale  string    `gorm:"type:varchar(10)" json:"wind_scale"`       // 风力等级
	WindSpeed  string    `gorm:"type:varchar(10)" json:"wind_speed"`       // 风速（km/h）
	Humidity   string    `gorm:"type:varchar(10)" json:"humidity"`         // 相对湿度（%）
	Precip     string    `gorm:"type:varchar(10)" json:"precip"`           // 过去 1 小时降水量（mm）
	Pressure   string    `gorm:"type:varchar(10)" json:"pressure"`         // 大气压强（hPa）
	Vis        string    `gorm:"type:varchar(10)" json:"vis"`              // 能见度（km）
	Cloud      string    `gorm:"type:varchar(10)" json:"cloud"`            // 云量（%）
	UpdateTime time.Time `gorm:"autoUpdateTime" json:"update_time"`        // 最后刷新时间
}

func (WeatherNowCache) TableName() string { return "weather_now_cache" }

// WeatherDailyCache 7 天逐天预报缓存。
// 联合主键 (location_id, fx_date)，每次刷新时批量 UPSERT。
type WeatherDailyCache struct {
	LocationID   string    `gorm:"primaryKey;type:varchar(20)" json:"location_id"`
	FxDate       string    `gorm:"primaryKey;type:varchar(10)" json:"fx_date"` // 预报日期（yyyy-MM-dd）
	StationID    string    `gorm:"type:varchar(32);index" json:"station_id"`
	Sunrise      string    `gorm:"type:varchar(10)" json:"sunrise"`          // 日出时间
	Sunset       string    `gorm:"type:varchar(10)" json:"sunset"`           // 日落时间
	TempMax      string    `gorm:"type:varchar(10)" json:"temp_max"`         // 最高温度（℃）
	TempMin      string    `gorm:"type:varchar(10)" json:"temp_min"`         // 最低温度（℃）
	IconDay      string    `gorm:"type:varchar(10)" json:"icon_day"`         // 白天天况图标
	TextDay      string    `gorm:"type:varchar(20)" json:"text_day"`         // 白天天况描述
	IconNight    string    `gorm:"type:varchar(10)" json:"icon_night"`       // 夜晚天况图标
	TextNight    string    `gorm:"type:varchar(20)" json:"text_night"`       // 夜晚天况描述
	WindDirDay   string    `gorm:"type:varchar(20)" json:"wind_dir_day"`     // 白天风向
	WindScaleDay string    `gorm:"type:varchar(10)" json:"wind_scale_day"`   // 白天风力等级
	Humidity     string    `gorm:"type:varchar(10)" json:"humidity"`         // 相对湿度（%）
	Precip       string    `gorm:"type:varchar(10)" json:"precip"`           // 降水量（mm）
	UvIndex      string    `gorm:"type:varchar(10)" json:"uv_index"`         // 紫外线指数
	UpdateTime   time.Time `gorm:"autoUpdateTime" json:"update_time"`
}

func (WeatherDailyCache) TableName() string { return "weather_daily_cache" }

// WeatherHourlyCache 逐小时预报缓存（默认获取 24 小时）。
// 联合主键 (location_id, fx_time)，每次刷新时批量 UPSERT。
type WeatherHourlyCache struct {
	LocationID string    `gorm:"primaryKey;type:varchar(20)" json:"location_id"`
	FxTime     string    `gorm:"primaryKey;type:varchar(20)" json:"fx_time"` // 预报时间（ISO8601）
	StationID  string    `gorm:"type:varchar(32);index" json:"station_id"`
	Temp       string    `gorm:"type:varchar(10)" json:"temp"`       // 温度（℃）
	Icon       string    `gorm:"type:varchar(10)" json:"icon"`       // 天气图标代码
	Text       string    `gorm:"type:varchar(20)" json:"text"`       // 天气描述
	WindDir    string    `gorm:"type:varchar(20)" json:"wind_dir"`   // 风向
	WindScale  string    `gorm:"type:varchar(10)" json:"wind_scale"` // 风力等级
	WindSpeed  string    `gorm:"type:varchar(10)" json:"wind_speed"` // 风速（km/h）
	Humidity   string    `gorm:"type:varchar(10)" json:"humidity"`   // 相对湿度（%）
	Precip     string    `gorm:"type:varchar(10)" json:"precip"`     // 降水量（mm）
	Pressure   string    `gorm:"type:varchar(10)" json:"pressure"`   // 大气压强（hPa）
	Cloud      string    `gorm:"type:varchar(10)" json:"cloud"`      // 云量（%）
	UpdateTime time.Time `gorm:"autoUpdateTime" json:"update_time"`
}

func (WeatherHourlyCache) TableName() string { return "weather_hourly_cache" }

// WeatherWarning 天气预警记录。
// 联合主键 (location_id, warning_id)，预警以追加模式存储，仅当生效期间保留。
type WeatherWarning struct {
	LocationID string    `gorm:"primaryKey;type:varchar(20)" json:"location_id"`
	WarningID  string    `gorm:"primaryKey;type:varchar(32)" json:"warning_id"`   // 和风天气预警唯一 ID
	StationID  string    `gorm:"type:varchar(32);index" json:"station_id"`
	Sender     string    `gorm:"type:varchar(50)" json:"sender"`                  // 发布单位
	PubTime    string    `gorm:"type:varchar(20)" json:"pub_time"`                // 发布时间
	Title      string    `gorm:"type:varchar(200)" json:"title"`                  // 预警标题
	StartTime  string    `gorm:"type:varchar(20)" json:"start_time"`              // 预警开始时间
	EndTime    string    `gorm:"type:varchar(20)" json:"end_time"`                // 预警结束时间
	Status     string    `gorm:"type:varchar(20)" json:"status"`                  // 预警状态
	Level      string    `gorm:"type:varchar(20)" json:"level"`                   // 预警级别（预警信号等级颜色）
	Type       string    `gorm:"type:varchar(20)" json:"type"`                    // 预警类型代码
	TypeName   string    `gorm:"type:varchar(50)" json:"type_name"`               // 预警类型名称
	Text       string    `gorm:"type:text" json:"text"`                           // 预警详细文字说明
	UpdateTime time.Time `gorm:"autoUpdateTime" json:"update_time"`
}

func (WeatherWarning) TableName() string { return "weather_warnings" }

// WeatherAirQuality 空气质量缓存，按 location_id 关联到电站区域。
type WeatherAirQuality struct {
	LocationID string    `gorm:"primaryKey;type:varchar(20)" json:"location_id"`
	StationID  string    `gorm:"type:varchar(32);index" json:"station_id"`
	PubTime    string    `gorm:"type:varchar(20)" json:"pub_time"`  // 数据发布时间
	Aqi        string    `gorm:"type:varchar(10)" json:"aqi"`       // AQI 指数
	Level      string    `gorm:"type:varchar(10)" json:"level"`     // 空气质量等级（数字）
	Category   string    `gorm:"type:varchar(20)" json:"category"`  // 空气质量类别（优/良等）
	Primary    string    `gorm:"type:varchar(20)" json:"primary"`   // 首要污染物
	Pm10       string    `gorm:"type:varchar(10)" json:"pm10"`      // PM10（μg/m³）
	Pm2p5      string    `gorm:"type:varchar(10)" json:"pm2p5"`     // PM2.5（μg/m³）
	No2        string    `gorm:"type:varchar(10)" json:"no2"`       // NO₂（μg/m³）
	So2        string    `gorm:"type:varchar(10)" json:"so2"`       // SO₂（μg/m³）
	Co         string    `gorm:"type:varchar(10)" json:"co"`        // CO（mg/m³）
	O3         string    `gorm:"type:varchar(10)" json:"o3"`        // O₃（μg/m³）
	UpdateTime time.Time `gorm:"autoUpdateTime" json:"update_time"` // 最后刷新时间
}

func (WeatherAirQuality) TableName() string { return "weather_air_quality" }
