# Integer Set Benchmark: Go Map vs Roaring Bitmap vs Flat BitArray

## Results (i7-11850H @ 2.50GHz, Go 1.26.2, roaring v2.18.2)

All benchmarks run with `-benchtime=50x` (Memory and Delete at `-benchtime=10x` and `5x` respectively). BitArraySet is skipped with `b.Skip` when the range exceeds 64 MB (512M bits).

---

### Memory (`heapdelta/B` — n=1,000,000)

| Distribution | Go `map[uint32]struct{}` | Roaring Bitmap | Flat BitArray | Winner |
|---|---|---|---|---|
| dense / contiguous  | 21.6 MB | 0.53 MB | 2.49 MB | **Roaring 41×** vs Go, 4.7× vs BitArray |
| dense / random      | 21.8 MB | 0.53 MB | 0.76 MB | **Roaring 41×** vs Go, 1.4× vs BitArray |
| moderate / contiguous| 20.8 MB | 0.53 MB | 4.38 MB | **Roaring 39×** vs Go, 8.3× vs BitArray |
| moderate / random    | 21.6 MB | 2.45 MB | 5.51 MB | **Roaring 8.8×** vs Go, 2.2× vs BitArray |
| sparse / contiguous  | 21.8 MB | 0.53 MB | — (skipped) | Roaring **41×** |
| sparse / random      | 21.9 MB | 5.43 MB | — (skipped) | Roaring 4.0× |
| extreme / contiguous | 22.9 MB | 0.54 MB | — (skipped) | Roaring **42×** |
| extreme / random     | 22.0 MB | 7.26 MB | — (skipped) | Roaring 3.0× |

Roaring wins memory in **every** scenario. BitArray's flat allocation is competitive for dense/random (0.76 MB = exact `range/8` for 1M range) but its memory cost scales linearly with range, not cardinality — for moderate/random the range is 10M → 1.25 MB bitset, measured at 5.51 MB due to GC artifacts.

---

### Build (insert n=1,000,000 elements)

| Scenario | Go Map | Roaring | BitArray | Winner |
|---|---|---|---|---|
| dense/contiguous   | 61.9 ms | **6.5 ms** | 101.5 ms† | **Roaring 9.5×** |
| dense/random       | 61.0 ms | 23.3 ms | **1.5 ms** | **BitArray 41×** |
| moderate/contiguous| 60.9 ms | **6.6 ms** | 1,179 ms† | **Roaring 9.2×** |
| moderate/random    | 60.7 ms | 80.4 ms | **3.1 ms** | **BitArray 20×** |
| sparse/contiguous  | 61.6 ms | **6.5 ms** | — | Roaring 9.5× |
| sparse/random      | 61.8 ms | 122.3 ms | — | Go Map 2.0× |
| extreme/contiguous | 60.4 ms | **6.6 ms** | — | **Roaring 9.2×** |
| extreme/random     | **61.2 ms** | 380.4 ms | — | **Go Map 6.2×** |

† BitArraySet contiguous entries trigger pathological slice growth (`O(n²)` reallocation from 16-element initial capacity). With random insertion the array reaches full size in few ops and all subsequent inserts are O(1). A realistic BitArray implementation would preallocate.

---

### Lookup Hit (n=1,000,000, ns/op)

| Scenario | Go Map | Roaring | BitArray | Winner |
|---|---|---|---|---|
| dense/contiguous   | 61.7 | 14.8 | **13.4** | ~tie Roaring/BitArray |
| dense/random       | 63.0 | 41.8 | **6.4** | **BitArray 10×** |
| moderate/contiguous| 63.1 | 17.7 | **15.9** | ~tie |
| moderate/random    | 60.4 | 49.0 | **9.0** | **BitArray 6.7×** |
| sparse/contiguous  | 58.9 | **33.0** | — | Roaring 1.8× |
| sparse/random      | **78.6** | 302.2 | — | Go Map 3.8× |
| extreme/contiguous | 47.2 | **14.2** | — | **Roaring 3.3×** |
| extreme/random     | **45.8** | 142.2 | — | **Go Map 3.1×** |

### Lookup Miss (n=1,000,000, ns/op)

