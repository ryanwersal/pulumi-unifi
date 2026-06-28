// SPDX-License-Identifier: Apache-2.0

package network

import (
	"testing"

	"github.com/filipowm/go-unifi/unifi"
)

func TestFirewallRuleRoundTrip(t *testing.T) {
	args := FirewallRuleArgs{
		Name:              "block-lan-to-iot",
		RuleIndex:         2010,
		Action:            ptr(FirewallRuleAction("drop")),
		Ruleset:           ptr("LAN_IN"),
		Enabled:           ptr(true),
		Logging:           ptr(true),
		IpSec:             ptr(FirewallRuleIpSec("match-none")),
		SettingPreference: ptr(FirewallRuleSettingPreference("manual")),

		ProtocolMatch: &FirewallRuleProtocolMatch{
			Protocol:      ptr("tcp_udp"),
			ProtocolV6:    ptr("tcp_udp"),
			MatchExcepted: ptr(true),
			IcmpTypename:  ptr("echo-request"),
		},

		Source: &FirewallRuleSource{
			Address:          ptr("192.168.1.0/24"),
			Port:             ptr("1024:65535"),
			Mac:              ptr("00:11:22:33:44:55"),
			NetworkType:      ptr(FirewallRuleNetworkType("NETv4")),
			FirewallGroupIds: []string{"fg-src1"},
		},

		Destination: &FirewallRuleDestination{
			Address:          ptr("192.168.50.0/24"),
			Port:             ptr("443"),
			NetworkId:        ptr("net-dst"),
			NetworkType:      ptr(FirewallRuleNetworkType("NETv4")),
			FirewallGroupIds: []string{"fg-dst1", "fg-dst2"},
		},

		ConnectionState: &FirewallRuleConnectionState{
			Established: ptr(true),
			New:         ptr(true),
			Related:     ptr(true),
			Invalid:     ptr(false),
		},
	}

	u := args.toUnifi("fr-123")
	if u.ID != "fr-123" {
		t.Fatalf("ID = %q, want fr-123", u.ID)
	}
	if u.Name != "block-lan-to-iot" {
		t.Fatalf("Name = %q", u.Name)
	}
	if u.RuleIndex != 2010 {
		t.Fatalf("RuleIndex = %d, want 2010", u.RuleIndex)
	}
	if u.Action != "drop" || u.Ruleset != "LAN_IN" {
		t.Fatalf("Action/Ruleset = %q/%q", u.Action, u.Ruleset)
	}
	if !u.Enabled {
		t.Fatalf("Enabled = false, want true")
	}
	if u.Protocol != "tcp_udp" || u.ProtocolV6 != "tcp_udp" {
		t.Fatalf("Protocol/ProtocolV6 = %q/%q", u.Protocol, u.ProtocolV6)
	}
	if !u.ProtocolMatchExcepted {
		t.Fatalf("ProtocolMatchExcepted = false, want true")
	}
	if u.ICMPTypename != "echo-request" {
		t.Fatalf("ICMPTypename = %q", u.ICMPTypename)
	}
	if u.SrcMACAddress != "00:11:22:33:44:55" {
		t.Fatalf("SrcMACAddress = %q", u.SrcMACAddress)
	}
	if u.DstNetworkID != "net-dst" {
		t.Fatalf("DstNetworkID = %q", u.DstNetworkID)
	}
	if len(u.SrcFirewallGroupIDs) != 1 || len(u.DstFirewallGroupIDs) != 2 {
		t.Fatalf("firewall group ids = %v / %v", u.SrcFirewallGroupIDs, u.DstFirewallGroupIDs)
	}
	if !u.StateEstablished || !u.StateNew || !u.StateRelated || u.StateInvalid {
		t.Fatalf("state flags wrong: %v %v %v %v", u.StateEstablished, u.StateNew, u.StateRelated, u.StateInvalid)
	}
	if !u.Logging || u.IPSec != "match-none" || u.SettingPreference != "manual" {
		t.Fatalf("logging/ipsec/pref = %v/%q/%q", u.Logging, u.IPSec, u.SettingPreference)
	}

	st := firewallRuleStateFrom(u, args)
	if st.FirewallRuleId != "fr-123" {
		t.Fatalf("FirewallRuleId = %q, want fr-123", st.FirewallRuleId)
	}
	if st.Name != "block-lan-to-iot" || st.RuleIndex != 2010 {
		t.Fatalf("state Name/RuleIndex = %q/%d", st.Name, st.RuleIndex)
	}
	if st.Action == nil || *st.Action != "drop" {
		t.Fatalf("state Action = %v", st.Action)
	}
	if st.Ruleset == nil || *st.Ruleset != "LAN_IN" {
		t.Fatalf("state Ruleset = %v", st.Ruleset)
	}

	// Protocol match group round-trip, including the reflect-when-true matchExcepted.
	if st.ProtocolMatch == nil {
		t.Fatal("protocolMatch group lost on round-trip")
	}
	if st.ProtocolMatch.Protocol == nil || *st.ProtocolMatch.Protocol != "tcp_udp" {
		t.Fatalf("state protocolMatch.protocol = %v", st.ProtocolMatch.Protocol)
	}
	if st.ProtocolMatch.MatchExcepted == nil || *st.ProtocolMatch.MatchExcepted != true {
		t.Fatalf("state protocolMatch.matchExcepted = %v, want true", st.ProtocolMatch.MatchExcepted)
	}
	if st.ProtocolMatch.IcmpTypename == nil || *st.ProtocolMatch.IcmpTypename != "echo-request" {
		t.Fatalf("state protocolMatch.icmpTypename = %v", st.ProtocolMatch.IcmpTypename)
	}

	// Source group round-trip.
	if st.Source == nil {
		t.Fatal("source group lost on round-trip")
	}
	if st.Source.Mac == nil || *st.Source.Mac != "00:11:22:33:44:55" {
		t.Fatalf("state source.mac = %v", st.Source.Mac)
	}
	if st.Source.Port == nil || *st.Source.Port != "1024:65535" {
		t.Fatalf("state source.port = %v", st.Source.Port)
	}
	if len(st.Source.FirewallGroupIds) != 1 || st.Source.FirewallGroupIds[0] != "fg-src1" {
		t.Fatalf("state source.firewallGroupIds = %v", st.Source.FirewallGroupIds)
	}

	// Destination group round-trip.
	if st.Destination == nil {
		t.Fatal("destination group lost on round-trip")
	}
	if st.Destination.NetworkId == nil || *st.Destination.NetworkId != "net-dst" {
		t.Fatalf("state destination.networkId = %v", st.Destination.NetworkId)
	}
	if len(st.Destination.FirewallGroupIds) != 2 {
		t.Fatalf("state destination.firewallGroupIds = %v", st.Destination.FirewallGroupIds)
	}

	// Connection state group round-trip. StateInvalid is explicitly false, set by
	// the user, so it must survive (reflect-when-true falls back to prior here).
	if st.ConnectionState == nil {
		t.Fatal("connectionState group lost on round-trip")
	}
	if st.ConnectionState.Established == nil || *st.ConnectionState.Established != true {
		t.Fatalf("state connectionState.established = %v", st.ConnectionState.Established)
	}
	if st.ConnectionState.New == nil || *st.ConnectionState.New != true {
		t.Fatalf("state connectionState.new = %v", st.ConnectionState.New)
	}
	if st.ConnectionState.Invalid == nil || *st.ConnectionState.Invalid != false {
		t.Fatalf("state connectionState.invalid = %v, want explicit false", st.ConnectionState.Invalid)
	}

	if st.IpSec == nil || *st.IpSec != "match-none" {
		t.Fatalf("state IpSec = %v", st.IpSec)
	}
}

