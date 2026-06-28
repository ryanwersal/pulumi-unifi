// SPDX-License-Identifier: Apache-2.0

package network

import (
	"context"
	"fmt"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/ryanwersal/pulumi-unifi/provider/config"
)

// Device manages the settings of an EXISTING, already-adopted UniFi network
// device (gateway, switch, access point, PDU, ...).
//
// Physical hardware is not created or destroyed through this API; the controller
// only exposes settings patches on devices it has already adopted. This resource
// therefore follows an ADOPTION model (like protect.Camera): it binds to a device
// by MAC, and Create/Update perform a read-modify-write — the current device is
// fetched, ONLY the explicitly-managed input fields are overlaid onto it, and the
// merged object is written back. Unmanaged settings (and port/outlet/radio
// overrides the program does not mention) are preserved. Delete is a no-op: the
// device is left in place, only released from Pulumi's management.
type Device struct{}

// DeviceEtherLightingArgs configures the EtherLighting LED strip on supported
// switches (e.g. the 24-port EtherLighting PoE switch).
type DeviceEtherLightingArgs struct {
	// LedMode: standard | etherlighting.
	LedMode *string `pulumi:"ledMode,optional"`
	// Mode of the lighting animation: speed | network.
	Mode *string `pulumi:"mode,optional"`
	// Behavior of the lighting: breath | steady.
	Behavior *string `pulumi:"behavior,optional"`
	// Brightness of the LED strip (1-100).
	Brightness *int `pulumi:"brightness,optional"`
}

// DeviceOutletOverride configures a single outlet on a PDU.
type DeviceOutletOverride struct {
	// Index is the 1-based outlet number (required key).
	Index int `pulumi:"index"`
	// Name is the outlet label.
	Name *string `pulumi:"name,optional"`
	// RelayState powers the outlet on (true) or off (false).
	RelayState *bool `pulumi:"relayState,optional"`
	// CycleEnabled allows the outlet to be power-cycled.
	CycleEnabled *bool `pulumi:"cycleEnabled,optional"`
}

// DeviceEthernetOverride assigns a network group to a physical ethernet port
// (mainly on gateways, e.g. mapping eth ports to WAN/WAN2/LAN groups).
type DeviceEthernetOverride struct {
	// Ifname is the interface name, e.g. "eth1" (required key).
	Ifname string `pulumi:"ifname"`
	// NetworkGroup is the assigned group, e.g. "LAN", "LAN2", "WAN", "WAN2".
	NetworkGroup *string `pulumi:"networkGroup,optional"`
}

// DeviceRadioOverride tunes a single radio of an access point. Entries are
// matched to the device's existing radios by band (radio).
type DeviceRadioOverride struct {
	// Radio band identifying which radio to tune: ng (2.4GHz) | na (5GHz) | ad (60GHz) | 6e (6GHz). Required key.
	Radio string `pulumi:"radio"`
	// Name is the controller's radio name (e.g. "wifi0").
	Name *string `pulumi:"name,optional"`
	// Channel is the radio channel, a number or "auto".
	Channel *string `pulumi:"channel,optional"`
	// ChannelWidth (HT) in MHz: 20 | 40 | 80 | 160 | 240 | 320 | 1080 | 2160 | 4320.
	ChannelWidth *int `pulumi:"channelWidth,optional"`
	// TxPower is the transmit power in dBm, or "auto". Honored only when txPowerMode is "custom".
	TxPower *string `pulumi:"txPower,optional"`
	// TxPowerMode: auto | medium | high | low | custom.
	TxPowerMode *string `pulumi:"txPowerMode,optional"`
	// MinRssi is the minimum RSSI (dBm, negative) below which clients are kicked. Setting it enables the feature unless minRssiEnabled is given.
	MinRssi *int `pulumi:"minRssi,optional"`
	// MinRssiEnabled toggles the minimum-RSSI kick feature.
	MinRssiEnabled *bool `pulumi:"minRssiEnabled,optional"`
	// AntennaGain in dBi (for devices with configurable antennas).
	AntennaGain *int `pulumi:"antennaGain,optional"`
}

