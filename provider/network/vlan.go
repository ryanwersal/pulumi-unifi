// SPDX-License-Identifier: Apache-2.0

package network

import (
	"context"
	"fmt"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/ryanwersal/pulumi-unifi/provider/config"
)

// Vlan is the controlling (marker) struct for a UniFi network/VLAN resource.
type Vlan struct{}

// NetworkIgmpQuerierSwitch is an IGMP querier address pinned to a switch.
type NetworkIgmpQuerierSwitch struct {
	// QuerierAddress is the IGMP querier IPv4 address.
	QuerierAddress string `pulumi:"querierAddress"`
	// SwitchMac is the MAC address of the switch acting as querier.
	SwitchMac string `pulumi:"switchMac"`
}

// NetworkWanDhcpOption is a custom DHCP option requested on the WAN interface.
type NetworkWanDhcpOption struct {
	// OptionNumber is the DHCP option code (1-254).
	OptionNumber int `pulumi:"optionNumber"`
	// Value is the option value.
	Value string `pulumi:"value"`
}

// NetworkWanProviderCapabilities advertises the ISP plan rates (used by SmartQueue).
type NetworkWanProviderCapabilities struct {
	// DownloadKilobitsPerSecond is the ISP advertised download rate in kbps.
	DownloadKilobitsPerSecond *int `pulumi:"downloadKilobitsPerSecond,optional"`
	// UploadKilobitsPerSecond is the ISP advertised upload rate in kbps.
	UploadKilobitsPerSecond *int `pulumi:"uploadKilobitsPerSecond,optional"`
}

// NetworkNatOutboundIp is a single outbound NAT IP / pool mapping.
type NetworkNatOutboundIp struct {
	// IpAddress is the single outbound NAT source IPv4 address.
	IpAddress *string `pulumi:"ipAddress,optional"`
	// IpAddressPool is a list of outbound NAT source IPs/ranges.
	IpAddressPool []string `pulumi:"ipAddressPool,optional"`
	// Mode selects the outbound NAT strategy: all | ip_address | ip_address_pool.
	Mode *string `pulumi:"mode,optional"`
	// WanNetworkGroup binds the mapping to a WAN interface group: WAN | WAN2.
	WanNetworkGroup *string `pulumi:"wanNetworkGroup,optional"`
}

// VlanDhcp groups the IPv4 DHCP-server settings for the network.
type VlanDhcp struct {
	// Enabled toggles the built-in DHCP server for this network.
	Enabled *bool `pulumi:"enabled,optional"`
	// Start is the first address of the DHCP range, e.g. 192.168.20.6.
	Start *string `pulumi:"start,optional"`
	// Stop is the last address of the DHCP range, e.g. 192.168.20.254.
	Stop *string `pulumi:"stop,optional"`
	// Lease is the DHCP lease time in seconds (default 86400).
	Lease *int `pulumi:"lease,optional"`
	// Dns1 is the first DHCP-advertised DNS server.
	Dns1 *string `pulumi:"dns1,optional"`
	// Dns2 is the second DHCP-advertised DNS server.
	Dns2 *string `pulumi:"dns2,optional"`
	// Dns3 is the third DHCP-advertised DNS server.
	Dns3 *string `pulumi:"dns3,optional"`
	// Dns4 is the fourth DHCP-advertised DNS server.
	Dns4 *string `pulumi:"dns4,optional"`
	// DnsEnabled advertises custom DNS servers via DHCP (otherwise the gateway is used).
	DnsEnabled *bool `pulumi:"dnsEnabled,optional"`
	// Gateway overrides the DHCP-advertised gateway address.
	Gateway *string `pulumi:"gateway,optional"`
	// GatewayEnabled toggles the custom DHCP gateway override.
	GatewayEnabled *bool `pulumi:"gatewayEnabled,optional"`
	// Ntp1 is the first DHCP-advertised NTP server.
	Ntp1 *string `pulumi:"ntp1,optional"`
	// Ntp2 is the second DHCP-advertised NTP server.
	Ntp2 *string `pulumi:"ntp2,optional"`
	// NtpEnabled advertises NTP servers via DHCP.
	NtpEnabled *bool `pulumi:"ntpEnabled,optional"`
	// Wins1 is the first DHCP-advertised WINS server.
	Wins1 *string `pulumi:"wins1,optional"`
	// Wins2 is the second DHCP-advertised WINS server.
	Wins2 *string `pulumi:"wins2,optional"`
	// WinsEnabled advertises WINS servers via DHCP.
	WinsEnabled *bool `pulumi:"winsEnabled,optional"`
	// BootEnabled enables DHCP network-boot (PXE) options.
	BootEnabled *bool `pulumi:"bootEnabled,optional"`
	// BootFilename is the boot file name handed to PXE clients.
	BootFilename *string `pulumi:"bootFilename,optional"`
	// BootServer is the next-server (boot server) IPv4 address.
	BootServer *string `pulumi:"bootServer,optional"`
	// TftpServer is the TFTP server advertised via DHCP option 66.
	TftpServer *string `pulumi:"tftpServer,optional"`
	// UnifiController advertises the UniFi controller (inform) address via DHCP.
	UnifiController *string `pulumi:"unifiController,optional"`
	// ConflictChecking probes for IP conflicts before leasing.
	ConflictChecking *bool `pulumi:"conflictChecking,optional"`
	// TimeOffset is the DHCP time offset (option 2) in seconds.
	TimeOffset *int `pulumi:"timeOffset,optional"`
	// TimeOffsetEnabled toggles advertising the time offset.
	TimeOffsetEnabled *bool `pulumi:"timeOffsetEnabled,optional"`
	// WpadUrl advertises a WPAD/proxy-autoconfig URL via DHCP.
	WpadUrl *string `pulumi:"wpadUrl,optional"`
	// RelayEnabled forwards DHCP requests to an external relay instead of serving locally.
	RelayEnabled *bool `pulumi:"relayEnabled,optional"`
	// GuardEnabled blocks rogue DHCP servers on this network.
	GuardEnabled *bool `pulumi:"guardEnabled,optional"`
}

// VlanDhcpV6 groups the stateful DHCPv6-server settings.
type VlanDhcpV6 struct {
	// Enabled enables the stateful DHCPv6 server.
	Enabled *bool `pulumi:"enabled,optional"`
	// Dns1 is the first DHCPv6-advertised DNS server.
	Dns1 *string `pulumi:"dns1,optional"`
	// Dns2 is the second DHCPv6-advertised DNS server.
	Dns2 *string `pulumi:"dns2,optional"`
	// Dns3 is the third DHCPv6-advertised DNS server.
	Dns3 *string `pulumi:"dns3,optional"`
	// Dns4 is the fourth DHCPv6-advertised DNS server.
	Dns4 *string `pulumi:"dns4,optional"`
	// DnsAuto uses upstream-provided DNS for DHCPv6 (default true) instead of manual servers.
	DnsAuto *bool `pulumi:"dnsAuto,optional"`
	// Lease is the DHCPv6 lease time in seconds (default 86400).
	Lease *int `pulumi:"lease,optional"`
	// Start is the first address of the DHCPv6 range.
	Start *string `pulumi:"start,optional"`
	// Stop is the last address of the DHCPv6 range.
	Stop *string `pulumi:"stop,optional"`
	// AllowSlaac allows SLAAC alongside DHCPv6.
	AllowSlaac *bool `pulumi:"allowSlaac,optional"`
}

// VlanIpv6 groups the IPv6 addressing / Router Advertisement / prefix-delegation settings.
type VlanIpv6 struct {
	// InterfaceType: none | static | pd | single_network.
	InterfaceType *string `pulumi:"interfaceType,optional"`
	// ClientAddressAssignment: slaac | dhcpv6.
	ClientAddressAssignment *string `pulumi:"clientAddressAssignment,optional"`
	// Subnet is the static IPv6 subnet (CIDR) when interfaceType=static.
	Subnet *string `pulumi:"subnet,optional"`
	// SettingPreference: auto | manual.
	SettingPreference *string `pulumi:"settingPreference,optional"`
	// RaEnabled enables IPv6 Router Advertisements.
	RaEnabled *bool `pulumi:"raEnabled,optional"`
	// RaPriority: high | medium | low.
	RaPriority *string `pulumi:"raPriority,optional"`
	// RaPreferredLifetime is the RA preferred lifetime in seconds (default 14400).
	RaPreferredLifetime *int `pulumi:"raPreferredLifetime,optional"`
	// RaValidLifetime is the RA valid lifetime in seconds (default 86400).
	RaValidLifetime *int `pulumi:"raValidLifetime,optional"`
	// PdInterface is the WAN used for prefix delegation: wan | wan2.
	PdInterface *string `pulumi:"pdInterface,optional"`
	// PdPrefixId is the hex prefix ID carved from the delegated prefix.
	PdPrefixId *string `pulumi:"pdPrefixId,optional"`
	// PdStart is the first address of the PD-derived range.
	PdStart *string `pulumi:"pdStart,optional"`
	// PdStop is the last address of the PD-derived range.
	PdStop *string `pulumi:"pdStop,optional"`
	// PdAutoPrefixIdEnabled lets the controller auto-assign the PD prefix ID.
	PdAutoPrefixIdEnabled *bool `pulumi:"pdAutoPrefixIdEnabled,optional"`
	// SingleNetworkInterface is the source network for single_network IPv6 mode.
	SingleNetworkInterface *string `pulumi:"singleNetworkInterface,optional"`
	// WanDelegationType (WAN networks): pd | single_network | none.
	WanDelegationType *string `pulumi:"wanDelegationType,optional"`
}

