// SPDX-License-Identifier: Apache-2.0

package network

import (
	"context"
	"fmt"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/ryanwersal/pulumi-unifi/provider/config"
)

// PortProfile is the controlling (marker) struct for a UniFi switch port
// profile. Port profiles are reusable port configurations that a switch
// Device's per-port override references via portProfileId (portconf_id).
type PortProfile struct{}

// PortProfileVlan groups the VLAN / forwarding settings for the port profile.
type PortProfileVlan struct {
	// Forward sets the VLAN forwarding mode: all (trunk) | native (access) |
	// customize (selective trunk, use with excludedNetworkIds) | disabled.
	// Defaults to "native".
	Forward *string `pulumi:"forward,optional"`
	// NativeNetworkId is the network `_id` used as the native (untagged) VLAN.
	NativeNetworkId *string `pulumi:"nativeNetworkId,optional"`
	// TaggedVlanMgmt controls tagged VLAN behavior: auto | block_all | custom.
	TaggedVlanMgmt *string `pulumi:"taggedVlanMgmt,optional"`
	// ExcludedNetworkIds lists network `_id`s to exclude when forward is "customize".
	ExcludedNetworkIds []string `pulumi:"excludedNetworkIds,optional"`
	// MulticastRouterNetworkIds lists network `_id`s acting as multicast routers.
	MulticastRouterNetworkIds []string `pulumi:"multicastRouterNetworkIds,optional"`
	// VoiceNetworkId is the network `_id` used for VoIP (voice VLAN) traffic.
	VoiceNetworkId *string `pulumi:"voiceNetworkId,optional"`
}

func (v *PortProfileVlan) Annotate(a infer.Annotator) {
	a.Describe(&v.Forward, "Forward sets the VLAN forwarding mode: all (trunk) | native (access) | "+
		"customize (selective trunk, use with excludedNetworkIds) | disabled. Defaults to \"native\".")
	a.Describe(&v.NativeNetworkId, "NativeNetworkId is the network `_id` used as the native (untagged) VLAN.")
	a.Describe(&v.TaggedVlanMgmt, "TaggedVlanMgmt controls tagged VLAN behavior: auto | block_all | custom.")
	a.Describe(&v.ExcludedNetworkIds, "ExcludedNetworkIds lists network `_id`s to exclude when forward is \"customize\".")
	a.Describe(&v.MulticastRouterNetworkIds, "MulticastRouterNetworkIds lists network `_id`s acting as multicast routers.")
	a.Describe(&v.VoiceNetworkId, "VoiceNetworkId is the network `_id` used for VoIP (voice VLAN) traffic.")
}

// PortProfileLink groups the physical-layer link settings (speed/duplex/FEC).
type PortProfileLink struct {
	// Autoneg enables auto-negotiation of speed/duplex. Defaults to true. When
	// true it overrides the manual speed and fullDuplex settings.
	Autoneg *bool `pulumi:"autoneg,optional"`
	// Speed is the fixed port speed in Mbps when autoneg is false: 10 | 100 |
	// 1000 | 2500 | 5000 | 10000 | 20000 | 25000 | 40000 | 50000 | 100000.
	Speed *int `pulumi:"speed,optional"`
	// FullDuplex enables full-duplex when autoneg is false. Defaults to false.
	FullDuplex *bool `pulumi:"fullDuplex,optional"`
	// FecMode is the forward error correction mode: rs-fec | fc-fec | default | disabled.
	FecMode *string `pulumi:"fecMode,optional"`
}

func (l *PortProfileLink) Annotate(a infer.Annotator) {
	a.Describe(&l.Autoneg, "Autoneg enables auto-negotiation of speed/duplex. Defaults to true. When "+
		"true it overrides the manual speed and fullDuplex settings.")
	a.Describe(&l.Speed, "Speed is the fixed port speed in Mbps when autoneg is false: 10 | 100 | "+
		"1000 | 2500 | 5000 | 10000 | 20000 | 25000 | 40000 | 50000 | 100000.")
	a.Describe(&l.FullDuplex, "FullDuplex enables full-duplex when autoneg is false. Defaults to false.")
	a.Describe(&l.FecMode, "FecMode is the forward error correction mode: rs-fec | fc-fec | default | disabled.")
}

