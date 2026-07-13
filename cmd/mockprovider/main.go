package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type FlightResult struct {
	ID       string `json:"id"`
	From     string `json:"from"`
	To       string `json:"to"`
	Date     string `json:"date"`
	Price    int    `json:"price"`
	Currency string `json:"currency"`
	Airline  string `json:"airline"`
}

type SearchResponse struct {
	Provider string         `json:"provider"`
	Results  []FlightResult `json:"results"`
}

type HotelResult struct {
	ID            string `json:"id"`
	City          string `json:"city"`
	CheckIn       string `json:"checkin"`
	CheckOut      string `json:"checkout"`
	Name          string `json:"name"`
	PricePerNight int    `json:"price_per_night"`
	Currency      string `json:"currency"`
}

type HotelSearchResponse struct {
	Provider string        `json:"provider"`
	Results  []HotelResult `json:"results"`
}

type ControlResponse struct {
	Status    string  `json:"status"`
	LatencyMs int     `json:"latency_ms"`
	ErrorRate float64 `json:"error_rate"`
}

type ControlRequest struct {
	LatencyMs int     `json:"latency_ms"`
	ErrorRate float64 `json:"error_rate"`
}

type ControlState struct {
	mu        sync.RWMutex
	latencyMs int
	errorRate float64
}

func (c *ControlState) Apply(req ControlRequest) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.latencyMs = req.LatencyMs
	c.errorRate = req.ErrorRate
}

func (c *ControlState) Snapshot() ControlRequest {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return ControlRequest{
		LatencyMs: c.latencyMs,
		ErrorRate: c.errorRate,
	}
}

func (c *ControlState) ShouldFail() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.errorRate <= 0 {
		return false
	}
	return rand.Float64() < c.errorRate
}

func (c *ControlState) LatencyMs() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.latencyMs
}

func sleepWithContext(ctx context.Context, latencyMs int) bool {
	if latencyMs <= 0 {
		return true
	}
	timer := time.NewTimer(time.Duration(latencyMs) * time.Millisecond)
	defer timer.Stop()
	select {
	case <-timer.C:
		return true
	case <-ctx.Done():
		return false
	}
}

func parseEnvInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		slog.Warn("invalid integer env, using fallback", "key", key, "value", raw, "fallback", fallback)
		return fallback
	}
	return value
}

func parseEnvFloat(key string, fallback float64) float64 {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		slog.Warn("invalid float env, using fallback", "key", key, "value", raw, "fallback", fallback)
		return fallback
	}
	return value
}

var cities = []struct{ from, to string }{
	{"SGN", "HAN"},
	{"SGN", "DAD"},
	{"HAN", "SGN"},
	{"DAD", "SGN"},
	{"SGN", "BKK"},
	{"HAN", "DAD"},
}

var airlines = []string{
	"Mock Airlines",
	"SkyHigh Express",
	"Pacific Wings",
	"Golden Dragon Air",
}

var defaultDates = []string{
	"2026-08-15",
	"2026-08-16",
	"2026-08-17",
	"2026-08-18",
	"2026-09-01",
}

var hotelCities = []string{"SGN", "HAN", "DAD", "BKK", "PQC"}

var hotelNames = []string{
	"Grand Plaza Hotel",
	"Seaside Resort",
	"City Center Inn",
	"Boutique Garden Hotel",
	"Business Tower Suites",
	"Riverside Lodge",
	"Heritage Boutique",
}

func generateResults(basePrice int) []FlightResult {
	count := rand.Intn(6) + 3 // 3-8 results
	results := make([]FlightResult, 0, count)
	for i := 0; i < count; i++ {
		route := cities[rand.Intn(len(cities))]
		priceVariation := int(float64(basePrice) * (0.8 + rand.Float64()*0.4)) // ±20%
		results = append(results, FlightResult{
			ID:       fmt.Sprintf("FL-%03d", rand.Intn(999)+1),
			From:     route.from,
			To:       route.to,
			Date:     defaultDates[rand.Intn(len(defaultDates))],
			Price:    priceVariation,
			Currency: "VND",
			Airline:  airlines[rand.Intn(len(airlines))],
		})
	}
	return results
}

