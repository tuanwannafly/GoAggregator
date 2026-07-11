package limiter

import (
	"context"
	"fmt"

	"github.com/yourusername/goaggregator/internal/domain"
	"golang.org/x/time/rate"
)

const (
	defaultRequestsPerSecond = 10
	defaultBurst             = 10
)

type Settings struct {
	RequestsPerSecond float64
	Burst             int
}

type Provider struct {
	next    domain.Provider
	limiter *rate.Limiter
}

func NewProvider(next domain.Provider) *Provider {
	return NewProviderWithSettings(next, DefaultSettings())
}

func NewProviderWithSettings(next domain.Provider, cfg Settings) *Provider {
	cfg = cfg.withDefaults()
	return &Provider{
		next:    next,
		limiter: rate.NewLimiter(rate.Limit(cfg.RequestsPerSecond), cfg.Burst),
	}
}

func NewProviders(providers []domain.Provider) []domain.Provider {
	return NewProvidersWithSettings(providers, DefaultSettings())
}

func NewProvidersWithSettings(providers []domain.Provider, cfg Settings) []domain.Provider {
	wrapped := make([]domain.Provider, 0, len(providers))
	for _, provider := range providers {
		wrapped = append(wrapped, NewProviderWithSettings(provider, cfg))
	}
	return wrapped
}

func DefaultSettings() Settings {
	return Settings{
		RequestsPerSecond: defaultRequestsPerSecond,
		Burst:             defaultBurst,
	}
}

func (s Settings) withDefaults() Settings {
	defaults := DefaultSettings()
	if s.RequestsPerSecond <= 0 {
		s.RequestsPerSecond = defaults.RequestsPerSecond
	}
	if s.Burst <= 0 {
		s.Burst = defaults.Burst
	}
	return s
}

func (p *Provider) Name() string {
	return p.next.Name()
}

func (p *Provider) Search(ctx context.Context, req domain.SearchRequest) (*domain.ProviderSearchResponse, error) {
	if !p.limiter.Allow() {
		return nil, fmt.Errorf("provider %q rate limited", p.Name())
	}
	return p.next.Search(ctx, req)
}
