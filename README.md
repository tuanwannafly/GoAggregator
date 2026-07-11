# GoAggregator

Flight/Hotel Search & Price Comparison Aggregator — demonstrates resilience patterns (fan-out/fan-in, circuit breaker, rate limiter, cache-aside) with mock providers controllable at runtime via `/control` endpoint.

## Architecture

```
Client → GoAggregator API → Redis (cache-aside, check first)
                          → [cache miss] → fan-out parallel to:
                                Provider A (fast, stable)
                                Provider B (slow ~1.5s)
                                Provider C (10% random errors)
                                Provider D (circuit open, skipped)
                                Provider E (random timeout)
                          → fan-in merge results → sort by price → cache → return
```

**Key Resilience Patterns:**
- **Fan-out/Fan-in**: Each provider called in goroutine with `context.WithTimeout` — slow provider only loses its own results, doesn't block others
- **Circuit Breaker** (per-provider): `sony/gobreaker` — opens after 5 requests with ≥60% failure rate, cooldown 30s, half-open allows 1 probe
- **Rate Limiter** (per-provider): Token bucket (`golang.org/x/time/rate`) — excess requests dropped locally, not sent
- **Cache-Aside** (Redis, TTL 30-60s): Short TTL because flight/hotel prices change frequently

## Quick Start

```bash
# 1. Start all services (API + 5 mock providers + Redis)
docker compose up --build

# 2. Verify health
curl http://localhost:8081/healthz
# {"status":"ok","service":"goaggregator-api"}

# 3. Search flights
curl "http://localhost:8081/search/flights?from=SGN&to=HAN&date=2026-08-15"

# 4. Search hotels
curl "http://localhost:8081/search/hotels?city=HAN&checkin=2026-08-15&checkout=2026-08-17"

# 5. Check provider circuit breaker status
curl http://localhost:8081/providers/status
```

## Mock Providers (Direct Access)

Each provider exposes `/search` and `/control` endpoints:

| Provider | Port | Base Price | Latency | Error Rate |
|----------|------|------------|---------|------------|
| provider-fast | 18080 | 1,200,000 | 50ms | 0% |
| provider-slow | 18081 | 1,350,000 | 1200ms | 0% |
| provider-flaky | 18082 | 1,100,000 | 200ms | 35% |
| provider-timeout | 18083 | 1,500,000 | 5000ms | 0% |
| provider-down | 18084 | 1,700,000 | 100ms | 100% |

```bash
# Check provider health
curl http://localhost:18080/healthz

# Search directly on a provider
curl "http://localhost:18080/search?from=SGN&to=HAN&date=2026-08-15"

# Control provider behavior at runtime (chaos engineering)
curl -X POST http://localhost:18080/control \
  -H "Content-Type: application/json" \
  -d '{"latency_ms": 2000, "error_rate": 0.5}'

# View current control settings
curl http://localhost:18080/control
```

## Demo Scripts

### 1. Circuit Breaker Trip Demo

This script demonstrates the circuit breaker opening after repeated failures, then recovering.

```bash
#!/bin/bash
# circuit-breaker-demo.sh

API="http://localhost:8081"
PROVIDER_DOWN="http://localhost:18084"  # provider-down (100% error rate)

echo "=== Circuit Breaker Demo ==="
echo "Provider 'provider-down' has 100% error rate"
echo ""

echo "--- Step 1: Check initial circuit state (should be Closed) ---"
curl -s "$API/providers/status" | jq '.providers[] | select(.name=="provider-down")'

echo ""
echo "--- Step 2: Trigger 6 searches (circuit opens after 5 failures @ 60% threshold) ---"
for i in {1..6}; do
  echo "Search #$i:"
  curl -s "$API/search/flights?from=SGN&to=HAN&date=2026-08-15" | jq '.meta | {providers_called, providers_succeeded, providers_failed: .providers_failed[]?.provider}'
  sleep 0.5
done

echo ""
echo "--- Step 3: Verify circuit is OPEN (provider skipped, no HTTP call made) ---"
curl -s "$API/providers/status" | jq '.providers[] | select(.name=="provider-down")'

echo ""
echo "--- Step 4: Wait for cooldown (30s default), then circuit goes HALF-OPEN ---"
echo "Waiting 35 seconds..."
sleep 35

echo ""
echo "--- Step 5: Next search probes half-open (1 request allowed) ---"
curl -s "$API/search/flights?from=SGN&to=HAN&date=2026-08-15" | jq '.meta | {providers_called, providers_succeeded, providers_failed: .providers_failed[]?.provider}'
curl -s "$API/providers/status" | jq '.providers[] | select(.name=="provider-down")'

echo ""
echo "--- Step 6: Fix provider (set error_rate=0), next search closes circuit ---"
curl -s -X POST "$PROVIDER_DOWN/control" -H "Content-Type: application/json" -d '{"latency_ms": 100, "error_rate": 0}'
curl -s "$API/search/flights?from=SGN&to=HAN&date=2026-08-15" | jq '.meta | {providers_called, providers_succeeded, providers_failed: .providers_failed[]?.provider}'
curl -s "$API/providers/status" | jq '.providers[] | select(.name=="provider-down")'

echo ""
echo "=== Demo Complete ==="
```