func generateHotelResults(basePrice int, city, checkIn, checkOut string) []HotelResult {
	count := rand.Intn(5) + 3 // 3-7 hotels
	results := make([]HotelResult, 0, count)
	requestedCity := city
	if requestedCity == "" {
		requestedCity = hotelCities[rand.Intn(len(hotelCities))]
	}
	for i := 0; i < count; i++ {
		priceVariation := int(float64(basePrice) * (0.8 + rand.Float64()*0.4)) // ±20%
		results = append(results, HotelResult{
			ID:            fmt.Sprintf("HT-%03d", rand.Intn(999)+1),
			City:          requestedCity,
			CheckIn:       checkIn,
			CheckOut:      checkOut,
			Name:          hotelNames[rand.Intn(len(hotelNames))],
			PricePerNight: priceVariation,
			Currency:      "VND",
		})
	}
	return results
}

func main() {
	rand.Seed(time.Now().UnixNano())

	providerName := os.Getenv("PROVIDER_NAME")
	if providerName == "" {
		providerName = "mock-provider"
	}

	basePriceStr := os.Getenv("BASE_PRICE_VND")
	basePrice := 1_500_000
	if basePriceStr != "" {
		if v, err := strconv.Atoi(basePriceStr); err == nil {
			basePrice = v
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	initialLatencyMs := parseEnvInt("LATENCY_MS", 0)
	initialErrorRate := parseEnvFloat("ERROR_RATE", 0)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	state := &ControlState{}
	if initialLatencyMs >= 0 && initialErrorRate >= 0 && initialErrorRate <= 1 {
		state.Apply(ControlRequest{LatencyMs: initialLatencyMs, ErrorRate: initialErrorRate})
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "provider": providerName})
	})

	respondWithChaos := func(c *gin.Context, endpoint string, build func() any) {
		if state.ShouldFail() {
			slog.Warn("simulated failure", "provider", providerName, "endpoint", endpoint)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "simulated failure"})
			return
		}

		latency := state.LatencyMs()
		if latency > 0 {
			slog.Info("applying latency", "provider", providerName, "latency_ms", latency, "endpoint", endpoint)
			if ok := sleepWithContext(c.Request.Context(), latency); !ok {
				slog.Warn("request cancelled during latency", "provider", providerName)
				return
			}
		}

		c.JSON(http.StatusOK, build())
	}

	r.GET("/search", func(c *gin.Context) {
		respondWithChaos(c, "flights", func() any {
			return SearchResponse{
				Provider: providerName,
				Results:  generateResults(basePrice),
			}
		})
	})

	r.GET("/search/hotels", func(c *gin.Context) {
		respondWithChaos(c, "hotels", func() any {
			return HotelSearchResponse{
				Provider: providerName,
				Results: generateHotelResults(
					basePrice,
					c.Query("city"),
					c.Query("checkin"),
					c.Query("checkout"),
				),
			}
		})
	})

	r.POST("/control", func(c *gin.Context) {
		var req ControlRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		if req.ErrorRate < 0 || req.ErrorRate > 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "error_rate must be between 0 and 1"})
			return
		}
		if req.LatencyMs < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "latency_ms must be >= 0"})
			return
		}
		state.Apply(req)
		slog.Info("control updated", "provider", providerName, "latency_ms", req.LatencyMs, "error_rate", req.ErrorRate)
		c.JSON(http.StatusOK, ControlResponse{Status: "ok", LatencyMs: req.LatencyMs, ErrorRate: req.ErrorRate})
	})

	r.GET("/control", func(c *gin.Context) {
		snapshot := state.Snapshot()
		c.JSON(http.StatusOK, ControlResponse{Status: "ok", LatencyMs: snapshot.LatencyMs, ErrorRate: snapshot.ErrorRate})
	})

	slog.Info("starting mock provider", "provider", providerName, "port", port, "base_price_vnd", basePrice, "latency_ms", state.LatencyMs(), "error_rate", state.Snapshot().ErrorRate)
	if err := r.Run(":" + port); err != nil {
		slog.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}
