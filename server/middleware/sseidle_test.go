// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// deadlineRecorder is a ResponseWriter that records the last write deadline
// set through http.ResponseController.
type deadlineRecorder struct {
	*httptest.ResponseRecorder
	deadline time.Time
}

func (d *deadlineRecorder) SetWriteDeadline(t time.Time) error {
	d.deadline = t
	return nil
}

// TestSSEIdleTimeoutOverridesHandlerDeadline is the canary for issue #78: the
// ADK sets a fixed write deadline when /run_sse starts, and if it survived it
// would cut the stream before the client's own timeout. The middleware must
// replace it, and refresh it on every write, with the per-request timeout.
func TestSSEIdleTimeoutOverridesHandlerDeadline(t *testing.T) {
	const timeout = 2 * time.Hour
	var afterHandlerDeadline time.Time
	rec := &deadlineRecorder{ResponseRecorder: httptest.NewRecorder()}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// What the ADK does at the start of RunSSEHandler.
		_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(time.Minute))
		afterHandlerDeadline = rec.deadline
		_, _ = w.Write([]byte("data: {}\n\n"))
	})
	h := SSEIdleTimeout(inner, func(*http.Request) time.Duration { return timeout })

	start := time.Now()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/agent/run_sse", nil))

	for name, got := range map[string]time.Time{"handler deadline": afterHandlerDeadline, "write deadline": rec.deadline} {
		if got.Before(start.Add(timeout)) {
			t.Fatalf("%s = %v from start, want at least %v", name, got.Sub(start), timeout)
		}
	}
}

// TestSSEIdleTimeoutIgnoresOtherPaths verifies only POST /run_sse is wrapped.
func TestSSEIdleTimeoutIgnoresOtherPaths(t *testing.T) {
	called := false
	h := SSEIdleTimeout(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := w.(*sseIdleWriter); ok {
			t.Fatal("non-SSE request was wrapped")
		}
	}), func(*http.Request) time.Duration { called = true; return time.Hour })

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/v1/agent/run", nil))
	if called {
		t.Fatal("timeout resolved for a non-SSE request")
	}
}
