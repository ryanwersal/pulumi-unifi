// SPDX-License-Identifier: Apache-2.0

package drive

import (
	"context"
	"errors"
	"fmt"

	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/ryanwersal/pulumi-unifi/provider/config"
	"github.com/ryanwersal/pulumi-unifi/provider/internal/driveapi"
)

// Share manages a UniFi Drive shared drive on a UNAS appliance.
//
// The reverse-engineered Drive API supports only CREATE and DELETE of a shared
// drive — there is no update endpoint — so every input is replace-only.
// Changing any of them REPLACES the share, which DELETES the old drive and all
// data in it. Set `pulumi.protect(true)` on drives whose data must not be lost.
//
// Requires the provider's `unasUrl` / `unasUsername` / `unasPassword` to be
// configured (the UNAS is a separate host from the main controller). This uses
// UniFi Drive's PRIVATE, unversioned API; treat it as best-effort across
// firmware.
type Share struct{}

// ShareArgs are the user-supplied inputs. All are replace-only (create+delete).
type ShareArgs struct {
	// Name is the shared drive's name. Must be unique on the appliance.
	Name string `pulumi:"name" provider:"replaceOnChanges"`
	// StoragePoolId pins the storage pool. Omit to use the appliance's first pool.
	StoragePoolId *string `pulumi:"storagePoolId,optional" provider:"replaceOnChanges"`
	// QuotaGib is the size limit in gibibytes. Omit (or <= 0) for no quota.
	QuotaGib *int `pulumi:"quotaGib,optional" provider:"replaceOnChanges"`
}

// ShareState is the persisted state: inputs plus appliance-assigned outputs.
type ShareState struct {
	ShareArgs
	// ShareId is the appliance-assigned drive identifier.
	ShareId string `pulumi:"shareId"`
	// PoolId is the storage pool the drive actually lives in (the resolved value
	// when StoragePoolId was omitted).
	PoolId string `pulumi:"poolId"`
	// ExportPath is the documented NFS export path for the drive, e.g.
	// /var/nfs/shared/<name>.
	ExportPath string `pulumi:"exportPath"`
}

func (s *Share) Annotate(a infer.Annotator) {
	a.Describe(&s, "A UniFi Drive shared drive on a UNAS appliance. Managed via UniFi Drive's PRIVATE /proxy/drive API "+
		"(the UNAS is a separate host — configure `unasUrl`/`unasUsername`/`unasPassword`). The API supports only "+
		"create and delete, so every input is replace-only: changing name, pool, or quota REPLACES the drive and "+
		"DELETES its data. Use `pulumi.protect(true)` to guard important drives.")
}

func (a *ShareArgs) Annotate(an infer.Annotator) {
	an.Describe(&a.Name, "Name is the shared drive's name. Must be unique on the appliance. Changing it replaces (and deletes) the drive.")
	an.Describe(&a.StoragePoolId, "StoragePoolId pins the storage pool. Omit to use the appliance's first pool.")
	an.Describe(&a.QuotaGib, "QuotaGib is the size limit in gibibytes. Omit (or <= 0) for no quota.")
}

func (s *ShareState) Annotate(a infer.Annotator) {
	a.Describe(&s.ShareId, "ShareId is the appliance-assigned drive identifier.")
	a.Describe(&s.PoolId, "PoolId is the storage pool the drive actually lives in (resolved when StoragePoolId is omitted).")
	a.Describe(&s.ExportPath, "ExportPath is the documented NFS export path for the drive, e.g. /var/nfs/shared/<name>.")
}

// shareStateFrom builds resource state from an appliance share. prior holds the
// user inputs so replace-only optionals echo back exactly (no spurious diff),
// while PoolId/ExportPath surface the appliance's actual values.
func shareStateFrom(s *driveapi.Share, prior ShareArgs) ShareState {
	return ShareState{
		ShareArgs: ShareArgs{
			Name:          s.Name,
			StoragePoolId: prior.StoragePoolId,
			QuotaGib:      prior.QuotaGib,
		},
		ShareId:    s.ID,
		PoolId:     s.StoragePoolID,
		ExportPath: s.ExportPath,
	}
}

func (a ShareArgs) toSpec() driveapi.ShareSpec {
	spec := driveapi.ShareSpec{Name: a.Name}
	if a.StoragePoolId != nil {
		spec.StoragePoolID = *a.StoragePoolId
	}
	if a.QuotaGib != nil {
		spec.QuotaGiB = int64(*a.QuotaGib)
	}
	return spec
}

func (Share) Create(ctx context.Context, req infer.CreateRequest[ShareArgs]) (infer.CreateResponse[ShareState], error) {
	if req.DryRun {
		return infer.CreateResponse[ShareState]{Output: ShareState{ShareArgs: req.Inputs}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	client, err := cfg.Drive()
	if err != nil {
		return infer.CreateResponse[ShareState]{}, err
	}
	created, err := client.CreateShare(ctx, req.Inputs.toSpec())
	if err != nil {
		return infer.CreateResponse[ShareState]{}, fmt.Errorf("create drive share %q: %w", req.Inputs.Name, err)
	}
	if created.ID == "" {
		return infer.CreateResponse[ShareState]{}, infer.ProviderErrorf("created drive share but appliance returned no ID")
	}
	return infer.CreateResponse[ShareState]{ID: created.ID, Output: shareStateFrom(created, req.Inputs)}, nil
}

func (Share) Read(ctx context.Context, req infer.ReadRequest[ShareArgs, ShareState]) (infer.ReadResponse[ShareArgs, ShareState], error) {
	cfg := infer.GetConfig[config.Config](ctx)
	client, err := cfg.Drive()
	if err != nil {
		return infer.ReadResponse[ShareArgs, ShareState]{}, err
	}
	share, err := client.GetShareByID(ctx, req.ID)
	if errors.Is(err, driveapi.ErrShareNotFound) {
		// Empty response marks the resource as deleted out-of-band.
		return infer.ReadResponse[ShareArgs, ShareState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[ShareArgs, ShareState]{}, fmt.Errorf("read drive share %q: %w", req.ID, err)
	}
	st := shareStateFrom(share, req.Inputs)
	return infer.ReadResponse[ShareArgs, ShareState]{ID: req.ID, Inputs: st.ShareArgs, State: st}, nil
}

func (Share) Delete(ctx context.Context, req infer.DeleteRequest[ShareState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[config.Config](ctx)
	client, err := cfg.Drive()
	if err != nil {
		return infer.DeleteResponse{}, err
	}
	if err := client.DeleteShare(ctx, req.ID); err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("delete drive share %q: %w", req.ID, err)
	}
	return infer.DeleteResponse{}, nil
}
