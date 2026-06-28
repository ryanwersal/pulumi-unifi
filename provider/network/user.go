// SPDX-License-Identifier: Apache-2.0

package network

import (
	"context"
	"fmt"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/ryanwersal/pulumi-unifi/provider/config"
)

// User is the controlling (marker) struct for a UniFi user (a known client /
// device). A user is a per-client configuration object on the controller keyed
// by MAC address: creating one registers the client and lets you assign a
// friendly name, user group, fixed IP, local DNS record, and so on.
type User struct{}

// UserArgs are the user-supplied inputs for a UniFi user (known client).
type UserArgs struct {
	// Mac is the client MAC address (e.g. "00:11:22:33:44:55"). It is the unique
	// key used to register and look up the client.
	Mac string `pulumi:"mac" provider:"replaceOnChanges"`
	// Name is a friendly name for the client.
	Name *string `pulumi:"name,optional"`
	// Note is free-form text stored alongside the client.
	Note *string `pulumi:"note,optional"`
	// UserGroupId is the ID of the user group (bandwidth profile) this client belongs to.
	UserGroupId *string `pulumi:"userGroupId,optional"`
	// NetworkId is the ID of the network the fixed IP is assigned from.
	NetworkId *string `pulumi:"networkId,optional"`

	// FixedIp is a static IPv4 address to assign to this client. Setting it enables UseFixedIp.
	FixedIp *string `pulumi:"fixedIp,optional"`
	// UseFixedIp toggles assigning the FixedIp to the client. Defaults to true when FixedIp is set.
	UseFixedIp *bool `pulumi:"useFixedIp,optional"`

	// Blocked, when true, blocks this client from accessing the network.
	Blocked *bool `pulumi:"blocked,optional"`

	// LocalDnsRecord is a local DNS hostname resolving to this client. Setting it enables LocalDnsRecordEnabled.
	LocalDnsRecord *string `pulumi:"localDnsRecord,optional"`
	// LocalDnsRecordEnabled toggles publishing the LocalDnsRecord. Defaults to true when LocalDnsRecord is set.
	LocalDnsRecordEnabled *bool `pulumi:"localDnsRecordEnabled,optional"`

	// DevIdOverride overrides the detected device fingerprint (device type id). 0 clears the override.
	DevIdOverride *int `pulumi:"devIdOverride,optional"`

	// VirtualNetworkOverrideEnabled toggles overriding the client's virtual network (VLAN). Defaults to true when VirtualNetworkOverrideId is set.
	VirtualNetworkOverrideEnabled *bool `pulumi:"virtualNetworkOverrideEnabled,optional"`
	// VirtualNetworkOverrideId is the ID of the virtual network (VLAN) the client is pinned to.
	VirtualNetworkOverrideId *string `pulumi:"virtualNetworkOverrideId,optional"`

	// FixedApEnabled toggles pinning the client to a fixed access point. Defaults to true when FixedApMac is set.
	FixedApEnabled *bool `pulumi:"fixedApEnabled,optional"`
	// FixedApMac is the MAC address of the access point the client is pinned to.
	FixedApMac *string `pulumi:"fixedApMac,optional"`
}

// UserState is the persisted state: inputs plus controller-assigned fields.
type UserState struct {
	UserArgs
	// UserId is the controller-assigned identifier (the UniFi `_id`).
	UserId string `pulumi:"userId"`
	// Hostname is the hostname the controller has observed for the client.
	Hostname string `pulumi:"hostname"`
	// Ip is the last-known IP address of the client (best-effort; may be empty).
	Ip string `pulumi:"ip"`
	// LastSeen is the Unix timestamp (seconds) the client was last seen.
	LastSeen int `pulumi:"lastSeen"`
}

// Annotate documents the resource. Must use a pointer receiver so the
// annotator can take the address of the resource and its fields.
func (u *User) Annotate(a infer.Annotator) {
	a.Describe(&u, "A UniFi user (a known client / device). Maps to a controller user object keyed by MAC "+
		"address. Creating this resource registers the client on the controller; deleting it removes the "+
		"user record. Supports assigning a friendly name, user group, fixed IP, local DNS record, virtual "+
		"network (VLAN) override, fixed access point, and blocking the client.")
}

