package network

import (
	"context"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/ryanwersal/pulumi-unifi/provider/config"
)

// staticRouteType is the upstream Routing.Type discriminator value used for
// static routes. The Routing object is generic (it also backs other routing
// objects), so this resource always pins Type to this value internally and does
// not expose the raw discriminator to the user.
const staticRouteType = "static-route"

// StaticRoute is the controlling (marker) struct for a UniFi static route.
type StaticRoute struct{}

// StaticRouteArgs are the user-supplied inputs for a static route.
type StaticRouteArgs struct {
	// Name of the static route.
	Name string `pulumi:"name"`
	// Network is the destination network in CIDR notation (IPv4 or IPv6), e.g. 10.0.0.0/24.
	Network string `pulumi:"network"`
	// StaticRouteType selects the route kind: nexthop-route | interface-route | blackhole.
	StaticRouteType string `pulumi:"staticRouteType"`
	// Enabled controls whether the route is active. Defaults to true.
	Enabled *bool `pulumi:"enabled,optional"`
	// Nexthop is the next-hop IP address (required when staticRouteType=nexthop-route).
	Nexthop *string `pulumi:"nexthop,optional"`
	// Distance is the administrative distance / metric of the route (1-255).
	Distance *int `pulumi:"distance,optional"`
	// Interface is the egress interface when staticRouteType=interface-route: WAN1 | WAN2 | a network ID.
	Interface *string `pulumi:"interface,optional"`
	// GatewayDevice is the MAC address of the gateway device that hosts the route.
	GatewayDevice *string `pulumi:"gatewayDevice,optional"`
	// GatewayType selects which gateway handles the route: default | switch.
	GatewayType *string `pulumi:"gatewayType,optional"`
}

// StaticRouteState is the persisted state: inputs plus controller-assigned fields.
type StaticRouteState struct {
	StaticRouteArgs
	// StaticRouteId is the controller-assigned identifier (the UniFi `_id`).
	StaticRouteId string `pulumi:"staticRouteId"`
}

// Annotate documents the resource. Must use a pointer receiver so the
// annotator can take the address of the resource and its fields.
func (r *StaticRoute) Annotate(a infer.Annotator) {
	a.Describe(&r, "A UniFi static route. Maps to a controller routing object with type=static-route, "+
		"directing traffic for a destination network via a next hop, an egress interface, or a blackhole.")
}

// toUnifi builds a go-unifi Routing from inputs. id is empty on create. The
// Type discriminator is always pinned to staticRouteType.
func (a StaticRouteArgs) toUnifi(id string) *unifi.Routing {
	r := &unifi.Routing{
		ID:                 id,
		Type:               staticRouteType,
		Name:               a.Name,
		StaticRouteNetwork: a.Network,
		StaticRouteType:    a.StaticRouteType,
		Enabled:            derefOr(a.Enabled, true),
	}
	if a.Nexthop != nil {
		r.StaticRouteNexthop = *a.Nexthop
	}
	if a.Distance != nil {
		r.StaticRouteDistance = *a.Distance
	}
	if a.Interface != nil {
		r.StaticRouteInterface = *a.Interface
	}
	if a.GatewayDevice != nil {
		r.GatewayDevice = *a.GatewayDevice
	}
	if a.GatewayType != nil {
		r.GatewayType = *a.GatewayType
	}
	return r
}

// staticRouteStrPtr reflects a controller string, falling back to the prior
// input when empty, to avoid spurious diffs on optional fields.
func staticRouteStrPtr(v string, prior *string) *string {
	if v != "" {
		return ptr(v)
	}
	return prior
}

// staticRouteStateFrom maps a controller Routing back into resource state. prior
// holds the user inputs so unset optional fields are preserved across the
// round-trip.
func staticRouteStateFrom(r *unifi.Routing, prior StaticRouteArgs) StaticRouteState {
	args := StaticRouteArgs{
		Name:            r.Name,
		Network:         r.StaticRouteNetwork,
		StaticRouteType: r.StaticRouteType,
		Enabled:         ptr(r.Enabled),
	}
	args.Nexthop = staticRouteStrPtr(r.StaticRouteNexthop, prior.Nexthop)
	if r.StaticRouteDistance != 0 {
		args.Distance = ptr(r.StaticRouteDistance)
	} else {
		args.Distance = prior.Distance
	}
	args.Interface = staticRouteStrPtr(r.StaticRouteInterface, prior.Interface)
	args.GatewayDevice = staticRouteStrPtr(r.GatewayDevice, prior.GatewayDevice)
	args.GatewayType = staticRouteStrPtr(r.GatewayType, prior.GatewayType)
	return StaticRouteState{StaticRouteArgs: args, StaticRouteId: r.ID}
}

// Create provisions a new static route.
func (StaticRoute) Create(ctx context.Context, req infer.CreateRequest[StaticRouteArgs]) (infer.CreateResponse[StaticRouteState], error) {
	if req.DryRun {
		return infer.CreateResponse[StaticRouteState]{Output: StaticRouteState{StaticRouteArgs: req.Inputs}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	created, err := cfg.Network().CreateRouting(ctx, cfg.ResolvedSite(), req.Inputs.toUnifi(""))
	if err != nil {
		return infer.CreateResponse[StaticRouteState]{}, err
	}
	return infer.CreateResponse[StaticRouteState]{ID: created.ID, Output: staticRouteStateFrom(created, req.Inputs)}, nil
}

// Read recovers state from the controller, enabling `pulumi import`.
func (StaticRoute) Read(ctx context.Context, req infer.ReadRequest[StaticRouteArgs, StaticRouteState]) (infer.ReadResponse[StaticRouteArgs, StaticRouteState], error) {
	cfg := infer.GetConfig[config.Config](ctx)
	r, err := cfg.Network().GetRouting(ctx, cfg.ResolvedSite(), req.ID)
	if err != nil {
		return infer.ReadResponse[StaticRouteArgs, StaticRouteState]{}, err
	}
	st := staticRouteStateFrom(r, req.Inputs)
	return infer.ReadResponse[StaticRouteArgs, StaticRouteState]{ID: req.ID, Inputs: st.StaticRouteArgs, State: st}, nil
}

// Update applies changed inputs in place.
func (StaticRoute) Update(ctx context.Context, req infer.UpdateRequest[StaticRouteArgs, StaticRouteState]) (infer.UpdateResponse[StaticRouteState], error) {
	if req.DryRun {
		return infer.UpdateResponse[StaticRouteState]{Output: StaticRouteState{StaticRouteArgs: req.Inputs, StaticRouteId: req.ID}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	updated, err := cfg.Network().UpdateRouting(ctx, cfg.ResolvedSite(), req.Inputs.toUnifi(req.ID))
	if err != nil {
		return infer.UpdateResponse[StaticRouteState]{}, err
	}
	return infer.UpdateResponse[StaticRouteState]{Output: staticRouteStateFrom(updated, req.Inputs)}, nil
}

// Delete removes the static route.
func (StaticRoute) Delete(ctx context.Context, req infer.DeleteRequest[StaticRouteState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[config.Config](ctx)
	return infer.DeleteResponse{}, cfg.Network().DeleteRouting(ctx, cfg.ResolvedSite(), req.ID)
}
