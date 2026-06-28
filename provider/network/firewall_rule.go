package network

import (
	"context"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/ryanwersal/pulumi-unifi/provider/config"
)

// FirewallRule is the controlling (marker) struct for a classic per-ruleset
// UniFi firewall rule resource.
type FirewallRule struct{}

// FirewallRuleProtocolMatch groups the protocol / ICMP-type match selectors.
type FirewallRuleProtocolMatch struct {
	// Protocol the rule matches (IPv4 rulesets), e.g. all | tcp | udp | tcp_udp |
	// icmp | a numeric protocol number, or empty for any.
	Protocol *string `pulumi:"protocol,optional"`
	// ProtocolV6 the rule matches (IPv6 rulesets), e.g. all | tcp | udp |
	// tcp_udp | icmpv6 | a numeric protocol number, or empty for any.
	ProtocolV6 *string `pulumi:"protocolV6,optional"`
	// MatchExcepted inverts the protocol match (match all except the specified
	// protocol).
	MatchExcepted *bool `pulumi:"matchExcepted,optional"`
	// IcmpTypename restricts an icmp rule to a specific ICMP type, e.g.
	// echo-request, destination-unreachable, time-exceeded.
	IcmpTypename *string `pulumi:"icmpTypename,optional"`
	// Icmpv6Typename restricts an icmpv6 rule to a specific ICMPv6 type, e.g.
	// echo-request, packet-too-big, neighbor-solicitation.
	Icmpv6Typename *string `pulumi:"icmpv6Typename,optional"`
}

func (pm *FirewallRuleProtocolMatch) Annotate(a infer.Annotator) {
	a.Describe(&pm.Protocol, "Protocol the rule matches (IPv4 rulesets), e.g. all | tcp | udp | tcp_udp | icmp | a numeric protocol number, or empty for any.")
	a.Describe(&pm.ProtocolV6, "ProtocolV6 the rule matches (IPv6 rulesets), e.g. all | tcp | udp | tcp_udp | icmpv6 | a numeric protocol number, or empty for any.")
	a.Describe(&pm.MatchExcepted, "MatchExcepted inverts the protocol match (match all except the specified protocol).")
	a.Describe(&pm.IcmpTypename, "IcmpTypename restricts an icmp rule to a specific ICMP type, e.g. echo-request, destination-unreachable, time-exceeded.")
	a.Describe(&pm.Icmpv6Typename, "Icmpv6Typename restricts an icmpv6 rule to a specific ICMPv6 type, e.g. echo-request, packet-too-big, neighbor-solicitation.")
}

// FirewallRuleSource groups the source-side match selectors.
type FirewallRuleSource struct {
	// Address is the source IPv4 address or CIDR to match.
	Address *string `pulumi:"address,optional"`
	// AddressIpv6 is the source IPv6 address or CIDR to match.
	AddressIpv6 *string `pulumi:"addressIpv6,optional"`
	// Port is the source port or port range to match.
	Port *string `pulumi:"port,optional"`
	// Mac is the source MAC address to match.
	Mac *string `pulumi:"mac,optional"`
	// NetworkId is the source network (firewall network conf) ID to match.
	NetworkId *string `pulumi:"networkId,optional"`
	// NetworkType selects how NetworkId is interpreted: ADDRv4 | NETv4.
	NetworkType *string `pulumi:"networkType,optional"`
	// FirewallGroupIds are the source firewall group IDs to match.
	FirewallGroupIds []string `pulumi:"firewallGroupIds,optional"`
}

func (s *FirewallRuleSource) Annotate(a infer.Annotator) {
	a.Describe(&s.Address, "Address is the source IPv4 address or CIDR to match.")
	a.Describe(&s.AddressIpv6, "AddressIpv6 is the source IPv6 address or CIDR to match.")
	a.Describe(&s.Port, "Port is the source port or port range to match.")
	a.Describe(&s.Mac, "Mac is the source MAC address to match.")
	a.Describe(&s.NetworkId, "NetworkId is the source network (firewall network conf) ID to match.")
	a.Describe(&s.NetworkType, "NetworkType selects how NetworkId is interpreted: ADDRv4 | NETv4.")
	a.Describe(&s.FirewallGroupIds, "FirewallGroupIds are the source firewall group IDs to match.")
}

