package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jainam-panchal/nextgen-training/capstone/internal/engine"
	"github.com/jainam-panchal/nextgen-training/capstone/internal/httpapi"
	"github.com/jainam-panchal/nextgen-training/capstone/internal/model"
	"github.com/jainam-panchal/nextgen-training/capstone/internal/sim"
)

func main() {
	addr := flag.String("addr", ":8080", "http listen address")
	tick := flag.Duration("tick", 1*time.Second, "simulation tick duration")
	flag.Parse()

	cfg := engine.DefaultConfig()
	cfg.TickDuration = *tick

	// Use the built-in 20-node demo graph with shortcuts.
	g := model.GenerateDemoGraph()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Create engine with real-time clock.
	clock := sim.NewRealClock(ctx, cfg.TickDuration)
	e, err := engine.New(cfg, g, clock)
	if err != nil {
		log.Fatal(err)
	}
	e.Start(ctx)
	defer e.Stop()

	// Mount REST API + pprof debug endpoints.
	api := httpapi.New(e)
	mux := http.NewServeMux()
	mux.Handle("/", api.Handler())
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	log.Printf("listening on %s", *addr)
	err = httpapi.RunHTTPServer(ctx, *addr, mux)
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}
}
