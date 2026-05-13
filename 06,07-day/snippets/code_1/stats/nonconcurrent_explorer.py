#!/usr/bin/env python3
"""Streamlit explorer for non-concurrent hashmap benchmark comparisons."""

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
def load_data(swiss_path: str, noswiss_path: str) -> tuple[pd.DataFrame, pd.DataFrame]:
    swiss = parse_file(Path(swiss_path), "swiss")
    noswiss = parse_file(Path(noswiss_path), "noswiss")
    data = pd.concat([swiss, noswiss], ignore_index=True)
    if data.empty:
        return data, pd.DataFrame()

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

    cols: list[str] = []
    for col in wide.columns:
        if isinstance(col, tuple):
            left = str(col[0]) if col[0] else ""
            right = str(col[1]) if len(col) > 1 and col[1] else ""
            if left and right:
                cols.append(f"{right}_{left}")
            elif left:
                cols.append(left)
            else:
                cols.append(right)
        else:
            cols.append(str(col))
    wide.columns = cols

    wide["delta_ns_pct"] = ((wide["noswiss_ns_op"] - wide["swiss_ns_op"]) / wide["swiss_ns_op"]) * 100.0
    wide["delta_b_pct"] = ((wide["noswiss_b_op"] - wide["swiss_b_op"]) / wide["swiss_b_op"].replace(0, pd.NA)) * 100.0
    wide["delta_allocs_pct"] = (
        (wide["noswiss_allocs_op"] - wide["swiss_allocs_op"]) / wide["swiss_allocs_op"].replace(0, pd.NA)
    ) * 100.0

    return grouped, wide


