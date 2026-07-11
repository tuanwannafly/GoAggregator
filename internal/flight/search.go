package flight

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/yourusername/goaggregator/internal/cache"
	"github.com/yourusername/goaggregator/internal/domain"
	"github.com/yourusername/goaggregator/internal/requestid"
)

type SearchService struct {
	providers []domain.Provider
	timeout   time.Duration
	cache     cache.Cache
	cacheTTL  time.Duration
}

type SearchResponse struct {
	Results []domain.FlightResult `json:"results"`
	Meta    SearchMeta            `json:"meta"`
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
	response *domain.ProviderSearchResponse
	err      error
}

func NewSearchService(providers []domain.Provider, timeout time.Duration) *SearchService {
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

func (s *SearchService) Search(ctx context.Context, req domain.SearchRequest) SearchResponse {
	ctx = requestid.NewContext(ctx)
	started := time.Now()
	cacheKey := flightCacheKey(req)
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
		Results: make([]domain.FlightResult, 0),
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
					Error:    "empty provider response",
				})
				continue
			}
			response.Results = append(response.Results, result.response.Results...)
			for i := len(response.Results) - len(result.response.Results); i < len(response.Results); i++ {
				if response.Results[i].Provider == "" {
					response.Results[i].Provider = result.response.Provider
				}
			}
		case <-ctx.Done():
			response.Meta.ProvidersFailed = append(response.Meta.ProvidersFailed, ProviderError{
				Provider: "request",
				Error:    fmt.Sprintf("request cancelled: %v", ctx.Err()),
			})
			response.Meta.DurationMs = time.Since(started).Milliseconds()
			return response
		}
	}

	response.Results = mergeSortResults(response.Results)
	response.Meta.ProvidersSucceeded = response.Meta.ProvidersCalled - len(response.Meta.ProvidersFailed)
	response.Meta.DurationMs = time.Since(started).Milliseconds()
	if s.cache != nil && s.cacheTTL > 0 {
		_ = s.cache.Set(ctx, cacheKey, response, s.cacheTTL)
	}
	return response
}

func flightCacheKey(req domain.SearchRequest) string {
	return fmt.Sprintf("flight:%s:%s:%s", strings.ToUpper(req.From), strings.ToUpper(req.To), req.Date)
}

func mergeSortResults(results []domain.FlightResult) []domain.FlightResult {
	deduped := make([]domain.FlightResult, 0, len(results))
	seen := make(map[string]struct{}, len(results))
	for _, result := range results {
		key := resultKey(result)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		deduped = append(deduped, result)
	}

	sort.SliceStable(deduped, func(i, j int) bool {
		if deduped[i].Price == deduped[j].Price {
			return deduped[i].ID < deduped[j].ID
		}
		return deduped[i].Price < deduped[j].Price
	})
	return deduped
}

func resultKey(result domain.FlightResult) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s|%d|%s", result.ID, result.From, result.To, result.Date, result.Airline, result.Price, result.Currency)
}
