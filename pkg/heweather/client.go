package heweather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"ccplatform/internal/config"
)

// Client 是和风天气 API v7 的 HTTP 客户端。
// Web API 类型使用 API Key 通过 URL 参数认证。
type Client struct {
	apiKey  string
	baseURL string
	airURL  string
	hc      *http.Client
}

// NewClient 根据配置创建和风天气客户端。
func NewClient(cfg config.HeWeatherConfig) *Client {
	return &Client{
		apiKey:  cfg.APIToken,
		baseURL: cfg.BaseURL,
		airURL:  cfg.AirURL,
		hc:      &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) FetchNowWeather(location string) (*HfNowResponse, error) {
	resp := &HfNowResponse{}
	err := c.getJSON(c.baseURL+"weather/now", location, resp)
	return resp, err
}

func (c *Client) Fetch7dForecast(location string) (*Hf7dResponse, error) {
	resp := &Hf7dResponse{}
	err := c.getJSON(c.baseURL+"weather/7d", location, resp)
	return resp, err
}

func (c *Client) Fetch24hForecast(location string) (*Hf24hResponse, error) {
	resp := &Hf24hResponse{}
	err := c.getJSON(c.baseURL+"weather/24h", location, resp)
	return resp, err
}

func (c *Client) FetchWarning(location string) (*HfWarningResponse, error) {
	resp := &HfWarningResponse{}
	err := c.getJSON(c.baseURL+"warning/now", location, resp)
	return resp, err
}

func (c *Client) FetchAirQuality(location string) (*HfAirQualityResponse, error) {
	resp := &HfAirQualityResponse{}
	err := c.getJSON(c.airURL+"air/now", location, resp)
	return resp, err
}

// getJSON 发起 GET 请求，通过 URL 参数传递 location 和 key（Web API 标准方式）。
func (c *Client) getJSON(apiURL, location string, dst interface{}) error {
	u, err := url.Parse(apiURL)
	if err != nil {
		return fmt.Errorf("parse url: %w", err)
	}

	q := u.Query()
	q.Set("location", location)
	q.Set("key", c.apiKey)
	u.RawQuery = q.Encode()

	resp, err := c.hc.Get(u.String())
	if err != nil {
		return fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("api returned %d: %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, dst); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	return nil
}