// DevicePortOverride configures a single switch port. Entries are matched to the
// device's existing port overrides by portIdx; unspecified fields on a matched
// port are left untouched.
type DevicePortOverride struct {
	// PortIdx is the 1-based physical port number (required key).
	PortIdx int `pulumi:"portIdx"`
	// Name is the port label.
	Name *string `pulumi:"name,optional"`
	// PoeMode: auto | pasv24 | passthrough | off.
	PoeMode *string `pulumi:"poeMode,optional"`
	// PortProfileId applies a saved port profile (the profile's `_id`, sent as portconf_id).
	PortProfileId *string `pulumi:"portProfileId,optional"`
	// OpMode: switch | mirror | aggregate.
	OpMode *string `pulumi:"opMode,optional"`
	// AggregateNumPorts is the number of consecutive ports in a link aggregation group (1-8).
	AggregateNumPorts *int `pulumi:"aggregateNumPorts,optional"`
	// NativeNetworkId is the untagged/native network for the port (native_networkconf_id).
	NativeNetworkId *string `pulumi:"nativeNetworkId,optional"`
	// TaggedVlanMgmt: auto | block_all | custom. With "custom", use excludedNetworkIds.
	TaggedVlanMgmt *string `pulumi:"taggedVlanMgmt,optional"`
	// ExcludedNetworkIds lists tagged networks to block on the port (excluded_networkconf_ids).
	ExcludedNetworkIds []string `pulumi:"excludedNetworkIds,optional"`
	// Forward: all | native | customize | disabled.
	Forward *string `pulumi:"forward,optional"`
	// Speed forces a link speed in Mbps: 10 | 100 | 1000 | 2500 | 5000 | 10000 | 20000 | 25000 | 40000 | 50000 | 100000.
	Speed *int `pulumi:"speed,optional"`
	// Autoneg enables link speed/duplex auto-negotiation.
	Autoneg *bool `pulumi:"autoneg,optional"`
	// FullDuplex forces full-duplex when autoneg is off.
	FullDuplex *bool `pulumi:"fullDuplex,optional"`
	// Isolation enables port isolation (no traffic to other isolated ports).
	Isolation *bool `pulumi:"isolation,optional"`
	// MirrorPortIdx is the source port to mirror when opMode is "mirror".
	MirrorPortIdx *int `pulumi:"mirrorPortIdx,optional"`
	// StormctrlType: level | rate.
	StormctrlType *string `pulumi:"stormctrlType,optional"`
	// StormctrlBroadcastEnabled enables broadcast storm control.
	StormctrlBroadcastEnabled *bool `pulumi:"stormctrlBroadcastEnabled,optional"`
	// StormctrlBroadcastLevel is the broadcast storm control level (0-100, percent).
	StormctrlBroadcastLevel *int `pulumi:"stormctrlBroadcastLevel,optional"`
	// StormctrlBroadcastRate is the broadcast storm control rate (pps).
	StormctrlBroadcastRate *int `pulumi:"stormctrlBroadcastRate,optional"`
	// StormctrlMcastEnabled enables multicast storm control.
	StormctrlMcastEnabled *bool `pulumi:"stormctrlMcastEnabled,optional"`
	// StormctrlMcastLevel is the multicast storm control level (0-100, percent).
	StormctrlMcastLevel *int `pulumi:"stormctrlMcastLevel,optional"`
	// StormctrlMcastRate is the multicast storm control rate (pps).
	StormctrlMcastRate *int `pulumi:"stormctrlMcastRate,optional"`
	// StormctrlUcastEnabled enables unknown-unicast storm control.
	StormctrlUcastEnabled *bool `pulumi:"stormctrlUcastEnabled,optional"`
	// StormctrlUcastLevel is the unknown-unicast storm control level (0-100, percent).
	StormctrlUcastLevel *int `pulumi:"stormctrlUcastLevel,optional"`
	// StormctrlUcastRate is the unknown-unicast storm control rate (pps).
	StormctrlUcastRate *int `pulumi:"stormctrlUcastRate,optional"`
	// PortSecurityEnabled restricts the port to specific MAC addresses.
	PortSecurityEnabled *bool `pulumi:"portSecurityEnabled,optional"`
	// PortSecurityMacAddress lists the MACs permitted when portSecurityEnabled is true.
	PortSecurityMacAddress []string `pulumi:"portSecurityMacAddress,optional"`
	// Dot1xCtrl: auto | force_authorized | force_unauthorized | mac_based | multi_host.
	Dot1xCtrl *string `pulumi:"dot1xCtrl,optional"`
	// Dot1xIdleTimeout is the 802.1X idle timeout in seconds.
	Dot1xIdleTimeout *int `pulumi:"dot1xIdleTimeout,optional"`
	// LldpmedEnabled enables LLDP-MED on the port.
	LldpmedEnabled *bool `pulumi:"lldpmedEnabled,optional"`
	// LldpmedNotifyEnabled enables LLDP-MED topology-change notifications.
	LldpmedNotifyEnabled *bool `pulumi:"lldpmedNotifyEnabled,optional"`
	// StpPortMode enables spanning-tree on the port.
	StpPortMode *bool `pulumi:"stpPortMode,optional"`
	// EgressRateLimitKbps caps egress bandwidth in kbps (requires egressRateLimitKbpsEnabled).
	EgressRateLimitKbps *int `pulumi:"egressRateLimitKbps,optional"`
	// EgressRateLimitKbpsEnabled toggles the egress rate limit.
	EgressRateLimitKbpsEnabled *bool `pulumi:"egressRateLimitKbpsEnabled,optional"`
	// PriorityQueue1Level is the QoS priority-queue 1 level (0-100).
	PriorityQueue1Level *int `pulumi:"priorityQueue1Level,optional"`
	// PriorityQueue2Level is the QoS priority-queue 2 level (0-100).
	PriorityQueue2Level *int `pulumi:"priorityQueue2Level,optional"`
	// PriorityQueue3Level is the QoS priority-queue 3 level (0-100).
	PriorityQueue3Level *int `pulumi:"priorityQueue3Level,optional"`
	// PriorityQueue4Level is the QoS priority-queue 4 level (0-100).
	PriorityQueue4Level *int `pulumi:"priorityQueue4Level,optional"`
	// FecMode: rs-fec | fc-fec | default | disabled (for SFP+/SFP28 ports).
	FecMode *string `pulumi:"fecMode,optional"`
	// VoiceNetworkId is the voice VLAN network for VoIP phones (voice_networkconf_id).
	VoiceNetworkId *string `pulumi:"voiceNetworkId,optional"`
	// PortKeepaliveEnabled enables PoE keepalive for legacy powered devices.
	PortKeepaliveEnabled *bool `pulumi:"portKeepaliveEnabled,optional"`
	// SettingPreference: auto (inherit profile) | manual (use these overrides).
	SettingPreference *string `pulumi:"settingPreference,optional"`
}

// DeviceLed groups the status-LED indicator settings, including the
// EtherLighting LED strip on supported switches.
type DeviceLed struct {
	// Override controls the status LED: default | on | off.
	Override *string `pulumi:"override,optional"`
	// OverrideColor is the LED color as a hex string, e.g. "#0000ff".
	OverrideColor *string `pulumi:"overrideColor,optional"`
	// OverrideColorBrightness is the LED brightness (0-100).
	OverrideColorBrightness *int `pulumi:"overrideColorBrightness,optional"`
	// EtherLighting configures the EtherLighting LED strip (supported switches).
	EtherLighting *DeviceEtherLightingArgs `pulumi:"etherLighting,optional"`
}

// DeviceSnmp groups the SNMP agent strings.
type DeviceSnmp struct {
	// Contact is the SNMP contact string.
	Contact *string `pulumi:"contact,optional"`
	// Location is the SNMP location string.
	Location *string `pulumi:"location,optional"`
}

// DeviceStp groups the spanning-tree bridge settings (distinct from the per-port
// STP toggle in portOverrides).
type DeviceStp struct {
	// Priority is the spanning-tree bridge priority: a multiple of 4096 from 0 to 61440.
	Priority *string `pulumi:"priority,optional"`
	// Version: stp | rstp | disabled.
	Version *string `pulumi:"version,optional"`
}

// DeviceSwitching groups the switch-wide L2 toggles plus the device-wide default
// PoE mode (per-port PoE lives in portOverrides).
type DeviceSwitching struct {
	// VlanEnabled enables 802.1Q VLAN switching on the device.
	VlanEnabled *bool `pulumi:"vlanEnabled,optional"`
	// JumboFrameEnabled enables jumbo frames switch-wide.
	JumboFrameEnabled *bool `pulumi:"jumboFrameEnabled,optional"`
	// FlowControlEnabled enables 802.3x flow control switch-wide.
	FlowControlEnabled *bool `pulumi:"flowControlEnabled,optional"`
	// PoeMode is the device-wide default PoE mode: auto | pasv24 | passthrough | off.
	PoeMode *string `pulumi:"poeMode,optional"`
}

// DeviceDot1x groups the switch-wide 802.1X settings (distinct from the per-port
// 802.1X controls in portOverrides).
type DeviceDot1x struct {
	// PortControlEnabled enables 802.1X port control switch-wide.
	PortControlEnabled *bool `pulumi:"portControlEnabled,optional"`
	// FallbackNetworkId is the fallback network for failed 802.1X auth (the network's `_id`).
	FallbackNetworkId *string `pulumi:"fallbackNetworkId,optional"`
}

