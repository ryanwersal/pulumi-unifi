// SPDX-License-Identifier: Apache-2.0

// Package driveapi is a minimal client for the PRIVATE UniFi Drive HTTP API
// served by a UNAS appliance (e.g. UNAS Pro) under /proxy/drive/api. UniFi Drive
// has no official, documented API; this rides the local, reverse-engineered
// surface the Drive web UI itself uses.
//
// The wire protocol, session handling, and NFS read-modify-write logic here are
// adapted from iperka/unifi-drive-storage-provider (Apache-2.0), which verified
// them against a UNAS Pro running UniFi Drive 4.2.6 / firmware 5.1.15. Two API
// styles coexist:
//
//   - v1 endpoints (/proxy/drive/api/v1/...) wrap the payload in an envelope
//     {"err":..., "type":"single|collection", "data":...}.
//   - v2 read endpoints (/proxy/drive/api/v2/...) return bare JSON.
//
// Unlike the Network and Protect applications, Drive runs on the UNAS appliance
// itself — a SEPARATE UniFi OS console from a UDM gateway — so this client is
// built against the UNAS host directly, not the primary controller. It does its
// own cookie+CSRF session login (the go-unifi client can't be reused: its
// constructor eagerly reads Network-app system info, which a NAS-only console
// does not serve).
package driveapi

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"
	"time"
)

// ErrShareNotFound is returned by share lookups and deletes when no share
// matches. Callers rely on it for idempotency (Delete of an already-gone share
// succeeds) and drift detection (Read returns an empty response).
var ErrShareNotFound = errors.New("driveapi: share not found")

// ErrExportNotFound is returned by NFS export lookups when no grant matches the
// (share, client) pair.
var ErrExportNotFound = errors.New("driveapi: nfs export not found")

// Share is a UniFi Drive shared drive.
type Share struct {
	// ID is the appliance-assigned identifier used for subsequent API calls.
	ID string
	// Name is the human-readable share name.
	Name string
	// StoragePoolID is the pool the drive lives in.
	StoragePoolID string
	// QuotaGiB is the size limit in gibibytes; -1 means unlimited.
	QuotaGiB int64
	// ExportPath is the documented NFS export path for the drive.
	ExportPath string
}

// ShareSpec describes a shared drive to create.
type ShareSpec struct {
	// Name is the share name (must be unique on the appliance).
	Name string
	// StoragePoolID pins the pool. Empty means the appliance's first pool.
	StoragePoolID string
	// QuotaGiB is the size limit in gibibytes; <= 0 means unlimited.
	QuotaGiB int64
}

// StoragePool is a UNAS storage pool.
type StoragePool struct {
	ID     string
	Number int
	Status string
}

// NFSExport is a single (share, client) NFS access grant.
type NFSExport struct {
	ShareID   string
	ShareName string
	// Client is a client IP or CIDR the appliance stores as one connection.
	Client string
	// Permission is "rw" or "ro".
	Permission string
}

// Client is the boundary to the UniFi Drive appliance. Implementations are safe
// for concurrent use.
type Client interface {
	// ListShares returns every shared drive.
	ListShares(ctx context.Context) ([]Share, error)
	// GetShareByID returns the share with the given ID, or ErrShareNotFound.
	GetShareByID(ctx context.Context, id string) (*Share, error)
	// CreateShare creates a shared drive. It errors if a share with the same
	// name already exists (share names are unique; this provider does not adopt
	// pre-existing shares).
	CreateShare(ctx context.Context, spec ShareSpec) (*Share, error)
	// DeleteShare removes the share with the given ID. An already-absent share
	// is not an error.
	DeleteShare(ctx context.Context, id string) error
	// ListStoragePools returns the appliance's storage pools.
	ListStoragePools(ctx context.Context) ([]StoragePool, error)

	// GetNFSExport returns the grant for (shareID, client), or ErrExportNotFound.
	GetNFSExport(ctx context.Context, shareID, client string) (*NFSExport, error)
	// EnsureNFSExport adds or refreshes an NFS grant for shareID to client at the
	// given permission ("rw"/"ro"). It is a read-modify-write of the global NFS
	// export settings, serialised per-client-process so concurrent calls do not
	// clobber each other.
	EnsureNFSExport(ctx context.Context, shareID, client, permission string) error
	// RemoveNFSExport removes the (shareID, client) grant. An already-absent
	// grant is not an error.
	RemoveNFSExport(ctx context.Context, shareID, client string) error
	// NFSServiceEnabled reports whether the appliance's global NFS service is on.
	// Exports are configured regardless, but are unreachable until it is enabled.
	NFSServiceEnabled(ctx context.Context) (bool, error)
}

