package hotel

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/yourusername/goaggregator/internal/cache"
	"github.com/yourusername/goaggregator/internal/domain"
)

type fakeHotelProvider struct {
	name    string
	delay   time.Duration
	err     error
	results []domain.HotelResult
	called  int
}

func (p *fakeHotelProvider) Name() string {
	return p.name
}

func (p *fakeHotelProvider) Search(ctx context.Context, req domain.HotelSearchRequest) (*domain.HotelProviderSearchResponse, error) {
	p.called++
	select {
	case <-time.After(p.delay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	if p.err != nil {
		return nil, p.err
	}
	if p.results != nil {
		return &domain.HotelProviderSearchResponse{
			Provider: p.name,
			Results:  p.results,
		}, nil
	}
	return &domain.HotelProviderSearchResponse{
		Provider: p.name,
		Results: []domain.HotelResult{{
			ID:            p.name + "-1",
			City:          req.City,
			CheckIn:       req.CheckIn,
			CheckOut:      req.CheckOut,
			Name:          "Test Hotel",
			PricePerNight: 1_000_000,
			Currency:      "VND",
		}},
	}, nil
}

func TestHotelSearchFansOutWithPerProviderTimeout(t *testing.T) {
	fast := &fakeHotelProvider{name: "fast", delay: 10 * time.Millisecond}
	slow := &fakeHotelProvider{name: "slow", delay: 200 * time.Millisecond}
	failing := &fakeHotelProvider{name: "failing", delay: 5 * time.Millisecond, err: errors.New("boom")}
	service := NewSearchService([]domain.HotelProvider{
		fast,
		slow,
		failing,
	}, 50*time.Millisecond)

	started := time.Now()
	resp := service.Search(context.Background(), domain.HotelSearchRequest{City: "DAD", CheckIn: "2026-08-15", CheckOut: "2026-08-17"})
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

func TestHotelSearchMergesDedupesAndSortsResultsByPrice(t *testing.T) {
	service := NewSearchService([]domain.HotelProvider{
		&fakeHotelProvider{
			name: "provider-a",
			results: []domain.HotelResult{
				{ID: "expensive", City: "DAD", CheckIn: "2026-08-15", CheckOut: "2026-08-17", PricePerNight: 2_000_000, Currency: "VND", Name: "A"},
				{ID: "duplicate", City: "DAD", CheckIn: "2026-08-15", CheckOut: "2026-08-17", PricePerNight: 1_000_000, Currency: "VND", Name: "B"},
			},
		},
		&fakeHotelProvider{
			name: "provider-b",
			results: []domain.HotelResult{
				{ID: "cheap", City: "DAD", CheckIn: "2026-08-15", CheckOut: "2026-08-17", PricePerNight: 500_000, Currency: "VND", Name: "C"},
				{ID: "duplicate", City: "DAD", CheckIn: "2026-08-15", CheckOut: "2026-08-17", PricePerNight: 1_000_000, Currency: "VND", Name: "B"},
			},
		},
	}, 50*time.Millisecond)

	resp := service.Search(context.Background(), domain.HotelSearchRequest{City: "DAD", CheckIn: "2026-08-15", CheckOut: "2026-08-17"})

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

type memoryCache struct {
	values map[string]SearchResponse
	sets   int
}

func newMemoryCache() *memoryCache {
	return &memoryCache{values: make(map[string]SearchResponse)}
}

func (c *memoryCache) Get(ctx context.Context, key string, dest any) error {
	value, ok := c.values[key]
	if !ok {
		return cache.ErrMiss
	}
	ptr, ok := dest.(*SearchResponse)
	if !ok {
		return errors.New("unexpected cache destination")
	}
	*ptr = value
	return nil
}

func (c *memoryCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	searchResp, ok := value.(SearchResponse)
	if !ok {
		return errors.New("unexpected cache value")
	}
	c.values[key] = searchResp
	c.sets++
	return nil
}

func TestHotelSearchUsesCacheAsideOnMissThenHit(t *testing.T) {
	provider := &fakeHotelProvider{name: "provider-cache"}
	cacheStore := newMemoryCache()
	service := NewSearchService([]domain.HotelProvider{provider}, 50*time.Millisecond).WithCache(cacheStore, time.Minute)
	req := domain.HotelSearchRequest{City: "DAD", CheckIn: "2026-08-15", CheckOut: "2026-08-17"}

	first := service.Search(context.Background(), req)
	if first.Meta.CacheHit {
		t.Fatal("first response should be a cache miss")
	}
	if provider.called != 1 {
		t.Fatalf("provider called = %d, want 1", provider.called)
	}
	if cacheStore.sets != 1 {
		t.Fatalf("cache sets = %d, want 1", cacheStore.sets)
	}

	second := service.Search(context.Background(), req)
	if !second.Meta.CacheHit {
		t.Fatal("second response should be a cache hit")
	}
	if provider.called != 1 {
		t.Fatalf("provider called = %d, want still 1", provider.called)
	}
	if second.Meta.ProvidersCalled != 0 || second.Meta.ProvidersSucceeded != 0 || len(second.Meta.ProvidersFailed) != 0 {
		t.Fatalf("cache hit should not call providers, meta=%+v", second.Meta)
	}
	if !reflect.DeepEqual(first.Results, second.Results) {
		t.Fatalf("cached results mismatch: first=%+v second=%+v", first.Results, second.Results)
	}
}