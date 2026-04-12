package eheim

import (
	"context"
	"fmt"
)

// FeederClient controls an Eheim autofeeder+ device via the REST API.
type FeederClient struct {
	*Client
	mac string
}

// NewFeederClient creates a client for a specific autofeeder on the Eheim mesh.
func NewFeederClient(host, mac string, opts ...Option) *FeederClient {
	return &FeederClient{
		Client: New(host, opts...),
		mac:    mac,
	}
}

// MAC returns the target feeder's MAC address.
func (f *FeederClient) MAC() string {
	return f.mac
}

// Status returns the current autofeeder state.
// GET /api/autofeeder?to=MAC
func (f *FeederClient) Status(ctx context.Context) (*FeederData, error) {
	var wire feederDataWire
	query := f.toQuery()
	if err := f.Get(ctx, "/api/autofeeder", query, &wire); err != nil {
		return nil, fmt.Errorf("fetching feeder status: %w", err)
	}
	return wireToFeederData(&wire)
}

// Feed triggers a single manual feeding.
// POST /api/autofeeder/feed
func (f *FeederClient) Feed(ctx context.Context) error {
	return f.Post(ctx, "/api/autofeeder/feed", f.toPayload())
}

// MarkDrumFull marks the food drum as full (resets fill level tracking).
// POST /api/autofeeder/full
func (f *FeederClient) MarkDrumFull(ctx context.Context) error {
	return f.Post(ctx, "/api/autofeeder/full", f.toPayload())
}

// SetSchedule sets the feeding schedule and feeding break option.
// POST /api/autofeeder/bio
func (f *FeederClient) SetSchedule(ctx context.Context, schedule WeekSchedule, feedingBreak bool) error {
	payload := map[string]any{
		"to":            f.mac,
		"configuration": scheduleToWire(schedule),
		"feedingBreak":  boolToInt(feedingBreak),
	}
	return f.Post(ctx, "/api/autofeeder/bio", payload)
}

// SetConfig sets overfeeding protection and filter sync.
// POST /api/autofeeder/config
func (f *FeederClient) SetConfig(ctx context.Context, update FeederConfigUpdate) error {
	// Fetch current state to preserve unchanged fields
	current, err := f.Status(ctx)
	if err != nil {
		return fmt.Errorf("fetching current config: %w", err)
	}

	overfeeding := current.Overfeeding
	if update.Overfeeding != nil {
		overfeeding = *update.Overfeeding
	}
	syncFilter := current.SyncFilter
	if update.SyncFilter != nil {
		syncFilter = *update.SyncFilter
	}

	payload := map[string]any{
		"to":          f.mac,
		"overfeeding": boolToInt(overfeeding),
		"sync":        syncFilter,
	}
	return f.Post(ctx, "/api/autofeeder/config", payload)
}

func (f *FeederClient) toQuery() map[string]string {
	if f.mac == "" {
		return nil
	}
	return map[string]string{"to": f.mac}
}

func (f *FeederClient) toPayload() map[string]string {
	return map[string]string{"to": f.mac}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
