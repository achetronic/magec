// Copyright 2025 Alberto Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package secrets lets agents and flow nodes use store secrets through
// placeholders without the raw values ever reaching the model: placeholders
// expand right before a tool or transform runs, and known values are redacted
// back to placeholders on every path that returns to the model or persists.
package secrets

import (
	"regexp"
	"strings"

	"github.com/achetronic/magec/server/store"
)

// minRedactLength keeps trivially short values out of the redactor: replacing
// something like "1234" would mangle unrelated text everywhere.
const minRedactLength = 6

// placeholderRE matches {{secret:KEY}} with optional inner whitespace.
var placeholderRE = regexp.MustCompile(`\{\{\s*secret:([A-Za-z_][A-Za-z0-9_]*)\s*\}\}`)

// Snapshot holds the secret material of one store generation, indexed for
// expansion and redaction. Build one per store rebuild and share it.
type Snapshot struct {
	byKey    map[string]string
	replacer *strings.Replacer
	keys     []string
}

// NewSnapshot indexes the given secrets. Values shorter than the redaction
// minimum still expand but are never redacted.
func NewSnapshot(secrets []store.Secret) *Snapshot {
	s := &Snapshot{byKey: make(map[string]string, len(secrets))}
	var pairs []string
	for _, sec := range secrets {
		if sec.Key == "" || sec.Value == "" {
			continue
		}
		s.byKey[sec.Key] = sec.Value
		s.keys = append(s.keys, sec.Key)
		if len(sec.Value) >= minRedactLength {
			pairs = append(pairs, sec.Value, "{{secret:"+sec.Key+"}}")
		}
	}
	s.replacer = strings.NewReplacer(pairs...)
	return s
}

// Keys returns every secret key in the snapshot.
func (s *Snapshot) Keys() []string {
	return s.keys
}

// Value returns the raw value for a key.
func (s *Snapshot) Value(key string) (string, bool) {
	v, ok := s.byKey[key]
	return v, ok
}

// Map returns key to value for the given allowed keys; nil allowed means all.
func (s *Snapshot) Map(allowed []string, allowAll bool) map[string]string {
	out := map[string]string{}
	if allowAll {
		for k, v := range s.byKey {
			out[k] = v
		}
		return out
	}
	for _, k := range allowed {
		if v, ok := s.byKey[k]; ok {
			out[k] = v
		}
	}
	return out
}

// ExpandString replaces every {{secret:KEY}} placeholder whose key exists in
// the snapshot and is allowed. Unknown or disallowed keys stay as-is.
func (s *Snapshot) ExpandString(text string, allowed map[string]string) string {
	if !strings.Contains(text, "{{") {
		return text
	}
	return placeholderRE.ReplaceAllStringFunc(text, func(match string) string {
		key := placeholderRE.FindStringSubmatch(match)[1]
		if v, ok := allowed[key]; ok {
			return v
		}
		return match
	})
}

// ExpandValue walks strings, maps and slices expanding placeholders, building
// copies along the way: the input is never mutated, so shared references
// (session events, recorded payloads) keep their placeholders.
func (s *Snapshot) ExpandValue(v any, allowed map[string]string) any {
	switch val := v.(type) {
	case string:
		return s.ExpandString(val, allowed)
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, item := range val {
			out[k] = s.ExpandValue(item, allowed)
		}
		return out
	case []any:
		out := make([]any, len(val))
		for i, item := range val {
			out[i] = s.ExpandValue(item, allowed)
		}
		return out
	default:
		return v
	}
}

// RedactString replaces every known secret value with its placeholder.
func (s *Snapshot) RedactString(text string) string {
	return s.replacer.Replace(text)
}

// RedactValue walks strings, maps and slices replacing secret values with
// their placeholders, building copies along the way.
func (s *Snapshot) RedactValue(v any) any {
	switch val := v.(type) {
	case string:
		return s.RedactString(val)
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, item := range val {
			out[k] = s.RedactValue(item)
		}
		return out
	case []any:
		out := make([]any, len(val))
		for i, item := range val {
			out[i] = s.RedactValue(item)
		}
		return out
	default:
		return v
	}
}
