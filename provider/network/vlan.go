package network

import (
	"context"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/ryanwersal/pulumi-unifi/provider/config"
)

// Vlan is the controlling (marker) struct for a UniFi network/VLAN resource.
type Vlan struct{}

// VlanArgs are the user-supplied inputs for a VLAN.
type VlanArgs struct {
	// Name of the network.
	Name string `pulumi:"name"`
	// Purpose: corporate | guest | vlan-only | wan | ... Defaults to "corporate".
	Purpose *string `pulumi:"purpose,optional"`
	// Vlan is the 802.1Q VLAN ID. When set, VLAN tagging is enabled.
	Vlan *int `pulumi:"vlan,optional"`
	// Subnet is the gateway IP/CIDR for the network, e.g. 192.168.20.1/24.
	Subnet *string `pulumi:"subnet,optional"`
	// DhcpEnabled toggles the built-in DHCP server for this network.
	DhcpEnabled *bool `pulumi:"dhcpEnabled,optional"`
	// DhcpStart is the first address of the DHCP range, e.g. 192.168.20.6.
	DhcpStart *string `pulumi:"dhcpStart,optional"`
	// DhcpStop is the last address of the DHCP range, e.g. 192.168.20.254.
	DhcpStop *string `pulumi:"dhcpStop,optional"`
	// Enabled controls whether the network is active. Defaults to true.
	Enabled *bool `pulumi:"enabled,optional"`
}

// VlanState is the persisted state: inputs plus controller-assigned fields.
type VlanState struct {
	VlanArgs
	// NetworkId is the controller-assigned identifier (the UniFi `_id`).
	NetworkId string `pulumi:"networkId"`
}

// Annotate documents the resource. Must use a pointer receiver so the
// annotator can take the address of the resource and its fields.
func (v *Vlan) Annotate(a infer.Annotator) {
	a.Describe(&v, "A UniFi Network (VLAN). Maps to a controller network configuration object.")
}

// toUnifi builds a go-unifi Network from inputs. id is empty on create.
func (a VlanArgs) toUnifi(id string) *unifi.Network {
	n := &unifi.Network{
		ID:           id,
		Name:         a.Name,
		Purpose:      derefOr(a.Purpose, "corporate"),
		Enabled:      derefOr(a.Enabled, true),
		NetworkGroup: "LAN",
	}
	if a.Vlan != nil {
		n.VLAN = *a.Vlan
		n.VLANEnabled = true
	}
	if a.Subnet != nil {
		n.IPSubnet = *a.Subnet
	}
	if a.DhcpEnabled != nil {
		n.DHCPDEnabled = *a.DhcpEnabled
	}
	if a.DhcpStart != nil {
		n.DHCPDStart = *a.DhcpStart
	}
	if a.DhcpStop != nil {
		n.DHCPDStop = *a.DhcpStop
	}
	return n
}

// stateFrom maps a controller Network back into resource state.
func stateFrom(n *unifi.Network) VlanState {
	args := VlanArgs{
		Name:    n.Name,
		Purpose: ptr(n.Purpose),
		Enabled: ptr(n.Enabled),
	}
	if n.VLANEnabled {
		args.Vlan = ptr(n.VLAN)
	}
	if n.IPSubnet != "" {
		args.Subnet = ptr(n.IPSubnet)
	}
	args.DhcpEnabled = ptr(n.DHCPDEnabled)
	if n.DHCPDStart != "" {
		args.DhcpStart = ptr(n.DHCPDStart)
	}
	if n.DHCPDStop != "" {
		args.DhcpStop = ptr(n.DHCPDStop)
	}
	return VlanState{VlanArgs: args, NetworkId: n.ID}
}

// Create provisions a new network.
func (Vlan) Create(ctx context.Context, req infer.CreateRequest[VlanArgs]) (infer.CreateResponse[VlanState], error) {
	if req.DryRun {
		return infer.CreateResponse[VlanState]{Output: VlanState{VlanArgs: req.Inputs}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	created, err := cfg.Network().CreateNetwork(ctx, cfg.ResolvedSite(), req.Inputs.toUnifi(""))
	if err != nil {
		return infer.CreateResponse[VlanState]{}, err
	}
	return infer.CreateResponse[VlanState]{ID: created.ID, Output: stateFrom(created)}, nil
}

// Read recovers state from the controller, enabling `pulumi import`.
func (Vlan) Read(ctx context.Context, req infer.ReadRequest[VlanArgs, VlanState]) (infer.ReadResponse[VlanArgs, VlanState], error) {
	cfg := infer.GetConfig[config.Config](ctx)
	n, err := cfg.Network().GetNetwork(ctx, cfg.ResolvedSite(), req.ID)
	if err != nil {
		return infer.ReadResponse[VlanArgs, VlanState]{}, err
	}
	st := stateFrom(n)
	return infer.ReadResponse[VlanArgs, VlanState]{ID: req.ID, Inputs: st.VlanArgs, State: st}, nil
}

// Update applies changed inputs in place.
func (Vlan) Update(ctx context.Context, req infer.UpdateRequest[VlanArgs, VlanState]) (infer.UpdateResponse[VlanState], error) {
	if req.DryRun {
		return infer.UpdateResponse[VlanState]{Output: VlanState{VlanArgs: req.Inputs, NetworkId: req.ID}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	updated, err := cfg.Network().UpdateNetwork(ctx, cfg.ResolvedSite(), req.Inputs.toUnifi(req.ID))
	if err != nil {
		return infer.UpdateResponse[VlanState]{}, err
	}
	return infer.UpdateResponse[VlanState]{Output: stateFrom(updated)}, nil
}

// Delete removes the network.
func (Vlan) Delete(ctx context.Context, req infer.DeleteRequest[VlanState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[config.Config](ctx)
	return infer.DeleteResponse{}, cfg.Network().DeleteNetwork(ctx, cfg.ResolvedSite(), req.ID)
}
