package output

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"gopkg.in/yaml.v3"
)

type Format string

const (
	Table Format = "table"
	JSON  Format = "json"
	YAML  Format = "yaml"
)

func ParseFormat(s string) Format {
	switch s {
	case "json":
		return JSON
	case "yaml":
		return YAML
	default:
		return Table
	}
}

type TableRow struct {
	Label string
	Value string
}

func Print(w io.Writer, format Format, data any, rows []TableRow) error {
	switch format {
	case JSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(stampedJSON(data))
	case YAML:
		return yaml.NewEncoder(w).Encode(stampedMap(data))
	default:
		return printTable(w, rows)
	}
}

// stampedJSON returns a map with run_at injected at the top level alongside
// the fields of data. If data cannot be marshaled to a map, it falls back to
// wrapping it under a "data" key.
func stampedJSON(data any) any {
	if data == nil {
		return map[string]any{"run_at": time.Now().Unix()}
	}
	b, err := json.Marshal(data)
	if err != nil {
		return data
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return data
	}
	m["run_at"] = time.Now().Unix()
	return m
}

// stampedMap returns an ordered-ish map for YAML with run_at at the top level.
func stampedMap(data any) any {
	if data == nil {
		return map[string]any{"run_at": time.Now().Unix()}
	}
	b, err := json.Marshal(data)
	if err != nil {
		return data
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return data
	}
	m["run_at"] = time.Now().Unix()
	return m
}

func printTable(w io.Writer, rows []TableRow) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, r := range rows {
		fmt.Fprintf(tw, "%s\t%s\n", r.Label, r.Value)
	}
	return tw.Flush()
}

func PrintList(w io.Writer, format Format, data any, headers []string, rows [][]string) error {
	switch format {
	case JSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(stampedJSON(data))
	case YAML:
		return yaml.NewEncoder(w).Encode(stampedMap(data))
	default:
		return printListTable(w, headers, rows)
	}
}

type LEDStatusData struct {
	White       int     `json:"white" yaml:"white"`
	Blue        int     `json:"blue" yaml:"blue"`
	Moon        int     `json:"moon" yaml:"moon"`
	Temperature float64 `json:"temperature" yaml:"temperature"`
}

func printListTable(w io.Writer, headers []string, rows [][]string) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for i, h := range headers {
		if i > 0 {
			fmt.Fprint(tw, "\t")
		}
		fmt.Fprint(tw, h)
	}
	fmt.Fprintln(tw)
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				fmt.Fprint(tw, "\t")
			}
			fmt.Fprint(tw, cell)
		}
		fmt.Fprintln(tw)
	}
	return tw.Flush()
}