// PortProfileStormControl groups the broadcast/multicast/unknown-unicast
// storm-control settings.
type PortProfileStormControl struct {
	// Type selects the storm-control metric: level | rate.
	Type *string `pulumi:"type,optional"`
	// BroadcastEnabled enables broadcast storm control. Defaults to false.
	BroadcastEnabled *bool `pulumi:"broadcastEnabled,optional"`
	// BroadcastLevel is the broadcast storm-control level (0-100).
	BroadcastLevel *int `pulumi:"broadcastLevel,optional"`
	// BroadcastRate is the broadcast rate in pps (0-14880000).
	BroadcastRate *int `pulumi:"broadcastRate,optional"`
	// MulticastEnabled enables multicast storm control. Defaults to false.
	MulticastEnabled *bool `pulumi:"multicastEnabled,optional"`
	// MulticastLevel is the multicast storm-control level (0-100).
	MulticastLevel *int `pulumi:"multicastLevel,optional"`
	// MulticastRate is the multicast rate in pps (0-14880000).
	MulticastRate *int `pulumi:"multicastRate,optional"`
	// UnknownUnicastEnabled enables unknown-unicast storm control. Defaults to false.
	UnknownUnicastEnabled *bool `pulumi:"unknownUnicastEnabled,optional"`
	// UnknownUnicastLevel is the unknown-unicast storm-control level (0-100).
	UnknownUnicastLevel *int `pulumi:"unknownUnicastLevel,optional"`
	// UnknownUnicastRate is the unknown-unicast rate in pps (0-14880000).
	UnknownUnicastRate *int `pulumi:"unknownUnicastRate,optional"`
}

func (s *PortProfileStormControl) Annotate(a infer.Annotator) {
	a.Describe(&s.Type, "Type selects the storm-control metric: level | rate.")
	a.Describe(&s.BroadcastEnabled, "BroadcastEnabled enables broadcast storm control. Defaults to false.")
	a.Describe(&s.BroadcastLevel, "BroadcastLevel is the broadcast storm-control level (0-100).")
	a.Describe(&s.BroadcastRate, "BroadcastRate is the broadcast rate in pps (0-14880000).")
	a.Describe(&s.MulticastEnabled, "MulticastEnabled enables multicast storm control. Defaults to false.")
	a.Describe(&s.MulticastLevel, "MulticastLevel is the multicast storm-control level (0-100).")
	a.Describe(&s.MulticastRate, "MulticastRate is the multicast rate in pps (0-14880000).")
	a.Describe(&s.UnknownUnicastEnabled, "UnknownUnicastEnabled enables unknown-unicast storm control. Defaults to false.")
	a.Describe(&s.UnknownUnicastLevel, "UnknownUnicastLevel is the unknown-unicast storm-control level (0-100).")
	a.Describe(&s.UnknownUnicastRate, "UnknownUnicastRate is the unknown-unicast rate in pps (0-14880000).")
}

// PortProfilePortSecurity groups the MAC-based port-security settings.
type PortProfilePortSecurity struct {
	// Enabled enables MAC-based port security. Defaults to false.
	Enabled *bool `pulumi:"enabled,optional"`
	// MacAddresses lists allowed MAC addresses when port security is on.
	MacAddresses []string `pulumi:"macAddresses,optional"`
}

func (ps *PortProfilePortSecurity) Annotate(a infer.Annotator) {
	a.Describe(&ps.Enabled, "Enabled enables MAC-based port security. Defaults to false.")
	a.Describe(&ps.MacAddresses, "MacAddresses lists allowed MAC addresses when port security is on.")
}

// PortProfileDot1x groups the 802.1X PNAC settings.
type PortProfileDot1x struct {
	// Ctrl is the 802.1X PNAC mode: auto | force_authorized |
	// force_unauthorized | mac_based | multi_host. Defaults to "force_authorized".
	Ctrl *string `pulumi:"ctrl,optional"`
	// IdleTimeout is the MAC-based 802.1X idle timeout in seconds (0-65535).
	// Defaults to 300.
	IdleTimeout *int `pulumi:"idleTimeout,optional"`
}

func (d *PortProfileDot1x) Annotate(a infer.Annotator) {
	a.Describe(&d.Ctrl, "Ctrl is the 802.1X PNAC mode: auto | force_authorized | "+
		"force_unauthorized | mac_based | multi_host. Defaults to \"force_authorized\".")
	a.Describe(&d.IdleTimeout, "IdleTimeout is the MAC-based 802.1X idle timeout in seconds (0-65535). "+
		"Defaults to 300.")
}

// PortProfileLldpMed groups the LLDP-MED protocol toggles.
type PortProfileLldpMed struct {
	// Enabled enables LLDP-MED. Defaults to true.
	Enabled *bool `pulumi:"enabled,optional"`
	// NotifyEnabled enables LLDP-MED topology-change notifications. Defaults to false.
	NotifyEnabled *bool `pulumi:"notifyEnabled,optional"`
}

