package breaker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sony/gobreaker"
	"github.com/yourusername/goaggregator/internal/domain"
	"github.com/yourusername/goaggregator/internal/flight"
	"github.com/yourusername/goaggregator/internal/provider"
)

func TestAggregatorSkipsHTTPProviderAfterCircuitOpens(t *testing.T) {
	var providerHits int32
	providerC := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&providerHits, 1)
		http.Error(w, "simulated failure", http.StatusInternalServerError)
	}))
	t.Cleanup(providerC.Close)

	httpProviders := provider.NewHTTPProviders([]string{providerC.URL}, 100*time.Millisecond)
	protectedProviders := NewFlightProvidersWithSettings(httpProviders, Settings{
		OpenCooldown:     time.Minute,
		HalfOpenRequests: 1,
	})
	service := flight.NewSearchService(protectedProviders, 100*time.Millisecond)
	request := domain.SearchRequest{From: "SGN", To: "HAN", Date: "2026-08-15"}

	for i := 0; i < 5; i++ {
		resp := service.Search(context.Background(), request)
		if resp.Meta.ProvidersCalled != 1 {
			t.Fatalf("call %d providers_called = %d, want 1", i+1, resp.Meta.ProvidersCalled)
		}
		if resp.Meta.ProvidersSucceeded != 0 {
			t.Fatalf("call %d providers_succeeded = %d, want 0", i+1, resp.Meta.ProvidersSucceeded)
		}
		if len(resp.Meta.ProvidersFailed) != 1 {
			t.Fatalf("call %d providers_failed = %d, want 1", i+1, len(resp.Meta.ProvidersFailed))
		}
	}

	if hits := atomic.LoadInt32(&providerHits); hits != 5 {
		t.Fatalf("provider HTTP hits after 5 calls = %d, want 5", hits)
	}

	breakerProvider, ok := protectedProviders[0].(*FlightProvider)
	if !ok {
		t.Fatalf("protected provider type = %T, want *breaker.FlightProvider", protectedProviders[0])
	}
	if breakerProvider.State() != gobreaker.StateOpen {
		t.Fatalf("breaker state = %s, want open", breakerProvider.State())
	}

	resp := service.Search(context.Background(), request)
	if resp.Meta.ProvidersCalled != 1 {
		t.Fatalf("6th call providers_called = %d, want 1", resp.Meta.ProvidersCalled)
	}
	if resp.Meta.ProvidersSucceeded != 0 {
		t.Fatalf("6th call providers_succeeded = %d, want 0", resp.Meta.ProvidersSucceeded)
	}
	if len(resp.Meta.ProvidersFailed) != 1 {
		t.Fatalf("6th call providers_failed = %d, want 1", len(resp.Meta.ProvidersFailed))
	}
	if len(resp.Results) != 0 {
		t.Fatalf("6th call results = %d, want 0", len(resp.Results))
	}
	if hits := atomic.LoadInt32(&providerHits); hits != 5 {
		t.Fatalf("provider HTTP hits after 6th call = %d, want still 5", hits)
	}
}