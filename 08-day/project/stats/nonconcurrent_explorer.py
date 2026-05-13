#!/usr/bin/env python3
from __future__ import annotations

import re
from pathlib import Path

import pandas as pd
import plotly.express as px
import streamlit as st


BENCH_RE = re.compile(
    r"^Benchmark(?P<name>\S+)-\d+\s+"
    r"(?P<iters>\d+)\s+"
    r"(?P<ns>[0-9]+(?:\.[0-9]+)?)\s+ns/op\s+"
    r"(?P<bop>[0-9]+(?:\.[0-9]+)?)\s+B/op\s+"
    r"(?P<allocs>[0-9]+(?:\.[0-9]+)?)\s+allocs/op$"
)
KV_RE = re.compile(r"([a-zA-Z_]+)=([0-9]+)")


def parse_impl(name: str) -> str:
    if "/BuiltinMap/" in name:
        return "BuiltinMap"
    if "/ShardedBuiltinMap/" in name:
        return "ShardedBuiltinMap"
    return "Unknown"


def parse_operation(name: str) -> str:
    for op in ("Build", "Read", "WriteUpdate", "DeleteOnly"):
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
        impl = parse_impl(name)
        n = keysize = shards = None
        for key, value in KV_RE.findall(name):
            if key == "n":
                n = int(value)
            elif key == "keysize":
                keysize = int(value)
            elif key == "shards":
                shards = int(value)

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


@st.cache_data(show_spinner=False)
def load_data(swiss_path: str, noswiss_path: str) -> pd.DataFrame:
    swiss = parse_file(Path(swiss_path), "swiss")
    noswiss = parse_file(Path(noswiss_path), "noswiss")
    data = pd.concat([swiss, noswiss], ignore_index=True)
    if data.empty:
        return pd.DataFrame()

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

    cols = []
    for col in wide.columns:
        if isinstance(col, tuple):
            left = str(col[0]) if col[0] else ""
            right = str(col[1]) if len(col) > 1 and col[1] else ""
            cols.append(f"{right}_{left}" if left and right else left or right)
        else:
            cols.append(str(col))
    wide.columns = cols

    wide["delta_ns_pct"] = ((wide["noswiss_ns_op"] - wide["swiss_ns_op"]) / wide["swiss_ns_op"]) * 100.0
    wide["delta_b_pct"] = ((wide["noswiss_b_op"] - wide["swiss_b_op"]) / wide["swiss_b_op"].replace(0, pd.NA)) * 100.0
    wide["delta_allocs_pct"] = (
        (wide["noswiss_allocs_op"] - wide["swiss_allocs_op"]) / wide["swiss_allocs_op"].replace(0, pd.NA)
    ) * 100.0
    return wide


def main() -> None:
    st.set_page_config(page_title="Swiss vs No-Swiss HashMap", layout="wide")
    st.title("Swiss vs No-Swiss HashMap (Non-Concurrent)")

    with st.sidebar:
        swiss_path = st.text_input("Swiss file", value="stats/swiss_nonconcurrent.txt")
        noswiss_path = st.text_input("No-Swiss file", value="stats/noswiss_nonconcurrent.txt")

    wide = load_data(swiss_path, noswiss_path)
    if wide.empty:
        st.error("No data parsed. Check benchmark output file paths.")
        return

    with st.sidebar:
        operations = sorted(wide["operation"].dropna().unique().tolist())
        impls = sorted(wide["impl"].dropna().unique().tolist())
        scales = sorted(int(x) for x in wide["n"].dropna().unique())
        key_sizes = sorted(int(x) for x in wide["keysize"].dropna().unique())
        shards = sorted(int(x) for x in wide["shards"].dropna().unique())

        selected_ops = st.multiselect("Operation", operations, default=operations)
        selected_impls = st.multiselect("Implementation", impls, default=impls)
        selected_scales = st.multiselect("Scale (n)", scales, default=scales)
        selected_key_sizes = st.multiselect("Key Size", key_sizes, default=key_sizes)
        selected_shards = st.multiselect("Shards", shards, default=shards)
        metric = st.selectbox("Metric", ["ns_op", "b_op", "allocs_op"], index=0)

    fwide = wide[
        wide["operation"].isin(selected_ops)
        & wide["impl"].isin(selected_impls)
        & wide["n"].isin(selected_scales)
        & wide["keysize"].isin(selected_key_sizes)
        & wide["shards"].isin(selected_shards)
    ].copy()
    if fwide.empty:
        st.warning("No rows for selected filters.")
        return

    metric_cols = [f"swiss_{metric}", f"noswiss_{metric}"]
    comp = (
        fwide.groupby(["operation", "impl", "n"], dropna=False)[metric_cols]
        .median()
        .reset_index()
        .melt(
            id_vars=["operation", "impl", "n"],
            value_vars=metric_cols,
            var_name="engine_metric",
            value_name=metric,
        )
    )

    st.subheader("Metric vs Scale")
    fig_scale = px.line(
        comp,
        x="n",
        y=metric,
        color="impl",
        line_dash="engine_metric",
        facet_row="operation",
        log_x=True,
        markers=True,
        title=f"{metric} by n",
    )
    fig_scale.update_layout(height=900)
    st.plotly_chart(fig_scale, use_container_width=True)

    delta_col = {"ns_op": "delta_ns_pct", "b_op": "delta_b_pct", "allocs_op": "delta_allocs_pct"}[metric]
    st.subheader("Delta % by Key Size (No-Swiss vs Swiss)")
    fig_delta = px.line(
        fwide.groupby(["operation", "impl", "keysize"], dropna=False)[delta_col]
        .median()
        .reset_index(),
        x="keysize",
        y=delta_col,
        color="impl",
        facet_row="operation",
        markers=True,
        title=delta_col,
    )
    fig_delta.add_hline(y=0, line_dash="dash", line_color="gray")
    fig_delta.update_layout(height=900)
    st.plotly_chart(fig_delta, use_container_width=True)

    st.subheader("Filtered Data")
    st.dataframe(
        fwide.sort_values(["operation", "impl", "n", "keysize", "shards"]),
        use_container_width=True,
        hide_index=True,
    )


if __name__ == "__main__":
    main()

