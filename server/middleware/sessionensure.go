// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"net/http/httptest"
)

// SessionEnsure intercepts session creation requests (POST to
// /apps/{app}/users/{user}/sessions/{session}) and makes them idempotent.
// If the session already exists, it returns the existing session without
// forwarding the create request. This prevents the ADK Create handler from
// overwriting accumulated session state (e.g. context guard summaries) on
// every message from chat clients that call ensureSession unconditionally.
//
// It works by issuing an internal GET for the same path against the inner
// handler chain. If the GET returns 200, the session exists and its response
// is returned directly. Otherwise the original POST is forwarded as usual.
//
// Place this middleware BEFORE SessionStateSeed in the chain so that
// existing sessions skip both the state seeding and the ADK Create call.
func SessionEnsure(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !isSessionCreatePath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		probe, err := http.NewRequestWithContext(r.Context(), http.MethodGet, r.URL.String(), nil)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		probe.Header = r.Header.Clone()

		rec := httptest.NewRecorder()
		next.ServeHTTP(rec, probe)

		if rec.Code == http.StatusOK {
			result := rec.Result()
			defer result.Body.Close()

			for k, vals := range rec.Header() {
				for _, v := range vals {
					w.Header().Add(k, v)
				}
			}
			w.WriteHeader(http.StatusOK)
			rec.Body.WriteTo(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}
