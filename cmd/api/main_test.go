package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/goaggregator/internal/domain"
	"github.com/yourusername/goaggregator/internal/flight"
	"github.com/yourusername/goaggregator/internal/hotel"
)

type apiTestProvider struct {
	lastReq domain.SearchRequest
}

func (p *apiTestProvider) Name() string {
	return "api-test-provider"
}

func (p *apiTestProvider) Search(ctx context.Context, req domain.SearchRequest) (*domain.ProviderSearchResponse, error) {
	p.lastReq = req
	return &domain.ProviderSearchResponse{
		Provider: p.Name(),
		Results: []domain.FlightResult{{
			ID:       "flight-1",
			From:     req.From,
			To:       req.To,
			Date:     req.Date,
			Price:    1_000_000,
			Currency: "VND",
			Airline:  "API Test Air",
		}},
	}, nil
}

type apiTestHotelProvider struct {
	lastReq domain.HotelSearchRequest
}

func (p *apiTestHotelProvider) Name() string {
	return "api-test-hotel-provider"
}

func (p *apiTestHotelProvider) Search(ctx context.Context, req domain.HotelSearchRequest) (*domain.HotelProviderSearchResponse, error) {
	p.lastReq = req
	return &domain.HotelProviderSearchResponse{
		Provider: p.Name(),
		Results: []domain.HotelResult{{
			Provider:      p.Name(),
			ID:            "hotel-1",
			City:          req.City,
			CheckIn:       req.CheckIn,
			CheckOut:      req.CheckOut,
			Name:          "API Test Hotel",
			PricePerNight: 500_000,
			Currency:      "VND",
		}},
	}, nil
}

func TestSearchFlightsHandlerReturnsResults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &apiTestProvider{}
	router := newRouter(flight.NewSearchService([]domain.Provider{provider}, 100*time.Millisecond), hotel.NewSearchService([]domain.HotelProvider{}, 100*time.Millisecond))

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/search/flights?from=SGN&to=HAN&date=2026-08-15", nil)
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", recorder.Code, recorder.Body.String())
	}
	if provider.lastReq != (domain.SearchRequest{From: "SGN", To: "HAN", Date: "2026-08-15"}) {
		t.Fatalf("provider request = %+v", provider.lastReq)
	}

	var resp flight.SearchResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Meta.ProvidersCalled != 1 || resp.Meta.ProvidersSucceeded != 1 {
		t.Fatalf("unexpected meta: %+v", resp.Meta)
	}
	if len(resp.Results) != 1 || resp.Results[0].Provider != "api-test-provider" {
		t.Fatalf("unexpected results: %+v", resp.Results)
	}
}

func TestSearchFlightsHandlerValidatesRequiredQueryParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := newRouter(flight.NewSearchService(nil, 100*time.Millisecond), hotel.NewSearchService([]domain.HotelProvider{}, 100*time.Millisecond))

	tests := []string{
		"/search/flights?to=HAN&date=2026-08-15",
		"/search/flights?from=SGN&date=2026-08-15",
		"/search/flights?from=SGN&to=HAN",
	}
	for _, target := range tests {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, target, nil)
		router.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("%s status = %d, want 400; body=%s", target, recorder.Code, recorder.Body.String())
		}
	}
}

func TestSearchHotelsHandlerReturnsResults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &apiTestHotelProvider{}
	router := newRouter(flight.NewSearchService(nil, 100*time.Millisecond), hotel.NewSearchService([]domain.HotelProvider{provider}, 100*time.Millisecond))

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/search/hotels?city=DAD&checkin=2026-08-15&checkout=2026-08-17", nil)
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", recorder.Code, recorder.Body.String())
	}

	var resp hotel.SearchResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Results == nil {
		t.Fatal("results should be an empty JSON array, not null")
	}
	if resp.Meta.ProvidersCalled != 1 || resp.Meta.ProvidersSucceeded != 1 {
		t.Fatalf("unexpected meta: %+v", resp.Meta)
	}
}

func TestSearchHotelsHandlerValidatesRequiredQueryParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := newRouter(flight.NewSearchService(nil, 100*time.Millisecond), hotel.NewSearchService([]domain.HotelProvider{}, 100*time.Millisecond))

	tests := []string{
		"/search/hotels?checkin=2026-08-15&checkout=2026-08-17",
		"/search/hotels?city=DAD&checkout=2026-08-17",
		"/search/hotels?city=DAD&checkin=2026-08-15",
	}
	for _, target := range tests {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, target, nil)
		router.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("%s status = %d, want 400; body=%s", target, recorder.Code, recorder.Body.String())
		}
	}
}