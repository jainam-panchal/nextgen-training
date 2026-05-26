# Roaring Bitmap vs Go Map — Integer Set Benchmark

## Results (i7-11850H @ 2.50GHz, Go 1.26.2, roaring v2.18.2)

### Memory (`heapdelta/B` — n=1,000,000)

| Distribution | Go `map[uint32]struct{}` | Roaring Bitmap | Ratio |
|---|---|---|---|
| dense / contiguous      | 21.8 MB | 0.53 MB | **41×** |
| dense / random          | 22.1 MB | 0.53 MB | **42×** |
| moderate / contiguous   | 21.8 MB | 0.53 MB | **41×** |
| moderate / random       | 21.6 MB | 5.08 MB | 4.3× |
| sparse / contiguous     | 21.8 MB | 0.53 MB | **41×** |
| sparse / random         | 21.7 MB | 7.46 MB | 2.9× |
| extreme / contiguous    | 21.8 MB | 0.54 MB | **40×** |
| extreme / random        | 21.8 MB | 13.1 MB | 1.7× |

Roaring wins memory in **every** scenario, but the margin collapses as sparsity increases: from 41× (dense/contiguous) down to 1.7× (extreme/random at 1M elements spread across the full uint32 range).

---

### Build (insert N=1,000,000 elements)

| Scenario | Go Map | Roaring | Winner |
|---|---|---|---|
| dense/contiguous   | 94.4 ms ± 10% | **9.8 ms ± 8%** | **Roaring 9.6× faster** |
| dense/random       | 87.7 ms ± 5%  | **32.8 ms ± 5%** | **Roaring 2.7× faster** |
| moderate/random    | 89.3 ms ± 5%  | 120.7 ms ± 10% | Go Map 1.4× faster |
| sparse/random      | 85.9 ms ± 4%  | 219.3 ms ± 8%  | **Go Map 2.6× faster** |
| extreme/random     | 88.7 ms ± 4%  | 572.4 ms ± 12% | **Go Map 6.5× faster** |

Roaring inserts fast when the bitmap stays dense (few chunks, bitmap containers). With random sparse data, each insertion opens a new array container and may trigger container-type transitions, cratering throughput.

---

### Lookup Hit (n=1,000,000, ns/op)

| Scenario | Go Map | Roaring | Winner |
|---|---|---|---|
| dense/contiguous   | 45.2 ns | **14.1 ns** | **Roaring 3.2× faster** |
| dense/random       | 47.4 ns | **26.4 ns** | **Roaring 1.8× faster** |
| moderate/contiguous| 49.6 ns | **14.0 ns** | **Roaring 3.5× faster** |
| moderate/random    | 45.6 ns | 46.2 ns | ~tie |
| sparse/random      | 54.0 ns | 162.7 ns  | **Go Map 3.0× faster** |
| extreme/random     | 48.3 ns | 267.9 ns  | **Go Map 5.5× faster** |

Roaring lookup is O(1) in bitmap containers (just a bitset test) but O(log n) in array containers (binary search) and can degrade with many sparse array containers.

### Lookup Miss (n=1,000,000, ns/op)

| Scenario | Go Map | Roaring | Winner |
|---|---|---|---|
| dense/contiguous   | 38.2 ns | **29.0 ns** | Roaring 1.3× |
| extreme/contiguous | 132.7 ns | **22.2 ns** | **Roaring 6.0×** |
| extreme/random     | 56.4 ns | 191.1 ns  | Go Map 3.4× |

---

### Iterate (n=1,000,000, ns/op)

| Scenario | Go Map | Roaring | Winner |
|---|---|---|---|
| dense/contiguous   | 15.7 ms | **9.6 ms** | Roaring 1.6× |
| dense/random       | 14.5 ms | **8.8 ms** | Roaring 1.6× |
| moderate/random    | 15.7 ms | **10.5 ms** | Roaring 1.5× |
| sparse/random      | 15.3 ms | **7.7 ms** | **Roaring 2.0×** |
| extreme/random     | 14.8 ms | **8.6 ms** | Roaring 1.7× |

Roaring's iterator is consistently faster — the internal container walk+scan is more cache-friendly than Go's hash-map iteration.

---

### Union (n=1,000,000, 50% overlap)

