package network

import (
	"testing"
)

// TestPortProfileToUnifiDefaults asserts the grouped always-send defaults are
// emitted on the controller object even with a minimal (no-groups) input.
func TestPortProfileToUnifiDefaults(t *testing.T) {
	args := PortProfileArgs{Name: "minimal"}
	u := args.toUnifi("")

	if u.Name != "minimal" {
		t.Errorf("name: got %q", u.Name)
	}
	// vlan.forward
	if u.Forward != "native" {
		t.Errorf("forward should default to native, got %q", u.Forward)
	}
	// link.autoneg / link.fullDuplex
	if !u.Autoneg {
		t.Error("autoneg should default to true")
	}
	if u.FullDuplex {
		t.Error("fullDuplex should default to false")
	}
	// stormControl.*Enabled
	if u.StormctrlBroadcastastEnabled || u.StormctrlMcastEnabled || u.StormctrlUcastEnabled {
		t.Error("storm-control enabled bools should default to false")
	}
	// portSecurity.enabled
	if u.PortSecurityEnabled {
		t.Error("portSecurityEnabled should default to false")
	}
	// dot1x.ctrl / dot1x.idleTimeout
	if u.Dot1XCtrl != "force_authorized" {
		t.Errorf("dot1xCtrl should default to force_authorized, got %q", u.Dot1XCtrl)
	}
	if u.Dot1XIDleTimeout != 300 {
		t.Errorf("dot1xIdleTimeout should default to 300, got %d", u.Dot1XIDleTimeout)
	}
	// lldpMed.enabled / lldpMed.notifyEnabled
	if !u.LldpmedEnabled {
		t.Error("lldpmedEnabled should default to true")
	}
	if u.LldpmedNotifyEnabled {
		t.Error("lldpmedNotifyEnabled should default to false")
	}
	// egressRateLimit.enabled
	if u.EgressRateLimitKbpsEnabled {
		t.Error("egressRateLimitKbpsEnabled should default to false")
	}
	// top-level always-send fields keep their flat handling.
	if u.OpMode != "switch" {
		t.Errorf("opMode should default to switch, got %q", u.OpMode)
	}
	if !u.StpPortMode {
		t.Error("stpPortMode should default to true")
	}
	if u.Isolation || u.PortKeepaliveEnabled {
		t.Error("isolation and portKeepaliveEnabled should default to false")
	}
}

