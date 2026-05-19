package middleware

import (
	"net/http"
	"strconv"
	"time"
)

type timingRecorder struct {
	http.ResponseWriter
	startedAt time.Time
	wrote     bool
}

func (recorder *timingRecorder) WriteHeader(status int) {
	if !recorder.wrote {
		duration := time.Since(recorder.startedAt).Milliseconds()
		recorder.Header().Set("X-Response-Time-Ms", strconv.FormatInt(duration, 10))
		recorder.wrote = true
	}

	recorder.ResponseWriter.WriteHeader(status)
}

func (recorder *timingRecorder) Write(body []byte) (int, error) {
	if !recorder.wrote {
		recorder.WriteHeader(http.StatusOK)
	}

	return recorder.ResponseWriter.Write(body)
}

func Timing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &timingRecorder{
			ResponseWriter: w,
			startedAt:      time.Now(),
		}

		next.ServeHTTP(recorder, r)
	})
}