// VlanIgmp groups the multicast / IGMP settings.
type VlanIgmp struct {
	// Snooping enables IGMP snooping to optimize multicast flooding.
	Snooping *bool `pulumi:"snooping,optional"`
	// ProxyUpstream marks this network as the IGMP proxy upstream.
	ProxyUpstream *bool `pulumi:"proxyUpstream,optional"`
	// ProxyFor selects downstream proxy scope: all | some | none.
	ProxyFor *string `pulumi:"proxyFor,optional"`
	// ProxyDownstreamNetworkIds lists downstream networks when proxyFor=some.
	ProxyDownstreamNetworkIds []string `pulumi:"proxyDownstreamNetworkIds,optional"`
	// FastLeave enables IGMP fast-leave processing.
	FastLeave *bool `pulumi:"fastLeave,optional"`
	// ForwardUnknownMulticast forwards unregistered multicast groups.
	ForwardUnknownMulticast *bool `pulumi:"forwardUnknownMulticast,optional"`
	// GroupMembership is the IGMP group membership interval (seconds).
	GroupMembership *int `pulumi:"groupMembership,optional"`
	// MaxResponse is the IGMP max response time (seconds).
	MaxResponse *int `pulumi:"maxResponse,optional"`
	// McrtrExpireTime is the multicast router expiry time (seconds).
	McrtrExpireTime *int `pulumi:"mcrtrExpireTime,optional"`
	// Suppression suppresses redundant IGMP membership reports.
	Suppression *bool `pulumi:"suppression,optional"`
	// QuerierSwitches pins IGMP querier addresses to specific switches.
	QuerierSwitches []NetworkIgmpQuerierSwitch `pulumi:"querierSwitches,optional"`
}

// VlanWan groups the WAN-uplink settings (used when purpose=wan).
type VlanWan struct {
	// Type (purpose=wan): disabled | dhcp | static | pppoe | dslite.
	Type *string `pulumi:"type,optional"`
	// TypeV6 (purpose=wan): disabled | slaac | dhcpv6 | static.
	TypeV6 *string `pulumi:"typeV6,optional"`
	// Ip is the static WAN IPv4 address.
	Ip *string `pulumi:"ip,optional"`
	// Ipv6 is the static WAN IPv6 address.
	Ipv6 *string `pulumi:"ipv6,optional"`
	// Netmask is the static WAN IPv4 netmask.
	Netmask *string `pulumi:"netmask,optional"`
	// Gateway is the static WAN IPv4 gateway.
	Gateway *string `pulumi:"gateway,optional"`
	// GatewayV6 is the static WAN IPv6 gateway.
	GatewayV6 *string `pulumi:"gatewayV6,optional"`
	// Dns1 is the first WAN DNS server.
	Dns1 *string `pulumi:"dns1,optional"`
	// Dns2 is the second WAN DNS server.
	Dns2 *string `pulumi:"dns2,optional"`
	// Dns3 is the third WAN DNS server.
	Dns3 *string `pulumi:"dns3,optional"`
	// Dns4 is the fourth WAN DNS server.
	Dns4 *string `pulumi:"dns4,optional"`
	// DnsPreference: auto | manual.
	DnsPreference *string `pulumi:"dnsPreference,optional"`
	// Ipv6Dns1 is the first WAN IPv6 DNS server.
	Ipv6Dns1 *string `pulumi:"ipv6Dns1,optional"`
	// Ipv6Dns2 is the second WAN IPv6 DNS server.
	Ipv6Dns2 *string `pulumi:"ipv6Dns2,optional"`
	// Ipv6DnsPreference: auto | manual.
	Ipv6DnsPreference *string `pulumi:"ipv6DnsPreference,optional"`
	// NetworkGroup is the WAN interface group: WAN | WAN2 | WAN_LTE_FAILOVER.
	NetworkGroup *string `pulumi:"networkGroup,optional"`
	// Vlan tags the WAN interface with a VLAN ID. Setting it enables WAN VLAN tagging.
	Vlan *int `pulumi:"vlan,optional"`
	// VlanEnabled toggles WAN VLAN tagging explicitly.
	VlanEnabled *bool `pulumi:"vlanEnabled,optional"`
	// Username is the PPPoE username (type=pppoe).
	Username *string `pulumi:"username,optional"`
	// Password is the PPPoE password (type=pppoe). Secret.
	Password *string `pulumi:"password,optional" provider:"secret"`
	// PppoeUsernameEnabled toggles sending the PPPoE username.
	PppoeUsernameEnabled *bool `pulumi:"pppoeUsernameEnabled,optional"`
	// PppoePasswordEnabled toggles sending the PPPoE password.
	PppoePasswordEnabled *bool `pulumi:"pppoePasswordEnabled,optional"`
	// SmartqEnabled enables SmartQueue QoS on the WAN.
	SmartqEnabled *bool `pulumi:"smartqEnabled,optional"`
	// SmartqUpRate is the SmartQueue upload limit in kbps.
	SmartqUpRate *int `pulumi:"smartqUpRate,optional"`
	// SmartqDownRate is the SmartQueue download limit in kbps.
	SmartqDownRate *int `pulumi:"smartqDownRate,optional"`
	// LoadBalanceType: failover-only | weighted.
	LoadBalanceType *string `pulumi:"loadBalanceType,optional"`
	// LoadBalanceWeight is the weighted load-balance weight (1-99).
	LoadBalanceWeight *int `pulumi:"loadBalanceWeight,optional"`
	// DhcpCos is the 802.1p CoS applied to WAN DHCP traffic (0-7).
	DhcpCos *int `pulumi:"dhcpCos,optional"`
	// DhcpOptions are custom DHCP options requested on the WAN.
	DhcpOptions []NetworkWanDhcpOption `pulumi:"dhcpOptions,optional"`
	// Dhcpv6PdSize is the IPv6 PD size to request from the ISP (48-64).
	Dhcpv6PdSize *int `pulumi:"dhcpv6PdSize,optional"`
	// Prefixlen is the static WAN IPv6 prefix length (1-128).
	Prefixlen *int `pulumi:"prefixlen,optional"`
	// EgressQos is the 802.1p priority for WAN egress (1-7).
	EgressQos *int `pulumi:"egressQos,optional"`
	// ProviderCapabilities advertises the ISP plan rates for SmartQueue.
	ProviderCapabilities *NetworkWanProviderCapabilities `pulumi:"providerCapabilities,optional"`
	// IpAliases are additional WAN IP aliases (CIDR).
	IpAliases []string `pulumi:"ipAliases,optional"`
	// DsliteRemoteHost is the DS-Lite AFTR remote host (type=dslite).
	DsliteRemoteHost *string `pulumi:"dsliteRemoteHost,optional"`
}

// VlanNat groups the source/outbound NAT settings.
type VlanNat struct {
	// Masquerade enables source NAT (masquerade) for this network.
	Masquerade *bool `pulumi:"masquerade,optional"`
	// OutboundIpAddresses configures outbound (source) NAT IP mappings.
	OutboundIpAddresses []NetworkNatOutboundIp `pulumi:"outboundIpAddresses,optional"`
}

