# GoAggregator — Kế Hoạch Triển Khai Dự Án
### Flight/Hotel Search & Price Comparison Aggregator (Go, Travel Tech Portfolio Project)

> Mục tiêu: xây service tổng hợp kết quả từ nhiều "provider" song song, chịu lỗi tốt (1 provider
> chết/chậm không kéo sập cả request) — thể hiện tư duy resilience, thứ mà 1 project CRUD thường
> không có cơ hội show.

**Timeline đề xuất:** 5 sprint x 1 tuần. Đây là project thứ 2 — nếu làm song song/nối tiếp với
GoBooking thì tổng ~8-10 tuần cho cả 2. Nếu deadline gấp, làm xong Sprint 0-2 là đã đủ chứng minh
được "fan-out/fan-in + circuit breaker hoạt động", đủ để đưa vào CV.

**Điểm khác biệt có chủ đích so với GoBooking:** service này **không có Postgres**, chỉ có Redis
cache — vì đây là service tổng hợp/đọc (aggregation, stateless), không sở hữu dữ liệu. Đây cũng là
1 quyết định kiến trúc đáng nói khi phỏng vấn: không phải service nào cũng cần DB riêng.

---

## 0. Phạm Vi Dự Án

**Trong phạm vi (MVP):**
- 3-5 mock provider server độc lập, có thể chỉnh độ trễ/tỷ lệ lỗi runtime
- Aggregator gọi song song các provider, timeout riêng từng provider
- Circuit breaker per-provider (tự ngắt khi provider lỗi liên tục)
- Rate limiter per-provider (không dội quá nhiều request vào 1 provider)
- Cache kết quả search theo TTL ngắn
- Response trả kèm metadata minh bạch (provider nào thành công/thất bại)

**Ngoài phạm vi:**
- Provider thật (chỉ mock)
- Thanh toán/booking (đây là service search-only, ghép với GoBooking là 1 hệ sinh thái)
- UI

---

## 1. Kiến Trúc & Tech Stack

| Thành phần | Lựa chọn | Lý do |
|---|---|---|
| Ngôn ngữ | Go 1.22+ | |
| Web framework | Gin | đồng bộ với GoBooking |
| Circuit breaker | `sony/gobreaker` | thư viện chuẩn, đủ để demo state machine Closed/Open/Half-Open |
| Rate limiter | `golang.org/x/time/rate` | token bucket per-provider |
| Cache | Redis (TTL 30-60s) | giảm tải provider, không cần cache lâu vì giá vé/phòng thay đổi liên tục |
| Concurrency | goroutine + channel + `context.WithTimeout` | fan-out/fan-in pattern |
| Logging | slog | structured, có provider name + duration mỗi lần gọi |
| Test | testing + testify + `-race` | test resilience (timeout, circuit trip) |
| Container | Docker Compose | chạy API + 5 mock provider + Redis bằng 1 lệnh |

**Sơ đồ luồng:**
```
Client → GoAggregator API → Redis (cache-aside, check trước)
                          → [cache miss] → fan-out song song tới:
                                Provider A (nhanh, ổn định)
                                Provider B (chậm ~1.5s)
                                Provider C (10% lỗi ngẫu nhiên)
                                Provider D (circuit đang mở, bị skip)
                                Provider E (timeout ngẫu nhiên)
                          → fan-in gom kết quả → merge/sort theo giá → cache lại → trả về
```

---

## 2. Cấu Trúc Thư Mục

```
goaggregator/
├── cmd/
│   ├── api/main.go
│   └── mockprovider/main.go   # server giả lập, chạy 5 instance config khác nhau
├── internal/
│   ├── config/
│   ├── domain/                 # FlightResult, HotelResult, Provider interface
│   ├── aggregator/              # fan-out/fan-in core
│   ├── resilience/
│   │   ├── breaker/             # wrap gobreaker per-provider
│   │   └── limiter/             # wrap rate.Limiter per-provider
│   ├── cache/                    # redis cache-aside
│   ├── provider/                  # HTTP client gọi từng mock provider
│   └── http/
│       ├── handler/
│       └── middleware/           # request-id, logging, recover
├── test/
│   ├── integration/
│   └── resilience/                # test timeout/circuit breaker/partial-failure
├── docker-compose.yml
├── Dockerfile
├── .github/workflows/ci.yml
└── README.md
```

---

## 3. Thiết Kế Mock Provider (nền tảng để test resilience)

