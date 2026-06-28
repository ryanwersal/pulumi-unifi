// SPDX-License-Identifier: Apache-2.0

package network

import (
	"context"
	"fmt"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/ryanwersal/pulumi-unifi/provider/config"
)

// FirewallZonePolicyAction is the traffic disposition applied by a policy.
type FirewallZonePolicyAction string

const (
	FirewallZonePolicyActionAllow  FirewallZonePolicyAction = "ALLOW"
	FirewallZonePolicyActionBlock  FirewallZonePolicyAction = "BLOCK"
	FirewallZonePolicyActionReject FirewallZonePolicyAction = "REJECT"
)

func (FirewallZonePolicyAction) Values() []infer.EnumValue[FirewallZonePolicyAction] {
	return []infer.EnumValue[FirewallZonePolicyAction]{
		{Name: "Allow", Value: FirewallZonePolicyActionAllow, Description: "Permit matching traffic."},
		{Name: "Block", Value: FirewallZonePolicyActionBlock, Description: "Silently drop matching traffic."},
		{Name: "Reject", Value: FirewallZonePolicyActionReject, Description: "Drop matching traffic and send a rejection response."},
	}
}

// FirewallZonePolicyIpVersion restricts a policy to an L3 IP version.
type FirewallZonePolicyIpVersion string

const (
	FirewallZonePolicyIpVersionBoth FirewallZonePolicyIpVersion = "BOTH"
	FirewallZonePolicyIpVersionIpv4 FirewallZonePolicyIpVersion = "IPV4"
	FirewallZonePolicyIpVersionIpv6 FirewallZonePolicyIpVersion = "IPV6"
)

func (FirewallZonePolicyIpVersion) Values() []infer.EnumValue[FirewallZonePolicyIpVersion] {
	return []infer.EnumValue[FirewallZonePolicyIpVersion]{
		{Name: "Both", Value: FirewallZonePolicyIpVersionBoth, Description: "Match both IPv4 and IPv6 traffic."},
		{Name: "Ipv4", Value: FirewallZonePolicyIpVersionIpv4, Description: "Match IPv4 traffic only."},
		{Name: "Ipv6", Value: FirewallZonePolicyIpVersionIpv6, Description: "Match IPv6 traffic only."},
	}
}

// FirewallZonePolicyConnectionStateType selects connection-state matching mode.
type FirewallZonePolicyConnectionStateType string

const (
	FirewallZonePolicyConnectionStateTypeAll         FirewallZonePolicyConnectionStateType = "ALL"
	FirewallZonePolicyConnectionStateTypeRespondOnly FirewallZonePolicyConnectionStateType = "RESPOND_ONLY"
	FirewallZonePolicyConnectionStateTypeCustom      FirewallZonePolicyConnectionStateType = "CUSTOM"
)

func (FirewallZonePolicyConnectionStateType) Values() []infer.EnumValue[FirewallZonePolicyConnectionStateType] {
	return []infer.EnumValue[FirewallZonePolicyConnectionStateType]{
		{Name: "All", Value: FirewallZonePolicyConnectionStateTypeAll, Description: "Match all connection states."},
		{Name: "RespondOnly", Value: FirewallZonePolicyConnectionStateTypeRespondOnly, Description: "Match only return (responding) traffic."},
		{Name: "Custom", Value: FirewallZonePolicyConnectionStateTypeCustom, Description: "Match the explicit connection states listed in connectionStates."},
	}
}

// FirewallZonePolicy is the controlling (marker) struct for a UniFi zone-based
// firewall policy. Zone-based firewall policies are used by current UDM/UDM-SE
// firmware (the "Zone-Based Firewall" feature) and replace the legacy rule list.
type FirewallZonePolicy struct{}

// FirewallZonePolicySourceArgs describes the source matching criteria for a
// zone-based firewall policy.
type FirewallZonePolicySourceArgs struct {
	// ZoneId is the controller ID of the source firewall zone.
	ZoneId string `pulumi:"zoneId"`
	// MatchingTarget selects what the source matches: ANY | CLIENT | NETWORK | IP | MAC.
	MatchingTarget *string `pulumi:"matchingTarget,optional"`
	// MatchingTargetType refines the match: OBJECT (a saved group) | SPECIFIC (inline values).
	MatchingTargetType *string `pulumi:"matchingTargetType,optional"`
	// NetworkIds lists the source network IDs (when matchingTarget=NETWORK).
	NetworkIds []string `pulumi:"networkIds,optional"`
	// IpGroupId references a saved IP group as the source (when matchingTarget=IP, matchingTargetType=OBJECT).
	IpGroupId *string `pulumi:"ipGroupId,optional"`
	// Ips lists inline source IPv4 addresses (when matchingTarget=IP, matchingTargetType=SPECIFIC).
	Ips []string `pulumi:"ips,optional"`
	// Mac is a single source MAC address (when matchingTarget=MAC).
	Mac *string `pulumi:"mac,optional"`
	// Macs lists source MAC addresses.
	Macs []string `pulumi:"macs,optional"`
	// ClientMacs lists source client MAC addresses (when matchingTarget=CLIENT).
	ClientMacs []string `pulumi:"clientMacs,optional"`
	// MatchMac toggles MAC-based matching of the source.
	MatchMac *bool `pulumi:"matchMac,optional"`
	// Port is the source port or port range/list (e.g. "80", "1000-2000", "80,443").
	Port *string `pulumi:"port,optional"`
	// PortGroupId references a saved port group as the source ports.
	PortGroupId *string `pulumi:"portGroupId,optional"`
	// PortMatchingType selects how source ports match: ANY | SPECIFIC | OBJECT.
	PortMatchingType *string `pulumi:"portMatchingType,optional"`
	// MatchOppositeIps inverts (negates) the source IP match.
	MatchOppositeIps *bool `pulumi:"matchOppositeIps,optional"`
	// MatchOppositePorts inverts (negates) the source port match.
	MatchOppositePorts *bool `pulumi:"matchOppositePorts,optional"`
	// MatchOppositeNetworks inverts (negates) the source network match.
	MatchOppositeNetworks *bool `pulumi:"matchOppositeNetworks,optional"`
}

