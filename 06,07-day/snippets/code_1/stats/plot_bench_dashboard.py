#!/usr/bin/env python3
"""Generate an interactive Plotly dashboard from benchstat comparison output."""

from __future__ import annotations

import argparse
import re
from pathlib import Path

import pandas as pd
import plotly.express as px
import plotly.graph_objects as go
from plotly.io import to_html


ROW_RE = re.compile(
    r"^(?P<name>[A-Za-z0-9_/=\-]+)\s+"
    r"(?P<swiss>[0-9]+(?:\.[0-9]+)?)(?P<swiss_unit>[num]?)\s*±\s*[^ ]+.*?"
    r"(?P<noswiss>[0-9]+(?:\.[0-9]+)?)(?P<noswiss_unit>[num]?)\s*±\s*"
)
NAME_PART_RE = re.compile(r"([a-zA-Z_]+)=([0-9]+)")
RATIO_RE = re.compile(r"([0-9]{2})Read([0-9]{1,2})Write")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--input", default="stats/comparision.txt", help="benchstat comparison file")
    parser.add_argument("--outdir", default="stats/plots", help="output directory")
    parser.add_argument(
        "--family",
        default="all",
        choices=["all", "ReadHeavy", "WriteHeavy", "MixedRatios", "ShardSweep"],
        help="optional family filter",
    )
    parser.add_argument("--strict", action="store_true", help="fail if any row can't be parsed")
    return parser.parse_args()


def unit_to_ns(value: float, unit: str) -> float:
    if unit == "":
        return value * 1_000_000_000.0
    if unit == "m":
        return value * 1_000_000.0
    if unit == "u":
        return value * 1_000.0
    if unit == "n":
        return value
    return value


def infer_family(name: str) -> str:
    if name.startswith("ConcurrentMapReadHeavy/"):
        return "ReadHeavy"
    if name.startswith("ConcurrentMapWriteHeavy/"):
        return "WriteHeavy"
    if name.startswith("ConcurrentMapMixedRatios/"):
        return "MixedRatios"
    if name.startswith("LockedShardedMapShardSweep"):
        return "ShardSweep"
    return "Unknown"


def infer_impl(name: str) -> str:
    if "/LockedBuiltinMap/" in name or name.endswith("/LockedBuiltinMap"):
        return "LockedBuiltinMap"
    if "/LockedShardedBuiltinMap/" in name or "LockedShardedMapShardSweep" in name:
        return "LockedShardedBuiltinMap"
    if "/SyncMap/" in name or name.endswith("/SyncMap"):
        return "SyncMap"
    return "Unknown"


def parse_dimensions(name: str) -> tuple[int | None, int | None, int | None, str | None]:
    n = keysize = shards = None
    ratio_label = None

    for key, value in NAME_PART_RE.findall(name):
        if key == "n":
            n = int(value)
        elif key == "keysize":
            keysize = int(value)
        elif key == "shards":
            shards = int(value)

    ratio_match = RATIO_RE.search(name)
    if ratio_match:
        ratio_label = f"{ratio_match.group(1)}/{ratio_match.group(2)}"

    return n, keysize, shards, ratio_label


def parse_benchstat(input_path: Path, strict: bool) -> tuple[pd.DataFrame, int]:
    rows: list[dict[str, object]] = []
    skipped = 0

    for line_no, line in enumerate(input_path.read_text(encoding="utf-8").splitlines(), start=1):
        line = line.strip()
        if not line:
            continue
        if not (line.startswith("ConcurrentMap") or line.startswith("LockedShardedMap")):
            continue

        match = ROW_RE.match(line)
        if not match:
            skipped += 1
            if strict:
                raise ValueError(f"unparsed benchmark row at line {line_no}: {line}")
            continue

        name = match.group("name")
        swiss = unit_to_ns(float(match.group("swiss")), match.group("swiss_unit"))
        noswiss = unit_to_ns(float(match.group("noswiss")), match.group("noswiss_unit"))
        if swiss <= 0 or noswiss <= 0:
            skipped += 1
            continue

        family = infer_family(name)
        impl = infer_impl(name)
        n, keysize, shards, ratio_label = parse_dimensions(name)

        delta_pct = ((noswiss - swiss) / swiss) * 100.0
        speedup = noswiss / swiss
        winner = "swiss" if swiss < noswiss else "noswiss"

        rows.append(
            {
                "benchmark": name,
                "family": family,
                "impl": impl,
                "n": n,
                "keysize": keysize,
                "shards": shards,
                "ratio": ratio_label,
                "swiss_ns_op": swiss,
                "noswiss_ns_op": noswiss,
                "delta_pct": delta_pct,
                "speedup": speedup,
                "winner": winner,
            }
        )

    return pd.DataFrame(rows), skipped


