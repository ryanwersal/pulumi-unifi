// SPDX-License-Identifier: Apache-2.0

package protect

import (
	"fmt"
	"strings"
)

// Small pointer helpers for mapping between optional Pulumi inputs (pointers)
// and value-typed wire-format fields.

func ptr[T any](v T) *T { return &v }

func derefOr[T any](p *T, def T) T {
	if p != nil {
		return *p
	}
	return def
}

// isProtectNotFound reports a 404 from the unified client, which exposes no
// typed not-found sentinel and only formats the status into the error text.
func isProtectNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "http code 404")
}

// wrap annotates a Protect error with caller context (op should read like
// `patch camera "abc123"`). Returns nil when err is nil.
func wrap(op string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", op, err)
}
