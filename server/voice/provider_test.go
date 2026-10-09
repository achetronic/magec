// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package voice

import (
	"context"
	"testing"
	"time"
)

func TestWithDefaultTimeoutAppliesDefault(t *testing.T) {
	ctx, cancel := WithDefaultTimeout(context.Background(), time.Minute)
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected a deadline")
	}
	if left := time.Until(deadline); left > time.Minute || left < 59*time.Second {
		t.Fatalf("expected ~1m deadline, got %v", left)
	}
}

func TestWithDefaultTimeoutKeepsCallerDeadline(t *testing.T) {
	parent, cancelParent := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancelParent()

	ctx, cancel := WithDefaultTimeout(parent, time.Minute)
	defer cancel()

	want, _ := parent.Deadline()
	got, _ := ctx.Deadline()
	if !got.Equal(want) {
		t.Fatalf("expected caller deadline %v to win, got %v", want, got)
	}
}