func (lm *PortProfileLldpMed) Annotate(a infer.Annotator) {
	a.Describe(&lm.Enabled, "Enabled enables LLDP-MED. Defaults to true.")
	a.Describe(&lm.NotifyEnabled, "NotifyEnabled enables LLDP-MED topology-change notifications. Defaults to false.")
}

// PortProfileEgressRateLimit groups the outbound rate-limiting settings.
type PortProfileEgressRateLimit struct {
	// Kbps is the outbound rate limit in kbps (64-9999999). Only applied when
	// enabled is true.
	Kbps *int `pulumi:"kbps,optional"`
	// Enabled enables outbound rate limiting. Defaults to false.
	Enabled *bool `pulumi:"enabled,optional"`
}

func (e *PortProfileEgressRateLimit) Annotate(a infer.Annotator) {
	a.Describe(&e.Kbps, "Kbps is the outbound rate limit in kbps (64-9999999). Only applied when "+
		"enabled is true.")
	a.Describe(&e.Enabled, "Enabled enables outbound rate limiting. Defaults to false.")
}

// PortProfilePriorityQueues groups the four QoS priority-queue levels.
type PortProfilePriorityQueues struct {
	// Queue1Level is the QoS priority queue 1 level (0-100).
	Queue1Level *int `pulumi:"queue1Level,optional"`
	// Queue2Level is the QoS priority queue 2 level (0-100).
	Queue2Level *int `pulumi:"queue2Level,optional"`
	// Queue3Level is the QoS priority queue 3 level (0-100).
	Queue3Level *int `pulumi:"queue3Level,optional"`
	// Queue4Level is the QoS priority queue 4 level (0-100).
	Queue4Level *int `pulumi:"queue4Level,optional"`
}

func (pq *PortProfilePriorityQueues) Annotate(a infer.Annotator) {
	a.Describe(&pq.Queue1Level, "Queue1Level is the QoS priority queue 1 level (0-100).")
	a.Describe(&pq.Queue2Level, "Queue2Level is the QoS priority queue 2 level (0-100).")
	a.Describe(&pq.Queue3Level, "Queue3Level is the QoS priority queue 3 level (0-100).")
	a.Describe(&pq.Queue4Level, "Queue4Level is the QoS priority queue 4 level (0-100).")
}

// PortProfileArgs are the user-supplied inputs for a port profile.
type PortProfileArgs struct {
	// Name is a descriptive name for the port profile.
	Name string `pulumi:"name"`

	// OpMode is the operation mode. Only "switch" is supported. Defaults to "switch".
	OpMode *string `pulumi:"opMode,optional"`
	// Isolation enables port isolation so devices on this profile cannot
	// communicate with each other. Defaults to false.
	Isolation *bool `pulumi:"isolation,optional"`
	// PoeMode controls Power-over-Ethernet: auto | off (the values go-unifi
	// accepts for a port profile). Per-port passthrough/pasv24 modes are set via
	// a Device port override's poeMode, not here.
	PoeMode *string `pulumi:"poeMode,optional"`
	// StpPortMode enables Spanning Tree Protocol on the port. Defaults to true.
	StpPortMode *bool `pulumi:"stpPortMode,optional"`
	// PortKeepaliveEnabled enables port keepalive. Defaults to false.
	PortKeepaliveEnabled *bool `pulumi:"portKeepaliveEnabled,optional"`
	// SettingPreference controls config source: auto | manual.
	SettingPreference *string `pulumi:"settingPreference,optional"`

	// Vlan groups the VLAN / forwarding settings.
	Vlan *PortProfileVlan `pulumi:"vlan,optional"`
	// Link groups the physical-layer link settings (speed/duplex/FEC).
	Link *PortProfileLink `pulumi:"link,optional"`
	// StormControl groups the storm-control settings.
	StormControl *PortProfileStormControl `pulumi:"stormControl,optional"`
	// PortSecurity groups the MAC-based port-security settings.
	PortSecurity *PortProfilePortSecurity `pulumi:"portSecurity,optional"`
	// Dot1x groups the 802.1X PNAC settings.
	Dot1x *PortProfileDot1x `pulumi:"dot1x,optional"`
	// LldpMed groups the LLDP-MED protocol toggles.
	LldpMed *PortProfileLldpMed `pulumi:"lldpMed,optional"`
	// EgressRateLimit groups the outbound rate-limiting settings.
	EgressRateLimit *PortProfileEgressRateLimit `pulumi:"egressRateLimit,optional"`
	// PriorityQueues groups the QoS priority-queue levels.
	PriorityQueues *PortProfilePriorityQueues `pulumi:"priorityQueues,optional"`
}

