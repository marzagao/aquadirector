// Package color provides minimal ANSI colorization for CLI output.
//
// Only the 8 base foreground codes are used so output stays legible on both
// light and dark terminal themes. Color is gated on a TTY check, honors the
// NO_COLOR environment variable, and can be forced with --color=always|never.
package color

import "os"

// Enabled controls whether the wrapper functions emit ANSI codes. It is set
// once by Init and read by every wrapper.
var Enabled bool

const (
	reset  = "\x1b[0m"
	cBold  = "\x1b[1m"
	cDim   = "\x1b[2m"
	cRed   = "\x1b[31m"
	cGreen = "\x1b[32m"
	cYel   = "\x1b[33m"
	cBlue  = "\x1b[34m"
)

// Init sets Enabled based on the requested mode. mode is one of
// "auto" (default), "always", or "never". In auto mode, color is enabled only
// when stdout is a terminal and NO_COLOR is unset.
func Init(mode string) {
	switch mode {
	case "always":
		Enabled = true
	case "never":
		Enabled = false
	default:
		Enabled = isTTY(os.Stdout) && os.Getenv("NO_COLOR") == ""
	}
}

func isTTY(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func wrap(code, s string) string {
	if !Enabled {
		return s
	}
	return code + s + reset
}

// Bold wraps s in a bold SGR sequence.
func Bold(s string) string { return wrap(cBold, s) }

// Dim wraps s in a faint/dim SGR sequence. A small number of terminals
// ignore this attribute; in that case s renders at normal intensity.
func Dim(s string) string { return wrap(cDim, s) }

// Red wraps s in the base red foreground color (31).
func Red(s string) string { return wrap(cRed, s) }

// Green wraps s in the base green foreground color (32).
func Green(s string) string { return wrap(cGreen, s) }

// Yellow wraps s in the base yellow foreground color (33).
func Yellow(s string) string { return wrap(cYel, s) }

// Blue wraps s in the base blue foreground color (34).
func Blue(s string) string { return wrap(cBlue, s) }

// StatusLabel colorizes a pH/ORP status word: critical→red, low/high→yellow,
// ok→green. Unknown values are returned unchanged.
func StatusLabel(label string) string {
	switch label {
	case "critical":
		return Red(label)
	case "low", "high":
		return Yellow(label)
	case "ok":
		return Green(label)
	default:
		return label
	}
}
