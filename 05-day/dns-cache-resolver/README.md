# DNS Cache Resolver (Day 5)

## Run

```bash
go run ./cmd/simulator
go test ./...
go test -bench=BenchmarkResolve -benchmem ./internal/resolver
```

## What is implemented

- Custom hash map store with:
- FNV-1a string hash
- bucket array + chaining (linked list per bucket)
- resize when load factor exceeds `0.75`
- Built-in map store for comparison
- DNS resolver with:
- cache-first resolve
- upstream fallback
- wildcard lookup (`*.example.com`)
- lazy TTL expiration on lookup
- periodic cleanup (`time.Ticker`)
- cache stats (hits/misses/rates/entry count/memory estimate)
- `AddRecord(domain, ip, ttl)` API
- JSON tags on `DNSRecord`

## Notes

- Higher cache `HitCount` means the record was returned from cache more often.
- Wildcard lookup is fallback only; exact domain lookup is attempted first.
- Cleanup runs in background every configured interval (simulator uses 30s).

## Benchmarks

Run:

  ```bash
  go test -bench=BenchmarkResolve -benchmem ./internal/resolver

  Output:
  goos: linux
  goarch: amd64
  pkg: dns-cache-resolver/internal/resolver
  cpu: 11th Gen Intel(R) Core(TM) i7-11850H @ 2.50GHz
  BenchmarkResolve/CustomMapStore/n=100-16                 8388062               135.6 ns/op             0 B/op          0 allocs/op
  BenchmarkResolve/InbuiltMapStore/n=100-16               10446400               113.3 ns/op             0 B/op          0 allocs/op
  BenchmarkResolve/CustomMapStore/n=1000-16                7679671               144.2 ns/op             0 B/op          0 allocs/op
  BenchmarkResolve/InbuiltMapStore/n=1000-16               9482930               124.3 ns/op             0 B/op          0 allocs/op
  BenchmarkResolve/CustomMapStore/n=10000-16               7347439               155.9 ns/op             0 B/op          0 allocs/op
  BenchmarkResolve/InbuiltMapStore/n=10000-16              8742204               133.3 ns/op             0 B/op          0 allocs/op
  BenchmarkResolve/CustomMapStore/n=100000-16              5815402               213.8 ns/op             0 B/op          0 allocs/op
  BenchmarkResolve/InbuiltMapStore/n=100000-16             7582772               160.9 ns/op             0 B/op          0 allocs/op
  BenchmarkResolve/CustomMapStore/n=500000-16              3411124               337.7 ns/op             0 B/op          0 allocs/op
  BenchmarkResolve/InbuiltMapStore/n=500000-16             3722890               321.1 ns/op             0 B/op          0 allocs/op
  BenchmarkResolve/CustomMapStore/n=1000000-16             3343378               361.3 ns/op             0 B/op          0 allocs/op
  BenchmarkResolve/InbuiltMapStore/n=1000000-16            3151083               387.8 ns/op             0 B/op          0 allocs/op
  PASS
  ok      dns-cache-resolver/internal/resolver    31.917s
```

  Short interpretation:

  - Built-in map is faster from n=100 to n=500000.
  - At n=1000000, custom map is slightly faster in this run (361.3 ns/op vs 387.8 ns/op).
  - Both implementations show 0 B/op and 0 allocs/op on this hot-cache resolve path.


What to compare:

- `CustomMapStore` vs `InbuiltMapStore`
- `ns/op` for throughput
- `B/op` for bytes allocated
- `allocs/op` for allocation count

## Collision handling proof

Collision-chain behavior is validated in:

- `internal/cache_store/custom_cache_store_test.go`

It intentionally finds colliding keys for the same bucket and confirms that:

- deleting one key does not remove the other collided entry
- chain integrity is preserved