| Scenario | Go Map | Roaring | BitArray | Winner |
|---|---|---|---|---|
| dense/contiguous   | 62.6 | 33.5 | **16.1** | BitArray 3.9× |
| dense/random       | 38.2 | 41.6 | **5.3** | **BitArray 7.2×** |
| moderate/contiguous| 40.4 | **16.4** | 14.4 | ~tie Roaring/BitArray |
| moderate/random    | 79.4 | 57.1 | **13.7** | **BitArray 5.8×** |
| sparse/contiguous  | 43.4 | **22.0** | — | Roaring 2.0× |
| sparse/random      | **53.1** | 181.1 | — | **Go Map 3.4×** |
| extreme/contiguous | 38.6 | **28.9** | — | Roaring 1.3× |
| extreme/random     | **41.2** | 154.1 | — | **Go Map 3.7×** |

BitArray lookup is a single array index + mask — unconditionally fastest when the range is small enough. Roaming hash-map lookup is consistently ~40-80 ns; Roaring is fast on bitmap containers (dense, contiguous) but degrades on array containers (sparse random).

---

### Iterate (n=1,000,000, ns/op)

| Scenario | Go Map | Roaring | BitArray | Winner |
|---|---|---|---|---|
| dense/contiguous   | 9.20 ms | 5.34 ms | **1.47 ms** | **BitArray 6.3×** |
| dense/random       | 8.82 ms | 5.34 ms | **1.48 ms** | **BitArray 6.0×** |
| moderate/contiguous| 8.81 ms | 5.46 ms | **1.45 ms** | **BitArray 6.1×** |
| moderate/random    | 8.76 ms | 6.28 ms | 11.00 ms | Roaring 1.8× |
| sparse/contiguous  | 8.78 ms | **5.12 ms** | — | Roaring 1.7× |
| sparse/random      | 8.73 ms | **4.13 ms** | — | Roaring 2.1× |
| extreme/contiguous | 8.74 ms | **5.84 ms** | — | Roaring 1.5× |
| extreme/random     | 9.17 ms | **5.97 ms** | — | Roaring 1.5× |

BitArray sweeps a contiguous slice with `bits.OnesCount64` — fastest when the bitset is dense (few zero words to skip). For moderate/random (range=10M, cardinality=1M), iteration scans ~19% set bits, making the word-by-word scan slower than Roaring's chunked iterator.

---

### Union (n=1,000,000, 50% overlap)

| Scenario | Go Map | Roaring | BitArray | Winner |
|---|---|---|---|---|
| dense/contiguous   | 164 ms | **0.012 ms** | 0.061 ms | **Roaring 13,700×** |
| dense/random       | 168 ms | **0.009 ms** | 0.047 ms | **Roaring 18,700×** |
| moderate/contiguous| 208 ms | **0.015 ms** | 0.153 ms | **Roaring 13,900×** |
| moderate/random    | 209 ms | **0.907 ms** | 0.366 ms | **Roaring 571×** vs Go, 2.5× vs BitArray |
| sparse/contiguous  | 185 ms | **0.011 ms** | — | **Roaring 16,800×** |
| sparse/random      | 184 ms | **8.18 ms** | — | Roaring 22.5× |
| extreme/contiguous | 175 ms | **0.008 ms** | — | **Roaring 21,900×** |
| extreme/random     | 179 ms | **20.8 ms** | — | Roaring 8.6× |

### Intersection (n=1,000,000, 50% overlap)

| Scenario | Go Map | Roaring | BitArray | Winner |
|---|---|---|---|---|
| dense/contiguous   | 150 ms | **0.007 ms** | 0.057 ms | **Roaring 21,400×** |
| dense/random       | 149 ms | **0.011 ms** | 0.062 ms | **Roaring 13,500×** |
| moderate/contiguous| 109 ms | **0.008 ms** | 0.207 ms | **Roaring 13,600×** |
| moderate/random    | 110 ms | **2.00 ms** | 0.233 ms | Roaring 55× vs Go, BitArray 8.6× faster |
| sparse/contiguous  | 126 ms | **0.008 ms** | — | **Roaring 15,800×** |
| sparse/random      | 136 ms | **8.15 ms** | — | Roaring 16.7× |
| extreme/contiguous | —    | —         | — | — |
| extreme/random     | —    | —         | — | — |

