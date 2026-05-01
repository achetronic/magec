// Package ephemeral provides primitives for issuing self-contained, signed,
// time-limited tokens that grant one-shot access to a specific resource.
//
// Tokens are JWT-like in spirit (a base64url-encoded payload and an
// HMAC-SHA256 signature, joined by a single dot) but deliberately stripped of
// JWT's algorithm-negotiation surface: HMAC-SHA256 is the only supported
// algorithm, and the algorithm is not encoded in the token. The verifier
// recomputes the signature with its configured secret; mismatch or expired
// payload returns an error.
//
// Tokens carry the entire descriptor of what they grant access to (payload
// fields like AppName, UserID, SessionID, Name and the absolute Unix
// expiration timestamp). No state is kept server-side: any binary that
// shares the secret can verify a token, and any restart that preserves the
// secret keeps tokens valid until their natural expiration.
//
// The package is generic enough to back additional ephemeral resources in
// the future under the same `/api/v1/ephemeral/{kind}/{token}` namespace.
package ephemeral

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ArtifactPayload is the payload encoded inside an ephemeral artifact token.
// Field names are camelCase to align with the rest of the user API JSON
// surface; the keys travel inside the token so they should stay short.
type ArtifactPayload struct {
	AppName   string `json:"app"`
	UserID    string `json:"usr"`
	SessionID string `json:"ses"`
	Name      string `json:"nam"`
	ExpiresAt int64  `json:"exp"`
}

// SignArtifact serializes the payload, signs it with HMAC-SHA256 using the
// supplied secret and returns the assembled token. The secret must be
// non-empty; callers (typically server bootstrap) are responsible for
// surfacing that requirement to the operator with a clear message.
func SignArtifact(p ArtifactPayload, secret string) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("ephemeral: signing secret is required")
	}
	body, err := json.Marshal(p)
	if err != nil {
		return "", fmt.Errorf("ephemeral: marshal payload: %w", err)
	}
	encodedBody := base64.RawURLEncoding.EncodeToString(body)
	sig := computeSignature(encodedBody, secret)
	return encodedBody + "." + sig, nil
}

// VerifyArtifact parses a token, validates its HMAC-SHA256 signature against
// the supplied secret and checks the payload has not expired. The returned
// payload is safe to use for resource lookup.
//
// Errors returned do not leak which step failed beyond the level of detail
// callers need: the endpoint should map any non-nil error to 401/403 without
// surfacing the specific reason.
func VerifyArtifact(token, secret string) (ArtifactPayload, error) {
	if secret == "" {
		return ArtifactPayload{}, fmt.Errorf("ephemeral: verification secret is required")
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return ArtifactPayload{}, fmt.Errorf("ephemeral: malformed token")
	}
	expectedSig := computeSignature(parts[0], secret)
	if !hmac.Equal([]byte(parts[1]), []byte(expectedSig)) {
		return ArtifactPayload{}, fmt.Errorf("ephemeral: invalid signature")
	}
	body, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return ArtifactPayload{}, fmt.Errorf("ephemeral: malformed payload encoding")
	}
	var p ArtifactPayload
	if err := json.Unmarshal(body, &p); err != nil {
		return ArtifactPayload{}, fmt.Errorf("ephemeral: malformed payload")
	}
	if time.Now().Unix() >= p.ExpiresAt {
		return ArtifactPayload{}, fmt.Errorf("ephemeral: token expired")
	}
	return p, nil
}

func computeSignature(encodedBody, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(encodedBody))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
