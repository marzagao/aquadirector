package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input string
		want  Format
	}{
		{"json", JSON},
		{"yaml", YAML},
		{"table", Table},
		{"", Table},
		{"unknown", Table},
	}

	for _, tt := range tests {
		got := ParseFormat(tt.input)
		if got != tt.want {
			t.Errorf("ParseFormat(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestPrint_Table(t *testing.T) {
	var buf bytes.Buffer
	rows := []TableRow{
		{Label: "Name", Value: "test"},
		{Label: "Status", Value: "ok"},
	}

	err := Print(&buf, Table, nil, rows)
	if err != nil {
		t.Fatalf("Print: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Name") || !strings.Contains(out, "test") {
		t.Errorf("table output missing expected content: %q", out)
	}
	if !strings.Contains(out, "Status") || !strings.Contains(out, "ok") {
		t.Errorf("table output missing expected content: %q", out)
	}
}

func TestPrint_JSON(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"name": "test"}

	err := Print(&buf, JSON, data, nil)
	if err != nil {
		t.Fatalf("Print: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `"name": "test"`) {
		t.Errorf("JSON output = %q, want to contain name:test", out)
	}
}

func TestPrint_JSON_RunAt(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"name": "test"}

	before := time.Now().Unix()
	err := Print(&buf, JSON, data, nil)
	after := time.Now().Unix()
	if err != nil {
		t.Fatalf("Print: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	runAt, ok := result["run_at"]
	if !ok {
		t.Fatal("JSON output missing run_at field")
	}
	ts, ok := runAt.(float64) // JSON numbers decode as float64
	if !ok {
		t.Fatalf("run_at is %T, want float64", runAt)
	}
	if int64(ts) < before || int64(ts) > after {
		t.Errorf("run_at = %d, want between %d and %d", int64(ts), before, after)
	}
	if result["name"] != "test" {
		t.Errorf("original fields missing from JSON output")
	}
}

func TestPrint_YAML(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"name": "test"}

	err := Print(&buf, YAML, data, nil)
	if err != nil {
		t.Fatalf("Print: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "name: test") {
		t.Errorf("YAML output = %q, want to contain name: test", out)
	}
}

func TestPrint_YAML_RunAt(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"name": "test"}

	err := Print(&buf, YAML, data, nil)
	if err != nil {
		t.Fatalf("Print: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "run_at:") {
		t.Errorf("YAML output missing run_at field: %q", out)
	}
	if !strings.Contains(out, "name: test") {
		t.Errorf("original fields missing from YAML output: %q", out)
	}
}

func TestPrint_Table_NoRunAt(t *testing.T) {
	var buf bytes.Buffer
	rows := []TableRow{{Label: "Status", Value: "ok"}}

	err := Print(&buf, Table, nil, rows)
	if err != nil {
		t.Fatalf("Print: %v", err)
	}

	if strings.Contains(buf.String(), "run_at") {
		t.Errorf("table output should not contain run_at")
	}
}

func TestPrintList_Table(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"IP", "MODEL"}
	rows := [][]string{
		{"192.168.1.1", "RSATO+"},
		{"192.168.1.2", "RSLED60"},
	}

	err := PrintList(&buf, Table, nil, headers, rows)
	if err != nil {
		t.Fatalf("PrintList: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "IP") || !strings.Contains(out, "RSATO+") {
		t.Errorf("list output missing expected content: %q", out)
	}
}

func TestPrintList_JSON(t *testing.T) {
	var buf bytes.Buffer
	data := []map[string]string{{"ip": "10.0.0.1"}}

	err := PrintList(&buf, JSON, data, nil, nil)
	if err != nil {
		t.Fatalf("PrintList: %v", err)
	}

	if !strings.Contains(buf.String(), "10.0.0.1") {
		t.Errorf("JSON list output missing expected content")
	}
}
