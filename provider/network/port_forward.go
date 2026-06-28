package network

import (
	"context"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/ryanwersal/pulumi-unifi/provider/config"
)

// PortForward is the controlling (marker) struct for a UniFi port-forwarding rule.
type PortForward struct{}

// PortForwardDestinationIp is one (destination IP, WAN interface) pair for a
// multi-WAN port-forwarding rule. It is the per-rule list element distinct from
// the singular destinationIp field.
type PortForwardDestinationIp struct {
	// DestinationIp is the public/destination IPv4 address this entry matches, or "any".
	DestinationIp *string `pulumi:"destinationIp,optional"`
	// Interface binds the entry to a WAN interface: wan | wan2.
	Interface *string `pulumi:"interface,optional"`
}

// PortForwardArgs are the user-supplied inputs for a port-forwarding rule.
type PortForwardArgs struct {
	// Name of the port-forwarding rule (1-128 characters).
	Name *string `pulumi:"name,optional"`
	// Enabled controls whether the rule is active. Defaults to true.
	Enabled *bool `pulumi:"enabled,optional"`
	// FwdPort is the internal destination port traffic is forwarded TO (single port or range).
	FwdPort *string `pulumi:"fwdPort,optional"`
	// DstPort is the external/WAN port traffic arrives ON, i.e. forwarded FROM (single port or range).
	DstPort *string `pulumi:"dstPort,optional"`
	// Fwd is the internal IPv4 address to forward traffic to.
	Fwd *string `pulumi:"fwd,optional"`
	// Proto selects the matched protocol: tcp_udp | tcp | udp. Defaults to "tcp_udp".
	Proto *string `pulumi:"proto,optional"`
	// Src restricts the source address: a single IPv4, range, CIDR, negation (!addr), or "any". Defaults to "any".
	Src *string `pulumi:"src,optional"`
	// Log enables logging of forwarded traffic. Defaults to false.
	Log *bool `pulumi:"log,optional"`
	// PfwdInterface selects the inbound WAN interface: wan | wan2 | both | all.
	PfwdInterface *string `pulumi:"pfwdInterface,optional"`
	// SrcFirewallGroupId references a firewall group used to restrict the source (when srcLimitingType=firewall_group).
	SrcFirewallGroupId *string `pulumi:"srcFirewallGroupId,optional"`
	// SrcLimitingEnabled enables restricting the rule by source address or firewall group.
	SrcLimitingEnabled *bool `pulumi:"srcLimitingEnabled,optional"`
	// SrcLimitingType selects how the source is limited: ip | firewall_group.
	SrcLimitingType *string `pulumi:"srcLimitingType,optional"`
	// DestinationIp is the public/destination IPv4 address this rule matches, or "any".
	DestinationIp *string `pulumi:"destinationIp,optional"`
	// DestinationIps maps destination IPs to specific WAN interfaces for multi-WAN setups.
	DestinationIps []PortForwardDestinationIp `pulumi:"destinationIps,optional"`
}

// PortForwardState is the persisted state: inputs plus controller-assigned fields.
type PortForwardState struct {
	PortForwardArgs
	// PortForwardId is the controller-assigned identifier (the UniFi `_id`).
	PortForwardId string `pulumi:"portForwardId"`
}

// Annotate documents the resource and its non-obvious fields.
func (p *PortForward) Annotate(a infer.Annotator) {
	a.Describe(&p, "A UniFi port-forwarding (destination NAT) rule. Forwards traffic arriving on a WAN "+
		"port (dstPort) to an internal host (fwd) and port (fwdPort).")
}

// Annotate documents the destination-IP list element.
func (d *PortForwardDestinationIp) Annotate(a infer.Annotator) {
	a.Describe(&d.Interface, "WAN interface this destination IP binds to: wan | wan2.")
}

// toUnifi builds a go-unifi PortForward from inputs. id is empty on create.
func (a PortForwardArgs) toUnifi(id string) *unifi.PortForward {
	pf := &unifi.PortForward{
		ID:      id,
		Enabled: derefOr(a.Enabled, true),
		Log:     derefOr(a.Log, false),
		Proto:   derefOr(a.Proto, "tcp_udp"),
		Src:     derefOr(a.Src, "any"),
	}
	if a.Name != nil {
		pf.Name = *a.Name
	}
	if a.FwdPort != nil {
		pf.FwdPort = *a.FwdPort
	}
	if a.DstPort != nil {
		pf.DstPort = *a.DstPort
	}
	if a.Fwd != nil {
		pf.Fwd = *a.Fwd
	}
	if a.PfwdInterface != nil {
		pf.PfwdInterface = *a.PfwdInterface
	}
	if a.SrcFirewallGroupId != nil {
		pf.SrcFirewallGroupID = *a.SrcFirewallGroupId
	}
	if a.SrcLimitingEnabled != nil {
		pf.SrcLimitingEnabled = *a.SrcLimitingEnabled
	}
	if a.SrcLimitingType != nil {
		pf.SrcLimitingType = *a.SrcLimitingType
	}
	if a.DestinationIp != nil {
		pf.DestinationIP = *a.DestinationIp
	}
	for _, dip := range a.DestinationIps {
		item := unifi.PortForwardDestinationIPs{}
		if dip.DestinationIp != nil {
			item.DestinationIP = *dip.DestinationIp
		}
		if dip.Interface != nil {
			item.Interface = *dip.Interface
		}
		pf.DestinationIPs = append(pf.DestinationIPs, item)
	}
	return pf
}

