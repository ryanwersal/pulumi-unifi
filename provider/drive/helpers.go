// SPDX-License-Identifier: Apache-2.0

// Package drive contains the UniFi Drive resources (shared drives and their NFS
// exports) managed on a UNAS appliance via its private /proxy/drive API. See
// provider/internal/driveapi.
package drive

// Small pointer helpers for mapping between optional Pulumi inputs (pointers)
// and value-typed wire fields.

func ptr[T any](v T) *T { return &v }

func derefOr[T any](p *T, def T) T {
	if p != nil {
		return *p
	}
	return def
}