// VlanArgs are the user-supplied inputs for a network/VLAN.
type VlanArgs struct {
	// Name of the network.
	Name string `pulumi:"name"`
	// Purpose: corporate | guest | vlan-only | wan | remote-user-vpn | site-vpn | vpn-client. Defaults to "corporate".
	Purpose *string `pulumi:"purpose,optional"`
	// Vlan is the 802.1Q VLAN ID. When set, VLAN tagging is enabled.
	Vlan *int `pulumi:"vlan,optional"`
	// Subnet is the gateway IP/CIDR for the network, e.g. 192.168.20.1/24.
	Subnet *string `pulumi:"subnet,optional"`
	// Enabled controls whether the network is active. Defaults to true.
	Enabled *bool `pulumi:"enabled,optional"`

	// NetworkGroup is the interface group: LAN, LAN2..LAN8 (or WAN/WAN2 for WAN networks). Defaults to "LAN".
	NetworkGroup *string `pulumi:"networkGroup,optional"`
	// DomainName is the DNS domain handed to clients.
	DomainName *string `pulumi:"domainName,optional"`
	// MdnsEnabled enables multicast DNS (Bonjour/mDNS) repeating on this network.
	MdnsEnabled *bool `pulumi:"mdnsEnabled,optional"`
	// InternetAccessEnabled controls whether clients may reach the internet. Defaults to true.
	InternetAccessEnabled *bool `pulumi:"internetAccessEnabled,optional"`
	// NetworkIsolationEnabled isolates clients from other networks. Defaults to false.
	NetworkIsolationEnabled *bool `pulumi:"networkIsolationEnabled,optional"`
	// AutoScaleEnabled lets the controller auto-size the subnet.
	AutoScaleEnabled *bool `pulumi:"autoScaleEnabled,optional"`
	// DpiEnabled enables Deep Packet Inspection on this network.
	DpiEnabled *bool `pulumi:"dpiEnabled,optional"`
	// GatewayType: default | switch.
	GatewayType *string `pulumi:"gatewayType,optional"`
	// SettingPreference: auto | manual. Controls whether device-specific overrides apply.
	SettingPreference *string `pulumi:"settingPreference,optional"`
	// InterfaceMtu sets the interface MTU. Setting it enables the MTU override.
	InterfaceMtu *int `pulumi:"interfaceMtu,optional"`
	// InterfaceMtuEnabled toggles the interface MTU override explicitly.
	InterfaceMtuEnabled *bool `pulumi:"interfaceMtuEnabled,optional"`
	// MacOverride clones a MAC address on the interface (commonly used on WAN).
	MacOverride *string `pulumi:"macOverride,optional"`
	// MacOverrideEnabled toggles the MAC clone override.
	MacOverrideEnabled *bool `pulumi:"macOverrideEnabled,optional"`
	// UpnpLanEnabled enables UPnP on this LAN.
	UpnpLanEnabled *bool `pulumi:"upnpLanEnabled,optional"`

	// Dhcp groups the IPv4 DHCP-server settings.
	Dhcp *VlanDhcp `pulumi:"dhcp,optional"`
	// DhcpV6 groups the stateful DHCPv6-server settings.
	DhcpV6 *VlanDhcpV6 `pulumi:"dhcpV6,optional"`
	// Ipv6 groups the IPv6 addressing / RA / prefix-delegation settings.
	Ipv6 *VlanIpv6 `pulumi:"ipv6,optional"`
	// Igmp groups the multicast / IGMP settings.
	Igmp *VlanIgmp `pulumi:"igmp,optional"`
	// Wan groups the WAN-uplink settings (used when purpose=wan).
	Wan *VlanWan `pulumi:"wan,optional"`
	// Nat groups the source/outbound NAT settings.
	Nat *VlanNat `pulumi:"nat,optional"`
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
	a.Describe(&v, "A UniFi Network (VLAN/LAN/WAN). Maps to a controller network configuration object. "+
		"Covers LAN/VLAN DHCP (v4 and v6), IGMP, IPv6, and WAN settings, grouped into nested objects "+
		"(dhcp, dhcpV6, ipv6, igmp, wan, nat). VPN-specific networks "+
		"(IPSec/OpenVPN/WireGuard/L2TP/PPTP) are not modeled.")
}

// Annotate documents the outbound NAT mapping fields.
func (m *NetworkNatOutboundIp) Annotate(a infer.Annotator) {
	a.Describe(&m.IpAddress, "IpAddress is the single outbound NAT source IPv4 address.")
	a.Describe(&m.IpAddressPool, "IpAddressPool is a list of outbound NAT source IPs/ranges.")
	a.Describe(&m.Mode, "Outbound NAT strategy: all | ip_address | ip_address_pool.")
	a.Describe(&m.WanNetworkGroup, "WAN interface group the mapping applies to: WAN | WAN2.")
}

func (s *NetworkIgmpQuerierSwitch) Annotate(a infer.Annotator) {
	a.Describe(&s.QuerierAddress, "QuerierAddress is the IGMP querier IPv4 address.")
	a.Describe(&s.SwitchMac, "SwitchMac is the MAC address of the switch acting as querier.")
}

func (o *NetworkWanDhcpOption) Annotate(a infer.Annotator) {
	a.Describe(&o.OptionNumber, "OptionNumber is the DHCP option code (1-254).")
	a.Describe(&o.Value, "Value is the option value.")
}

func (c *NetworkWanProviderCapabilities) Annotate(a infer.Annotator) {
	a.Describe(&c.DownloadKilobitsPerSecond, "DownloadKilobitsPerSecond is the ISP advertised download rate in kbps.")
	a.Describe(&c.UploadKilobitsPerSecond, "UploadKilobitsPerSecond is the ISP advertised upload rate in kbps.")
}

func (d *VlanDhcp) Annotate(a infer.Annotator) {
	a.Describe(&d.Enabled, "Enabled toggles the built-in DHCP server for this network.")
	a.Describe(&d.Start, "Start is the first address of the DHCP range, e.g. 192.168.20.6.")
	a.Describe(&d.Stop, "Stop is the last address of the DHCP range, e.g. 192.168.20.254.")
	a.Describe(&d.Lease, "Lease is the DHCP lease time in seconds (default 86400).")
	a.Describe(&d.Dns1, "Dns1 is the first DHCP-advertised DNS server.")
	a.Describe(&d.Dns2, "Dns2 is the second DHCP-advertised DNS server.")
	a.Describe(&d.Dns3, "Dns3 is the third DHCP-advertised DNS server.")
	a.Describe(&d.Dns4, "Dns4 is the fourth DHCP-advertised DNS server.")
	a.Describe(&d.DnsEnabled, "DnsEnabled advertises custom DNS servers via DHCP (otherwise the gateway is used).")
	a.Describe(&d.Gateway, "Gateway overrides the DHCP-advertised gateway address.")
	a.Describe(&d.GatewayEnabled, "GatewayEnabled toggles the custom DHCP gateway override.")
	a.Describe(&d.Ntp1, "Ntp1 is the first DHCP-advertised NTP server.")
	a.Describe(&d.Ntp2, "Ntp2 is the second DHCP-advertised NTP server.")
	a.Describe(&d.NtpEnabled, "NtpEnabled advertises NTP servers via DHCP.")
	a.Describe(&d.Wins1, "Wins1 is the first DHCP-advertised WINS server.")
	a.Describe(&d.Wins2, "Wins2 is the second DHCP-advertised WINS server.")
	a.Describe(&d.WinsEnabled, "WinsEnabled advertises WINS servers via DHCP.")
	a.Describe(&d.BootEnabled, "BootEnabled enables DHCP network-boot (PXE) options.")
	a.Describe(&d.BootFilename, "BootFilename is the boot file name handed to PXE clients.")
	a.Describe(&d.BootServer, "BootServer is the next-server (boot server) IPv4 address.")
	a.Describe(&d.TftpServer, "TftpServer is the TFTP server advertised via DHCP option 66.")
	a.Describe(&d.UnifiController, "UnifiController advertises the UniFi controller (inform) address via DHCP.")
	a.Describe(&d.ConflictChecking, "ConflictChecking probes for IP conflicts before leasing.")
	a.Describe(&d.TimeOffset, "TimeOffset is the DHCP time offset (option 2) in seconds.")
	a.Describe(&d.TimeOffsetEnabled, "TimeOffsetEnabled toggles advertising the time offset.")
	a.Describe(&d.WpadUrl, "WpadUrl advertises a WPAD/proxy-autoconfig URL via DHCP.")
	a.Describe(&d.RelayEnabled, "RelayEnabled forwards DHCP requests to an external relay instead of serving locally.")
	a.Describe(&d.GuardEnabled, "GuardEnabled blocks rogue DHCP servers on this network.")
}

func (d *VlanDhcpV6) Annotate(a infer.Annotator) {
	a.Describe(&d.Enabled, "Enabled enables the stateful DHCPv6 server.")
	a.Describe(&d.Dns1, "Dns1 is the first DHCPv6-advertised DNS server.")
	a.Describe(&d.Dns2, "Dns2 is the second DHCPv6-advertised DNS server.")
	a.Describe(&d.Dns3, "Dns3 is the third DHCPv6-advertised DNS server.")
	a.Describe(&d.Dns4, "Dns4 is the fourth DHCPv6-advertised DNS server.")
	a.Describe(&d.DnsAuto, "DnsAuto uses upstream-provided DNS for DHCPv6 (default true) instead of manual servers.")
	a.Describe(&d.Lease, "Lease is the DHCPv6 lease time in seconds (default 86400).")
	a.Describe(&d.Start, "Start is the first address of the DHCPv6 range.")
	a.Describe(&d.Stop, "Stop is the last address of the DHCPv6 range.")
	a.Describe(&d.AllowSlaac, "AllowSlaac allows SLAAC alongside DHCPv6.")
}

func (v6 *VlanIpv6) Annotate(a infer.Annotator) {
	a.Describe(&v6.InterfaceType, "InterfaceType: none | static | pd | single_network.")
	a.Describe(&v6.ClientAddressAssignment, "ClientAddressAssignment: slaac | dhcpv6.")
	a.Describe(&v6.Subnet, "Subnet is the static IPv6 subnet (CIDR) when interfaceType=static.")
	a.Describe(&v6.SettingPreference, "SettingPreference: auto | manual.")
	a.Describe(&v6.RaEnabled, "RaEnabled enables IPv6 Router Advertisements.")
	a.Describe(&v6.RaPriority, "RaPriority: high | medium | low.")
	a.Describe(&v6.RaPreferredLifetime, "RaPreferredLifetime is the RA preferred lifetime in seconds (default 14400).")
	a.Describe(&v6.RaValidLifetime, "RaValidLifetime is the RA valid lifetime in seconds (default 86400).")
	a.Describe(&v6.PdInterface, "PdInterface is the WAN used for prefix delegation: wan | wan2.")
	a.Describe(&v6.PdPrefixId, "PdPrefixId is the hex prefix ID carved from the delegated prefix.")
	a.Describe(&v6.PdStart, "PdStart is the first address of the PD-derived range.")
	a.Describe(&v6.PdStop, "PdStop is the last address of the PD-derived range.")
	a.Describe(&v6.PdAutoPrefixIdEnabled, "PdAutoPrefixIdEnabled lets the controller auto-assign the PD prefix ID.")
	a.Describe(&v6.SingleNetworkInterface, "SingleNetworkInterface is the source network for single_network IPv6 mode.")
	a.Describe(&v6.WanDelegationType, "WanDelegationType (WAN networks): pd | single_network | none.")
}

