package color

import (
	"strings"
	"testing"
)

func TestDisabled(t *testing.T) {
	Enabled = false
	cases := []struct {
		name string
		fn   func(string) string
	}{
		{"Red", Red},
		{"Green", Green},
		{"Yellow", Yellow},
		{"Blue", Blue},
		{"Bold", Bold},
		{"Dim", Dim},
	}
	for _, tc := range cases {
		if got := tc.fn("hello"); got != "hello" {
			t.Errorf("%s with Enabled=false: got %q, want %q", tc.name, got, "hello")
		}
	}
}

func TestEnabled(t *testing.T) {
	Enabled = true
	t.Cleanup(func() { Enabled = false })

	cases := []struct {
		name string
		fn   func(string) string
		code string
	}{
		{"Red", Red, "\x1b[31m"},
		{"Green", Green, "\x1b[32m"},
		{"Yellow", Yellow, "\x1b[33m"},
		{"Blue", Blue, "\x1b[34m"},
		{"Bold", Bold, "\x1b[1m"},
		{"Dim", Dim, "\x1b[2m"},
	}
	for _, tc := range cases {
		got := tc.fn("hello")
		if !strings.HasPrefix(got, tc.code) {
			t.Errorf("%s: missing prefix %q in %q", tc.name, tc.code, got)
		}
		if !strings.HasSuffix(got, "\x1b[0m") {
			t.Errorf("%s: missing reset suffix in %q", tc.name, got)
		}
		if !strings.Contains(got, "hello") {
			t.Errorf("%s: missing payload in %q", tc.name, got)
		}
	}
}

func TestStatusLabel(t *testing.T) {
	Enabled = true
	t.Cleanup(func() { Enabled = false })

	cases := []struct {
		label string
		code  string
	}{
		{"critical", "\x1b[31m"},
		{"low", "\x1b[33m"},
		{"high", "\x1b[33m"},
		{"ok", "\x1b[32m"},
	}
	for _, tc := range cases {
		got := StatusLabel(tc.label)
		if !strings.Contains(got, tc.code) {
			t.Errorf("StatusLabel(%q): missing code %q in %q", tc.label, tc.code, got)
		}
	}

	if got := StatusLabel("weird"); got != "weird" {
		t.Errorf("StatusLabel(\"weird\"): got %q, want unchanged", got)
	}
}

func TestInit(t *testing.T) {
	Init("always")
	if !Enabled {
		t.Error("Init(always): Enabled should be true")
	}

	Init("never")
	if Enabled {
		t.Error("Init(never): Enabled should be false")
	}

	t.Setenv("NO_COLOR", "1")
	Init("auto")
	if Enabled {
		t.Error("Init(auto) with NO_COLOR: Enabled should be false")
	}

	Enabled = false
}
