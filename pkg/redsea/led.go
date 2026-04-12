package redsea

import (
	"context"
	"fmt"
)

type LEDClient struct {
	*Client
}

func NewLEDClient(ip string, opts ...Option) *LEDClient {
	return &LEDClient{Client: New(ip, opts...)}
}

func (l *LEDClient) ManualState(ctx context.Context) (*LEDManualState, error) {
	var state LEDManualState
	if err := l.Get(ctx, "/manual", &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func (l *LEDClient) SetManual(ctx context.Context, set LEDManualSet) error {
	return l.Post(ctx, "/manual", set, nil)
}

func (l *LEDClient) SetTimer(ctx context.Context, set LEDTimerSet) error {
	return l.Post(ctx, "/timer", set, nil)
}

func (l *LEDClient) Schedule(ctx context.Context, day int) (*LEDSchedule, error) {
	if day < 1 || day > 7 {
		return nil, fmt.Errorf("day must be 1-7, got %d", day)
	}
	var sched LEDSchedule
	if err := l.Get(ctx, fmt.Sprintf("/auto/%d", day), &sched); err != nil {
		return nil, err
	}
	return &sched, nil
}

func (l *LEDClient) SetSchedule(ctx context.Context, day int, sched *LEDSchedule) error {
	if day < 1 || day > 7 {
		return fmt.Errorf("day must be 1-7, got %d", day)
	}
	return l.Put(ctx, fmt.Sprintf("/auto/%d", day), sched, nil)
}

func (l *LEDClient) Acclimation(ctx context.Context) (*LEDAcclimation, error) {
	var accl LEDAcclimation
	if err := l.Get(ctx, "/acclimation", &accl); err != nil {
		return nil, err
	}
	return &accl, nil
}

func (l *LEDClient) SetAcclimation(ctx context.Context, accl *LEDAcclimation) error {
	return l.Put(ctx, "/acclimation", accl, nil)
}

func (l *LEDClient) MoonPhase(ctx context.Context) (*LEDMoonPhase, error) {
	var mp LEDMoonPhase
	if err := l.Get(ctx, "/moonphase", &mp); err != nil {
		return nil, err
	}
	return &mp, nil
}

func (l *LEDClient) Dashboard(ctx context.Context) (*LEDDashboard, error) {
	var dash LEDDashboard
	if err := l.Get(ctx, "/dashboard", &dash); err != nil {
		return nil, err
	}
	return &dash, nil
}

func (l *LEDClient) Mode(ctx context.Context) (string, error) {
	return l.Client.Mode(ctx)
}

func (l *LEDClient) SetMode(ctx context.Context, mode string) error {
	return l.Client.SetMode(ctx, mode)
}
