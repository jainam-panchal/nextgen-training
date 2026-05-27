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
| 10 | 48 ns/op (80 B, 1 alloc) | 375 ns/op (160 B, 10 allocs) | **7.8× slice** |
| 100 | 340 ns/op (896 B, 1 alloc) | 3,590 ns/op (1,600 B, 100 allocs) | **10.5× slice** |
| 1,000 | 3,216 ns/op (8,192 B, 1 alloc) | 39,404 ns/op (16,000 B, 1,000 allocs) | **12× slice** |
| 10,000 | 21,597 ns/op (81,920 B, 1 alloc) | 367,618 ns/op (160,000 B, 10,000 allocs) | **17× slice** |
| 100,000 | 198,054 ns/op (802,819 B, 1 alloc) | 5,256,163 ns/op (1,600,005 B, 100,000 allocs) | **26.5× slice** |

Slice's single amortized-grow allocation crushes the linked list's per-node allocation overhead. The gap widens with size as GC pressure grows.

### Iterate (sequential traversal)

| n | Slice | Linked List | Ratio |
|---|---|---|---|
| 10 | 4.7 ns/op | 6.7 ns/op | 1.4× slice |
| 100 | 46 ns/op | 158 ns/op | **3.4× slice** |
| 1,000 | 499 ns/op | 2,045 ns/op | **4.1× slice** |
| 10,000 | 5,294 ns/op | 19,181 ns/op | **3.6× slice** |
| 100,000 | 57,755 ns/op | 185,152 ns/op | **3.2× slice** |

**CPU cache effect:** Slice iterates at ~2 GB/s (sequential prefetch). Linked list pointer-chases across heap, suffering a cache miss per node (~60–100 ns each). The 3–4× gap at scale directly reflects DRAM latency vs L1/L2 hit.

### Random Access

Uses shuffled indices to avoid worst-case bias (previous version used descending order).

| n | Slice | Linked List | Ratio |
|---|---|---|---|
| 10 | 6.4 ns/op | 41 ns/op | 6.3× slice |
| 100 | 63 ns/op | 3,968 ns/op | **63× slice** |
| 1,000 | 565 ns/op | 780,913 ns/op | **1,383× slice** |
| 10,000 | 6,805 ns/op | 74,594,211 ns/op | **10,962× slice** |
| 100,000 | 81,362 ns/op | 7,451,785,825 ns/op (7.45 s) | **91,588× slice** |

This is the most dramatic gap. Slice's O(1) indexed access vs linked list's O(n) pointer chase per element. At 100k elements, each access walks ~50k nodes on average — billions of cache-miss-laden pointer dereferences.

### Insert at Front

| n | Slice | Linked List | Ratio |
|---|---|---|---|
| 10 | 677 ns/op (616 B, 20 allocs) | 417 ns/op (160 B, 10 allocs) | **1.6× list** |
| 100 | 18,854 ns/op (43,912 B, 200 allocs) | 3,708 ns/op (1,600 B, 100 allocs) | **5.1× list** |
| 1,000 | 1,256,488 ns/op (4,290,117 B, 2,000 allocs) | 43,287 ns/op (16,000 B, 1,000 allocs) | **29× list** |
| 10,000 | 81,232,695 ns/op (428,682,913 B, 20,005 allocs) | 374,521 ns/op (160,000 B, 10,000 allocs) | **217× list** |
| 100,000 | 10,463,990,507 ns/op (40,399,207,808 B, 201,354 allocs) | 4,696,411 ns/op (1,600,000 B, 100,000 allocs) | **2,228× list** |

**Linked list's strongest win.** Slice's `append([]int{v}, s...)` copies the entire array per insert — O(n²) time and allocation. Linked list's O(1) head pointer swap dominates. At small n (10), the slice is competitive because allocation overhead per node dominates.

### Build and Drain from Front

Measures full cycle: build a structure of size n, then delete all elements from the front. Pure delete-from-front is trivially fast for both (O(1) pointer bump), so the meaningful comparison is the whole build + drain workload (like a queue).

| n | Slice | Linked List | Ratio |
|---|---|---|---|
| 10 | 48 ns/op (80 B, 1 alloc) | 454 ns/op (160 B, 10 allocs) | **9.4× slice** |
| 100 | 290 ns/op (896 B, 1 alloc) | 4,207 ns/op (1,600 B, 100 allocs) | **14.5× slice** |
| 1,000 | 2,538 ns/op (8,192 B, 1 alloc) | 56,447 ns/op (16,000 B, 1,000 allocs) | **22× slice** |
| 10,000 | 27,287 ns/op (81,920 B, 1 alloc) | 481,916 ns/op (160,000 B, 10,000 allocs) | **17.7× slice** |
| 100,000 | 252,355 ns/op (802,827 B, 1 alloc) | 5,595,654 ns/op (1,600,001 B, 100,000 allocs) | **22× slice** |