// FirewallZonePolicyDestinationArgs describes the destination matching criteria
// for a zone-based firewall policy.
type FirewallZonePolicyDestinationArgs struct {
	// ZoneId is the controller ID of the destination firewall zone.
	ZoneId string `pulumi:"zoneId"`
	// MatchingTarget selects what the destination matches: ANY | APP | APP_CATEGORY | IP | REGION | WEB.
	MatchingTarget *string `pulumi:"matchingTarget,optional"`
	// MatchingTargetType refines the match: ANY | OBJECT (a saved group) | SPECIFIC (inline values).
	MatchingTargetType *string `pulumi:"matchingTargetType,optional"`
	// IpGroupId references a saved IP group as the destination (when matchingTarget=IP, matchingTargetType=OBJECT).
	IpGroupId *string `pulumi:"ipGroupId,optional"`
	// Ips lists inline destination IPv4 addresses (when matchingTarget=IP, matchingTargetType=SPECIFIC).
	Ips []string `pulumi:"ips,optional"`
	// AppIds lists application IDs to match (when matchingTarget=APP).
	AppIds []string `pulumi:"appIds,optional"`
	// AppCategoryIds lists application category IDs to match (when matchingTarget=APP_CATEGORY).
	AppCategoryIds []string `pulumi:"appCategoryIds,optional"`
	// Regions lists geographic regions to match (when matchingTarget=REGION).
	Regions []string `pulumi:"regions,optional"`
	// WebDomains lists web domains to match (when matchingTarget=WEB).
	WebDomains []string `pulumi:"webDomains,optional"`
	// Port is the destination port or port range/list (e.g. "80", "1000-2000", "80,443").
	Port *string `pulumi:"port,optional"`
	// PortGroupId references a saved port group as the destination ports.
	PortGroupId *string `pulumi:"portGroupId,optional"`
	// PortMatchingType selects how destination ports match: ANY | SPECIFIC | OBJECT.
	PortMatchingType *string `pulumi:"portMatchingType,optional"`
	// MatchOppositeIps inverts (negates) the destination IP match.
	MatchOppositeIps *bool `pulumi:"matchOppositeIps,optional"`
	// MatchOppositePorts inverts (negates) the destination port match.
	MatchOppositePorts *bool `pulumi:"matchOppositePorts,optional"`
}

// FirewallZonePolicyScheduleArgs describes the optional temporal enforcement
// window for a zone-based firewall policy.
type FirewallZonePolicyScheduleArgs struct {
	// Mode selects the schedule pattern: ALWAYS | EVERY_DAY | EVERY_WEEK | ONE_TIME_ONLY | CUSTOM.
	Mode *string `pulumi:"mode,optional"`
	// TimeAllDay applies the policy for the entire day (ignores the time range).
	TimeAllDay *bool `pulumi:"timeAllDay,optional"`
	// TimeRangeStart is the daily start time in HH:MM (24-hour) format.
	TimeRangeStart *string `pulumi:"timeRangeStart,optional"`
	// TimeRangeEnd is the daily end time in HH:MM (24-hour) format.
	TimeRangeEnd *string `pulumi:"timeRangeEnd,optional"`
	// Date is a single enforcement date in YYYY-MM-DD format (mode=ONE_TIME_ONLY).
	Date *string `pulumi:"date,optional"`
	// DateStart is the range start date in YYYY-MM-DD format.
	DateStart *string `pulumi:"dateStart,optional"`
	// DateEnd is the range end date in YYYY-MM-DD format.
	DateEnd *string `pulumi:"dateEnd,optional"`
	// RepeatOnDays lists recurring days: mon | tue | wed | thu | fri | sat | sun (mode=EVERY_WEEK/CUSTOM).
	RepeatOnDays []string `pulumi:"repeatOnDays,optional"`
}

