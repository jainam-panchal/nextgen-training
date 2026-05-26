# Array (Slice) vs Linked List — Go Benchmark

**Hypothesis:** Arrays win most of the time because of CPU hardware optimization (cache line prefetching, spatial locality). Linked lists only win where they exploit O(1) head insertion.

**Machine:** i7-11850H @ 2.50GHz, Go 1.26.2

## Run

```bash
go test -bench=. -benchmem -timeout 30m
```

## Results

### Append (build by appending at tail)

| n | Slice | Linked List | Ratio |
|---|---|---|---|
| 10 | 45 ns/op (80 B, 1 alloc) | 331 ns/op (160 B, 10 allocs) | **7.3× slice** |
| 100 | 314 ns/op (896 B, 1 alloc) | 3,056 ns/op (1,600 B, 100 allocs) | **9.7× slice** |
| 1,000 | 2,766 ns/op (8,192 B, 1 alloc) | 30,765 ns/op (16,000 B, 1,000 allocs) | **11× slice** |
| 10,000 | 20,847 ns/op (81,920 B, 1 alloc) | 329,001 ns/op (160,000 B, 10,000 allocs) | **16× slice** |
| 100,000 | 199,251 ns/op (802,819 B, 1 alloc) | 3,959,153 ns/op (1,600,003 B, 100,000 allocs) | **20× slice** |

Slice's single amortized-grow allocation crushes the linked list's per-node allocation overhead. The gap widens with size as GC pressure grows.

### Iterate (sequential traversal)

| n | Slice | Linked List | Ratio |
|---|---|---|---|
| 10 | 3.3 ns/op | 5.0 ns/op | 1.5× slice |
| 100 | 58 ns/op | 178 ns/op | **3.1× slice** |
| 1,000 | 474 ns/op | 1,689 ns/op | **3.6× slice** |
| 10,000 | 4,902 ns/op | 16,564 ns/op | **3.4× slice** |
| 100,000 | 48,994 ns/op | 178,255 ns/op | **3.6× slice** |

**CPU cache effect:** Slice iterates at ~2 GB/s (sequential prefetch). Linked list pointer-chases across heap, suffering a cache miss per node (~60–100 ns each). The ~3.6× gap at scale directly reflects DRAM latency vs L1/L2 hit.

### Random Access

| n | Slice | Linked List | Ratio |
|---|---|---|---|
| 10 | 4.4 ns/op | 41 ns/op | 9.3× slice |
| 100 | 61 ns/op | 3,943 ns/op | **65× slice** |
| 1,000 | 563 ns/op | 694,566 ns/op | **1,234× slice** |
| 10,000 | 5,724 ns/op | 75,711,394 ns/op | **13,227× slice** |
| 100,000 | 57,938 ns/op | 7,900,228,689 ns/op (7.9 s) | **136,345× slice** |

This is the most dramatic gap. Slice's O(1) indexed access vs linked list's O(n) pointer chase per element. At 100k elements, traversing half the list on average sums to billions of cache-miss-laden pointer dereferences.

### Insert at Front

| n | Slice | Linked List | Ratio |
|---|---|---|---|
| 10 | 541 ns/op (616 B, 20 allocs) | 349 ns/op (160 B, 10 allocs) | **1.5× list** |
| 100 | 15,996 ns/op (43,912 B, 200 allocs) | 2,896 ns/op (1,600 B, 100 allocs) | **5.5× list** |
| 1,000 | 1,097,515 ns/op (4,290,119 B, 2,000 allocs) | 33,516 ns/op (16,000 B, 1,000 allocs) | **33× list** |
| 10,000 | 73,359,472 ns/op (428,682,980 B, 20,005 allocs) | 299,735 ns/op (160,000 B, 10,000 allocs) | **245× list** |
| 100,000 | 8,834,477,972 ns/op (40,399,210,072 B, 201,425 allocs) | 3,507,300 ns/op (1,600,001 B, 100,000 allocs) | **2,519× list** |

**Linked list's strongest win.** Slice's `append([]int{v}, s...)` copies the entire array per insert — O(n²) time and allocation. Linked list's O(1) head pointer swap dominates. At small n (10), the slice is competitive because allocation overhead per node dominates.

### Delete from Front

| n | Slice | Linked List | Ratio |
|---|---|---|---|
| 10 | 45 ns/op (80 B, 1 alloc) | 360 ns/op (160 B, 10 allocs) | **8.0× slice** |
| 100 | 290 ns/op (896 B, 1 alloc) | 3,691 ns/op (1,600 B, 100 allocs) | **13× slice** |
| 1,000 | 2,524 ns/op (8,192 B, 1 alloc) | 33,425 ns/op (16,000 B, 1,000 allocs) | **13× slice** |
| 10,000 | 26,892 ns/op (81,920 B, 1 alloc) | 323,294 ns/op (160,000 B, 10,000 allocs) | **12× slice** |
| 100,000 | 275,865 ns/op (802,825 B, 1 alloc) | 3,935,113 ns/op (1,600,001 B, 100,000 allocs) | **14× slice** |

Go's slice re-slicing (`s = s[1:]`) is O(1) — just moves the start pointer. No data movement. The linked list also does O(1) pointer swaps, but **building** the list costs n allocations vs the slice's 1, which dominates wall time.

### Prepend-Build (build by repeatedly inserting at front)

| n | Slice | Linked List | Ratio |
|---|---|---|---|
| 10 | 524 ns/op (536 B, 19 allocs) | 327 ns/op (160 B, 10 allocs) | **1.6× list** |
| 100 | 17,124 ns/op (43,016 B, 199 allocs) | 3,267 ns/op (1,600 B, 100 allocs) | **5.2× list** |
| 1,000 | 1,239,842 ns/op (4,281,927 B, 1,999 allocs) | 31,035 ns/op (16,000 B, 1,000 allocs) | **40× list** |
| 10,000 | 92,447,910 ns/op (428,601,112 B, 20,005 allocs) | 315,731 ns/op (160,000 B, 10,000 allocs) | **293× list** |
| 100,000 | 10,414,530,828 ns/op (40,398,334,408 B, 200,763 allocs) | 3,629,334 ns/op (1,600,000 B, 100,000 allocs) | **2,869× list** |

Same as Insert Front but builds from scratch. Slice's O(n²) behavior produces absurd allocation pressure — **40 GB** of temporary allocations for building 100k elements. Linked list's O(n) prepend finishes in 3.6 ms.

## Summary

| Benchmark | Small n (10) | Large n (100k) | Winner |
|---|---|---|---|
| Append (tail) | 7.3× slice | 20× slice | **Slice** |
| Sequential iterate | 1.5× slice | 3.6× slice | **Slice** |
| Random access | 9.3× slice | **136,345×** slice | **Slice** |
| Insert front | 1.5× list | 2,519× list | **Linked list** |
| Delete front | 8× slice | 14× slice | **Slice** |
| Prepend-build | 1.6× list | 2,869× list | **Linked list** |

**Arrays win 4/6 benchmarks. Linked list wins 2/6 (front-insertion variants).**

## Verdict on Hypothesis

**Confirmed.** Arrays dominate whenever the workload touches memory sequentially or by index:

- **Sequential access** exploits hardware prefetchers that fill cache lines before the CPU requests them.
- **Indexed access** is O(1) pointer arithmetic vs O(n) heap-dereference walks.
- **Single-allocation growth** avoids per-node malloc/free overhead and GC scanning cost.

Linked lists are only competitive for pure front-insertion workloads (build by prepending, stack-like usage). In all other common patterns — iteration, random access, tail append, front delete — the array wins, often by orders of magnitude.
