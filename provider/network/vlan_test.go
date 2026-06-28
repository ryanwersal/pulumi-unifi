// SPDX-License-Identifier: Apache-2.0

package network

import "testing"

// TestVlanRoundTrip builds a representative spread of inputs (across the nested
// facet groups), maps them to the go-unifi Network, maps the result back, and
// asserts the important fields survive the round-trip.
func TestVlanRoundTrip(t *testing.T) {
	args := VlanArgs{
		Name:                    "lab",
		Purpose:                 ptr(VlanPurposeCorporate),
		Vlan:                    ptr(30),
		Subnet:                  ptr("192.168.30.1/24"),
		Enabled:                 ptr(true),
		NetworkGroup:            ptr("LAN"),
		DomainName:              ptr("lab.example.com"),
		MdnsEnabled:             ptr(true),
		InternetAccessEnabled:   ptr(false),
		NetworkIsolationEnabled: ptr(true),
		AutoScaleEnabled:        ptr(false),
		DpiEnabled:              ptr(true),
		GatewayType:             ptr(VlanGatewayType("default")),
		SettingPreference:       ptr(VlanPreference("manual")),
		InterfaceMtu:            ptr(1500),

		Dhcp: &VlanDhcp{
			Enabled:          ptr(true),
			Start:            ptr("192.168.30.6"),
			Stop:             ptr("192.168.30.254"),
			Lease:            ptr(7200),
			Dns1:             ptr("1.1.1.1"),
			Dns2:             ptr("8.8.8.8"),
			DnsEnabled:       ptr(true),
			Gateway:          ptr("192.168.30.1"),
			GatewayEnabled:   ptr(true),
			Ntp1:             ptr("192.168.30.1"),
			NtpEnabled:       ptr(true),
			Wins1:            ptr("192.168.30.2"),
			WinsEnabled:      ptr(true),
			BootEnabled:      ptr(true),
			BootFilename:     ptr("pxelinux.0"),
			BootServer:       ptr("192.168.30.3"),
			TftpServer:       ptr("192.168.30.3"),
			UnifiController:  ptr("192.168.30.4"),
			ConflictChecking: ptr(true),
			TimeOffset:       ptr(3600),
			RelayEnabled:     ptr(false),
			GuardEnabled:     ptr(true),
		},

		DhcpV6: &VlanDhcpV6{
			Enabled:    ptr(true),
			Dns1:       ptr("2001:4860:4860::8888"),
			DnsAuto:    ptr(false),
			Lease:      ptr(86400),
			Start:      ptr("::2"),
			Stop:       ptr("::7d1"),
			AllowSlaac: ptr(true),
		},

		Igmp: &VlanIgmp{
			Snooping:                  ptr(true),
			ProxyUpstream:             ptr(true),
			ProxyFor:                  ptr(VlanIgmpProxyFor("some")),
			ProxyDownstreamNetworkIds: []string{"abc123"},
			GroupMembership:           ptr(260),
			QuerierSwitches: []NetworkIgmpQuerierSwitch{
				{QuerierAddress: "192.168.30.1", SwitchMac: "aa:bb:cc:dd:ee:ff"},
			},
		},

		Ipv6: &VlanIpv6{
			InterfaceType:           ptr(VlanIpv6InterfaceType("pd")),
			ClientAddressAssignment: ptr(VlanIpv6ClientAddressAssignment("slaac")),
			RaEnabled:               ptr(true),
			RaPriority:              ptr(VlanIpv6RaPriority("high")),
			RaPreferredLifetime:     ptr(14400),
			PdInterface:             ptr(VlanIpv6PdInterface("wan")),
			PdPrefixId:              ptr("1"),
		},

		Wan: &VlanWan{
			Type:                 ptr(VlanWanType("pppoe")),
			TypeV6:               ptr(VlanWanTypeV6("dhcpv6")),
			Ip:                   ptr("203.0.113.10"),
			Netmask:              ptr("255.255.255.0"),
			Gateway:              ptr("203.0.113.1"),
			Dns1:                 ptr("9.9.9.9"),
			DnsPreference:        ptr(VlanPreference("manual")),
			NetworkGroup:         ptr(VlanWanNetworkGroup("WAN")),
			Vlan:                 ptr(35),
			Username:             ptr("isp-user"),
			Password:             ptr("s3cr3t"),
			PppoeUsernameEnabled: ptr(true),
			SmartqEnabled:        ptr(true),
			SmartqUpRate:         ptr(40000),
			SmartqDownRate:       ptr(900000),
			LoadBalanceType:      ptr(VlanWanLoadBalanceType("weighted")),
			LoadBalanceWeight:    ptr(50),
			DhcpCos:              ptr(3),
			EgressQos:            ptr(5),
			DhcpOptions: []NetworkWanDhcpOption{
				{OptionNumber: 43, Value: "0104ABCDEF"},
			},
			ProviderCapabilities: &NetworkWanProviderCapabilities{
				DownloadKilobitsPerSecond: ptr(1000000),
				UploadKilobitsPerSecond:   ptr(40000),
			},
			IpAliases: []string{"203.0.113.20/24"},
		},

		Nat: &VlanNat{
			Masquerade: ptr(true),
			OutboundIpAddresses: []NetworkNatOutboundIp{
				{IpAddress: ptr("203.0.113.30"), Mode: ptr(VlanNatOutboundMode("ip_address")), WanNetworkGroup: ptr(VlanNatWanGroup("WAN"))},
			},
		},
	}

	n := args.toUnifi("net-1")

	// Spot-check the upstream mapping for fields with non-obvious names / behavior.
	if n.ID != "net-1" {
		t.Fatalf("ID: got %q want net-1", n.ID)
	}
	if !n.VLANEnabled || n.VLAN != 30 {
		t.Fatalf("VLAN tagging not enabled: enabled=%v vlan=%d", n.VLANEnabled, n.VLAN)
	}
	if n.NetworkGroup != "LAN" {
		t.Fatalf("NetworkGroup: got %q want LAN", n.NetworkGroup)
	}
	if n.InternetAccessEnabled {
		t.Fatalf("InternetAccessEnabled: got true want false")
	}
	if !n.InterfaceMtuEnabled {
		t.Fatalf("InterfaceMtuEnabled: setting interfaceMtu should enable the override")
	}
	if n.DHCPDLeaseTime != 7200 {
		t.Fatalf("DHCPDLeaseTime: got %d want 7200", n.DHCPDLeaseTime)
	}
	if !n.DHCPguardEnabled {
		t.Fatalf("DHCPguardEnabled: got %v want true", n.DHCPguardEnabled)
	}
	if n.DHCPDV6DNSAuto {
		t.Fatalf("DHCPDV6DNSAuto: got true want false (explicitly set)")
	}
	if !n.IsNAT {
		t.Fatalf("IsNAT: got false want true (nat.masquerade)")
	}
	if n.XWANPassword != "s3cr3t" {
		t.Fatalf("XWANPassword: got %q want s3cr3t", n.XWANPassword)
	}
	if !n.WANVLANEnabled || n.WANVLAN != 35 {
		t.Fatalf("WAN VLAN: enabled=%v vlan=%d", n.WANVLANEnabled, n.WANVLAN)
	}
	if len(n.WANDHCPOptions) != 1 || n.WANDHCPOptions[0].OptionNumber != 43 {
		t.Fatalf("WANDHCPOptions not mapped: %+v", n.WANDHCPOptions)
	}
	if n.WANProviderCapabilities.DownloadKilobitsPerSecond != 1000000 {
		t.Fatalf("WANProviderCapabilities download: got %d", n.WANProviderCapabilities.DownloadKilobitsPerSecond)
	}
	if len(n.NATOutboundIPAddresses) != 1 || n.NATOutboundIPAddresses[0].Mode != "ip_address" {
		t.Fatalf("NATOutboundIPAddresses not mapped: %+v", n.NATOutboundIPAddresses)
	}
	if len(n.IGMPQuerierSwitches) != 1 || n.IGMPQuerierSwitches[0].SwitchMAC != "aa:bb:cc:dd:ee:ff" {
		t.Fatalf("IGMPQuerierSwitches not mapped: %+v", n.IGMPQuerierSwitches)
	}

	// Round-trip back into state, using the original inputs as prior.
	st := vlanStateFrom(n, args)
	if st.NetworkId != "net-1" {
		t.Fatalf("NetworkId: got %q want net-1", st.NetworkId)
	}

	out := st.VlanArgs
	vlanEqStrP(t, "purpose", (*string)(out.Purpose), (*string)(args.Purpose))
	vlanEqIntP(t, "vlan", out.Vlan, args.Vlan)
	vlanEqStrP(t, "subnet", out.Subnet, args.Subnet)
	vlanEqBoolP(t, "enabled", out.Enabled, args.Enabled)
	vlanEqStrP(t, "networkGroup", out.NetworkGroup, args.NetworkGroup)
	vlanEqStrP(t, "domainName", out.DomainName, args.DomainName)
	vlanEqBoolP(t, "mdnsEnabled", out.MdnsEnabled, args.MdnsEnabled)
	vlanEqBoolP(t, "internetAccessEnabled", out.InternetAccessEnabled, args.InternetAccessEnabled)
	vlanEqBoolP(t, "networkIsolationEnabled", out.NetworkIsolationEnabled, args.NetworkIsolationEnabled)
	vlanEqBoolP(t, "dpiEnabled", out.DpiEnabled, args.DpiEnabled)
	vlanEqStrP(t, "gatewayType", (*string)(out.GatewayType), (*string)(args.GatewayType))
	vlanEqStrP(t, "settingPreference", (*string)(out.SettingPreference), (*string)(args.SettingPreference))
	vlanEqIntP(t, "interfaceMtu", out.InterfaceMtu, args.InterfaceMtu)

	if out.Dhcp == nil {
		t.Fatal("dhcp group lost on round-trip")
	}
	vlanEqBoolP(t, "dhcp.enabled", out.Dhcp.Enabled, args.Dhcp.Enabled)
	vlanEqStrP(t, "dhcp.start", out.Dhcp.Start, args.Dhcp.Start)
	vlanEqStrP(t, "dhcp.stop", out.Dhcp.Stop, args.Dhcp.Stop)
	vlanEqIntP(t, "dhcp.lease", out.Dhcp.Lease, args.Dhcp.Lease)
	vlanEqStrP(t, "dhcp.dns1", out.Dhcp.Dns1, args.Dhcp.Dns1)
	vlanEqStrP(t, "dhcp.bootFilename", out.Dhcp.BootFilename, args.Dhcp.BootFilename)
	vlanEqBoolP(t, "dhcp.guardEnabled", out.Dhcp.GuardEnabled, args.Dhcp.GuardEnabled)

	if out.DhcpV6 == nil {
		t.Fatal("dhcpV6 group lost on round-trip")
	}
	vlanEqBoolP(t, "dhcpV6.enabled", out.DhcpV6.Enabled, args.DhcpV6.Enabled)
	vlanEqStrP(t, "dhcpV6.dns1", out.DhcpV6.Dns1, args.DhcpV6.Dns1)
	vlanEqBoolP(t, "dhcpV6.dnsAuto", out.DhcpV6.DnsAuto, args.DhcpV6.DnsAuto)
	vlanEqStrP(t, "dhcpV6.start", out.DhcpV6.Start, args.DhcpV6.Start)

	if out.Igmp == nil {
		t.Fatal("igmp group lost on round-trip")
	}
	vlanEqStrP(t, "igmp.proxyFor", (*string)(out.Igmp.ProxyFor), (*string)(args.Igmp.ProxyFor))
	vlanEqIntP(t, "igmp.groupMembership", out.Igmp.GroupMembership, args.Igmp.GroupMembership)
	if len(out.Igmp.QuerierSwitches) != 1 || out.Igmp.QuerierSwitches[0].SwitchMac != "aa:bb:cc:dd:ee:ff" {
		t.Fatalf("igmp.querierSwitches round-trip: %+v", out.Igmp.QuerierSwitches)
	}

	if out.Ipv6 == nil {
		t.Fatal("ipv6 group lost on round-trip")
	}
	vlanEqStrP(t, "ipv6.interfaceType", (*string)(out.Ipv6.InterfaceType), (*string)(args.Ipv6.InterfaceType))
	vlanEqStrP(t, "ipv6.raPriority", (*string)(out.Ipv6.RaPriority), (*string)(args.Ipv6.RaPriority))
	vlanEqIntP(t, "ipv6.raPreferredLifetime", out.Ipv6.RaPreferredLifetime, args.Ipv6.RaPreferredLifetime)

	if out.Wan == nil {
		t.Fatal("wan group lost on round-trip")
	}
	vlanEqStrP(t, "wan.type", (*string)(out.Wan.Type), (*string)(args.Wan.Type))
	vlanEqStrP(t, "wan.typeV6", (*string)(out.Wan.TypeV6), (*string)(args.Wan.TypeV6))
	vlanEqStrP(t, "wan.gateway", out.Wan.Gateway, args.Wan.Gateway)
	vlanEqStrP(t, "wan.networkGroup", (*string)(out.Wan.NetworkGroup), (*string)(args.Wan.NetworkGroup))
	vlanEqIntP(t, "wan.vlan", out.Wan.Vlan, args.Wan.Vlan)
	vlanEqStrP(t, "wan.username", out.Wan.Username, args.Wan.Username)
	// Secret is preserved from prior because the controller does not echo it.
	vlanEqStrP(t, "wan.password", out.Wan.Password, args.Wan.Password)
	vlanEqStrP(t, "wan.loadBalanceType", (*string)(out.Wan.LoadBalanceType), (*string)(args.Wan.LoadBalanceType))
	vlanEqIntP(t, "wan.smartqUpRate", out.Wan.SmartqUpRate, args.Wan.SmartqUpRate)
	vlanEqIntP(t, "wan.egressQos", out.Wan.EgressQos, args.Wan.EgressQos)
	if len(out.Wan.DhcpOptions) != 1 || out.Wan.DhcpOptions[0].OptionNumber != 43 || out.Wan.DhcpOptions[0].Value != "0104ABCDEF" {
		t.Fatalf("wan.dhcpOptions round-trip: %+v", out.Wan.DhcpOptions)
	}
	if out.Wan.ProviderCapabilities == nil ||
		out.Wan.ProviderCapabilities.DownloadKilobitsPerSecond == nil ||
		*out.Wan.ProviderCapabilities.DownloadKilobitsPerSecond != 1000000 {
		t.Fatalf("wan.providerCapabilities round-trip: %+v", out.Wan.ProviderCapabilities)
	}
	if len(out.Wan.IpAliases) != 1 || out.Wan.IpAliases[0] != "203.0.113.20/24" {
		t.Fatalf("wan.ipAliases round-trip: %+v", out.Wan.IpAliases)
	}

	if out.Nat == nil {
		t.Fatal("nat group lost on round-trip")
	}
	vlanEqBoolP(t, "nat.masquerade", out.Nat.Masquerade, args.Nat.Masquerade)
	if len(out.Nat.OutboundIpAddresses) != 1 || out.Nat.OutboundIpAddresses[0].Mode == nil ||
		*out.Nat.OutboundIpAddresses[0].Mode != "ip_address" {
		t.Fatalf("nat.outboundIpAddresses round-trip: %+v", out.Nat.OutboundIpAddresses)
	}
}

