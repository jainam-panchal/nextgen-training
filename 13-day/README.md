# Day 14 - Real-Time Auction Platform API

Detailed technical architecture:
- [Architecture Deep Dive](docs/ARCHITECTURE.md)
- [Testing and Performance Report](docs/TESTING_AND_PERFORMANCE.md)
- [SSE PoC](docs/SSE_POC.md)

## What Is Implemented
- Core domain from Day 13 is preserved and reused:
  - Max-heap per item for highest bid.
  - HashMaps for users/items/bids.
  - Category tree for browse-by-subtree.
  - Linked list for bid history tracking.
  - Undo stack per user+item for retract.
- REST API implemented using `net/http`.
- SSE endpoint implemented for live updates.
- Middleware stack implemented:
  - panic recovery
  - request logging
  - auth simulation (`X-User-ID`)
  - generic token-bucket rate limiting configured for bid route (`5 bids/user/min`)
  - response timing header
- pprof endpoint enabled under `/debug/pprof/*`.

## Requirement Matrix
| Requirement | Status | Evidence |
|---|---|---|
| Day 13 data models (`User`,`Item`,`Bid`) | Met | `internal/models/models.go` |
| Day 13 DS usage (heap/hashmap/tree/stack/linked list) | Met | `internal/auction/store.go`, `internal/ds/*` |
| Place/Retract/End/Browse core logic | Met | `internal/auction/service.go` |
| Concurrent bid handling with per-item mutex | Met | `internal/auction/service.go` (`getItemLock`) |
| Day 13 mandatory tests (100 goroutines, retract, end, race) | Met | `service_test.go`, `service_simulation_test.go`, `go test -race -count=5 ./...` |
| Day 14 REST endpoints | Met | `internal/handlers/handlers.go` |
| Day 14 SSE fan-out with channel per watcher | Met | `SubscribeItem/UnsubscribeItem`, `publishToWatchers` |
| Day 14 middleware stack | Met | `internal/middleware/middleware.go`, `cmd/server/main.go` |
| Day 14 concurrent simulation (50 users, 5 items, 10 bids, ~30s) | Met | `TestConcurrentBiddingSimulation50Users5Items` |
| Day 14 profiling outputs (CPU/Memory/Mutex/Goroutine/Flame) | Met | `profiles/cpu/*`, `profiles/memory/*`, `profiles/mutex/*`, `profiles/goroutine/*` |
| Benchmark throughput and evidence | Met | `profiles/runs/core_benchmark.txt`, `profiles/runs/fanout_stable.txt` |
| Trade-off analysis + architecture docs | Met | `docs/ARCHITECTURE.md`, `docs/TESTING_AND_PERFORMANCE.md` |

## API Endpoints
- `POST /users`
- `POST /items`
- `POST /items/:id/bid`
- `DELETE /items/:id/bid/last`
- `GET /items/:id`
- `GET /items?category=X`
- `GET /items/:id/bids`
- `POST /items/:id/end`
- `GET /stats`
- `GET /items/:id/live` (SSE)

## Concurrency and Safety
- Lock order invariant: `itemLock(itemID)` then `store.mu`.
- Non-blocking event publish to avoid blocking bid path.
- SSE fan-out uses buffered watcher channels and disconnect cleanup.
- Bid amount hardening includes finite number validation (`NaN/Inf` rejected).

## Tests Executed
```bash
go test ./...
go test -race -count=3 ./...
```

Current result snapshot:
```text
ok  	realtime-auction/internal/auction
ok  	realtime-auction/internal/handlers
ok  	realtime-auction/internal/ds/heap
ok  	realtime-auction/internal/ds/linkedlist
ok  	realtime-auction/internal/ds/stack
ok  	realtime-auction/internal/ds/tree
```

## Profiling Commands
Regenerate all benchmark/profile artifacts (organized under `profiles/`):
```bash
mkdir -p profiles/{cpu,memory,mutex,goroutine,runs}

# Full core benchmark report
go test -run=^$ -bench=BenchmarkCore -benchtime=1x -benchmem ./internal/auction \
  | tee profiles/runs/core_benchmark.txt

# Stable fanout benchmark report
go test -run=^$ -bench=BenchmarkCore_LivePublish_Fanout -benchmem -benchtime=3s -count=10 ./internal/auction \
  | tee profiles/runs/fanout_stable.txt

# Full-suite profiles (CPU, memory, mutex)
go test -run=^$ -bench=BenchmarkCore -benchtime=1s -benchmem \
  -cpuprofile=profiles/cpu/cpu_all.prof \
  -memprofile=profiles/memory/mem_all.prof \
  -mutexprofile=profiles/mutex/mutex_all.prof \
  ./internal/auction \
  | tee profiles/runs/core_all_profile_run.txt

# Top reports + flame graph
go tool pprof -top profiles/cpu/cpu_all.prof > profiles/cpu/cpu_all_top.txt
go tool pprof -svg profiles/cpu/cpu_all.prof > profiles/cpu/cpu_all_flame.svg
go tool pprof -top profiles/memory/mem_all.prof > profiles/memory/mem_all_top.txt
go tool pprof -top profiles/mutex/mutex_all.prof > profiles/mutex/mutex_all_top.txt

# Contention-only profile set (hot bid path)
go test -run=^$ -bench=BenchmarkCore_PlaceBid_Contention -benchtime=3s -benchmem \
  -cpuprofile=profiles/cpu/cpu.prof \
  -memprofile=profiles/memory/mem.prof \
  -mutexprofile=profiles/mutex/mutex.prof \
  ./internal/auction \
  | tee profiles/runs/core_contention_profile_run.txt

go tool pprof -top profiles/cpu/cpu.prof > profiles/cpu/cpu_top.txt
go tool pprof -top profiles/memory/mem.prof > profiles/memory/mem_top.txt
go tool pprof -top profiles/mutex/mutex.prof > profiles/mutex/mutex_top.txt

# Goroutine snapshot from pprof HTTP endpoint
# Start server in one shell:
#   go run ./cmd/server
# Then in another shell:
curl -fsS http://localhost:8081/debug/pprof/goroutine?debug=1 > profiles/goroutine/goroutine_debug.txt
```

Quick inspect:
```bash
go tool pprof -top profiles/cpu/cpu_all.prof
go tool pprof -top profiles/memory/mem_all.prof
go tool pprof -top profiles/mutex/mutex_all.prof
```

## Bruno Collection
- Folder: `bruno/auction-api`
- Open/import this exact folder in Bruno.
- Local environment defaults to `http://localhost:8081` (matches server default).
- Grouped request folders:
  - `00-system` (`health`, `stats`)
  - `01-users`
  - `02-items`
  - `03-bids`
  - `04-live`

## Profile Findings (Latest)
- CPU (`profiles/cpu/cpu_all_top.txt`): runtime scheduling/GC + lock paths dominate under full benchmark mix.
- Memory (`profiles/memory/mem_all_top.txt`): highest allocation pressure comes from browse/category traversal and bid-path append structures.
- Mutex (`profiles/mutex/mutex_all_top.txt`): contention is concentrated in bid hot path (`PlaceBid`) under parallel benchmarks.
- Goroutine (`profiles/goroutine/goroutine_debug.txt`): bounded server goroutines in idle snapshot; no per-item dispatcher leak path in current design.