def fig_winner_distribution(df: pd.DataFrame) -> go.Figure:
    if df.empty:
        return go.Figure()
    grouped = (
        df.groupby(["family", "impl", "winner"], dropna=False)
        .size()
        .reset_index(name="count")
        .sort_values(["family", "impl", "winner"])
    )
    fig = px.bar(
        grouped,
        x="family",
        y="count",
        color="winner",
        facet_col="impl",
        barmode="stack",
        title="Who Wins More Often (Lower ns/op Wins)",
        labels={"count": "Benchmark Cases"},
        color_discrete_map={"swiss": "#1f77b4", "noswiss": "#ff7f0e"},
    )
    fig.update_layout(height=460)
    return fig


def fig_median_delta_by_family(df: pd.DataFrame) -> go.Figure:
    if df.empty:
        return go.Figure()
    grouped = (
        df.groupby(["family", "impl"], dropna=False)["delta_pct"]
        .median()
        .reset_index()
        .sort_values(["family", "impl"])
    )
    fig = px.bar(
        grouped,
        x="family",
        y="delta_pct",
        color="impl",
        barmode="group",
        title="Median Delta % by Workload (noswiss vs swiss)",
        labels={"delta_pct": "Median Delta %"},
    )
    fig.add_hline(y=0, line_dash="dash", line_color="gray")
    fig.update_layout(height=420)
    return fig


def fig_growth_by_scale(df: pd.DataFrame) -> go.Figure:
    if df.empty:
        return go.Figure()
    data = df[df["n"].notna() & (df["family"] != "ShardSweep")].copy()
    if data.empty:
        return go.Figure()
    grouped = (
        data.groupby(["family", "impl", "n"], dropna=False)[["swiss_ns_op", "noswiss_ns_op"]]
        .mean()
        .reset_index()
        .sort_values(["family", "impl", "n"])
    )
    # Growth factor vs smallest n in each (family, impl)
    grouped["swiss_growth"] = grouped.groupby(["family", "impl"])["swiss_ns_op"].transform(lambda s: s / s.iloc[0])
    grouped["noswiss_growth"] = grouped.groupby(["family", "impl"])["noswiss_ns_op"].transform(lambda s: s / s.iloc[0])

    swiss = grouped[["family", "impl", "n", "swiss_growth"]].rename(columns={"swiss_growth": "growth"})
    swiss["engine"] = "swiss"
    noswiss = grouped[["family", "impl", "n", "noswiss_growth"]].rename(columns={"noswiss_growth": "growth"})
    noswiss["engine"] = "noswiss"
    melted = pd.concat([swiss, noswiss], ignore_index=True)

    fig = px.line(
        melted,
        x="n",
        y="growth",
        color="impl",
        line_dash="engine",
        facet_row="family",
        log_x=True,
        markers=True,
        title="Growth Factor by Scale (lower growth is better)",
        labels={"growth": "Growth Factor (vs smallest n)", "n": "Scale (n)"},
    )
    fig.update_layout(height=900)
    return fig


def fig_delta_by_keysize(df: pd.DataFrame) -> go.Figure:
    data = df[df["keysize"].notna() & (df["family"] != "ShardSweep")].copy()
    if data.empty:
        return go.Figure()
    grouped = (
        data.groupby(["family", "impl", "keysize"], dropna=False)["delta_pct"]
        .mean()
        .reset_index()
        .sort_values(["family", "impl", "keysize"])
    )
    fig = px.line(
        grouped,
        x="keysize",
        y="delta_pct",
        color="impl",
        facet_row="family",
        markers=True,
        title="Delta % by Key Size (noswiss vs swiss)",
        labels={"delta_pct": "Delta %", "keysize": "Key Size (bytes)", "impl": "Map"},
    )
    fig.update_layout(height=900)
    return fig


def fig_shard_sweep(df: pd.DataFrame) -> go.Figure:
    data = df[(df["family"] == "ShardSweep") & df["shards"].notna() & df["n"].notna()].copy()
    if data.empty:
        return go.Figure()
    grouped = (
        data.groupby(["shards", "n"], dropna=False)[["swiss_ns_op", "noswiss_ns_op"]]
        .mean()
        .reset_index()
        .sort_values(["n", "shards"])
    )
    fig = go.Figure()
    for n in sorted(grouped["n"].dropna().unique()):
        sample = grouped[grouped["n"] == n]
        fig.add_trace(
            go.Scatter(
                x=sample["shards"],
                y=sample["swiss_ns_op"],
                mode="lines+markers",
                name=f"swiss (n={n})",
            )
        )
        fig.add_trace(
            go.Scatter(
                x=sample["shards"],
                y=sample["noswiss_ns_op"],
                mode="lines+markers",
                name=f"noswiss (n={n})",
            )
        )
    fig.update_layout(
        title="Shard Sweep Comparison (90/10 Mixed)",
        xaxis_title="Shard Count",
        yaxis_title="ns/op",
        height=560,
    )
    return fig