func (d *UserArgs) Annotate(a infer.Annotator) {
	a.Describe(&d.Mac, "Mac is the client MAC address (e.g. \"00:11:22:33:44:55\"). It is the unique key used to register and look up the client.")
	a.Describe(&d.Name, "Name is a friendly name for the client.")
	a.Describe(&d.Note, "Note is free-form text stored alongside the client.")
	a.Describe(&d.UserGroupId, "UserGroupId is the ID of the user group (bandwidth profile) this client belongs to.")
	a.Describe(&d.NetworkId, "NetworkId is the ID of the network the fixed IP is assigned from.")
	a.Describe(&d.FixedIp, "FixedIp is a static IPv4 address to assign to this client. Setting it enables UseFixedIp.")
	a.Describe(&d.UseFixedIp, "UseFixedIp toggles assigning the FixedIp to the client. Defaults to true when FixedIp is set.")
	a.Describe(&d.Blocked, "Blocked, when true, blocks this client from accessing the network.")
	a.Describe(&d.LocalDnsRecord, "LocalDnsRecord is a local DNS hostname resolving to this client. Setting it enables LocalDnsRecordEnabled.")
	a.Describe(&d.LocalDnsRecordEnabled, "LocalDnsRecordEnabled toggles publishing the LocalDnsRecord. Defaults to true when LocalDnsRecord is set.")
	a.Describe(&d.DevIdOverride, "DevIdOverride overrides the detected device fingerprint (device type id). 0 clears the override.")
	a.Describe(&d.VirtualNetworkOverrideEnabled, "VirtualNetworkOverrideEnabled toggles overriding the client's virtual network (VLAN). Defaults to true when VirtualNetworkOverrideId is set.")
	a.Describe(&d.VirtualNetworkOverrideId, "VirtualNetworkOverrideId is the ID of the virtual network (VLAN) the client is pinned to.")
	a.Describe(&d.FixedApEnabled, "FixedApEnabled toggles pinning the client to a fixed access point. Defaults to true when FixedApMac is set.")
	a.Describe(&d.FixedApMac, "FixedApMac is the MAC address of the access point the client is pinned to.")
}

func (d *UserState) Annotate(a infer.Annotator) {
	a.Describe(&d.UserId, "UserId is the controller-assigned identifier (the UniFi `_id`).")
	a.Describe(&d.Hostname, "Hostname is the hostname the controller has observed for the client.")
	a.Describe(&d.Ip, "Ip is the last-known IP address of the client (best-effort; may be empty).")
	a.Describe(&d.LastSeen, "LastSeen is the Unix timestamp (seconds) the client was last seen.")
}

// toUnifi builds a go-unifi User from inputs. id is empty on create.
func (a UserArgs) toUnifi(id string) *unifi.User {
	u := &unifi.User{
		ID:  id,
		MAC: a.Mac,
	}
	if a.Name != nil {
		u.Name = *a.Name
	}
	if a.Note != nil {
		u.Note = *a.Note
	}
	if a.UserGroupId != nil {
		u.UserGroupID = *a.UserGroupId
	}
	if a.NetworkId != nil {
		u.NetworkID = *a.NetworkId
	}
	if a.FixedIp != nil {
		u.FixedIP = *a.FixedIp
		u.UseFixedIP = true
	}
	if a.UseFixedIp != nil {
		u.UseFixedIP = *a.UseFixedIp
	}
	if a.Blocked != nil {
		u.Blocked = *a.Blocked
	}
	if a.LocalDnsRecord != nil {
		u.LocalDNSRecord = *a.LocalDnsRecord
		u.LocalDNSRecordEnabled = true
	}
	if a.LocalDnsRecordEnabled != nil {
		u.LocalDNSRecordEnabled = *a.LocalDnsRecordEnabled
	}
	if a.DevIdOverride != nil {
		u.DevIdOverride = *a.DevIdOverride
	}
	if a.VirtualNetworkOverrideId != nil {
		u.VirtualNetworkOverrideID = *a.VirtualNetworkOverrideId
		u.VirtualNetworkOverrideEnabled = true
	}
	if a.VirtualNetworkOverrideEnabled != nil {
		u.VirtualNetworkOverrideEnabled = *a.VirtualNetworkOverrideEnabled
	}
	if a.FixedApMac != nil {
		u.FixedApMAC = *a.FixedApMac
		u.FixedApEnabled = true
	}
	if a.FixedApEnabled != nil {
		u.FixedApEnabled = *a.FixedApEnabled
	}
	return u
}