func (args *PortProfileArgs) Annotate(a infer.Annotator) {
	a.Describe(&args.Name, "Name is a descriptive name for the port profile.")
	a.Describe(&args.OpMode, "OpMode is the operation mode. Only \"switch\" is supported. Defaults to \"switch\".")
	a.Describe(&args.Isolation, "Isolation enables port isolation so devices on this profile cannot "+
		"communicate with each other. Defaults to false.")
	a.Describe(&args.PoeMode, "PoeMode controls Power-over-Ethernet: auto | off (the values go-unifi "+
		"accepts for a port profile). Per-port passthrough/pasv24 modes are set via "+
		"a Device port override's poeMode, not here.")
	a.Describe(&args.StpPortMode, "StpPortMode enables Spanning Tree Protocol on the port. Defaults to true.")
	a.Describe(&args.PortKeepaliveEnabled, "PortKeepaliveEnabled enables port keepalive. Defaults to false.")
	a.Describe(&args.SettingPreference, "SettingPreference controls config source: auto | manual.")
	a.Describe(&args.Vlan, "Vlan groups the VLAN / forwarding settings.")
	a.Describe(&args.Link, "Link groups the physical-layer link settings (speed/duplex/FEC).")
	a.Describe(&args.StormControl, "StormControl groups the storm-control settings.")
	a.Describe(&args.PortSecurity, "PortSecurity groups the MAC-based port-security settings.")
	a.Describe(&args.Dot1x, "Dot1x groups the 802.1X PNAC settings.")
	a.Describe(&args.LldpMed, "LldpMed groups the LLDP-MED protocol toggles.")
	a.Describe(&args.EgressRateLimit, "EgressRateLimit groups the outbound rate-limiting settings.")
	a.Describe(&args.PriorityQueues, "PriorityQueues groups the QoS priority-queue levels.")
}

// PortProfileState is the persisted state: inputs plus controller-assigned fields.
type PortProfileState struct {
	PortProfileArgs
	// PortProfileId is the controller-assigned identifier (the UniFi `_id`),
	// referenced by a Device port override's portProfileId (portconf_id).
	PortProfileId string `pulumi:"portProfileId"`
}

func (st *PortProfileState) Annotate(a infer.Annotator) {
	a.Describe(&st.PortProfileId, "PortProfileId is the controller-assigned identifier (the UniFi `_id`), "+
		"referenced by a Device port override's portProfileId (portconf_id).")
}

// Annotate documents the resource. Must use a pointer receiver so the
// annotator can take the address of the resource and its fields.
func (p *PortProfile) Annotate(a infer.Annotator) {
	a.Describe(&p, "A UniFi switch port profile (portconf). A reusable collection of port settings "+
		"grouped into nested objects (vlan, link, stormControl, portSecurity, dot1x, lldpMed, "+
		"egressRateLimit, priorityQueues) that can be applied to switch ports via a Device port "+
		"override's portProfileId.")
}