def main() -> None:
    st.set_page_config(page_title="Non-Concurrent HashMap Explorer", layout="wide")
    st.title("Non-Concurrent HashMap Comparison")
    st.caption("Simple payload-based comparison across operation, key size, scale, shards, and engine.")

    with st.sidebar:
        st.header("Data")
        swiss_path = st.text_input("Swiss file", value="stats/swiss_nonconcurrent.txt")
        noswiss_path = st.text_input("No-Swiss file", value="stats/noswiss_nonconcurrent.txt")
        st.divider()

    long_df, wide_df = load_data(swiss_path, noswiss_path)
    if wide_df.empty:
        st.error("No benchmark rows parsed. Check file paths and format.")
        return

    with st.sidebar:
        st.header("Filters")
        operations = sorted(wide_df["operation"].dropna().unique().tolist())
        impls = sorted(wide_df["impl"].dropna().unique().tolist())
        scales = sorted(int(x) for x in wide_df["n"].dropna().unique())
        key_sizes = sorted(int(x) for x in wide_df["keysize"].dropna().unique())
        shards = sorted(int(x) for x in wide_df["shards"].dropna().unique())

        st.subheader("Preset")
        preset = st.selectbox(
            "Quick preset",
            options=["Read + Builtin (no shard)", "All data", "Sharded only"],
            index=0,
            key="preset",
        )

        if preset == "Read + Builtin (no shard)":
            preset_ops = ["Read"] if "Read" in operations else operations
            preset_impls = ["BuiltinMap"] if "BuiltinMap" in impls else impls
            preset_shards = [0] if 0 in shards else shards
        elif preset == "Sharded only":
            preset_ops = operations
            preset_impls = ["ShardedBuiltinMap"] if "ShardedBuiltinMap" in impls else impls
            preset_shards = [s for s in shards if s != 0] or shards
        else:
            preset_ops = operations
            preset_impls = impls
            preset_shards = shards

        if "last_preset" not in st.session_state:
            st.session_state["last_preset"] = preset

        # Auto-apply whenever preset changes so filters don't stay stuck.
        if st.session_state["last_preset"] != preset:
            st.session_state["selected_ops"] = preset_ops
            st.session_state["selected_impls"] = preset_impls
            st.session_state["selected_scales"] = scales
            st.session_state["selected_key_sizes"] = key_sizes
            st.session_state["selected_shards"] = preset_shards
            st.session_state["metric"] = "ns_op"
            st.session_state["view_mode"] = "aggregated"
            st.session_state["use_log_x"] = True
            st.session_state["last_preset"] = preset

        if st.button("Reset Filters to Preset"):
            st.session_state["selected_ops"] = preset_ops
            st.session_state["selected_impls"] = preset_impls
            st.session_state["selected_scales"] = scales
            st.session_state["selected_key_sizes"] = key_sizes
            st.session_state["selected_shards"] = preset_shards
            st.session_state["metric"] = "ns_op"
            st.session_state["view_mode"] = "aggregated"
            st.session_state["use_log_x"] = True

        if "selected_ops" not in st.session_state:
            st.session_state["selected_ops"] = preset_ops
        if "selected_impls" not in st.session_state:
            st.session_state["selected_impls"] = preset_impls
        if "selected_scales" not in st.session_state:
            st.session_state["selected_scales"] = scales
        if "selected_key_sizes" not in st.session_state:
            st.session_state["selected_key_sizes"] = key_sizes
        if "selected_shards" not in st.session_state:
            st.session_state["selected_shards"] = preset_shards
        if "metric" not in st.session_state:
            st.session_state["metric"] = "ns_op"
        if "view_mode" not in st.session_state:
            st.session_state["view_mode"] = "aggregated"
        if "use_log_x" not in st.session_state:
            st.session_state["use_log_x"] = True

        selected_ops = st.multiselect("Operation", options=operations, key="selected_ops")
        selected_impls = st.multiselect("Implementation", options=impls, key="selected_impls")
        selected_scales = st.multiselect("Scale (n)", options=scales, key="selected_scales")
        selected_key_sizes = st.multiselect("Key Size", options=key_sizes, key="selected_key_sizes")
        selected_shards = st.multiselect("Shards", options=shards, key="selected_shards")
        metric = st.selectbox("Metric", options=["ns_op", "b_op", "allocs_op"], key="metric")
        view_mode = st.selectbox("View", options=["aggregated", "raw"], key="view_mode")
        use_log_x = st.toggle("Log scale on n", key="use_log_x")

    fwide = wide_df[
        wide_df["operation"].isin(selected_ops)
        & wide_df["impl"].isin(selected_impls)
        & wide_df["n"].isin(selected_scales)
        & wide_df["keysize"].isin(selected_key_sizes)
        & wide_df["shards"].isin(selected_shards)
    ].copy()
    if fwide.empty:
        st.warning("No rows for selected filters.")
        return

    # Create long comparison view for selected metric.
    metric_cols = [f"swiss_{metric}", f"noswiss_{metric}"]
    if view_mode == "aggregated":
        # One clean line per (operation, impl, engine) by collapsing payload variants.
        comp_base = (
            fwide.groupby(["operation", "impl", "n"], dropna=False)[metric_cols]
            .median()
            .reset_index()
        )
    else:
        comp_base = fwide[["operation", "impl", "n", "keysize", "shards", *metric_cols]].copy()

    id_vars = ["operation", "impl", "n"]
    if "keysize" in comp_base.columns:
        id_vars.append("keysize")
    if "shards" in comp_base.columns:
        id_vars.append("shards")

    comp = comp_base.melt(
        id_vars=id_vars,
        value_vars=metric_cols,
        var_name="engine_metric",
        value_name=metric,
    )
    comp["engine"] = comp["engine_metric"].str.replace(f"_{metric}", "", regex=False)
    if "keysize" not in comp.columns:
        comp["keysize"] = 0
    if "shards" not in comp.columns:
        comp["shards"] = 0
    comp = comp.sort_values(["operation", "impl", "engine", "keysize", "shards", "n"])
    comp["series_id"] = (
        comp["impl"].astype(str)
        + "|"
        + comp["engine"].astype(str)
        + "|k="
        + comp["keysize"].astype(str)
        + "|s="
        + comp["shards"].astype(str)
    )

    st.subheader("Quick Summary")
    c1, c2, c3 = st.columns(3)
    with c1:
        st.metric("Selected Rows", f"{len(fwide)}")
    with c2:
        st.metric("Operations", f"{fwide['operation'].nunique()}")
    with c3:
        st.metric("Implementations", f"{fwide['impl'].nunique()}")

    st.info("Use preset + reset for quick workflows, then fine-tune filters manually.")

    st.subheader("Engine Comparison by Payload Size")
    fig_payload = px.line(
        comp,
        x="n",
        y=metric,
        color="impl",
        line_dash="engine",
        line_group="series_id",
        facet_row="operation",
        markers=True,
        log_x=use_log_x,
        hover_data=["keysize", "shards"],
        title=f"{metric} vs n ({view_mode})",
    )
    fig_payload.update_layout(height=900)
    fig_payload.update_yaxes(title=metric)
    fig_payload.update_xaxes(title="n (payload size)")
    st.plotly_chart(fig_payload, use_container_width=True)

    st.subheader("Delta % by Key Size (No-Swiss vs Swiss)")
    delta_col = {
        "ns_op": "delta_ns_pct",
        "b_op": "delta_b_pct",
        "allocs_op": "delta_allocs_pct",
    }[metric]
    delta = (
        fwide.groupby(["operation", "impl", "keysize"], dropna=False)[delta_col]
        .median()
        .reset_index()
        .sort_values(["operation", "impl", "keysize"])
    )
    fig_delta = px.line(
        delta,
        x="keysize",
        y=delta_col,
        color="impl",
        facet_row="operation",
        markers=True,
        title=f"{delta_col} by keysize",
    )
    fig_delta.add_hline(y=0, line_dash="dash", line_color="gray")
    fig_delta.update_layout(height=900)
    fig_delta.update_xaxes(title="keysize")
    fig_delta.update_yaxes(title=delta_col)
    st.plotly_chart(fig_delta, use_container_width=True)

    st.subheader("Data Table")
    st.dataframe(
        fwide.sort_values(["operation", "impl", "n", "keysize", "shards"]),
        use_container_width=True,
        hide_index=True,
    )


if __name__ == "__main__":
    main()
