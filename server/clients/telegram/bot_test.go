// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package telegram

import (
	"errors"
	"fmt"
	"testing"

	"github.com/mymmrac/telego/telegoapi"
)

// isParseEntitiesError must match the error telego actually returns: the
// description carries Telegram's category prefix ("Bad Request: ...") and the
// *telegoapi.Error is wrapped with "api: %w".
func TestIsParseEntitiesError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "real telegram error with category prefix",
			err: fmt.Errorf("api: %w", &telegoapi.Error{
				ErrorCode:   400,
				Description: "Bad Request: can't parse entities: Can't find end of the entity starting at byte offset 194",
			}),
			want: true,
		},
		{
			name: "bare parse entities description without prefix",
			err: fmt.Errorf("api: %w", &telegoapi.Error{
				ErrorCode:   400,
				Description: "can't parse entities: ...",
			}),
			want: true,
		},
		{
			name: "other 400 error",
			err: fmt.Errorf("api: %w", &telegoapi.Error{
				ErrorCode:   400,
				Description: "Bad Request: message is too long",
			}),
			want: false,
		},
		{
			name: "non-400 parse entities error",
			err: fmt.Errorf("api: %w", &telegoapi.Error{
				ErrorCode:   500,
				Description: "can't parse entities: ...",
			}),
			want: false,
		},
		{
			name: "unrelated error",
			err:  errors.New("network unreachable"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isParseEntitiesError(tc.err); got != tc.want {
				t.Errorf("isParseEntitiesError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}