Đây là phần hay bị bỏ qua nhưng lại là thứ làm project này "thật" — mock provider không chỉ trả
data random mà còn có **control endpoint** để bật/tắt độ trễ, tỷ lệ lỗi lúc runtime (giống chaos
engineering thu nhỏ), phục vụ viết test resilience mà không cần restart server.

```go
// cmd/mockprovider/main.go
// GET  /search           -> trả kết quả giả (giá random quanh 1 mốc base price)
// POST /control           -> body: {"latency_ms": 2000, "error_rate": 0.3}
//                            đổi hành vi provider ngay lập tức, dùng trong test
```

Chạy 5 instance qua docker-compose với env khác nhau: `provider-fast`, `provider-slow`,
`provider-flaky`, `provider-timeout`, `provider-down` — mô phỏng đúng các tình huống thực tế.

---

## 4. Git Workflow

Giống GoBooking: `main` + `feature/<sprint>-<mô-tả>`, Conventional Commits, tag theo sprint
(`v0.1.0`...), mỗi user story = 1 PR tự review.

---

## 5. Sprint Planning

### Sprint 0 — Nền tảng & Mock Providers (2-3 ngày)
**Mục tiêu:** có 5 mock provider chạy được qua Docker Compose, aggregator gọi thẳng qua HTTP thành công 1 provider.

| ID | User Story | Tasks | Branch | Điểm |
|---|---|---|---|---|
| US-01 | Là dev, tôi muốn có mock provider server trả kết quả giả kèm độ trễ/lỗi cấu hình được | `cmd/mockprovider`, endpoint `/search` + `/control` | `feature/s0-mock-provider` | 5 |
| US-02 | Là dev, tôi muốn 5 provider chạy song song qua Docker Compose với hành vi khác nhau | docker-compose service x5 với env riêng | `feature/s0-compose-providers` | 3 |
| US-03 | Là dev, tôi muốn API skeleton có health check + config + logger | `internal/config`, slog, `/healthz` | `feature/s0-base-server` | 3 |
| US-04 | Là dev, tôi muốn CI lint + test tự động | GitHub Actions cơ bản | `feature/s0-ci-pipeline` | 2 |

**DoD:** `docker compose up` chạy 5 provider + API, `curl` từng provider `/search` trả JSON hợp lệ.

---

### Sprint 1 — Fan-out/Fan-in Core (1 tuần)
**Mục tiêu:** aggregator gọi song song tất cả provider, 1 provider chậm không chặn kết quả các provider khác.

| ID | User Story | Tasks | Branch | Điểm |
|---|---|---|---|---|
| US-05 | Là dev, tôi muốn định nghĩa `Provider` interface thống nhất | `domain.Provider`, implement HTTP client cho từng mock | `feature/s1-provider-interface` | 3 |
| US-06 | Là khách, tôi muốn search flight gọi song song nhiều provider và có timeout riêng từng provider | Fan-out goroutine + `context.WithTimeout` mỗi provider (như snippet đã có), fan-in qua channel | `feature/s1-fanout-fanin-flights` | 8 |
| US-07 | Là khách, tôi muốn kết quả được gộp và sắp xếp theo giá | Merge + sort ascending theo price, dedupe nếu trùng | `feature/s1-merge-sort-results` | 3 |
| US-08 | Là dev, tôi muốn chứng minh provider chậm không làm chậm toàn bộ response | Test: 1 provider set latency 5s (qua `/control`), assert tổng response time ≈ timeout config (~3s), không phải 5s | `feature/s1-timeout-isolation-test` | 5 |

**AC mẫu (US-08):**
- Given provider B được set `latency_ms=5000` qua control endpoint
- And aggregator timeout mỗi provider là 3s
- When gọi `SearchFlights`
- Then response trả về trong ~3s (không đợi provider B), kết quả thiếu provider B nhưng vẫn có
  kết quả từ các provider khác

---

### Sprint 2 — Circuit Breaker & Rate Limiter (1 tuần, trọng tâm)
**Mục tiêu:** hệ thống tự bảo vệ khi 1 provider liên tục lỗi, không dội quá tải vào provider.