Go's slice re-slicing (`s = s[1:]`) is O(1) — just moves the start pointer. No data movement. The linked list also does O(1) pointer swaps, but **building** the list costs n allocations vs the slice's 1, which dominates wall time.

### Random Insert

Insert n elements at random positions into a pre-built structure of size n. Both are O(n²) — slice shifts via memmove, linked list pointer-chases to find each insertion point. The RNG source is seeded once and reset between iterations (not re-created).

| n | Slice | Linked List | Ratio |
|---|---|---|---|
| 10 | 355 ns/op (240 B, 2 allocs) | 926 ns/op (320 B, 20 allocs) | **2.6× slice** |
| 100 | 4,130 ns/op (2,688 B, 2 allocs) | 22,616 ns/op (3,200 B, 200 allocs) | **5.5× slice** |
| 1,000 | 141,876 ns/op (38,912 B, 3 allocs) | 1,452,367 ns/op (32,000 B, 2,000 allocs) | **10× slice** |
| 10,000 | 11,346,520 ns/op (507,904 B, 4 allocs) | 170,959,331 ns/op (320,002 B, 20,000 allocs) | **15× slice** |
| 100,000 | 3,845,226,964 ns/op (6,635,536 B, 6 allocs) | 34,389,724,336 ns/op (3,200,016 B, 200,001 allocs) | **8.9× slice** |

**The slice is 3–15× faster** because:
- **memmove is cache-friendly linear copy** (prefetcher pulls in cache lines ahead)
- **Linked list traversal is pointer chasing** — random-address heap dereferences miss L1/L2/L3 and hit DRAM (~60–100 ns per miss)
- **Allocation count:** slice does O(log n) capacity-doubling reallocs vs linked list's O(n) per-node allocs

The hardware prefetcher makes memmove look cheap even though both are theoretically O(n²).

### Prepend-Build (build by repeatedly inserting at front)

| n | Slice | Linked List | Ratio |
|---|---|---|---|
| 10 | 541 ns/op (536 B, 19 allocs) | 418 ns/op (160 B, 10 allocs) | **1.3× list** |
| 100 | 16,180 ns/op (43,016 B, 199 allocs) | 4,292 ns/op (1,600 B, 100 allocs) | **3.8× list** |
| 1,000 | 1,222,705 ns/op (4,281,929 B, 1,999 allocs) | 48,161 ns/op (16,000 B, 1,000 allocs) | **25× list** |
| 10,000 | 106,142,524 ns/op (428,601,267 B, 20,006 allocs) | 367,844 ns/op (160,000 B, 10,000 allocs) | **288× list** |
| 100,000 | 11,890,264,879 ns/op (40,398,350,664 B, 200,903 allocs) | 6,521,933 ns/op (1,600,000 B, 100,000 allocs) | **1,823× list** |

Same as Insert Front but builds from scratch. Slice's O(n²) behavior produces absurd allocation pressure — **40 GB** of temporary allocations for building 100k elements. Linked list's O(n) prepend finishes in 6.5 ms.

## Summary

| Benchmark | Small n (10) | Large n (100k) | Winner |
|---|---|---|---|
| Append (tail) | 7.8× slice | 26.5× slice | **Slice** |
| Sequential iterate | 1.4× slice | 3.2× slice | **Slice** |
| Random access | 6.3× slice | **91,588×** slice | **Slice** |
| Random insert | 2.6× slice | 8.9× slice | **Slice** |
| Build and drain | 9.4× slice | 22× slice | **Slice** |
| Insert front | 1.6× list | 2,228× list | **Linked list** |
| Prepend-build | 1.3× list | 1,823× list | **Linked list** |

**Arrays win 5/7 benchmarks. Linked list wins 2/7 (front-insertion variants).**

## Verdict on Hypothesis

**Confirmed.** Arrays dominate whenever the workload touches memory sequentially or by index:

- **Sequential access** exploits hardware prefetchers that fill cache lines before the CPU requests them.
- **Indexed access** is O(1) pointer arithmetic vs O(n) heap-dereference walks.
- **Single-allocation growth** avoids per-node malloc/free overhead and GC scanning cost.

Linked lists are only competitive for pure front-insertion workloads (build by prepending, stack-like usage). In all other common patterns — iteration, random access, tail append, front delete — the array wins, often by orders of magnitude.

## Methodology Notes

- **RNG overhead eliminated:** Random benchmarks create the RNG source once and `Seed()` it between iterations, avoiding per-iteration allocation.
- **Random access uses shuffled indices** (not descending order) to avoid worst-case bias.
- **Build and Drain** measures the full build + drain cycle, since pure delete-from-front is O(1) for both and near-zero time.
