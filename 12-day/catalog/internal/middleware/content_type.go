package middleware

import (
	"net/http"
	"strings"
)

func JSONOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			contentType := r.Header.Get("Content-Type")

			if contentType != "" &&
				contentType != "application/json" &&
				!strings.HasPrefix(contentType, "application/json;") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnsupportedMediaType)
				_, _ = w.Write([]byte(`{"error":"Content-Type must be application/json"}`))
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
