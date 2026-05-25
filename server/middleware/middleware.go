package middleware

import (
	"crypto/subtle"
	"fmt"
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

	check := newPasswordCheck(password)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		// Static files served by the admin UI live outside /api/ and must
		// pass through unauthenticated so the SPA can load before the user
		// types the password.
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}

		// /auth/check lets the admin UI validate a password without
		// triggering the rate limiter — otherwise an interactive prompt
		// would lock the user's IP out after a handful of typos.
		if r.URL.Path == "/api/v1/admin/auth/check" {
			if check.matches(r) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"ok":true}`))
				return
			}
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		if !check.allow(w, r) {
			return
		}
		next.ServeHTTP(w, r)
	})
}

// BearerAuth protects an HTTP handler with `Authorization: Bearer <password>`
// authentication. Used by surfaces that serve at the root path (no `/api/`
// prefix), such as the embedded MCP server. Shares the rate-limited bearer
// check with [AdminAuth]; the only behavioural difference is that this
// middleware does not carve out the SPA static-file and auth-check paths.
//
// If password is empty, all requests pass through (open mode) and emitting
// a warning is the caller's responsibility.
func BearerAuth(next http.Handler, password string) http.Handler {
	if password == "" {
		return next
	}

	check := newPasswordCheck(password)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		if !check.allow(w, r) {
			return
		}
		next.ServeHTTP(w, r)
	})
}

// passwordCheck is the shared state for the bearer-token middlewares:
// the constant-time-comparable secret and the per-IP rate limiter that
// throttles repeated bad attempts.
type passwordCheck struct {
	passwordBytes []byte
	limiter       *rateLimiter
}

// newPasswordCheck returns a checker pre-wired with a 5-failure-per-minute
// rate limiter. The cleanup goroutine runs for the lifetime of the process,
// which is fine because middlewares are constructed once at startup.
func newPasswordCheck(password string) *passwordCheck {
	rl := newRateLimiter(5, time.Minute)
	go rl.cleanup(30 * time.Second)
	return &passwordCheck{
		passwordBytes: []byte(password),
		limiter:       rl,
	}
}

// matches reports whether the request carries the expected bearer token.
// It does not touch the rate limiter and never writes to the response;
// callers use it for endpoints where validation must stay silent.
func (c *passwordCheck) matches(r *http.Request) bool {
	token := extractBearerToken(r)
	if token == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(token), c.passwordBytes) == 1
}

// allow runs the full rate-limited bearer check. When it returns true the
// caller may serve the request; when it returns false the response has
// already been written (401 or 429) and the caller must stop.
func (c *passwordCheck) allow(w http.ResponseWriter, r *http.Request) bool {
	ip := extractIP(r)
	if !c.limiter.allow(ip) {
		w.Header().Set("Retry-After", "60")
		writeJSONError(w, http.StatusTooManyRequests, "too many failed attempts, try again later")
		return false
	}
	token := extractBearerToken(r)
	if token == "" {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return false
	}
	if subtle.ConstantTimeCompare([]byte(token), c.passwordBytes) != 1 {
		c.limiter.record(ip)
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return false
	}
	return true
}

// writeJSONError emits the canonical {"error":"..."} body that every other
// middleware in this package returns on a hard failure.
func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	http.Error(w, fmt.Sprintf(`{"error":%q}`, message), status)
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
