// SPDX-License-Identifier: Apache-2.0

package driveapi

import (
	"context"
	"net/http"
)

// --- NFS export config (/proxy/drive/api/v1/services/nfs/advanced-settings) ---
//
// NFS access is modelled globally as a list of connections, one per client IP,
// each listing the shared drives that client may access and at what permission.
// Granting a drive to a client is a read-modify-write of this structure.

type nfsAdvancedSettings struct {
	Connections []nfsConnection `json:"connections"`
}

type nfsConnection struct {
	Client       string        `json:"client"`
	Async        bool          `json:"async"`
	ErrorCodes   []string      `json:"errorCodes"`
	SharedDrives []nfsDriveAcl `json:"sharedDrives"`
}

type nfsDriveAcl struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Status           string `json:"status"`
	EncryptionStatus string `json:"encryptionStatus"`
	// Permission is "rw" or "ro".
	Permission string `json:"permission"`
}

// nfsSettings is GET .../services/nfs/settings -> {"enable":bool}.
type nfsSettings struct {
	Enable bool `json:"enable"`
}

// GetNFSExport implements Client.
func (c *httpClient) GetNFSExport(ctx context.Context, shareID, client string) (*NFSExport, error) {
	var settings nfsAdvancedSettings
	if err := c.callV1(ctx, http.MethodGet, apiV1NFSAdvanced, nil, &settings); err != nil {
		return nil, err
	}
	for _, conn := range settings.Connections {
		if conn.Client != client {
			continue
		}
		for _, d := range conn.SharedDrives {
			if d.ID == shareID {
				return &NFSExport{ShareID: shareID, ShareName: d.Name, Client: client, Permission: d.Permission}, nil
			}
		}
	}
	return nil, ErrExportNotFound
}

// EnsureNFSExport implements Client via a read-modify-write of the global NFS
// export settings, serialised by nfsMu.
func (c *httpClient) EnsureNFSExport(ctx context.Context, shareID, client, permission string) error {
	c.nfsMu.Lock()
	defer c.nfsMu.Unlock()

	// Resolve the drive name for the ACL entry (also validates the share exists).
	drives, err := c.listDrives(ctx)
	if err != nil {
		return err
	}
	var name string
	for i := range drives {
		if drives[i].ID == shareID {
			name = drives[i].Name
			break
		}
	}
	if name == "" {
		return ErrShareNotFound
	}

	var settings nfsAdvancedSettings
	if err := c.callV1(ctx, http.MethodGet, apiV1NFSAdvanced, nil, &settings); err != nil {
		return err
	}
	acl := nfsDriveAcl{ID: shareID, Name: name, Status: "active", EncryptionStatus: "unencrypted", Permission: permission}
	settings.addDriveToClient(client, acl)
	return c.callV1(ctx, http.MethodPut, apiV1NFSAdvanced, settings, nil)
}

// RemoveNFSExport implements Client via a read-modify-write, serialised by nfsMu.
func (c *httpClient) RemoveNFSExport(ctx context.Context, shareID, client string) error {
	c.nfsMu.Lock()
	defer c.nfsMu.Unlock()

	var settings nfsAdvancedSettings
	if err := c.callV1(ctx, http.MethodGet, apiV1NFSAdvanced, nil, &settings); err != nil {
		return err
	}
	if !settings.removeDriveForClient(shareID, client) {
		return nil // nothing to remove
	}
	return c.callV1(ctx, http.MethodPut, apiV1NFSAdvanced, settings, nil)
}

// removeDriveFromNFS strips a drive id from every client's sharedDrives (and
// drops connections left empty), then writes the result back. Serialised by
// nfsMu. A no-op write is skipped. Used when a share is deleted.
func (c *httpClient) removeDriveFromNFS(ctx context.Context, driveID string) error {
	c.nfsMu.Lock()
	defer c.nfsMu.Unlock()

	var settings nfsAdvancedSettings
	if err := c.callV1(ctx, http.MethodGet, apiV1NFSAdvanced, nil, &settings); err != nil {
		return err
	}
	if !settings.removeDrive(driveID) {
		return nil
	}
	return c.callV1(ctx, http.MethodPut, apiV1NFSAdvanced, settings, nil)
}

// NFSServiceEnabled implements Client.
func (c *httpClient) NFSServiceEnabled(ctx context.Context) (bool, error) {
	var svc nfsSettings
	if err := c.callV1(ctx, http.MethodGet, apiV1NFSSettings, nil, &svc); err != nil {
		return false, err
	}
	return svc.Enable, nil
}

// addDriveToClient ensures the connection for client includes acl (idempotent,
// refreshing an existing entry so permission changes land).
func (s *nfsAdvancedSettings) addDriveToClient(client string, acl nfsDriveAcl) {
	for i := range s.Connections {
		if s.Connections[i].Client == client {
			for j := range s.Connections[i].SharedDrives {
				if s.Connections[i].SharedDrives[j].ID == acl.ID {
					s.Connections[i].SharedDrives[j] = acl // refresh (e.g. permission)
					return
				}
			}
			s.Connections[i].SharedDrives = append(s.Connections[i].SharedDrives, acl)
			return
		}
	}
	s.Connections = append(s.Connections, nfsConnection{
		Client:       client,
		ErrorCodes:   []string{},
		SharedDrives: []nfsDriveAcl{acl},
	})
}

// removeDriveForClient removes the (driveID, client) grant, dropping the
// connection if it becomes empty. Returns true if anything changed.
func (s *nfsAdvancedSettings) removeDriveForClient(driveID, client string) bool {
	changed := false
	kept := s.Connections[:0]
	for _, conn := range s.Connections {
		if conn.Client == client {
			drives := conn.SharedDrives[:0]
			for _, d := range conn.SharedDrives {
				if d.ID == driveID {
					changed = true
					continue
				}
				drives = append(drives, d)
			}
			conn.SharedDrives = drives
		}
		if len(conn.SharedDrives) > 0 {
			kept = append(kept, conn)
		} else {
			changed = true
		}
	}
	s.Connections = kept
	return changed
}

// removeDrive deletes driveID from all connections, dropping any that become
// empty. Returns true if anything changed.
func (s *nfsAdvancedSettings) removeDrive(driveID string) bool {
	changed := false
	kept := s.Connections[:0]
	for _, conn := range s.Connections {
		drives := conn.SharedDrives[:0]
		for _, d := range conn.SharedDrives {
			if d.ID == driveID {
				changed = true
				continue
			}
			drives = append(drives, d)
		}
		conn.SharedDrives = drives
		if len(conn.SharedDrives) > 0 {
			kept = append(kept, conn)
		} else {
			changed = true
		}
	}
	s.Connections = kept
	return changed
}
