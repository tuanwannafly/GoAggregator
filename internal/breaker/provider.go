package breaker

import (
	"context"
	"fmt"
	"time"

	"github.com/sony/gobreaker"
	"github.com/yourusername/goaggregator/internal/domain"
)

const (
	defaultMinRequests       = 5
	defaultFailureThreshold  = 0.60
	defaultOpenStateCooldown = 30 * time.Second
	defaultHalfOpenRequests  = 1
)

type Settings struct {
	MinRequests      uint32
	FailureThreshold float64
	OpenCooldown     time.Duration
	HalfOpenRequests uint32
}

type Provider struct {
	next    domain.Provider
	breaker *gobreaker.CircuitBreaker
}

func NewProvider(next domain.Provider) *Provider {
	return NewProviderWithSettings(next, DefaultSettings())
}

func NewProviderWithSettings(next domain.Provider, cfg Settings) *Provider {
	cfg = cfg.withDefaults()
	settings := gobreaker.Settings{
		Name:        next.Name(),
		Timeout:     cfg.OpenCooldown,
		MaxRequests: cfg.HalfOpenRequests,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			if counts.Requests < cfg.MinRequests {
				return false
			}
			failureRate := float64(counts.TotalFailures) / float64(counts.Requests)
			return failureRate >= cfg.FailureThreshold
		},
	}

	return &Provider{
		next:    next,
		breaker: gobreaker.NewCircuitBreaker(settings),
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
		MinRequests:      defaultMinRequests,
		FailureThreshold: defaultFailureThreshold,
		OpenCooldown:     defaultOpenStateCooldown,
		HalfOpenRequests: defaultHalfOpenRequests,
	}
}

func (s Settings) withDefaults() Settings {
	defaults := DefaultSettings()
	if s.MinRequests == 0 {
		s.MinRequests = defaults.MinRequests
	}
	if s.FailureThreshold <= 0 {
		s.FailureThreshold = defaults.FailureThreshold
	}
	if s.OpenCooldown <= 0 {
		s.OpenCooldown = defaults.OpenCooldown
	}
	if s.HalfOpenRequests == 0 {
		s.HalfOpenRequests = defaults.HalfOpenRequests
	}
	return s
}

func (p *Provider) Name() string {
	return p.next.Name()
}

func (p *Provider) Search(ctx context.Context, req domain.SearchRequest) (*domain.ProviderSearchResponse, error) {
	result, err := p.breaker.Execute(func() (interface{}, error) {
		return p.next.Search(ctx, req)
	})
	if err != nil {
		return nil, fmt.Errorf("provider %q circuit breaker: %w", p.Name(), err)
	}

	searchResp, ok := result.(*domain.ProviderSearchResponse)
	if !ok {
		return nil, fmt.Errorf("provider %q circuit breaker returned unexpected response", p.Name())
	}
	return searchResp, nil
}

func (p *Provider) State() gobreaker.State {
	return p.breaker.State()
}
