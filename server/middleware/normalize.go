// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"unicode"
)

// SnakeCaseNormalize intercepts POST requests to /run and /run_sse and
// recursively converts any snake_case JSON keys to camelCase before the
// ADK handler processes the request. This allows API clients to send
// either convention while ADK's strict decoder (DisallowUnknownFields)
// receives the camelCase keys it expects.
//
// Keys that are already camelCase pass through unchanged. Single-word
// keys (no underscore) are never modified. When both snake_case and
// camelCase versions of the same key coexist in an object, the
// camelCase value wins and the snake_case duplicate is dropped.
func SnakeCaseNormalize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !isRunPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		body, err := io.ReadAll(r.Body)
		r.Body.Close()
		if err != nil || len(body) == 0 {
			r.Body = io.NopCloser(bytes.NewReader(body))
			next.ServeHTTP(w, r)
			return
		}

		normalized, changed := normalizeJSON(body)
		if !changed {
			r.Body = io.NopCloser(bytes.NewReader(body))
			next.ServeHTTP(w, r)
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(normalized))
		r.ContentLength = int64(len(normalized))
		next.ServeHTTP(w, r)
	})
}

// isRunPath returns true if the URL path ends with /run or /run_sse.
func isRunPath(path string) bool {
	return strings.HasSuffix(path, "/run") || strings.HasSuffix(path, "/run_sse")
}

// normalizeJSON parses raw JSON bytes, recursively converts snake_case
// keys to camelCase, and re-serializes. Returns the original bytes and
// false if no conversion was needed or parsing fails.
func normalizeJSON(data []byte) ([]byte, bool) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return data, false
	}
	result, changed := normalizeValue(v)
	if !changed {
		return data, false
	}
	out, err := json.Marshal(result)
	if err != nil {
		return data, false
	}
	return out, true
}

// normalizeValue walks any JSON value recursively, converting snake_case
// object keys to camelCase.
func normalizeValue(v interface{}) (interface{}, bool) {
	switch val := v.(type) {
	case map[string]interface{}:
		return normalizeObject(val)
	case []interface{}:
		return normalizeArray(val)
	default:
		return v, false
	}
}

// normalizeObject converts snake_case keys in a JSON object to camelCase
// and recurses into values. When both forms exist (e.g. "app_name" and
// "appName"), the camelCase value takes precedence.
func normalizeObject(obj map[string]interface{}) (map[string]interface{}, bool) {
	result := make(map[string]interface{}, len(obj))
	changed := false

	for key, val := range obj {
		camel := snakeToCamel(key)
		newVal, valChanged := normalizeValue(val)
		if valChanged {
			changed = true
		}
		if camel != key {
			changed = true
			if _, hasCamel := obj[camel]; hasCamel {
				continue
			}
			result[camel] = newVal
		} else {
			result[key] = newVal
		}
	}
	return result, changed
}

// normalizeArray recurses into each element of a JSON array.
func normalizeArray(arr []interface{}) ([]interface{}, bool) {
	changed := false
	result := make([]interface{}, len(arr))
	for i, elem := range arr {
		newElem, elemChanged := normalizeValue(elem)
		if elemChanged {
			changed = true
		}
		result[i] = newElem
	}
	return result, changed
}

// snakeToCamel converts a snake_case string to camelCase.
// Single-word strings and strings without underscores are returned as-is.
// Examples: "app_name" → "appName", "user_id" → "userId", "text" → "text".
func snakeToCamel(s string) string {
	if !strings.Contains(s, "_") {
		return s
	}
	parts := strings.Split(s, "_")
	var b strings.Builder
	b.Grow(len(s))
	for i, part := range parts {
		if part == "" {
			continue
		}
		if i == 0 {
			b.WriteString(part)
			continue
		}
		runes := []rune(part)
		runes[0] = unicode.ToUpper(runes[0])
		b.WriteString(string(runes))
	}
	return b.String()
}
