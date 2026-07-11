package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/goaggregator/internal/breaker"
	"github.com/yourusername/goaggregator/internal/cache"
	"github.com/yourusername/goaggregator/internal/config"
	"github.com/yourusername/goaggregator/internal/domain"
	"github.com/yourusername/goaggregator/internal/flight"
	"github.com/yourusername/goaggregator/internal/hotel"
	"github.com/yourusername/goaggregator/internal/limiter"
	"github.com/yourusername/goaggregator/internal/provider"
)

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.SlogLevel()}))
	slog.SetDefault(logger)

	flightProviders := provider.NewHTTPProviders(cfg.ProviderHosts, time.Duration(cfg.ProviderTimeoutMs)*time.Millisecond)
	flightProviders = breaker.NewFlightProvidersWithSettings(flightProviders, breaker.Settings{
		OpenCooldown:     time.Duration(cfg.BreakerCooldownMs) * time.Millisecond,
		HalfOpenRequests: uint32(cfg.BreakerHalfOpenCalls),
	})
	flightProviders = limiter.NewFlightProvidersWithSettings(flightProviders, limiter.Settings{
		RequestsPerSecond: float64(cfg.ProviderRateLimitRPS),
		Burst:             cfg.ProviderRateLimitBurst,
	})

	hotelProviders := provider.NewHTTPHotelProviders(cfg.ProviderHosts, time.Duration(cfg.ProviderTimeoutMs)*time.Millisecond)
	hotelProviders = breaker.NewHotelProvidersWithSettings(hotelProviders, breaker.Settings{
		OpenCooldown:     time.Duration(cfg.BreakerCooldownMs) * time.Millisecond,
		HalfOpenRequests: uint32(cfg.BreakerHalfOpenCalls),
	})
	hotelProviders = limiter.NewHotelProvidersWithSettings(hotelProviders, limiter.Settings{
		RequestsPerSecond: float64(cfg.ProviderRateLimitRPS),
		Burst:             cfg.ProviderRateLimitBurst,
	})

	flightCache := cache.NewRedisCache(cfg.RedisAddr, cfg.RedisDB)
	defer flightCache.Close()
	flightSearch := flight.NewSearchService(flightProviders, time.Duration(cfg.ProviderTimeoutMs)*time.Millisecond).
		WithCache(flightCache, time.Duration(cfg.CacheTTLSeconds)*time.Second)

	hotelCache := cache.NewRedisCache(cfg.RedisAddr, cfg.RedisDB)
	defer hotelCache.Close()
	hotelSearch := hotel.NewSearchService(hotelProviders, time.Duration(cfg.ProviderTimeoutMs)*time.Millisecond).
		WithCache(hotelCache, time.Duration(cfg.CacheTTLSeconds)*time.Second)

	r := newRouter(flightSearch, hotelSearch, flightProviders, hotelProviders)

	slog.Info("starting server",
		"port", cfg.Port,
		"log_level", cfg.LogLevel,
		"provider_timeout_ms", cfg.ProviderTimeoutMs,
		"provider_rate_limit_rps", cfg.ProviderRateLimitRPS,
		"provider_rate_limit_burst", cfg.ProviderRateLimitBurst,
		"breaker_cooldown_ms", cfg.BreakerCooldownMs,
		"breaker_half_open_calls", cfg.BreakerHalfOpenCalls,
		"redis_addr", cfg.RedisAddr,
		"cache_ttl_seconds", cfg.CacheTTLSeconds,
		"providers", len(flightProviders),
	)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("forced shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}

func newRouter(flightSearch *flight.SearchService, hotelSearch *hotel.SearchService, flightProviders []domain.Provider, hotelProviders []domain.HotelProvider) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "goaggregator-api",
		})
	})

	r.GET("/search/flights", searchFlightsHandler(flightSearch))
	r.GET("/search/hotels", searchHotelsHandler(hotelSearch))

	r.GET("/providers/status", providersStatusHandler(flightProviders, hotelProviders))
	return r
}

func providersStatusHandler(flightProviders []domain.Provider, hotelProviders []domain.HotelProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		type ProviderStatus struct {
			Name  string `json:"name"`
			Type  string `json:"type"` // "flight" or "hotel"
			State string `json:"state"`
		}

		var statuses []ProviderStatus

		for _, p := range flightProviders {
			if bp, ok := p.(*breaker.FlightProvider); ok {
				statuses = append(statuses, ProviderStatus{
					Name:  bp.Name(),
					Type:  "flight",
					State: bp.State().String(),
				})
			}
		}

		for _, p := range hotelProviders {
			if bp, ok := p.(*breaker.HotelProvider); ok {
				statuses = append(statuses, ProviderStatus{
					Name:  bp.Name(),
					Type:  "hotel",
					State: bp.State().String(),
				})
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"providers": statuses,
		})
	}
}

func searchFlightsHandler(flightSearch *flight.SearchService) gin.HandlerFunc {
	return func(c *gin.Context) {
		searchReq := domain.SearchRequest{
			From: c.Query("from"),
			To:   c.Query("to"),
			Date: c.Query("date"),
		}
		if searchReq.From == "" || searchReq.To == "" || searchReq.Date == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "from, to, and date query params are required"})
			return
		}

		response := flightSearch.Search(c.Request.Context(), searchReq)
		c.JSON(http.StatusOK, response)
	}
}

func searchHotelsHandler(hotelSearch *hotel.SearchService) gin.HandlerFunc {
	return func(c *gin.Context) {
		searchReq := domain.HotelSearchRequest{
			City:     c.Query("city"),
			CheckIn:  c.Query("checkin"),
			CheckOut: c.Query("checkout"),
		}
		if searchReq.City == "" || searchReq.CheckIn == "" || searchReq.CheckOut == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "city, checkin, and checkout query params are required"})
			return
		}

		response := hotelSearch.Search(c.Request.Context(), searchReq)
		c.JSON(http.StatusOK, response)
	}
}