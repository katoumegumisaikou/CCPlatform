package repository

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"ccplatform/internal/config"
)

// InfluxWriter 负责把 MQTT 上行遥测数据写入 InfluxDB v2。
// 当前实现直接调用 InfluxDB HTTP Write API，避免额外引入官方 client 的运行时依赖。
type InfluxWriter struct {
	// endpoint 是最终写入地址，形如:
	// http://host:8086/api/v2/write?bucket=xxx&org=xxx&precision=ns
	endpoint string
	token    string       // InfluxDB v2 API Token，用于 Authorization 头
	client   *http.Client // 带超时的 HTTP Client，防止写入请求长期阻塞 MQTT 消息处理
}

// NewInfluxWriter 根据配置创建 InfluxDB 写入器。
// 当配置未启用时返回 nil，调用方可以继续传递该 nil writer；
// WritePoint 会把 nil receiver 视为“未启用”，直接跳过写入。
func NewInfluxWriter(cfg config.InfluxDBConfig) (*InfluxWriter, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	if cfg.URL == "" || cfg.Token == "" || cfg.Org == "" || cfg.Bucket == "" {
		return nil, fmt.Errorf("influxdb enabled but url/token/org/bucket is empty")
	}

	base, err := url.Parse(strings.TrimRight(cfg.URL, "/"))
	if err != nil {
		return nil, fmt.Errorf("parse influxdb url: %w", err)
	}

	// InfluxDB v2 写入接口固定为 /api/v2/write，
	// org、bucket 和 precision 通过 query string 传递。
	base.Path = "/api/v2/write"
	q := base.Query()
	q.Set("org", cfg.Org)
	q.Set("bucket", cfg.Bucket)
	q.Set("precision", "ns")
	base.RawQuery = q.Encode()

	return &InfluxWriter{
		endpoint: base.String(),
		token:    cfg.Token,
		client:   &http.Client{Timeout: 3 * time.Second},
	}, nil
}

// WritePoint 写入单条时序点。
// measurement 对应 InfluxDB 的 measurement，tags 用于低基数字段（如 robot_id），
// fields 存放实际指标值，ts 是数据采集时间；ts 为空时会在 lineProtocol 中回退为当前时间。
func (w *InfluxWriter) WritePoint(measurement string, tags map[string]string, fields map[string]interface{}, ts time.Time) error {
	// 未启用 InfluxDB 时 NewInfluxWriter 返回 nil，这里直接跳过，保持主链路无侵入。
	if w == nil {
		return nil
	}
	line, err := lineProtocol(measurement, tags, fields, ts)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, w.endpoint, bytes.NewBufferString(line+"\n"))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Token "+w.token)
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")

	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("influxdb write failed: %s %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

// lineProtocol 把业务字段转换成 InfluxDB line protocol。
// 格式为: measurement,tag_key=tag_value field_key=field_value timestamp
// 为了让测试和日志输出稳定，tags 和 fields 都会先按 key 排序。
func lineProtocol(measurement string, tags map[string]string, fields map[string]interface{}, ts time.Time) (string, error) {
	if measurement == "" || len(fields) == 0 {
		return "", fmt.Errorf("measurement and fields are required")
	}
	if ts.IsZero() {
		ts = time.Now()
	}

	var b strings.Builder
	b.WriteString(escapeKey(measurement))

	// tag 位于逗号分隔区，只适合放 robot_id、station_id 这类低基数字段；
	// 空 tag 值没有查询价值，直接跳过。
	for _, k := range sortedKeys(tags) {
		if tags[k] == "" {
			continue
		}
		b.WriteByte(',')
		b.WriteString(escapeKey(k))
		b.WriteByte('=')
		b.WriteString(escapeKey(tags[k]))
	}

	// field 是真正写入的指标值。formatField 返回 false 的类型会被忽略，
	// 避免把 InfluxDB 不支持的 Go 类型写成非法 line protocol。
	first := true
	for _, k := range sortedKeys(fields) {
		v, ok := formatField(fields[k])
		if !ok {
			continue
		}
		if first {
			b.WriteByte(' ')
			first = false
		} else {
			b.WriteByte(',')
		}
		b.WriteString(escapeKey(k))
		b.WriteByte('=')
		b.WriteString(v)
	}
	if first {
		return "", fmt.Errorf("no writable fields")
	}

	b.WriteByte(' ')
	b.WriteString(strconv.FormatInt(ts.UnixNano(), 10))
	return b.String(), nil
}

// sortedKeys 返回 map 的有序 key 列表。
// Go map 遍历顺序不稳定，排序后生成的 line protocol 更利于测试和排查问题。
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// escapeKey 转义 measurement、tag key、tag value 和 field key 中的特殊字符。
// InfluxDB line protocol 要求空格、逗号、等号和反斜杠在这些位置必须转义。
func escapeKey(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, " ", `\ `)
	s = strings.ReplaceAll(s, ",", `\,`)
	s = strings.ReplaceAll(s, "=", `\=`)
	return s
}

// formatField 把 Go 值转换成 InfluxDB field value 字面量。
// 整数需要带 i 后缀，无符号整数带 u 后缀；字符串需要引号并转义反斜杠和双引号。
// time.Time 会写成 UnixNano 整数，便于保留设备时间字段的时间精度。
// 返回 false 表示该类型当前不支持写入。
func formatField(v interface{}) (string, bool) {
	switch x := v.(type) {
	case int:
		return strconv.FormatInt(int64(x), 10) + "i", true
	case int8:
		return strconv.FormatInt(int64(x), 10) + "i", true
	case int64:
		return strconv.FormatInt(x, 10) + "i", true
	case uint:
		return strconv.FormatUint(uint64(x), 10) + "u", true
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64), true
	case string:
		return `"` + strings.ReplaceAll(strings.ReplaceAll(x, `\`, `\\`), `"`, `\"`) + `"`, true
	case bool:
		return strconv.FormatBool(x), true
	case time.Time:
		if x.IsZero() {
			return "", false
		}
		return strconv.FormatInt(x.UnixNano(), 10) + "i", true
	default:
		return "", false
	}
}