// DeviceOutlet groups the PDU outlet settings: the feature toggles plus the
// per-outlet overrides.
type DeviceOutlet struct {
	// Enabled enables PDU outlet control.
	Enabled *bool `pulumi:"enabled,optional"`
	// PowerCycleEnabled enables scheduled power cycling for PDU outlets.
	PowerCycleEnabled *bool `pulumi:"powerCycleEnabled,optional"`
	// Overrides configures individual PDU outlets.
	Overrides []DeviceOutletOverride `pulumi:"overrides,optional"`
}

// DeviceVrrp groups the gateway VRRP high-availability role/priority.
type DeviceVrrp struct {
	// Mode is the gateway VRRP role: primary | secondary.
	Mode *string `pulumi:"mode,optional"`
	// Priority is the VRRP priority (10-200).
	Priority *int `pulumi:"priority,optional"`
}

// DeviceLcm groups the front LCD/touchscreen display settings.
type DeviceLcm struct {
	// Brightness is the front display brightness (1-100). Setting it overrides the global default.
	Brightness *int `pulumi:"brightness,optional"`
	// IdleTimeout is the front display idle timeout in seconds (10-3600). Setting it overrides the global default.
	IdleTimeout *int `pulumi:"idleTimeout,optional"`
	// NightModeBegins is the night-mode start time, "HH:MM".
	NightModeBegins *string `pulumi:"nightModeBegins,optional"`
	// NightModeEnds is the night-mode end time, "HH:MM".
	NightModeEnds *string `pulumi:"nightModeEnds,optional"`
	// OrientationOverride rotates the front display: 0 | 90 | 180 | 270.
	OrientationOverride *int `pulumi:"orientationOverride,optional"`
}

// DeviceArgs are the user-supplied inputs. Every field is optional except the
// MAC, which binds the resource to an already-adopted device. Most settings are
// grouped into nested facet objects (led, snmp, stp, switching, dot1x, outlet,
// vrrp, lcm); the large per-port/radio/ethernet override lists stay top-level.
type DeviceArgs struct {
	// Mac is the device MAC address (required). The device must already be adopted on the controller.
	Mac string `pulumi:"mac" provider:"replaceOnChanges"`
	// Name is the device's display name.
	Name *string `pulumi:"name,optional"`
	// Disabled administratively disables the device.
	Disabled *bool `pulumi:"disabled,optional"`
	// MgmtNetworkId pins the device's management network (the network's `_id`).
	MgmtNetworkId *string `pulumi:"mgmtNetworkId,optional"`

	// Led groups the status-LED indicator settings.
	Led *DeviceLed `pulumi:"led,optional"`
	// Snmp groups the SNMP agent strings.
	Snmp *DeviceSnmp `pulumi:"snmp,optional"`
	// Stp groups the spanning-tree bridge settings.
	Stp *DeviceStp `pulumi:"stp,optional"`
	// Switching groups the switch-wide L2 settings.
	Switching *DeviceSwitching `pulumi:"switching,optional"`
	// Dot1x groups the switch-wide 802.1X settings.
	Dot1x *DeviceDot1x `pulumi:"dot1x,optional"`
	// Outlet groups the PDU outlet settings.
	Outlet *DeviceOutlet `pulumi:"outlet,optional"`
	// Vrrp groups the gateway VRRP high-availability settings.
	Vrrp *DeviceVrrp `pulumi:"vrrp,optional"`
	// Lcm groups the front LCD/touchscreen display settings.
	Lcm *DeviceLcm `pulumi:"lcm,optional"`

	// PortOverrides configures individual switch ports (the centerpiece for switches).
	PortOverrides []DevicePortOverride `pulumi:"portOverrides,optional"`
	// RadioTable tunes individual access-point radios.
	RadioTable []DeviceRadioOverride `pulumi:"radioTable,optional"`
	// EthernetOverrides assigns network groups to physical ethernet ports (gateways).
	EthernetOverrides []DeviceEthernetOverride `pulumi:"ethernetOverrides,optional"`
}

// DeviceState is the persisted state: inputs plus read-only device facts.
type DeviceState struct {
	DeviceArgs
	// DeviceId is the controller-assigned identifier (the UniFi `_id`).
	DeviceId string `pulumi:"deviceId"`
	// Model is the device model code, e.g. "USPRPS" or "U6-Enterprise" (read-only).
	Model string `pulumi:"model"`
	// Type is the device type, e.g. "usw", "uap", "ugw" (read-only).
	Type string `pulumi:"type"`
	// State is the connection state, e.g. "Connected", "Pending" (read-only).
	State string `pulumi:"state"`
	// Adopted indicates whether the controller has adopted the device (read-only).
	Adopted bool `pulumi:"adopted"`
}

func (d *Device) Annotate(a infer.Annotator) {
	a.Describe(&d, "Manage settings of an existing, already-adopted UniFi network device "+
		"(gateway, switch, access point, PDU). Adoption model: binds to a device by MAC and "+
		"read-modify-writes only the managed fields onto the controller's current device object; "+
		"unmanaged settings and unlisted port/outlet/radio overrides are preserved. Delete is a no-op.")
}

func (e *DeviceEtherLightingArgs) Annotate(a infer.Annotator) {
	a.Describe(&e.LedMode, "LedMode: standard | etherlighting.")
	a.Describe(&e.Mode, "Mode of the lighting animation: speed | network.")
	a.Describe(&e.Behavior, "Behavior of the lighting: breath | steady.")
	a.Describe(&e.Brightness, "Brightness of the LED strip (1-100).")
}

func (o *DeviceOutletOverride) Annotate(a infer.Annotator) {
	a.Describe(&o.Index, "Index is the 1-based outlet number (required key).")
	a.Describe(&o.Name, "Name is the outlet label.")
	a.Describe(&o.RelayState, "RelayState powers the outlet on (true) or off (false).")
	a.Describe(&o.CycleEnabled, "CycleEnabled allows the outlet to be power-cycled.")
}

func (e *DeviceEthernetOverride) Annotate(a infer.Annotator) {
	a.Describe(&e.Ifname, "Ifname is the interface name, e.g. \"eth1\" (required key).")
	a.Describe(&e.NetworkGroup, "NetworkGroup is the assigned group, e.g. \"LAN\", \"LAN2\", \"WAN\", \"WAN2\".")
}

func (r *DeviceRadioOverride) Annotate(a infer.Annotator) {
	a.Describe(&r.Radio, "Radio band identifying which radio to tune: ng (2.4GHz) | na (5GHz) | ad (60GHz) | 6e (6GHz). Required key.")
	a.Describe(&r.Name, "Name is the controller's radio name (e.g. \"wifi0\").")
	a.Describe(&r.Channel, "Channel is the radio channel, a number or \"auto\".")
	a.Describe(&r.ChannelWidth, "ChannelWidth (HT) in MHz: 20 | 40 | 80 | 160 | 240 | 320 | 1080 | 2160 | 4320.")
	a.Describe(&r.TxPower, "TxPower is the transmit power in dBm, or \"auto\". Honored only when txPowerMode is \"custom\".")
	a.Describe(&r.TxPowerMode, "TxPowerMode: auto | medium | high | low | custom.")
	a.Describe(&r.MinRssi, "MinRssi is the minimum RSSI (dBm, negative) below which clients are kicked. Setting it enables the feature unless minRssiEnabled is given.")
	a.Describe(&r.MinRssiEnabled, "MinRssiEnabled toggles the minimum-RSSI kick feature.")
	a.Describe(&r.AntennaGain, "AntennaGain in dBi (for devices with configurable antennas).")
}

