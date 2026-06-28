package protect

import "strings"

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