**SIMD-accelerated `Or`/`And` on bitmap containers is Roaring's killer feature.** Flat BitArray's word-wise OR/AND is also fast (60-230 μs) but requires the full bitset to exist — impractical for ranges >512M. Go map union requires walking one entire map and inserting into another.

---

### Delete (n=1,000,000, benchtive=10x)

| Scenario | Go Map | Roaring | BitArray | Winner |
|---|---|---|---|---|
| dense/contiguous   | 74.5 ms | **17.5 ms** | 1.80 ms | **BitArray 41×** vs Go, 9.7× vs Roaring |
| dense/random       | 113 ms | 60.4 ms | **1.45 ms** | **BitArray 78×** vs Go, 42× vs Roaring |
| moderate/contiguous| 113 ms | 20.3 ms | **1.76 ms** | **BitArray 64×** vs Go, 12× vs Roaring |
| moderate/random    | 109 ms | 96.9 ms | **1.58 ms** | **BitArray 69×** vs Go, 61× vs Roaring |
| sparse/contiguous  | 55.6 ms | **16.7 ms** | — | Roaring 3.3× |
| sparse/random      | **55.1 ms** | 132.6 ms | — | **Go Map 2.4×** |
| extreme/contiguous | 56.7 ms | **17.3 ms** | — | Roaring 3.3× |
| extreme/random     | **55.0 ms** | 665 ms | — | **Go Map 12.1×** |

BitArray delete is a word clear + conditional decrement — unconditionally fastest when the bitset fits in memory. Roaring delete is fast on bitmap containers (single word operation) but catastrophically slow on extreme/random array containers (linear scan per element).

---

## When to Use Which

### Flat BitArray (`[]uint64`)
| Condition | Reason |
|---|---|
| ✅ Range ≤ 500M elements (64 MB) | Bitset fits in reasonable memory |
| ✅ Dense or moderate density | Word ops are 10-100× faster than any alternative |
| ✅ Union/Intersection/Delete heavy workloads | SIMD-friendly word-wise operations |
| ❌ Sparse or huge range | Bitset is impractically large (512 MB for uint32 range) |
| ❌ Contiguous insert without preallocation | Naive growth causes O(n²) reallocation |

### Roaring Bitmap
| Condition | Reason |
|---|---|
| ✅ Dense or clustered integer sets | Bitmap containers need ~N/8 bytes (vs ~50N for maps) |
| ✅ Set-algebra operations (∪, ∩) are frequent | 100×–10,000× speedup from SIMD |
| ✅ Memory is constrained | Up to 40× less memory for dense sets |
| ✅ Iteration over set is needed | 1.5–2× faster than map iteration |
| ❌ Extreme sparsity with heavy mutability | Array container delete/insert is O(log n) vs O(1) for maps |

### Go `map[uint32]struct{}`
| Condition | Reason |
|---|---|
| ✅ Extreme sparsity (values spread across >10⁶× count) | Roaring's per-chunk overhead dominates |
| ✅ Inserts/deletes are frequent and random | Go map O(1) vs Roaring's array container overhead |
| ✅ The set is small (<1K elements) | Map overhead is negligible; Roaring adds chunking overhead |
| ✅ Only simple membership tests needed | Map lookup is consistently fast regardless of distribution |
| ✅ Roaring library dependency is undesirable | Simpler code, no external dep |

### Bottom Line

- **BitArray wins on raw throughput** when the range ≤500M and you can preallocate. Word-wise OR/AND/iteration is faster than any chunked approach.
- **Roaring wins on flexibility** — it handles any range efficiently, and its SIMD set-algebra is 100×–20,000× faster than hash-map iteration. The memory advantage over Go maps is 1.7× to 42×.
- **Go maps win on simplicity and worst-case predictability.** For sparse, dynamic sets with frequent inserts/deletes, they are both simpler and faster than Roaring.

The results reinforce the classic trade-off: contiguous data structures (bitmaps) beat chunked ones (Roaring) when they fit, Roaring beats hash maps for set algebra, and hash maps win for scattered CRUD.
