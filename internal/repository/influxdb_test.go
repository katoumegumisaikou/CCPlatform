package repository

import (
	"strings"
	"testing"
	"time"
)

func TestLineProtocol(t *testing.T) {
	line, err := lineProtocol(
		"robot status",
		map[string]string{"robot_id": "r,1"},
		map[string]interface{}{"battery": int8(80), "speed": 1.25, "state": "ok"},
		time.Unix(1, 2),
	)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`robot\ status,robot_id=r\,1 `,
		`battery=80i`,
		`speed=1.25`,
		`state="ok"`,
		` 1000000002`,
	} {
		if !strings.Contains(line, want) {
			t.Fatalf("line %q missing %q", line, want)
		}
	}
}

func TestLineProtocolRobotTelemetryFields(t *testing.T) {
	line, err := lineProtocol(
		"robot_status",
		map[string]string{"robot_id": "robot-001"},
		map[string]interface{}{
			"battery_voltage":      48.6,
			"fault_status":         int8(1),
			"run_duration_seconds": int64(3600),
			"signal_strength":      78,
			"device_time":          time.Unix(1700000000, 123),
		},
		time.Unix(1700000001, 0),
	)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`robot_status,robot_id=robot-001 `,
		`battery_voltage=48.6`,
		`fault_status=1i`,
		`run_duration_seconds=3600i`,
		`signal_strength=78i`,
		`device_time=1700000000000000123i`,
		` 1700000001000000000`,
	} {
		if !strings.Contains(line, want) {
			t.Fatalf("line %q missing %q", line, want)
		}
	}
}