func (p *DevicePortOverride) Annotate(a infer.Annotator) {
	a.Describe(&p.PortIdx, "PortIdx is the 1-based physical port number (required key).")
	a.Describe(&p.Name, "Name is the port label.")
	a.Describe(&p.PoeMode, "PoeMode: auto | pasv24 | passthrough | off.")
	a.Describe(&p.PortProfileId, "PortProfileId applies a saved port profile (the profile's `_id`, sent as portconf_id).")
	a.Describe(&p.OpMode, "OpMode: switch | mirror | aggregate.")
	a.Describe(&p.AggregateNumPorts, "AggregateNumPorts is the number of consecutive ports in a link aggregation group (1-8).")
	a.Describe(&p.NativeNetworkId, "NativeNetworkId is the untagged/native network for the port (native_networkconf_id).")
	a.Describe(&p.TaggedVlanMgmt, "TaggedVlanMgmt: auto | block_all | custom. With \"custom\", use excludedNetworkIds.")
	a.Describe(&p.ExcludedNetworkIds, "ExcludedNetworkIds lists tagged networks to block on the port (excluded_networkconf_ids).")
	a.Describe(&p.Forward, "Forward: all | native | customize | disabled.")
	a.Describe(&p.Speed, "Speed forces a link speed in Mbps: 10 | 100 | 1000 | 2500 | 5000 | 10000 | 20000 | 25000 | 40000 | 50000 | 100000.")
	a.Describe(&p.Autoneg, "Autoneg enables link speed/duplex auto-negotiation.")
	a.Describe(&p.FullDuplex, "FullDuplex forces full-duplex when autoneg is off.")
	a.Describe(&p.Isolation, "Isolation enables port isolation (no traffic to other isolated ports).")
	a.Describe(&p.MirrorPortIdx, "MirrorPortIdx is the source port to mirror when opMode is \"mirror\".")
	a.Describe(&p.StormctrlType, "StormctrlType: level | rate.")
	a.Describe(&p.StormctrlBroadcastEnabled, "StormctrlBroadcastEnabled enables broadcast storm control.")
	a.Describe(&p.StormctrlBroadcastLevel, "StormctrlBroadcastLevel is the broadcast storm control level (0-100, percent).")
	a.Describe(&p.StormctrlBroadcastRate, "StormctrlBroadcastRate is the broadcast storm control rate (pps).")
	a.Describe(&p.StormctrlMcastEnabled, "StormctrlMcastEnabled enables multicast storm control.")
	a.Describe(&p.StormctrlMcastLevel, "StormctrlMcastLevel is the multicast storm control level (0-100, percent).")
	a.Describe(&p.StormctrlMcastRate, "StormctrlMcastRate is the multicast storm control rate (pps).")
	a.Describe(&p.StormctrlUcastEnabled, "StormctrlUcastEnabled enables unknown-unicast storm control.")
	a.Describe(&p.StormctrlUcastLevel, "StormctrlUcastLevel is the unknown-unicast storm control level (0-100, percent).")
	a.Describe(&p.StormctrlUcastRate, "StormctrlUcastRate is the unknown-unicast storm control rate (pps).")
	a.Describe(&p.PortSecurityEnabled, "PortSecurityEnabled restricts the port to specific MAC addresses.")
	a.Describe(&p.PortSecurityMacAddress, "PortSecurityMacAddress lists the MACs permitted when portSecurityEnabled is true.")
	a.Describe(&p.Dot1xCtrl, "Dot1xCtrl: auto | force_authorized | force_unauthorized | mac_based | multi_host.")
	a.Describe(&p.Dot1xIdleTimeout, "Dot1xIdleTimeout is the 802.1X idle timeout in seconds.")
	a.Describe(&p.LldpmedEnabled, "LldpmedEnabled enables LLDP-MED on the port.")
	a.Describe(&p.LldpmedNotifyEnabled, "LldpmedNotifyEnabled enables LLDP-MED topology-change notifications.")
	a.Describe(&p.StpPortMode, "StpPortMode enables spanning-tree on the port.")
	a.Describe(&p.EgressRateLimitKbps, "EgressRateLimitKbps caps egress bandwidth in kbps (requires egressRateLimitKbpsEnabled).")
	a.Describe(&p.EgressRateLimitKbpsEnabled, "EgressRateLimitKbpsEnabled toggles the egress rate limit.")
	a.Describe(&p.PriorityQueue1Level, "PriorityQueue1Level is the QoS priority-queue 1 level (0-100).")
	a.Describe(&p.PriorityQueue2Level, "PriorityQueue2Level is the QoS priority-queue 2 level (0-100).")
	a.Describe(&p.PriorityQueue3Level, "PriorityQueue3Level is the QoS priority-queue 3 level (0-100).")
	a.Describe(&p.PriorityQueue4Level, "PriorityQueue4Level is the QoS priority-queue 4 level (0-100).")
	a.Describe(&p.FecMode, "FecMode: rs-fec | fc-fec | default | disabled (for SFP+/SFP28 ports).")
	a.Describe(&p.VoiceNetworkId, "VoiceNetworkId is the voice VLAN network for VoIP phones (voice_networkconf_id).")
	a.Describe(&p.PortKeepaliveEnabled, "PortKeepaliveEnabled enables PoE keepalive for legacy powered devices.")
	a.Describe(&p.SettingPreference, "SettingPreference: auto (inherit profile) | manual (use these overrides).")
}

func (l *DeviceLed) Annotate(a infer.Annotator) {
	a.Describe(&l.Override, "Override controls the status LED: default | on | off.")
	a.Describe(&l.OverrideColor, "OverrideColor is the LED color as a hex string, e.g. \"#0000ff\".")
	a.Describe(&l.OverrideColorBrightness, "OverrideColorBrightness is the LED brightness (0-100).")
	a.Describe(&l.EtherLighting, "EtherLighting configures the EtherLighting LED strip (supported switches).")
}

func (s *DeviceSnmp) Annotate(a infer.Annotator) {
	a.Describe(&s.Contact, "Contact is the SNMP contact string.")
	a.Describe(&s.Location, "Location is the SNMP location string.")
}

func (s *DeviceStp) Annotate(a infer.Annotator) {
	a.Describe(&s.Priority, "Priority is the spanning-tree bridge priority: a multiple of 4096 from 0 to 61440.")
	a.Describe(&s.Version, "Version: stp | rstp | disabled.")
}