// TestPortProfileRoundTrip exercises every nested group and asserts the fields
// survive the toUnifi -> stateFrom round-trip.
func TestPortProfileRoundTrip(t *testing.T) {
	args := PortProfileArgs{
		Name:                 "trunk",
		OpMode:               ptr("switch"),
		Isolation:            ptr(true),
		PoeMode:              ptr("off"),
		StpPortMode:          ptr(false),
		PortKeepaliveEnabled: ptr(true),
		SettingPreference:    ptr("manual"),

		Vlan: &PortProfileVlan{
			Forward:                   ptr("customize"),
			NativeNetworkId:           ptr("net-native"),
			TaggedVlanMgmt:            ptr("custom"),
			ExcludedNetworkIds:        []string{"net-a", "net-b"},
			MulticastRouterNetworkIds: []string{"net-mcast"},
			VoiceNetworkId:            ptr("net-voice"),
		},
		Link: &PortProfileLink{
			Autoneg:    ptr(false),
			Speed:      ptr(1000),
			FullDuplex: ptr(true),
			FecMode:    ptr("rs-fec"),
		},
		StormControl: &PortProfileStormControl{
			Type:                  ptr("rate"),
			BroadcastEnabled:      ptr(true),
			BroadcastLevel:        ptr(50),
			BroadcastRate:         ptr(1000),
			MulticastEnabled:      ptr(true),
			MulticastLevel:        ptr(40),
			MulticastRate:         ptr(2000),
			UnknownUnicastEnabled: ptr(true),
			UnknownUnicastLevel:   ptr(30),
			UnknownUnicastRate:    ptr(3000),
		},
		PortSecurity: &PortProfilePortSecurity{
			Enabled:      ptr(true),
			MacAddresses: []string{"aa:bb:cc:dd:ee:ff"},
		},
		Dot1x: &PortProfileDot1x{
			Ctrl:        ptr("mac_based"),
			IdleTimeout: ptr(120),
		},
		LldpMed: &PortProfileLldpMed{
			Enabled:       ptr(false),
			NotifyEnabled: ptr(true),
		},
		EgressRateLimit: &PortProfileEgressRateLimit{
			Kbps:    ptr(5000),
			Enabled: ptr(true),
		},
		PriorityQueues: &PortProfilePriorityQueues{
			Queue1Level: ptr(10),
			Queue2Level: ptr(20),
			Queue3Level: ptr(30),
			Queue4Level: ptr(40),
		},
	}

	u := args.toUnifi("pp-1")
	if u.ID != "pp-1" {
		t.Fatalf("id not propagated: %q", u.ID)
	}

	st := portProfileStateFrom(u, args)
	if st.PortProfileId != "pp-1" {
		t.Errorf("computed portProfileId: got %q", st.PortProfileId)
	}
	got := st.PortProfileArgs

	// Top-level fields.
	if got.Name != "trunk" {
		t.Errorf("name: got %q", got.Name)
	}
	if derefOr(got.OpMode, "") != "switch" {
		t.Errorf("opMode: got %v", got.OpMode)
	}
	if derefOr(got.Isolation, false) != true {
		t.Errorf("isolation: got %v", got.Isolation)
	}
	if derefOr(got.PoeMode, "") != "off" {
		t.Errorf("poeMode: got %v", got.PoeMode)
	}
	if derefOr(got.StpPortMode, true) != false {
		t.Errorf("stpPortMode: got %v", got.StpPortMode)
	}
	if derefOr(got.PortKeepaliveEnabled, false) != true {
		t.Errorf("portKeepaliveEnabled: got %v", got.PortKeepaliveEnabled)
	}
	if derefOr(got.SettingPreference, "") != "manual" {
		t.Errorf("settingPreference: got %v", got.SettingPreference)
	}

	// vlan group.
	if got.Vlan == nil {
		t.Fatal("vlan group lost on round-trip")
	}
	if derefOr(got.Vlan.Forward, "") != "customize" {
		t.Errorf("vlan.forward: got %v", got.Vlan.Forward)
	}
	if derefOr(got.Vlan.NativeNetworkId, "") != "net-native" {
		t.Errorf("vlan.nativeNetworkId: got %v", got.Vlan.NativeNetworkId)
	}
	if derefOr(got.Vlan.TaggedVlanMgmt, "") != "custom" {
		t.Errorf("vlan.taggedVlanMgmt: got %v", got.Vlan.TaggedVlanMgmt)
	}
	if len(got.Vlan.ExcludedNetworkIds) != 2 || got.Vlan.ExcludedNetworkIds[0] != "net-a" {
		t.Errorf("vlan.excludedNetworkIds: got %v", got.Vlan.ExcludedNetworkIds)
	}
	if len(got.Vlan.MulticastRouterNetworkIds) != 1 || got.Vlan.MulticastRouterNetworkIds[0] != "net-mcast" {
		t.Errorf("vlan.multicastRouterNetworkIds: got %v", got.Vlan.MulticastRouterNetworkIds)
	}
	if derefOr(got.Vlan.VoiceNetworkId, "") != "net-voice" {
		t.Errorf("vlan.voiceNetworkId: got %v", got.Vlan.VoiceNetworkId)
	}

	// link group.
	if got.Link == nil {
		t.Fatal("link group lost on round-trip")
	}
	if derefOr(got.Link.Autoneg, true) != false {
		t.Errorf("link.autoneg: got %v", got.Link.Autoneg)
	}
	if derefOr(got.Link.Speed, 0) != 1000 {
		t.Errorf("link.speed: got %v", got.Link.Speed)
	}
	if derefOr(got.Link.FullDuplex, false) != true {
		t.Errorf("link.fullDuplex: got %v", got.Link.FullDuplex)
	}
	if derefOr(got.Link.FecMode, "") != "rs-fec" {
		t.Errorf("link.fecMode: got %v", got.Link.FecMode)
	}

	// stormControl group.
	if got.StormControl == nil {
		t.Fatal("stormControl group lost on round-trip")
	}
	if derefOr(got.StormControl.Type, "") != "rate" {
		t.Errorf("stormControl.type: got %v", got.StormControl.Type)
	}
	if derefOr(got.StormControl.BroadcastEnabled, false) != true ||
		derefOr(got.StormControl.BroadcastLevel, 0) != 50 ||
		derefOr(got.StormControl.BroadcastRate, 0) != 1000 {
		t.Errorf("stormControl broadcast round-trip failed")
	}
	if derefOr(got.StormControl.MulticastEnabled, false) != true ||
		derefOr(got.StormControl.MulticastLevel, 0) != 40 ||
		derefOr(got.StormControl.MulticastRate, 0) != 2000 {
		t.Errorf("stormControl multicast round-trip failed")
	}
	if derefOr(got.StormControl.UnknownUnicastEnabled, false) != true ||
		derefOr(got.StormControl.UnknownUnicastLevel, 0) != 30 ||
		derefOr(got.StormControl.UnknownUnicastRate, 0) != 3000 {
		t.Errorf("stormControl unknownUnicast round-trip failed")
	}

	// portSecurity group.
	if got.PortSecurity == nil {
		t.Fatal("portSecurity group lost on round-trip")
	}
	if derefOr(got.PortSecurity.Enabled, false) != true || len(got.PortSecurity.MacAddresses) != 1 {
		t.Errorf("portSecurity round-trip failed")
	}

	// dot1x group.
	if got.Dot1x == nil {
		t.Fatal("dot1x group lost on round-trip")
	}
	if derefOr(got.Dot1x.Ctrl, "") != "mac_based" || derefOr(got.Dot1x.IdleTimeout, 0) != 120 {
		t.Errorf("dot1x round-trip failed")
	}

	// lldpMed group.
	if got.LldpMed == nil {
		t.Fatal("lldpMed group lost on round-trip")
	}
	if derefOr(got.LldpMed.Enabled, true) != false || derefOr(got.LldpMed.NotifyEnabled, false) != true {
		t.Errorf("lldpMed round-trip failed")
	}

	// egressRateLimit group.
	if got.EgressRateLimit == nil {
		t.Fatal("egressRateLimit group lost on round-trip")
	}
	if derefOr(got.EgressRateLimit.Kbps, 0) != 5000 || derefOr(got.EgressRateLimit.Enabled, false) != true {
		t.Errorf("egressRateLimit round-trip failed")
	}

	// priorityQueues group.
	if got.PriorityQueues == nil {
		t.Fatal("priorityQueues group lost on round-trip")
	}
	if derefOr(got.PriorityQueues.Queue1Level, 0) != 10 ||
		derefOr(got.PriorityQueues.Queue2Level, 0) != 20 ||
		derefOr(got.PriorityQueues.Queue3Level, 0) != 30 ||
		derefOr(got.PriorityQueues.Queue4Level, 0) != 40 {
		t.Errorf("priorityQueues round-trip failed")
	}
}

