package breaker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/sony/gobreaker"
	"github.com/yourusername/goaggregator/internal/domain"
	"github.com/yourusername/goaggregator/internal/requestid"
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

type FlightProvider struct {
	next    domain.Provider
	breaker *gobreaker.CircuitBreaker
}

type HotelProvider struct {
	next    domain.HotelProvider
	breaker *gobreaker.CircuitBreaker
}

func NewFlightProvider(next domain.Provider) *FlightProvider {
	return NewFlightProviderWithSettings(next, DefaultSettings())
}

func NewFlightProviderWithSettings(next domain.Provider, cfg Settings) *FlightProvider {
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

	return &FlightProvider{
		next:    next,
		breaker: gobreaker.NewCircuitBreaker(settings),
	}
}

func NewHotelProvider(next domain.HotelProvider) *HotelProvider {
	return NewHotelProviderWithSettings(next, DefaultSettings())
}

func NewHotelProviderWithSettings(next domain.HotelProvider, cfg Settings) *HotelProvider {
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

	return &HotelProvider{
		next:    next,
		breaker: gobreaker.NewCircuitBreaker(settings),
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

func (p *FlightProvider) Name() string {
	return p.next.Name()
}

func (p *FlightProvider) Search(ctx context.Context, req domain.SearchRequest) (*domain.ProviderSearchResponse, error) {
	started := time.Now()
	cbState := p.breaker.State()

	result, err := p.breaker.Execute(func() (interface{}, error) {
		return p.next.Search(ctx, req)
	})

	status := "success"
	if err != nil {
		status = "error"
	}

	slog.Info("provider call",
		"request_id", requestid.FromContext(ctx),
		"provider", p.Name(),
		"duration_ms", time.Since(started).Milliseconds(),
		"status", status,
		"circuit_state", cbState.String(),
	)

	if err != nil {
		return nil, fmt.Errorf("provider %q circuit breaker: %w", p.Name(), err)
	}

	searchResp, ok := result.(*domain.ProviderSearchResponse)
	if !ok {
		return nil, fmt.Errorf("provider %q circuit breaker returned unexpected response", p.Name())
	}
	return searchResp, nil
}

func (p *FlightProvider) State() gobreaker.State {
	return p.breaker.State()
}

func (p *HotelProvider) Name() string {
	return p.next.Name()
}

func (p *HotelProvider) Search(ctx context.Context, req domain.HotelSearchRequest) (*domain.HotelProviderSearchResponse, error) {
	started := time.Now()
	cbState := p.breaker.State()

	result, err := p.breaker.Execute(func() (interface{}, error) {
		return p.next.Search(ctx, req)
	})

	status := "success"
	if err != nil {
		status = "error"
	}

	slog.Info("provider call",
		"request_id", requestid.FromContext(ctx),
		"provider", p.Name(),
		"duration_ms", time.Since(started).Milliseconds(),
		"status", status,
		"circuit_state", cbState.String(),
	)

	if err != nil {
		return nil, fmt.Errorf("provider %q circuit breaker: %w", p.Name(), err)
	}

	searchResp, ok := result.(*domain.HotelProviderSearchResponse)
	if !ok {
		return nil, fmt.Errorf("provider %q circuit breaker returned unexpected response", p.Name())
	}
	return searchResp, nil
}

func (p *HotelProvider) State() gobreaker.State {
	return p.breaker.State()
}