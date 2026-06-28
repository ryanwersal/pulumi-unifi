// SPDX-License-Identifier: Apache-2.0

package network

import "testing"

func TestStaticRouteRoundTrip(t *testing.T) {
	args := StaticRouteArgs{
		Name:            "to-lab",
		Network:         "10.20.0.0/24",
		StaticRouteType: "nexthop-route",
		Enabled:         ptr(true),
		Nexthop:         ptr("192.168.1.1"),
		Distance:        ptr(5),
		Interface:       ptr("WAN1"),
		GatewayDevice:   ptr("aa:bb:cc:dd:ee:ff"),
		GatewayType:     ptr("default"),
	}

	u := args.toUnifi("route-123")

	if u.ID != "route-123" {
		t.Errorf("ID = %q, want route-123", u.ID)
	}
	if u.Type != staticRouteType {
		t.Errorf("Type = %q, want %q", u.Type, staticRouteType)
	}
	if u.Name != "to-lab" {
		t.Errorf("Name = %q, want to-lab", u.Name)
	}
	if u.StaticRouteNetwork != "10.20.0.0/24" {
		t.Errorf("StaticRouteNetwork = %q", u.StaticRouteNetwork)
	}
	if u.StaticRouteType != "nexthop-route" {
		t.Errorf("StaticRouteType = %q", u.StaticRouteType)
	}
	if !u.Enabled {
		t.Errorf("Enabled = false, want true")
	}
	if u.StaticRouteNexthop != "192.168.1.1" {
		t.Errorf("StaticRouteNexthop = %q", u.StaticRouteNexthop)
	}
	if u.StaticRouteDistance != 5 {
		t.Errorf("StaticRouteDistance = %d, want 5", u.StaticRouteDistance)
	}
	if u.StaticRouteInterface != "WAN1" {
		t.Errorf("StaticRouteInterface = %q", u.StaticRouteInterface)
	}
	if u.GatewayDevice != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("GatewayDevice = %q", u.GatewayDevice)
	}
	if u.GatewayType != "default" {
		t.Errorf("GatewayType = %q", u.GatewayType)
	}

	st := staticRouteStateFrom(u, args)

	if st.StaticRouteId != "route-123" {
		t.Errorf("StaticRouteId = %q, want route-123", st.StaticRouteId)
	}
	if st.Name != args.Name {
		t.Errorf("Name = %q, want %q", st.Name, args.Name)
	}
	if st.Network != args.Network {
		t.Errorf("Network = %q, want %q", st.Network, args.Network)
	}
	if st.StaticRouteType != args.StaticRouteType {
		t.Errorf("StaticRouteType = %q, want %q", st.StaticRouteType, args.StaticRouteType)
	}
	if st.Enabled == nil || !*st.Enabled {
		t.Errorf("Enabled did not survive round-trip")
	}
	if st.Nexthop == nil || *st.Nexthop != "192.168.1.1" {
		t.Errorf("Nexthop did not survive round-trip")
	}
	if st.Distance == nil || *st.Distance != 5 {
		t.Errorf("Distance did not survive round-trip")
	}
	if st.Interface == nil || *st.Interface != "WAN1" {
		t.Errorf("Interface did not survive round-trip")
	}
	if st.GatewayDevice == nil || *st.GatewayDevice != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("GatewayDevice did not survive round-trip")
	}
	if st.GatewayType == nil || *st.GatewayType != "default" {
		t.Errorf("GatewayType did not survive round-trip")
	}
}

func TestStaticRouteBlackholeDefaults(t *testing.T) {
	args := StaticRouteArgs{
		Name:            "drop-bogons",
		Network:         "192.0.2.0/24",
		StaticRouteType: "blackhole",
	}

	u := args.toUnifi("")

	if u.ID != "" {
		t.Errorf("ID = %q, want empty on create", u.ID)
	}
	if u.Type != staticRouteType {
		t.Errorf("Type = %q, want %q", u.Type, staticRouteType)
	}
	if !u.Enabled {
		t.Errorf("Enabled default = false, want true")
	}
	if u.StaticRouteNexthop != "" {
		t.Errorf("StaticRouteNexthop = %q, want empty", u.StaticRouteNexthop)
	}

	st := staticRouteStateFrom(u, args)
	if st.Nexthop != nil {
		t.Errorf("Nexthop = %v, want nil (preserved unset)", *st.Nexthop)
	}
	if st.Distance != nil {
		t.Errorf("Distance = %v, want nil (preserved unset)", *st.Distance)
	}
}
