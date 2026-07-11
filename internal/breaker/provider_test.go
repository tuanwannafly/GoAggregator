package breaker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sony/gobreaker"
	"github.com/yourusername/goaggregator/internal/domain"
)

type scriptedProvider struct {
	name   string
	errors []bool
	called int
	block  func()
}

func (p *scriptedProvider) Name() string {
	return p.name
}

func (p *scriptedProvider) Search(ctx context.Context, req domain.SearchRequest) (*domain.ProviderSearchResponse, error) {
	p.called++
	if p.block != nil {
		p.block()
	}
	if p.called <= len(p.errors) && p.errors[p.called-1] {
		return nil, errors.New("provider failed")
	}
	return &domain.ProviderSearchResponse{Provider: p.name}, nil
}

func TestFlightProviderCircuitBreakerTripsAtSixtyPercentFailuresOverFiveRequests(t *testing.T) {
	base := &scriptedProvider{
		name:   "provider-flaky",
		errors: []bool{true, true, false, false, true, false},
	}
	protected := NewFlightProvider(base)

	for i := 0; i < 5; i++ {
		_, _ = protected.Search(context.Background(), domain.SearchRequest{})
	}

	if protected.State() != gobreaker.StateOpen {
		t.Fatalf("breaker state = %s, want open", protected.State())
	}

	_, err := protected.Search(context.Background(), domain.SearchRequest{})
	if err == nil {
		t.Fatal("expected open circuit error")
	}
	if base.called != 5 {
		t.Fatalf("provider called = %d, want 5; open breaker should skip the 6th call", base.called)
	}
}

func TestFlightProviderCircuitBreakerStaysClosedBelowSixtyPercentFailures(t *testing.T) {
	base := &scriptedProvider{
		name:   "provider-mostly-ok",
		errors: []bool{true, true, false, false, false, false},
	}
	protected := NewFlightProvider(base)

	for i := 0; i < 5; i++ {
		_, _ = protected.Search(context.Background(), domain.SearchRequest{})
	}

	if protected.State() != gobreaker.StateClosed {
		t.Fatalf("breaker state = %s, want closed", protected.State())
	}

	_, err := protected.Search(context.Background(), domain.SearchRequest{})
	if err != nil {
		t.Fatalf("unexpected error while breaker should be closed: %v", err)
	}
	if base.called != 6 {
		t.Fatalf("provider called = %d, want 6", base.called)
	}
}

func TestFlightProviderCircuitBreakerRetriesAfterCooldownInHalfOpen(t *testing.T) {
	base := &scriptedProvider{
		name:   "provider-recovering",
		errors: []bool{true, true, false, false, true, false, false},
	}
	protected := NewFlightProviderWithSettings(base, Settings{
		OpenCooldown:     20 * time.Millisecond,
		HalfOpenRequests: 1,
	})

	for i := 0; i < 5; i++ {
		_, _ = protected.Search(context.Background(), domain.SearchRequest{})
	}
	if protected.State() != gobreaker.StateOpen {
		t.Fatalf("breaker state = %s, want open", protected.State())
	}

	time.Sleep(25 * time.Millisecond)
	_, err := protected.Search(context.Background(), domain.SearchRequest{})
	if err != nil {
		t.Fatalf("half-open retry should call recovered provider: %v", err)
	}
	if protected.State() != gobreaker.StateClosed {
		t.Fatalf("breaker state = %s, want closed after successful half-open retry", protected.State())
	}
	if base.called != 6 {
		t.Fatalf("provider called = %d, want 6", base.called)
	}
}

func TestFlightProviderCircuitBreakerLimitsHalfOpenRequests(t *testing.T) {
	base := &scriptedProvider{
		name:   "provider-limited-half-open",
		errors: []bool{true, true, false, false, true, false},
	}
	protected := NewFlightProviderWithSettings(base, Settings{
		OpenCooldown:     20 * time.Millisecond,
		HalfOpenRequests: 1,
	})

	for i := 0; i < 5; i++ {
		_, _ = protected.Search(context.Background(), domain.SearchRequest{})
	}
	time.Sleep(25 * time.Millisecond)

	enteredHalfOpen := make(chan struct{})
	releaseHalfOpen := make(chan struct{})
	base.block = func() {
		close(enteredHalfOpen)
		<-releaseHalfOpen
	}

	firstDone := make(chan error, 1)
	go func() {
		_, err := protected.Search(context.Background(), domain.SearchRequest{})
		firstDone <- err
	}()
	<-enteredHalfOpen

	_, err := protected.Search(context.Background(), domain.SearchRequest{})
	if err == nil {
		t.Fatal("expected second half-open request to be rejected")
	}
	if base.called != 6 {
		t.Fatalf("provider called = %d, want 6; second half-open request should not call provider", base.called)
	}

	close(releaseHalfOpen)
	if err := <-firstDone; err != nil {
		t.Fatalf("first half-open request should succeed: %v", err)
	}
}