// FirewallRuleDestination groups the destination-side match selectors. Note
// there is no mac field (asymmetric with source).
type FirewallRuleDestination struct {
	// Address is the destination IPv4 address or CIDR to match.
	Address *string `pulumi:"address,optional"`
	// AddressIpv6 is the destination IPv6 address or CIDR to match.
	AddressIpv6 *string `pulumi:"addressIpv6,optional"`
	// Port is the destination port or port range to match.
	Port *string `pulumi:"port,optional"`
	// NetworkId is the destination network (firewall network conf) ID to match.
	NetworkId *string `pulumi:"networkId,optional"`
	// NetworkType selects how NetworkId is interpreted: ADDRv4 | NETv4.
	NetworkType *string `pulumi:"networkType,optional"`
	// FirewallGroupIds are the destination firewall group IDs to match.
	FirewallGroupIds []string `pulumi:"firewallGroupIds,optional"`
}

func (d *FirewallRuleDestination) Annotate(a infer.Annotator) {
	a.Describe(&d.Address, "Address is the destination IPv4 address or CIDR to match.")
	a.Describe(&d.AddressIpv6, "AddressIpv6 is the destination IPv6 address or CIDR to match.")
	a.Describe(&d.Port, "Port is the destination port or port range to match.")
	a.Describe(&d.NetworkId, "NetworkId is the destination network (firewall network conf) ID to match.")
	a.Describe(&d.NetworkType, "NetworkType selects how NetworkId is interpreted: ADDRv4 | NETv4.")
	a.Describe(&d.FirewallGroupIds, "FirewallGroupIds are the destination firewall group IDs to match.")
}

// FirewallRuleConnectionState groups the conntrack-state match toggles.
type FirewallRuleConnectionState struct {
	// Established matches packets in the established connection state.
	Established *bool `pulumi:"established,optional"`
	// New matches packets in the new connection state.
	New *bool `pulumi:"new,optional"`
	// Related matches packets in the related connection state.
	Related *bool `pulumi:"related,optional"`
	// Invalid matches packets in the invalid connection state.
	Invalid *bool `pulumi:"invalid,optional"`
}

func (cs *FirewallRuleConnectionState) Annotate(a infer.Annotator) {
	a.Describe(&cs.Established, "Established matches packets in the established connection state.")
	a.Describe(&cs.New, "New matches packets in the new connection state.")
	a.Describe(&cs.Related, "Related matches packets in the related connection state.")
	a.Describe(&cs.Invalid, "Invalid matches packets in the invalid connection state.")
}

// FirewallRuleArgs are the user-supplied inputs for a firewall rule.
type FirewallRuleArgs struct {
	// Name of the firewall rule (1-128 characters).
	Name string `pulumi:"name"`
	// RuleIndex is the ordering index of the rule within its ruleset. Must be
	// >= 2000 and < 3000, or >= 4000 and < 5000.
	RuleIndex int `pulumi:"ruleIndex"`
	// Action taken on matching traffic: accept | drop | reject.
	Action *string `pulumi:"action,optional"`
	// Ruleset the rule belongs to, from the perspective of the security gateway:
	// WAN_IN | WAN_OUT | WAN_LOCAL | LAN_IN | LAN_OUT | LAN_LOCAL | GUEST_IN |
	// GUEST_OUT | GUEST_LOCAL | WANv6_IN | WANv6_OUT | WANv6_LOCAL | LANv6_IN |
	// LANv6_OUT | LANv6_LOCAL | GUESTv6_IN | GUESTv6_OUT | GUESTv6_LOCAL.
	Ruleset *string `pulumi:"ruleset,optional"`
	// Enabled controls whether the rule is active. Defaults to true.
	Enabled *bool `pulumi:"enabled,optional"`

	// Logging enables logging of packets that match this rule.
	Logging *bool `pulumi:"logging,optional"`
	// IpSec matches on IPsec encapsulation: match-ipsec | match-none.
	IpSec *string `pulumi:"ipSec,optional"`
	// SettingPreference controls whether the rule uses automatic or manual
	// settings: auto | manual.
	SettingPreference *string `pulumi:"settingPreference,optional"`

	// ProtocolMatch groups the protocol / ICMP-type match selectors.
	ProtocolMatch *FirewallRuleProtocolMatch `pulumi:"protocolMatch,optional"`
	// Source groups the source-side match selectors.
	Source *FirewallRuleSource `pulumi:"source,optional"`
	// Destination groups the destination-side match selectors.
	Destination *FirewallRuleDestination `pulumi:"destination,optional"`
	// ConnectionState groups the conntrack-state match toggles.
	ConnectionState *FirewallRuleConnectionState `pulumi:"connectionState,optional"`
}

