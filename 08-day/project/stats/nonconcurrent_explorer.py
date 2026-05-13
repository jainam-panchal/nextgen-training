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
    return data


def main() -> None:
    st.set_page_config(page_title="Swiss vs No-Swiss HashMap", layout="wide")
    st.title("Swiss vs No-Swiss HashMap (Non-Concurrent)")

    with st.sidebar:
        swiss_path = st.text_input("Swiss file", value="stats/swiss_nonconcurrent.txt")
        noswiss_path = st.text_input("No-Swiss file", value="stats/noswiss_nonconcurrent.txt")

    long_data = load_data(swiss_path, noswiss_path)
    if long_data.empty:
        st.error("No data parsed. Check benchmark output file paths.")
        return

    with st.sidebar:
        operations = sorted(long_data["operation"].dropna().unique().tolist())
        scales = sorted(int(x) for x in long_data["n"].dropna().unique())
        key_sizes = sorted(int(x) for x in long_data["keysize"].dropna().unique())
        default_op = "Read" if "Read" in operations else operations[0]

        selected_op = st.selectbox("Operation", operations, index=operations.index(default_op))
        selected_scales = st.multiselect("Scale (n)", scales, default=scales)
        selected_key_sizes = st.multiselect("Key Size", key_sizes, default=key_sizes)
        metric = st.selectbox("Metric", ["ns_op", "b_op", "allocs_op"], index=0)

    flong = long_data[
        (long_data["operation"] == selected_op)
        & long_data["n"].isin(selected_scales)
        & long_data["keysize"].isin(selected_key_sizes)
    ].copy()
    if flong.empty:
        st.warning("No rows for selected filters.")
        return

    st.caption(
        "Rows by operation: "
        + ", ".join(f"{k}={v}" for k, v in flong.groupby("operation").size().to_dict().items())
    )

    comp = (
        flong.groupby(["engine", "operation", "impl", "n"], dropna=False)[metric]
        .median()
        .reset_index()
    )

    st.subheader("Direct Comparison by Scale")
    fig_scale = px.line(
        comp,
        x="n",
        y=metric,
        color="engine",
        line_dash="impl",
        symbol="impl",
        facet_row="operation",
        log_x=True,
        markers=True,
        title=f"{metric} by n",
    )
    fig_scale.update_layout(height=700)
    st.plotly_chart(fig_scale, use_container_width=True)

    st.subheader("Summary Table")
    summary = flong.groupby(["operation", "n", "keysize", "impl", "engine"], dropna=False)[metric].median().reset_index()
    st.dataframe(
        summary.sort_values(["operation", "n", "keysize", "impl", "engine"]),
        use_container_width=True,
        hide_index=True,
    )


if __name__ == "__main__":
    main()
