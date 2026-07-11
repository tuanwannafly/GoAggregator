package flight

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/yourusername/goaggregator/internal/domain"
	"github.com/yourusername/goaggregator/internal/provider"
)

type controllableProvider struct {
	server *httptest.Server
	mu     sync.RWMutex
	name   string
	delay  time.Duration
}

type controlRequest struct {
	LatencyMs int `json:"latency_ms"`
}

func newControllableProvider(t *testing.T, name string) *controllableProvider {
	t.Helper()

	p := &controllableProvider{name: name}
	mux := http.NewServeMux()
	mux.HandleFunc("/search", p.handleSearch)
	mux.HandleFunc("/control", p.handleControl)
	p.server = httptest.NewServer(mux)
	t.Cleanup(p.server.Close)
	return p
}

func (p *controllableProvider) URL() string {
	return p.server.URL
}

func (p *controllableProvider) handleSearch(w http.ResponseWriter, r *http.Request) {
	p.mu.RLock()
	delay := p.delay
	p.mu.RUnlock()

	if delay > 0 {
		select {
		case <-time.After(delay):
		case <-r.Context().Done():
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(domain.ProviderSearchResponse{
		Provider: p.name,
		Results: []domain.FlightResult{{
			ID:       p.name + "-1",
			From:     r.URL.Query().Get("from"),
			To:       r.URL.Query().Get("to"),
			Date:     r.URL.Query().Get("date"),
			Price:    1_000_000,
			Currency: "VND",
			Airline:  p.name,
		}},
	})
}

func (p *controllableProvider) handleControl(w http.ResponseWriter, r *http.Request) {
	var req controlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid control body", http.StatusBadRequest)
		return
	}

	p.mu.Lock()
	p.delay = time.Duration(req.LatencyMs) * time.Millisecond
	p.mu.Unlock()
	w.WriteHeader(http.StatusOK)
}

func TestSearchTimeoutIsolationWithHTTPProviders(t *testing.T) {
	fastProvider := newControllableProvider(t, "provider-fast")
	slowProvider := newControllableProvider(t, "provider-slow")

	controlBody := `{"latency_ms":5000}`
	resp, err := http.Post(slowProvider.URL()+"/control", "application/json", strings.NewReader(controlBody))
	if err != nil {
		t.Fatalf("set slow provider latency: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("set slow provider latency status = %d, want 200", resp.StatusCode)
	}

	providers := provider.NewHTTPProviders([]string{fastProvider.URL(), slowProvider.URL()}, 3*time.Second)
	service := NewSearchService(providers, 3*time.Second)

	started := time.Now()
	searchResp := service.Search(context.Background(), domain.SearchRequest{From: "SGN", To: "HAN", Date: "2026-08-15"})
	elapsed := time.Since(started)

	if elapsed < 3*time.Second || elapsed >= 4*time.Second {
		t.Fatalf("search elapsed = %s, want around 3s and below slow provider latency", elapsed)
	}
	if searchResp.Meta.ProvidersCalled != 2 {
		t.Fatalf("providers called = %d, want 2", searchResp.Meta.ProvidersCalled)
	}
	if searchResp.Meta.ProvidersSucceeded != 1 {
		t.Fatalf("providers succeeded = %d, want 1", searchResp.Meta.ProvidersSucceeded)
	}
	if len(searchResp.Meta.ProvidersFailed) != 1 {
		t.Fatalf("providers failed = %d, want 1", len(searchResp.Meta.ProvidersFailed))
	}
	if len(searchResp.Results) != 1 || searchResp.Results[0].Provider != "provider-fast" {
		t.Fatalf("expected only fast provider results, got %+v", searchResp.Results)
	}
}
