package cli

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"jainam-panchal/nextgen-training/stock-price-tracker/internal/config"
	"jainam-panchal/nextgen-training/stock-price-tracker/internal/tracker"
)

func Run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// buffer update logs
	resizeLogFile, err := os.OpenFile("buffer-resize.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer resizeLogFile.Close()

	// init all datapoints
	trackers := make(map[string]*tracker.StockTracker, len(cfg.Symbols))
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for _, symbol := range cfg.Symbols {
		stockTracker, err := tracker.NewStockTracker(symbol, cfg.WindowSize)
		if err != nil {
			return err
		}

		_, _ = stockTracker.SeedInitialPoint(time.Now(), rng)
		trackers[symbol] = stockTracker
	}

	ticker := time.NewTicker(cfg.UpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case now := <-ticker.C:
			fmt.Print("\033[H\033[2J")
			fmt.Println("=== Stock Price Dashboard ===")
			fmt.Printf("Updated: %s | Interval: %s | Window: %d\n\n", now.Format("15:04:05"), cfg.UpdateInterval, cfg.WindowSize)
			fmt.Printf("%-12s %-10s %10s %10s %10s %10s %10s %10s\n", "SYMBOL", "STATUS", "PRICE", "SMA", "MIN", "MAX", "VOLUME", "POINTS")
			fmt.Println("-----------------------------------------------------------------------------------------------")

			for _, symbol := range cfg.Symbols {
				stockTracker := trackers[symbol]
				point := stockTracker.NextPricePoint(now, rng)
				oldCap, newCap := stockTracker.AddPrice(point)
				minPrice, maxPrice := stockTracker.MinMax()
				if newCap > oldCap {
					fmt.Fprintf(
						resizeLogFile,
						"%s | %s | buffer resized | %d -> %d\n",
						now.Format(time.RFC3339),
						symbol,
						oldCap,
						newCap,
					)
				}

				fmt.Printf(
					"%-12s %-10s %10.2f %10.2f %10.2f %10.2f %10d %10d\n",
					stockTracker.Symbol,
					stockTracker.Status,
					stockTracker.CurrentPrice(),
					stockTracker.SMA(),
					minPrice,
					maxPrice,
					point.Volume,
					len(stockTracker.Window),
				)
			}

			fmt.Println("\nResize events are written to buffer-resize.log")
		}
	}
}