**Run it:**
```bash
chmod +x circuit-breaker-demo.sh
./circuit-breaker-demo.sh
```

**Expected Output Progression:**
```
Step 1: State = "closed"
Step 2: 5 searches → all fail, 6th search → circuit opens, provider skipped
Step 3: State = "open", providers_failed shows "circuit breaker: open"
Step 4: Wait 30s cooldown
Step 5: State = "half-open", 1 probe request allowed
Step 6: Provider fixed → probe succeeds → State = "closed"
```

---

### 2. Chaos Engineering Demo (Partial Failure)

Simulate 2/5 providers failing, verify aggregator returns 200 with partial results.

```bash
#!/bin/bash
# chaos-demo.sh

API="http://localhost:8081"
FLAKY="http://localhost:18082"
DOWN="http://localhost:18084"

echo "=== Chaos Engineering Demo: 2/5 Providers Fail ==="
echo ""

echo "--- Initial state: all providers healthy ---"
curl -s "$API/search/flights?from=SGN&to=HAN&date=2026-08-15" | jq '.meta | {providers_called, providers_succeeded, providers_failed: .providers_failed[]?.provider}'

echo ""
echo "--- Enable 100% error on provider-flaky and provider-down ---"
curl -s -X POST "$FLAKY/control" -H "Content-Type: application/json" -d '{"latency_ms": 200, "error_rate": 1.0}'
curl -s -X POST "$DOWN/control" -H "Content-Type: application/json" -d '{"latency_ms": 100, "error_rate": 1.0}'

echo ""
echo "--- Search with 2 providers failing ---"
curl -s "$API/search/flights?from=SGN&to=HAN&date=2026-08-15" | jq '.meta | {providers_called, providers_succeeded, providers_failed: .providers_failed[]?.provider}'

echo ""
echo "--- Results still returned from 3 healthy providers ---"
curl -s "$API/search/flights?from=SGN&to=HAN&date=2026-08-15" | jq '.results[] | {provider, price, airline}'

echo ""
echo "--- Restore providers ---"
curl -s -X POST "$FLAKY/control" -H "Content-Type: application/json" -d '{"latency_ms": 200, "error_rate": 0.35}'
curl -s -X POST "$DOWN/control" -H "Content-Type: application/json" -d '{"latency_ms": 100, "error_rate": 1.0}'

echo ""
echo "=== Demo Complete ==="
```

---

### 3. Timeout Isolation Demo

Prove slow provider doesn't block fast ones.

```bash
#!/bin/bash
# timeout-demo.sh

API="http://localhost:8081"
TIMEOUT_PROV="http://localhost:18083"

echo "=== Timeout Isolation Demo ==="
echo ""

echo "--- Set provider-timeout to 5s latency (aggregator timeout is 3s) ---"
curl -s -X POST "$TIMEOUT_PROV/control" -H "Content-Type: application/json" -d '{"latency_ms": 5000, "error_rate": 0}'

echo ""
echo "--- Search (should complete in ~3s, not 5s) ---"
time curl -s "$API/search/flights?from=SGN&to=HAN&date=2026-08-15" | jq '.meta | {duration_ms, providers_called, providers_succeeded, providers_failed: .providers_failed[]?.provider}'

echo ""
echo "--- Results from fast providers only ---"
curl -s "$API/search/flights?from=SGN&to=HAN&date=2026-08-15" | jq '.results[] | {provider, price}'

echo ""
echo "--- Restore ---"
curl -s -X POST "$TIMEOUT_PROV/control" -H "Content-Type: application/json" -d '{"latency_ms": 5000, "error_rate": 0}'
```

---

### 4. Rate Limiter Demo

```bash
#!/bin/bash
# rate-limiter-demo.sh

API="http://localhost:8081"

echo "=== Rate Limiter Demo ==="
echo "Default: 10 req/s per provider, burst 20"
echo ""

echo "--- Burst 25 rapid searches (last 5 should be rate-limited) ---"
for i in {1..25}; do
  curl -s "$API/search/flights?from=SGN&to=HAN&date=2026-08-15" | jq -r '.meta.providers_failed[]?.error // "ok"' | head -1
done

echo ""
echo "Check logs for 'rate limited' messages on provider-flaky (lowest burst tolerance)"
```

