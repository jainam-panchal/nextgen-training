# Testing and Performance

This project now follows standard Go test/benchmark patterns only:
- all performance workloads are in `*_test.go`
- profiling is produced via `go test` flags
- generated artifacts go to `profiles/` (ignored by git)

## 1) Correctness and Race Safety

```bash
go test ./...
go test -race -count=5 ./...
```

## 2) Benchmarks (single command)

```bash
go test -run=^$ -bench=BenchmarkCore -benchmem -count=5 ./internal/auction
```

Benchmarks included:
- `BenchmarkCore_PlaceBid_NoContention`
- `BenchmarkCore_PlaceBid_Contention`
- `BenchmarkCore_RetractLastBid`
- `BenchmarkCore_EndAuction`
- `BenchmarkCore_BrowseCategory_Subtree`
- `BenchmarkCore_LivePublish_Fanout`

## 3) Profiles from Benchmarks (Canonical Paths)

```bash
# all-core profile set
go test -run=^$ -bench=BenchmarkCore -benchtime=1s -benchmem \
  -cpuprofile=profiles/cpu/cpu_all.prof \
  -memprofile=profiles/memory/mem_all.prof \
  -mutexprofile=profiles/mutex/mutex_all.prof \
  ./internal/auction \
  | tee profiles/runs/core_all_profile_run.txt

# contention-only profile set (hot bid path)
go test -run=^$ -bench=BenchmarkCore_PlaceBid_Contention -benchtime=3s -benchmem \
  -cpuprofile=profiles/cpu/cpu.prof \
  -memprofile=profiles/memory/mem.prof \
  -mutexprofile=profiles/mutex/mutex.prof \
  ./internal/auction \
  | tee profiles/runs/core_contention_profile_run.txt
```

Read results:

```bash
go tool pprof -top profiles/cpu/cpu_all.prof > profiles/cpu/cpu_all_top.txt
go tool pprof -top profiles/memory/mem_all.prof > profiles/memory/mem_all_top.txt
go tool pprof -top profiles/mutex/mutex_all.prof > profiles/mutex/mutex_all_top.txt
```

Flame graph:

```bash
go tool pprof -svg profiles/cpu/cpu_all.prof > profiles/cpu/cpu_all_flame.svg
```

Goroutine snapshot:

```bash
# run server in one shell: go run ./cmd/server
curl -fsS http://localhost:8081/debug/pprof/goroutine?debug=1 > profiles/goroutine/goroutine_debug.txt
```

## 4) Evidence Snippets

Latest benchmark outputs are stored in:
- `profiles/runs/core_benchmark.txt`
- `profiles/runs/fanout_stable.txt`

Latest profile summaries are stored in:
- `profiles/cpu/cpu_all_top.txt`
- `profiles/memory/mem_all_top.txt`
- `profiles/mutex/mutex_all_top.txt`
- `profiles/goroutine/goroutine_debug.txt`
## 5) Mid/High Scale Test

The project already includes a deterministic concurrent simulation test:

```bash
go test -run TestConcurrentBiddingSimulation50Users5Items -race ./internal/auction
```

This covers 50 users, 5 items, and concurrent bidding with final winner validation.