// toUnifi builds a go-unifi PortProfile from inputs. id is empty on create.
//
// Several controller fields have no upstream omitempty and carry non-zero
// controller defaults, so they are always sent even when the user omits the
// enclosing nested group; omitting them (Go zero values) would silently change
// behavior. The default is read from the group when both the group and the
// member are set, otherwise the documented default is used.
func (a PortProfileArgs) toUnifi(id string) *unifi.PortProfile {
	p := &unifi.PortProfile{
		ID:                   id,
		Name:                 a.Name,
		Isolation:            derefOr(a.Isolation, false),
		PortKeepaliveEnabled: derefOr(a.PortKeepaliveEnabled, false),
		StpPortMode:          derefOr(a.StpPortMode, true),
		OpMode:               derefOr(a.OpMode, "switch"),
	}
	if a.PoeMode != nil {
		p.PoeMode = *a.PoeMode
	}
	if a.SettingPreference != nil {
		p.SettingPreference = *a.SettingPreference
	}

	// VLAN / forwarding. forward defaults to "native" and is always sent.
	p.Forward = "native"
	if v := a.Vlan; v != nil {
		if v.Forward != nil {
			p.Forward = *v.Forward
		}
		if v.NativeNetworkId != nil {
			p.NATiveNetworkID = *v.NativeNetworkId
		}
		if v.TaggedVlanMgmt != nil {
			p.TaggedVLANMgmt = *v.TaggedVlanMgmt
		}
		if v.ExcludedNetworkIds != nil {
			p.ExcludedNetworkIDs = v.ExcludedNetworkIds
		}
		if v.MulticastRouterNetworkIds != nil {
			p.MulticastRouterNetworkIDs = v.MulticastRouterNetworkIds
		}
		if v.VoiceNetworkId != nil {
			p.VoiceNetworkID = *v.VoiceNetworkId
		}
	}

	// Link. autoneg (true) and fullDuplex (false) are always sent.
	p.Autoneg = true
	p.FullDuplex = false
	if l := a.Link; l != nil {
		if l.Autoneg != nil {
			p.Autoneg = *l.Autoneg
		}
		if l.FullDuplex != nil {
			p.FullDuplex = *l.FullDuplex
		}
		if l.Speed != nil {
			p.Speed = *l.Speed
		}
		if l.FecMode != nil {
			p.FecMode = *l.FecMode
		}
	}

	// Storm control. The three *Enabled bools are always sent (default false).
	p.StormctrlBroadcastastEnabled = false
	p.StormctrlMcastEnabled = false
	p.StormctrlUcastEnabled = false
	if s := a.StormControl; s != nil {
		if s.BroadcastEnabled != nil {
			p.StormctrlBroadcastastEnabled = *s.BroadcastEnabled
		}
		if s.MulticastEnabled != nil {
			p.StormctrlMcastEnabled = *s.MulticastEnabled
		}
		if s.UnknownUnicastEnabled != nil {
			p.StormctrlUcastEnabled = *s.UnknownUnicastEnabled
		}
		if s.Type != nil {
			p.StormctrlType = *s.Type
		}
		if s.BroadcastLevel != nil {
			p.StormctrlBroadcastastLevel = *s.BroadcastLevel
		}
		if s.BroadcastRate != nil {
			p.StormctrlBroadcastastRate = *s.BroadcastRate
		}
		if s.MulticastLevel != nil {
			p.StormctrlMcastLevel = *s.MulticastLevel
		}
		if s.MulticastRate != nil {
			p.StormctrlMcastRate = *s.MulticastRate
		}
		if s.UnknownUnicastLevel != nil {
			p.StormctrlUcastLevel = *s.UnknownUnicastLevel
		}
		if s.UnknownUnicastRate != nil {
			p.StormctrlUcastRate = *s.UnknownUnicastRate
		}
	}

	// Port security. enabled (false) is always sent.
	p.PortSecurityEnabled = false
	if ps := a.PortSecurity; ps != nil {
		if ps.Enabled != nil {
			p.PortSecurityEnabled = *ps.Enabled
		}
		if ps.MacAddresses != nil {
			p.PortSecurityMACAddress = ps.MacAddresses
		}
	}

	// 802.1X. ctrl ("force_authorized") and idleTimeout (300) are always sent.
	p.Dot1XCtrl = "force_authorized"
	p.Dot1XIDleTimeout = 300
	if d := a.Dot1x; d != nil {
		if d.Ctrl != nil {
			p.Dot1XCtrl = *d.Ctrl
		}
		if d.IdleTimeout != nil {
			p.Dot1XIDleTimeout = *d.IdleTimeout
		}
	}

	// LLDP-MED. enabled (true) and notifyEnabled (false) are always sent.
	p.LldpmedEnabled = true
	p.LldpmedNotifyEnabled = false
	if lm := a.LldpMed; lm != nil {
		if lm.Enabled != nil {
			p.LldpmedEnabled = *lm.Enabled
		}
		if lm.NotifyEnabled != nil {
			p.LldpmedNotifyEnabled = *lm.NotifyEnabled
		}
	}

	// Egress rate limiting. enabled (false) is always sent. Setting kbps does
	// NOT auto-enable the limit; the two are independent.
	p.EgressRateLimitKbpsEnabled = false
	if e := a.EgressRateLimit; e != nil {
		if e.Enabled != nil {
			p.EgressRateLimitKbpsEnabled = *e.Enabled
		}
		if e.Kbps != nil {
			p.EgressRateLimitKbps = *e.Kbps
		}
	}

	// Priority queues.
	if pq := a.PriorityQueues; pq != nil {
		if pq.Queue1Level != nil {
			p.PriorityQueue1Level = *pq.Queue1Level
		}
		if pq.Queue2Level != nil {
			p.PriorityQueue2Level = *pq.Queue2Level
		}
		if pq.Queue3Level != nil {
			p.PriorityQueue3Level = *pq.Queue3Level
		}
		if pq.Queue4Level != nil {
			p.PriorityQueue4Level = *pq.Queue4Level
		}
	}

	return p
}

