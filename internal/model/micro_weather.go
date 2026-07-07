package model

import "time"

// MicroWeatherStation 微气象站设备模型，部署在电站现场采集微气象数据。
// 当前为硬件预留接口，支持手动录入和管理，后续可对接 MQTT 数据上行。
type MicroWeatherStation struct {
	StationCode    string     `gorm:"primaryKey;type:varchar(50)" json:"station_code"`      // 气象站编码（唯一标识）
	StationName    string     `gorm:"type:varchar(100);not null" json:"station_name"`        // 气象站名称
	PowerStationID string     `gorm:"type:varchar(32);index" json:"power_station_id"`        // 关联的光伏电站 ID
	Longitude      float64    `gorm:"type:decimal(10,7)" json:"longitude"`                   // 部署经度
	Latitude       float64    `gorm:"type:decimal(10,7)" json:"latitude"`                    // 部署纬度
	Status         int8       `gorm:"type:tinyint;default:1" json:"status"`                  // 状态: 0停用 1正常 2故障
	InstallTime    *time.Time `gorm:"type:datetime" json:"install_time"`                     // 安装时间
	CreateTime     time.Time  `gorm:"autoCreateTime" json:"create_time"`
	UpdateTime     time.Time  `gorm:"autoUpdateTime" json:"update_time"`
}

func (MicroWeatherStation) TableName() string { return "micro_weather_stations" }

// MicroWeatherData 微气象站观测数据记录。
// 支持手动录入或 API 写入，后续可对接现场设备 MQTT 数据上行。
type MicroWeatherData struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	StationCode    string    `gorm:"type:varchar(50);index" json:"station_code"`  // 关联微气象站编码
	Temperature    float64   `gorm:"type:decimal(5,2)" json:"temperature"`        // 温度（℃）
	Humidity       float64   `gorm:"type:decimal(5,2)" json:"humidity"`           // 相对湿度（%）
	WindSpeed      float64   `gorm:"type:decimal(5,2)" json:"wind_speed"`         // 风速（m/s）
	WindDirection  float64   `gorm:"type:decimal(5,2)" json:"wind_direction"`     // 风向（度）
	Pressure       float64   `gorm:"type:decimal(8,2)" json:"pressure"`           // 大气压强（hPa）
	Rainfall       float64   `gorm:"type:decimal(6,2)" json:"rainfall"`           // 降雨量（mm）
	SolarRadiation float64   `gorm:"type:decimal(8,2)" json:"solar_radiation"`    // 水平总辐射（W/m²）
	DustDensity    float64   `gorm:"type:decimal(6,2)" json:"dust_density"`       // 灰尘沉降量（μg/m³）
	PM25           float64   `gorm:"type:decimal(6,2)" json:"pm25"`               // PM2.5 浓度（μg/m³）
	PM10           float64   `gorm:"type:decimal(6,2)" json:"pm10"`               // PM10 浓度（μg/m³）
	RecordTime     time.Time `gorm:"type:datetime;index" json:"record_time"`      // 数据采集时间
}

func (MicroWeatherData) TableName() string { return "micro_weather_data" }