func (r *FirewallRuleArgs) Annotate(a infer.Annotator) {
	a.Describe(&r.Name, "Name of the firewall rule (1-128 characters).")
	a.Describe(&r.RuleIndex, "RuleIndex is the ordering index of the rule within its ruleset. Must be >= 2000 and < 3000, or >= 4000 and < 5000.")
	a.Describe(&r.Action, "Action taken on matching traffic: accept | drop | reject.")
	a.Describe(&r.Ruleset, "Ruleset the rule belongs to, from the perspective of the security gateway: WAN_IN | WAN_OUT | WAN_LOCAL | LAN_IN | LAN_OUT | LAN_LOCAL | GUEST_IN | GUEST_OUT | GUEST_LOCAL | WANv6_IN | WANv6_OUT | WANv6_LOCAL | LANv6_IN | LANv6_OUT | LANv6_LOCAL | GUESTv6_IN | GUESTv6_OUT | GUESTv6_LOCAL.")
	a.Describe(&r.Enabled, "Enabled controls whether the rule is active. Defaults to true.")
	a.Describe(&r.Logging, "Logging enables logging of packets that match this rule.")
	a.Describe(&r.IpSec, "IpSec matches on IPsec encapsulation: match-ipsec | match-none.")
	a.Describe(&r.SettingPreference, "SettingPreference controls whether the rule uses automatic or manual settings: auto | manual.")
	a.Describe(&r.ProtocolMatch, "ProtocolMatch groups the protocol / ICMP-type match selectors.")
	a.Describe(&r.Source, "Source groups the source-side match selectors.")
	a.Describe(&r.Destination, "Destination groups the destination-side match selectors.")
	a.Describe(&r.ConnectionState, "ConnectionState groups the conntrack-state match toggles.")
}

// FirewallRuleState is the persisted state: inputs plus controller-assigned fields.
type FirewallRuleState struct {
	FirewallRuleArgs
	// FirewallRuleId is the controller-assigned identifier (the UniFi `_id`).
	FirewallRuleId string `pulumi:"firewallRuleId"`
}

func (s *FirewallRuleState) Annotate(a infer.Annotator) {
	a.Describe(&s.FirewallRuleId, "FirewallRuleId is the controller-assigned identifier (the UniFi `_id`).")
}

// Annotate documents the resource. Must use a pointer receiver so the
// annotator can take the address of the resource and its fields.
func (r *FirewallRule) Annotate(a infer.Annotator) {
	a.Describe(&r, "A classic (per-ruleset) UniFi firewall rule. Rules are evaluated in order "+
		"within a ruleset, controlled by ruleIndex. Use firewall groups, networks, addresses, "+
		"ports and connection-state matching to permit or block traffic. Match selectors are "+
		"grouped into nested objects (protocolMatch, source, destination, connectionState).")
}

