package flight

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/yourusername/goaggregator/internal/domain"
)

type SearchService struct {
	providers []domain.Provider
	timeout   time.Duration
}

type SearchResponse struct {
	Results []domain.FlightResult `json:"results"`
	Meta    SearchMeta            `json:"meta"`
}

type SearchMeta struct {
	ProvidersCalled    int             `json:"providers_called"`
	ProvidersSucceeded int             `json:"providers_succeeded"`
	ProvidersFailed    []ProviderError `json:"providers_failed"`
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
	}
}

func (s *SearchService) Search(ctx context.Context, req domain.SearchRequest) SearchResponse {
	started := time.Now()
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
	return response
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
