package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jainam-panchal/nextgen-training/capstone/internal/engine"
)

// Server wraps an engine.Engine with HTTP handlers.
type Server struct {
	e   *engine.Engine
	mux *http.ServeMux
}

// New creates an HTTP API server wrapping the given engine.
func New(e *engine.Engine) *Server {
	s := &Server{e: e, mux: http.NewServeMux()}
	s.routes()
	return s
}

// Handler returns the HTTP handler (for use with http.Server or httptest).
func (s *Server) Handler() http.Handler { return s.mux }

// routes registers all API endpoints.
func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.handleHealth)
	s.mux.HandleFunc("POST /vehicles", s.handleRegisterVehicle)
	s.mux.HandleFunc("GET /route", s.handleRoute)
	s.mux.HandleFunc("POST /emergency", s.handleEmergency)
	s.mux.HandleFunc("GET /congestion", s.handleCongestion)
	s.mux.HandleFunc("GET /signals/", s.handleSignalStatus)
	s.mux.HandleFunc("GET /stats", s.handleStats)
}

// GET /healthz — liveness check.
func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// POST /vehicles — register a new vehicle.
func (s *Server) handleRegisterVehicle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req engine.RegisterVehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("decode json: %w", err))
		return
	}
	resp, err := s.e.RegisterVehicle(ctx, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// GET /route?from=<id>&to=<id> — shortest path query.
func (s *Server) handleRoute(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()
	from, err := atoiParam(q.Get("from"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	to, err := atoiParam(q.Get("to"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	resp, err := s.e.Route(ctx, from, to)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// POST /emergency — dispatch an emergency vehicle (creates corridor + preempts signals).
func (s *Server) handleEmergency(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req engine.DispatchEmergencyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("decode json: %w", err))
		return
	}
	resp, err := s.e.DispatchEmergency(ctx, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// GET /congestion — congestion report for all roads.
func (s *Server) handleCongestion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	resp, err := s.e.Congestion(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// GET /signals/{id} — signal status at one intersection.
func (s *Server) handleSignalStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/signals/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusBadRequest, errors.New("missing id"))
		return
	}
	id, err := atoiParam(parts[0])
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	resp, err := s.e.SignalStatus(ctx, id)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// GET /stats — simulation statistics.
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	resp, err := s.e.Stats(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func atoiParam(v string) (int, error) {
	if v == "" {
		return 0, errors.New("missing parameter")
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("invalid int: %w", err)
	}
	return i, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

// RunHTTPServer runs an http.Server until ctx is cancelled, then gracefully shuts down.
func RunHTTPServer(ctx context.Context, addr string, h http.Handler) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
		return ctx.Err()
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