func (s *DeviceSwitching) Annotate(a infer.Annotator) {
	a.Describe(&s.VlanEnabled, "VlanEnabled enables 802.1Q VLAN switching on the device.")
	a.Describe(&s.JumboFrameEnabled, "JumboFrameEnabled enables jumbo frames switch-wide.")
	a.Describe(&s.FlowControlEnabled, "FlowControlEnabled enables 802.3x flow control switch-wide.")
	a.Describe(&s.PoeMode, "PoeMode is the device-wide default PoE mode: auto | pasv24 | passthrough | off.")
}

func (x *DeviceDot1x) Annotate(a infer.Annotator) {
	a.Describe(&x.PortControlEnabled, "PortControlEnabled enables 802.1X port control switch-wide.")
	a.Describe(&x.FallbackNetworkId, "FallbackNetworkId is the fallback network for failed 802.1X auth (the network's `_id`).")
}

func (o *DeviceOutlet) Annotate(a infer.Annotator) {
	a.Describe(&o.Enabled, "Enabled enables PDU outlet control.")
	a.Describe(&o.PowerCycleEnabled, "PowerCycleEnabled enables scheduled power cycling for PDU outlets.")
	a.Describe(&o.Overrides, "Overrides configures individual PDU outlets.")
}

func (v *DeviceVrrp) Annotate(a infer.Annotator) {
	a.Describe(&v.Mode, "Mode is the gateway VRRP role: primary | secondary.")
	a.Describe(&v.Priority, "Priority is the VRRP priority (10-200).")
}

func (l *DeviceLcm) Annotate(a infer.Annotator) {
	a.Describe(&l.Brightness, "Brightness is the front display brightness (1-100). Setting it overrides the global default.")
	a.Describe(&l.IdleTimeout, "IdleTimeout is the front display idle timeout in seconds (10-3600). Setting it overrides the global default.")
	a.Describe(&l.NightModeBegins, "NightModeBegins is the night-mode start time, \"HH:MM\".")
	a.Describe(&l.NightModeEnds, "NightModeEnds is the night-mode end time, \"HH:MM\".")
	a.Describe(&l.OrientationOverride, "OrientationOverride rotates the front display: 0 | 90 | 180 | 270.")
}

func (d *DeviceArgs) Annotate(a infer.Annotator) {
	a.Describe(&d.Mac, "Mac is the device MAC address (required). The device must already be adopted on the controller.")
	a.Describe(&d.Name, "Name is the device's display name.")
	a.Describe(&d.Disabled, "Disabled administratively disables the device.")
	a.Describe(&d.MgmtNetworkId, "MgmtNetworkId pins the device's management network (the network's `_id`).")
	a.Describe(&d.Led, "Led groups the status-LED indicator settings.")
	a.Describe(&d.Snmp, "Snmp groups the SNMP agent strings.")
	a.Describe(&d.Stp, "Stp groups the spanning-tree bridge settings.")
	a.Describe(&d.Switching, "Switching groups the switch-wide L2 settings.")
	a.Describe(&d.Dot1x, "Dot1x groups the switch-wide 802.1X settings.")
	a.Describe(&d.Outlet, "Outlet groups the PDU outlet settings.")
	a.Describe(&d.Vrrp, "Vrrp groups the gateway VRRP high-availability settings.")
	a.Describe(&d.Lcm, "Lcm groups the front LCD/touchscreen display settings.")
	a.Describe(&d.PortOverrides, "PortOverrides configures individual switch ports (the centerpiece for switches).")
	a.Describe(&d.RadioTable, "RadioTable tunes individual access-point radios.")
	a.Describe(&d.EthernetOverrides, "EthernetOverrides assigns network groups to physical ethernet ports (gateways).")
}

func (s *DeviceState) Annotate(a infer.Annotator) {
	a.Describe(&s.DeviceId, "DeviceId is the controller-assigned identifier (the UniFi `_id`).")
	a.Describe(&s.Model, "Model is the device model code, e.g. \"USPRPS\" or \"U6-Enterprise\" (read-only).")
	a.Describe(&s.Type, "Type is the device type, e.g. \"usw\", \"uap\", \"ugw\" (read-only).")
	a.Describe(&s.State, "State is the connection state, e.g. \"Connected\", \"Pending\" (read-only).")
	a.Describe(&s.Adopted, "Adopted indicates whether the controller has adopted the device (read-only).")
}

// applyTo overlays the managed input fields onto a device fetched from the
// controller. It NEVER builds a device from zero, so unmanaged settings survive.
// Each nested facet group is an optional pointer: an absent group means "manage
// nothing in this facet", so its applyTo runs only when the group is non-nil.
func (a DeviceArgs) applyTo(d *unifi.Device) {
	if a.Name != nil {
		d.Name = *a.Name
	}
	if a.Disabled != nil {
		d.Disabled = *a.Disabled
	}
	if a.MgmtNetworkId != nil {
		d.MgmtNetworkID = *a.MgmtNetworkId
	}

	if a.Led != nil {
		a.Led.applyTo(d)
	}
	if a.Snmp != nil {
		a.Snmp.applyTo(d)
	}
	if a.Stp != nil {
		a.Stp.applyTo(d)
	}
	if a.Switching != nil {
		a.Switching.applyTo(d)
	}
	if a.Dot1x != nil {
		a.Dot1x.applyTo(d)
	}
	if a.Outlet != nil {
		a.Outlet.applyTo(d)
	}
	if a.Vrrp != nil {
		a.Vrrp.applyTo(d)
	}
	if a.Lcm != nil {
		a.Lcm.applyTo(d)
	}

	for _, po := range a.PortOverrides {
		po.applyTo(devicePortOverrideRef(d, po.PortIdx))
	}
	for _, ro := range a.RadioTable {
		ro.applyTo(deviceRadioRef(d, ro.Radio))
	}
	for _, eo := range a.EthernetOverrides {
		eo.applyTo(deviceEthernetRef(d, eo.Ifname))
	}
}

// applyTo overlays the managed LED settings (including the EtherLighting strip).
func (l DeviceLed) applyTo(d *unifi.Device) {
	if l.Override != nil {
		d.LedOverride = *l.Override
	}
	if l.OverrideColor != nil {
		d.LedOverrideColor = *l.OverrideColor
	}
	if l.OverrideColorBrightness != nil {
		d.LedOverrideColorBrightness = *l.OverrideColorBrightness
	}
	if l.EtherLighting != nil {
		l.EtherLighting.applyTo(&d.EtherLighting)
	}
}

// applyTo overlays the managed SNMP agent strings.
func (s DeviceSnmp) applyTo(d *unifi.Device) {
	if s.Contact != nil {
		d.SnmpContact = *s.Contact
	}
	if s.Location != nil {
		d.SnmpLocation = *s.Location
	}
}

