// Package config provides configuration structs and default values.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Symbols        []string
	WindowSize     int
	UpdateInterval time.Duration
}

var defaultSymbols = []string{"RELIANCE", "TCS", "INFY", "HDFCBANK", "ITC"}

const (
	defaultWindowSize       = 100
	defaultUpdateIntervalMs = 500
)

func Load() (Config, error) {
	symbols := defaultSymbols

	rawSymbols := strings.TrimSpace(os.Getenv("STOCK_SYMBOLS"))
	if rawSymbols != "" {
		parts := strings.Split(rawSymbols, ",")
		symbols = make([]string, 0, len(parts))

		for _, part := range parts {
			symbol := strings.ToUpper(strings.TrimSpace(part))
			if symbol != "" {
				symbols = append(symbols, symbol)
			}
		}

		if len(symbols) < 5 {
			return Config{}, fmt.Errorf("CONFIG || STOCK_SYMBOLS must contain at least five symbols")
		}
	}

	windowSize, err := loadPositiveInt("STOCK_WINDOW_SIZE", defaultWindowSize)
	if err != nil {
		return Config{}, err
	}

	updateIntervalMs, err := loadPositiveInt("STOCK_UPDATE_INTERVAL_MS", defaultUpdateIntervalMs)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Symbols:        symbols,
		WindowSize:     windowSize,
		UpdateInterval: time.Duration(updateIntervalMs) * time.Millisecond,
	}, nil
}

func loadPositiveInt(key string, defaultValue int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("CONFIG || %s must be a valid integer", key)
	}

	if value <= 0 {
		return 0, fmt.Errorf("CONFIG || %s must be greater than 0", key)
	}

	return value, nil
}
