## Problem Statement

Build a stock price tracking system in Go:

1. STRUCTURE:

- Track prices for multiple stocks (at least 5: RELIANCE, TCS, INFY, HDFCBANK, ITC)
- Each stock maintains a dynamic price history (slice of PricePoint)
- PricePoint: { Price float64, Timestamp time.Time, Volume int }

2. CORE FEATURES:

- Simulate real-time updates: random price changes every 100ms (use goroutine — preview)
- Maintain a sliding window of last N prices (configurable: default 100)
- Calculate: current price, min/max in window, simple moving average (SMA)
- Display live dashboard (clear screen + reprint)

3. DYNAMIC ARRAY FOCUS:

- Start with capacity 10, observe growth as prices arrive
- Implement circular buffer variant for the sliding window
- Log every resize event: "Buffer resized: cap 64 → 128"
- Implement manual shrink when window slides past old data

4. GO REQUIREMENTS:

- Use struct with methods (not free functions)
- Implement Stringer interface for display
- Handle edge cases: empty history, single data point

## Approximate Memory Usage

These numbers are approximate for a 64-bit Go runtime.

- `PricePoint` uses about **40 bytes**
- `StockTracker` struct itself uses about **80 bytes**
- most memory is used by the `Window` slice, because it stores recent `PricePoint` values

## Memory By Window Size

The `Window` slice grows based on how many recent price points are stored.

- window size `10` -> about **400 bytes**
- window size `100` -> about **4000 bytes** (~4 KB)
- window size `500` -> about **20000 bytes** (~20 KB)

So one tracker at window size `100` uses roughly:

- tracker struct: **80 bytes**
- window data: **4000 bytes**
- total: about **4080 bytes**

For 5 stocks at window size `100`, total tracker data is about:

- **20400 bytes** (~20 KB)

The `Window` slice starts with capacity `10` and grows with `append()`. When old points are removed, the active window gets smaller, but the old backing array may still stay in memory. That is why the project also uses manual shrink logic.

## Naive Slice Comparison

If we kept full history in a normal slice and stored `1,000,000` price points for one stock:

- one `PricePoint` ~= **40 bytes**
- `1,000,000 * 40` ~= **40,000,000 bytes**
- total ~= **38 MB per stock**

With the current rolling window approach at size `100`: window data ~= **4000 bytes** (~4 KB per stock)
