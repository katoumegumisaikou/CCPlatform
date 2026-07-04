package service

import (
	"ccplatform/internal/config"
	"ccplatform/internal/model"
	"ccplatform/internal/repository"
	"ccplatform/pkg/heweather"
	"fmt"
	"log"
	"time"
)

// WeatherService 气象业务逻辑层，封装和风天气 API 的数据获取、缓存刷新和查询。
type WeatherService struct {
	weatherRepo *repository.WeatherRepo
	stationRepo *repository.StationRepo
	client      *heweather.Client
}

// NewWeatherService 创建 WeatherService 实例。
func NewWeatherService() *WeatherService {
	return &WeatherService{
		weatherRepo: repository.NewWeatherRepo(),
		stationRepo: repository.NewStationRepo(),
		client:      heweather.NewClient(config.Cfg.HeWeather),
	}
}

// RefreshStationWeather 从和风天气 API 拉取指定电站的最新气象数据并写入缓存。
func (s *WeatherService) RefreshStationWeather(stationID string) error {
	station, err := s.stationRepo.GetByID(stationID)
	if err != nil {
		return fmt.Errorf("get station %s: %w", stationID, err)
	}

	loc := fmt.Sprintf("%.2f,%.2f", station.Longitude, station.Latitude)
	log.Printf("[Weather] Refreshing weather for station %s (location=%s)", stationID, loc)

	errCh := make(chan error, 5)

	go func() { errCh <- s.refreshNow(loc, stationID) }()
	go func() { errCh <- s.refreshDaily(loc, stationID) }()
	go func() { errCh <- s.refreshHourly(loc, stationID) }()
	go func() { errCh <- s.refreshWarning(loc, stationID) }()
	go func() { errCh <- s.refreshAirQuality(loc, stationID) }()

	var errs []error
	for i := 0; i < 5; i++ {
		if err := <-errCh; err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 5 {
		return fmt.Errorf("all 5 HeFeng APIs failed: %v", errs)
	}
	if len(errs) > 0 {
		log.Printf("[Weather] Station %s: %d/%d APIs failed", stationID, len(errs), 5)
	}
	return nil
}

// RefreshAllStationsWeather 刷新所有启用电站的天气数据。
func (s *WeatherService) RefreshAllStationsWeather() error {
	stations, err := s.stationRepo.GetAll()
	if err != nil {
		return fmt.Errorf("get all stations: %w", err)
	}

	log.Printf("[Weather] Refreshing weather for %d station(s)", len(stations))
	for _, st := range stations {
		if err := s.RefreshStationWeather(st.StationID); err != nil {
			log.Printf("[Weather] Station %s refresh error: %v", st.StationID, err)
		}
		time.Sleep(500 * time.Millisecond)
	}
	return nil
}

// ─── 查询方法 ────────────────────────────────────────────────

func (s *WeatherService) GetNowWeather(stationID string) (*model.WeatherNowCache, error) {
	return s.weatherRepo.GetNowByStationID(stationID)
}

func (s *WeatherService) GetDailyForecast(stationID string, days int) ([]model.WeatherDailyCache, error) {
	return s.weatherRepo.GetDailyByStationID(stationID, days)
}

func (s *WeatherService) GetHourlyForecast(stationID string, hours int) ([]model.WeatherHourlyCache, error) {
	return s.weatherRepo.GetHourlyByStationID(stationID, hours)
}

func (s *WeatherService) GetActiveWarnings(stationID string) ([]model.WeatherWarning, error) {
	return s.weatherRepo.GetActiveWarnings(stationID)
}

func (s *WeatherService) GetWarningHistory(stationID string, page, size int, startDate, endDate string) ([]model.WeatherWarning, int64, error) {
	return s.weatherRepo.ListWarnings(stationID, page, size, startDate, endDate)
}

func (s *WeatherService) GetAirQuality(stationID string) (*model.WeatherAirQuality, error) {
	return s.weatherRepo.GetAirQualityByStationID(stationID)
}

// ─── 内部刷新方法 ────────────────────────────────────────────

func (s *WeatherService) refreshNow(loc, stationID string) error {
	resp, err := s.client.FetchNowWeather(loc)
	if err != nil {
		return fmt.Errorf("fetch now: %w", err)
	}
	if resp.Code != "200" {
		return fmt.Errorf("now api code=%s", resp.Code)
	}

	cache := &model.WeatherNowCache{
		LocationID: loc,
		StationID:  stationID,
		ObsTime:    resp.Now.ObsTime,
		Temp:       resp.Now.Temp,
		FeelsLike:  resp.Now.FeelsLike,
		Icon:       resp.Now.Icon,
		Text:       resp.Now.Text,
		WindDir:    resp.Now.WindDir,
		WindScale:  resp.Now.WindScale,
		WindSpeed:  resp.Now.WindSpeed,
		Humidity:   resp.Now.Humidity,
		Precip:     resp.Now.Precip,
		Pressure:   resp.Now.Pressure,
		Vis:        resp.Now.Vis,
		Cloud:      resp.Now.Cloud,
	}
	return s.weatherRepo.UpsertNow(cache)
}

func (s *WeatherService) refreshDaily(loc, stationID string) error {
	resp, err := s.client.Fetch7dForecast(loc)
	if err != nil {
		return fmt.Errorf("fetch 7d: %w", err)
	}
	if resp.Code != "200" {
		return fmt.Errorf("7d api code=%s", resp.Code)
	}

	caches := make([]model.WeatherDailyCache, 0, len(resp.Daily))
	for _, d := range resp.Daily {
		caches = append(caches, model.WeatherDailyCache{
			LocationID:   loc,
			FxDate:       d.FxDate,
			StationID:    stationID,
			Sunrise:      d.Sunrise,
			Sunset:       d.Sunset,
			TempMax:      d.TempMax,
			TempMin:      d.TempMin,
			IconDay:      d.IconDay,
			TextDay:      d.TextDay,
			IconNight:    d.IconNight,
			TextNight:    d.TextNight,
			WindDirDay:   d.WindDirDay,
			WindScaleDay: d.WindScaleDay,
			Humidity:     d.Humidity,
			Precip:       d.Precip,
			UvIndex:      d.UvIndex,
		})
	}
	return s.weatherRepo.BatchUpsertDaily(caches)
}

func (s *WeatherService) refreshHourly(loc, stationID string) error {
	resp, err := s.client.Fetch24hForecast(loc)
	if err != nil {
		return fmt.Errorf("fetch 24h: %w", err)
	}
	if resp.Code != "200" {
		return fmt.Errorf("24h api code=%s", resp.Code)
	}

	caches := make([]model.WeatherHourlyCache, 0, len(resp.Hourly))
	for _, h := range resp.Hourly {
		caches = append(caches, model.WeatherHourlyCache{
			LocationID: loc,
			FxTime:     h.FxTime,
			StationID:  stationID,
			Temp:       h.Temp,
			Icon:       h.Icon,
			Text:       h.Text,
			WindDir:    h.WindDir,
			WindScale:  h.WindScale,
			WindSpeed:  h.WindSpeed,
			Humidity:   h.Humidity,
			Precip:     h.Precip,
			Pressure:   h.Pressure,
			Cloud:      h.Cloud,
		})
	}
	return s.weatherRepo.BatchUpsertHourly(caches)
}

func (s *WeatherService) refreshWarning(loc, stationID string) error {
	resp, err := s.client.FetchWarning(loc)
	if err != nil {
		return fmt.Errorf("fetch warning: %w", err)
	}
	if resp.Code != "200" {
		log.Printf("[Weather] Warning api code=%s for %s", resp.Code, stationID)
		return nil
	}

	for _, w := range resp.Warning {
		warning := &model.WeatherWarning{
			LocationID: loc,
			WarningID:  w.ID,
			StationID:  stationID,
			Sender:     w.Sender,
			PubTime:    w.PubTime,
			Title:      w.Title,
			StartTime:  w.StartTime,
			EndTime:    w.EndTime,
			Status:     w.Status,
			Level:      w.Level,
			Type:       w.Type,
			TypeName:   w.TypeName,
			Text:       w.Text,
		}
		if err := s.weatherRepo.UpsertWarning(warning); err != nil {
			log.Printf("[Weather] Upsert warning %s: %v", w.ID, err)
		}
	}
	return nil
}

func (s *WeatherService) refreshAirQuality(loc, stationID string) error {
	resp, err := s.client.FetchAirQuality(loc)
	if err != nil {
		log.Printf("[Weather] Air quality fetch error (may be unsupported on free plan): %v", err)
		return nil // 免费套餐不支持空气质量，不视为错误
	}
	if resp.Code != "200" {
		log.Printf("[Weather] Air quality api code=%s (free plan may not include this)", resp.Code)
		return nil
	}

	air := &model.WeatherAirQuality{
		LocationID: loc,
		StationID:  stationID,
		PubTime:    resp.Now.PubTime,
		Aqi:        resp.Now.Aqi,
		Level:      resp.Now.Level,
		Category:   resp.Now.Category,
		Primary:    resp.Now.Primary,
		Pm10:       resp.Now.Pm10,
		Pm2p5:      resp.Now.Pm2p5,
		No2:        resp.Now.No2,
		So2:        resp.Now.So2,
		Co:         resp.Now.Co,
		O3:         resp.Now.O3,
	}
	return s.weatherRepo.UpsertAirQuality(air)
}
