// SPDX-License-Identifier: Apache-2.0

package drive

import (
	"context"
	"errors"
	"fmt"
	"strings"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/ryanwersal/pulumi-unifi/provider/config"
	"github.com/ryanwersal/pulumi-unifi/provider/internal/driveapi"
)

// NfsPermission is the access level of an NFS export grant.
type NfsPermission string

const (
	NfsPermissionRw NfsPermission = "rw"
	NfsPermissionRo NfsPermission = "ro"
)

func (NfsPermission) Values() []infer.EnumValue[NfsPermission] {
	return []infer.EnumValue[NfsPermission]{
		{Name: "rw", Value: NfsPermissionRw, Description: "Read-write access."},
		{Name: "ro", Value: NfsPermissionRo, Description: "Read-only access."},
	}
}

// NfsExport grants an NFS client access to a UniFi Drive shared drive.
//
// On the appliance, NFS access is one global list of (client -> drives) grants;
// this resource owns a single (share, client) entry within it. Create/Update/
// Delete are read-modify-writes of that global list, serialised within the
// provider process, so multiple NfsExport resources in one `pulumi up` are safe.
// Concurrent writes from SEPARATE processes against the same appliance can still
// race — avoid running two Drive stacks at once.
type NfsExport struct{}

// NfsExportArgs are the user-supplied inputs.
type NfsExportArgs struct {
	// ShareId is the drive to export (the `shareId` of a unifi:drive:Share).
	ShareId string `pulumi:"shareId" provider:"replaceOnChanges"`
	// Client is the NFS client IP or CIDR allowed to mount the drive.
	Client string `pulumi:"client" provider:"replaceOnChanges"`
	// Permission is the access level: "rw" (default) or "ro".
	Permission *NfsPermission `pulumi:"permission,optional"`
}

// NfsExportState is the persisted state: inputs plus resolved outputs.
type NfsExportState struct {
	NfsExportArgs
	// ShareName is the exported drive's name (resolved from ShareId).
	ShareName string `pulumi:"shareName"`
}

func (n *NfsExport) Annotate(a infer.Annotator) {
	a.Describe(&n, "An NFS export grant giving one client access to a UniFi Drive shared drive. Owns a single "+
		"(share, client) entry in the appliance's global NFS export list; writes are read-modify-writes serialised "+
		"within the provider process. Requires the global NFS service to be enabled on the appliance.")
}

func (n *NfsExportArgs) Annotate(a infer.Annotator) {
	a.Describe(&n.ShareId, "ShareId is the drive to export (the shareId of a unifi:drive:Share).")
	a.Describe(&n.Client, "Client is the NFS client IP or CIDR allowed to mount the drive.")
	a.Describe(&n.Permission, `Permission is the access level: "rw" (default) or "ro".`)
	a.SetDefault(&n.Permission, NfsPermissionRw)
}

func (n *NfsExportState) Annotate(a infer.Annotator) {
	a.Describe(&n.ShareName, "ShareName is the exported drive's name (resolved from ShareId).")
}

// exportID composes the resource ID from its natural key.
func exportID(shareID, client string) string { return shareID + "/" + client }

// parseExportID splits an export ID back into (shareID, client). Share IDs carry
// no slash; client CIDRs do, so split on the FIRST slash only.
func parseExportID(id string) (shareID, client string, ok bool) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func (n NfsExportArgs) permission() string {
	return string(derefOr(n.Permission, NfsPermissionRw))
}

