#!/usr/bin/env python3
"""quant_engine.py —— 离线量化指标计算（Sharpe / MaxDD / 年化收益）。

输入：从 stdin 读 CSV (date,close)，或 --pair/--tf 从 PG 拉 price_daily（需环境
DATABASE_URL）。

输出 JSON：sharpe / max_drawdown / annual_return / n_bars。
"""
import argparse
import json
import math
import os
import sys


def load_csv(path):
    rows = []
    if path == "-":
        f = sys.stdin
    else:
        f = open(path, "r")
    for line in f:
        line = line.strip()
        if not line or line.startswith("#") or line.lower().startswith("date"):
            continue
        parts = line.split(",")
        if len(parts) >= 2:
            try:
                rows.append(float(parts[1]))
            except ValueError:
                pass
    return rows


def load_pg(pair, tf):
    try:
        import psycopg2
    except ImportError:
        print("psycopg2 not installed", file=sys.stderr)
        sys.exit(2)
    dsn = os.getenv("DATABASE_URL")
    if not dsn:
        print("DATABASE_URL not set", file=sys.stderr)
        sys.exit(2)
    table = "price_daily" if tf in ("1d", "d", "daily", "") else "price_intraday"
    sql = f"SELECT close FROM {table} WHERE symbol=%s AND close>0 ORDER BY time ASC"
    args = [pair]
    if table == "price_intraday":
        sql += " AND interval=%s"
        args.append(tf)
    conn = psycopg2.connect(dsn)
    cur = conn.cursor()
    cur.execute(sql, args)
    return [r[0] for r in cur.fetchall()]


def compute(closes):
    if len(closes) < 5:
        return {"error": "insufficient bars", "n_bars": len(closes)}
    rets = [closes[i] / closes[i - 1] - 1 for i in range(1, len(closes))]
    mean = sum(rets) / len(rets)
    var = sum((r - mean) ** 2 for r in rets) / len(rets)
    std = math.sqrt(var)
    sharpe = (mean / std * math.sqrt(252)) if std > 0 else 0.0
    eq = 1.0
    peak = 1.0
    mdd = 0.0
    for r in rets:
        eq *= 1 + r
        if eq > peak:
            peak = eq
        dd = (peak - eq) / peak if peak > 0 else 0.0
        if dd > mdd:
            mdd = dd
    annual = (closes[-1] / closes[0]) ** (252 / len(rets)) - 1
    return {
        "sharpe": round(sharpe, 4),
        "max_drawdown": round(mdd, 4),
        "annual_return": round(annual, 4),
        "n_bars": len(closes),
    }


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--pair")
    ap.add_argument("--tf", default="1d")
    ap.add_argument("--csv", default="-")
    args = ap.parse_args()

    if args.pair:
        closes = load_pg(args.pair, args.tf)
    else:
        closes = load_csv(args.csv)
    print(json.dumps(compute(closes)))


if __name__ == "__main__":
    main()
