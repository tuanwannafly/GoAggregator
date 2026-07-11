package hotel

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/yourusername/goaggregator/internal/cache"
	"github.com/yourusername/goaggregator/internal/domain"
)

type SearchService struct {
	providers []domain.HotelProvider
	timeout   time.Duration
	cache     cache.Cache
	cacheTTL  time.Duration
}

type SearchResponse struct {
	Results []domain.HotelResult `json:"results"`
	Meta    SearchMeta           `json:"meta"`
}

type SearchMeta struct {
	ProvidersCalled    int             `json:"providers_called"`
	ProvidersSucceeded int             `json:"providers_succeeded"`
	ProvidersFailed    []ProviderError `json:"providers_failed"`
	CacheHit           bool            `json:"cache_hit"`
	DurationMs         int64           `json:"duration_ms"`
}

type ProviderError struct {
	Provider string `json:"provider"`
	Error    string `json:"error"`
}

type providerResult struct {
	provider string
	response *domain.HotelProviderSearchResponse
	err      error
}

func NewSearchService(providers []domain.HotelProvider, timeout time.Duration) *SearchService {
	return &SearchService{
		providers: providers,
		timeout:   timeout,
		cache:     cache.NoopCache{},
	}
}

func (s *SearchService) WithCache(cache cache.Cache, ttl time.Duration) *SearchService {
	if cache != nil && ttl > 0 {
		s.cache = cache
		s.cacheTTL = ttl
	}
	return s
}

func (s *SearchService) Search(ctx context.Context, req domain.HotelSearchRequest) SearchResponse {
	started := time.Now()
	cacheKey := hotelCacheKey(req)
	if s.cache != nil && s.cacheTTL > 0 {
		var cached SearchResponse
		if err := s.cache.Get(ctx, cacheKey, &cached); err == nil {
			cached.Meta.ProvidersCalled = 0
			cached.Meta.ProvidersSucceeded = 0
			cached.Meta.ProvidersFailed = []ProviderError{}
			cached.Meta.CacheHit = true
			cached.Meta.DurationMs = time.Since(started).Milliseconds()
			return cached
		} else if !errors.Is(err, cache.ErrMiss) {
			// Cache is best-effort; provider search remains the source of truth.
			_ = err
		}
	}

	resultsCh := make(chan providerResult, len(s.providers))

	for _, p := range s.providers {
		provider := p
		go func() {
			providerCtx, cancel := context.WithTimeout(ctx, s.timeout)
			defer cancel()

			resp, err := provider.Search(providerCtx, req)
			resultsCh <- providerResult{
				provider: provider.Name(),
				response: resp,
				err:      err,
			}
		}()
	}

	response := SearchResponse{
		Results: make([]domain.HotelResult, 0),
		Meta: SearchMeta{
			ProvidersCalled: len(s.providers),
			ProvidersFailed: make([]ProviderError, 0),
		},
	}

	for i := 0; i < len(s.providers); i++ {
		select {
		case result := <-resultsCh:
			if result.err != nil {
				response.Meta.ProvidersFailed = append(response.Meta.ProvidersFailed, ProviderError{
					Provider: result.provider,
					Error:    result.err.Error(),
				})
				continue
			}
			if result.response == nil {
				response.Meta.ProvidersFailed = append(response.Meta.ProvidersFailed, ProviderError{
					Provider: result.provider,
					Error:    "empty response",
				})
				continue
			}
			response.Meta.ProvidersSucceeded++
			for _, r := range result.response.Results {
				response.Results = append(response.Results, domain.HotelResult{
					Provider:      result.provider,
					ID:            r.ID,
					City:          r.City,
					CheckIn:       r.CheckIn,
					CheckOut:      r.CheckOut,
					Name:          r.Name,
					PricePerNight: r.PricePerNight,
					Currency:      r.Currency,
				})
			}
		case <-ctx.Done():
			response.Meta.ProvidersFailed = append(response.Meta.ProvidersFailed, ProviderError{
				Provider: "context",
				Error:    ctx.Err().Error(),
			})
		}
	}

	response.Results = mergeSortHotelResults(response.Results)
	response.Meta.DurationMs = time.Since(started).Milliseconds()

	if s.cache != nil && s.cacheTTL > 0 && !response.Meta.CacheHit {
		_ = s.cache.Set(ctx, cacheKey, response, s.cacheTTL)
	}

	return response
}

func mergeSortHotelResults(results []domain.HotelResult) []domain.HotelResult {
	deduped := make([]domain.HotelResult, 0, len(results))
	seen := make(map[string]struct{}, len(results))
	for _, result := range results {
		key := hotelResultKey(result)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		deduped = append(deduped, result)
	}

	sort.SliceStable(deduped, func(i, j int) bool {
		if deduped[i].PricePerNight == deduped[j].PricePerNight {
			return deduped[i].ID < deduped[j].ID
		}
		return deduped[i].PricePerNight < deduped[j].PricePerNight
	})
	return deduped
}

func hotelResultKey(result domain.HotelResult) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s|%d|%s", result.ID, result.City, result.CheckIn, result.CheckOut, result.Name, result.PricePerNight, result.Currency)
}

func hotelCacheKey(req domain.HotelSearchRequest) string {
	return "hotel:" + req.City + ":" + req.CheckIn + ":" + req.CheckOut
}