// toUnifi builds a go-unifi FirewallRule from inputs. id is empty on create.
func (a FirewallRuleArgs) toUnifi(id string) *unifi.FirewallRule {
	r := &unifi.FirewallRule{
		ID:        id,
		Name:      a.Name,
		RuleIndex: a.RuleIndex,
		Enabled:   derefOr(a.Enabled, true),
	}
	if a.Action != nil {
		r.Action = *a.Action
	}
	if a.Ruleset != nil {
		r.Ruleset = *a.Ruleset
	}

	// Protocol match.
	if pm := a.ProtocolMatch; pm != nil {
		if pm.Protocol != nil {
			r.Protocol = *pm.Protocol
		}
		if pm.ProtocolV6 != nil {
			r.ProtocolV6 = *pm.ProtocolV6
		}
		if pm.MatchExcepted != nil {
			r.ProtocolMatchExcepted = *pm.MatchExcepted
		}
		if pm.IcmpTypename != nil {
			r.ICMPTypename = *pm.IcmpTypename
		}
		if pm.Icmpv6Typename != nil {
			r.ICMPv6Typename = *pm.Icmpv6Typename
		}
	}

	// Source.
	if s := a.Source; s != nil {
		if s.Address != nil {
			r.SrcAddress = *s.Address
		}
		if s.AddressIpv6 != nil {
			r.SrcAddressIPV6 = *s.AddressIpv6
		}
		if s.Port != nil {
			r.SrcPort = *s.Port
		}
		if s.Mac != nil {
			r.SrcMACAddress = *s.Mac
		}
		if s.NetworkId != nil {
			r.SrcNetworkID = *s.NetworkId
		}
		if s.NetworkType != nil {
			r.SrcNetworkType = *s.NetworkType
		}
		if s.FirewallGroupIds != nil {
			r.SrcFirewallGroupIDs = s.FirewallGroupIds
		}
	}

	// Destination.
	if d := a.Destination; d != nil {
		if d.Address != nil {
			r.DstAddress = *d.Address
		}
		if d.AddressIpv6 != nil {
			r.DstAddressIPV6 = *d.AddressIpv6
		}
		if d.Port != nil {
			r.DstPort = *d.Port
		}
		if d.NetworkId != nil {
			r.DstNetworkID = *d.NetworkId
		}
		if d.NetworkType != nil {
			r.DstNetworkType = *d.NetworkType
		}
		if d.FirewallGroupIds != nil {
			r.DstFirewallGroupIDs = d.FirewallGroupIds
		}
	}

	// Connection state.
	if cs := a.ConnectionState; cs != nil {
		if cs.Established != nil {
			r.StateEstablished = *cs.Established
		}
		if cs.New != nil {
			r.StateNew = *cs.New
		}
		if cs.Related != nil {
			r.StateRelated = *cs.Related
		}
		if cs.Invalid != nil {
			r.StateInvalid = *cs.Invalid
		}
	}

	if a.Logging != nil {
		r.Logging = *a.Logging
	}
	if a.IpSec != nil {
		r.IPSec = *a.IpSec
	}
	if a.SettingPreference != nil {
		r.SettingPreference = *a.SettingPreference
	}

	return r
}

// firewallRuleStrPtr reflects a controller string, falling back to the prior
// input when empty.
func firewallRuleStrPtr(v string, prior *string) *string {
	if v != "" {
		return ptr(v)
	}
	return prior
}

// firewallRuleBoolPtr reflects a controller bool when the user set it or when it
// is true, otherwise leaves the optional input unset to avoid spurious diffs.
func firewallRuleBoolPtr(v bool, prior *bool) *bool {
	if v {
		return ptr(v)
	}
	return prior
}

// isZero reports whether no source member is set (so the group round-trips as nil).
func (s FirewallRuleSource) isZero() bool {
	return s.Address == nil && s.AddressIpv6 == nil && s.Port == nil && s.Mac == nil &&
		s.NetworkId == nil && s.NetworkType == nil && s.FirewallGroupIds == nil
}

// isZero reports whether no destination member is set (so the group round-trips as nil).
func (d FirewallRuleDestination) isZero() bool {
	return d.Address == nil && d.AddressIpv6 == nil && d.Port == nil &&
		d.NetworkId == nil && d.NetworkType == nil && d.FirewallGroupIds == nil
}