| ID | User Story | Tasks | Branch | Điểm |
|---|---|---|---|---|
| US-09 | Là hệ thống, tôi muốn tự ngắt gọi tới provider lỗi liên tục | Wrap `gobreaker.CircuitBreaker` per-provider, `ReadyToTrip`: mở mạch khi tỷ lệ lỗi ≥60% trong 5 request gần nhất | `feature/s2-circuit-breaker` | 8 |
| US-10 | Là hệ thống, tôi muốn circuit tự thử lại provider sau thời gian nghỉ | Config `Timeout` (cooldown) + `MaxRequests` cho trạng thái Half-Open | `feature/s2-circuit-half-open` | 3 |
| US-11 | Là hệ thống, tôi muốn không dội quá N request/giây vào 1 provider | `rate.Limiter` per-provider, request vượt quota bị bỏ qua provider đó (không lỗi toàn bộ) | `feature/s2-rate-limiter-provider` | 5 |
| US-12 | Là dev, tôi muốn test chứng minh circuit breaker mở đúng lúc | Set provider `error_rate=1.0`, gọi 5 lần, assert lần thứ 6 bị skip ngay (không gọi HTTP thật) — verify bằng đếm số lần handler mock provider nhận request | `feature/s2-circuit-breaker-test` | 5 |

**AC mẫu (US-12):**
- Given provider C có `error_rate=1.0`
- When gọi aggregator 6 lần liên tiếp
- Then 5 lần đầu có gọi HTTP tới provider C (đều lỗi), từ lần thứ 6 circuit ở trạng thái Open →
  bị skip, không tốn round-trip HTTP nữa (assert qua counter ở mock provider hoặc qua log)

---

### Sprint 3 — Caching & API Layer (1 tuần)
**Mục tiêu:** hoàn thiện API public, giảm tải provider bằng cache.

| ID | User Story | Tasks | Branch | Điểm |
|---|---|---|---|---|
| US-13 | Là khách, tôi muốn gọi `GET /search/flights?from=&to=&date=` | Handler + validate query param | `feature/s3-search-flights-api` | 3 |
| US-14 | Là khách, tôi muốn gọi `GET /search/hotels?city=&checkin=&checkout=` | Handler tương tự cho hotel | `feature/s3-search-hotels-api` | 3 |
| US-15 | Là hệ thống, tôi muốn cache kết quả search theo route+date để giảm tải provider | Cache-aside: check Redis trước, miss thì fan-out rồi set lại TTL 30-60s | `feature/s3-redis-cache-aside` | 5 |
| US-16 | Là khách, tôi muốn biết provider nào thành công/thất bại trong response | Field `meta`: `providers_called`, `providers_succeeded`, `providers_failed`, `cache_hit`, `duration_ms` | `feature/s3-response-meta-transparency` | 3 |

**Response mẫu (US-16) — điểm cộng senior-level, thể hiện tư duy "graceful degradation":**
```json
{
  "results": [ { "provider": "provider-fast", "price": 1250000, "...": "..." } ],
  "meta": {
    "providers_called": 5,
    "providers_succeeded": 3,
    "providers_failed": ["provider-flaky", "provider-down"],
    "cache_hit": false,
    "duration_ms": 312
  }
}
```

---

### Sprint 4 — Observability & Hardening (1 tuần)
**Mục tiêu:** sẵn sàng demo, có bằng chứng đo lường được resilience hoạt động.

| ID | User Story | Tasks | Branch | Điểm |
|---|---|---|---|---|
| US-17 | Là dev, tôi muốn log mỗi lần gọi provider kèm duration + kết quả | slog structured log: `request_id, provider, duration_ms, status, circuit_state` | `feature/s4-structured-logging` | 3 |
| US-18 | Là dev, tôi muốn xem trạng thái circuit breaker của từng provider | `GET /providers/status` (debug/admin endpoint) trả state Closed/Open/Half-Open mỗi provider | `feature/s4-provider-status-endpoint` | 3 |
| US-19 | Là dev, tôi muốn integration test full flow có chaos | Test: bật lỗi 2/5 provider qua `/control`, gọi search, assert response vẫn 200 với `providers_succeeded=3` | `feature/s4-integration-chaos-test` | 5 |
| US-20 | Là dev, tôi muốn README có demo script bật/tắt lỗi provider trực tiếp | README: kiến trúc, cách chạy, script `curl` demo circuit breaker trip trực tiếp | `feature/s4-readme-demo-script` | 3 |

---

### Sprint 5 — Stretch Goals (tuỳ thời gian)

- OpenTelemetry tracing xuyên suốt aggregator → provider (span cho mỗi lệnh gọi song song)
- Prometheus metrics: histogram latency per-provider, gauge trạng thái circuit breaker
- Load test bằng `k6`: so sánh p99 latency khi 1 provider khỏe mạnh vs khi 1 provider die hẳn,
  chứng minh circuit breaker giữ latency ổn định thay vì tăng vọt — số liệu này rất mạnh để bỏ
  vào CV ("giảm p99 latency X% khi provider lỗi nhờ circuit breaker")