// TestVlanDefaults asserts the documented defaults are applied when inputs are unset.
func TestVlanDefaults(t *testing.T) {
	n := VlanArgs{Name: "minimal"}.toUnifi("")
	if n.Purpose != "corporate" {
		t.Fatalf("default purpose: got %q want corporate", n.Purpose)
	}
	if !n.Enabled {
		t.Fatalf("default enabled: got false want true")
	}
	if n.NetworkGroup != "LAN" {
		t.Fatalf("default networkGroup: got %q want LAN", n.NetworkGroup)
	}
	if !n.InternetAccessEnabled {
		t.Fatalf("default internetAccessEnabled: got false want true")
	}
	// dhcpdv6_dns_auto has no omitempty upstream and defaults true on the
	// controller; an unset value (even with the whole dhcpV6 group omitted) must
	// still be sent as true, not the Go zero.
	if !n.DHCPDV6DNSAuto {
		t.Fatalf("default dhcpV6.dnsAuto: got false want true")
	}
	if n.VLANEnabled {
		t.Fatalf("VLANEnabled should be false when no vlan is set")
	}
}

// TestVlanEmptyGroupsStayNil asserts that round-tripping a minimal network does
// not synthesize non-nil facet groups out of controller zero values.
func TestVlanEmptyGroupsStayNil(t *testing.T) {
	n := VlanArgs{Name: "minimal"}.toUnifi("")
	n.ID = "net-min"
	st := vlanStateFrom(n, VlanArgs{Name: "minimal"})
	out := st.VlanArgs
	// The controller object built from a bare network has no WAN/IPv6/NAT config,
	// so those groups must round-trip as nil (no spurious "always shows changes").
	if out.Wan != nil {
		t.Errorf("wan should be nil for a minimal network, got %+v", out.Wan)
	}
	if out.Ipv6 != nil {
		t.Errorf("ipv6 should be nil for a minimal network, got %+v", out.Ipv6)
	}
	if out.Nat != nil {
		t.Errorf("nat should be nil for a minimal network, got %+v", out.Nat)
	}
	// dhcpV6.dnsAuto must not force the group on when the user never set it.
	if out.DhcpV6 != nil {
		t.Errorf("dhcpV6 should be nil when unset (dnsAuto must not force it), got %+v", out.DhcpV6)
	}
}

func vlanEqStrP(t *testing.T, name string, got, want *string) {
	t.Helper()
	if (got == nil) != (want == nil) || (got != nil && *got != *want) {
		t.Fatalf("%s: got %v want %v", name, vlanDerefStr(got), vlanDerefStr(want))
	}
}

func vlanEqIntP(t *testing.T, name string, got, want *int) {
	t.Helper()
	if (got == nil) != (want == nil) || (got != nil && *got != *want) {
		t.Fatalf("%s: got %v want %v", name, vlanDerefInt(got), vlanDerefInt(want))
	}
}

func vlanEqBoolP(t *testing.T, name string, got, want *bool) {
	t.Helper()
	if (got == nil) != (want == nil) || (got != nil && *got != *want) {
		t.Fatalf("%s: got %v want %v", name, vlanDerefBool(got), vlanDerefBool(want))
	}
}

func vlanDerefStr(p *string) string {
	if p == nil {
		return "<nil>"
	}
	return *p
}

func vlanDerefInt(p *int) any {
	if p == nil {
		return "<nil>"
	}
	return *p
}

func vlanDerefBool(p *bool) any {
	if p == nil {
		return "<nil>"
	}
	return *p
}
