package alerts

import (
	"bytes"
	"fmt"
	"text/template"
	"time"
)

type Severity int

const (
	SeverityInfo Severity = iota
	SeverityWarning
	SeverityCritical
)

func ParseSeverity(s string) Severity {
	switch s {
	case "warning":
		return SeverityWarning
	case "critical":
		return SeverityCritical
	default:
		return SeverityInfo
	}
}

func (s Severity) String() string {
	switch s {
	case SeverityWarning:
		return "warning"
	case SeverityCritical:
		return "critical"
	default:
		return "info"
	}
}

type Rule struct {
	Name      string
	Source    string
	Metric    string
	Operator  string
	Threshold any
	Severity  Severity
	Message   string
}

type AlertResult struct {
	Rule      Rule
	Value     any
	Triggered bool
	Timestamp time.Time
	Message   string
}

func (r *Rule) Evaluate(value any) (bool, error) {
	if value == nil {
		return false, nil
	}
	switch r.Operator {
	case ">":
		return compareNumeric(value, r.Threshold, func(a, b float64) bool { return a > b })
	case "<":
		return compareNumeric(value, r.Threshold, func(a, b float64) bool { return a < b })
	case ">=":
		return compareNumeric(value, r.Threshold, func(a, b float64) bool { return a >= b })
	case "<=":
		return compareNumeric(value, r.Threshold, func(a, b float64) bool { return a <= b })
	case "==":
		return fmt.Sprintf("%v", value) == fmt.Sprintf("%v", r.Threshold), nil
	case "!=":
		return fmt.Sprintf("%v", value) != fmt.Sprintf("%v", r.Threshold), nil
	default:
		return false, fmt.Errorf("unknown operator: %s", r.Operator)
	}
}

func (r *Rule) FormatMessage(value any) string {
	if r.Message == "" {
		return fmt.Sprintf("%s: %v %s %v", r.Name, value, r.Operator, r.Threshold)
	}

	tmpl, err := template.New("msg").Parse(r.Message)
	if err != nil {
		return r.Message
	}

	data := map[string]any{
		"Name":      r.Name,
		"Value":     value,
		"Threshold": r.Threshold,
		"Severity":  r.Severity.String(),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return r.Message
	}
	return buf.String()
}

func compareNumeric(value, threshold any, cmp func(float64, float64) bool) (bool, error) {
	v, err := toFloat64(value)
	if err != nil {
		return false, fmt.Errorf("value: %w", err)
	}
	t, err := toFloat64(threshold)
	if err != nil {
		return false, fmt.Errorf("threshold: %w", err)
	}
	return cmp(v, t), nil
}

func toFloat64(v any) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case int32:
		return float64(val), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}