| Scenario | Go Map | Roaring | Winner |
|---|---|---|---|
| dense/contiguous   | 204.5 ms | **0.011 ms** | **Roaring 18,600×** |
| dense/random       | 204.1 ms | **0.013 ms** | **Roaring 15,800×** |
| moderate/contiguous| 225.0 ms | **0.029 ms** | **Roaring 7,800×** |
| moderate/random    | 277.1 ms | **1.64 ms** | **Roaring 170×** |
| sparse/random      | 217.7 ms | **12.2 ms** | **Roaring 18×** |
| extreme/random     | 249.1 ms | **42.1 ms** | **Roaring 5.9×** |

This is Roaring's killer feature. SIMD-accelerated `Or` on bitmap containers is orders of magnitude faster than hash-walking + insertion into a new map. The advantage narrows with sparsity as more array containers force element-by-element merge.

### Intersection (n=1,000,000, 50% overlap)

| Scenario | Go Map | Roaring | Winner |
|---|---|---|---|
| dense/contiguous   | 236.9 ms | **0.025 ms** | **Roaring 9,300×** |
| dense/random       | 204.1 ms | **0.014 ms** | **Roaring 14,200×** |
| moderate/contiguous| 131.3 ms | **0.011 ms** | **Roaring 11,800×** |
| moderate/random    | 130.9 ms | **3.05 ms** | **Roaring 43×** |
| sparse/random      | 130.1 ms | **10.4 ms** | **Roaring 12.5×** |
| extreme/random     | 141.5 ms | **26.2 ms** | **Roaring 5.4×** |

Same pattern as union. The `And` operation on bitmap containers uses SIMD popcount & AND. Even for sparse random, Roaring holds an advantage.

---

### Delete (n=1,000,000)

| Scenario | Go Map | Roaring | Winner |
|---|---|---|---|
| dense/contiguous   | 75.9 ms | **23.8 ms** | Roaring 3.2× |
| moderate/contiguous| 71.7 ms | **23.0 ms** | Roaring 3.1× |
| moderate/random    | 72.8 ms | 122.2 ms  | Go Map 1.7× |
| dense/random       | 71.3 ms | 86.9 ms   | Go Map 1.2× |
| sparse/random      | 70.4 ms | 192.0 ms  | **Go Map 2.7×** |
| extreme/random     | 74.4 ms | 1219 ms   | **Go Map 16.4×** |

Go map delete is O(1) amortized with zero extra work. Roaring's delete on array containers requires a linear scan. On bitmap containers it's a single word operation. For extreme sparsity, removing from thousands of sparse array containers is catastrophic.

---

## Summary: When to Use Which

### Roaring Bitmap is the right choice when:

| Condition | Reason |
|---|---|
| ✅ Dense or clustered integer sets | Bitmap containers need ~N/8 bytes (vs ~50N for maps) |
| ✅ Sets of 10K+ elements | Overhead amortizes; map memory is ~38 KB at N=1K and grows linearly |
| ✅ Set-algebra operations (∪, ∩, −) are frequent | 100×–10,000× speedup from SIMD |
| ✅ Memory is constrained | Up to 40× less memory for dense sets |
| ✅ Iteration over set is needed | 1.5–2× faster than map iteration |

### Go `map[uint32]struct{}` is the right choice when:

| Condition | Reason |
|---|---|
| ✅ Extreme sparsity (values spread across >10⁶× count) | Roaring's per-chunk overhead dominates |
| ✅ Inserts/deletes are frequent and random | Go map O(1) vs Roaring's O(log n) array operations |
| ✅ The set is small (<1K elements) | Map overhead is negligible; Roaring adds chunking overhead |
| ✅ Only simple membership tests needed | Map lookup is consistently fast regardless of distribution |
| ✅ Roaring library dependency is undesirable | Simpler code, no external dep |

### Bottom Line

- **Excellent fit:** Dense integer sets, especially with set operations. The `roaring.Or`/`And` speedup over manual map iteration is the single most compelling reason to adopt it.
- **Good fit:** Moderate density, read-heavy workloads, constrained memory.
- **Poor fit:** Extreme sparsity with heavy mutability (inserts/deletes). Go's map is simpler and faster for scattered, dynamic integer sets.
