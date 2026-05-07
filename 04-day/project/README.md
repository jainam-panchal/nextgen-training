# Day 4

## Run

```bash
go run ./cmd/simulator
go test ./...
go test -bench=. -benchmem ./internal/router
```

## Benchmarks

Command used:

```bash
go test -bench=. -benchmem ./...
```

Output:

```text
goos: linux
goarch: amd64
pkg: jainamp-panchal/nextgen-training/packet-router/internal/router
cpu: 11th Gen Intel(R) Core(TM) i7-11850H @ 2.50GHz
BenchmarkRouterLinkedList/n=100-16          111291      11451 ns/op      16120 B/op       205 allocs/op
BenchmarkRouterLinkedList/n=1000-16          10000     137961 ns/op     160121 B/op      2005 allocs/op
BenchmarkRouterLinkedList/n=10000-16           850    1252386 ns/op    1600131 B/op     20005 allocs/op
BenchmarkRouterLinkedList/n=100000-16           80   13337502 ns/op   16000123 B/op    200005 allocs/op
BenchmarkRouterLinkedList/n=500000-16           15   69859642 ns/op   80000128 B/op   1000005 allocs/op
BenchmarkRouterLinkedList/n=1000000-16           7  149356718 ns/op  160000152 B/op   2000005 allocs/op
BenchmarkRouterSlice/n=100-16               147002      12643 ns/op      16912 B/op       131 allocs/op
BenchmarkRouterSlice/n=1000-16               12127     126736 ns/op     169296 B/op      1049 allocs/op
BenchmarkRouterSlice/n=10000-16               1074    1118572 ns/op    1703892 B/op     10068 allocs/op
BenchmarkRouterSlice/n=100000-16                96   12505448 ns/op   18434257 B/op    100104 allocs/op
BenchmarkRouterSlice/n=500000-16                15   68924732 ns/op   94106162 B/op    500136 allocs/op
BenchmarkRouterSlice/n=1000000-16                8  143145062 ns/op  188617424 B/op   1000151 allocs/op
BenchmarkRouterRingBuffer/n=100-16          132057      13937 ns/op      16816 B/op       114 allocs/op
BenchmarkRouterRingBuffer/n=1000-16          10000     110491 ns/op     168816 B/op      1029 allocs/op
BenchmarkRouterRingBuffer/n=10000-16          1071     951742 ns/op    1636084 B/op     10044 allocs/op
BenchmarkRouterRingBuffer/n=100000-16           99   11910487 ns/op   16912368 B/op    100061 allocs/op
BenchmarkRouterRingBuffer/n=500000-16           16   66951903 ns/op   83556336 B/op    500074 allocs/op
BenchmarkRouterRingBuffer/n=1000000-16           8  136023215 ns/op  167090672 B/op   1000079 allocs/op
PASS
ok  	jainamp-panchal/nextgen-training/packet-router/internal/router	34.421s
```

## What was noticed

- linked list had the highest allocation count in every run
- slice was strong on smaller and medium sizes
- ring buffer stayed competitive and used less memory than slice at larger sizes
- all implementations got more expensive as packet count increased

## Reorder behavior

- higher priority packets are processed first
- because of that, packets can be reordered across priorities
- inside the same priority queue, FIFO order is preserved