// ensure applies the grant and returns the resolved state.
func ensureExport(ctx context.Context, client driveapi.Client, args NfsExportArgs) (NfsExportState, error) {
	if err := client.EnsureNFSExport(ctx, args.ShareId, args.Client, args.permission()); err != nil {
		return NfsExportState{}, fmt.Errorf("ensure nfs export of share %q to %q: %w", args.ShareId, args.Client, err)
	}
	// Warn (don't fail) if the global NFS service is off — the grant is stored
	// but unreachable until an admin enables NFS in the Drive settings.
	if on, err := client.NFSServiceEnabled(ctx); err == nil && !on {
		p.GetLogger(ctx).Warning("the appliance's global NFS service is disabled; this export will not be reachable until NFS is enabled in UniFi Drive settings")
	}
	exp, err := client.GetNFSExport(ctx, args.ShareId, args.Client)
	if err != nil {
		return NfsExportState{}, fmt.Errorf("read back nfs export of share %q to %q: %w", args.ShareId, args.Client, err)
	}
	return NfsExportState{
		NfsExportArgs: NfsExportArgs{ShareId: args.ShareId, Client: args.Client, Permission: ptr(NfsPermission(exp.Permission))},
		ShareName:     exp.ShareName,
	}, nil
}

func (NfsExport) Create(ctx context.Context, req infer.CreateRequest[NfsExportArgs]) (infer.CreateResponse[NfsExportState], error) {
	if req.DryRun {
		return infer.CreateResponse[NfsExportState]{Output: NfsExportState{NfsExportArgs: req.Inputs}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	client, err := cfg.Drive()
	if err != nil {
		return infer.CreateResponse[NfsExportState]{}, err
	}
	st, err := ensureExport(ctx, client, req.Inputs)
	if err != nil {
		return infer.CreateResponse[NfsExportState]{}, err
	}
	return infer.CreateResponse[NfsExportState]{ID: exportID(req.Inputs.ShareId, req.Inputs.Client), Output: st}, nil
}

func (NfsExport) Read(ctx context.Context, req infer.ReadRequest[NfsExportArgs, NfsExportState]) (infer.ReadResponse[NfsExportArgs, NfsExportState], error) {
	shareID, client, ok := parseExportID(req.ID)
	if !ok {
		return infer.ReadResponse[NfsExportArgs, NfsExportState]{}, fmt.Errorf("invalid nfs export id %q (want <shareId>/<client>)", req.ID)
	}
	cfg := infer.GetConfig[config.Config](ctx)
	drive, err := cfg.Drive()
	if err != nil {
		return infer.ReadResponse[NfsExportArgs, NfsExportState]{}, err
	}
	exp, err := drive.GetNFSExport(ctx, shareID, client)
	if errors.Is(err, driveapi.ErrExportNotFound) {
		return infer.ReadResponse[NfsExportArgs, NfsExportState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[NfsExportArgs, NfsExportState]{}, fmt.Errorf("read nfs export %q: %w", req.ID, err)
	}
	st := NfsExportState{
		NfsExportArgs: NfsExportArgs{ShareId: shareID, Client: client, Permission: ptr(NfsPermission(exp.Permission))},
		ShareName:     exp.ShareName,
	}
	return infer.ReadResponse[NfsExportArgs, NfsExportState]{ID: req.ID, Inputs: st.NfsExportArgs, State: st}, nil
}

func (NfsExport) Update(ctx context.Context, req infer.UpdateRequest[NfsExportArgs, NfsExportState]) (infer.UpdateResponse[NfsExportState], error) {
	if req.DryRun {
		return infer.UpdateResponse[NfsExportState]{Output: NfsExportState{NfsExportArgs: req.Inputs}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	client, err := cfg.Drive()
	if err != nil {
		return infer.UpdateResponse[NfsExportState]{}, err
	}
	st, err := ensureExport(ctx, client, req.Inputs)
	if err != nil {
		return infer.UpdateResponse[NfsExportState]{}, err
	}
	return infer.UpdateResponse[NfsExportState]{Output: st}, nil
}

func (NfsExport) Delete(ctx context.Context, req infer.DeleteRequest[NfsExportState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[config.Config](ctx)
	client, err := cfg.Drive()
	if err != nil {
		return infer.DeleteResponse{}, err
	}
	if err := client.RemoveNFSExport(ctx, req.State.ShareId, req.State.Client); err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("delete nfs export %q: %w", req.ID, err)
	}
	return infer.DeleteResponse{}, nil
}
