// SPDX-License-Identifier: Apache-2.0

package drive

import (
	"testing"

	"github.com/ryanwersal/pulumi-unifi/provider/internal/driveapi"
)

func TestParseExportID(t *testing.T) {
	cases := []struct {
		id      string
		shareID string
		client  string
		ok      bool
	}{
		{"d-1/10.0.0.5", "d-1", "10.0.0.5", true},
		// CIDR client contains a slash: split on the FIRST slash only.
		{"d-1/10.0.0.0/24", "d-1", "10.0.0.0/24", true},
		{"abc123/192.168.1.0/24", "abc123", "192.168.1.0/24", true},
		{"noslash", "", "", false},
		{"/leadingempty", "", "", false},
		{"trailingempty/", "", "", false},
	}
	for _, c := range cases {
		shareID, client, ok := parseExportID(c.id)
		if ok != c.ok || shareID != c.shareID || client != c.client {
			t.Errorf("parseExportID(%q) = (%q, %q, %v), want (%q, %q, %v)", c.id, shareID, client, ok, c.shareID, c.client, c.ok)
		}
	}
	// exportID round-trips through parseExportID, even with a CIDR client.
	if got := exportID("d-1", "10.0.0.0/24"); got != "d-1/10.0.0.0/24" {
		t.Errorf("exportID = %q", got)
	}
}

func TestShareStateFromPreservesOptionals(t *testing.T) {
	s := &driveapi.Share{ID: "d1", Name: "media", StoragePoolID: "pool-1", QuotaGiB: -1, ExportPath: "/var/nfs/shared/media"}

	// User omitted pool and quota: inputs stay nil (no spurious diff), while the
	// appliance's actual pool/export surface as outputs.
	st := shareStateFrom(s, ShareArgs{Name: "media"})
	if st.StoragePoolId != nil || st.QuotaGib != nil {
		t.Errorf("omitted optionals should stay nil, got pool=%v quota=%v", st.StoragePoolId, st.QuotaGib)
	}
	if st.ShareId != "d1" || st.PoolId != "pool-1" || st.ExportPath != "/var/nfs/shared/media" {
		t.Errorf("unexpected outputs %+v", st)
	}

	// User set pool and quota: echoed back exactly.
	st = shareStateFrom(s, ShareArgs{Name: "media", StoragePoolId: ptr("pool-1"), QuotaGib: ptr(100)})
	if st.StoragePoolId == nil || *st.StoragePoolId != "pool-1" || st.QuotaGib == nil || *st.QuotaGib != 100 {
		t.Errorf("set optionals should echo, got pool=%v quota=%v", st.StoragePoolId, st.QuotaGib)
	}
}

func TestShareArgsToSpec(t *testing.T) {
	spec := ShareArgs{Name: "x", StoragePoolId: ptr("p2"), QuotaGib: ptr(50)}.toSpec()
	if spec.Name != "x" || spec.StoragePoolID != "p2" || spec.QuotaGiB != 50 {
		t.Errorf("unexpected spec %+v", spec)
	}
	// Omitted optionals become the appliance defaults (empty pool => first pool,
	// 0 => unlimited, both handled downstream).
	spec = ShareArgs{Name: "y"}.toSpec()
	if spec.StoragePoolID != "" || spec.QuotaGiB != 0 {
		t.Errorf("omitted optionals should be zero-valued, got %+v", spec)
	}
}

func TestNfsExportPermissionDefault(t *testing.T) {
	if got := (NfsExportArgs{}).permission(); got != "rw" {
		t.Errorf("default permission = %q, want rw", got)
	}
	if got := (NfsExportArgs{Permission: ptr(NfsPermissionRo)}).permission(); got != "ro" {
		t.Errorf("permission = %q, want ro", got)
	}
}