// Config holds the connection settings for a Client.
type Config struct {
	// Host is the UNAS appliance address. A scheme is optional; https is assumed.
	Host string
	// Username / Password authenticate a LOCAL UniFi OS admin account on the
	// appliance (cloud/SSO accounts do not work for the local API).
	Username string
	Password string
	// InsecureSkipVerify disables TLS verification (UNAS ships a self-signed cert).
	InsecureSkipVerify bool
	// Timeout bounds each HTTP request. Zero uses defaultTimeout.
	Timeout time.Duration
	// StoragePoolID pins which pool new drives are created in when a Share does
	// not specify one. Empty means the appliance's first pool.
	StoragePoolID string
	// MaxRetries is how many times a request is retried after HTTP 429 (UniFi OS
	// rate-limits /api/auth/login). Zero uses defaultMaxRetries.
	MaxRetries int
}

// httpClient talks to the reverse-engineered UniFi Drive local API. It manages a
// cookie-based session (TOKEN cookie) plus the X-Csrf-Token required on writes.
type httpClient struct {
	cfg  Config
	base string
	http *http.Client

	mu       sync.Mutex
	csrf     string
	loggedIn bool

	poolOnce sync.Once
	poolID   string
	poolErr  error

	// nfsMu serialises read-modify-write updates to the global NFS export
	// settings so concurrent EnsureNFSExport/RemoveNFSExport calls (Pulumi runs
	// resource ops as goroutines in one process) don't clobber each other.
	nfsMu sync.Mutex

	maxRetries  int
	backoffBase time.Duration
}

const (
	defaultTimeout     = 30 * time.Second
	defaultMaxRetries  = 4
	defaultBackoffBase = 500 * time.Millisecond
	maxBackoff         = 10 * time.Second
)

// errRateLimited is the sentinel for an HTTP 429, so the retry loop can tell
// throttling apart from other failures.
var errRateLimited = errors.New("driveapi: rate limited (HTTP 429)")

func isRateLimited(err error) bool { return errors.Is(err, errRateLimited) }

// API paths. v1 endpoints are enveloped ({err,type,data}); v2 reads are not.
const (
	apiV1Shared      = "/proxy/drive/api/v1/shared"
	apiV2Storage     = "/proxy/drive/api/v2/storage"
	apiV1NFSSettings = "/proxy/drive/api/v1/services/nfs/settings"
	apiV1NFSAdvanced = "/proxy/drive/api/v1/services/nfs/advanced-settings"
	apiV1BatchOp     = "/proxy/drive/api/v1/systems/storage/shared/batch-operation"
)

// New builds a Client from cfg. It does not contact the appliance; the first API
// call performs the login.
func New(cfg Config) Client { return newHTTPClient(cfg) }

func newHTTPClient(cfg Config) *httpClient {
	host := strings.TrimRight(cfg.Host, "/")
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "https://" + host
	}
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}
	maxRetries := cfg.MaxRetries
	if maxRetries == 0 {
		maxRetries = defaultMaxRetries
	}
	jar, _ := cookiejar.New(nil)
	return &httpClient{
		cfg:         cfg,
		base:        host,
		maxRetries:  maxRetries,
		backoffBase: defaultBackoffBase,
		http: &http.Client{
			Timeout: timeout,
			Jar:     jar,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.InsecureSkipVerify}, //nolint:gosec // opt-in for self-signed UNAS certs
			},
		},
	}
}

// --- auth ---