// userStateFrom maps a controller User back into resource state. prior holds the
// user inputs so unset optional fields are preserved across the round-trip.
func userStateFrom(u *unifi.User, prior UserArgs) UserState {
	args := UserArgs{
		Mac: u.MAC,
	}
	args.Name = vlanStrPtr(u.Name, prior.Name)
	args.Note = vlanStrPtr(u.Note, prior.Note)
	args.UserGroupId = vlanStrPtr(u.UserGroupID, prior.UserGroupId)
	args.NetworkId = vlanStrPtr(u.NetworkID, prior.NetworkId)
	args.FixedIp = vlanStrPtr(u.FixedIP, prior.FixedIp)
	args.UseFixedIp = vlanBoolPtr(u.UseFixedIP, prior.UseFixedIp)
	args.Blocked = vlanBoolPtr(u.Blocked, prior.Blocked)
	args.LocalDnsRecord = vlanStrPtr(u.LocalDNSRecord, prior.LocalDnsRecord)
	args.LocalDnsRecordEnabled = vlanBoolPtr(u.LocalDNSRecordEnabled, prior.LocalDnsRecordEnabled)
	args.DevIdOverride = vlanIntPtr(u.DevIdOverride, prior.DevIdOverride)
	args.VirtualNetworkOverrideEnabled = vlanBoolPtr(u.VirtualNetworkOverrideEnabled, prior.VirtualNetworkOverrideEnabled)
	args.VirtualNetworkOverrideId = vlanStrPtr(u.VirtualNetworkOverrideID, prior.VirtualNetworkOverrideId)
	args.FixedApEnabled = vlanBoolPtr(u.FixedApEnabled, prior.FixedApEnabled)
	args.FixedApMac = vlanStrPtr(u.FixedApMAC, prior.FixedApMac)

	return UserState{
		UserArgs: args,
		UserId:   u.ID,
		Hostname: u.Hostname,
		Ip:       u.IP,
		LastSeen: u.LastSeen,
	}
}

// Create registers a new known client on the controller.
func (User) Create(ctx context.Context, req infer.CreateRequest[UserArgs]) (infer.CreateResponse[UserState], error) {
	if req.DryRun {
		return infer.CreateResponse[UserState]{Output: UserState{UserArgs: req.Inputs}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	created, err := cfg.Network().CreateUser(ctx, cfg.ResolvedSite(), req.Inputs.toUnifi(""))
	if err != nil {
		return infer.CreateResponse[UserState]{}, wrap(fmt.Sprintf("create user %q (site %q)", req.Inputs.Mac, cfg.ResolvedSite()), err)
	}
	if created.ID == "" {
		return infer.CreateResponse[UserState]{}, infer.ProviderErrorf("created user but controller returned no ID")
	}
	return infer.CreateResponse[UserState]{ID: created.ID, Output: userStateFrom(created, req.Inputs)}, nil
}

// Read recovers state from the controller, enabling `pulumi import`.
func (User) Read(ctx context.Context, req infer.ReadRequest[UserArgs, UserState]) (infer.ReadResponse[UserArgs, UserState], error) {
	cfg := infer.GetConfig[config.Config](ctx)
	u, err := cfg.Network().GetUser(ctx, cfg.ResolvedSite(), req.ID)
	if notFound(err) {
		return infer.ReadResponse[UserArgs, UserState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[UserArgs, UserState]{}, wrap(fmt.Sprintf("read user %q (site %q)", req.ID, cfg.ResolvedSite()), err)
	}
	st := userStateFrom(u, req.Inputs)
	return infer.ReadResponse[UserArgs, UserState]{ID: req.ID, Inputs: st.UserArgs, State: st}, nil
}

// Update applies changed inputs in place.
func (User) Update(ctx context.Context, req infer.UpdateRequest[UserArgs, UserState]) (infer.UpdateResponse[UserState], error) {
	if req.DryRun {
		return infer.UpdateResponse[UserState]{Output: UserState{UserArgs: req.Inputs, UserId: req.ID}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	updated, err := cfg.Network().UpdateUser(ctx, cfg.ResolvedSite(), req.Inputs.toUnifi(req.ID))
	if err != nil {
		return infer.UpdateResponse[UserState]{}, wrap(fmt.Sprintf("update user %q (site %q)", req.ID, cfg.ResolvedSite()), err)
	}
	return infer.UpdateResponse[UserState]{Output: userStateFrom(updated, req.Inputs)}, nil
}

// Delete removes the user record from the controller. DeleteUser uses the REST
// endpoint keyed by the controller id; the MAC-keyed DeleteUserByMAC (forget-sta)
// is also available on the client but the REST delete matches the id we track.
func (User) Delete(ctx context.Context, req infer.DeleteRequest[UserState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[config.Config](ctx)
	err := cfg.Network().DeleteUser(ctx, cfg.ResolvedSite(), req.ID)
	if notFound(err) {
		return infer.DeleteResponse{}, nil
	}
	return infer.DeleteResponse{}, wrap(fmt.Sprintf("delete user %q (site %q)", req.ID, cfg.ResolvedSite()), err)
}