// FirewallZonePolicyMatchingArgs groups the flow-level traffic-match criteria
// (protocol, IP version, connection-state, and IPsec selection) that are not
// tied to either the source or destination zone.
type FirewallZonePolicyMatchingArgs struct {
	// IpVersion restricts the policy to an L3 version: BOTH | IPV4 | IPV6.
	IpVersion *FirewallZonePolicyIpVersion `pulumi:"ipVersion,optional"`
	// Protocol filters by IP protocol: all | tcp_udp | tcp | udp | icmp | icmpv6 | igmp | esp | ah | gre | ... (see UniFi docs).
	Protocol *string `pulumi:"protocol,optional"`
	// MatchOppositeProtocol inverts (negates) the protocol match.
	MatchOppositeProtocol *bool `pulumi:"matchOppositeProtocol,optional"`
	// ConnectionStateType selects connection-state matching: ALL | RESPOND_ONLY | CUSTOM.
	ConnectionStateType *FirewallZonePolicyConnectionStateType `pulumi:"connectionStateType,optional"`
	// ConnectionStates lists states to match when connectionStateType=CUSTOM: ESTABLISHED | NEW | RELATED | INVALID.
	ConnectionStates []string `pulumi:"connectionStates,optional"`
	// MatchIpSec enables matching on IPsec-encapsulated traffic.
	MatchIpSec *bool `pulumi:"matchIpSec,optional"`
	// MatchIpSecType selects IPsec matching mode: MATCH_IP_SEC | MATCH_NON_IP_SEC.
	MatchIpSecType *string `pulumi:"matchIpSecType,optional"`
}

// FirewallZonePolicyArgs are the user-supplied inputs for a zone-based firewall policy.
type FirewallZonePolicyArgs struct {
	// Name is the policy identifier shown in the controller.
	Name string `pulumi:"name"`
	// Action is the traffic disposition: ALLOW | BLOCK | REJECT.
	Action FirewallZonePolicyAction `pulumi:"action"`
	// Source is the source zone and matching criteria.
	Source FirewallZonePolicySourceArgs `pulumi:"source"`
	// Destination is the destination zone and matching criteria.
	Destination FirewallZonePolicyDestinationArgs `pulumi:"destination"`

	// Enabled controls whether the policy is active. Defaults to true.
	Enabled *bool `pulumi:"enabled,optional"`
	// Description is free-form documentation for the policy.
	Description *string `pulumi:"description,optional"`
	// Index is the policy priority/ordering rank (lower is evaluated first).
	Index *int `pulumi:"index,optional"`
	// Logging enables syslog logging for traffic matching this policy.
	Logging *bool `pulumi:"logging,optional"`
	// CreateAllowRespond automatically allows reverse-direction (return) traffic.
	CreateAllowRespond *bool `pulumi:"createAllowRespond,optional"`
	// Matching groups the flow-level traffic-match criteria (protocol, IP version,
	// connection-state, and IPsec selection).
	Matching *FirewallZonePolicyMatchingArgs `pulumi:"matching,optional"`
	// Schedule is the optional temporal enforcement window. When unset the policy is always active.
	Schedule *FirewallZonePolicyScheduleArgs `pulumi:"schedule,optional"`
}

// FirewallZonePolicyState is the persisted state: inputs plus controller-assigned fields.
type FirewallZonePolicyState struct {
	FirewallZonePolicyArgs
	// FirewallZonePolicyId is the controller-assigned identifier (the UniFi `_id`).
	FirewallZonePolicyId string `pulumi:"firewallZonePolicyId"`
}

// Annotate documents the resource. Must use a pointer receiver so the
// annotator can take the address of the resource and its fields.
func (p *FirewallZonePolicy) Annotate(a infer.Annotator) {
	a.Describe(&p, "A UniFi zone-based firewall policy (the Zone-Based Firewall feature on current "+
		"UDM/UDM-SE/UXG firmware). Each policy matches traffic flowing from a source zone to a "+
		"destination zone and applies an action. This resource requires a controller with the "+
		"zone-based firewall enabled; it is not available on older controllers that still use the "+
		"legacy firewall rule list.")
}

// Annotate documents the source matching fields.
func (s *FirewallZonePolicySourceArgs) Annotate(a infer.Annotator) {
	a.Describe(&s.ZoneId, "ZoneId is the controller ID of the source firewall zone.")
	a.Describe(&s.MatchingTarget, "What the source matches: ANY | CLIENT | NETWORK | IP | MAC.")
	a.Describe(&s.MatchingTargetType, "Refines the source match: OBJECT (saved group) | SPECIFIC (inline values).")
	a.Describe(&s.NetworkIds, "NetworkIds lists the source network IDs (when matchingTarget=NETWORK).")
	a.Describe(&s.IpGroupId, "IpGroupId references a saved IP group as the source (when matchingTarget=IP, matchingTargetType=OBJECT).")
	a.Describe(&s.Ips, "Ips lists inline source IPv4 addresses (when matchingTarget=IP, matchingTargetType=SPECIFIC).")
	a.Describe(&s.Mac, "Mac is a single source MAC address (when matchingTarget=MAC).")
	a.Describe(&s.Macs, "Macs lists source MAC addresses.")
	a.Describe(&s.ClientMacs, "ClientMacs lists source client MAC addresses (when matchingTarget=CLIENT).")
	a.Describe(&s.MatchMac, "MatchMac toggles MAC-based matching of the source.")
	a.Describe(&s.Port, "Port is the source port or port range/list (e.g. \"80\", \"1000-2000\", \"80,443\").")
	a.Describe(&s.PortGroupId, "PortGroupId references a saved port group as the source ports.")
	a.Describe(&s.PortMatchingType, "PortMatchingType selects how source ports match: ANY | SPECIFIC | OBJECT.")
	a.Describe(&s.MatchOppositeIps, "MatchOppositeIps inverts (negates) the source IP match.")
	a.Describe(&s.MatchOppositePorts, "MatchOppositePorts inverts (negates) the source port match.")
	a.Describe(&s.MatchOppositeNetworks, "MatchOppositeNetworks inverts (negates) the source network match.")
}

