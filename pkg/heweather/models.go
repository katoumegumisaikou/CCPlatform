// Package heweather 封装和风天气 API v7 的 HTTP 调用和响应数据结构。
//
// 和风天气 API 返回的数值字段均为字符串类型（如温度 "24"、湿度 "72"），
// 调用方需要在业务层自行转换为数值类型进行计算。
package heweather

// ─── 实时天气 ───────────────────────────────────────────────

// HfNowResponse 实时天气 API (/v7/weather/now) 的顶层响应。
type HfNowResponse struct {
	Code       string    `json:"code"`       // API 状态码，"200" 表示成功
	UpdateTime string    `json:"updateTime"` // 当前 API 的最近更新时间（ISO8601）
	FxLink     string    `json:"fxLink"`     // 响应式天气页面链接
	Now        HfNowData `json:"now"`        // 实况天气数据
	Refer      HfRefer   `json:"refer"`      // 数据来源与授权信息
}

// HfNowData 实时天气的核心数据字段。
type HfNowData struct {
	ObsTime   string `json:"obsTime"`   // 数据观测时间
	Temp      string `json:"temp"`      // 温度（℃）
	FeelsLike string `json:"feelsLike"` // 体感温度（℃）
	Icon      string `json:"icon"`      // 天气图标代码
	Text      string `json:"text"`      // 天气状况文字描述
	Wind360   string `json:"wind360"`   // 风向 360 角度
	WindDir   string `json:"windDir"`   // 风向文字描述
	WindScale string `json:"windScale"` // 风力等级
	WindSpeed string `json:"windSpeed"` // 风速（km/h）
	Humidity  string `json:"humidity"`  // 相对湿度（%）
	Precip    string `json:"precip"`    // 过去 1 小时降水量（mm）
	Pressure  string `json:"pressure"`  // 大气压强（hPa）
	Vis       string `json:"vis"`       // 能见度（km）
	Cloud     string `json:"cloud"`     // 云量（%，可能为空）
	Dew       string `json:"dew"`       // 露点温度（℃）
}

// ─── 7 天预报 ───────────────────────────────────────────────

// Hf7dResponse 7 天天气预报 API (/v7/weather/7d) 的顶层响应。
type Hf7dResponse struct {
	Code       string         `json:"code"`
	UpdateTime string         `json:"updateTime"`
	FxLink     string         `json:"fxLink"`
	Daily      []HfDailyData  `json:"daily"`
	Refer      HfRefer        `json:"refer"`
}

// HfDailyData 单日预报的核心数据字段。
type HfDailyData struct {
	FxDate         string `json:"fxDate"`         // 预报日期（yyyy-MM-dd）
	Sunrise        string `json:"sunrise"`        // 日出时间（HH:mm）
	Sunset         string `json:"sunset"`         // 日落时间（HH:mm）
	Moonrise       string `json:"moonrise"`       // 月升时间
	Moonset        string `json:"moonset"`        // 月落时间
	MoonPhase      string `json:"moonPhase"`      // 月相名称
	TempMax        string `json:"tempMax"`        // 最高温度（℃）
	TempMin        string `json:"tempMin"`        // 最低温度（℃）
	IconDay        string `json:"iconDay"`        // 白天天况图标代码
	TextDay        string `json:"textDay"`        // 白天天况文字描述
	IconNight      string `json:"iconNight"`      // 夜晚天况图标代码
	TextNight      string `json:"textNight"`      // 夜晚天况文字描述
	WindDirDay     string `json:"windDirDay"`     // 白天风向
	WindScaleDay   string `json:"windScaleDay"`   // 白天风力等级
	WindSpeedDay   string `json:"windSpeedDay"`   // 白天风速（km/h）
	WindDirNight   string `json:"windDirNight"`   // 夜间风向
	WindScaleNight string `json:"windScaleNight"` // 夜间风力等级
	WindSpeedNight string `json:"windSpeedNight"` // 夜间风速（km/h）
	Humidity       string `json:"humidity"`       // 相对湿度（%）
	Precip         string `json:"precip"`         // 降水量（mm）
	Pressure       string `json:"pressure"`       // 大气压强（hPa）
	Vis            string `json:"vis"`            // 能见度（km）
	Cloud          string `json:"cloud"`          // 云量（%）
	UvIndex        string `json:"uvIndex"`        // 紫外线指数
}

// ─── 逐小时预报 ─────────────────────────────────────────────

// Hf24hResponse 逐小时天气预报 API (/v7/weather/24h) 的顶层响应。
type Hf24hResponse struct {
	Code       string          `json:"code"`
	UpdateTime string          `json:"updateTime"`
	FxLink     string          `json:"fxLink"`
	Hourly     []HfHourlyData  `json:"hourly"`
	Refer      HfRefer         `json:"refer"`
}

