// SPDX-License-Identifier: Apache-2.0

package network

import "testing"

// TestFirewallZonePolicyRoundTrip builds a representative spread of inputs,
// maps them to the go-unifi FirewallZonePolicy, maps the result back, and
// asserts the important fields survive the round-trip.
func TestFirewallZonePolicyRoundTrip(t *testing.T) {
	args := FirewallZonePolicyArgs{
		Name:               "lan-to-wan-block",
		Action:             "BLOCK",
		Enabled:            ptr(true),
		Description:        ptr("block lab to internet"),
		Index:              ptr(2),
		Logging:            ptr(true),
		CreateAllowRespond: ptr(true),
		Matching: &FirewallZonePolicyMatchingArgs{
			IpVersion:             ptr("IPV4"),
			Protocol:              ptr("tcp_udp"),
			MatchOppositeProtocol: ptr(false),
			ConnectionStateType:   ptr("CUSTOM"),
			ConnectionStates:      []string{"NEW", "ESTABLISHED"},
			MatchIpSec:            ptr(true),
			MatchIpSecType:        ptr("MATCH_IP_SEC"),
		},
		Source: FirewallZonePolicySourceArgs{
			ZoneId:             "zone-lan",
			MatchingTarget:     ptr("NETWORK"),
			MatchingTargetType: ptr("OBJECT"),
			NetworkIds:         []string{"net-1", "net-2"},
			Port:               ptr("1000-2000"),
			PortMatchingType:   ptr("SPECIFIC"),
			MatchOppositeIps:   ptr(true),
		},
		Destination: FirewallZonePolicyDestinationArgs{
			ZoneId:             "zone-wan",
			MatchingTarget:     ptr("IP"),
			MatchingTargetType: ptr("SPECIFIC"),
			Ips:                []string{"8.8.8.8"},
			Port:               ptr("443"),
			PortMatchingType:   ptr("SPECIFIC"),
		},
		Schedule: &FirewallZonePolicyScheduleArgs{
			Mode:           ptr("EVERY_WEEK"),
			TimeAllDay:     ptr(false),
			TimeRangeStart: ptr("09:00"),
			TimeRangeEnd:   ptr("17:00"),
			RepeatOnDays:   []string{"mon", "tue", "wed"},
		},
	}

	u := args.toUnifi("")
	if u.ID != "" {
		t.Fatalf("expected empty ID on create, got %q", u.ID)
	}

	st := firewallZonePolicyStateFrom(u, args)

	if st.Name != "lan-to-wan-block" {
		t.Errorf("name: got %q", st.Name)
	}
	if st.Action != "BLOCK" {
		t.Errorf("action: got %q", st.Action)
	}
	if st.Enabled == nil || !*st.Enabled {
		t.Errorf("enabled did not survive round-trip: %v", st.Enabled)
	}
	if st.Description == nil || *st.Description != "block lab to internet" {
		t.Errorf("description did not survive round-trip: %v", st.Description)
	}
	if st.Index == nil || *st.Index != 2 {
		t.Errorf("index did not survive round-trip: %v", st.Index)
	}
	if st.CreateAllowRespond == nil || !*st.CreateAllowRespond {
		t.Errorf("createAllowRespond did not survive round-trip: %v", st.CreateAllowRespond)
	}

	// Matching.
	if st.Matching == nil {
		t.Fatalf("matching did not survive round-trip")
	}
	if st.Matching.IpVersion == nil || *st.Matching.IpVersion != "IPV4" {
		t.Errorf("matching.ipVersion did not survive round-trip: %v", st.Matching.IpVersion)
	}
	if st.Matching.Protocol == nil || *st.Matching.Protocol != "tcp_udp" {
		t.Errorf("matching.protocol did not survive round-trip: %v", st.Matching.Protocol)
	}
	if st.Matching.ConnectionStateType == nil || *st.Matching.ConnectionStateType != "CUSTOM" {
		t.Errorf("matching.connectionStateType did not survive round-trip: %v", st.Matching.ConnectionStateType)
	}
	if len(st.Matching.ConnectionStates) != 2 || st.Matching.ConnectionStates[0] != "NEW" {
		t.Errorf("matching.connectionStates did not survive round-trip: %v", st.Matching.ConnectionStates)
	}
	if st.Matching.MatchIpSec == nil || !*st.Matching.MatchIpSec {
		t.Errorf("matching.matchIpSec did not survive round-trip: %v", st.Matching.MatchIpSec)
	}
	if st.Matching.MatchIpSecType == nil || *st.Matching.MatchIpSecType != "MATCH_IP_SEC" {
		t.Errorf("matching.matchIpSecType did not survive round-trip: %v", st.Matching.MatchIpSecType)
	}

	// Source.
	if st.Source.ZoneId != "zone-lan" {
		t.Errorf("source.zoneId: got %q", st.Source.ZoneId)
	}
	if st.Source.MatchingTarget == nil || *st.Source.MatchingTarget != "NETWORK" {
		t.Errorf("source.matchingTarget did not survive round-trip: %v", st.Source.MatchingTarget)
	}
	if len(st.Source.NetworkIds) != 2 || st.Source.NetworkIds[1] != "net-2" {
		t.Errorf("source.networkIds did not survive round-trip: %v", st.Source.NetworkIds)
	}
	if st.Source.Port == nil || *st.Source.Port != "1000-2000" {
		t.Errorf("source.port did not survive round-trip: %v", st.Source.Port)
	}
	if st.Source.MatchOppositeIps == nil || !*st.Source.MatchOppositeIps {
		t.Errorf("source.matchOppositeIps did not survive round-trip: %v", st.Source.MatchOppositeIps)
	}

	// Destination.
	if st.Destination.ZoneId != "zone-wan" {
		t.Errorf("destination.zoneId: got %q", st.Destination.ZoneId)
	}
	if st.Destination.MatchingTarget == nil || *st.Destination.MatchingTarget != "IP" {
		t.Errorf("destination.matchingTarget did not survive round-trip: %v", st.Destination.MatchingTarget)
	}
	if len(st.Destination.Ips) != 1 || st.Destination.Ips[0] != "8.8.8.8" {
		t.Errorf("destination.ips did not survive round-trip: %v", st.Destination.Ips)
	}
	if st.Destination.Port == nil || *st.Destination.Port != "443" {
		t.Errorf("destination.port did not survive round-trip: %v", st.Destination.Port)
	}

	// Schedule.
	if st.Schedule == nil {
		t.Fatalf("schedule did not survive round-trip")
	}
	if st.Schedule.Mode == nil || *st.Schedule.Mode != "EVERY_WEEK" {
		t.Errorf("schedule.mode did not survive round-trip: %v", st.Schedule.Mode)
	}
	if st.Schedule.TimeRangeStart == nil || *st.Schedule.TimeRangeStart != "09:00" {
		t.Errorf("schedule.timeRangeStart did not survive round-trip: %v", st.Schedule.TimeRangeStart)
	}
	if len(st.Schedule.RepeatOnDays) != 3 || st.Schedule.RepeatOnDays[0] != "mon" {
		t.Errorf("schedule.repeatOnDays did not survive round-trip: %v", st.Schedule.RepeatOnDays)
	}
}