- Kết nối với GoBooking: aggregator trả kết quả search → GoBooking xử lý hold/booking, thành 1
  hệ sinh thái 2 service hoàn chỉnh

---

## 6. API Spec Đầy Đủ

| Method | Path | Mô tả |
|---|---|---|
| GET | `/healthz` | health check |
| GET | `/search/flights?from=&to=&date=` | tổng hợp kết quả chuyến bay từ các provider |
| GET | `/search/hotels?city=&checkin=&checkout=` | tổng hợp kết quả phòng |
| GET | `/providers/status` | debug: trạng thái circuit breaker từng provider |
| POST (mock provider) | `/control` | chỉnh latency/error_rate runtime, chỉ dùng nội bộ để test |

---

## 7. Testing Strategy

| Loại | Mục tiêu |
|---|---|
| Unit test | merge/sort logic, refund... (không áp dụng ở đây), circuit breaker config |
| Resilience test | timeout isolation (US-08), circuit breaker trip/reset (US-12) |
| Integration test (chaos) | bật lỗi ngẫu nhiên nhiều provider cùng lúc, assert response vẫn hợp lệ với partial results |
| Load test (stretch) | đo p99 latency có/không có circuit breaker |

```bash
go vet ./...
golangci-lint run
go test -race -cover ./...
```

---

## 8. CI/CD Pipeline (tóm tắt, giống GoBooking)

```yaml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    services:
      redis: { image: redis:7, ports: ["6379:6379"] }
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }
      - run: go vet ./...
      - run: golangci-lint run
      - run: go test -race -cover ./...
      - run: docker build -t goaggregator:ci .
```

---

## 9. Tài Liệu Cần Chuẩn Bị Cho CV/Phỏng Vấn

1. **README.md** — kiến trúc, sơ đồ fan-out/fan-in, cách chạy demo, script chaos-test
2. **Bằng chứng đo lường:** log/screenshot circuit breaker chuyển trạng thái Closed → Open →
   Half-Open → Closed khi chạy demo script
3. **So sánh có/không circuit breaker** (nếu làm Sprint 5 load test) — số liệu latency cụ thể

---

## 10. Câu Chuyện Để Kể Khi Phỏng Vấn

1. **Vấn đề:** search cần gọi nhiều nguồn dữ liệu (nhiều hãng bay/OTA), nếu gọi tuần tự thì chậm,
   nếu gọi song song không kiểm soát thì 1 provider chết có thể kéo theo timeout toàn bộ request
   hoặc dội quá tải khi provider đang gặp sự cố (cascading failure)
2. **Giải pháp:**
   - Fan-out/fan-in bằng goroutine + channel, mỗi provider có `context.WithTimeout` riêng nên
     provider chậm chỉ mất kết quả của chính nó, không ảnh hưởng tổng thời gian phản hồi
   - Circuit breaker per-provider: sau N lần lỗi liên tiếp, tự "ngắt mạch" — vừa bảo vệ hệ thống
     của mình (không tốn thời gian chờ 1 provider chắc chắn sẽ lỗi), vừa bảo vệ provider đang gặp
     sự cố (không dội thêm traffic vào lúc nó đang yếu — nguyên lý giống backpressure)
   - Cache TTL ngắn giảm tải cho provider ở các query trùng lặp (VD nhiều người cùng search
     SGN→HAN cùng ngày trong 1 phút)
3. **Bằng chứng:** test giả lập provider lỗi 100%, chứng minh circuit breaker mở đúng sau ngưỡng
   cấu hình và ngừng gọi HTTP thật — không chỉ code chạy được mà đo được hành vi đúng như thiết kế
4. **Trade-off đáng nói:** cache TTL ngắn (30-60s) vì giá vé/phòng biến động liên tục — khác với
   cache thông thường ưu tiên TTL dài để giảm tải, ở đây phải cân bằng giữa độ tươi của dữ liệu và
   hiệu năng, đây là quyết định có chủ đích chứ không phải mặc định

---

## 11. Checklist Timeline Tổng

- [ ] Tuần 1: Sprint 0 + Sprint 1 (mock providers + fan-out/fan-in)
- [ ] Tuần 2: Sprint 2 (circuit breaker + rate limiter) ← **ưu tiên cao nhất nếu thiếu thời gian**
- [ ] Tuần 3: Sprint 3 (caching + API hoàn chỉnh)
- [ ] Tuần 4: Sprint 4 (observability + docs) — sẵn sàng đưa vào CV
- [ ] Tuần 5 (nếu còn thời gian): Sprint 5 stretch goals