// applyTo overlays the managed spanning-tree bridge settings.
func (s DeviceStp) applyTo(d *unifi.Device) {
	if s.Priority != nil {
		d.StpPriority = *s.Priority
	}
	if s.Version != nil {
		d.StpVersion = *s.Version
	}
}

// applyTo overlays the managed switch-wide L2 settings.
func (s DeviceSwitching) applyTo(d *unifi.Device) {
	if s.VlanEnabled != nil {
		d.SwitchVLANEnabled = *s.VlanEnabled
	}
	if s.JumboFrameEnabled != nil {
		d.JumboframeEnabled = *s.JumboFrameEnabled
	}
	if s.FlowControlEnabled != nil {
		d.FlowctrlEnabled = *s.FlowControlEnabled
	}
	if s.PoeMode != nil {
		d.PoeMode = *s.PoeMode
	}
}

// applyTo overlays the managed switch-wide 802.1X settings.
func (x DeviceDot1x) applyTo(d *unifi.Device) {
	if x.PortControlEnabled != nil {
		d.Dot1XPortctrlEnabled = *x.PortControlEnabled
	}
	if x.FallbackNetworkId != nil {
		d.Dot1XFallbackNetworkID = *x.FallbackNetworkId
	}
}

// applyTo overlays the managed PDU outlet settings, appending/overlaying the
// per-outlet overrides keyed by index.
func (o DeviceOutlet) applyTo(d *unifi.Device) {
	if o.Enabled != nil {
		d.OutletEnabled = *o.Enabled
	}
	if o.PowerCycleEnabled != nil {
		d.OutletPowerCycleEnabled = *o.PowerCycleEnabled
	}
	for _, oo := range o.Overrides {
		oo.applyTo(deviceOutletOverrideRef(d, oo.Index))
	}
}

// applyTo overlays the managed gateway VRRP settings.
func (v DeviceVrrp) applyTo(d *unifi.Device) {
	if v.Mode != nil {
		d.GatewayVrrpMode = *v.Mode
	}
	if v.Priority != nil {
		d.GatewayVrrpPriority = *v.Priority
	}
}

// applyTo overlays the managed front-display settings. Setting brightness or the
// idle timeout also flips the matching controller override bool on, mirroring the
// console behavior (these *Override flags have no input field of their own).
func (l DeviceLcm) applyTo(d *unifi.Device) {
	if l.Brightness != nil {
		d.LcmBrightness = *l.Brightness
		d.LcmBrightnessOverride = true
	}
	if l.IdleTimeout != nil {
		d.LcmIDleTimeout = *l.IdleTimeout
		d.LcmIDleTimeoutOverride = true
	}
	if l.NightModeBegins != nil {
		d.LcmNightModeBegins = *l.NightModeBegins
	}
	if l.NightModeEnds != nil {
		d.LcmNightModeEnds = *l.NightModeEnds
	}
	if l.OrientationOverride != nil {
		d.LcmOrientationOverride = *l.OrientationOverride
	}
}

func (e DeviceEtherLightingArgs) applyTo(u *unifi.DeviceEtherLighting) {
	if e.LedMode != nil {
		u.LedMode = *e.LedMode
	}
	if e.Mode != nil {
		u.Mode = *e.Mode
	}
	if e.Behavior != nil {
		u.Behavior = *e.Behavior
	}
	if e.Brightness != nil {
		u.Brightness = *e.Brightness
	}
}

func (o DeviceOutletOverride) applyTo(u *unifi.DeviceOutletOverrides) {
	u.Index = o.Index
	if o.Name != nil {
		u.Name = *o.Name
	}
	if o.RelayState != nil {
		u.RelayState = *o.RelayState
	}
	if o.CycleEnabled != nil {
		u.CycleEnabled = *o.CycleEnabled
	}
}

func (e DeviceEthernetOverride) applyTo(u *unifi.DeviceEthernetOverrides) {
	u.Ifname = e.Ifname
	if e.NetworkGroup != nil {
		u.NetworkGroup = *e.NetworkGroup
	}
}

func (r DeviceRadioOverride) applyTo(u *unifi.DeviceRadioTable) {
	u.Radio = r.Radio
	if r.Name != nil {
		u.Name = *r.Name
	}
	if r.Channel != nil {
		u.Channel = *r.Channel
	}
	if r.ChannelWidth != nil {
		u.Ht = *r.ChannelWidth
	}
	if r.TxPower != nil {
		u.TxPower = *r.TxPower
	}
	if r.TxPowerMode != nil {
		u.TxPowerMode = *r.TxPowerMode
	}
	if r.MinRssi != nil {
		u.MinRssi = *r.MinRssi
		if r.MinRssiEnabled == nil {
			u.MinRssiEnabled = true
		}
	}
	if r.MinRssiEnabled != nil {
		u.MinRssiEnabled = *r.MinRssiEnabled
	}
	if r.AntennaGain != nil {
		u.AntennaGain = *r.AntennaGain
	}
}

