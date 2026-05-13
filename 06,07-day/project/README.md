# Ride Sharing Dispatch (Day 6/7)

In-memory ride dispatch engine in Go with custom data structures for request ordering, driver lookup, active ride tracking, and history.

## Quickstart

```bash
go run ./cmd
go test ./...
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | tail -n 1
```

## System Architecture

```text
Rider Request
    |
    v
Dispatcher ---------------------------------------------+
    |                                                   |
    | enqueue request                                   | nearest-driver search
    v                                                   v
RideRequestPriorityQueue (Min-Heap by RequestTime)   DriverStore
                                                     - driversByID (HashMap)
                                                     - availableDriversByBlock (Zone Index)
    |
    | assign matched request
    v
ActiveRideTracker (DoublyLinkedList + rideID->node map)
    |
    | complete ride
    v
RideHistoryStore (HashMap)
```

## Why These Data Structures

- `HashMap (driversByID)`: fast driver lookup/update by ID, simple state management.
- `Zone index map`: narrows nearest-driver scan to neighboring geo buckets instead of full scan.
- `Min-Heap queue`: always serves oldest pending request first in `O(log n)`.
- `DLL + map for active rides`: keeps insertion order and enables `O(1)` removal by ride ID.
- `HashMap history`: fast save/get for rides; full scans needed for aggregate queries.

## Complexity Summary

| Operation | Complexity |
|---|---|
| Register/Get Driver | O(1) average |
| Go online/offline/assign/release | O(1) average |
| Enqueue ride request | O(log n) |
| Peek/Pop next request | O(1) / O(log n) |
| Find nearest driver | O(k), `k = candidates in scanned blocks` |
| Add active ride | O(1) |
| Remove active ride by ID | O(1) |
| Save/Get history ride | O(1) average |
| List rides / ListByDriver | O(m) |

## Tests and Coverage

Current total coverage:

```text
69.7% of statements
```

Package coverage snapshot:

- `internal/dispatch`: 38.0%
- `internal/driver`: 88.3%
- `internal/ridequeue`: 87.0%
- `internal/rides`: 95.7%
- `internal/ds/heap`: 86.7%
- `internal/ds/linkedlist`: 91.8%
- `internal/config`: 100.0%
- `internal/geo`: 100.0%
- `internal/rider`: 100.0%

## Benchmarks

Benchmark command:

```bash
go test -bench=. -benchmem ./...
```

Latest output snapshot:

```text
goos: linux
goarch: amd64
cpu: 11th Gen Intel(R) Core(TM) i7-11850H @ 2.50GHz

pkg: ride-sharing/internal/dispatch
BenchmarkDispatcherRequestAndProcess/drivers=10-16      146686    7189 ns/op   2657 B/op  8 allocs/op
BenchmarkDispatcherRequestAndProcess/drivers=100-16     113379   10769 ns/op   2650 B/op  8 allocs/op
BenchmarkDispatcherRequestAndProcess/drivers=1000-16     22771   44270 ns/op   2629 B/op  8 allocs/op

pkg: ride-sharing/internal/driver
BenchmarkFindNearestAvailable/n=100-16                 337381    3752 ns/op   2048 B/op  1 allocs/op
BenchmarkFindNearestAvailable/n=1000-16                312525    4008 ns/op   2048 B/op  1 allocs/op
BenchmarkFindNearestAvailable/n=10000-16               294030    4039 ns/op   2048 B/op  1 allocs/op

pkg: ride-sharing/internal/ridequeue
BenchmarkRideRequestQueuePushPop/n=100-16             2012374   603.4 ns/op    144 B/op  5 allocs/op
BenchmarkRideRequestQueuePushPop/n=1000-16            2086239   600.1 ns/op    144 B/op  5 allocs/op
BenchmarkRideRequestQueuePushPop/n=10000-16           1656345   720.6 ns/op    144 B/op  5 allocs/op

pkg: ride-sharing/internal/rides
BenchmarkActiveRideTrackerAddRemove/n=100-16           950156    1079 ns/op    288 B/op  8 allocs/op
BenchmarkActiveRideTrackerAddRemove/n=1000-16          477085    2456 ns/op    288 B/op  8 allocs/op
BenchmarkActiveRideTrackerAddRemove/n=10000-16          68458   17032 ns/op    288 B/op  8 allocs/op
```