// Annotate documents the destination matching fields.
func (d *FirewallZonePolicyDestinationArgs) Annotate(a infer.Annotator) {
	a.Describe(&d.ZoneId, "ZoneId is the controller ID of the destination firewall zone.")
	a.Describe(&d.MatchingTarget, "What the destination matches: ANY | APP | APP_CATEGORY | IP | REGION | WEB.")
	a.Describe(&d.MatchingTargetType, "Refines the destination match: ANY | OBJECT (saved group) | SPECIFIC (inline values).")
	a.Describe(&d.IpGroupId, "IpGroupId references a saved IP group as the destination (when matchingTarget=IP, matchingTargetType=OBJECT).")
	a.Describe(&d.Ips, "Ips lists inline destination IPv4 addresses (when matchingTarget=IP, matchingTargetType=SPECIFIC).")
	a.Describe(&d.AppIds, "AppIds lists application IDs to match (when matchingTarget=APP).")
	a.Describe(&d.AppCategoryIds, "AppCategoryIds lists application category IDs to match (when matchingTarget=APP_CATEGORY).")
	a.Describe(&d.Regions, "Regions lists geographic regions to match (when matchingTarget=REGION).")
	a.Describe(&d.WebDomains, "WebDomains lists web domains to match (when matchingTarget=WEB).")
	a.Describe(&d.Port, "Port is the destination port or port range/list (e.g. \"80\", \"1000-2000\", \"80,443\").")
	a.Describe(&d.PortGroupId, "PortGroupId references a saved port group as the destination ports.")
	a.Describe(&d.PortMatchingType, "PortMatchingType selects how destination ports match: ANY | SPECIFIC | OBJECT.")
	a.Describe(&d.MatchOppositeIps, "MatchOppositeIps inverts (negates) the destination IP match.")
	a.Describe(&d.MatchOppositePorts, "MatchOppositePorts inverts (negates) the destination port match.")
}

// Annotate documents the schedule fields.
func (s *FirewallZonePolicyScheduleArgs) Annotate(a infer.Annotator) {
	a.Describe(&s.Mode, "Schedule pattern: ALWAYS | EVERY_DAY | EVERY_WEEK | ONE_TIME_ONLY | CUSTOM.")
	a.Describe(&s.TimeAllDay, "TimeAllDay applies the policy for the entire day (ignores the time range).")
	a.Describe(&s.TimeRangeStart, "TimeRangeStart is the daily start time in HH:MM (24-hour) format.")
	a.Describe(&s.TimeRangeEnd, "TimeRangeEnd is the daily end time in HH:MM (24-hour) format.")
	a.Describe(&s.Date, "Date is a single enforcement date in YYYY-MM-DD format (mode=ONE_TIME_ONLY).")
	a.Describe(&s.DateStart, "DateStart is the range start date in YYYY-MM-DD format.")
	a.Describe(&s.DateEnd, "DateEnd is the range end date in YYYY-MM-DD format.")
	a.Describe(&s.RepeatOnDays, "RepeatOnDays lists recurring days: mon | tue | wed | thu | fri | sat | sun (mode=EVERY_WEEK/CUSTOM).")
}

// Annotate documents the matching fields.
func (m *FirewallZonePolicyMatchingArgs) Annotate(a infer.Annotator) {
	a.Describe(&m.IpVersion, "L3 version the policy matches: BOTH | IPV4 | IPV6.")
	a.Describe(&m.Protocol, "Protocol filters by IP protocol: all | tcp_udp | tcp | udp | icmp | icmpv6 | igmp | esp | ah | gre | ... (see UniFi docs).")
	a.Describe(&m.MatchOppositeProtocol, "MatchOppositeProtocol inverts (negates) the protocol match.")
	a.Describe(&m.ConnectionStateType, "Connection-state matching: ALL | RESPOND_ONLY | CUSTOM.")
	a.Describe(&m.ConnectionStates, "ConnectionStates lists states to match when connectionStateType=CUSTOM: ESTABLISHED | NEW | RELATED | INVALID.")
	a.Describe(&m.MatchIpSec, "MatchIpSec enables matching on IPsec-encapsulated traffic.")
	a.Describe(&m.MatchIpSecType, "IPsec matching mode: MATCH_IP_SEC | MATCH_NON_IP_SEC.")
}

