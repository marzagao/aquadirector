package alerts

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"text/template"
)

type Notifier interface {
	Notify(ctx context.Context, alert AlertResult) error
	MinSeverity() Severity
}

// StdoutNotifier prints alerts to stderr.
type StdoutNotifier struct {
	minSeverity Severity
}

func NewStdoutNotifier(minSeverity Severity) *StdoutNotifier {
	return &StdoutNotifier{minSeverity: minSeverity}
}

func (n *StdoutNotifier) MinSeverity() Severity { return n.minSeverity }

func (n *StdoutNotifier) Notify(_ context.Context, alert AlertResult) error {
	prefix := "[" + alert.Rule.Severity.String() + "]"
	fmt.Fprintf(os.Stderr, "%s %s: %s\n", prefix, alert.Rule.Name, alert.Message)
	return nil
}

// WebhookNotifier sends HTTP requests.
type WebhookNotifier struct {
	minSeverity  Severity
	url          string
	method       string
	headers      map[string]string
	bodyTemplate string
}

func NewWebhookNotifier(minSeverity Severity, url, method string, headers map[string]string, bodyTemplate string) *WebhookNotifier {
	if method == "" {
		method = "POST"
	}
	return &WebhookNotifier{
		minSeverity:  minSeverity,
		url:          url,
		method:       method,
		headers:      headers,
		bodyTemplate: bodyTemplate,
	}
}

func (n *WebhookNotifier) MinSeverity() Severity { return n.minSeverity }

func (n *WebhookNotifier) Notify(ctx context.Context, alert AlertResult) error {
	body, err := renderTemplate(n.bodyTemplate, alert)
	if err != nil {
		return fmt.Errorf("rendering body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, n.method, n.url, bytes.NewReader([]byte(body)))
	if err != nil {
		return err
	}
	for k, v := range n.headers {
		req.Header.Set(k, v)
	}
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return nil
}

// CommandNotifier executes a shell command.
type CommandNotifier struct {
	minSeverity Severity
	command     string
	args        []string
}

func NewCommandNotifier(minSeverity Severity, command string, args []string) *CommandNotifier {
	return &CommandNotifier{
		minSeverity: minSeverity,
		command:     command,
		args:        args,
	}
}

func (n *CommandNotifier) MinSeverity() Severity { return n.minSeverity }

func (n *CommandNotifier) Notify(ctx context.Context, alert AlertResult) error {
	var renderedArgs []string
	for _, arg := range n.args {
		rendered, err := renderTemplate(arg, alert)
		if err != nil {
			renderedArgs = append(renderedArgs, arg)
		} else {
			renderedArgs = append(renderedArgs, rendered)
		}
	}

	cmd := exec.CommandContext(ctx, n.command, renderedArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func renderTemplate(tmplStr string, alert AlertResult) (string, error) {
	tmpl, err := template.New("alert").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	data := map[string]any{
		"Name":      alert.Rule.Name,
		"Value":     alert.Value,
		"Threshold": alert.Rule.Threshold,
		"Severity":  alert.Rule.Severity.String(),
		"Message":   alert.Message,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
