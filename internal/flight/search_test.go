package flight

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yourusername/goaggregator/internal/domain"
)

type fakeProvider struct {
	name    string
	delay   time.Duration
	err     error
	results []domain.FlightResult
}

func (p fakeProvider) Name() string {
	return p.name
}

func (p fakeProvider) Search(ctx context.Context, req domain.SearchRequest) (*domain.ProviderSearchResponse, error) {
	select {
	case <-time.After(p.delay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	if p.err != nil {
		return nil, p.err
	}
	if p.results != nil {
		return &domain.ProviderSearchResponse{
			Provider: p.name,
			Results:  p.results,
		}, nil
	}
	return &domain.ProviderSearchResponse{
		Provider: p.name,
		Results:  []domain.FlightResult{{ID: p.name + "-1", From: req.From, To: req.To, Date: req.Date, Price: 1_000_000}},
	}, nil
}

func TestSearchFansOutWithPerProviderTimeout(t *testing.T) {
	service := NewSearchService([]domain.Provider{
		fakeProvider{name: "fast", delay: 10 * time.Millisecond},
		fakeProvider{name: "slow", delay: 200 * time.Millisecond},
		fakeProvider{name: "failing", delay: 5 * time.Millisecond, err: errors.New("boom")},
	}, 50*time.Millisecond)

	started := time.Now()
	resp := service.Search(context.Background(), domain.SearchRequest{From: "SGN", To: "HAN", Date: "2026-08-15"})
	elapsed := time.Since(started)

	if elapsed >= 150*time.Millisecond {
		t.Fatalf("search waited too long for slow provider: %s", elapsed)
	}
	if resp.Meta.ProvidersCalled != 3 {
		t.Fatalf("providers called = %d, want 3", resp.Meta.ProvidersCalled)
	}
	if resp.Meta.ProvidersSucceeded != 1 {
		t.Fatalf("providers succeeded = %d, want 1", resp.Meta.ProvidersSucceeded)
	}
	if len(resp.Meta.ProvidersFailed) != 2 {
		t.Fatalf("providers failed = %d, want 2", len(resp.Meta.ProvidersFailed))
	}
	if len(resp.Results) != 1 || resp.Results[0].Provider != "fast" {
		t.Fatalf("unexpected results: %+v", resp.Results)
	}
}

func TestSearchMergesDedupesAndSortsResultsByPrice(t *testing.T) {
	service := NewSearchService([]domain.Provider{
		fakeProvider{
			name: "provider-a",
			results: []domain.FlightResult{
				{ID: "expensive", From: "SGN", To: "HAN", Date: "2026-08-15", Price: 2_000_000, Currency: "VND", Airline: "A"},
				{ID: "duplicate", From: "SGN", To: "HAN", Date: "2026-08-15", Price: 1_000_000, Currency: "VND", Airline: "B"},
			},
		},
		fakeProvider{
			name: "provider-b",
			results: []domain.FlightResult{
				{ID: "cheap", From: "SGN", To: "HAN", Date: "2026-08-15", Price: 500_000, Currency: "VND", Airline: "C"},
				{ID: "duplicate", From: "SGN", To: "HAN", Date: "2026-08-15", Price: 1_000_000, Currency: "VND", Airline: "B"},
			},
		},
	}, 50*time.Millisecond)

	resp := service.Search(context.Background(), domain.SearchRequest{From: "SGN", To: "HAN", Date: "2026-08-15"})

	if resp.Meta.ProvidersSucceeded != 2 {
		t.Fatalf("providers succeeded = %d, want 2", resp.Meta.ProvidersSucceeded)
	}
	if len(resp.Results) != 3 {
		t.Fatalf("results = %d, want 3: %+v", len(resp.Results), resp.Results)
	}
	wantIDs := []string{"cheap", "duplicate", "expensive"}
	for i, want := range wantIDs {
		if resp.Results[i].ID != want {
			t.Fatalf("result[%d].ID = %q, want %q; results=%+v", i, resp.Results[i].ID, want, resp.Results)
		}
	}
	if resp.Results[0].Provider != "provider-b" || resp.Results[1].Provider == "" || resp.Results[2].Provider != "provider-a" {
		t.Fatalf("provider attribution missing or unexpected: %+v", resp.Results)
	}
}
