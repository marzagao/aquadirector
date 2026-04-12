package alerts

import "testing"

func TestRuleEvaluateNumeric(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		operator  string
		threshold any
		want      bool
	}{
		{"greater true", 28.0, ">", 27.0, true},
		{"greater false", 26.0, ">", 27.0, false},
		{"less true", 23.0, "<", 24.0, true},
		{"less false", 25.0, "<", 24.0, false},
		{"gte equal", 27.0, ">=", 27.0, true},
		{"lte equal", 24.0, "<=", 24.0, true},
		{"int value", 500, "<", 1000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Rule{Operator: tt.operator, Threshold: tt.threshold}
			got, err := r.Evaluate(tt.value)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Evaluate(%v %s %v) = %v, want %v", tt.value, tt.operator, tt.threshold, got, tt.want)
			}
		})
	}
}

func TestRuleEvaluateString(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		operator  string
		threshold any
		want      bool
	}{
		{"eq true", "dry", "==", "dry", true},
		{"eq false", "wet", "==", "dry", false},
		{"neq true", "wet", "!=", "dry", true},
		{"neq false", "dry", "!=", "dry", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Rule{Operator: tt.operator, Threshold: tt.threshold}
			got, err := r.Evaluate(tt.value)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Evaluate(%v %s %v) = %v, want %v", tt.value, tt.operator, tt.threshold, got, tt.want)
			}
		})
	}
}

func TestRuleFormatMessage(t *testing.T) {
	r := Rule{
		Name:      "temp_high",
		Operator:  ">",
		Threshold: 27.0,
		Message:   "Temperature is {{.Value}}C (threshold: {{.Threshold}}C)",
	}

	msg := r.FormatMessage(28.5)
	expected := "Temperature is 28.5C (threshold: 27C)"
	if msg != expected {
		t.Errorf("FormatMessage = %q, want %q", msg, expected)
	}
}

func TestSeverityParsing(t *testing.T) {
	if ParseSeverity("info") != SeverityInfo {
		t.Error("expected SeverityInfo")
	}
	if ParseSeverity("warning") != SeverityWarning {
		t.Error("expected SeverityWarning")
	}
	if ParseSeverity("critical") != SeverityCritical {
		t.Error("expected SeverityCritical")
	}
	if ParseSeverity("unknown") != SeverityInfo {
		t.Error("expected SeverityInfo for unknown")
	}
}
