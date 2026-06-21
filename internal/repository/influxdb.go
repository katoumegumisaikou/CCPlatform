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

// InfluxWriter writes MQTT telemetry to InfluxDB v2 using the HTTP write API.
type InfluxWriter struct {
	endpoint string
	token    string
	client   *http.Client
}

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

func (w *InfluxWriter) WritePoint(measurement string, tags map[string]string, fields map[string]interface{}, ts time.Time) error {
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

func lineProtocol(measurement string, tags map[string]string, fields map[string]interface{}, ts time.Time) (string, error) {
	if measurement == "" || len(fields) == 0 {
		return "", fmt.Errorf("measurement and fields are required")
	}
	if ts.IsZero() {
		ts = time.Now()
	}

	var b strings.Builder
	b.WriteString(escapeKey(measurement))

	for _, k := range sortedKeys(tags) {
		if tags[k] == "" {
			continue
		}
		b.WriteByte(',')
		b.WriteString(escapeKey(k))
		b.WriteByte('=')
		b.WriteString(escapeKey(tags[k]))
	}

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

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func escapeKey(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, " ", `\ `)
	s = strings.ReplaceAll(s, ",", `\,`)
	s = strings.ReplaceAll(s, "=", `\=`)
	return s
}

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
	default:
		return "", false
	}
}
