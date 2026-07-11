package limiter

import (
	"context"
	"testing"
	"time"

	"github.com/yourusername/goaggregator/internal/domain"
	"github.com/yourusername/goaggregator/internal/flight"
)

type countingProvider struct {
	name   string
	called int
}

func (p *countingProvider) Name() string {
	return p.name
}

func (p *countingProvider) Search(ctx context.Context, req domain.SearchRequest) (*domain.ProviderSearchResponse, error) {
	p.called++
	return &domain.ProviderSearchResponse{
		Provider: p.name,
		Results: []domain.FlightResult{{
			Provider: p.name,
			ID:       p.name + "-1",
			From:     req.From,
			To:       req.To,
			Date:     req.Date,
			Price:    1_000_000,
			Currency: "VND",
			Airline:  p.name,
		}},
	}, nil
}

func TestProviderRateLimiterSkipsRequestsOverQuota(t *testing.T) {
	base := &countingProvider{name: "provider-limited"}
	limited := NewFlightProviderWithSettings(base, Settings{RequestsPerSecond: 1, Burst: 1})

	_, err := limited.Search(context.Background(), domain.SearchRequest{})
	if err != nil {
		t.Fatalf("first request should pass: %v", err)
	}

	_, err = limited.Search(context.Background(), domain.SearchRequest{})
	if err == nil {
		t.Fatal("second immediate request should be rate limited")
	}
	if base.called != 1 {
		t.Fatalf("provider called = %d, want 1", base.called)
	}
}

func TestRateLimitedProviderDoesNotFailWholeFlightSearch(t *testing.T) {
	limitedBase := &countingProvider{name: "provider-limited"}
	healthyBase := &countingProvider{name: "provider-healthy"}
	limited := NewFlightProviderWithSettings(limitedBase, Settings{RequestsPerSecond: 1, Burst: 1})

	_, err := limited.Search(context.Background(), domain.SearchRequest{})
	if err != nil {
		t.Fatalf("preload limited provider token: %v", err)
	}

	service := flight.NewSearchService([]domain.Provider{limited, healthyBase}, 100*time.Millisecond)
	resp := service.Search(context.Background(), domain.SearchRequest{From: "SGN", To: "HAN", Date: "2026-08-15"})

	if resp.Meta.ProvidersCalled != 2 {
		t.Fatalf("providers called = %d, want 2", resp.Meta.ProvidersCalled)
	}
	if resp.Meta.ProvidersSucceeded != 1 {
		t.Fatalf("providers succeeded = %d, want 1", resp.Meta.ProvidersSucceeded)
	}
	if len(resp.Meta.ProvidersFailed) != 1 {
		t.Fatalf("providers failed = %d, want 1", len(resp.Meta.ProvidersFailed))
	}
	if len(resp.Results) != 1 || resp.Results[0].Provider != "provider-healthy" {
		t.Fatalf("expected healthy provider result only, got %+v", resp.Results)
	}
	if limitedBase.called != 1 {
		t.Fatalf("limited provider called = %d, want 1; over-quota search should be skipped", limitedBase.called)
	}
	if healthyBase.called != 1 {
		t.Fatalf("healthy provider called = %d, want 1", healthyBase.called)
	}
}