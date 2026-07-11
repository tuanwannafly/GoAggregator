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

type FlightProvider struct {
	next    domain.Provider
	limiter *rate.Limiter
}

type HotelProvider struct {
	next    domain.HotelProvider
	limiter *rate.Limiter
}

func NewFlightProvider(next domain.Provider) *FlightProvider {
	return NewFlightProviderWithSettings(next, DefaultSettings())
}

func NewFlightProviderWithSettings(next domain.Provider, cfg Settings) *FlightProvider {
	cfg = cfg.withDefaults()
	return &FlightProvider{
		next:    next,
		limiter: rate.NewLimiter(rate.Limit(cfg.RequestsPerSecond), cfg.Burst),
	}
}

func NewHotelProvider(next domain.HotelProvider) *HotelProvider {
	return NewHotelProviderWithSettings(next, DefaultSettings())
}

func NewHotelProviderWithSettings(next domain.HotelProvider, cfg Settings) *HotelProvider {
	cfg = cfg.withDefaults()
	return &HotelProvider{
		next:    next,
		limiter: rate.NewLimiter(rate.Limit(cfg.RequestsPerSecond), cfg.Burst),
	}
}

func NewFlightProviders(providers []domain.Provider) []domain.Provider {
	return NewFlightProvidersWithSettings(providers, DefaultSettings())
}

func NewFlightProvidersWithSettings(providers []domain.Provider, cfg Settings) []domain.Provider {
	wrapped := make([]domain.Provider, 0, len(providers))
	for _, provider := range providers {
		wrapped = append(wrapped, NewFlightProviderWithSettings(provider, cfg))
	}
	return wrapped
}

func NewHotelProviders(providers []domain.HotelProvider) []domain.HotelProvider {
	return NewHotelProvidersWithSettings(providers, DefaultSettings())
}

func NewHotelProvidersWithSettings(providers []domain.HotelProvider, cfg Settings) []domain.HotelProvider {
	wrapped := make([]domain.HotelProvider, 0, len(providers))
	for _, provider := range providers {
		wrapped = append(wrapped, NewHotelProviderWithSettings(provider, cfg))
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

func (p *FlightProvider) Name() string {
	return p.next.Name()
}

func (p *FlightProvider) Search(ctx context.Context, req domain.SearchRequest) (*domain.ProviderSearchResponse, error) {
	if !p.limiter.Allow() {
		return nil, fmt.Errorf("provider %q rate limited", p.Name())
	}
	return p.next.Search(ctx, req)
}

func (p *HotelProvider) Name() string {
	return p.next.Name()
}

func (p *HotelProvider) Search(ctx context.Context, req domain.HotelSearchRequest) (*domain.HotelProviderSearchResponse, error) {
	if !p.limiter.Allow() {
		return nil, fmt.Errorf("provider %q rate limited", p.Name())
	}
	return p.next.Search(ctx, req)
}