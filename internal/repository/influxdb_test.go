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
