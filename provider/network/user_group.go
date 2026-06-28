// SPDX-License-Identifier: Apache-2.0

package network

import (
	"context"
	"fmt"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/ryanwersal/pulumi-unifi/provider/config"
)

// UserGroup is the controlling (marker) struct for a UniFi user group resource.
// User groups apply per-client bandwidth (QoS rate) limits.
type UserGroup struct{}

// UserGroupArgs are the user-supplied inputs for a user group.
type UserGroupArgs struct {
	// Name of the user group.
	Name string `pulumi:"name"`
	// QosRateMaxDown is the maximum download rate in kbps. -1 means unlimited. Defaults to -1.
	QosRateMaxDown *int `pulumi:"qosRateMaxDown,optional"`
	// QosRateMaxUp is the maximum upload rate in kbps. -1 means unlimited. Defaults to -1.
	QosRateMaxUp *int `pulumi:"qosRateMaxUp,optional"`
}

func (g *UserGroupArgs) Annotate(a infer.Annotator) {
	a.Describe(&g.Name, "Name of the user group.")
	a.Describe(&g.QosRateMaxDown, "QosRateMaxDown is the maximum download rate in kbps. -1 means unlimited. Defaults to -1.")
	a.Describe(&g.QosRateMaxUp, "QosRateMaxUp is the maximum upload rate in kbps. -1 means unlimited. Defaults to -1.")
}

// UserGroupState is the persisted state: inputs plus controller-assigned fields.
type UserGroupState struct {
	UserGroupArgs
	// UserGroupId is the controller-assigned identifier (the UniFi `_id`).
	UserGroupId string `pulumi:"userGroupId"`
}

func (g *UserGroupState) Annotate(a infer.Annotator) {
	a.Describe(&g.UserGroupId, "UserGroupId is the controller-assigned identifier (the UniFi `_id`).")
}

// Annotate documents the resource. Must use a pointer receiver so the
// annotator can take the address of the resource and its fields.
func (g *UserGroup) Annotate(a infer.Annotator) {
	a.Describe(&g, "A UniFi user group, applying per-client bandwidth (QoS rate) limits "+
		"that can be assigned to clients.")
}

// toUnifi builds a go-unifi UserGroup from inputs. id is empty on create.
func (a UserGroupArgs) toUnifi(id string) *unifi.UserGroup {
	return &unifi.UserGroup{
		ID:             id,
		Name:           a.Name,
		QOSRateMaxDown: derefOr(a.QosRateMaxDown, -1),
		QOSRateMaxUp:   derefOr(a.QosRateMaxUp, -1),
	}
}

// userGroupStateFrom maps a controller UserGroup back into resource state. prior
// holds the user inputs so unset optional fields are preserved across the round-trip.
func userGroupStateFrom(u *unifi.UserGroup, prior UserGroupArgs) UserGroupState {
	args := UserGroupArgs{
		Name:           u.Name,
		QosRateMaxDown: ptr(u.QOSRateMaxDown),
		QosRateMaxUp:   ptr(u.QOSRateMaxUp),
	}
	return UserGroupState{UserGroupArgs: args, UserGroupId: u.ID}
}

// Create provisions a new user group.
func (UserGroup) Create(ctx context.Context, req infer.CreateRequest[UserGroupArgs]) (infer.CreateResponse[UserGroupState], error) {
	if req.DryRun {
		return infer.CreateResponse[UserGroupState]{Output: UserGroupState{UserGroupArgs: req.Inputs}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	created, err := cfg.Network().CreateUserGroup(ctx, cfg.ResolvedSite(), req.Inputs.toUnifi(""))
	if err != nil {
		return infer.CreateResponse[UserGroupState]{}, wrap(fmt.Sprintf("create user group %q (site %q)", req.Inputs.Name, cfg.ResolvedSite()), err)
	}
	return infer.CreateResponse[UserGroupState]{ID: created.ID, Output: userGroupStateFrom(created, req.Inputs)}, nil
}

// Read recovers state from the controller, enabling `pulumi import`.
func (UserGroup) Read(ctx context.Context, req infer.ReadRequest[UserGroupArgs, UserGroupState]) (infer.ReadResponse[UserGroupArgs, UserGroupState], error) {
	cfg := infer.GetConfig[config.Config](ctx)
	u, err := cfg.Network().GetUserGroup(ctx, cfg.ResolvedSite(), req.ID)
	if notFound(err) {
		return infer.ReadResponse[UserGroupArgs, UserGroupState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[UserGroupArgs, UserGroupState]{}, wrap(fmt.Sprintf("read user group %q (site %q)", req.ID, cfg.ResolvedSite()), err)
	}
	st := userGroupStateFrom(u, req.Inputs)
	return infer.ReadResponse[UserGroupArgs, UserGroupState]{ID: req.ID, Inputs: st.UserGroupArgs, State: st}, nil
}

// Update applies changed inputs in place.
func (UserGroup) Update(ctx context.Context, req infer.UpdateRequest[UserGroupArgs, UserGroupState]) (infer.UpdateResponse[UserGroupState], error) {
	if req.DryRun {
		return infer.UpdateResponse[UserGroupState]{Output: UserGroupState{UserGroupArgs: req.Inputs, UserGroupId: req.ID}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	updated, err := cfg.Network().UpdateUserGroup(ctx, cfg.ResolvedSite(), req.Inputs.toUnifi(req.ID))
	if err != nil {
		return infer.UpdateResponse[UserGroupState]{}, wrap(fmt.Sprintf("update user group %q (site %q)", req.ID, cfg.ResolvedSite()), err)
	}
	return infer.UpdateResponse[UserGroupState]{Output: userGroupStateFrom(updated, req.Inputs)}, nil
}

// Delete removes the user group.
func (UserGroup) Delete(ctx context.Context, req infer.DeleteRequest[UserGroupState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[config.Config](ctx)
	err := cfg.Network().DeleteUserGroup(ctx, cfg.ResolvedSite(), req.ID)
	if notFound(err) {
		return infer.DeleteResponse{}, nil
	}
	return infer.DeleteResponse{}, wrap(fmt.Sprintf("delete user group %q (site %q)", req.ID, cfg.ResolvedSite()), err)
}