// HfHourlyData 单小时预报的核心数据字段。
type HfHourlyData struct {
	FxTime    string `json:"fxTime"`    // 预报时间（ISO8601）
	Temp      string `json:"temp"`      // 温度（℃）
	Icon      string `json:"icon"`      // 天气图标代码
	Text      string `json:"text"`      // 天气状况文字描述
	Wind360   string `json:"wind360"`   // 风向 360 角度
	WindDir   string `json:"windDir"`   // 风向文字描述
	WindScale string `json:"windScale"` // 风力等级
	WindSpeed string `json:"windSpeed"` // 风速（km/h）
	Humidity  string `json:"humidity"`  // 相对湿度（%）
	Pop       string `json:"pop"`       // 降水概率（%）
	Precip    string `json:"precip"`    // 降水量（mm）
	Pressure  string `json:"pressure"`  // 大气压强（hPa）
	Cloud     string `json:"cloud"`     // 云量（%）
	Dew       string `json:"dew"`       // 露点温度（℃）
}

// ─── 天气预警 ───────────────────────────────────────────────

// HfWarningResponse 天气预警 API (/v7/warning/now) 的顶层响应。
type HfWarningResponse struct {
	Code       string           `json:"code"`
	UpdateTime string           `json:"updateTime"`
	FxLink     string           `json:"fxLink"`
	Warning    []HfWarningData  `json:"warning"`
	Refer      HfRefer          `json:"refer"`
}

// HfWarningData 单条预警信息的核心字段。
type HfWarningData struct {
	ID        string `json:"id"`        // 预警 ID
	Sender    string `json:"sender"`    // 发布单位
	PubTime   string `json:"pubTime"`   // 发布时间
	Title     string `json:"title"`     // 预警标题
	StartTime string `json:"startTime"` // 预警开始时间
	EndTime   string `json:"endTime"`   // 预警结束时间
	Status    string `json:"status"`    // 预警状态
	Level     string `json:"level"`     // 预警级别
	Type      string `json:"type"`      // 预警类型代码
	TypeName  string `json:"typeName"`  // 预警类型名称
	Text      string `json:"text"`      // 预警详细文字
	Related   string `json:"related"`   // 关联的预警 ID
}

// ─── 空气质量 ────────────────────────────────────────────────

// HfAirQualityResponse 空气质量 API 的顶层响应。注意和风天气的空气质量接口
// 使用独立的子域名 devapi.qweather.com/airquality/，响应结构也与其他天气接口不同。
type HfAirQualityResponse struct {
	Code       string         `json:"code"`
	UpdateTime string         `json:"updateTime"`
	FxLink     string         `json:"fxLink"`
	Now        HfAirNowData   `json:"now"`
	Station    []HfAirStation `json:"station"`
	Refer      HfRefer        `json:"refer"`
}

// HfAirNowData 空气质量实时数据。
type HfAirNowData struct {
	PubTime  string `json:"pubTime"`  // 数据发布时间
	Aqi      string `json:"aqi"`      // AQI 指数
	Level    string `json:"level"`    // 空气质量等级（数字）
	Category string `json:"category"` // 空气质量类别（优/良/轻度污染等）
	Primary  string `json:"primary"`  // 首要污染物
	Pm10     string `json:"pm10"`     // PM10 浓度（μg/m³）
	Pm2p5    string `json:"pm2p5"`    // PM2.5 浓度（μg/m³）
	No2      string `json:"no2"`      // NO₂ 浓度（μg/m³）
	So2      string `json:"so2"`      // SO₂ 浓度（μg/m³）
	Co       string `json:"co"`       // CO 浓度（mg/m³）
	O3       string `json:"o3"`       // O₃ 浓度（μg/m³）
}

// HfAirStation 空气质量监测站信息。
type HfAirStation struct {
	Name     string `json:"name"`     // 监测站名称
	ID       string `json:"id"`       // 监测站 ID
	PubTime  string `json:"pubTime"`  // 发布时间
	Aqi      string `json:"aqi"`      // AQI
	Level    string `json:"level"`    // 等级
	Category string `json:"category"` // 类别
	Primary  string `json:"primary"`  // 首要污染物
	Pm10     string `json:"pm10"`
	Pm2p5    string `json:"pm2p5"`
	No2      string `json:"no2"`
	So2      string `json:"so2"`
	Co       string `json:"co"`
	O3       string `json:"o3"`
}

// ─── 公共 ───────────────────────────────────────────────────

// HfRefer 数据来源与授权声明。
type HfRefer struct {
	Sources []string `json:"sources"` // 数据来源列表
	License []string `json:"license"` // 授权声明
}
