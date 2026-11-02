// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"strings"
	"time"
)

type sseIdleWriter struct {
	http.ResponseWriter
	rc      *http.ResponseController
	timeout time.Duration
}

func (w *sseIdleWriter) Write(b []byte) (int, error) {
	w.rc.SetWriteDeadline(time.Now().Add(w.timeout))
	return w.ResponseWriter.Write(b)
}

// SetWriteDeadline ignores the deadline the wrapped handler asks for (the ADK
// sets a fixed one when /run_sse starts) and applies the idle timeout instead,
// so the client's timeout is the only limit on the stream.
func (w *sseIdleWriter) SetWriteDeadline(time.Time) error {
	return w.rc.SetWriteDeadline(time.Now().Add(w.timeout))
}

func (w *sseIdleWriter) Flush() {
	w.rc.Flush()
}

func (w *sseIdleWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// SSEIdleTimeout wraps an HTTP handler to automatically timeout
// long-running Server-Sent Events (/run_sse) requests if no data
// is written to the response within the duration timeoutFor returns
// for that request (the calling client's SSE timeout).
func SSEIdleTimeout(next http.Handler, timeoutFor func(*http.Request) time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !(strings.HasSuffix(r.URL.Path, "/run_sse") && r.Method == "POST") {
			next.ServeHTTP(w, r)
			return
		}

		iw := &sseIdleWriter{
			ResponseWriter: w,
			rc:             http.NewResponseController(w),
			timeout:        timeoutFor(r),
		}
		next.ServeHTTP(iw, r)
	})
}
