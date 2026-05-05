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
