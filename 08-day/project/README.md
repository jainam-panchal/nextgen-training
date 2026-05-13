# Minimal Swiss vs No-Swiss HashMap Comparison

Pure non-concurrent hashmap comparison (no locks, no parallel benchmark loops).

## Scope

- Implementations:
  - `BuiltinMap`
  - `ShardedBuiltinMap`
- Operations:
  - `Build`
  - `Read`
  - `WriteUpdate`
  - `DeleteOnly`
- Matrix (timely):
  - `n`: `1000`, `10000`, `100000`
  - `keysize`: `16`, `64`, `256`
  - `shards`: `1`, `16`, `64` (sharded only)

## Run Benchmarks

Swiss:

```bash
go test -run='^$' -bench='BenchmarkNonConcurrent(Build|Read|WriteUpdate|DeleteOnly)$' -benchmem -count=1 -timeout=30m ./... > stats/swiss_nonconcurrent.txt
```

No-Swiss:

```bash
~/Downloads/go1.23.0.linux-amd64/go/bin/go test -run='^$' -bench='BenchmarkNonConcurrent(Build|Read|WriteUpdate|DeleteOnly)$' -benchmem -count=1 -timeout=30m ./... > stats/noswiss_nonconcurrent.txt
```

Comparison:

```bash
benchstat stats/swiss_nonconcurrent.txt stats/noswiss_nonconcurrent.txt > stats/comparison_nonconcurrent.txt
```

## Visualization (Streamlit)

Create venv + install:

```bash
python3 -m venv .venv
.venv/bin/pip install -r requirements.txt
```

Run:

```bash
.venv/bin/streamlit run stats/nonconcurrent_explorer.py
```