func (g *VlanIgmp) Annotate(a infer.Annotator) {
	a.Describe(&g.Snooping, "Snooping enables IGMP snooping to optimize multicast flooding.")
	a.Describe(&g.ProxyUpstream, "ProxyUpstream marks this network as the IGMP proxy upstream.")
	a.Describe(&g.ProxyFor, "ProxyFor selects downstream proxy scope: all | some | none.")
	a.Describe(&g.ProxyDownstreamNetworkIds, "ProxyDownstreamNetworkIds lists downstream networks when proxyFor=some.")
	a.Describe(&g.FastLeave, "FastLeave enables IGMP fast-leave processing.")
	a.Describe(&g.ForwardUnknownMulticast, "ForwardUnknownMulticast forwards unregistered multicast groups.")
	a.Describe(&g.GroupMembership, "GroupMembership is the IGMP group membership interval (seconds).")
	a.Describe(&g.MaxResponse, "MaxResponse is the IGMP max response time (seconds).")
	a.Describe(&g.McrtrExpireTime, "McrtrExpireTime is the multicast router expiry time (seconds).")
	a.Describe(&g.Suppression, "Suppression suppresses redundant IGMP membership reports.")
	a.Describe(&g.QuerierSwitches, "QuerierSwitches pins IGMP querier addresses to specific switches.")
}

func (w *VlanWan) Annotate(a infer.Annotator) {
	a.Describe(&w.Type, "Type (purpose=wan): disabled | dhcp | static | pppoe | dslite.")
	a.Describe(&w.TypeV6, "TypeV6 (purpose=wan): disabled | slaac | dhcpv6 | static.")
	a.Describe(&w.Ip, "Ip is the static WAN IPv4 address.")
	a.Describe(&w.Ipv6, "Ipv6 is the static WAN IPv6 address.")
	a.Describe(&w.Netmask, "Netmask is the static WAN IPv4 netmask.")
	a.Describe(&w.Gateway, "Gateway is the static WAN IPv4 gateway.")
	a.Describe(&w.GatewayV6, "GatewayV6 is the static WAN IPv6 gateway.")
	a.Describe(&w.Dns1, "Dns1 is the first WAN DNS server.")
	a.Describe(&w.Dns2, "Dns2 is the second WAN DNS server.")
	a.Describe(&w.Dns3, "Dns3 is the third WAN DNS server.")
	a.Describe(&w.Dns4, "Dns4 is the fourth WAN DNS server.")
	a.Describe(&w.DnsPreference, "DnsPreference: auto | manual.")
	a.Describe(&w.Ipv6Dns1, "Ipv6Dns1 is the first WAN IPv6 DNS server.")
	a.Describe(&w.Ipv6Dns2, "Ipv6Dns2 is the second WAN IPv6 DNS server.")
	a.Describe(&w.Ipv6DnsPreference, "Ipv6DnsPreference: auto | manual.")
	a.Describe(&w.NetworkGroup, "NetworkGroup is the WAN interface group: WAN | WAN2 | WAN_LTE_FAILOVER.")
	a.Describe(&w.Vlan, "Vlan tags the WAN interface with a VLAN ID. Setting it enables WAN VLAN tagging.")
	a.Describe(&w.VlanEnabled, "VlanEnabled toggles WAN VLAN tagging explicitly.")
	a.Describe(&w.Username, "Username is the PPPoE username (type=pppoe).")
	a.Describe(&w.Password, "Password is the PPPoE password (type=pppoe). Secret.")
	a.Describe(&w.PppoeUsernameEnabled, "PppoeUsernameEnabled toggles sending the PPPoE username.")
	a.Describe(&w.PppoePasswordEnabled, "PppoePasswordEnabled toggles sending the PPPoE password.")
	a.Describe(&w.SmartqEnabled, "SmartqEnabled enables SmartQueue QoS on the WAN.")
	a.Describe(&w.SmartqUpRate, "SmartqUpRate is the SmartQueue upload limit in kbps.")
	a.Describe(&w.SmartqDownRate, "SmartqDownRate is the SmartQueue download limit in kbps.")
	a.Describe(&w.LoadBalanceType, "LoadBalanceType: failover-only | weighted.")
	a.Describe(&w.LoadBalanceWeight, "LoadBalanceWeight is the weighted load-balance weight (1-99).")
	a.Describe(&w.DhcpCos, "DhcpCos is the 802.1p CoS applied to WAN DHCP traffic (0-7).")
	a.Describe(&w.DhcpOptions, "DhcpOptions are custom DHCP options requested on the WAN.")
	a.Describe(&w.Dhcpv6PdSize, "Dhcpv6PdSize is the IPv6 PD size to request from the ISP (48-64).")
	a.Describe(&w.Prefixlen, "Prefixlen is the static WAN IPv6 prefix length (1-128).")
	a.Describe(&w.EgressQos, "EgressQos is the 802.1p priority for WAN egress (1-7).")
	a.Describe(&w.ProviderCapabilities, "ProviderCapabilities advertises the ISP plan rates for SmartQueue.")
	a.Describe(&w.IpAliases, "IpAliases are additional WAN IP aliases (CIDR).")
	a.Describe(&w.DsliteRemoteHost, "DsliteRemoteHost is the DS-Lite AFTR remote host (type=dslite).")
}

func (nat *VlanNat) Annotate(a infer.Annotator) {
	a.Describe(&nat.Masquerade, "Masquerade enables source NAT (masquerade) for this network.")
	a.Describe(&nat.OutboundIpAddresses, "OutboundIpAddresses configures outbound (source) NAT IP mappings.")
}

func (d *VlanArgs) Annotate(a infer.Annotator) {
	a.Describe(&d.Name, "Name of the network.")
	a.Describe(&d.Purpose, "Purpose: corporate | guest | vlan-only | wan | remote-user-vpn | site-vpn | vpn-client. Defaults to \"corporate\".")
	a.Describe(&d.Vlan, "Vlan is the 802.1Q VLAN ID. When set, VLAN tagging is enabled.")
	a.Describe(&d.Subnet, "Subnet is the gateway IP/CIDR for the network, e.g. 192.168.20.1/24.")
	a.Describe(&d.Enabled, "Enabled controls whether the network is active. Defaults to true.")
	a.Describe(&d.NetworkGroup, "NetworkGroup is the interface group: LAN, LAN2..LAN8 (or WAN/WAN2 for WAN networks). Defaults to \"LAN\".")
	a.Describe(&d.DomainName, "DomainName is the DNS domain handed to clients.")
	a.Describe(&d.MdnsEnabled, "MdnsEnabled enables multicast DNS (Bonjour/mDNS) repeating on this network.")
	a.Describe(&d.InternetAccessEnabled, "InternetAccessEnabled controls whether clients may reach the internet. Defaults to true.")
	a.Describe(&d.NetworkIsolationEnabled, "NetworkIsolationEnabled isolates clients from other networks. Defaults to false.")
	a.Describe(&d.AutoScaleEnabled, "AutoScaleEnabled lets the controller auto-size the subnet.")
	a.Describe(&d.DpiEnabled, "DpiEnabled enables Deep Packet Inspection on this network.")
	a.Describe(&d.GatewayType, "GatewayType: default | switch.")
	a.Describe(&d.SettingPreference, "SettingPreference: auto | manual. Controls whether device-specific overrides apply.")
	a.Describe(&d.InterfaceMtu, "InterfaceMtu sets the interface MTU. Setting it enables the MTU override.")
	a.Describe(&d.InterfaceMtuEnabled, "InterfaceMtuEnabled toggles the interface MTU override explicitly.")
	a.Describe(&d.MacOverride, "MacOverride clones a MAC address on the interface (commonly used on WAN).")
	a.Describe(&d.MacOverrideEnabled, "MacOverrideEnabled toggles the MAC clone override.")
	a.Describe(&d.UpnpLanEnabled, "UpnpLanEnabled enables UPnP on this LAN.")
	a.Describe(&d.Dhcp, "Dhcp groups the IPv4 DHCP-server settings.")
	a.Describe(&d.DhcpV6, "DhcpV6 groups the stateful DHCPv6-server settings.")
	a.Describe(&d.Ipv6, "Ipv6 groups the IPv6 addressing / RA / prefix-delegation settings.")
	a.Describe(&d.Igmp, "Igmp groups the multicast / IGMP settings.")
	a.Describe(&d.Wan, "Wan groups the WAN-uplink settings (used when purpose=wan).")
	a.Describe(&d.Nat, "Nat groups the source/outbound NAT settings.")
}

