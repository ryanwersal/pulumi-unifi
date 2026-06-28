// SPDX-License-Identifier: Apache-2.0

package network

import (
	"testing"

	"github.com/filipowm/go-unifi/unifi"
)

func TestFirewallGroupRoundTrip(t *testing.T) {
	args := FirewallGroupArgs{
		Name:         "blocked-ips",
		GroupType:    ptr(FirewallGroupTypeAddressGroup),
		GroupMembers: []string{"1.2.3.4", "10.0.0.0/24"},
	}

	u := args.toUnifi("fg-123")
	if u.ID != "fg-123" {
		t.Fatalf("ID = %q, want fg-123", u.ID)
	}
	if u.Name != "blocked-ips" {
		t.Fatalf("Name = %q, want blocked-ips", u.Name)
	}
	if u.GroupType != "address-group" {
		t.Fatalf("GroupType = %q, want address-group", u.GroupType)
	}
	if len(u.GroupMembers) != 2 || u.GroupMembers[0] != "1.2.3.4" || u.GroupMembers[1] != "10.0.0.0/24" {
		t.Fatalf("GroupMembers = %v, want [1.2.3.4 10.0.0.0/24]", u.GroupMembers)
	}

	st := firewallGroupStateFrom(u, args)
	if st.FirewallGroupId != "fg-123" {
		t.Fatalf("FirewallGroupId = %q, want fg-123", st.FirewallGroupId)
	}
	if st.Name != "blocked-ips" {
		t.Fatalf("state Name = %q, want blocked-ips", st.Name)
	}
	if st.GroupType == nil || *st.GroupType != "address-group" {
		t.Fatalf("state GroupType = %v, want address-group", st.GroupType)
	}
	if len(st.GroupMembers) != 2 {
		t.Fatalf("state GroupMembers = %v, want 2 entries", st.GroupMembers)
	}
}

func TestFirewallGroupDefaultsAndPriorPreserved(t *testing.T) {
	// No group type set: defaults to address-group.
	args := FirewallGroupArgs{Name: "g"}
	u := args.toUnifi("")
	if u.GroupType != "address-group" {
		t.Fatalf("default GroupType = %q, want address-group", u.GroupType)
	}

	// Controller returns empty members/type; prior inputs should be preserved.
	prior := FirewallGroupArgs{Name: "g", GroupType: ptr(FirewallGroupTypePortGroup), GroupMembers: []string{"80", "443"}}
	st := firewallGroupStateFrom(&unifi.FirewallGroup{ID: "x", Name: "g"}, prior)
	if len(st.GroupMembers) != 2 {
		t.Fatalf("preserved GroupMembers = %v, want 2 entries", st.GroupMembers)
	}
	if st.GroupType == nil || *st.GroupType != "port-group" {
		t.Fatalf("preserved GroupType = %v, want port-group", st.GroupType)
	}
}
