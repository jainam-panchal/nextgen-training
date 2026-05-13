# Non-Concurrent HashMap Benchmarks

## Python venv setup

From this directory:

```bash
cd /home/jainam-panchal/repos/nextgen-training/06,07-day/snippets/code_1
python3 -m venv .venv
.venv/bin/pip install -r requirements.txt
```

If `python3 -m venv` fails on your system, use:

```bash
pip3 install --user --break-system-packages virtualenv
~/.local/bin/virtualenv .venv
.venv/bin/pip install -r requirements.txt
```

## Run non-concurrent benchmarks

Swiss:

```bash
go test -run='^$' -bench='BenchmarkNonConcurrent(Build|Read|WriteUpdate|Delete|DeleteOnly)$' -benchmem -timeout=60m ./... > stats/swiss_nonconcurrent.txt
```

No-Swiss:

```bash
~/Downloads/go1.23.0.linux-amd64/go/bin/go test -run='^$' -bench='BenchmarkNonConcurrent(Build|Read|WriteUpdate|Delete|DeleteOnly)$' -benchmem -timeout=60m ./... > stats/noswiss_nonconcurrent.txt
```

Comparison:

```bash
benchstat stats/swiss_nonconcurrent.txt stats/noswiss_nonconcurrent.txt > stats/comparison_nonconcurrent.txt
```

## Run interactive explorer

```bash
.venv/bin/streamlit run stats/nonconcurrent_explorer.py
```