func (s *VlanState) Annotate(a infer.Annotator) {
	a.Describe(&s.NetworkId, "NetworkId is the controller-assigned identifier (the UniFi `_id`).")
}

// toUnifi builds a go-unifi Network from inputs. id is empty on create.
func (a VlanArgs) toUnifi(id string) *unifi.Network {
	n := &unifi.Network{
		ID:                    id,
		Name:                  a.Name,
		Purpose:               derefOr(a.Purpose, "corporate"),
		Enabled:               derefOr(a.Enabled, true),
		NetworkGroup:          derefOr(a.NetworkGroup, "LAN"),
		InternetAccessEnabled: derefOr(a.InternetAccessEnabled, true),
	}
	if a.Vlan != nil {
		n.VLAN = *a.Vlan
		n.VLANEnabled = true
	}
	if a.Subnet != nil {
		n.IPSubnet = *a.Subnet
	}

	// General / LAN.
	if a.DomainName != nil {
		n.DomainName = *a.DomainName
	}
	if a.MdnsEnabled != nil {
		n.MdnsEnabled = *a.MdnsEnabled
	}
	if a.NetworkIsolationEnabled != nil {
		n.NetworkIsolationEnabled = *a.NetworkIsolationEnabled
	}
	if a.AutoScaleEnabled != nil {
		n.AutoScaleEnabled = *a.AutoScaleEnabled
	}
	if a.DpiEnabled != nil {
		n.DPIEnabled = *a.DpiEnabled
	}
	if a.GatewayType != nil {
		n.GatewayType = *a.GatewayType
	}
	if a.SettingPreference != nil {
		n.SettingPreference = *a.SettingPreference
	}
	if a.InterfaceMtu != nil {
		n.InterfaceMtu = *a.InterfaceMtu
		n.InterfaceMtuEnabled = true
	}
	if a.InterfaceMtuEnabled != nil {
		n.InterfaceMtuEnabled = *a.InterfaceMtuEnabled
	}
	if a.MacOverride != nil {
		n.MACOverride = *a.MacOverride
	}
	if a.MacOverrideEnabled != nil {
		n.MACOverrideEnabled = *a.MacOverrideEnabled
	}
	if a.UpnpLanEnabled != nil {
		n.UpnpLanEnabled = *a.UpnpLanEnabled
	}

	// DHCP server.
	if d := a.Dhcp; d != nil {
		if d.Enabled != nil {
			n.DHCPDEnabled = *d.Enabled
		}
		if d.Start != nil {
			n.DHCPDStart = *d.Start
		}
		if d.Stop != nil {
			n.DHCPDStop = *d.Stop
		}
		if d.Lease != nil {
			n.DHCPDLeaseTime = *d.Lease
		}
		if d.Dns1 != nil {
			n.DHCPDDNS1 = *d.Dns1
		}
		if d.Dns2 != nil {
			n.DHCPDDNS2 = *d.Dns2
		}
		if d.Dns3 != nil {
			n.DHCPDDNS3 = *d.Dns3
		}
		if d.Dns4 != nil {
			n.DHCPDDNS4 = *d.Dns4
		}
		if d.DnsEnabled != nil {
			n.DHCPDDNSEnabled = *d.DnsEnabled
		}
		if d.Gateway != nil {
			n.DHCPDGateway = *d.Gateway
		}
		if d.GatewayEnabled != nil {
			n.DHCPDGatewayEnabled = *d.GatewayEnabled
		}
		if d.Ntp1 != nil {
			n.DHCPDNtp1 = *d.Ntp1
		}
		if d.Ntp2 != nil {
			n.DHCPDNtp2 = *d.Ntp2
		}
		if d.NtpEnabled != nil {
			n.DHCPDNtpEnabled = *d.NtpEnabled
		}
		if d.Wins1 != nil {
			n.DHCPDWins1 = *d.Wins1
		}
		if d.Wins2 != nil {
			n.DHCPDWins2 = *d.Wins2
		}
		if d.WinsEnabled != nil {
			n.DHCPDWinsEnabled = *d.WinsEnabled
		}
		if d.BootEnabled != nil {
			n.DHCPDBootEnabled = *d.BootEnabled
		}
		if d.BootFilename != nil {
			n.DHCPDBootFilename = *d.BootFilename
		}
		if d.BootServer != nil {
			n.DHCPDBootServer = *d.BootServer
		}
		if d.TftpServer != nil {
			n.DHCPDTFTPServer = *d.TftpServer
		}
		if d.UnifiController != nil {
			n.DHCPDUnifiController = *d.UnifiController
		}
		if d.ConflictChecking != nil {
			n.DHCPDConflictChecking = *d.ConflictChecking
		}
		if d.TimeOffset != nil {
			n.DHCPDTimeOffset = *d.TimeOffset
		}
		if d.TimeOffsetEnabled != nil {
			n.DHCPDTimeOffsetEnabled = *d.TimeOffsetEnabled
		}
		if d.WpadUrl != nil {
			n.DHCPDWPAdUrl = *d.WpadUrl
		}
		if d.RelayEnabled != nil {
			n.DHCPRelayEnabled = *d.RelayEnabled
		}
		if d.GuardEnabled != nil {
			n.DHCPguardEnabled = *d.GuardEnabled
		}
	}

	// DHCPv6.
	if d := a.DhcpV6; d != nil {
		if d.Enabled != nil {
			n.DHCPDV6Enabled = *d.Enabled
		}
		if d.Dns1 != nil {
			n.DHCPDV6DNS1 = *d.Dns1
		}
		if d.Dns2 != nil {
			n.DHCPDV6DNS2 = *d.Dns2
		}
		if d.Dns3 != nil {
			n.DHCPDV6DNS3 = *d.Dns3
		}
		if d.Dns4 != nil {
			n.DHCPDV6DNS4 = *d.Dns4
		}
		if d.Lease != nil {
			n.DHCPDV6LeaseTime = *d.Lease
		}
		if d.Start != nil {
			n.DHCPDV6Start = *d.Start
		}
		if d.Stop != nil {
			n.DHCPDV6Stop = *d.Stop
		}
		if d.AllowSlaac != nil {
			n.DHCPDV6AllowSlaac = *d.AllowSlaac
		}
	}
	// dhcpdv6_dns_auto has no omitempty upstream and the controller default is
	// true, so always send it even when the dhcpV6 group is omitted; omitting it
	// (zero value false) would silently switch DHCPv6 clients off upstream/auto
	// DNS. Mirrors internetAccessEnabled.
	n.DHCPDV6DNSAuto = true
	if a.DhcpV6 != nil && a.DhcpV6.DnsAuto != nil {
		n.DHCPDV6DNSAuto = *a.DhcpV6.DnsAuto
	}

	// IGMP.
	if g := a.Igmp; g != nil {
		if g.Snooping != nil {
			n.IGMPSnooping = *g.Snooping
		}
		if g.ProxyUpstream != nil {
			n.IGMPProxyUpstream = *g.ProxyUpstream
		}
		if g.ProxyFor != nil {
			n.IGMPProxyFor = *g.ProxyFor
		}
		if g.ProxyDownstreamNetworkIds != nil {
			n.IGMPProxyDownstreamNetworkIDs = g.ProxyDownstreamNetworkIds
		}
		if g.FastLeave != nil {
			n.IGMPFastleave = *g.FastLeave
		}
		if g.ForwardUnknownMulticast != nil {
			n.IGMPForwardUnknownMulticast = *g.ForwardUnknownMulticast
		}
		if g.GroupMembership != nil {
			n.IGMPGroupmembership = *g.GroupMembership
		}
		if g.MaxResponse != nil {
			n.IGMPMaxresponse = *g.MaxResponse
		}
		if g.McrtrExpireTime != nil {
			n.IGMPMcrtrexpiretime = *g.McrtrExpireTime
		}
		if g.Suppression != nil {
			n.IGMPSupression = *g.Suppression
		}
		for _, q := range g.QuerierSwitches {
			n.IGMPQuerierSwitches = append(n.IGMPQuerierSwitches, unifi.NetworkIGMPQuerierSwitches{
				QuerierAddress: q.QuerierAddress,
				SwitchMAC:      q.SwitchMac,
			})
		}
	}

	// IPv6.
	if v6 := a.Ipv6; v6 != nil {
		if v6.InterfaceType != nil {
			n.IPV6InterfaceType = *v6.InterfaceType
		}
		if v6.ClientAddressAssignment != nil {
			n.IPV6ClientAddressAssignment = *v6.ClientAddressAssignment
		}
		if v6.Subnet != nil {
			n.IPV6Subnet = *v6.Subnet
		}
		if v6.SettingPreference != nil {
			n.IPV6SettingPreference = *v6.SettingPreference
		}
		if v6.RaEnabled != nil {
			n.IPV6RaEnabled = *v6.RaEnabled
		}
		if v6.RaPriority != nil {
			n.IPV6RaPriority = *v6.RaPriority
		}
		if v6.RaPreferredLifetime != nil {
			n.IPV6RaPreferredLifetime = *v6.RaPreferredLifetime
		}
		if v6.RaValidLifetime != nil {
			n.IPV6RaValidLifetime = *v6.RaValidLifetime
		}
		if v6.PdInterface != nil {
			n.IPV6PDInterface = *v6.PdInterface
		}
		if v6.PdPrefixId != nil {
			n.IPV6PDPrefixid = *v6.PdPrefixId
		}
		if v6.PdStart != nil {
			n.IPV6PDStart = *v6.PdStart
		}
		if v6.PdStop != nil {
			n.IPV6PDStop = *v6.PdStop
		}
		if v6.PdAutoPrefixIdEnabled != nil {
			n.IPV6PDAutoPrefixidEnabled = *v6.PdAutoPrefixIdEnabled
		}
		if v6.SingleNetworkInterface != nil {
			n.IPV6SingleNetworkInterface = *v6.SingleNetworkInterface
		}
		if v6.WanDelegationType != nil {
			n.IPV6WANDelegationType = *v6.WanDelegationType
		}
	}

	// WAN.
	if w := a.Wan; w != nil {
		if w.Type != nil {
			n.WANType = *w.Type
		}
		if w.TypeV6 != nil {
			n.WANTypeV6 = *w.TypeV6
		}
		if w.Ip != nil {
			n.WANIP = *w.Ip
		}
		if w.Ipv6 != nil {
			n.WANIPV6 = *w.Ipv6
		}
		if w.Netmask != nil {
			n.WANNetmask = *w.Netmask
		}
		if w.Gateway != nil {
			n.WANGateway = *w.Gateway
		}
		if w.GatewayV6 != nil {
			n.WANGatewayV6 = *w.GatewayV6
		}
		if w.Dns1 != nil {
			n.WANDNS1 = *w.Dns1
		}
		if w.Dns2 != nil {
			n.WANDNS2 = *w.Dns2
		}
		if w.Dns3 != nil {
			n.WANDNS3 = *w.Dns3
		}
		if w.Dns4 != nil {
			n.WANDNS4 = *w.Dns4
		}
		if w.DnsPreference != nil {
			n.WANDNSPreference = *w.DnsPreference
		}
		if w.Ipv6Dns1 != nil {
			n.WANIPV6DNS1 = *w.Ipv6Dns1
		}
		if w.Ipv6Dns2 != nil {
			n.WANIPV6DNS2 = *w.Ipv6Dns2
		}
		if w.Ipv6DnsPreference != nil {
			n.WANIPV6DNSPreference = *w.Ipv6DnsPreference
		}
		if w.NetworkGroup != nil {
			n.WANNetworkGroup = *w.NetworkGroup
		}
		if w.Vlan != nil {
			n.WANVLAN = *w.Vlan
			n.WANVLANEnabled = true
		}
		if w.VlanEnabled != nil {
			n.WANVLANEnabled = *w.VlanEnabled
		}
		if w.Username != nil {
			n.WANUsername = *w.Username
		}
		if w.Password != nil {
			n.XWANPassword = *w.Password
		}
		if w.PppoeUsernameEnabled != nil {
			n.WANPppoeUsernameEnabled = *w.PppoeUsernameEnabled
		}
		if w.PppoePasswordEnabled != nil {
			n.WANPppoePasswordEnabled = *w.PppoePasswordEnabled
		}
		if w.SmartqEnabled != nil {
			n.WANSmartqEnabled = *w.SmartqEnabled
		}
		if w.SmartqUpRate != nil {
			n.WANSmartqUpRate = *w.SmartqUpRate
		}
		if w.SmartqDownRate != nil {
			n.WANSmartqDownRate = *w.SmartqDownRate
		}
		if w.LoadBalanceType != nil {
			n.WANLoadBalanceType = *w.LoadBalanceType
		}
		if w.LoadBalanceWeight != nil {
			n.WANLoadBalanceWeight = *w.LoadBalanceWeight
		}
		if w.DhcpCos != nil {
			n.WANDHCPCos = *w.DhcpCos
		}
		for _, o := range w.DhcpOptions {
			n.WANDHCPOptions = append(n.WANDHCPOptions, unifi.NetworkWANDHCPOptions{
				OptionNumber: o.OptionNumber,
				Value:        o.Value,
			})
		}
		if w.Dhcpv6PdSize != nil {
			n.WANDHCPv6PDSize = *w.Dhcpv6PdSize
		}
		if w.Prefixlen != nil {
			n.WANPrefixlen = *w.Prefixlen
		}
		if w.EgressQos != nil {
			n.WANEgressQOS = *w.EgressQos
		}
		if w.ProviderCapabilities != nil {
			caps := unifi.NetworkWANProviderCapabilities{}
			if w.ProviderCapabilities.DownloadKilobitsPerSecond != nil {
				caps.DownloadKilobitsPerSecond = *w.ProviderCapabilities.DownloadKilobitsPerSecond
			}
			if w.ProviderCapabilities.UploadKilobitsPerSecond != nil {
				caps.UploadKilobitsPerSecond = *w.ProviderCapabilities.UploadKilobitsPerSecond
			}
			n.WANProviderCapabilities = caps
		}
		if w.IpAliases != nil {
			n.WANIPAliases = w.IpAliases
		}
		if w.DsliteRemoteHost != nil {
			n.WANDsliteRemoteHost = *w.DsliteRemoteHost
		}
	}

	// NAT.
	if nat := a.Nat; nat != nil {
		if nat.Masquerade != nil {
			n.IsNAT = *nat.Masquerade
		}
		for _, na := range nat.OutboundIpAddresses {
			item := unifi.NetworkNATOutboundIPAddresses{
				IPAddressPool: na.IpAddressPool,
			}
			if na.IpAddress != nil {
				item.IPAddress = *na.IpAddress
			}
			if na.Mode != nil {
				item.Mode = *na.Mode
			}
			if na.WanNetworkGroup != nil {
				item.WANNetworkGroup = *na.WanNetworkGroup
			}
			n.NATOutboundIPAddresses = append(n.NATOutboundIPAddresses, item)
		}
	}

	return n
}