// firewallRuleProtocolMatchFrom reconstructs the protocolMatch group, returning
// nil when no member is set. MatchExcepted preserves the reflect-when-true bool
// semantics (emit when the controller value is true or a prior input was set).
func firewallRuleProtocolMatchFrom(u *unifi.FirewallRule, prior *FirewallRuleProtocolMatch) *FirewallRuleProtocolMatch {
	var p FirewallRuleProtocolMatch
	if prior != nil {
		p = *prior
	}
	pm := FirewallRuleProtocolMatch{
		Protocol:       firewallRuleStrPtr(u.Protocol, p.Protocol),
		ProtocolV6:     firewallRuleStrPtr(u.ProtocolV6, p.ProtocolV6),
		MatchExcepted:  firewallRuleBoolPtr(u.ProtocolMatchExcepted, p.MatchExcepted),
		IcmpTypename:   firewallRuleStrPtr(u.ICMPTypename, p.IcmpTypename),
		Icmpv6Typename: firewallRuleStrPtr(u.ICMPv6Typename, p.Icmpv6Typename),
	}
	if pm == (FirewallRuleProtocolMatch{}) {
		return nil
	}
	return &pm
}

// firewallRuleSourceFrom reconstructs the source group, returning nil when no
// member is set. The firewallGroupIds list keeps its len>0 ? controller : prior
// preservation.
func firewallRuleSourceFrom(u *unifi.FirewallRule, prior *FirewallRuleSource) *FirewallRuleSource {
	var p FirewallRuleSource
	if prior != nil {
		p = *prior
	}
	s := FirewallRuleSource{
		Address:     firewallRuleStrPtr(u.SrcAddress, p.Address),
		AddressIpv6: firewallRuleStrPtr(u.SrcAddressIPV6, p.AddressIpv6),
		Port:        firewallRuleStrPtr(u.SrcPort, p.Port),
		Mac:         firewallRuleStrPtr(u.SrcMACAddress, p.Mac),
		NetworkId:   firewallRuleStrPtr(u.SrcNetworkID, p.NetworkId),
		NetworkType: firewallRuleStrPtr(u.SrcNetworkType, p.NetworkType),
	}
	if len(u.SrcFirewallGroupIDs) > 0 {
		s.FirewallGroupIds = u.SrcFirewallGroupIDs
	} else {
		s.FirewallGroupIds = p.FirewallGroupIds
	}
	if s.isZero() {
		return nil
	}
	return &s
}

// firewallRuleDestinationFrom reconstructs the destination group, returning nil
// when no member is set. The firewallGroupIds list keeps its
// len>0 ? controller : prior preservation.
func firewallRuleDestinationFrom(u *unifi.FirewallRule, prior *FirewallRuleDestination) *FirewallRuleDestination {
	var p FirewallRuleDestination
	if prior != nil {
		p = *prior
	}
	d := FirewallRuleDestination{
		Address:     firewallRuleStrPtr(u.DstAddress, p.Address),
		AddressIpv6: firewallRuleStrPtr(u.DstAddressIPV6, p.AddressIpv6),
		Port:        firewallRuleStrPtr(u.DstPort, p.Port),
		NetworkId:   firewallRuleStrPtr(u.DstNetworkID, p.NetworkId),
		NetworkType: firewallRuleStrPtr(u.DstNetworkType, p.NetworkType),
	}
	if len(u.DstFirewallGroupIDs) > 0 {
		d.FirewallGroupIds = u.DstFirewallGroupIDs
	} else {
		d.FirewallGroupIds = p.FirewallGroupIds
	}
	if d.isZero() {
		return nil
	}
	return &d
}

// firewallRuleConnectionStateFrom reconstructs the connectionState group,
// returning nil when no member is set. All four bools preserve the
// reflect-when-true semantics (emit when the controller value is true or a prior
// input was set).
func firewallRuleConnectionStateFrom(u *unifi.FirewallRule, prior *FirewallRuleConnectionState) *FirewallRuleConnectionState {
	var p FirewallRuleConnectionState
	if prior != nil {
		p = *prior
	}
	cs := FirewallRuleConnectionState{
		Established: firewallRuleBoolPtr(u.StateEstablished, p.Established),
		New:         firewallRuleBoolPtr(u.StateNew, p.New),
		Related:     firewallRuleBoolPtr(u.StateRelated, p.Related),
		Invalid:     firewallRuleBoolPtr(u.StateInvalid, p.Invalid),
	}
	if cs == (FirewallRuleConnectionState{}) {
		return nil
	}
	return &cs
}

