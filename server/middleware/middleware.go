package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/felixge/httpsnoop"

	"github.com/achetronic/magec/server/store"
)

// AccessLog logs every HTTP request with method, path, status code,
// duration, and response size.
func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		m := httpsnoop.CaptureMetrics(next, w, r)

		logFn := slog.Info
		if m.Code >= 400 {
			logFn = slog.Warn
		}
		logFn("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", m.Code,
			"duration", time.Since(start).Round(time.Millisecond),
			"bytes", m.Written,
		)
	})
}

// ClientAuth protects API endpoints with client token authentication.
// Static files, health checks, CORS preflight, and voice-events pass through.
// If no clients exist in the store, all requests pass through (open mode).
func ClientAuth(next http.Handler, dataStore *store.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if r.Method == http.MethodOptions ||
			path == "/api/v1/health" ||
			path == "/api/v1/voice/events" ||
			strings.HasPrefix(path, "/api/v1/webhooks/") ||
			!strings.HasPrefix(path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}

		clients := dataStore.ListClients()
		if len(clients) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		token := r.Header.Get("Authorization")
		hasToken := strings.HasPrefix(token, "Bearer ")

		if hasToken {
			token = strings.TrimPrefix(token, "Bearer ")
			cl, ok := dataStore.GetClientByToken(token)
			if !ok || !cl.Enabled {
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, `{"error":"invalid or disabled client token"}`, http.StatusUnauthorized)
				return
			}
			r.Header.Set("X-Client-ID", cl.ID)
			next.ServeHTTP(w, r)
			return
		}

		if path == "/api/v1/client/info" {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"missing or invalid Authorization header"}`, http.StatusUnauthorized)
	})
}

// CORS adds permissive CORS headers to all responses and handles
// OPTIONS preflight requests.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
