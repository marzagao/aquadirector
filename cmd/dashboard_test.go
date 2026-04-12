package cmd

import "testing"

func TestPhStatus(t *testing.T) {
	tests := []struct {
		ph   float64
		want string
	}{
		{7.7, "critical"},
		{7.79, "critical"},
		{7.8, "low"},
		{8.0, "low"},
		{8.09, "low"},
		{8.1, "ok"},
		{8.3, "ok"},
		{8.31, "high"},
		{9.0, "high"},
	}
	for _, tt := range tests {
		got := phStatus(tt.ph)
		if got != tt.want {
			t.Errorf("phStatus(%.2f) = %q, want %q", tt.ph, got, tt.want)
		}
	}
}

func TestOrpStatus(t *testing.T) {
	tests := []struct {
		orp  int
		want string
	}{
		{0, "critical"},
		{99, "critical"},
		{100, "low"},
		{199, "low"},
		{200, "ok"},
		{450, "ok"},
		{451, "high"},
		{600, "high"},
	}
	for _, tt := range tests {
		got := orpStatus(tt.orp)
		if got != tt.want {
			t.Errorf("orpStatus(%d) = %q, want %q", tt.orp, got, tt.want)
		}
	}
}