// firewallRuleStateFrom maps a controller FirewallRule back into resource state.
// prior holds the user inputs so unset optional fields are preserved across the
// round-trip.
func firewallRuleStateFrom(u *unifi.FirewallRule, prior FirewallRuleArgs) FirewallRuleState {
	args := FirewallRuleArgs{
		Name:      u.Name,
		RuleIndex: u.RuleIndex,
		Enabled:   ptr(u.Enabled),
	}
	args.Action = firewallRuleStrPtr(u.Action, prior.Action)
	args.Ruleset = firewallRuleStrPtr(u.Ruleset, prior.Ruleset)

	args.Logging = firewallRuleBoolPtr(u.Logging, prior.Logging)
	args.IpSec = firewallRuleStrPtr(u.IPSec, prior.IpSec)
	args.SettingPreference = firewallRuleStrPtr(u.SettingPreference, prior.SettingPreference)

	// Nested facets.
	args.ProtocolMatch = firewallRuleProtocolMatchFrom(u, prior.ProtocolMatch)
	args.Source = firewallRuleSourceFrom(u, prior.Source)
	args.Destination = firewallRuleDestinationFrom(u, prior.Destination)
	args.ConnectionState = firewallRuleConnectionStateFrom(u, prior.ConnectionState)

	return FirewallRuleState{FirewallRuleArgs: args, FirewallRuleId: u.ID}
}

// Create provisions a new firewall rule.
func (FirewallRule) Create(ctx context.Context, req infer.CreateRequest[FirewallRuleArgs]) (infer.CreateResponse[FirewallRuleState], error) {
	if req.DryRun {
		return infer.CreateResponse[FirewallRuleState]{Output: FirewallRuleState{FirewallRuleArgs: req.Inputs}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	created, err := cfg.Network().CreateFirewallRule(ctx, cfg.ResolvedSite(), req.Inputs.toUnifi(""))
	if err != nil {
		return infer.CreateResponse[FirewallRuleState]{}, err
	}
	return infer.CreateResponse[FirewallRuleState]{ID: created.ID, Output: firewallRuleStateFrom(created, req.Inputs)}, nil
}

// Read recovers state from the controller, enabling `pulumi import`.
func (FirewallRule) Read(ctx context.Context, req infer.ReadRequest[FirewallRuleArgs, FirewallRuleState]) (infer.ReadResponse[FirewallRuleArgs, FirewallRuleState], error) {
	cfg := infer.GetConfig[config.Config](ctx)
	u, err := cfg.Network().GetFirewallRule(ctx, cfg.ResolvedSite(), req.ID)
	if notFound(err) {
		return infer.ReadResponse[FirewallRuleArgs, FirewallRuleState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[FirewallRuleArgs, FirewallRuleState]{}, err
	}
	st := firewallRuleStateFrom(u, req.Inputs)
	return infer.ReadResponse[FirewallRuleArgs, FirewallRuleState]{ID: req.ID, Inputs: st.FirewallRuleArgs, State: st}, nil
}

// Update applies changed inputs in place.
func (FirewallRule) Update(ctx context.Context, req infer.UpdateRequest[FirewallRuleArgs, FirewallRuleState]) (infer.UpdateResponse[FirewallRuleState], error) {
	if req.DryRun {
		return infer.UpdateResponse[FirewallRuleState]{Output: FirewallRuleState{FirewallRuleArgs: req.Inputs, FirewallRuleId: req.ID}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	updated, err := cfg.Network().UpdateFirewallRule(ctx, cfg.ResolvedSite(), req.Inputs.toUnifi(req.ID))
	if err != nil {
		return infer.UpdateResponse[FirewallRuleState]{}, err
	}
	return infer.UpdateResponse[FirewallRuleState]{Output: firewallRuleStateFrom(updated, req.Inputs)}, nil
}

// Delete removes the firewall rule.
func (FirewallRule) Delete(ctx context.Context, req infer.DeleteRequest[FirewallRuleState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[config.Config](ctx)
	return infer.DeleteResponse{}, cfg.Network().DeleteFirewallRule(ctx, cfg.ResolvedSite(), req.ID)
}
