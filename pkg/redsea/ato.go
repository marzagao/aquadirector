package redsea

import "context"

type ATOClient struct {
	*Client
}

func NewATOClient(ip string, opts ...Option) *ATOClient {
	return &ATOClient{Client: New(ip, opts...)}
}

func (a *ATOClient) Dashboard(ctx context.Context) (*ATODashboard, error) {
	var dash ATODashboard
	if err := a.Get(ctx, "/dashboard", &dash); err != nil {
		return nil, err
	}
	return &dash, nil
}

func (a *ATOClient) Configuration(ctx context.Context) (*ATOConfiguration, error) {
	var cfg ATOConfiguration
	if err := a.Get(ctx, "/configuration", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (a *ATOClient) SetConfiguration(ctx context.Context, update map[string]any) error {
	return a.Put(ctx, "/configuration", update, nil)
}

func (a *ATOClient) Resume(ctx context.Context) error {
	return a.Post(ctx, "/resume", nil, nil)
}

func (a *ATOClient) SetVolume(ctx context.Context, volumeML int) error {
	payload := map[string]int{"volume": volumeML}
	return a.Post(ctx, "/update-volume", payload, nil)
}

func (a *ATOClient) Mode(ctx context.Context) (string, error) {
	return a.Client.Mode(ctx)
}

func (a *ATOClient) SetMode(ctx context.Context, mode string) error {
	return a.Client.SetMode(ctx, mode)
}