// Annotate documents the policy input fields.
func (p *FirewallZonePolicyArgs) Annotate(a infer.Annotator) {
	a.Describe(&p.Name, "Name is the policy identifier shown in the controller.")
	a.Describe(&p.Action, "Action is the traffic disposition: ALLOW | BLOCK | REJECT.")
	a.Describe(&p.Source, "Source is the source zone and matching criteria.")
	a.Describe(&p.Destination, "Destination is the destination zone and matching criteria.")
	a.Describe(&p.Enabled, "Enabled controls whether the policy is active. Defaults to true.")
	a.SetDefault(&p.Enabled, true)
	a.Describe(&p.Description, "Description is free-form documentation for the policy.")
	a.Describe(&p.Index, "Index is the policy priority/ordering rank (lower is evaluated first).")
	a.Describe(&p.Logging, "Logging enables syslog logging for traffic matching this policy.")
	a.Describe(&p.CreateAllowRespond, "CreateAllowRespond automatically allows reverse-direction (return) traffic.")
	a.Describe(&p.Matching, "Matching groups the flow-level traffic-match criteria (protocol, IP version, connection-state, and IPsec selection).")
	a.Describe(&p.Schedule, "Schedule is the optional temporal enforcement window. When unset the policy is always active.")
}

// Annotate documents the controller-assigned output fields.
func (s *FirewallZonePolicyState) Annotate(a infer.Annotator) {
	a.Describe(&s.FirewallZonePolicyId, "FirewallZonePolicyId is the controller-assigned identifier (the UniFi `_id`).")
}

// toUnifi builds a go-unifi FirewallZonePolicy from inputs. id is empty on create.
func (a FirewallZonePolicyArgs) toUnifi(id string) *unifi.FirewallZonePolicy {
	p := &unifi.FirewallZonePolicy{
		ID:                 id,
		Name:               a.Name,
		Action:             string(a.Action),
		Enabled:            derefOr(a.Enabled, true),
		Logging:            derefOr(a.Logging, false),
		CreateAllowRespond: derefOr(a.CreateAllowRespond, false),
		Source:             a.Source.toUnifi(),
		Destination:        a.Destination.toUnifi(),
	}
	if a.Description != nil {
		p.Description = *a.Description
	}
	if a.Index != nil {
		p.Index = *a.Index
	}

	// Matching. matchIpSec and matchOppositeProtocol have no upstream omitempty
	// and must be sent unconditionally (controller default false); read them from
	// the matching group when present, otherwise emit the default even when the
	// whole group is omitted. The remaining members are plain optionals.
	var m FirewallZonePolicyMatchingArgs
	if a.Matching != nil {
		m = *a.Matching
	}
	p.MatchIPSec = derefOr(m.MatchIpSec, false)
	p.MatchOppositeProtocol = derefOr(m.MatchOppositeProtocol, false)
	if m.IpVersion != nil {
		p.IPVersion = string(*m.IpVersion)
	}
	if m.Protocol != nil {
		p.Protocol = *m.Protocol
	}
	if m.ConnectionStateType != nil {
		p.ConnectionStateType = string(*m.ConnectionStateType)
	}
	if m.ConnectionStates != nil {
		p.ConnectionStates = m.ConnectionStates
	}
	if m.MatchIpSecType != nil {
		p.MatchIPSecType = *m.MatchIpSecType
	}

	if a.Schedule != nil {
		p.Schedule = a.Schedule.toUnifi()
	}
	return p
}

// toUnifi builds the upstream source matching struct.
func (s FirewallZonePolicySourceArgs) toUnifi() unifi.FirewallZonePolicySource {
	out := unifi.FirewallZonePolicySource{
		ZoneID:                s.ZoneId,
		MatchMAC:              derefOr(s.MatchMac, false),
		MatchOppositeIPs:      derefOr(s.MatchOppositeIps, false),
		MatchOppositePorts:    derefOr(s.MatchOppositePorts, false),
		MatchOppositeNetworks: derefOr(s.MatchOppositeNetworks, false),
	}
	if s.MatchingTarget != nil {
		out.MatchingTarget = *s.MatchingTarget
	}
	if s.MatchingTargetType != nil {
		out.MatchingTargetType = *s.MatchingTargetType
	}
	if s.NetworkIds != nil {
		out.NetworkIDs = s.NetworkIds
	}
	if s.IpGroupId != nil {
		out.IPGroupID = *s.IpGroupId
	}
	if s.Ips != nil {
		out.IPs = s.Ips
	}
	if s.Mac != nil {
		out.MAC = *s.Mac
	}
	if s.Macs != nil {
		out.MACs = s.Macs
	}
	if s.ClientMacs != nil {
		out.ClientMACs = s.ClientMacs
	}
	if s.Port != nil {
		out.Port = *s.Port
	}
	if s.PortGroupId != nil {
		out.PortGroupID = *s.PortGroupId
	}
	if s.PortMatchingType != nil {
		out.PortMatchingType = *s.PortMatchingType
	}
	return out
}

