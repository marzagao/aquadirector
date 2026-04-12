package alerts

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type MetricFetcher interface {
	Fetch(ctx context.Context, source, metric string) (any, error)
}

type Engine struct {
	rules     []Rule
	fetcher   MetricFetcher
	notifiers []Notifier
}

func NewEngine(rules []Rule, fetcher MetricFetcher, notifiers []Notifier) *Engine {
	return &Engine{
		rules:     rules,
		fetcher:   fetcher,
		notifiers: notifiers,
	}
}

func (e *Engine) Check(ctx context.Context) ([]AlertResult, error) {
	var results []AlertResult

	for _, rule := range e.rules {
		value, err := e.fetcher.Fetch(ctx, rule.Source, rule.Metric)
		if err != nil {
			results = append(results, AlertResult{
				Rule:      rule,
				Triggered: false,
				Timestamp: time.Now(),
				Message:   fmt.Sprintf("fetch error: %v", err),
			})
			continue
		}

		triggered, err := rule.Evaluate(value)
		if err != nil {
			results = append(results, AlertResult{
				Rule:      rule,
				Value:     value,
				Triggered: false,
				Timestamp: time.Now(),
				Message:   fmt.Sprintf("eval error: %v", err),
			})
			continue
		}

		msg := ""
		if triggered {
			msg = rule.FormatMessage(value)
		}

		results = append(results, AlertResult{
			Rule:      rule,
			Value:     value,
			Triggered: triggered,
			Timestamp: time.Now(),
			Message:   msg,
		})
	}

	return results, nil
}

func (e *Engine) Notify(ctx context.Context, results []AlertResult) error {
	var errs []string
	for _, result := range results {
		if !result.Triggered {
			continue
		}
		for _, n := range e.notifiers {
			if result.Rule.Severity < n.MinSeverity() {
				continue
			}
			if err := n.Notify(ctx, result); err != nil {
				errs = append(errs, fmt.Sprintf("%s: %v", result.Rule.Name, err))
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("notification errors: %s", strings.Join(errs, "; "))
	}
	return nil
}