// TestPortProfileEmptyGroupsStayNil asserts that round-tripping a minimal port
// profile does not synthesize non-nil facet groups out of controller defaults
// (notably the always-send defaults like forward/autoneg/dot1x/lldpMed).
func TestPortProfileEmptyGroupsStayNil(t *testing.T) {
	u := PortProfileArgs{Name: "minimal"}.toUnifi("")
	u.ID = "pp-min"
	st := portProfileStateFrom(u, PortProfileArgs{Name: "minimal"})
	got := st.PortProfileArgs

	if got.Vlan != nil {
		t.Errorf("vlan should be nil for a minimal profile, got %+v", got.Vlan)
	}
	if got.Link != nil {
		t.Errorf("link should be nil for a minimal profile, got %+v", got.Link)
	}
	if got.StormControl != nil {
		t.Errorf("stormControl should be nil for a minimal profile, got %+v", got.StormControl)
	}
	if got.PortSecurity != nil {
		t.Errorf("portSecurity should be nil for a minimal profile, got %+v", got.PortSecurity)
	}
	if got.Dot1x != nil {
		t.Errorf("dot1x should be nil for a minimal profile, got %+v", got.Dot1x)
	}
	if got.LldpMed != nil {
		t.Errorf("lldpMed should be nil for a minimal profile, got %+v", got.LldpMed)
	}
	if got.EgressRateLimit != nil {
		t.Errorf("egressRateLimit should be nil for a minimal profile, got %+v", got.EgressRateLimit)
	}
	if got.PriorityQueues != nil {
		t.Errorf("priorityQueues should be nil for a minimal profile, got %+v", got.PriorityQueues)
	}
}