func TestFirewallRuleDefaultsAndPriorPreserved(t *testing.T) {
	// Enabled defaults to true when unset.
	args := FirewallRuleArgs{Name: "r", RuleIndex: 4001}
	u := args.toUnifi("")
	if !u.Enabled {
		t.Fatalf("default Enabled = false, want true")
	}

	// Controller returns empty optional fields; prior inputs should be preserved,
	// including the prior values nested in the source group and the reflect-when-
	// true logging flag.
	prior := FirewallRuleArgs{
		Name:      "r",
		RuleIndex: 4001,
		Action:    ptr(FirewallRuleAction("accept")),
		Ruleset:   ptr("WAN_IN"),
		Logging:   ptr(true),
		Source: &FirewallRuleSource{
			FirewallGroupIds: []string{"keep-me"},
		},
	}
	st := firewallRuleStateFrom(&unifi.FirewallRule{ID: "x", Name: "r", RuleIndex: 4001}, prior)
	if st.Action == nil || *st.Action != "accept" {
		t.Fatalf("preserved Action = %v", st.Action)
	}
	if st.Ruleset == nil || *st.Ruleset != "WAN_IN" {
		t.Fatalf("preserved Ruleset = %v", st.Ruleset)
	}
	if st.Source == nil {
		t.Fatal("source group lost when only firewallGroupIds carried over from prior")
	}
	if len(st.Source.FirewallGroupIds) != 1 || st.Source.FirewallGroupIds[0] != "keep-me" {
		t.Fatalf("preserved source.firewallGroupIds = %v", st.Source.FirewallGroupIds)
	}
	if st.Logging == nil || *st.Logging != true {
		t.Fatalf("preserved Logging = %v", st.Logging)
	}
}

// TestFirewallRuleEmptyGroupsStayNil asserts that round-tripping a minimal rule
// does not synthesize non-nil facet groups out of controller zero values.
func TestFirewallRuleEmptyGroupsStayNil(t *testing.T) {
	u := FirewallRuleArgs{Name: "minimal", RuleIndex: 2001}.toUnifi("")
	u.ID = "fr-min"
	st := firewallRuleStateFrom(u, FirewallRuleArgs{Name: "minimal", RuleIndex: 2001})
	out := st.FirewallRuleArgs
	if out.ProtocolMatch != nil {
		t.Errorf("protocolMatch should be nil for a minimal rule, got %+v", out.ProtocolMatch)
	}
	if out.Source != nil {
		t.Errorf("source should be nil for a minimal rule, got %+v", out.Source)
	}
	if out.Destination != nil {
		t.Errorf("destination should be nil for a minimal rule, got %+v", out.Destination)
	}
	if out.ConnectionState != nil {
		t.Errorf("connectionState should be nil for a minimal rule, got %+v", out.ConnectionState)
	}
}
