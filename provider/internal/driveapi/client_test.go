// SPDX-License-Identifier: Apache-2.0

package driveapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// fakeUNAS is a stateful stand-in for the UniFi Drive API surface this client
// uses: login, storage pools, the enveloped shared-drive collection, the
// batch-operation delete, and the global NFS advanced-settings.
type fakeUNAS struct {
	mu       sync.Mutex
	drives   []apiDrive
	nfs      nfsAdvancedSettings
	nfsOn    bool
	seq      int
	srv      *httptest.Server
	loginHit int32
}

func newFakeUNAS(t *testing.T) *fakeUNAS {
	t.Helper()
	f := &fakeUNAS{nfsOn: true}
	mux := http.NewServeMux()

	mux.HandleFunc("/api/auth/csrf", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Csrf-Token", "csrf-1")
		_ = json.NewEncoder(w).Encode(csrfResponse{CSRFToken: "csrf-1"})
	})
	mux.HandleFunc("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&f.loginHit, 1)
		if r.Header.Get("X-Csrf-Token") == "" {
			t.Error("login missing CSRF header")
		}
		http.SetCookie(w, &http.Cookie{Name: "TOKEN", Value: "tok"})
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/proxy/drive/api/v2/storage", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(storageResponse{Pools: []storagePool{
			{ID: "pool-1", Number: 1, Status: "ok"},
			{ID: "pool-2", Number: 2, Status: "ok"},
		}})
	})
	mux.HandleFunc("/proxy/drive/api/v1/shared", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		switch r.Method {
		case http.MethodGet:
			writeEnvelope(w, "collection", f.drives)
		case http.MethodPost:
			var req createDriveRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			f.seq++
			d := apiDrive{ID: "d" + strconv.Itoa(f.seq), Name: req.Name, StoragePoolID: req.StoragePoolID, Quota: req.Quota, Status: "active"}
			f.drives = append(f.drives, d)
			writeEnvelope(w, "single", d)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/proxy/drive/api/v1/systems/storage/shared/batch-operation", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		var req batchOperationRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.Action == "delete" {
			kept := f.drives[:0]
			for _, d := range f.drives {
				drop := false
				for _, n := range req.Names {
					if d.Name == n {
						drop = true
					}
				}
				if !drop {
					kept = append(kept, d)
				}
			}
			f.drives = kept
		}
		writeEnvelope(w, "single", "OK")
	})
	mux.HandleFunc("/proxy/drive/api/v1/services/nfs/settings", func(w http.ResponseWriter, _ *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		writeEnvelope(w, "single", nfsSettings{Enable: f.nfsOn})
	})
	mux.HandleFunc("/proxy/drive/api/v1/services/nfs/advanced-settings", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		switch r.Method {
		case http.MethodGet:
			writeEnvelope(w, "single", f.nfs)
		case http.MethodPut:
			var s nfsAdvancedSettings
			_ = json.NewDecoder(r.Body).Decode(&s)
			f.nfs = s
			writeEnvelope(w, "single", "OK")
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	f.srv = httptest.NewServer(mux)
	t.Cleanup(f.srv.Close)
	return f
}

func writeEnvelope(w http.ResponseWriter, typ string, data any) {
	raw, _ := json.Marshal(data)
	_ = json.NewEncoder(w).Encode(envelope{Type: typ, Data: raw})
}

func newTestClient(t *testing.T, f *fakeUNAS) *httpClient {
	t.Helper()
	c := newHTTPClient(Config{Host: f.srv.URL, Username: "u", Password: "p"})
	c.backoffBase = time.Millisecond
	return c
}

func TestShareLifecycle(t *testing.T) {
	f := newFakeUNAS(t)
	c := newTestClient(t, f)
	ctx := context.Background()

	if _, err := c.GetShareByID(ctx, "nope"); err != ErrShareNotFound {
		t.Fatalf("want ErrShareNotFound, got %v", err)
	}

	share, err := c.CreateShare(ctx, ShareSpec{Name: "media", QuotaGiB: 100})
	if err != nil {
		t.Fatalf("CreateShare: %v", err)
	}
	if share.ID == "" || share.Name != "media" || share.QuotaGiB != 100 {
		t.Fatalf("unexpected share %+v", share)
	}
	if share.StoragePoolID != "pool-1" {
		t.Errorf("pool = %q, want pool-1 (first pool default)", share.StoragePoolID)
	}
	if share.ExportPath != "/var/nfs/shared/media" {
		t.Errorf("export = %q", share.ExportPath)
	}

	// Duplicate name errors (no adoption).
	if _, err := c.CreateShare(ctx, ShareSpec{Name: "media"}); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("duplicate create should error, got %v", err)
	}

	got, err := c.GetShareByID(ctx, share.ID)
	if err != nil || got.Name != "media" {
		t.Fatalf("GetShareByID: %+v %v", got, err)
	}

	if err := c.DeleteShare(ctx, share.ID); err != nil {
		t.Fatalf("DeleteShare: %v", err)
	}
	if _, err := c.GetShareByID(ctx, share.ID); err != ErrShareNotFound {
		t.Errorf("share should be gone, got %v", err)
	}
	// Deleting an already-gone share is a no-op.
	if err := c.DeleteShare(ctx, share.ID); err != nil {
		t.Errorf("delete of absent share should succeed, got %v", err)
	}
}