// TestFirewallZonePolicyDefaultsAndScheduleOmitted verifies create defaults and
// that an omitted schedule stays nil through the round-trip.
func TestFirewallZonePolicyDefaultsAndScheduleOmitted(t *testing.T) {
	args := FirewallZonePolicyArgs{
		Name:   "allow-all",
		Action: "ALLOW",
		Source: FirewallZonePolicySourceArgs{ZoneId: "zone-a"},
		Destination: FirewallZonePolicyDestinationArgs{
			ZoneId: "zone-b",
		},
	}

	u := args.toUnifi("")
	if !u.Enabled {
		t.Errorf("expected enabled to default to true")
	}
	// matchIpSec and matchOppositeProtocol are always-send bools: they must still
	// serialize (defaulting to false) even when the whole matching group is omitted.
	if u.MatchIPSec {
		t.Errorf("expected matchIpSec to default to false")
	}
	if u.MatchOppositeProtocol {
		t.Errorf("expected matchOppositeProtocol to default to false")
	}

	st := firewallZonePolicyStateFrom(u, args)
	if st.Schedule != nil {
		t.Errorf("expected schedule to remain nil when omitted, got %+v", st.Schedule)
	}
	// An unused matching group should round-trip as nil to avoid spurious diffs.
	if st.Matching != nil {
		t.Errorf("expected matching to remain nil when omitted, got %+v", st.Matching)
	}
	// An optional bool the user never set and the controller reports as false
	// should remain nil to avoid spurious diffs.
	if st.Logging != nil {
		t.Errorf("expected logging to remain nil, got %v", *st.Logging)
	}
}
