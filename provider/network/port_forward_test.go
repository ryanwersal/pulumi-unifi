// SPDX-License-Identifier: Apache-2.0

package network

import "testing"

// TestPortForwardRoundTrip builds a representative spread of inputs, maps them
// to the go-unifi PortForward, maps the result back, and asserts the important
// fields survive the round-trip.
func TestPortForwardRoundTrip(t *testing.T) {
	args := PortForwardArgs{
		Name:               ptr("ssh"),
		Enabled:            ptr(true),
		FwdPort:            ptr("22"),
		DstPort:            ptr("2222"),
		Fwd:                ptr("192.168.1.10"),
		Proto:              ptr("tcp"),
		Src:                ptr("203.0.113.0/24"),
		Log:                ptr(true),
		PfwdInterface:      ptr("wan"),
		SrcFirewallGroupId: ptr("abc123"),
		SrcLimitingEnabled: ptr(true),
		SrcLimitingType:    ptr("firewall_group"),
		DestinationIp:      ptr("198.51.100.5"),
		DestinationIps: []PortForwardDestinationIp{
			{DestinationIp: ptr("198.51.100.5"), Interface: ptr("wan")},
			{DestinationIp: ptr("198.51.100.6"), Interface: ptr("wan2")},
		},
	}

	u := args.toUnifi("pf-1")
	if u.ID != "pf-1" {
		t.Fatalf("ID = %q, want pf-1", u.ID)
	}
	if u.Name != "ssh" || u.FwdPort != "22" || u.DstPort != "2222" || u.Fwd != "192.168.1.10" {
		t.Fatalf("core fields not mapped: %+v", u)
	}
	if u.Proto != "tcp" || u.Src != "203.0.113.0/24" || !u.Log || !u.Enabled {
		t.Fatalf("proto/src/log/enabled not mapped: %+v", u)
	}
	if u.PfwdInterface != "wan" || u.SrcFirewallGroupID != "abc123" || !u.SrcLimitingEnabled || u.SrcLimitingType != "firewall_group" {
		t.Fatalf("src limiting fields not mapped: %+v", u)
	}
	if u.DestinationIP != "198.51.100.5" || len(u.DestinationIPs) != 2 {
		t.Fatalf("destination fields not mapped: %+v", u)
	}
	if u.DestinationIPs[1].DestinationIP != "198.51.100.6" || u.DestinationIPs[1].Interface != "wan2" {
		t.Fatalf("destination list element not mapped: %+v", u.DestinationIPs)
	}

	st := portForwardStateFrom(u, args)
	if st.PortForwardId != "pf-1" {
		t.Fatalf("PortForwardId = %q, want pf-1", st.PortForwardId)
	}
	if derefOr(st.Name, "") != "ssh" || derefOr(st.FwdPort, "") != "22" || derefOr(st.DstPort, "") != "2222" {
		t.Fatalf("name/ports not round-tripped: %+v", st.PortForwardArgs)
	}
	if derefOr(st.Fwd, "") != "192.168.1.10" || derefOr(st.Proto, "") != "tcp" || derefOr(st.Src, "") != "203.0.113.0/24" {
		t.Fatalf("fwd/proto/src not round-tripped: %+v", st.PortForwardArgs)
	}
	if !derefOr(st.Log, false) || !derefOr(st.Enabled, false) || !derefOr(st.SrcLimitingEnabled, false) {
		t.Fatalf("bool fields not round-tripped: %+v", st.PortForwardArgs)
	}
	if derefOr(st.SrcLimitingType, "") != "firewall_group" || derefOr(st.SrcFirewallGroupId, "") != "abc123" {
		t.Fatalf("src limiting not round-tripped: %+v", st.PortForwardArgs)
	}
	if derefOr(st.DestinationIp, "") != "198.51.100.5" || len(st.DestinationIps) != 2 {
		t.Fatalf("destination not round-tripped: %+v", st.PortForwardArgs)
	}
	if derefOr(st.DestinationIps[0].Interface, "") != "wan" {
		t.Fatalf("destination list element not round-tripped: %+v", st.DestinationIps)
	}
}

// TestPortForwardDefaults verifies the documented defaults are applied when the
// optional inputs are left unset.
func TestPortForwardDefaults(t *testing.T) {
	u := PortForwardArgs{}.toUnifi("")
	if u.Proto != "tcp_udp" {
		t.Fatalf("Proto default = %q, want tcp_udp", u.Proto)
	}
	if u.Src != "any" {
		t.Fatalf("Src default = %q, want any", u.Src)
	}
	if !u.Enabled {
		t.Fatalf("Enabled default = %v, want true", u.Enabled)
	}
	if u.Log {
		t.Fatalf("Log default = %v, want false", u.Log)
	}
}
