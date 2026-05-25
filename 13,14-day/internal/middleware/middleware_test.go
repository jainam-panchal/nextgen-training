package middleware

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAuthSimulationSetsUserIDInContext(t *testing.T) {
	var got int
	h := AuthSimulation(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		if v, ok := r.Context().Value(UserIDContextKey).(int); ok {
			got = v
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-User-ID", "42")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if got != 42 {
		t.Fatalf("expected user id 42 in context, got %d", got)
	}
}

func TestAuthSimulationIgnoresInvalidHeader(t *testing.T) {
	var found bool
	h := AuthSimulation(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		_, found = r.Context().Value(UserIDContextKey).(int)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-User-ID", "not-an-int")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if found {
		t.Fatal("expected invalid X-User-ID to be ignored")
	}
}

func TestResponseTimingHeaderIsSet(t *testing.T) {
	h := ResponseTiming(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Header().Get("X-Response-Time") == "" {
		t.Fatal("expected X-Response-Time header to be set")
	}
}

func TestRecoveryHandlesPanic(t *testing.T) {
	h := Recovery(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("boom")
	}))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
	if rr.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected application/json content type, got %q", rr.Header().Get("Content-Type"))
	}
}

func TestRateLimiterBlocksAfterFiveBidsPerMinute(t *testing.T) {
	rl := NewBidRateLimiter(
		5,
		time.Minute,
		func(r *http.Request) bool { return r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/items/") && strings.HasSuffix(r.URL.Path, "/bid") },
		KeyFromUserContext,
	)
	h := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, "ok")
	}))

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/items/1/bid", nil)
		req = req.WithContext(context.WithValue(req.Context(), UserIDContextKey, 7))
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusCreated {
			t.Fatalf("expected 201 on request %d, got %d", i+1, rr.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/items/1/bid", nil)
	req = req.WithContext(context.WithValue(req.Context(), UserIDContextKey, 7))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 on sixth request, got %d", rr.Code)
	}
}

func TestRateLimiterOnlyAppliesToBidEndpoint(t *testing.T) {
	rl := NewBidRateLimiter(
		5,
		time.Minute,
		func(r *http.Request) bool { return r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/items/") && strings.HasSuffix(r.URL.Path, "/bid") },
		KeyFromUserContext,
	)
	h := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	// Non-bid path should bypass limiter.
	for i := 0; i < 8; i++ {
		req := httptest.NewRequest(http.MethodPost, "/items/1/end", nil)
		req = req.WithContext(context.WithValue(req.Context(), UserIDContextKey, 7))
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusCreated {
			t.Fatalf("expected non-bid path to bypass limiter, got %d", rr.Code)
		}
	}
}

func TestRateLimiterWithoutUserIDBypasses(t *testing.T) {
	rl := NewBidRateLimiter(
		5,
		time.Minute,
		func(r *http.Request) bool { return r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/items/") && strings.HasSuffix(r.URL.Path, "/bid") },
		KeyFromUserContext,
	)
	h := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	for i := 0; i < 8; i++ {
		req := httptest.NewRequest(http.MethodPost, "/items/1/bid", nil)
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusCreated {
			t.Fatalf("expected request without user id to bypass limiter, got %d", rr.Code)
		}
	}
}

func TestKeyFromUserContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req = req.WithContext(context.WithValue(req.Context(), UserIDContextKey, 17))
	key, ok := KeyFromUserContext(req)
	if !ok || key != "17" {
		t.Fatalf("expected key 17, got %q (ok=%v)", key, ok)
	}
}