def build_scale_recommendations(df: pd.DataFrame) -> pd.DataFrame:
    data = df[df["n"].notna()].copy()
    if data.empty:
        return pd.DataFrame(columns=["n", "best_impl", "median_ns"])
    grouped = data.groupby(["n", "impl"], dropna=False)[["swiss_ns_op", "noswiss_ns_op"]].median().reset_index()
    grouped["median_ns"] = grouped[["swiss_ns_op", "noswiss_ns_op"]].min(axis=1)
    idx = grouped.groupby("n")["median_ns"].idxmin()
    out = grouped.loc[idx, ["n", "impl", "median_ns"]].sort_values("n").reset_index(drop=True)
    out = out.rename(columns={"impl": "best_impl"})
    return out


def recommendation_html(rec_df: pd.DataFrame) -> str:
    if rec_df.empty:
        return "<p>No recommendation rows available.</p>"
    rows = [
        "<table border='1' cellpadding='6' cellspacing='0'>",
        "<tr><th>Scale (n)</th><th>Recommended Map</th><th>Median Best ns/op</th></tr>",
    ]
    for _, row in rec_df.iterrows():
        rows.append(f"<tr><td>{int(row['n'])}</td><td>{row['best_impl']}</td><td>{row['median_ns']:.2f}</td></tr>")
    rows.append("</table>")
    return "\n".join(rows)


def write_dashboard(out_path: Path, figures: list[tuple[str, go.Figure]], summary: dict[str, object]) -> None:
    parts = [
        "<html><head><meta charset='utf-8'><title>Benchmark Dashboard</title></head><body>",
        "<h1>Swiss vs No-Swiss Benchmark Dashboard</h1>",
        (
            f"<p><b>Rows:</b> {summary['rows']} | <b>Skipped:</b> {summary['skipped']} | "
            f"<b>Families:</b> {summary['families']} | <b>Impls:</b> {summary['impls']}</p>"
        ),
        (
            "<p><b>Interpretation:</b> Positive Delta % means <code>noswiss</code> is slower than "
            "<code>swiss</code>. Negative Delta % means <code>noswiss</code> is faster.</p>"
        ),
        "<p><b>Growth Equation:</b> growth_factor = ns_op_at_n / ns_op_at_smallest_n. Lower is better.</p>",
        "<p><b>Optimization Score:</b> score = alpha*ns_op + beta*growth_factor. Start with alpha=0.7, beta=0.3.</p>",
        "<h2>Recommendation by Scale (Simple Median Rule)</h2>",
        summary["recommendation_table_html"],
        "<p>Interactive controls: zoom/pan and legend click to hide/show traces.</p>",
    ]
    for title, fig in figures:
        if len(fig.data) == 0:
            continue
        parts.append(f"<h2>{title}</h2>")
        parts.append(to_html(fig, include_plotlyjs="cdn", full_html=False))
    parts.append("</body></html>")
    out_path.write_text("\n".join(parts), encoding="utf-8")


def main() -> None:
    args = parse_args()
    input_path = Path(args.input)
    outdir = Path(args.outdir)
    outdir.mkdir(parents=True, exist_ok=True)

    df, skipped = parse_benchstat(input_path, args.strict)
    if df.empty:
        raise SystemExit("no benchmark rows parsed from input")

    if args.family != "all":
        df = df[df["family"] == args.family].copy()
        if df.empty:
            raise SystemExit(f"no rows left after family filter: {args.family}")

    csv_path = outdir / "benchmark_parsed.csv"
    df.to_csv(csv_path, index=False)
    rec_df = build_scale_recommendations(df)

    summary = {
        "rows": len(df),
        "skipped": skipped,
        "families": ", ".join(sorted(df["family"].dropna().unique())),
        "impls": ", ".join(sorted(df["impl"].dropna().unique())),
        "recommendation_table_html": recommendation_html(rec_df),
    }
    figures = [
        ("Winner Distribution", fig_winner_distribution(df)),
        ("Median Delta by Workload", fig_median_delta_by_family(df)),
        ("Growth Factor by Scale", fig_growth_by_scale(df)),
        ("Delta % by Key Size", fig_delta_by_keysize(df)),
        ("Shard Sweep", fig_shard_sweep(df)),
    ]
    html_path = outdir / "benchmark_dashboard.html"
    write_dashboard(html_path, figures, summary)

    print(f"Wrote CSV: {csv_path}")
    print(f"Wrote HTML: {html_path}")
    print(f"Rows parsed: {len(df)} | skipped: {skipped}")


if __name__ == "__main__":
    main()