---

## API Reference

| Method | Path | Description |
|--------|------|-------------|
| GET | `/healthz` | Health check |
| GET | `/search/flights?from=&to=&date=` | Aggregate flight search |
| GET | `/search/hotels?city=&checkin=&checkout=` | Aggregate hotel search |
| GET | `/providers/status` | Circuit breaker state per provider |
| POST (mock) | `/control` | Set `latency_ms`, `error_rate` |
| GET (mock) | `/control` | View current control settings |

**Response Format (Search):**
```json
{
  "results": [
    { "provider": "provider-fast", "id": "FL-123", "from": "SGN", "to": "HAN", "date": "2026-08-15", "price": 1250000, "currency": "VND", "airline": "Mock Airlines" }
  ],
  "meta": {
    "providers_called": 5,
    "providers_succeeded": 3,
    "providers_failed": [
      { "provider": "provider-flaky", "error": "provider \"provider-flaky\" returned status 500" },
      { "provider": "provider-down", "error": "provider \"provider-down\" returned status 500" }
    ],
    "cache_hit": false,
    "duration_ms": 312
  }
}
```

## Configuration (Environment Variables)

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8081` | API server port |
| `LOG_LEVEL` | `info` | slog level (debug, info, warn, error) |
| `PROVIDER_TIMEOUT_MS` | `3000` | Per-provider HTTP timeout |
| `PROVIDER_HOSTS` | (comma-separated) | Mock provider URLs |
| `REDIS_ADDR` | `redis:6379` | Redis address |
| `CACHE_TTL_SECONDS` | `45` | Cache TTL |
| `PROVIDER_RATE_LIMIT_RPS` | `10` | Requests/sec per provider |
| `PROVIDER_RATE_LIMIT_BURST` | `20` | Token bucket burst |
| `BREAKER_COOLDOWN_MS` | `30000` | Circuit breaker open→half-open cooldown |
| `BREAKER_HALF_OPEN_CALLS` | `1` | Max requests in half-open state |

## Testing

```bash
# Unit + integration tests (with race detector where CGO available)
go test -race ./...

# Specific integration chaos test
go test -v ./test/integration/... -run TestIntegrationChaosTwoProvidersFail

# All tests
go test ./...
```

## Project Structure

```
goaggregator/
├── cmd/
│   ├── api/main.go              # API server entry point
│   └── mockprovider/main.go     # Mock provider server (5 instances via docker-compose)
├── internal/
│   ├── config/                  # Environment config
│   ├── domain/                  # FlightResult, HotelResult, Provider interfaces
│   ├── aggregator/              # Fan-out/fan-in core (flight/hotel search services)
│   ├── resilience/
│   │   ├── breaker/             # Circuit breaker wrapper (sony/gobreaker)
│   │   └── limiter/             # Rate limiter wrapper (token bucket)
│   ├── cache/                   # Redis + Noop + Memory cache implementations
│   ├── provider/                # HTTP clients for mock providers
│   └── http/                    # Gin handlers, middleware
├── test/integration/            # Chaos integration tests
├── docker-compose.yml
├── Dockerfile
└── README.md
```

## Interview Talking Points

1. **Fan-out/Fan-in with per-provider timeout** — Goroutine + `context.WithTimeout` per provider ensures one slow provider doesn't block others. Total latency ≈ max(timeout, slowest healthy provider).

2. **Circuit Breaker per provider** — `sony/gobreaker` wraps each HTTP client. After 5 failures (60% threshold), circuit opens → skips HTTP calls entirely for 30s → half-open allows 1 probe → closes on success. Protects both aggregator (no wasted time) and downstream (no added load during outage).

3. **Rate Limiter per provider** — Token bucket prevents thundering herd. Excess requests dropped locally (not sent), so they don't count as failures or trigger circuit breaker.

4. **Short TTL Cache (30-60s)** — Flight/hotel prices change frequently. Cache-aside with short TTL balances freshness vs. load reduction. Cache key includes route + date.

5. **Transparent Response Metadata** — Response includes `providers_called`, `providers_succeeded`, `providers_failed[]`, `cache_hit`, `duration_ms`. Caller knows exactly what happened — graceful degradation is visible, not hidden.

6. **Runtime Controllable Mocks** — `/control` endpoint on each provider enables chaos testing without restarts. Realistic failure injection for CI/CD resilience validation.

7. **No Database by Design** — Aggregation service is stateless read-only. Redis only for cache. Architectural decision: not every service needs its own DB.

## License

MIT