package network

import (
	"context"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/ryanwersal/pulumi-unifi/provider/config"
)

// FirewallGroup is the controlling (marker) struct for a UniFi firewall group
// resource. A firewall group is a reusable set of addresses or ports that can
// be referenced from firewall rules.
type FirewallGroup struct{}

// FirewallGroupArgs are the user-supplied inputs for a firewall group.
type FirewallGroupArgs struct {
	// Name of the firewall group (1-64 characters).
	Name string `pulumi:"name"`
	// GroupType selects the kind of group: address-group | port-group | ipv6-address-group.
	GroupType *string `pulumi:"groupType,optional"`
	// GroupMembers are the entries in the group. The accepted format depends on
	// GroupType: IPv4 addresses/CIDRs for address-group, port numbers/ranges for
	// port-group, or IPv6 addresses/CIDRs for ipv6-address-group.
	GroupMembers []string `pulumi:"groupMembers,optional"`
}

// FirewallGroupState is the persisted state: inputs plus controller-assigned fields.
type FirewallGroupState struct {
	FirewallGroupArgs
	// FirewallGroupId is the controller-assigned identifier (the UniFi `_id`).
	FirewallGroupId string `pulumi:"firewallGroupId"`
}

// Annotate documents the resource. Must use a pointer receiver so the
// annotator can take the address of the resource and its fields.
func (g *FirewallGroup) Annotate(a infer.Annotator) {
	a.Describe(&g, "A UniFi firewall group: a reusable set of addresses or ports referenced by firewall rules.")
}

func (d *FirewallGroupArgs) Annotate(a infer.Annotator) {
	a.Describe(&d.Name, "Name of the firewall group (1-64 characters).")
	a.Describe(&d.GroupType, "GroupType selects the kind of group: address-group | port-group | ipv6-address-group.")
	a.Describe(&d.GroupMembers, "GroupMembers are the entries in the group. The accepted format depends on GroupType: IPv4 addresses/CIDRs for address-group, port numbers/ranges for port-group, or IPv6 addresses/CIDRs for ipv6-address-group.")
}

func (s *FirewallGroupState) Annotate(a infer.Annotator) {
	a.Describe(&s.FirewallGroupId, "FirewallGroupId is the controller-assigned identifier (the UniFi `_id`).")
}

// toUnifi builds a go-unifi FirewallGroup from inputs. id is empty on create.
func (a FirewallGroupArgs) toUnifi(id string) *unifi.FirewallGroup {
	g := &unifi.FirewallGroup{
		ID:           id,
		Name:         a.Name,
		GroupType:    derefOr(a.GroupType, "address-group"),
		GroupMembers: a.GroupMembers,
	}
	return g
}

// firewallGroupStateFrom maps a controller FirewallGroup back into resource
// state. prior holds the user inputs so unset optional fields are preserved
// across the round-trip.
func firewallGroupStateFrom(g *unifi.FirewallGroup, prior FirewallGroupArgs) FirewallGroupState {
	args := FirewallGroupArgs{
		Name: g.Name,
	}
	if g.GroupType != "" {
		args.GroupType = ptr(g.GroupType)
	} else {
		args.GroupType = prior.GroupType
	}
	if len(g.GroupMembers) > 0 {
		args.GroupMembers = g.GroupMembers
	} else {
		args.GroupMembers = prior.GroupMembers
	}
	return FirewallGroupState{FirewallGroupArgs: args, FirewallGroupId: g.ID}
}

// Create provisions a new firewall group.
func (FirewallGroup) Create(ctx context.Context, req infer.CreateRequest[FirewallGroupArgs]) (infer.CreateResponse[FirewallGroupState], error) {
	if req.DryRun {
		return infer.CreateResponse[FirewallGroupState]{Output: FirewallGroupState{FirewallGroupArgs: req.Inputs}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	created, err := cfg.Network().CreateFirewallGroup(ctx, cfg.ResolvedSite(), req.Inputs.toUnifi(""))
	if err != nil {
		return infer.CreateResponse[FirewallGroupState]{}, err
	}
	return infer.CreateResponse[FirewallGroupState]{ID: created.ID, Output: firewallGroupStateFrom(created, req.Inputs)}, nil
}

// Read recovers state from the controller, enabling `pulumi import`.
func (FirewallGroup) Read(ctx context.Context, req infer.ReadRequest[FirewallGroupArgs, FirewallGroupState]) (infer.ReadResponse[FirewallGroupArgs, FirewallGroupState], error) {
	cfg := infer.GetConfig[config.Config](ctx)
	g, err := cfg.Network().GetFirewallGroup(ctx, cfg.ResolvedSite(), req.ID)
	if notFound(err) {
		return infer.ReadResponse[FirewallGroupArgs, FirewallGroupState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[FirewallGroupArgs, FirewallGroupState]{}, err
	}
	st := firewallGroupStateFrom(g, req.Inputs)
	return infer.ReadResponse[FirewallGroupArgs, FirewallGroupState]{ID: req.ID, Inputs: st.FirewallGroupArgs, State: st}, nil
}

// Update applies changed inputs in place.
func (FirewallGroup) Update(ctx context.Context, req infer.UpdateRequest[FirewallGroupArgs, FirewallGroupState]) (infer.UpdateResponse[FirewallGroupState], error) {
	if req.DryRun {
		return infer.UpdateResponse[FirewallGroupState]{Output: FirewallGroupState{FirewallGroupArgs: req.Inputs, FirewallGroupId: req.ID}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	updated, err := cfg.Network().UpdateFirewallGroup(ctx, cfg.ResolvedSite(), req.Inputs.toUnifi(req.ID))
	if err != nil {
		return infer.UpdateResponse[FirewallGroupState]{}, err
	}
	return infer.UpdateResponse[FirewallGroupState]{Output: firewallGroupStateFrom(updated, req.Inputs)}, nil
}

// Delete removes the firewall group.
func (FirewallGroup) Delete(ctx context.Context, req infer.DeleteRequest[FirewallGroupState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[config.Config](ctx)
	return infer.DeleteResponse{}, cfg.Network().DeleteFirewallGroup(ctx, cfg.ResolvedSite(), req.ID)
}
