package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "net/http/pprof"

	"catalog/internal/handlers"
	"catalog/internal/middleware"
	"catalog/internal/store"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	productStore := store.New()
	productHandler := handlers.NewProductHandler(productStore)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /products", productHandler.Create)
	mux.HandleFunc("GET /products", productHandler.List)
	mux.HandleFunc("GET /products/stats", productHandler.Stats)
	mux.HandleFunc("GET /products/{id}", productHandler.Get)
	mux.HandleFunc("PUT /products/{id}", productHandler.Update)
	mux.HandleFunc("DELETE /products/{id}", productHandler.Delete)

	mux.Handle("/debug/", http.DefaultServeMux)

	handler := middleware.Logging(
		middleware.Recovery(
			middleware.JSONOnly(
				middleware.Timing(mux),
			),
		),
	)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		slog.Info("server listening", "addr", server.Addr)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	slog.Info("server shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server shutdown failed", "error", err)
		return
	}

	slog.Info("server stopped")
}
