package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type ctxKey string

const UserIDContextKey ctxKey = "user_id"

type statusWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

type timingWriter struct {
	http.ResponseWriter
	start time.Time
	set   bool
}

func (w *timingWriter) WriteHeader(statusCode int) {
	if !w.set {
		w.Header().Set("X-Response-Time", time.Since(w.start).String())
		w.set = true
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *timingWriter) Write(b []byte) (int, error) {
	if !w.set {
		w.Header().Set("X-Response-Time", time.Since(w.start).String())
		w.set = true
	}
	return w.ResponseWriter.Write(b)
}

func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"code":  "panic_recovered",
					"error": "internal server error",
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func ResponseTiming(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tw := &timingWriter{ResponseWriter: w, start: time.Now()}
		next.ServeHTTP(tw, r)
	})
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(sw, r)
		log.Printf("method=%s path=%s status=%d duration=%s", r.Method, r.URL.Path, sw.statusCode, time.Since(start))
	})
}

func AuthSimulation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userHeader := r.Header.Get("X-User-ID")
		if userHeader != "" {
			if id, err := strconv.Atoi(userHeader); err == nil {
				ctx := context.WithValue(r.Context(), UserIDContextKey, id)
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	})
}

type tokenBucket struct {
	tokens     int
	lastRefill time.Time
}

type MatchFunc func(*http.Request) bool
type KeyFunc func(*http.Request) (string, bool)

type BidRateLimiter struct {
	mu          sync.Mutex
	buckets     map[string]*tokenBucket
	limit       int
	window      time.Duration
	match       MatchFunc
	extractKey  KeyFunc
	rejectCode  int
	rejectBody  map[string]string
}

func NewBidRateLimiter(limit int, window time.Duration, match MatchFunc, extractKey KeyFunc) *BidRateLimiter {
	if limit <= 0 {
		limit = 5
	}
	if window <= 0 {
		window = time.Minute
	}
	return &BidRateLimiter{
		buckets:    make(map[string]*tokenBucket),
		limit:      limit,
		window:     window,
		match:      match,
		extractKey: extractKey,
		rejectCode: http.StatusTooManyRequests,
		rejectBody: map[string]string{
			"code":  "rate_limited",
			"error": "max 5 bids per user per minute",
		},
	}
}

func (rl *BidRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rl.match != nil && !rl.match(r) {
			next.ServeHTTP(w, r)
			return
		}

		if rl.extractKey == nil {
			next.ServeHTTP(w, r)
			return
		}
		key, ok := rl.extractKey(r)
		if !ok || key == "" {
			next.ServeHTTP(w, r)
			return
		}

		if !rl.allow(key) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(rl.rejectCode)
			_ = json.NewEncoder(w).Encode(rl.rejectBody)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func KeyFromUserContext(r *http.Request) (string, bool) {
	id, ok := r.Context().Value(UserIDContextKey).(int)
	if !ok || id == 0 {
		return "", false
	}
	return strconv.Itoa(id), true
}

func (rl *BidRateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now().UTC()
	bucket, exists := rl.buckets[key]
	if !exists {
		rl.buckets[key] = &tokenBucket{tokens: rl.limit - 1, lastRefill: now}
		return true
	}

	if now.Sub(bucket.lastRefill) >= rl.window {
		bucket.tokens = rl.limit
		bucket.lastRefill = now
	}

	if bucket.tokens <= 0 {
		return false
	}

	bucket.tokens--
	return true
}
