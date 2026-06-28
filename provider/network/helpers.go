package network

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/filipowm/go-unifi/unifi"
)

// Small pointer helpers for mapping between optional Pulumi inputs (pointers)
// and the value-typed fields of the go-unifi structs.

func ptr[T any](v T) *T { return &v }

func derefOr[T any](p *T, def T) T {
	if p != nil {
		return *p
	}
	return def
}

// notFound reports whether err is go-unifi's ErrNotFound or an HTTP 404, so a
// Read can return an empty ReadResponse instead of erroring.
func notFound(err error) bool {
	if errors.Is(err, unifi.ErrNotFound) {
		return true
	}
	var se *unifi.ServerError
	if errors.As(err, &se) {
		return se.StatusCode == http.StatusNotFound
	}
	return false
}

// wrap annotates a controller error with caller context (op should read like
// `create vlan "lab" (site "default")`). Returns nil when err is nil.
func wrap(op string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", op, err)
}