func (c *httpClient) login(ctx context.Context) error {
	// Warm up CSRF (best-effort; some firmware only sets it on login).
	if req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+"/api/auth/csrf", nil); err == nil {
		if resp, err := c.http.Do(req); err == nil {
			c.captureCSRF(resp)
			var body csrfResponse
			if json.NewDecoder(resp.Body).Decode(&body) == nil && body.CSRFToken != "" {
				c.csrf = body.CSRFToken
			}
			_ = resp.Body.Close()
		}
	}

	payload, err := json.Marshal(loginRequest{Username: c.cfg.Username, Password: c.cfg.Password}) //nolint:gosec // login inherently sends the password to the appliance
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/api/auth/login", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.csrf != "" {
		req.Header.Set("X-Csrf-Token", c.csrf)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("driveapi login: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	c.captureCSRF(resp)
	switch {
	case resp.StatusCode == http.StatusTooManyRequests:
		return fmt.Errorf("driveapi login: %w", errRateLimited)
	case resp.StatusCode == http.StatusUnauthorized, resp.StatusCode == http.StatusForbidden:
		return fmt.Errorf("driveapi login: authentication failed (HTTP %d); use a LOCAL UniFi OS admin account on the UNAS appliance, not a cloud/SSO login", resp.StatusCode)
	case resp.StatusCode >= 400:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("driveapi login: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	c.loggedIn = true
	return nil
}

func (c *httpClient) captureCSRF(resp *http.Response) {
	for _, h := range []string{"X-Csrf-Token", "X-Updated-Csrf-Token"} {
		if v := resp.Header.Get(h); v != "" {
			c.csrf = v
		}
	}
}

func (c *httpClient) ensureLogin(ctx context.Context) error {
	if c.loggedIn {
		return nil
	}
	return c.login(ctx)
}

// --- low-level request plumbing ---

// doJSON issues a single request and returns the status and raw body. Caller
// holds c.mu (for CSRF access/update).
func (c *httpClient) doJSON(ctx context.Context, method, urlPath string, body any) (int, []byte, error) {
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return 0, nil, err
		}
		reader = bytes.NewReader(buf)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base+urlPath, reader)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.csrf != "" {
		req.Header.Set("X-Csrf-Token", c.csrf)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	c.captureCSRF(resp)
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, respBody, nil
}

// request performs an authenticated request, transparently handling two
// transient conditions: a single re-login on 401, and exponential backoff +
// retry on 429. Returns the status and raw body.
func (c *httpClient) request(ctx context.Context, method, urlPath string, body any) (int, []byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	reauthed := false
	for attempt := 0; ; attempt++ {
		if err := c.ensureLogin(ctx); err != nil {
			if isRateLimited(err) && attempt < c.maxRetries {
				if werr := c.backoff(ctx, attempt); werr != nil {
					return 0, nil, werr
				}
				continue
			}
			return 0, nil, err
		}

		status, respBody, err := c.doJSON(ctx, method, urlPath, body)
		if err != nil {
			return status, respBody, err
		}

		switch {
		case status == http.StatusUnauthorized && !reauthed:
			// Session expired — re-login once and retry the call.
			reauthed = true
			c.loggedIn = false
			continue
		case status == http.StatusTooManyRequests && attempt < c.maxRetries:
			if werr := c.backoff(ctx, attempt); werr != nil {
				return status, respBody, werr
			}
			continue
		}
		return status, respBody, nil
	}
}

// backoff sleeps for an exponentially increasing delay (capped), respecting ctx
// cancellation. attempt is zero-based.
func (c *httpClient) backoff(ctx context.Context, attempt int) error {
	d := c.backoffBase << attempt
	if d > maxBackoff || d <= 0 {
		d = maxBackoff
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// callV1 performs an enveloped v1 request and unmarshals the inner data into out
// (when non-nil), mapping the envelope error and HTTP status to a Go error.
func (c *httpClient) callV1(ctx context.Context, method, urlPath string, body, out any) error {
	status, raw, err := c.request(ctx, method, urlPath, body)
	if err != nil {
		return fmt.Errorf("driveapi: %s %s: %w", method, urlPath, err)
	}
	var env envelope
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &env); err != nil {
			// Not an envelope; treat non-2xx as error.
			if status >= 400 {
				return fmt.Errorf("driveapi: %s %s: HTTP %d: %s", method, urlPath, status, strings.TrimSpace(string(raw)))
			}
			return fmt.Errorf("driveapi: %s %s: decode: %w", method, urlPath, err)
		}
	}
	if env.Err != nil {
		return fmt.Errorf("driveapi: %s %s: %s", method, urlPath, env.Err.Error())
	}
	if status >= 400 {
		return fmt.Errorf("driveapi: %s %s: HTTP %d", method, urlPath, status)
	}
	if out != nil && len(env.Data) > 0 {
		if err := json.Unmarshal(env.Data, out); err != nil {
			return fmt.Errorf("driveapi: %s %s: decode data: %w", method, urlPath, err)
		}
	}
	return nil
}

// getV2 performs a non-enveloped v2 GET and unmarshals into out.
func (c *httpClient) getV2(ctx context.Context, urlPath string, out any) error {
	status, raw, err := c.request(ctx, http.MethodGet, urlPath, nil)
	if err != nil {
		return fmt.Errorf("driveapi: GET %s: %w", urlPath, err)
	}
	if status >= 400 {
		return fmt.Errorf("driveapi: GET %s: HTTP %d: %s", urlPath, status, strings.TrimSpace(string(raw)))
	}
	if out != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("driveapi: GET %s: decode: %w", urlPath, err)
		}
	}
	return nil
}

// compile-time assertion
var _ Client = (*httpClient)(nil)