// vlanStrPtr reflects a controller string, falling back to the prior input when empty.
func vlanStrPtr(v string, prior *string) *string {
	if v != "" {
		return ptr(v)
	}
	return prior
}

// vlanIntPtr reflects a controller int, falling back to the prior input when zero.
func vlanIntPtr(v int, prior *int) *int {
	if v != 0 {
		return ptr(v)
	}
	return prior
}

// vlanBoolPtr reflects a controller bool when the user set it or when it is true,
// otherwise leaves the optional input unset to avoid spurious diffs.
func vlanBoolPtr(v bool, prior *bool) *bool {
	if v {
		return ptr(v)
	}
	return prior
}

// isZero reports whether no IGMP member is set (so the group round-trips as nil).
func (g VlanIgmp) isZero() bool {
	return g.Snooping == nil && g.ProxyUpstream == nil && g.ProxyFor == nil &&
		g.ProxyDownstreamNetworkIds == nil && g.FastLeave == nil && g.ForwardUnknownMulticast == nil &&
		g.GroupMembership == nil && g.MaxResponse == nil && g.McrtrExpireTime == nil &&
		g.Suppression == nil && g.QuerierSwitches == nil
}

// isZero reports whether no WAN member is set (so the group round-trips as nil).
func (w VlanWan) isZero() bool {
	return w.Type == nil && w.TypeV6 == nil && w.Ip == nil && w.Ipv6 == nil && w.Netmask == nil &&
		w.Gateway == nil && w.GatewayV6 == nil && w.Dns1 == nil && w.Dns2 == nil && w.Dns3 == nil &&
		w.Dns4 == nil && w.DnsPreference == nil && w.Ipv6Dns1 == nil && w.Ipv6Dns2 == nil &&
		w.Ipv6DnsPreference == nil && w.NetworkGroup == nil && w.Vlan == nil && w.VlanEnabled == nil &&
		w.Username == nil && w.Password == nil && w.PppoeUsernameEnabled == nil && w.PppoePasswordEnabled == nil &&
		w.SmartqEnabled == nil && w.SmartqUpRate == nil && w.SmartqDownRate == nil && w.LoadBalanceType == nil &&
		w.LoadBalanceWeight == nil && w.DhcpCos == nil && w.DhcpOptions == nil && w.Dhcpv6PdSize == nil &&
		w.Prefixlen == nil && w.EgressQos == nil && w.ProviderCapabilities == nil && w.IpAliases == nil &&
		w.DsliteRemoteHost == nil
}

// isZero reports whether no NAT member is set (so the group round-trips as nil).
func (nat VlanNat) isZero() bool {
	return nat.Masquerade == nil && nat.OutboundIpAddresses == nil
}

