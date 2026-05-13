#!/usr/bin/env python3
"""Build a simple comparison dashboard for non-concurrent map benchmarks."""

from __future__ import annotations

import argparse
import re
from pathlib import Path

import pandas as pd
import plotly.express as px
from plotly.io import to_html


BENCH_RE = re.compile(
    r"^Benchmark(?P<name>\S+)-\d+\s+"
    r"(?P<iters>\d+)\s+"
    r"(?P<ns>[0-9]+(?:\.[0-9]+)?)\s+ns/op\s+"
    r"(?P<bop>[0-9]+(?:\.[0-9]+)?)\s+B/op\s+"
    r"(?P<allocs>[0-9]+(?:\.[0-9]+)?)\s+allocs/op$"
)
KV_RE = re.compile(r"([a-zA-Z_]+)=([0-9]+)")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--swiss", default="stats/swiss_nonconcurrent.txt", help="swiss benchmark output")
    parser.add_argument("--noswiss", default="stats/noswiss_nonconcurrent.txt", help="no-swiss benchmark output")
    parser.add_argument("--outdir", default="stats/plots", help="output directory")
    return parser.parse_args()


def parse_impl(name: str) -> str:
    if "/BuiltinMap/" in name:
        return "BuiltinMap"
    if "/ShardedBuiltinMap/" in name:
        return "ShardedBuiltinMap"
    return "Unknown"


def parse_operation(name: str) -> str:
    for op in ("Build", "Read", "WriteUpdate", "DeleteOnly", "Delete"):
        if name.startswith(f"NonConcurrent{op}/"):
            return op
    return "Unknown"


def parse_file(path: Path, engine: str) -> pd.DataFrame:
    rows: list[dict[str, object]] = []
    if not path.exists():
        return pd.DataFrame()

    for line in path.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        m = BENCH_RE.match(line)
        if not m:
            continue
        name = m.group("name")
        n = keysize = shards = None
        for key, value in KV_RE.findall(name):
            if key == "n":
                n = int(value)
            elif key == "keysize":
                keysize = int(value)
            elif key == "shards":
                shards = int(value)

        impl = parse_impl(name)
        if impl == "BuiltinMap" and shards is None:
            shards = 0

        rows.append(
            {
                "engine": engine,
                "benchmark": name,
                "operation": parse_operation(name),
                "impl": impl,
                "n": n,
                "keysize": keysize,
                "shards": shards,
                "ns_op": float(m.group("ns")),
                "b_op": float(m.group("bop")),
                "allocs_op": float(m.group("allocs")),
            }
        )
    return pd.DataFrame(rows)