// isZero reports whether no vlan member is set (so the group round-trips as nil).
func (g PortProfileVlan) isZero() bool {
	return g.Forward == nil && g.NativeNetworkId == nil && g.TaggedVlanMgmt == nil &&
		g.ExcludedNetworkIds == nil && g.MulticastRouterNetworkIds == nil && g.VoiceNetworkId == nil
}

// isZero reports whether no portSecurity member is set (so it round-trips as nil).
func (g PortProfilePortSecurity) isZero() bool {
	return g.Enabled == nil && g.MacAddresses == nil
}

// portProfileVlanFrom reconstructs the vlan group, returning nil when no member is set.
func portProfileVlanFrom(u *unifi.PortProfile, prior *PortProfileVlan) *PortProfileVlan {
	var p PortProfileVlan
	if prior != nil {
		p = *prior
	}
	g := PortProfileVlan{
		NativeNetworkId: vlanStrPtr(u.NATiveNetworkID, p.NativeNetworkId),
		TaggedVlanMgmt:  vlanStrPtr(u.TaggedVLANMgmt, p.TaggedVlanMgmt),
		VoiceNetworkId:  vlanStrPtr(u.VoiceNetworkID, p.VoiceNetworkId),
	}
	// forward defaults to "native" on the controller; only reflect it when the
	// user previously set it so the group can round-trip as nil.
	if p.Forward != nil {
		g.Forward = vlanStrPtr(u.Forward, p.Forward)
	}
	if len(u.ExcludedNetworkIDs) > 0 {
		g.ExcludedNetworkIds = u.ExcludedNetworkIDs
	} else {
		g.ExcludedNetworkIds = p.ExcludedNetworkIds
	}
	if len(u.MulticastRouterNetworkIDs) > 0 {
		g.MulticastRouterNetworkIds = u.MulticastRouterNetworkIDs
	} else {
		g.MulticastRouterNetworkIds = p.MulticastRouterNetworkIds
	}
	if g.isZero() {
		return nil
	}
	return &g
}

// portProfileLinkFrom reconstructs the link group, returning nil when no member is set.
func portProfileLinkFrom(u *unifi.PortProfile, prior *PortProfileLink) *PortProfileLink {
	var p PortProfileLink
	if prior != nil {
		p = *prior
	}
	g := PortProfileLink{
		Speed:      vlanIntPtr(u.Speed, p.Speed),
		FullDuplex: vlanBoolPtr(u.FullDuplex, p.FullDuplex),
		FecMode:    vlanStrPtr(u.FecMode, p.FecMode),
	}
	// autoneg defaults to true on the controller; only reflect it when the user
	// previously set it so the group can round-trip as nil.
	if p.Autoneg != nil {
		g.Autoneg = ptr(u.Autoneg)
	}
	if g == (PortProfileLink{}) {
		return nil
	}
	return &g
}

// portProfileStormControlFrom reconstructs the stormControl group, returning nil
// when no member is set.
func portProfileStormControlFrom(u *unifi.PortProfile, prior *PortProfileStormControl) *PortProfileStormControl {
	var p PortProfileStormControl
	if prior != nil {
		p = *prior
	}
	g := PortProfileStormControl{
		Type:                  vlanStrPtr(u.StormctrlType, p.Type),
		BroadcastEnabled:      vlanBoolPtr(u.StormctrlBroadcastastEnabled, p.BroadcastEnabled),
		BroadcastLevel:        vlanIntPtr(u.StormctrlBroadcastastLevel, p.BroadcastLevel),
		BroadcastRate:         vlanIntPtr(u.StormctrlBroadcastastRate, p.BroadcastRate),
		MulticastEnabled:      vlanBoolPtr(u.StormctrlMcastEnabled, p.MulticastEnabled),
		MulticastLevel:        vlanIntPtr(u.StormctrlMcastLevel, p.MulticastLevel),
		MulticastRate:         vlanIntPtr(u.StormctrlMcastRate, p.MulticastRate),
		UnknownUnicastEnabled: vlanBoolPtr(u.StormctrlUcastEnabled, p.UnknownUnicastEnabled),
		UnknownUnicastLevel:   vlanIntPtr(u.StormctrlUcastLevel, p.UnknownUnicastLevel),
		UnknownUnicastRate:    vlanIntPtr(u.StormctrlUcastRate, p.UnknownUnicastRate),
	}
	if g == (PortProfileStormControl{}) {
		return nil
	}
	return &g
}