func (p DevicePortOverride) applyTo(u *unifi.DevicePortOverrides) {
	u.PortIDX = p.PortIdx
	if p.Name != nil {
		u.Name = *p.Name
	}
	if p.PoeMode != nil {
		u.PoeMode = *p.PoeMode
	}
	if p.PortProfileId != nil {
		u.PortProfileID = *p.PortProfileId
	}
	if p.OpMode != nil {
		u.OpMode = *p.OpMode
	}
	if p.AggregateNumPorts != nil {
		u.AggregateNumPorts = *p.AggregateNumPorts
	}
	if p.NativeNetworkId != nil {
		u.NATiveNetworkID = *p.NativeNetworkId
	}
	if p.TaggedVlanMgmt != nil {
		u.TaggedVLANMgmt = *p.TaggedVlanMgmt
	}
	if p.ExcludedNetworkIds != nil {
		u.ExcludedNetworkIDs = p.ExcludedNetworkIds
	}
	if p.Forward != nil {
		u.Forward = *p.Forward
	}
	if p.Speed != nil {
		u.Speed = *p.Speed
	}
	if p.Autoneg != nil {
		u.Autoneg = *p.Autoneg
	}
	if p.FullDuplex != nil {
		u.FullDuplex = *p.FullDuplex
	}
	if p.Isolation != nil {
		u.Isolation = *p.Isolation
	}
	if p.MirrorPortIdx != nil {
		u.MirrorPortIDX = *p.MirrorPortIdx
	}
	if p.StormctrlType != nil {
		u.StormctrlType = *p.StormctrlType
	}
	if p.StormctrlBroadcastEnabled != nil {
		u.StormctrlBroadcastastEnabled = *p.StormctrlBroadcastEnabled
	}
	if p.StormctrlBroadcastLevel != nil {
		u.StormctrlBroadcastastLevel = *p.StormctrlBroadcastLevel
	}
	if p.StormctrlBroadcastRate != nil {
		u.StormctrlBroadcastastRate = *p.StormctrlBroadcastRate
	}
	if p.StormctrlMcastEnabled != nil {
		u.StormctrlMcastEnabled = *p.StormctrlMcastEnabled
	}
	if p.StormctrlMcastLevel != nil {
		u.StormctrlMcastLevel = *p.StormctrlMcastLevel
	}
	if p.StormctrlMcastRate != nil {
		u.StormctrlMcastRate = *p.StormctrlMcastRate
	}
	if p.StormctrlUcastEnabled != nil {
		u.StormctrlUcastEnabled = *p.StormctrlUcastEnabled
	}
	if p.StormctrlUcastLevel != nil {
		u.StormctrlUcastLevel = *p.StormctrlUcastLevel
	}
	if p.StormctrlUcastRate != nil {
		u.StormctrlUcastRate = *p.StormctrlUcastRate
	}
	if p.PortSecurityEnabled != nil {
		u.PortSecurityEnabled = *p.PortSecurityEnabled
	}
	if p.PortSecurityMacAddress != nil {
		u.PortSecurityMACAddress = p.PortSecurityMacAddress
	}
	if p.Dot1xCtrl != nil {
		u.Dot1XCtrl = *p.Dot1xCtrl
	}
	if p.Dot1xIdleTimeout != nil {
		u.Dot1XIDleTimeout = *p.Dot1xIdleTimeout
	}
	if p.LldpmedEnabled != nil {
		u.LldpmedEnabled = *p.LldpmedEnabled
	}
	if p.LldpmedNotifyEnabled != nil {
		u.LldpmedNotifyEnabled = *p.LldpmedNotifyEnabled
	}
	if p.StpPortMode != nil {
		u.StpPortMode = *p.StpPortMode
	}
	if p.EgressRateLimitKbps != nil {
		u.EgressRateLimitKbps = *p.EgressRateLimitKbps
	}
	if p.EgressRateLimitKbpsEnabled != nil {
		u.EgressRateLimitKbpsEnabled = *p.EgressRateLimitKbpsEnabled
	}
	if p.PriorityQueue1Level != nil {
		u.PriorityQueue1Level = *p.PriorityQueue1Level
	}
	if p.PriorityQueue2Level != nil {
		u.PriorityQueue2Level = *p.PriorityQueue2Level
	}
	if p.PriorityQueue3Level != nil {
		u.PriorityQueue3Level = *p.PriorityQueue3Level
	}
	if p.PriorityQueue4Level != nil {
		u.PriorityQueue4Level = *p.PriorityQueue4Level
	}
	if p.FecMode != nil {
		u.FecMode = *p.FecMode
	}
	if p.VoiceNetworkId != nil {
		u.VoiceNetworkID = *p.VoiceNetworkId
	}
	if p.PortKeepaliveEnabled != nil {
		u.PortKeepaliveEnabled = *p.PortKeepaliveEnabled
	}
	if p.SettingPreference != nil {
		u.SettingPreference = *p.SettingPreference
	}
}

// devicePortOverrideRef returns a pointer to the existing port override with the
// given index, appending a new one (keyed by portIdx) if none exists.
func devicePortOverrideRef(d *unifi.Device, idx int) *unifi.DevicePortOverrides {
	for i := range d.PortOverrides {
		if d.PortOverrides[i].PortIDX == idx {
			return &d.PortOverrides[i]
		}
	}
	d.PortOverrides = append(d.PortOverrides, unifi.DevicePortOverrides{PortIDX: idx})
	return &d.PortOverrides[len(d.PortOverrides)-1]
}

// deviceOutletOverrideRef returns a pointer to the existing outlet override with
// the given index, appending a new one if none exists.
func deviceOutletOverrideRef(d *unifi.Device, idx int) *unifi.DeviceOutletOverrides {
	for i := range d.OutletOverrides {
		if d.OutletOverrides[i].Index == idx {
			return &d.OutletOverrides[i]
		}
	}
	d.OutletOverrides = append(d.OutletOverrides, unifi.DeviceOutletOverrides{Index: idx})
	return &d.OutletOverrides[len(d.OutletOverrides)-1]
}

// deviceRadioRef returns a pointer to the existing radio with the given band,
// appending a new one if none exists.
func deviceRadioRef(d *unifi.Device, radio string) *unifi.DeviceRadioTable {
	for i := range d.RadioTable {
		if d.RadioTable[i].Radio == radio {
			return &d.RadioTable[i]
		}
	}
	d.RadioTable = append(d.RadioTable, unifi.DeviceRadioTable{Radio: radio})
	return &d.RadioTable[len(d.RadioTable)-1]
}

// deviceEthernetRef returns a pointer to the existing ethernet override with the
// given interface name, appending a new one if none exists.
func deviceEthernetRef(d *unifi.Device, ifname string) *unifi.DeviceEthernetOverrides {
	for i := range d.EthernetOverrides {
		if d.EthernetOverrides[i].Ifname == ifname {
			return &d.EthernetOverrides[i]
		}
	}
	d.EthernetOverrides = append(d.EthernetOverrides, unifi.DeviceEthernetOverrides{Ifname: ifname})
	return &d.EthernetOverrides[len(d.EthernetOverrides)-1]
}