// toUnifi builds the upstream destination matching struct.
func (d FirewallZonePolicyDestinationArgs) toUnifi() unifi.FirewallZonePolicyDestination {
	out := unifi.FirewallZonePolicyDestination{
		ZoneID:             d.ZoneId,
		MatchOppositeIPs:   derefOr(d.MatchOppositeIps, false),
		MatchOppositePorts: derefOr(d.MatchOppositePorts, false),
	}
	if d.MatchingTarget != nil {
		out.MatchingTarget = *d.MatchingTarget
	}
	if d.MatchingTargetType != nil {
		out.MatchingTargetType = *d.MatchingTargetType
	}
	if d.IpGroupId != nil {
		out.IPGroupID = *d.IpGroupId
	}
	if d.Ips != nil {
		out.IPs = d.Ips
	}
	if d.AppIds != nil {
		out.AppIDs = d.AppIds
	}
	if d.AppCategoryIds != nil {
		out.AppCategoryIDs = d.AppCategoryIds
	}
	if d.Regions != nil {
		out.Regions = d.Regions
	}
	if d.WebDomains != nil {
		out.WebDomains = d.WebDomains
	}
	if d.Port != nil {
		out.Port = *d.Port
	}
	if d.PortGroupId != nil {
		out.PortGroupID = *d.PortGroupId
	}
	if d.PortMatchingType != nil {
		out.PortMatchingType = *d.PortMatchingType
	}
	return out
}

// toUnifi builds the upstream schedule struct.
func (s FirewallZonePolicyScheduleArgs) toUnifi() unifi.FirewallZonePolicySchedule {
	out := unifi.FirewallZonePolicySchedule{
		TimeAllDay: derefOr(s.TimeAllDay, false),
	}
	if s.Mode != nil {
		out.Mode = *s.Mode
	}
	if s.TimeRangeStart != nil {
		out.TimeRangeStart = *s.TimeRangeStart
	}
	if s.TimeRangeEnd != nil {
		out.TimeRangeEnd = *s.TimeRangeEnd
	}
	if s.Date != nil {
		out.Date = *s.Date
	}
	if s.DateStart != nil {
		out.DateStart = *s.DateStart
	}
	if s.DateEnd != nil {
		out.DateEnd = *s.DateEnd
	}
	if s.RepeatOnDays != nil {
		out.RepeatOnDays = s.RepeatOnDays
	}
	return out
}

// fzpStrPtr reflects a controller string, falling back to the prior input when empty.
func fzpStrPtr(v string, prior *string) *string {
	if v != "" {
		return ptr(v)
	}
	return prior
}

// fzpIntPtr reflects a controller int, falling back to the prior input when zero.
func fzpIntPtr(v int, prior *int) *int {
	if v != 0 {
		return ptr(v)
	}
	return prior
}

// fzpBoolPtr reflects a controller bool when the user set it or when it is true,
// otherwise leaves the optional input unset to avoid spurious diffs.
func fzpBoolPtr(v bool, prior *bool) *bool {
	if v {
		return ptr(v)
	}
	return prior
}

// fzpStrSlice reflects a controller string slice, preserving the prior input when empty.
func fzpStrSlice(v []string, prior []string) []string {
	if len(v) > 0 {
		return v
	}
	return prior
}

// firewallZonePolicySourceFrom maps a controller source back into resource inputs.
func firewallZonePolicySourceFrom(u unifi.FirewallZonePolicySource, prior FirewallZonePolicySourceArgs) FirewallZonePolicySourceArgs {
	return FirewallZonePolicySourceArgs{
		ZoneId:                u.ZoneID,
		MatchingTarget:        fzpStrPtr(u.MatchingTarget, prior.MatchingTarget),
		MatchingTargetType:    fzpStrPtr(u.MatchingTargetType, prior.MatchingTargetType),
		NetworkIds:            fzpStrSlice(u.NetworkIDs, prior.NetworkIds),
		IpGroupId:             fzpStrPtr(u.IPGroupID, prior.IpGroupId),
		Ips:                   fzpStrSlice(u.IPs, prior.Ips),
		Mac:                   fzpStrPtr(u.MAC, prior.Mac),
		Macs:                  fzpStrSlice(u.MACs, prior.Macs),
		ClientMacs:            fzpStrSlice(u.ClientMACs, prior.ClientMacs),
		MatchMac:              fzpBoolPtr(u.MatchMAC, prior.MatchMac),
		Port:                  fzpStrPtr(u.Port, prior.Port),
		PortGroupId:           fzpStrPtr(u.PortGroupID, prior.PortGroupId),
		PortMatchingType:      fzpStrPtr(u.PortMatchingType, prior.PortMatchingType),
		MatchOppositeIps:      fzpBoolPtr(u.MatchOppositeIPs, prior.MatchOppositeIps),
		MatchOppositePorts:    fzpBoolPtr(u.MatchOppositePorts, prior.MatchOppositePorts),
		MatchOppositeNetworks: fzpBoolPtr(u.MatchOppositeNetworks, prior.MatchOppositeNetworks),
	}
}