// vlanDhcpFrom reconstructs the dhcp group, returning nil when no member is set.
func vlanDhcpFrom(n *unifi.Network, prior *VlanDhcp) *VlanDhcp {
	var p VlanDhcp
	if prior != nil {
		p = *prior
	}
	d := VlanDhcp{
		Enabled:           ptr(n.DHCPDEnabled),
		Start:             vlanStrPtr(n.DHCPDStart, p.Start),
		Stop:              vlanStrPtr(n.DHCPDStop, p.Stop),
		Lease:             vlanIntPtr(n.DHCPDLeaseTime, p.Lease),
		Dns1:              vlanStrPtr(n.DHCPDDNS1, p.Dns1),
		Dns2:              vlanStrPtr(n.DHCPDDNS2, p.Dns2),
		Dns3:              vlanStrPtr(n.DHCPDDNS3, p.Dns3),
		Dns4:              vlanStrPtr(n.DHCPDDNS4, p.Dns4),
		DnsEnabled:        vlanBoolPtr(n.DHCPDDNSEnabled, p.DnsEnabled),
		Gateway:           vlanStrPtr(n.DHCPDGateway, p.Gateway),
		GatewayEnabled:    vlanBoolPtr(n.DHCPDGatewayEnabled, p.GatewayEnabled),
		Ntp1:              vlanStrPtr(n.DHCPDNtp1, p.Ntp1),
		Ntp2:              vlanStrPtr(n.DHCPDNtp2, p.Ntp2),
		NtpEnabled:        vlanBoolPtr(n.DHCPDNtpEnabled, p.NtpEnabled),
		Wins1:             vlanStrPtr(n.DHCPDWins1, p.Wins1),
		Wins2:             vlanStrPtr(n.DHCPDWins2, p.Wins2),
		WinsEnabled:       vlanBoolPtr(n.DHCPDWinsEnabled, p.WinsEnabled),
		BootEnabled:       vlanBoolPtr(n.DHCPDBootEnabled, p.BootEnabled),
		BootFilename:      vlanStrPtr(n.DHCPDBootFilename, p.BootFilename),
		BootServer:        vlanStrPtr(n.DHCPDBootServer, p.BootServer),
		TftpServer:        vlanStrPtr(n.DHCPDTFTPServer, p.TftpServer),
		UnifiController:   vlanStrPtr(n.DHCPDUnifiController, p.UnifiController),
		ConflictChecking:  vlanBoolPtr(n.DHCPDConflictChecking, p.ConflictChecking),
		TimeOffset:        vlanIntPtr(n.DHCPDTimeOffset, p.TimeOffset),
		TimeOffsetEnabled: vlanBoolPtr(n.DHCPDTimeOffsetEnabled, p.TimeOffsetEnabled),
		WpadUrl:           vlanStrPtr(n.DHCPDWPAdUrl, p.WpadUrl),
		RelayEnabled:      vlanBoolPtr(n.DHCPRelayEnabled, p.RelayEnabled),
		GuardEnabled:      vlanBoolPtr(n.DHCPguardEnabled, p.GuardEnabled),
	}
	if d == (VlanDhcp{}) {
		return nil
	}
	return &d
}

// vlanDhcpV6From reconstructs the dhcpV6 group. DnsAuto is reflected from the
// controller (default true) only when the user previously set it, to avoid
// forcing the group on for networks that never configured DHCPv6.
func vlanDhcpV6From(n *unifi.Network, prior *VlanDhcpV6) *VlanDhcpV6 {
	var p VlanDhcpV6
	if prior != nil {
		p = *prior
	}
	d := VlanDhcpV6{
		Enabled:    vlanBoolPtr(n.DHCPDV6Enabled, p.Enabled),
		Dns1:       vlanStrPtr(n.DHCPDV6DNS1, p.Dns1),
		Dns2:       vlanStrPtr(n.DHCPDV6DNS2, p.Dns2),
		Dns3:       vlanStrPtr(n.DHCPDV6DNS3, p.Dns3),
		Dns4:       vlanStrPtr(n.DHCPDV6DNS4, p.Dns4),
		Lease:      vlanIntPtr(n.DHCPDV6LeaseTime, p.Lease),
		Start:      vlanStrPtr(n.DHCPDV6Start, p.Start),
		Stop:       vlanStrPtr(n.DHCPDV6Stop, p.Stop),
		AllowSlaac: vlanBoolPtr(n.DHCPDV6AllowSlaac, p.AllowSlaac),
	}
	// Only reflect dnsAuto into the input when the user expressed an opinion;
	// otherwise the always-true controller value would force the group non-nil.
	if p.DnsAuto != nil {
		d.DnsAuto = ptr(n.DHCPDV6DNSAuto)
	}
	if d == (VlanDhcpV6{}) {
		return nil
	}
	return &d
}

// vlanIpv6From reconstructs the ipv6 group, returning nil when no member is set.
func vlanIpv6From(n *unifi.Network, prior *VlanIpv6) *VlanIpv6 {
	var p VlanIpv6
	if prior != nil {
		p = *prior
	}
	v6 := VlanIpv6{
		InterfaceType:           vlanStrPtr(n.IPV6InterfaceType, p.InterfaceType),
		ClientAddressAssignment: vlanStrPtr(n.IPV6ClientAddressAssignment, p.ClientAddressAssignment),
		Subnet:                  vlanStrPtr(n.IPV6Subnet, p.Subnet),
		SettingPreference:       vlanStrPtr(n.IPV6SettingPreference, p.SettingPreference),
		RaEnabled:               vlanBoolPtr(n.IPV6RaEnabled, p.RaEnabled),
		RaPriority:              vlanStrPtr(n.IPV6RaPriority, p.RaPriority),
		RaPreferredLifetime:     vlanIntPtr(n.IPV6RaPreferredLifetime, p.RaPreferredLifetime),
		RaValidLifetime:         vlanIntPtr(n.IPV6RaValidLifetime, p.RaValidLifetime),
		PdInterface:             vlanStrPtr(n.IPV6PDInterface, p.PdInterface),
		PdPrefixId:              vlanStrPtr(n.IPV6PDPrefixid, p.PdPrefixId),
		PdStart:                 vlanStrPtr(n.IPV6PDStart, p.PdStart),
		PdStop:                  vlanStrPtr(n.IPV6PDStop, p.PdStop),
		PdAutoPrefixIdEnabled:   vlanBoolPtr(n.IPV6PDAutoPrefixidEnabled, p.PdAutoPrefixIdEnabled),
		SingleNetworkInterface:  vlanStrPtr(n.IPV6SingleNetworkInterface, p.SingleNetworkInterface),
		WanDelegationType:       vlanStrPtr(n.IPV6WANDelegationType, p.WanDelegationType),
	}
	if v6 == (VlanIpv6{}) {
		return nil
	}
	return &v6
}

// vlanIgmpFrom reconstructs the igmp group, returning nil when no member is set.
func vlanIgmpFrom(n *unifi.Network, prior *VlanIgmp) *VlanIgmp {
	var p VlanIgmp
	if prior != nil {
		p = *prior
	}
	g := VlanIgmp{
		Snooping:                vlanBoolPtr(n.IGMPSnooping, p.Snooping),
		ProxyUpstream:           vlanBoolPtr(n.IGMPProxyUpstream, p.ProxyUpstream),
		ProxyFor:                vlanStrPtr(n.IGMPProxyFor, p.ProxyFor),
		FastLeave:               vlanBoolPtr(n.IGMPFastleave, p.FastLeave),
		ForwardUnknownMulticast: vlanBoolPtr(n.IGMPForwardUnknownMulticast, p.ForwardUnknownMulticast),
		GroupMembership:         vlanIntPtr(n.IGMPGroupmembership, p.GroupMembership),
		MaxResponse:             vlanIntPtr(n.IGMPMaxresponse, p.MaxResponse),
		McrtrExpireTime:         vlanIntPtr(n.IGMPMcrtrexpiretime, p.McrtrExpireTime),
		Suppression:             vlanBoolPtr(n.IGMPSupression, p.Suppression),
	}
	if len(n.IGMPProxyDownstreamNetworkIDs) > 0 {
		g.ProxyDownstreamNetworkIds = n.IGMPProxyDownstreamNetworkIDs
	} else {
		g.ProxyDownstreamNetworkIds = p.ProxyDownstreamNetworkIds
	}
	if len(n.IGMPQuerierSwitches) > 0 {
		for _, q := range n.IGMPQuerierSwitches {
			g.QuerierSwitches = append(g.QuerierSwitches, NetworkIgmpQuerierSwitch{
				QuerierAddress: q.QuerierAddress,
				SwitchMac:      q.SwitchMAC,
			})
		}
	} else {
		g.QuerierSwitches = p.QuerierSwitches
	}
	if g.isZero() {
		return nil
	}
	return &g
}

