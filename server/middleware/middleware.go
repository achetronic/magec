package middleware

import (
	"crypto/subtle"
	"log/slog"
	"net/http"
	"strings"
	"sync"
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
			(strings.HasPrefix(path, "/api/v1/a2a/") && strings.HasSuffix(path, "/.well-known/agent-card.json")) ||
			strings.HasPrefix(path, "/api/v1/webhooks/") ||
			strings.HasPrefix(path, "/api/v1/ephemeral/") ||
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

// AdminAuth protects admin API endpoints with password authentication.
// If password is empty, all requests pass through (open mode).
// Uses constant-time comparison and per-IP rate limiting.
func AdminAuth(next http.Handler, password string) http.Handler {
	if password == "" {
		return next
	}

	rl := newRateLimiter(5, time.Minute)
	go rl.cleanup(30 * time.Second)

	passwordBytes := []byte(password)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		path := r.URL.Path

		if !strings.HasPrefix(path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}

		if path == "/api/v1/admin/auth/check" {
			token := extractBearerToken(r)
			if token == "" || subtle.ConstantTimeCompare([]byte(token), passwordBytes) != 1 {
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"ok":true}`))
			return
		}

		ip := extractIP(r)
		if !rl.allow(ip) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")
			http.Error(w, `{"error":"too many failed attempts, try again later"}`, http.StatusTooManyRequests)
			return
		}

		token := extractBearerToken(r)
		if token == "" {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		if subtle.ConstantTimeCompare([]byte(token), passwordBytes) != 1 {
			rl.record(ip)
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// BearerAuth protects an HTTP handler with `Authorization: Bearer <password>`
// authentication. Used by surfaces that serve at the root path (no `/api/`
// prefix), such as the embedded MCP server. Reuses the same constant-time
// comparison and per-IP rate limiter as [AdminAuth].
//
// If password is empty, all requests pass through (open mode) and a warning
// is the caller's responsibility.
func BearerAuth(next http.Handler, password string) http.Handler {
	if password == "" {
		return next
	}

	rl := newRateLimiter(5, time.Minute)
	go rl.cleanup(30 * time.Second)

	passwordBytes := []byte(password)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		ip := extractIP(r)
		if !rl.allow(ip) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")
			http.Error(w, `{"error":"too many failed attempts, try again later"}`, http.StatusTooManyRequests)
			return
		}

		token := extractBearerToken(r)
		if token == "" {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		if subtle.ConstantTimeCompare([]byte(token), passwordBytes) != 1 {
			rl.record(ip)
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}

func extractIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		parts := strings.SplitN(fwd, ",", 2)
		return strings.TrimSpace(parts[0])
	}
	host := r.RemoteAddr
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		return host[:idx]
	}
	return host
}

type rateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	max      int
	window   time.Duration
}

func newRateLimiter(max int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		attempts: make(map[string][]time.Time),
		max:      max,
		window:   window,
	}
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-rl.window)

	var recent []time.Time
	for _, t := range rl.attempts[ip] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	rl.attempts[ip] = recent
	return len(recent) < rl.max
}

func (rl *rateLimiter) record(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.attempts[ip] = append(rl.attempts[ip], time.Now())
}

func (rl *rateLimiter) cleanup(interval time.Duration) {
	for {
		time.Sleep(interval)
		rl.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-rl.window)
		for ip, times := range rl.attempts {
			var recent []time.Time
			for _, t := range times {
				if t.After(cutoff) {
					recent = append(recent, t)
				}
			}
			if len(recent) == 0 {
				delete(rl.attempts, ip)
			} else {
				rl.attempts[ip] = recent
			}
		}
		rl.mu.Unlock()
	}
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
