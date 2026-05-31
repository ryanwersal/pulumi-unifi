package network

// Small pointer helpers for mapping between optional Pulumi inputs (pointers)
// and the value-typed fields of the go-unifi structs.

func ptr[T any](v T) *T { return &v }

func derefOr[T any](p *T, def T) T {
	if p != nil {
		return *p
	}
	return def
}