def main() -> None:
    args = parse_args()
    outdir = Path(args.outdir)
    outdir.mkdir(parents=True, exist_ok=True)

    swiss = parse_file(Path(args.swiss), "swiss")
    noswiss = parse_file(Path(args.noswiss), "noswiss")
    data = pd.concat([swiss, noswiss], ignore_index=True)
    if data.empty:
        raise SystemExit("No benchmark rows parsed from input files.")

    grouped = (
        data.groupby(["engine", "benchmark", "operation", "impl", "n", "keysize", "shards"], dropna=False)[
            ["ns_op", "b_op", "allocs_op"]
        ]
        .mean()
        .reset_index()
    )
    wide = grouped.pivot_table(
        index=["benchmark", "operation", "impl", "n", "keysize", "shards"],
        columns="engine",
        values=["ns_op", "b_op", "allocs_op"],
        aggfunc="mean",
    ).reset_index()
    normalized_cols: list[str] = []
    for col in wide.columns:
        if isinstance(col, tuple):
            left = str(col[0]) if col[0] else ""
            right = str(col[1]) if len(col) > 1 and col[1] else ""
            if left and right:
                normalized_cols.append(f"{right}_{left}")
            elif left:
                normalized_cols.append(left)
            else:
                normalized_cols.append(right)
        else:
            normalized_cols.append(str(col))
    wide.columns = normalized_cols

    # Delta: positive => noswiss slower.
    wide["delta_ns_pct"] = ((wide["noswiss_ns_op"] - wide["swiss_ns_op"]) / wide["swiss_ns_op"]) * 100.0
    wide["delta_b_pct"] = ((wide["noswiss_b_op"] - wide["swiss_b_op"]) / wide["swiss_b_op"].replace(0, pd.NA)) * 100.0
    wide["delta_allocs_pct"] = (
        (wide["noswiss_allocs_op"] - wide["swiss_allocs_op"]) / wide["swiss_allocs_op"].replace(0, pd.NA)
    ) * 100.0

    fig_winner = px.bar(
        wide.assign(winner=wide.apply(lambda r: "swiss" if r["swiss_ns_op"] < r["noswiss_ns_op"] else "noswiss", axis=1))
        .groupby(["operation", "impl", "winner"], dropna=False)
        .size()
        .reset_index(name="count"),
        x="operation",
        y="count",
        color="winner",
        facet_col="impl",
        barmode="stack",
        title="Winner Count by Operation (Lower ns/op Wins)",
    )
    fig_winner.update_layout(height=460)

    fig_delta = px.line(
        wide.groupby(["operation", "impl", "keysize"], dropna=False)["delta_ns_pct"].median().reset_index(),
        x="keysize",
        y="delta_ns_pct",
        color="impl",
        facet_row="operation",
        markers=True,
        title="Latency Delta % vs Key Size (noswiss vs swiss)",
    )
    fig_delta.update_layout(height=1000)

    fig_payload_ns = px.line(
        wide.groupby(["operation", "impl", "n"], dropna=False)[["swiss_ns_op", "noswiss_ns_op"]]
        .median()
        .reset_index()
        .melt(id_vars=["operation", "impl", "n"], var_name="engine_metric", value_name="ns_op"),
        x="n",
        y="ns_op",
        color="impl",
        line_dash="engine_metric",
        facet_row="operation",
        log_x=True,
        markers=True,
        title="Latency by Payload Size (n)",
    )
    fig_payload_ns.update_layout(height=1000)

    mem_series = (
        wide.groupby(["operation", "impl", "n"], dropna=False)[["swiss_b_op", "noswiss_b_op"]]
        .median()
        .reset_index()
        .melt(id_vars=["operation", "impl", "n"], var_name="engine_metric", value_name="b_op")
    )
    # Hide operations that are fully zero in B/op to avoid misleading flat plots.
    non_zero_ops = (
        mem_series.groupby("operation", dropna=False)["b_op"]
        .max()
        .reset_index()
        .query("b_op > 0")["operation"]
        .tolist()
    )
    mem_series = mem_series[mem_series["operation"].isin(non_zero_ops)]
    fig_payload_mem = px.line(
        mem_series,
        x="n",
        y="b_op",
        color="impl",
        line_dash="engine_metric",
        facet_row="operation",
        log_x=True,
        markers=True,
        title="Memory by Payload Size (B/op, non-zero operations only)",
    )
    fig_payload_mem.update_layout(height=1000)

    csv_path = outdir / "nonconcurrent_parsed.csv"
    wide.to_csv(csv_path, index=False)

    html_path = outdir / "nonconcurrent_dashboard.html"
    html_parts = [
        "<html><head><meta charset='utf-8'><title>Non-Concurrent HashMap Dashboard</title></head><body>",
        "<h1>Non-Concurrent HashMap Comparison Dashboard</h1>",
        (
            "<p><b>Interpretation:</b> Positive delta means <code>noswiss</code> slower; "
            "negative delta means <code>noswiss</code> faster.</p>"
        ),
        "<p>Simple comparison only: operation, payload size, key size, implementation.</p>",
        "<h2>Winner Distribution</h2>",
        to_html(fig_winner, include_plotlyjs="cdn", full_html=False),
        "<h2>Delta by Key Size</h2>",
        to_html(fig_delta, include_plotlyjs=False, full_html=False),
        "<h2>Latency by Payload Size</h2>",
        to_html(fig_payload_ns, include_plotlyjs=False, full_html=False),
        "<h2>Memory by Payload Size</h2>",
        to_html(fig_payload_mem, include_plotlyjs=False, full_html=False),
        "</body></html>",
    ]
    html_path.write_text("\n".join(html_parts), encoding="utf-8")

    print(f"Wrote CSV: {csv_path}")
    print(f"Wrote HTML: {html_path}")
    print(f"Rows parsed: swiss={len(swiss)} noswiss={len(noswiss)} merged={len(wide)}")


if __name__ == "__main__":
    main()