func TestCreateShareUnlimitedQuota(t *testing.T) {
	f := newFakeUNAS(t)
	c := newTestClient(t, f)
	share, err := c.CreateShare(context.Background(), ShareSpec{Name: "vault", QuotaGiB: 0})
	if err != nil {
		t.Fatalf("CreateShare: %v", err)
	}
	if share.QuotaGiB != -1 {
		t.Errorf("quota = %d, want -1 (unlimited)", share.QuotaGiB)
	}
}

func TestCreateSharePinnedPool(t *testing.T) {
	f := newFakeUNAS(t)
	c := newTestClient(t, f)
	share, err := c.CreateShare(context.Background(), ShareSpec{Name: "x", StoragePoolID: "pool-2"})
	if err != nil {
		t.Fatalf("CreateShare: %v", err)
	}
	if share.StoragePoolID != "pool-2" {
		t.Errorf("pool = %q, want pool-2", share.StoragePoolID)
	}
}

func TestListStoragePools(t *testing.T) {
	f := newFakeUNAS(t)
	c := newTestClient(t, f)
	pools, err := c.ListStoragePools(context.Background())
	if err != nil {
		t.Fatalf("ListStoragePools: %v", err)
	}
	if len(pools) != 2 || pools[0].ID != "pool-1" {
		t.Errorf("unexpected pools %+v", pools)
	}
}

func TestNFSExportRoundTrip(t *testing.T) {
	f := newFakeUNAS(t)
	c := newTestClient(t, f)
	ctx := context.Background()

	share, err := c.CreateShare(ctx, ShareSpec{Name: "media"})
	if err != nil {
		t.Fatalf("CreateShare: %v", err)
	}

	if _, err := c.GetNFSExport(ctx, share.ID, "10.0.0.5"); err != ErrExportNotFound {
		t.Fatalf("want ErrExportNotFound, got %v", err)
	}

	if err := c.EnsureNFSExport(ctx, share.ID, "10.0.0.5", "rw"); err != nil {
		t.Fatalf("EnsureNFSExport: %v", err)
	}
	exp, err := c.GetNFSExport(ctx, share.ID, "10.0.0.5")
	if err != nil {
		t.Fatalf("GetNFSExport: %v", err)
	}
	if exp.Permission != "rw" || exp.ShareName != "media" || exp.Client != "10.0.0.5" {
		t.Errorf("unexpected export %+v", exp)
	}

	// Update permission via re-ensure (RMW refresh).
	if err := c.EnsureNFSExport(ctx, share.ID, "10.0.0.5", "ro"); err != nil {
		t.Fatalf("re-EnsureNFSExport: %v", err)
	}
	exp, _ = c.GetNFSExport(ctx, share.ID, "10.0.0.5")
	if exp.Permission != "ro" {
		t.Errorf("permission = %q, want ro", exp.Permission)
	}
	// A second client on the same drive is a distinct connection.
	if err := c.EnsureNFSExport(ctx, share.ID, "10.0.0.6", "rw"); err != nil {
		t.Fatalf("second client: %v", err)
	}
	if len(f.nfs.Connections) != 2 {
		t.Errorf("expected 2 connections, got %d", len(f.nfs.Connections))
	}

	if err := c.RemoveNFSExport(ctx, share.ID, "10.0.0.5"); err != nil {
		t.Fatalf("RemoveNFSExport: %v", err)
	}
	if _, err := c.GetNFSExport(ctx, share.ID, "10.0.0.5"); err != ErrExportNotFound {
		t.Errorf("export should be gone, got %v", err)
	}
	// Removing an absent grant is a no-op.
	if err := c.RemoveNFSExport(ctx, share.ID, "10.0.0.5"); err != nil {
		t.Errorf("remove of absent grant should succeed, got %v", err)
	}
}