// portProfilePortSecurityFrom reconstructs the portSecurity group, returning nil
// when no member is set.
func portProfilePortSecurityFrom(u *unifi.PortProfile, prior *PortProfilePortSecurity) *PortProfilePortSecurity {
	var p PortProfilePortSecurity
	if prior != nil {
		p = *prior
	}
	g := PortProfilePortSecurity{
		Enabled: vlanBoolPtr(u.PortSecurityEnabled, p.Enabled),
	}
	if len(u.PortSecurityMACAddress) > 0 {
		g.MacAddresses = u.PortSecurityMACAddress
	} else {
		g.MacAddresses = p.MacAddresses
	}
	if g.isZero() {
		return nil
	}
	return &g
}

// portProfileDot1xFrom reconstructs the dot1x group, returning nil when no member is set.
func portProfileDot1xFrom(u *unifi.PortProfile, prior *PortProfileDot1x) *PortProfileDot1x {
	var p PortProfileDot1x
	if prior != nil {
		p = *prior
	}
	g := PortProfileDot1x{}
	// ctrl ("force_authorized") and idleTimeout (300) default non-zero on the
	// controller; only reflect them when the user previously set them so the
	// group can round-trip as nil.
	if p.Ctrl != nil {
		g.Ctrl = vlanStrPtr(u.Dot1XCtrl, p.Ctrl)
	}
	if p.IdleTimeout != nil {
		g.IdleTimeout = vlanIntPtr(u.Dot1XIDleTimeout, p.IdleTimeout)
	}
	if g == (PortProfileDot1x{}) {
		return nil
	}
	return &g
}

// portProfileLldpMedFrom reconstructs the lldpMed group, returning nil when no member is set.
func portProfileLldpMedFrom(u *unifi.PortProfile, prior *PortProfileLldpMed) *PortProfileLldpMed {
	var p PortProfileLldpMed
	if prior != nil {
		p = *prior
	}
	g := PortProfileLldpMed{
		NotifyEnabled: vlanBoolPtr(u.LldpmedNotifyEnabled, p.NotifyEnabled),
	}
	// enabled defaults to true on the controller; only reflect it when the user
	// previously set it so the group can round-trip as nil.
	if p.Enabled != nil {
		g.Enabled = ptr(u.LldpmedEnabled)
	}
	if g == (PortProfileLldpMed{}) {
		return nil
	}
	return &g
}

// portProfileEgressRateLimitFrom reconstructs the egressRateLimit group,
// returning nil when no member is set.
func portProfileEgressRateLimitFrom(u *unifi.PortProfile, prior *PortProfileEgressRateLimit) *PortProfileEgressRateLimit {
	var p PortProfileEgressRateLimit
	if prior != nil {
		p = *prior
	}
	g := PortProfileEgressRateLimit{
		Kbps:    vlanIntPtr(u.EgressRateLimitKbps, p.Kbps),
		Enabled: vlanBoolPtr(u.EgressRateLimitKbpsEnabled, p.Enabled),
	}
	if g == (PortProfileEgressRateLimit{}) {
		return nil
	}
	return &g
}

// portProfilePriorityQueuesFrom reconstructs the priorityQueues group, returning
// nil when no member is set.
func portProfilePriorityQueuesFrom(u *unifi.PortProfile, prior *PortProfilePriorityQueues) *PortProfilePriorityQueues {
	var p PortProfilePriorityQueues
	if prior != nil {
		p = *prior
	}
	g := PortProfilePriorityQueues{
		Queue1Level: vlanIntPtr(u.PriorityQueue1Level, p.Queue1Level),
		Queue2Level: vlanIntPtr(u.PriorityQueue2Level, p.Queue2Level),
		Queue3Level: vlanIntPtr(u.PriorityQueue3Level, p.Queue3Level),
		Queue4Level: vlanIntPtr(u.PriorityQueue4Level, p.Queue4Level),
	}
	if g == (PortProfilePriorityQueues{}) {
		return nil
	}
	return &g
}