// firewallZonePolicyDestinationFrom maps a controller destination back into resource inputs.
func firewallZonePolicyDestinationFrom(u unifi.FirewallZonePolicyDestination, prior FirewallZonePolicyDestinationArgs) FirewallZonePolicyDestinationArgs {
	return FirewallZonePolicyDestinationArgs{
		ZoneId:             u.ZoneID,
		MatchingTarget:     fzpStrPtr(u.MatchingTarget, prior.MatchingTarget),
		MatchingTargetType: fzpStrPtr(u.MatchingTargetType, prior.MatchingTargetType),
		IpGroupId:          fzpStrPtr(u.IPGroupID, prior.IpGroupId),
		Ips:                fzpStrSlice(u.IPs, prior.Ips),
		AppIds:             fzpStrSlice(u.AppIDs, prior.AppIds),
		AppCategoryIds:     fzpStrSlice(u.AppCategoryIDs, prior.AppCategoryIds),
		Regions:            fzpStrSlice(u.Regions, prior.Regions),
		WebDomains:         fzpStrSlice(u.WebDomains, prior.WebDomains),
		Port:               fzpStrPtr(u.Port, prior.Port),
		PortGroupId:        fzpStrPtr(u.PortGroupID, prior.PortGroupId),
		PortMatchingType:   fzpStrPtr(u.PortMatchingType, prior.PortMatchingType),
		MatchOppositeIps:   fzpBoolPtr(u.MatchOppositeIPs, prior.MatchOppositeIps),
		MatchOppositePorts: fzpBoolPtr(u.MatchOppositePorts, prior.MatchOppositePorts),
	}
}

// firewallZonePolicyScheduleFrom maps a controller schedule back into resource inputs.
func firewallZonePolicyScheduleFrom(u unifi.FirewallZonePolicySchedule, prior *FirewallZonePolicyScheduleArgs) *FirewallZonePolicyScheduleArgs {
	if prior == nil && u.Mode == "" && !u.TimeAllDay && u.TimeRangeStart == "" && u.TimeRangeEnd == "" &&
		u.Date == "" && u.DateStart == "" && u.DateEnd == "" && len(u.RepeatOnDays) == 0 {
		return nil
	}
	var priorVal FirewallZonePolicyScheduleArgs
	if prior != nil {
		priorVal = *prior
	}
	return &FirewallZonePolicyScheduleArgs{
		Mode:           fzpStrPtr(u.Mode, priorVal.Mode),
		TimeAllDay:     fzpBoolPtr(u.TimeAllDay, priorVal.TimeAllDay),
		TimeRangeStart: fzpStrPtr(u.TimeRangeStart, priorVal.TimeRangeStart),
		TimeRangeEnd:   fzpStrPtr(u.TimeRangeEnd, priorVal.TimeRangeEnd),
		Date:           fzpStrPtr(u.Date, priorVal.Date),
		DateStart:      fzpStrPtr(u.DateStart, priorVal.DateStart),
		DateEnd:        fzpStrPtr(u.DateEnd, priorVal.DateEnd),
		RepeatOnDays:   fzpStrSlice(u.RepeatOnDays, priorVal.RepeatOnDays),
	}
}

// isZero reports whether no matching member is set (so the group round-trips as nil).
func (m FirewallZonePolicyMatchingArgs) isZero() bool {
	return m.IpVersion == nil && m.Protocol == nil && m.MatchOppositeProtocol == nil &&
		m.ConnectionStateType == nil && len(m.ConnectionStates) == 0 && m.MatchIpSec == nil &&
		m.MatchIpSecType == nil
}

// firewallZonePolicyMatchingFrom reconstructs the matching group, returning nil
// when no member is set so an unused group round-trips as nil.
func firewallZonePolicyMatchingFrom(u *unifi.FirewallZonePolicy, prior *FirewallZonePolicyMatchingArgs) *FirewallZonePolicyMatchingArgs {
	var p FirewallZonePolicyMatchingArgs
	if prior != nil {
		p = *prior
	}
	m := FirewallZonePolicyMatchingArgs{
		Protocol:              fzpStrPtr(u.Protocol, p.Protocol),
		MatchOppositeProtocol: fzpBoolPtr(u.MatchOppositeProtocol, p.MatchOppositeProtocol),
		ConnectionStates:      fzpStrSlice(u.ConnectionStates, p.ConnectionStates),
		MatchIpSec:            fzpBoolPtr(u.MatchIPSec, p.MatchIpSec),
		MatchIpSecType:        fzpStrPtr(u.MatchIPSecType, p.MatchIpSecType),
	}
	// Enum fields: reflect the controller value when set, else keep prior.
	if u.IPVersion != "" {
		m.IpVersion = ptr(FirewallZonePolicyIpVersion(u.IPVersion))
	} else {
		m.IpVersion = p.IpVersion
	}
	if u.ConnectionStateType != "" {
		m.ConnectionStateType = ptr(FirewallZonePolicyConnectionStateType(u.ConnectionStateType))
	} else {
		m.ConnectionStateType = p.ConnectionStateType
	}
	if m.isZero() {
		return nil
	}
	return &m
}

