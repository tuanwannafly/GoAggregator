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
	"github.com/yourusername/goaggregator/internal/config"
	"github.com/yourusername/goaggregator/internal/domain"
	"github.com/yourusername/goaggregator/internal/flight"
	"github.com/yourusername/goaggregator/internal/limiter"
	"github.com/yourusername/goaggregator/internal/provider"
)

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.SlogLevel()}))
	slog.SetDefault(logger)
	providers := provider.NewHTTPProviders(cfg.ProviderHosts, time.Duration(cfg.ProviderTimeoutMs)*time.Millisecond)
	providers = breaker.NewProvidersWithSettings(providers, breaker.Settings{
		OpenCooldown:     time.Duration(cfg.BreakerCooldownMs) * time.Millisecond,
		HalfOpenRequests: uint32(cfg.BreakerHalfOpenCalls),
	})
	providers = limiter.NewProvidersWithSettings(providers, limiter.Settings{
		RequestsPerSecond: float64(cfg.ProviderRateLimitRPS),
		Burst:             cfg.ProviderRateLimitBurst,
	})
	flightSearch := flight.NewSearchService(providers, time.Duration(cfg.ProviderTimeoutMs)*time.Millisecond)

	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "goaggregator-api",
		})
	})

	r.GET("/search/flights", func(c *gin.Context) {
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
	})

	slog.Info("starting server",
		"port", cfg.Port,
		"log_level", cfg.LogLevel,
		"provider_timeout_ms", cfg.ProviderTimeoutMs,
		"provider_rate_limit_rps", cfg.ProviderRateLimitRPS,
		"provider_rate_limit_burst", cfg.ProviderRateLimitBurst,
		"breaker_cooldown_ms", cfg.BreakerCooldownMs,
		"breaker_half_open_calls", cfg.BreakerHalfOpenCalls,
		"providers", len(providers),
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
