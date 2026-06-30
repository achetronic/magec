// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package ephemeral

import (
	"strings"
	"testing"
	"time"
)

const testSecret = "test-secret-do-not-leak"

func validPayload() ArtifactPayload {
	return ArtifactPayload{
		AppName:   "agent-1",
		UserID:    "default_user",
		SessionID: "sess-1",
		Name:      "file.pdf",
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}
}

func TestSignAndVerify_RoundTrip(t *testing.T) {
	p := validPayload()
	token, err := SignArtifact(p, testSecret)
	if err != nil {
		t.Fatalf("SignArtifact: %v", err)
	}
	got, err := VerifyArtifact(token, testSecret)
	if err != nil {
		t.Fatalf("VerifyArtifact: %v", err)
	}
	if got != p {
		t.Fatalf("payload mismatch: got %+v want %+v", got, p)
	}
}

func TestSign_EmptySecret(t *testing.T) {
	if _, err := SignArtifact(validPayload(), ""); err == nil {
		t.Fatalf("expected error for empty secret")
	}
}

func TestVerify_EmptySecret(t *testing.T) {
	token, _ := SignArtifact(validPayload(), testSecret)
	if _, err := VerifyArtifact(token, ""); err == nil {
		t.Fatalf("expected error for empty secret")
	}
}

func TestVerify_WrongSecret(t *testing.T) {
	token, _ := SignArtifact(validPayload(), testSecret)
	if _, err := VerifyArtifact(token, "another-secret"); err == nil {
		t.Fatalf("expected signature mismatch")
	}
}

func TestVerify_TamperedPayload(t *testing.T) {
	token, _ := SignArtifact(validPayload(), testSecret)
	parts := strings.SplitN(token, ".", 2)
	// Replace first byte of payload.
	tampered := "X" + parts[0][1:] + "." + parts[1]
	if _, err := VerifyArtifact(tampered, testSecret); err == nil {
		t.Fatalf("expected error for tampered payload")
	}
}

func TestVerify_TamperedSignature(t *testing.T) {
	token, _ := SignArtifact(validPayload(), testSecret)
	parts := strings.SplitN(token, ".", 2)
	tampered := parts[0] + "." + "X" + parts[1][1:]
	if _, err := VerifyArtifact(tampered, testSecret); err == nil {
		t.Fatalf("expected error for tampered signature")
	}
}

func TestVerify_Malformed(t *testing.T) {
	cases := []string{
		"",
		"notoken",
		"only.one",
		"a.b.c",
		".sig",
		"payload.",
	}
	for _, c := range cases {
		if _, err := VerifyArtifact(c, testSecret); err == nil {
			t.Fatalf("expected error for %q", c)
		}
	}
}

func TestVerify_Expired(t *testing.T) {
	p := validPayload()
	p.ExpiresAt = time.Now().Add(-1 * time.Second).Unix()
	token, _ := SignArtifact(p, testSecret)
	_, err := VerifyArtifact(token, testSecret)
	if err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expected expired error, got %v", err)
	}
}

func TestVerify_NotYetExpiredAtBoundary(t *testing.T) {
	p := validPayload()
	// 1 second from now: must still verify.
	p.ExpiresAt = time.Now().Add(1 * time.Second).Unix()
	token, _ := SignArtifact(p, testSecret)
	if _, err := VerifyArtifact(token, testSecret); err != nil {
		t.Fatalf("token should still be valid 1s before expiry: %v", err)
	}
}