// firewallZonePolicyStateFrom maps a controller FirewallZonePolicy back into resource state.
// prior holds the user inputs so unset optional fields are preserved across the round-trip.
func firewallZonePolicyStateFrom(u *unifi.FirewallZonePolicy, prior FirewallZonePolicyArgs) FirewallZonePolicyState {
	args := FirewallZonePolicyArgs{
		Name:               u.Name,
		Action:             FirewallZonePolicyAction(u.Action),
		Enabled:            fzpBoolPtr(u.Enabled, prior.Enabled),
		Description:        fzpStrPtr(u.Description, prior.Description),
		Index:              fzpIntPtr(u.Index, prior.Index),
		Logging:            fzpBoolPtr(u.Logging, prior.Logging),
		CreateAllowRespond: fzpBoolPtr(u.CreateAllowRespond, prior.CreateAllowRespond),
		Source:             firewallZonePolicySourceFrom(u.Source, prior.Source),
		Destination:        firewallZonePolicyDestinationFrom(u.Destination, prior.Destination),
		Matching:           firewallZonePolicyMatchingFrom(u, prior.Matching),
		Schedule:           firewallZonePolicyScheduleFrom(u.Schedule, prior.Schedule),
	}
	return FirewallZonePolicyState{FirewallZonePolicyArgs: args, FirewallZonePolicyId: u.ID}
}

// Create provisions a new zone-based firewall policy.
func (FirewallZonePolicy) Create(ctx context.Context, req infer.CreateRequest[FirewallZonePolicyArgs]) (infer.CreateResponse[FirewallZonePolicyState], error) {
	if req.DryRun {
		return infer.CreateResponse[FirewallZonePolicyState]{Output: FirewallZonePolicyState{FirewallZonePolicyArgs: req.Inputs}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	created, err := cfg.Network().CreateFirewallZonePolicy(ctx, cfg.ResolvedSite(), req.Inputs.toUnifi(""))
	if err != nil {
		return infer.CreateResponse[FirewallZonePolicyState]{}, wrap(fmt.Sprintf("create firewall zone policy %q (site %q)", req.Inputs.Name, cfg.ResolvedSite()), err)
	}
	return infer.CreateResponse[FirewallZonePolicyState]{ID: created.ID, Output: firewallZonePolicyStateFrom(created, req.Inputs)}, nil
}

// Read recovers state from the controller, enabling `pulumi import`.
func (FirewallZonePolicy) Read(ctx context.Context, req infer.ReadRequest[FirewallZonePolicyArgs, FirewallZonePolicyState]) (infer.ReadResponse[FirewallZonePolicyArgs, FirewallZonePolicyState], error) {
	cfg := infer.GetConfig[config.Config](ctx)
	p, err := cfg.Network().GetFirewallZonePolicy(ctx, cfg.ResolvedSite(), req.ID)
	if notFound(err) {
		return infer.ReadResponse[FirewallZonePolicyArgs, FirewallZonePolicyState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[FirewallZonePolicyArgs, FirewallZonePolicyState]{}, wrap(fmt.Sprintf("read firewall zone policy %q (site %q)", req.ID, cfg.ResolvedSite()), err)
	}
	st := firewallZonePolicyStateFrom(p, req.Inputs)
	return infer.ReadResponse[FirewallZonePolicyArgs, FirewallZonePolicyState]{ID: req.ID, Inputs: st.FirewallZonePolicyArgs, State: st}, nil
}

// Update applies changed inputs in place.
func (FirewallZonePolicy) Update(ctx context.Context, req infer.UpdateRequest[FirewallZonePolicyArgs, FirewallZonePolicyState]) (infer.UpdateResponse[FirewallZonePolicyState], error) {
	if req.DryRun {
		return infer.UpdateResponse[FirewallZonePolicyState]{Output: FirewallZonePolicyState{FirewallZonePolicyArgs: req.Inputs, FirewallZonePolicyId: req.ID}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	updated, err := cfg.Network().UpdateFirewallZonePolicy(ctx, cfg.ResolvedSite(), req.Inputs.toUnifi(req.ID))
	if err != nil {
		return infer.UpdateResponse[FirewallZonePolicyState]{}, wrap(fmt.Sprintf("update firewall zone policy %q (site %q)", req.ID, cfg.ResolvedSite()), err)
	}
	return infer.UpdateResponse[FirewallZonePolicyState]{Output: firewallZonePolicyStateFrom(updated, req.Inputs)}, nil
}

// Delete removes the zone-based firewall policy.
func (FirewallZonePolicy) Delete(ctx context.Context, req infer.DeleteRequest[FirewallZonePolicyState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[config.Config](ctx)
	err := cfg.Network().DeleteFirewallZonePolicy(ctx, cfg.ResolvedSite(), req.ID)
	if notFound(err) {
		return infer.DeleteResponse{}, nil
	}
	return infer.DeleteResponse{}, wrap(fmt.Sprintf("delete firewall zone policy %q (site %q)", req.ID, cfg.ResolvedSite()), err)
}
