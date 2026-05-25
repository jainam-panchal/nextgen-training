package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"realtime-auction/internal/auction"
	"realtime-auction/internal/handlers"
	"realtime-auction/internal/middleware"
)

func main() {
	runtime.SetMutexProfileFraction(1)

	store := auction.NewAuctionStore()
	service := auction.NewAuctionService(store)
	handler := handlers.NewAuctionHandler(service)

	mux := http.NewServeMux()
	handler.Register(mux)
	mux.Handle("/debug/pprof/", http.DefaultServeMux)

	limiter := middleware.NewBidRateLimiter(
		5,
		time.Minute,
		func(r *http.Request) bool {
			if r.Method != http.MethodPost {
				return false
			}
			parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
			if len(parts) != 3 || parts[0] != "items" || parts[2] != "bid" {
				return false
			}
			_, err := strconv.Atoi(parts[1])
			return err == nil
		},
		middleware.KeyFromUserContext,
	)
	root := middleware.Chain(
		mux,
		middleware.Recovery,
		middleware.Logging,
		middleware.AuthSimulation,
		limiter.Middleware,
		middleware.ResponseTiming,
	)

	server := &http.Server{
		Addr:              ":8081",
		Handler:           root,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("server starting on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