// portProfileStateFrom maps a controller PortProfile back into resource state.
// prior carries the user inputs so unset optional fields are preserved across
// the round-trip and the optional nested groups round-trip as nil when unused.
func portProfileStateFrom(u *unifi.PortProfile, prior PortProfileArgs) PortProfileState {
	args := PortProfileArgs{
		Name:                 u.Name,
		Isolation:            ptr(u.Isolation),
		PortKeepaliveEnabled: ptr(u.PortKeepaliveEnabled),
		StpPortMode:          ptr(u.StpPortMode),
	}
	if u.OpMode != "" {
		args.OpMode = ptr(u.OpMode)
	}
	if u.PoeMode != "" {
		args.PoeMode = ptr(u.PoeMode)
	}
	if u.SettingPreference != "" {
		args.SettingPreference = ptr(u.SettingPreference)
	}

	// Nested facets.
	args.Vlan = portProfileVlanFrom(u, prior.Vlan)
	args.Link = portProfileLinkFrom(u, prior.Link)
	args.StormControl = portProfileStormControlFrom(u, prior.StormControl)
	args.PortSecurity = portProfilePortSecurityFrom(u, prior.PortSecurity)
	args.Dot1x = portProfileDot1xFrom(u, prior.Dot1x)
	args.LldpMed = portProfileLldpMedFrom(u, prior.LldpMed)
	args.EgressRateLimit = portProfileEgressRateLimitFrom(u, prior.EgressRateLimit)
	args.PriorityQueues = portProfilePriorityQueuesFrom(u, prior.PriorityQueues)

	return PortProfileState{PortProfileArgs: args, PortProfileId: u.ID}
}

// Create provisions a new port profile.
func (PortProfile) Create(ctx context.Context, req infer.CreateRequest[PortProfileArgs]) (infer.CreateResponse[PortProfileState], error) {
	if req.DryRun {
		return infer.CreateResponse[PortProfileState]{Output: PortProfileState{PortProfileArgs: req.Inputs}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	created, err := cfg.Network().CreatePortProfile(ctx, cfg.ResolvedSite(), req.Inputs.toUnifi(""))
	if err != nil {
		return infer.CreateResponse[PortProfileState]{}, wrap(fmt.Sprintf("create port profile %q (site %q)", req.Inputs.Name, cfg.ResolvedSite()), err)
	}
	return infer.CreateResponse[PortProfileState]{ID: created.ID, Output: portProfileStateFrom(created, req.Inputs)}, nil
}

// Read recovers state from the controller, enabling `pulumi import`.
func (PortProfile) Read(ctx context.Context, req infer.ReadRequest[PortProfileArgs, PortProfileState]) (infer.ReadResponse[PortProfileArgs, PortProfileState], error) {
	cfg := infer.GetConfig[config.Config](ctx)
	u, err := cfg.Network().GetPortProfile(ctx, cfg.ResolvedSite(), req.ID)
	if notFound(err) {
		return infer.ReadResponse[PortProfileArgs, PortProfileState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[PortProfileArgs, PortProfileState]{}, wrap(fmt.Sprintf("read port profile %q (site %q)", req.ID, cfg.ResolvedSite()), err)
	}
	st := portProfileStateFrom(u, req.Inputs)
	return infer.ReadResponse[PortProfileArgs, PortProfileState]{ID: req.ID, Inputs: st.PortProfileArgs, State: st}, nil
}

// Update applies changed inputs in place.
func (PortProfile) Update(ctx context.Context, req infer.UpdateRequest[PortProfileArgs, PortProfileState]) (infer.UpdateResponse[PortProfileState], error) {
	if req.DryRun {
		return infer.UpdateResponse[PortProfileState]{Output: PortProfileState{PortProfileArgs: req.Inputs, PortProfileId: req.ID}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	updated, err := cfg.Network().UpdatePortProfile(ctx, cfg.ResolvedSite(), req.Inputs.toUnifi(req.ID))
	if err != nil {
		return infer.UpdateResponse[PortProfileState]{}, wrap(fmt.Sprintf("update port profile %q (site %q)", req.ID, cfg.ResolvedSite()), err)
	}
	return infer.UpdateResponse[PortProfileState]{Output: portProfileStateFrom(updated, req.Inputs)}, nil
}

// Delete removes the port profile.
func (PortProfile) Delete(ctx context.Context, req infer.DeleteRequest[PortProfileState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[config.Config](ctx)
	err := cfg.Network().DeletePortProfile(ctx, cfg.ResolvedSite(), req.ID)
	if notFound(err) {
		return infer.DeleteResponse{}, nil
	}
	return infer.DeleteResponse{}, wrap(fmt.Sprintf("delete port profile %q (site %q)", req.ID, cfg.ResolvedSite()), err)
}
