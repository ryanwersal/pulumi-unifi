// SPDX-License-Identifier: Apache-2.0

package driveapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// --- v1 envelope ---

// envelope is the v1 wrapper: {"err":..., "type":"single|collection", "data":...}.
type envelope struct {
	Err  *apiError       `json:"err"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type apiError struct {
	Msg      string          `json:"msg"`
	Code     string          `json:"code"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

func (e *apiError) Error() string {
	if e == nil {
		return ""
	}
	if e.Code != "" {
		return e.Msg + " (" + e.Code + ")"
	}
	return e.Msg
}

// --- auth wire types ---

type csrfResponse struct {
	CSRFToken string `json:"csrfToken"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// --- storage pools (GET /proxy/drive/api/v2/storage) ---

type storageResponse struct {
	Pools []storagePool `json:"pools"`
}

type storagePool struct {
	ID     string `json:"id"`
	Number int    `json:"number"`
	Status string `json:"status"`
}

// --- shared drives (GET/POST /proxy/drive/api/v1/shared) ---

type apiDrive struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	StoragePoolID string `json:"storagePoolId"`
	// Quota is the size limit in GB; -1 means unlimited.
	Quota  int64  `json:"quota"`
	Usage  int64  `json:"usage"`
	Status string `json:"status"`
}

// createDriveRequest is the body of POST /proxy/drive/api/v1/shared. Empty
// members/groups and security "none" mirror what the UI sends for a plain shared
// drive (no per-user ACLs — access is governed by the NFS export).
type createDriveRequest struct {
	Name          string   `json:"name"`
	StoragePoolID string   `json:"storagePoolId"`
	Quota         int64    `json:"quota"`
	Members       []string `json:"members"`
	Groups        []string `json:"groups"`
	Security      string   `json:"security"`
}

// batchOperationRequest is the body of the batch-operation endpoint. Shared
// drives are deleted by NAME via action "delete" (the UI first "deactivate"s
// them).
type batchOperationRequest struct {
	Action string   `json:"action"`
	Names  []string `json:"names"`
}

// DefaultExportBase is the documented NFS export base path for a drive:
//
//	mount -t nfs <host>:/var/nfs/shared/<Drive Name> /mnt
const DefaultExportBase = "/var/nfs/shared"

func nfsExportPath(name string) string { return DefaultExportBase + "/" + name }

func toShare(d apiDrive) *Share {
	return &Share{
		ID:            d.ID,
		Name:          d.Name,
		StoragePoolID: d.StoragePoolID,
		QuotaGiB:      d.Quota,
		ExportPath:    nfsExportPath(d.Name),
	}
}

// normalizeQuota converts a requested GiB quota into the appliance's wire value:
// <= 0 becomes -1 (unlimited).
func normalizeQuota(gib int64) int64 {
	if gib <= 0 {
		return -1
	}
	return gib
}

// listDrives fetches the raw shared-drive collection.
func (c *httpClient) listDrives(ctx context.Context) ([]apiDrive, error) {
	var drives []apiDrive
	if err := c.callV1(ctx, http.MethodGet, apiV1Shared, nil, &drives); err != nil {
		return nil, err
	}
	return drives, nil
}

// ListShares implements Client.
func (c *httpClient) ListShares(ctx context.Context) ([]Share, error) {
	drives, err := c.listDrives(ctx)
	if err != nil {
		return nil, err
	}
	shares := make([]Share, 0, len(drives))
	for i := range drives {
		shares = append(shares, *toShare(drives[i]))
	}
	return shares, nil
}

// GetShareByID implements Client.
func (c *httpClient) GetShareByID(ctx context.Context, id string) (*Share, error) {
	drives, err := c.listDrives(ctx)
	if err != nil {
		return nil, err
	}
	for i := range drives {
		if drives[i].ID == id {
			return toShare(drives[i]), nil
		}
	}
	return nil, ErrShareNotFound
}

// ListStoragePools implements Client.
func (c *httpClient) ListStoragePools(ctx context.Context) ([]StoragePool, error) {
	var sr storageResponse
	if err := c.getV2(ctx, apiV2Storage, &sr); err != nil {
		return nil, err
	}
	pools := make([]StoragePool, 0, len(sr.Pools))
	for _, p := range sr.Pools {
		pools = append(pools, StoragePool(p))
	}
	return pools, nil
}

// resolvePoolID returns the pool new drives are created in, caching the result.
// It honours Config.StoragePoolID, else picks the first pool reported.
func (c *httpClient) resolvePoolID(ctx context.Context) (string, error) {
	c.poolOnce.Do(func() {
		if c.cfg.StoragePoolID != "" {
			c.poolID = c.cfg.StoragePoolID
			return
		}
		var sr storageResponse
		if err := c.getV2(ctx, apiV2Storage, &sr); err != nil {
			c.poolErr = err
			return
		}
		if len(sr.Pools) == 0 {
			c.poolErr = fmt.Errorf("driveapi: no storage pools reported by appliance")
			return
		}
		c.poolID = sr.Pools[0].ID
	})
	return c.poolID, c.poolErr
}

// CreateShare implements Client. It errors if a share with the same name already
// exists rather than adopting it.
func (c *httpClient) CreateShare(ctx context.Context, spec ShareSpec) (*Share, error) {
	drives, err := c.listDrives(ctx)
	if err != nil {
		return nil, err
	}
	for i := range drives {
		if drives[i].Name == spec.Name {
			return nil, fmt.Errorf("driveapi: shared drive %q already exists", spec.Name)
		}
	}

	poolID := spec.StoragePoolID
	if poolID == "" {
		poolID, err = c.resolvePoolID(ctx)
		if err != nil {
			return nil, err
		}
	}
	req := createDriveRequest{
		Name:          spec.Name,
		StoragePoolID: poolID,
		Quota:         normalizeQuota(spec.QuotaGiB),
		Members:       []string{},
		Groups:        []string{},
		Security:      "none",
	}
	var created apiDrive
	if err := c.callV1(ctx, http.MethodPost, apiV1Shared, req, &created); err != nil {
		return nil, err
	}
	// Some firmware omits echoed fields on create; backfill from the request.
	if created.StoragePoolID == "" {
		created.StoragePoolID = poolID
	}
	if created.Name == "" {
		created.Name = spec.Name
	}
	return toShare(created), nil
}

// DeleteShare implements Client. Deletion is by NAME via the batch-operation
// endpoint (there is no per-id DELETE route), mirroring the UI's
// deactivate->delete sequence. A missing drive is treated as success.
func (c *httpClient) DeleteShare(ctx context.Context, id string) error {
	drives, err := c.listDrives(ctx)
	if err != nil {
		return err
	}
	var name string
	for i := range drives {
		if drives[i].ID == id {
			name = drives[i].Name
			break
		}
	}
	if name == "" {
		return nil // already gone
	}

	// Strip the drive from the global NFS export settings first, so deleting it
	// leaves no orphan allowlist entries behind (best-effort).
	_ = c.removeDriveFromNFS(ctx, id)

	// Deactivate first (best-effort; some firmware requires inactive before delete).
	_ = c.batchOp(ctx, "deactivate", name)
	if err := c.batchOp(ctx, "delete", name); err != nil {
		if strings.Contains(err.Error(), "record not found") {
			return nil
		}
		return err
	}
	return nil
}

// batchOp runs a shared-drive batch operation (deactivate/delete) by name.
func (c *httpClient) batchOp(ctx context.Context, action, name string) error {
	return c.callV1(ctx, http.MethodPost, apiV1BatchOp, batchOperationRequest{Action: action, Names: []string{name}}, nil)
}
