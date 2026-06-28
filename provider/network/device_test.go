package network

import (
	"testing"

	"github.com/filipowm/go-unifi/unifi"
)

// TestDeviceApplyToAndStateRoundTrip exercises the adoption read-modify-write:
// a representative spread of managed inputs is overlaid onto a fetched device
// that already carries unmanaged settings (and a pre-existing port override),
// then mapped back to state. It asserts the managed fields land, the unmanaged
// settings survive, and the read-only facts and round-tripped inputs are correct.
func TestDeviceApplyToAndStateRoundTrip(t *testing.T) {
	args := DeviceArgs{
		Mac:      "aa:bb:cc:dd:ee:ff",
		Name:     ptr("Core Switch"),
		Disabled: ptr(false),
		Snmp:     &DeviceSnmp{Location: ptr("rack-1")},
		Stp:      &DeviceStp{Priority: ptr("4096")},
		Switching: &DeviceSwitching{
			VlanEnabled: ptr(true),
			PoeMode:     ptr("auto"),
		},
		Led: &DeviceLed{
			Override: ptr("on"),
			EtherLighting: &DeviceEtherLightingArgs{
				LedMode:    ptr("etherlighting"),
				Mode:       ptr("speed"),
				Behavior:   ptr("breath"),
				Brightness: ptr(80),
			},
		},
		Lcm: &DeviceLcm{Brightness: ptr(50)},
		Outlet: &DeviceOutlet{
			Overrides: []DeviceOutletOverride{
				{Index: 1, Name: ptr("NAS"), RelayState: ptr(true), CycleEnabled: ptr(false)},
			},
		},
		RadioTable: []DeviceRadioOverride{
			{Radio: "na", Channel: ptr("36"), ChannelWidth: ptr(80), TxPower: ptr("18"), TxPowerMode: ptr("custom"), MinRssi: ptr(-75)},
		},
		EthernetOverrides: []DeviceEthernetOverride{
			{Ifname: "eth1", NetworkGroup: ptr("WAN2")},
		},
		PortOverrides: []DevicePortOverride{
			// Overlays the pre-existing port 1 (only sets PoE; name must survive).
			{PortIdx: 1, PoeMode: ptr("off")},
			// A brand-new override for port 5 with a broad spread of fields.
			{
				PortIdx:                    5,
				Name:                       ptr("Uplink"),
				OpMode:                     ptr("switch"),
				NativeNetworkId:            ptr("net123"),
				TaggedVlanMgmt:             ptr("custom"),
				ExcludedNetworkIds:         []string{"netA", "netB"},
				Forward:                    ptr("customize"),
				Speed:                      ptr(1000),
				Autoneg:                    ptr(true),
				StormctrlType:              ptr("level"),
				StormctrlBroadcastEnabled:  ptr(true),
				StormctrlBroadcastLevel:    ptr(50),
				PortSecurityEnabled:        ptr(true),
				PortSecurityMacAddress:     []string{"11:22:33:44:55:66"},
				Dot1xCtrl:                  ptr("auto"),
				EgressRateLimitKbps:        ptr(1000),
				EgressRateLimitKbpsEnabled: ptr(true),
				PriorityQueue1Level:        ptr(10),
				FecMode:                    ptr("default"),
				VoiceNetworkId:             ptr("voiceNet"),
				SettingPreference:          ptr("manual"),
			},
		},
	}

	// Simulate the device as fetched from the controller: adopted, with an
	// unmanaged SNMP contact and a pre-existing override on port 1.
	d := &unifi.Device{
		ID:          "dev-1",
		MAC:         "aa:bb:cc:dd:ee:ff",
		Model:       "USW-Pro-24-PoE",
		Type:        "usw",
		Adopted:     true,
		State:       unifi.DeviceStateConnected,
		SnmpContact: "noc@example.com",
		PortOverrides: []unifi.DevicePortOverrides{
			{PortIDX: 1, Name: "keep-me", Isolation: true},
		},
	}

	args.applyTo(d)

	// Managed top-level fields land.
	if d.Name != "Core Switch" {
		t.Errorf("Name = %q, want %q", d.Name, "Core Switch")
	}
	if d.PoeMode != "auto" {
		t.Errorf("PoeMode = %q, want auto", d.PoeMode)
	}
	if d.StpPriority != "4096" {
		t.Errorf("StpPriority = %q, want 4096", d.StpPriority)
	}
	if !d.SwitchVLANEnabled {
		t.Error("SwitchVLANEnabled = false, want true")
	}
	// Setting LcmBrightness must flip the override flag on.
	if d.LcmBrightness != 50 || !d.LcmBrightnessOverride {
		t.Errorf("LcmBrightness = %d override = %v, want 50/true", d.LcmBrightness, d.LcmBrightnessOverride)
	}

	// Unmanaged setting survives the read-modify-write.
	if d.SnmpContact != "noc@example.com" {
		t.Errorf("unmanaged SnmpContact was clobbered: %q", d.SnmpContact)
	}

	// EtherLighting nested object applied.
	if d.EtherLighting.LedMode != "etherlighting" || d.EtherLighting.Brightness != 80 {
		t.Errorf("EtherLighting = %+v", d.EtherLighting)
	}

	// Outlet, radio, ethernet overrides appended.
	if len(d.OutletOverrides) != 1 || d.OutletOverrides[0].Name != "NAS" || !d.OutletOverrides[0].RelayState {
		t.Errorf("OutletOverrides = %+v", d.OutletOverrides)
	}
	if len(d.RadioTable) != 1 || d.RadioTable[0].Channel != "36" || d.RadioTable[0].Ht != 80 || d.RadioTable[0].TxPower != "18" {
		t.Errorf("RadioTable = %+v", d.RadioTable)
	}
	// MinRssi auto-enables the feature when minRssiEnabled is not given.
	if d.RadioTable[0].MinRssi != -75 || !d.RadioTable[0].MinRssiEnabled {
		t.Errorf("MinRssi = %d enabled = %v, want -75/true", d.RadioTable[0].MinRssi, d.RadioTable[0].MinRssiEnabled)
	}
	if len(d.EthernetOverrides) != 1 || d.EthernetOverrides[0].NetworkGroup != "WAN2" {
		t.Errorf("EthernetOverrides = %+v", d.EthernetOverrides)
	}

	// Port overrides: pre-existing port 1 is overlaid (PoE set, name kept),
	// new port 5 is appended.
	if len(d.PortOverrides) != 2 {
		t.Fatalf("len(PortOverrides) = %d, want 2", len(d.PortOverrides))
	}
	p1 := findPort(t, d, 1)
	if p1.PoeMode != "off" {
		t.Errorf("port 1 PoeMode = %q, want off", p1.PoeMode)
	}
	if p1.Name != "keep-me" {
		t.Errorf("port 1 Name = %q, want keep-me (overlay must not clobber)", p1.Name)
	}
	if !p1.Isolation {
		t.Error("port 1 Isolation = false, want true (unmanaged field must survive)")
	}
	p5 := findPort(t, d, 5)
	if p5.Name != "Uplink" || p5.NATiveNetworkID != "net123" || p5.TaggedVLANMgmt != "custom" {
		t.Errorf("port 5 = %+v", p5)
	}
	if len(p5.ExcludedNetworkIDs) != 2 || p5.Speed != 1000 || !p5.Autoneg {
		t.Errorf("port 5 = %+v", p5)
	}
	if !p5.StormctrlBroadcastastEnabled || p5.StormctrlBroadcastastLevel != 50 {
		t.Errorf("port 5 stormctrl = %+v", p5)
	}
	if !p5.PortSecurityEnabled || len(p5.PortSecurityMACAddress) != 1 || p5.Dot1XCtrl != "auto" {
		t.Errorf("port 5 security = %+v", p5)
	}
	if p5.EgressRateLimitKbps != 1000 || !p5.EgressRateLimitKbpsEnabled || p5.PriorityQueue1Level != 10 {
		t.Errorf("port 5 rate/qos = %+v", p5)
	}
	if p5.FecMode != "default" || p5.VoiceNetworkID != "voiceNet" || p5.SettingPreference != "manual" {
		t.Errorf("port 5 misc = %+v", p5)
	}

	// State mapping: read-only facts and round-tripped managed inputs.
	st := deviceStateFrom(d, args)
	if st.DeviceId != "dev-1" || st.Model != "USW-Pro-24-PoE" || st.Type != "usw" {
		t.Errorf("state facts = %+v", st)
	}
	if st.State != "Connected" {
		t.Errorf("State = %q, want Connected", st.State)
	}
	if !st.Adopted {
		t.Error("Adopted = false, want true")
	}
	if st.Mac != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("Mac = %q", st.Mac)
	}
	if derefOr(st.Name, "") != "Core Switch" {
		t.Errorf("state Name = %v", st.Name)
	}
	// switching.poeMode now lives in the nested switching group.
	if st.Switching == nil || derefOr(st.Switching.PoeMode, "") != "auto" {
		t.Errorf("state switching = %+v", st.Switching)
	}
	// Unmanaged scalar (not in prior) stays nil — no spurious diff.
	if st.MgmtNetworkId != nil {
		t.Errorf("unmanaged MgmtNetworkId leaked into state: %v", *st.MgmtNetworkId)
	}
	// Unmanaged facet (not in prior) round-trips as a nil group.
	if st.Vrrp != nil {
		t.Errorf("unmanaged vrrp group leaked into state: %+v", st.Vrrp)
	}
	// Nested override lists are preserved from prior inputs (port overrides stay
	// top-level; outlet overrides now live under the outlet group).
	if len(st.PortOverrides) != 2 {
		t.Errorf("state port overrides not preserved: ports=%d", len(st.PortOverrides))
	}
	if st.Outlet == nil || len(st.Outlet.Overrides) != 1 {
		t.Errorf("state outlet overrides not preserved: %+v", st.Outlet)
	}
}

func findPort(t *testing.T, d *unifi.Device, idx int) *unifi.DevicePortOverrides {
	t.Helper()
	for i := range d.PortOverrides {
		if d.PortOverrides[i].PortIDX == idx {
			return &d.PortOverrides[i]
		}
	}
	t.Fatalf("port %d not found in %+v", idx, d.PortOverrides)
	return nil
}