// portForwardStrPtr reflects a controller string, falling back to the prior input when empty.
func portForwardStrPtr(v string, prior *string) *string {
	if v != "" {
		return ptr(v)
	}
	return prior
}

// portForwardStateFrom maps a controller PortForward back into resource state.
// prior holds the user inputs so unset optional fields are preserved across the
// round-trip rather than producing spurious diffs.
func portForwardStateFrom(pf *unifi.PortForward, prior PortForwardArgs) PortForwardState {
	args := PortForwardArgs{
		Enabled: ptr(pf.Enabled),
		Log:     ptr(pf.Log),
	}
	args.Name = portForwardStrPtr(pf.Name, prior.Name)
	args.FwdPort = portForwardStrPtr(pf.FwdPort, prior.FwdPort)
	args.DstPort = portForwardStrPtr(pf.DstPort, prior.DstPort)
	args.Fwd = portForwardStrPtr(pf.Fwd, prior.Fwd)
	args.Proto = portForwardStrPtr(pf.Proto, prior.Proto)
	args.Src = portForwardStrPtr(pf.Src, prior.Src)
	args.PfwdInterface = portForwardStrPtr(pf.PfwdInterface, prior.PfwdInterface)
	args.SrcFirewallGroupId = portForwardStrPtr(pf.SrcFirewallGroupID, prior.SrcFirewallGroupId)
	args.SrcLimitingType = portForwardStrPtr(pf.SrcLimitingType, prior.SrcLimitingType)
	args.DestinationIp = portForwardStrPtr(pf.DestinationIP, prior.DestinationIp)
	if prior.SrcLimitingEnabled != nil || pf.SrcLimitingEnabled {
		args.SrcLimitingEnabled = ptr(pf.SrcLimitingEnabled)
	}
	if len(pf.DestinationIPs) > 0 {
		for _, dip := range pf.DestinationIPs {
			item := PortForwardDestinationIp{}
			if dip.DestinationIP != "" {
				item.DestinationIp = ptr(dip.DestinationIP)
			}
			if dip.Interface != "" {
				item.Interface = ptr(dip.Interface)
			}
			args.DestinationIps = append(args.DestinationIps, item)
		}
	} else {
		args.DestinationIps = prior.DestinationIps
	}
	return PortForwardState{PortForwardArgs: args, PortForwardId: pf.ID}
}

// Create provisions a new port-forwarding rule.
func (PortForward) Create(ctx context.Context, req infer.CreateRequest[PortForwardArgs]) (infer.CreateResponse[PortForwardState], error) {
	if req.DryRun {
		return infer.CreateResponse[PortForwardState]{Output: PortForwardState{PortForwardArgs: req.Inputs}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	created, err := cfg.Network().CreatePortForward(ctx, cfg.ResolvedSite(), req.Inputs.toUnifi(""))
	if err != nil {
		return infer.CreateResponse[PortForwardState]{}, err
	}
	return infer.CreateResponse[PortForwardState]{ID: created.ID, Output: portForwardStateFrom(created, req.Inputs)}, nil
}

// Read recovers state from the controller, enabling `pulumi import`.
func (PortForward) Read(ctx context.Context, req infer.ReadRequest[PortForwardArgs, PortForwardState]) (infer.ReadResponse[PortForwardArgs, PortForwardState], error) {
	cfg := infer.GetConfig[config.Config](ctx)
	pf, err := cfg.Network().GetPortForward(ctx, cfg.ResolvedSite(), req.ID)
	if err != nil {
		return infer.ReadResponse[PortForwardArgs, PortForwardState]{}, err
	}
	st := portForwardStateFrom(pf, req.Inputs)
	return infer.ReadResponse[PortForwardArgs, PortForwardState]{ID: req.ID, Inputs: st.PortForwardArgs, State: st}, nil
}

// Update applies changed inputs in place.
func (PortForward) Update(ctx context.Context, req infer.UpdateRequest[PortForwardArgs, PortForwardState]) (infer.UpdateResponse[PortForwardState], error) {
	if req.DryRun {
		return infer.UpdateResponse[PortForwardState]{Output: PortForwardState{PortForwardArgs: req.Inputs, PortForwardId: req.ID}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	updated, err := cfg.Network().UpdatePortForward(ctx, cfg.ResolvedSite(), req.Inputs.toUnifi(req.ID))
	if err != nil {
		return infer.UpdateResponse[PortForwardState]{}, err
	}
	return infer.UpdateResponse[PortForwardState]{Output: portForwardStateFrom(updated, req.Inputs)}, nil
}

// Delete removes the port-forwarding rule.
func (PortForward) Delete(ctx context.Context, req infer.DeleteRequest[PortForwardState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[config.Config](ctx)
	return infer.DeleteResponse{}, cfg.Network().DeletePortForward(ctx, cfg.ResolvedSite(), req.ID)
}