func TestEnsureNFSExportUnknownShare(t *testing.T) {
	f := newFakeUNAS(t)
	c := newTestClient(t, f)
	if err := c.EnsureNFSExport(context.Background(), "ghost", "10.0.0.5", "rw"); err != ErrShareNotFound {
		t.Fatalf("want ErrShareNotFound, got %v", err)
	}
}

func TestDeleteShareStripsNFS(t *testing.T) {
	f := newFakeUNAS(t)
	c := newTestClient(t, f)
	ctx := context.Background()
	share, _ := c.CreateShare(ctx, ShareSpec{Name: "media"})
	if err := c.EnsureNFSExport(ctx, share.ID, "10.0.0.5", "rw"); err != nil {
		t.Fatalf("EnsureNFSExport: %v", err)
	}
	if err := c.DeleteShare(ctx, share.ID); err != nil {
		t.Fatalf("DeleteShare: %v", err)
	}
	if len(f.nfs.Connections) != 0 {
		t.Errorf("deleting the share should strip its NFS grants, got %+v", f.nfs.Connections)
	}
}

func TestNFSServiceEnabled(t *testing.T) {
	f := newFakeUNAS(t)
	c := newTestClient(t, f)
	on, err := c.NFSServiceEnabled(context.Background())
	if err != nil || !on {
		t.Fatalf("want enabled, got %v %v", on, err)
	}
	f.mu.Lock()
	f.nfsOn = false
	f.mu.Unlock()
	on, _ = c.NFSServiceEnabled(context.Background())
	if on {
		t.Error("want disabled")
	}
}

func TestEnvelopeErrorSurfaced(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/login", func(w http.ResponseWriter, _ *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "TOKEN", Value: "tok"})
	})
	mux.HandleFunc("/proxy/drive/api/v1/shared", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(envelope{Err: &apiError{Msg: "boom", Code: "ErrInternal"}})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := newHTTPClient(Config{Host: srv.URL, Username: "u", Password: "p"})
	_, err := c.ListShares(context.Background())
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("want enveloped error surfaced, got %v", err)
	}
}

func TestRetriesOn429ThenSucceeds(t *testing.T) {
	var loginHits int32
	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/login", func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&loginHits, 1) <= 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		http.SetCookie(w, &http.Cookie{Name: "TOKEN", Value: "tok"})
	})
	mux.HandleFunc("/proxy/drive/api/v1/shared", func(w http.ResponseWriter, _ *http.Request) {
		writeEnvelope(w, "collection", []apiDrive{})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := newHTTPClient(Config{Host: srv.URL, Username: "u", Password: "p"})
	c.backoffBase = time.Millisecond
	if _, err := c.ListShares(context.Background()); err != nil {
		t.Fatalf("want success after retrying through 429s, got %v", err)
	}
	if loginHits < 3 {
		t.Errorf("expected >=3 login attempts, got %d", loginHits)
	}
}

func TestHostSchemeDefaulting(t *testing.T) {
	c := newHTTPClient(Config{Host: "10.0.0.10"})
	if !strings.HasPrefix(c.base, "https://") {
		t.Errorf("base = %q, want https scheme", c.base)
	}
}

func TestAddDriveToClientAndRemove(t *testing.T) {
	s := &nfsAdvancedSettings{}
	acl := nfsDriveAcl{ID: "d1", Name: "n", Permission: "rw"}
	s.addDriveToClient("10.0.0.1", acl)
	s.addDriveToClient("10.0.0.1", acl) // idempotent
	if len(s.Connections) != 1 || len(s.Connections[0].SharedDrives) != 1 {
		t.Fatalf("expected 1 connection/1 drive, got %+v", s.Connections)
	}
	if !s.removeDriveForClient("d1", "10.0.0.1") {
		t.Fatal("expected change")
	}
	if len(s.Connections) != 0 {
		t.Fatalf("connection should be dropped when empty, got %+v", s.Connections)
	}
	if s.removeDriveForClient("d1", "10.0.0.1") {
		t.Error("removing absent grant should report no change")
	}
}
