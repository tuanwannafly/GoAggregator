# GoAggregator

**Flight & Hotel Search & Price Comparison Aggregator**

A production-grade demonstration of distributed system resilience patterns including fan-out/fan-in, circuit breaker, rate limiter, and cache-aside strategies. Features mock providers controllable at runtime via `/control` endpoint for comprehensive chaos engineering.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Features](#features)
- [API Reference](#api-reference)
- [Configuration](#configuration)
- [Testing](#testing)
- [Demo Scripts](#demo-scripts)
- [Project Structure](#project-structure)
- [Resilience Patterns](#resilience-patterns)
- [License](#license)

## Overview

GoAggregator is a high-performance flight and hotel price aggregation service built with Go. It demonstrates enterprise-grade resilience patterns needed for real-world distributed systems, including graceful degradation, fault isolation, and chaos engineering capabilities.

### Key Capabilities

- **Multi-Provider Aggregation**: Simultaneously query multiple flight and hotel providers
- **Resilient Design**: Built-in fault tolerance with circuit breakers and rate limiters
- **Chaos Engineering**: Runtime-controllable mock providers for testing failure scenarios
- **Caching**: Redis-based cache-aside pattern with configurable TTL
- **Transparency**: Full observability with detailed response metadata

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Client Request                                  │
│                    GET /search/flights?from=SGN&to=HAN                      │
└─────────────────────────────────┬───────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           GoAggregator API (Port 8081)                       │
│  ┌─────────────────────────────────────────────────────────────────────────┐ │
│  │                        Cache-Aside Check (Redis)                        │ │
│  │                    TTL: 30-60s (short due to price volatility)          │ │
│  └─────────────────────────────────────────────────────────────────────────┘ │
│                                  │                                           │
│                    [Cache Miss] ─┴──────────────────────────────────────────│
│                                  │                                           │
│                                  ▼                                           │
│  ┌─────────────────────────────────────────────────────────────────────────┐ │
│  │                      Fan-Out (Parallel Goroutines)                      │ │
│  │                                                                               │ │
│  │    ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │ │
│  │    │ Provider A  │  │ Provider B  │  │ Provider C  │  │ Provider D  │    │ │
│  │    │ (Fast/50ms) │  │ (Slow/1.2s) │  │ (Flaky/35%) │  │ (Timeout/5s)│    │ │
│  │    └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘    │ │
│  │                                                                               │ │
│  │    Per-Provider: Context Timeout | Circuit Breaker | Rate Limiter          │ │
│  └─────────────────────────────────────────────────────────────────────────┘ │
│                                  │                                           │
│                                  ▼                                           │
│  ┌─────────────────────────────────────────────────────────────────────────┐ │
│  │                       Fan-In (Result Aggregation)                        │ │
│  │                    Sort by Price → Cache → Return                         │ │
│  └─────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Data Flow

1. **Request Received**: Client sends flight/hotel search request
2. **Cache Check**: Redis cache-aside pattern checks for cached results
3. **Fan-Out**: Parallel queries to all configured providers using goroutines
4. **Fault Isolation**: Each provider wrapped with timeout, circuit breaker, rate limiter
5. **Fan-In**: Results merged, sorted by price
6. **Response**: Cached + returned with full metadata

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)
- curl or any HTTP client

### 1. Start All Services

```bash
# Clone the repository
git clone https://github.com/tuanwannafly/GoAggregator.git
cd GoAggregator

# Start all services (API + 5 mock providers + Redis)
docker compose up --build
```

### 2. Verify Health

```bash
# Check API health
curl http://localhost:8081/healthz
# Response: {"status":"ok","service":"goaggregator-api"}

# Check provider health
curl http://localhost:18080/healthz
# Response: {"status":"ok","service":"provider-fast"}
```

### 3. Search Flights

```bash
curl "http://localhost:8081/search/flights?from=SGN&to=HAN&date=2026-08-15"
```

### 4. Search Hotels

```bash
curl "http://localhost:8081/search/hotels?city=HAN&checkin=2026-08-15&checkout=2026-08-17"
```

### 5. Check Provider Status

```bash
# View circuit breaker states for all providers
curl http://localhost:8081/providers/status
```

## Features

### Mock Providers

| Provider | Port | Base Price | Latency | Error Rate | Behavior |
|----------|------|------------|---------|------------|----------|
| provider-fast | 18080 | 1,200,000 VND | 50ms | 0% | Fast, reliable |
| provider-slow | 18081 | 1,350,000 VND | 1200ms | 0% | Simulates slow API |
| provider-flaky | 18082 | 1,100,000 VND | 200ms | 35% | Random failures |
| provider-timeout | 18083 | 1,500,000 VND | 5000ms | 0% | Always times out |
| provider-down | 18084 | 1,700,000 VND | 100ms | 100% | Always fails |

### Direct Provider Access

```bash
# Search directly on a provider
curl "http://localhost:18080/search?from=SGN&to=HAN&date=2026-08-15"

# Control provider behavior at runtime (chaos engineering)
curl -X POST http://localhost:18080/control \
  -H "Content-Type: application/json" \
  -d '{"latency_ms": 2000, "error_rate": 0.5}'

# View current control settings
curl http://localhost:18080/control
```

## API Reference

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/healthz` | Health check endpoint |
| GET | `/search/flights?from=&to=&date=` | Aggregate flight search |
| GET | `/search/hotels?city=&checkin=&checkout=` | Aggregate hotel search |
| GET | `/providers/status` | Circuit breaker state per provider |
| POST | `/control` | Set `latency_ms`, `error_rate` (mock) |
| GET | `/control` | View current control settings (mock) |

### Search Response Format

```json
{
  "results": [
    {
      "provider": "provider-fast",
      "id": "FL-123",
      "from": "SGN",
      "to": "HAN",
      "date": "2026-08-15",
      "price": 1250000,
      "currency": "VND",
      "airline": "Mock Airlines"
    }
  ],
  "meta": {
    "providers_called": 5,
    "providers_succeeded": 3,
    "providers_failed": [
      {
        "provider": "provider-flaky",
        "error": "provider \"provider-flaky\" returned status 500"
      },
      {
        "provider": "provider-down",
        "error": "circuit breaker: open"
      }
    ],
    "cache_hit": false,
    "duration_ms": 312
  }
}
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8081` | API server port |
| `LOG_LEVEL` | `info` | Logging level (debug, info, warn, error) |
| `PROVIDER_TIMEOUT_MS` | `3000` | Per-provider HTTP timeout |
| `PROVIDER_HOSTS` | (comma-separated) | Mock provider URLs |
| `REDIS_ADDR` | `redis:6379` | Redis address |
| `CACHE_TTL_SECONDS` | `45` | Cache TTL in seconds |
| `PROVIDER_RATE_LIMIT_RPS` | `10` | Requests per second per provider |
| `PROVIDER_RATE_LIMIT_BURST` | `20` | Token bucket burst size |
| `BREAKER_COOLDOWN_MS` | `30000` | Circuit breaker cooldown |
| `BREAKER_HALF_OPEN_CALLS` | `1` | Max requests in half-open state |

## Testing

### Run All Tests

```bash
# Unit + integration tests with race detector
go test -race ./...

# All tests without race detector
go test ./...

# Specific integration test
go test -v ./test/integration/... -run TestIntegrationChaosTwoProvidersFail
```

### Demo Scripts

#### 1. Circuit Breaker Demo

Demonstrates circuit breaker opening after repeated failures, then recovering.

```bash
#!/bin/bash
# circuit-breaker-demo.sh

API="http://localhost:8081"
PROVIDER_DOWN="http://localhost:18084"

echo "=== Circuit Breaker Demo ==="
echo "Provider 'provider-down' has 100% error rate"

echo "--- Step 1: Check initial circuit state ---"
curl -s "$API/providers/status" | jq '.providers[] | select(.name=="provider-down")'

echo "--- Step 2: Trigger 6 searches ---"
for i in {1..6}; do
  echo "Search #$i:"
  curl -s "$API/search/flights?from=SGN&to=HAN&date=2026-08-15" | jq '.meta | {providers_called, providers_succeeded, providers_failed: .providers_failed[]?.provider}'
  sleep 0.5
done

echo "--- Step 3: Verify circuit is OPEN ---"
curl -s "$API/providers/status" | jq '.providers[] | select(.name=="provider-down")'

echo "--- Step 4: Wait for cooldown (30s) ---"
sleep 35

echo "--- Step 5: Next search probes half-open ---"
curl -s "$API/search/flights?from=SGN&to=HAN&date=2026-08-15" | jq '.meta | {providers_called, providers_succeeded, providers_failed: .providers_failed[]?.provider}'
curl -s "$API/providers/status" | jq '.providers[] | select(.name=="provider-down")'

echo "--- Step 6: Fix provider ---"
curl -s -X POST "$PROVIDER_DOWN/control" -H "Content-Type: application/json" -d '{"latency_ms": 100, "error_rate": 0}'
curl -s "$API/search/flights?from=SGN&to=HAN&date=2026-08-15" | jq '.meta | {providers_called, providers_succeeded, providers_failed: .providers_failed[]?.provider}'
curl -s "$API/providers/status" | jq '.providers[] | select(.name=="provider-down")'
```

#### 2. Chaos Engineering Demo

Simulates 2/5 providers failing, verifying aggregator returns partial results.

```bash
#!/bin/bash
# chaos-demo.sh

API="http://localhost:8081"
FLAKY="http://localhost:18082"
DOWN="http://localhost:18084"

echo "=== Chaos Engineering Demo: 2/5 Providers Fail ==="

echo "--- Initial state: all providers healthy ---"
curl -s "$API/search/flights?from=SGN&to=HAN&date=2026-08-15" | jq '.meta | {providers_called, providers_succeeded, providers_failed: .providers_failed[]?.provider}'

echo "--- Enable 100% error on provider-flaky and provider-down ---"
curl -s -X POST "$FLAKY/control" -H "Content-Type: application/json" -d '{"latency_ms": 200, "error_rate": 1.0}'
curl -s -X POST "$DOWN/control" -H "Content-Type: application/json" -d '{"latency_ms": 100, "error_rate": 1.0}'

echo "--- Search with 2 providers failing ---"
curl -s "$API/search/flights?from=SGN&to=HAN&date=2026-08-15" | jq '.meta | {providers_called, providers_succeeded, providers_failed: .providers_failed[]?.provider}'

echo "--- Results still returned from 3 healthy providers ---"
curl -s "$API/search/flights?from=SGN&to=HAN&date=2026-08-15" | jq '.results[] | {provider, price, airline}'

echo "--- Restore providers ---"
curl -s -X POST "$FLAKY/control" -H "Content-Type: application/json" -d '{"latency_ms": 200, "error_rate": 0.35}'
curl -s -X POST "$DOWN/control" -H "Content-Type: application/json" -d '{"latency_ms": 100, "error_rate": 1.0}'
```

#### 3. Timeout Isolation Demo

Proves slow provider doesn't block fast ones.

```bash
#!/bin/bash
# timeout-demo.sh

API="http://localhost:8081"
TIMEOUT_PROV="http://localhost:18083"

echo "=== Timeout Isolation Demo ==="

echo "--- Set provider-timeout to 5s latency (aggregator timeout is 3s) ---"
curl -s -X POST "$TIMEOUT_PROV/control" -H "Content-Type: application/json" -d '{"latency_ms": 5000, "error_rate": 0}'

echo "--- Search (should complete in ~3s, not 5s) ---"
time curl -s "$API/search/flights?from=SGN&to=HAN&date=2026-08-15" | jq '.meta | {duration_ms, providers_called, providers_succeeded, providers_failed: .providers_failed[]?.provider}'

echo "--- Results from fast providers only ---"
curl -s "$API/search/flights?from=SGN&to=HAN&date=2026-08-15" | jq '.results[] | {provider, price}'
```

#### 4. Rate Limiter Demo

```bash
#!/bin/bash
# rate-limiter-demo.sh

API="http://localhost:8081"

echo "=== Rate Limiter Demo ==="
echo "Default: 10 req/s per provider, burst 20"

echo "--- Burst 25 rapid searches (last 5 should be rate-limited) ---"
for i in {1..25}; do
  curl -s "$API/search/flights?from=SGN&to=HAN&date=2026-08-15" | jq -r '.meta.providers_failed[]?.error // "ok"' | head -1
done
```

## Project Structure

```
goaggregator/
├── cmd/
│   ├── api/main.go              # API server entry point
│   └── mockprovider/main.go     # Mock provider server
├── internal/
│   ├── config/                  # Environment configuration
│   ├── domain/                  # FlightResult, HotelResult, Provider interfaces
│   ├── flight/                  # Flight search service
│   ├── hotel/                   # Hotel search service
│   ├── aggregator/              # Fan-out/fan-in core logic
│   ├── resilience/
│   │   ├── breaker/             # Circuit breaker wrapper (sony/gobreaker)
│   │   └── limiter/             # Rate limiter wrapper (token bucket)
│   ├── cache/                   # Redis + Noop + Memory cache implementations
│   ├── provider/                # HTTP clients for mock providers
│   ├── http/                    # Gin handlers, middleware
│   └── telemetry/               # Tracing and logging
├── test/integration/            # Chaos integration tests
├── frontend/                    # Next.js frontend application
├── scripts/                     # Demo and utility scripts
├── docker-compose.yml           # Docker orchestration
├── Dockerfile                   # Container definition
└── README.md                    # This file
```

## Resilience Patterns

### 1. Fan-Out/Fan-In

Each provider is called in a goroutine with `context.WithTimeout`. This ensures:
- Slow providers only lose their own results
- Fast providers return quickly regardless of slow ones
- Total latency ≈ max(timeout, slowest healthy provider)

### 2. Circuit Breaker (Per-Provider)

Implemented using `sony/gobreaker`:
- **Opens** after 5 requests with ≥60% failure rate
- **Cooldown** of 30 seconds before attempting recovery
- **Half-open** state allows 1 probe request
- **Closes** on successful probe

### 3. Rate Limiter (Per-Provider)

Token bucket implementation using `golang.org/x/time/rate`:
- Default: 10 requests/second per provider
- Burst size: 20 requests
- Excess requests dropped locally (not sent)
- Does not count as failures or trigger circuit breaker

### 4. Cache-Aside (Redis)

- TTL: 30-60 seconds (short due to price volatility)
- Cache key includes route + date
- Reduces load on providers
- Improves response time for repeated queries

### 5. Transparent Response Metadata

Every response includes:
- `providers_called`: Total providers queried
- `providers_succeeded`: Successful responses
- `providers_failed[]`: Failed providers with error details
- `cache_hit`: Whether result came from cache
- `duration_ms`: Total processing time

This makes graceful degradation visible to callers, not hidden.

## License

MIT License

Copyright (c) 2026 GoAggregator Contributors

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