// deviceStateFrom maps a controller device back into resource state. The managed
// input fields are taken from prior (the adoption model preserves the program's
// declared values, avoiding spurious diffs); managed scalar settings are
// refreshed from the device so genuine drift surfaces. The nested override lists
// are preserved from prior rather than reconstructed. Read-only facts come from
// the device.
func deviceStateFrom(d *unifi.Device, prior DeviceArgs) DeviceState {
	args := prior
	if args.Mac == "" {
		args.Mac = d.MAC
	}
	if prior.Name != nil {
		args.Name = ptr(d.Name)
	}
	if prior.Disabled != nil {
		args.Disabled = ptr(d.Disabled)
	}
	if prior.MgmtNetworkId != nil {
		args.MgmtNetworkId = ptr(d.MgmtNetworkID)
	}

	// Reflect each managed scalar from the device per nested member, refreshing
	// only the members the program declared. A group is allocated in the output
	// solely when prior carried it (so unmanaged facets round-trip as nil and do
	// not produce spurious diffs). The already-nested members preserved from
	// prior (led.etherLighting, outlet.overrides) are carried in the struct copy.
	if prior.Led != nil {
		led := *prior.Led
		if prior.Led.Override != nil {
			led.Override = ptr(d.LedOverride)
		}
		if prior.Led.OverrideColor != nil {
			led.OverrideColor = ptr(d.LedOverrideColor)
		}
		if prior.Led.OverrideColorBrightness != nil {
			led.OverrideColorBrightness = ptr(d.LedOverrideColorBrightness)
		}
		args.Led = &led
	}
	if prior.Snmp != nil {
		snmp := *prior.Snmp
		if prior.Snmp.Contact != nil {
			snmp.Contact = ptr(d.SnmpContact)
		}
		if prior.Snmp.Location != nil {
			snmp.Location = ptr(d.SnmpLocation)
		}
		args.Snmp = &snmp
	}
	if prior.Stp != nil {
		stp := *prior.Stp
		if prior.Stp.Priority != nil {
			stp.Priority = ptr(d.StpPriority)
		}
		if prior.Stp.Version != nil {
			stp.Version = ptr(d.StpVersion)
		}
		args.Stp = &stp
	}
	if prior.Switching != nil {
		switching := *prior.Switching
		if prior.Switching.VlanEnabled != nil {
			switching.VlanEnabled = ptr(d.SwitchVLANEnabled)
		}
		if prior.Switching.JumboFrameEnabled != nil {
			switching.JumboFrameEnabled = ptr(d.JumboframeEnabled)
		}
		if prior.Switching.FlowControlEnabled != nil {
			switching.FlowControlEnabled = ptr(d.FlowctrlEnabled)
		}
		if prior.Switching.PoeMode != nil {
			switching.PoeMode = ptr(d.PoeMode)
		}
		args.Switching = &switching
	}
	if prior.Dot1x != nil {
		dot1x := *prior.Dot1x
		if prior.Dot1x.PortControlEnabled != nil {
			dot1x.PortControlEnabled = ptr(d.Dot1XPortctrlEnabled)
		}
		if prior.Dot1x.FallbackNetworkId != nil {
			dot1x.FallbackNetworkId = ptr(d.Dot1XFallbackNetworkID)
		}
		args.Dot1x = &dot1x
	}
	if prior.Outlet != nil {
		outlet := *prior.Outlet
		if prior.Outlet.Enabled != nil {
			outlet.Enabled = ptr(d.OutletEnabled)
		}
		if prior.Outlet.PowerCycleEnabled != nil {
			outlet.PowerCycleEnabled = ptr(d.OutletPowerCycleEnabled)
		}
		args.Outlet = &outlet
	}
	if prior.Vrrp != nil {
		vrrp := *prior.Vrrp
		if prior.Vrrp.Mode != nil {
			vrrp.Mode = ptr(d.GatewayVrrpMode)
		}
		if prior.Vrrp.Priority != nil {
			vrrp.Priority = ptr(d.GatewayVrrpPriority)
		}
		args.Vrrp = &vrrp
	}
	if prior.Lcm != nil {
		lcm := *prior.Lcm
		if prior.Lcm.Brightness != nil {
			lcm.Brightness = ptr(d.LcmBrightness)
		}
		if prior.Lcm.IdleTimeout != nil {
			lcm.IdleTimeout = ptr(d.LcmIDleTimeout)
		}
		if prior.Lcm.NightModeBegins != nil {
			lcm.NightModeBegins = ptr(d.LcmNightModeBegins)
		}
		if prior.Lcm.NightModeEnds != nil {
			lcm.NightModeEnds = ptr(d.LcmNightModeEnds)
		}
		if prior.Lcm.OrientationOverride != nil {
			lcm.OrientationOverride = ptr(d.LcmOrientationOverride)
		}
		args.Lcm = &lcm
	}
	return DeviceState{
		DeviceArgs: args,
		DeviceId:   d.ID,
		Model:      d.Model,
		Type:       d.Type,
		State:      d.State.String(),
		Adopted:    d.Adopted,
	}
}

// Create binds the resource to an already-adopted device and applies settings.
func (Device) Create(ctx context.Context, req infer.CreateRequest[DeviceArgs]) (infer.CreateResponse[DeviceState], error) {
	if req.DryRun {
		return infer.CreateResponse[DeviceState]{Output: DeviceState{DeviceArgs: req.Inputs}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	d, err := cfg.Network().GetDeviceByMAC(ctx, cfg.ResolvedSite(), req.Inputs.Mac)
	if err != nil {
		return infer.CreateResponse[DeviceState]{}, fmt.Errorf("device %q must already be adopted on the controller: %w", req.Inputs.Mac, err)
	}
	if !d.Adopted {
		return infer.CreateResponse[DeviceState]{}, fmt.Errorf("device %q (model %s) is present but not adopted; adopt it in the UniFi console first", req.Inputs.Mac, d.Model)
	}
	req.Inputs.applyTo(d)
	updated, err := cfg.Network().UpdateDevice(ctx, cfg.ResolvedSite(), d)
	if err != nil {
		return infer.CreateResponse[DeviceState]{}, wrap(fmt.Sprintf("create device %q (site %q)", req.Inputs.Mac, cfg.ResolvedSite()), err)
	}
	return infer.CreateResponse[DeviceState]{ID: updated.ID, Output: deviceStateFrom(updated, req.Inputs)}, nil
}

// Read recovers state from the controller (also enables `pulumi import`).
func (Device) Read(ctx context.Context, req infer.ReadRequest[DeviceArgs, DeviceState]) (infer.ReadResponse[DeviceArgs, DeviceState], error) {
	cfg := infer.GetConfig[config.Config](ctx)
	d, err := cfg.Network().GetDevice(ctx, cfg.ResolvedSite(), req.ID)
	if notFound(err) {
		return infer.ReadResponse[DeviceArgs, DeviceState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[DeviceArgs, DeviceState]{}, wrap(fmt.Sprintf("read device %q (site %q)", req.ID, cfg.ResolvedSite()), err)
	}
	st := deviceStateFrom(d, req.State.DeviceArgs)
	return infer.ReadResponse[DeviceArgs, DeviceState]{ID: req.ID, Inputs: st.DeviceArgs, State: st}, nil
}

// Update read-modify-writes the managed fields onto the controller's device.
func (Device) Update(ctx context.Context, req infer.UpdateRequest[DeviceArgs, DeviceState]) (infer.UpdateResponse[DeviceState], error) {
	if req.DryRun {
		return infer.UpdateResponse[DeviceState]{Output: DeviceState{DeviceArgs: req.Inputs, DeviceId: req.ID}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	d, err := cfg.Network().GetDevice(ctx, cfg.ResolvedSite(), req.ID)
	if err != nil {
		return infer.UpdateResponse[DeviceState]{}, wrap(fmt.Sprintf("get device %q (site %q)", req.ID, cfg.ResolvedSite()), err)
	}
	req.Inputs.applyTo(d)
	updated, err := cfg.Network().UpdateDevice(ctx, cfg.ResolvedSite(), d)
	if err != nil {
		return infer.UpdateResponse[DeviceState]{}, wrap(fmt.Sprintf("update device %q (site %q)", req.ID, cfg.ResolvedSite()), err)
	}
	return infer.UpdateResponse[DeviceState]{Output: deviceStateFrom(updated, req.Inputs)}, nil
}

// Delete is a no-op: the physical device is left in place and adopted, only
// released from Pulumi's management. Its current settings are not reverted.
func (Device) Delete(_ context.Context, _ infer.DeleteRequest[DeviceState]) (infer.DeleteResponse, error) {
	return infer.DeleteResponse{}, nil
}