// vlanWanFrom reconstructs the wan group, returning nil when no member is set.
func vlanWanFrom(n *unifi.Network, prior *VlanWan) *VlanWan {
	var p VlanWan
	if prior != nil {
		p = *prior
	}
	w := VlanWan{
		Type:              vlanStrPtr(n.WANType, p.Type),
		TypeV6:            vlanStrPtr(n.WANTypeV6, p.TypeV6),
		Ip:                vlanStrPtr(n.WANIP, p.Ip),
		Ipv6:              vlanStrPtr(n.WANIPV6, p.Ipv6),
		Netmask:           vlanStrPtr(n.WANNetmask, p.Netmask),
		Gateway:           vlanStrPtr(n.WANGateway, p.Gateway),
		GatewayV6:         vlanStrPtr(n.WANGatewayV6, p.GatewayV6),
		Dns1:              vlanStrPtr(n.WANDNS1, p.Dns1),
		Dns2:              vlanStrPtr(n.WANDNS2, p.Dns2),
		Dns3:              vlanStrPtr(n.WANDNS3, p.Dns3),
		Dns4:              vlanStrPtr(n.WANDNS4, p.Dns4),
		DnsPreference:     vlanStrPtr(n.WANDNSPreference, p.DnsPreference),
		Ipv6Dns1:          vlanStrPtr(n.WANIPV6DNS1, p.Ipv6Dns1),
		Ipv6Dns2:          vlanStrPtr(n.WANIPV6DNS2, p.Ipv6Dns2),
		Ipv6DnsPreference: vlanStrPtr(n.WANIPV6DNSPreference, p.Ipv6DnsPreference),
		NetworkGroup:      vlanStrPtr(n.WANNetworkGroup, p.NetworkGroup),
		VlanEnabled:       vlanBoolPtr(n.WANVLANEnabled, p.VlanEnabled),
		Username:          vlanStrPtr(n.WANUsername, p.Username),
		// The controller never echoes the PPPoE password; preserve the user input.
		Password:             p.Password,
		PppoeUsernameEnabled: vlanBoolPtr(n.WANPppoeUsernameEnabled, p.PppoeUsernameEnabled),
		PppoePasswordEnabled: vlanBoolPtr(n.WANPppoePasswordEnabled, p.PppoePasswordEnabled),
		SmartqEnabled:        vlanBoolPtr(n.WANSmartqEnabled, p.SmartqEnabled),
		SmartqUpRate:         vlanIntPtr(n.WANSmartqUpRate, p.SmartqUpRate),
		SmartqDownRate:       vlanIntPtr(n.WANSmartqDownRate, p.SmartqDownRate),
		LoadBalanceType:      vlanStrPtr(n.WANLoadBalanceType, p.LoadBalanceType),
		LoadBalanceWeight:    vlanIntPtr(n.WANLoadBalanceWeight, p.LoadBalanceWeight),
		DhcpCos:              vlanIntPtr(n.WANDHCPCos, p.DhcpCos),
		Dhcpv6PdSize:         vlanIntPtr(n.WANDHCPv6PDSize, p.Dhcpv6PdSize),
		Prefixlen:            vlanIntPtr(n.WANPrefixlen, p.Prefixlen),
		EgressQos:            vlanIntPtr(n.WANEgressQOS, p.EgressQos),
		DsliteRemoteHost:     vlanStrPtr(n.WANDsliteRemoteHost, p.DsliteRemoteHost),
	}
	if n.WANVLANEnabled {
		w.Vlan = ptr(n.WANVLAN)
	} else {
		w.Vlan = p.Vlan
	}
	if len(n.WANDHCPOptions) > 0 {
		for _, o := range n.WANDHCPOptions {
			w.DhcpOptions = append(w.DhcpOptions, NetworkWanDhcpOption{
				OptionNumber: o.OptionNumber,
				Value:        o.Value,
			})
		}
	} else {
		w.DhcpOptions = p.DhcpOptions
	}
	if n.WANProviderCapabilities.DownloadKilobitsPerSecond != 0 || n.WANProviderCapabilities.UploadKilobitsPerSecond != 0 {
		caps := &NetworkWanProviderCapabilities{}
		if n.WANProviderCapabilities.DownloadKilobitsPerSecond != 0 {
			caps.DownloadKilobitsPerSecond = ptr(n.WANProviderCapabilities.DownloadKilobitsPerSecond)
		}
		if n.WANProviderCapabilities.UploadKilobitsPerSecond != 0 {
			caps.UploadKilobitsPerSecond = ptr(n.WANProviderCapabilities.UploadKilobitsPerSecond)
		}
		w.ProviderCapabilities = caps
	} else {
		w.ProviderCapabilities = p.ProviderCapabilities
	}
	if len(n.WANIPAliases) > 0 {
		w.IpAliases = n.WANIPAliases
	} else {
		w.IpAliases = p.IpAliases
	}
	if w.isZero() {
		return nil
	}
	return &w
}

// vlanNatFrom reconstructs the nat group, returning nil when no member is set.
func vlanNatFrom(n *unifi.Network, prior *VlanNat) *VlanNat {
	var p VlanNat
	if prior != nil {
		p = *prior
	}
	nat := VlanNat{Masquerade: vlanBoolPtr(n.IsNAT, p.Masquerade)}
	if len(n.NATOutboundIPAddresses) > 0 {
		for _, na := range n.NATOutboundIPAddresses {
			item := NetworkNatOutboundIp{}
			if na.IPAddress != "" {
				item.IpAddress = ptr(na.IPAddress)
			}
			if len(na.IPAddressPool) > 0 {
				item.IpAddressPool = na.IPAddressPool
			}
			if na.Mode != "" {
				item.Mode = ptr(na.Mode)
			}
			if na.WANNetworkGroup != "" {
				item.WanNetworkGroup = ptr(na.WANNetworkGroup)
			}
			nat.OutboundIpAddresses = append(nat.OutboundIpAddresses, item)
		}
	} else {
		nat.OutboundIpAddresses = p.OutboundIpAddresses
	}
	if nat.isZero() {
		return nil
	}
	return &nat
}

// vlanStateFrom maps a controller Network back into resource state. prior holds
// the user inputs so write-only/secret values (e.g. the PPPoE password) and
// unset optional fields are preserved across the round-trip.
func vlanStateFrom(n *unifi.Network, prior VlanArgs) VlanState {
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

	// General / LAN.
	args.NetworkGroup = vlanStrPtr(n.NetworkGroup, prior.NetworkGroup)
	args.DomainName = vlanStrPtr(n.DomainName, prior.DomainName)
	args.MdnsEnabled = vlanBoolPtr(n.MdnsEnabled, prior.MdnsEnabled)
	args.InternetAccessEnabled = ptr(n.InternetAccessEnabled)
	args.NetworkIsolationEnabled = vlanBoolPtr(n.NetworkIsolationEnabled, prior.NetworkIsolationEnabled)
	args.AutoScaleEnabled = vlanBoolPtr(n.AutoScaleEnabled, prior.AutoScaleEnabled)
	args.DpiEnabled = vlanBoolPtr(n.DPIEnabled, prior.DpiEnabled)
	args.GatewayType = vlanStrPtr(n.GatewayType, prior.GatewayType)
	args.SettingPreference = vlanStrPtr(n.SettingPreference, prior.SettingPreference)
	args.InterfaceMtu = vlanIntPtr(n.InterfaceMtu, prior.InterfaceMtu)
	args.InterfaceMtuEnabled = vlanBoolPtr(n.InterfaceMtuEnabled, prior.InterfaceMtuEnabled)
	args.MacOverride = vlanStrPtr(n.MACOverride, prior.MacOverride)
	args.MacOverrideEnabled = vlanBoolPtr(n.MACOverrideEnabled, prior.MacOverrideEnabled)
	args.UpnpLanEnabled = vlanBoolPtr(n.UpnpLanEnabled, prior.UpnpLanEnabled)

	// Nested facets.
	args.Dhcp = vlanDhcpFrom(n, prior.Dhcp)
	args.DhcpV6 = vlanDhcpV6From(n, prior.DhcpV6)
	args.Ipv6 = vlanIpv6From(n, prior.Ipv6)
	args.Igmp = vlanIgmpFrom(n, prior.Igmp)
	args.Wan = vlanWanFrom(n, prior.Wan)
	args.Nat = vlanNatFrom(n, prior.Nat)

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
		return infer.CreateResponse[VlanState]{}, wrap(fmt.Sprintf("create network %q (site %q)", req.Inputs.Name, cfg.ResolvedSite()), err)
	}
	return infer.CreateResponse[VlanState]{ID: created.ID, Output: vlanStateFrom(created, req.Inputs)}, nil
}

// Read recovers state from the controller, enabling `pulumi import`.
func (Vlan) Read(ctx context.Context, req infer.ReadRequest[VlanArgs, VlanState]) (infer.ReadResponse[VlanArgs, VlanState], error) {
	cfg := infer.GetConfig[config.Config](ctx)
	n, err := cfg.Network().GetNetwork(ctx, cfg.ResolvedSite(), req.ID)
	if notFound(err) {
		return infer.ReadResponse[VlanArgs, VlanState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[VlanArgs, VlanState]{}, wrap(fmt.Sprintf("read network %q (site %q)", req.ID, cfg.ResolvedSite()), err)
	}
	st := vlanStateFrom(n, req.Inputs)
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
		return infer.UpdateResponse[VlanState]{}, wrap(fmt.Sprintf("update network %q (site %q)", req.ID, cfg.ResolvedSite()), err)
	}
	return infer.UpdateResponse[VlanState]{Output: vlanStateFrom(updated, req.Inputs)}, nil
}

// Delete removes the network.
func (Vlan) Delete(ctx context.Context, req infer.DeleteRequest[VlanState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[config.Config](ctx)
	err := cfg.Network().DeleteNetwork(ctx, cfg.ResolvedSite(), req.ID)
	if notFound(err) {
		return infer.DeleteResponse{}, nil
	}
	return infer.DeleteResponse{}, wrap(fmt.Sprintf("delete network %q (site %q)", req.ID, cfg.ResolvedSite()), err)
}
