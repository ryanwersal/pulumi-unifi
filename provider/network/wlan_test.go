// SPDX-License-Identifier: Apache-2.0

package network

import "testing"

func TestWlanRoundTrip(t *testing.T) {
	args := WlanArgs{
		Name:         "Enterprise",
		NetworkId:    "net123",
		WlanGroupId:  ptr("grp1"),
		UserGroupId:  ptr("ug1"),
		Enabled:      ptr(true),
		Security:     ptr(WlanSecurityWpaPsk),
		Passphrase:   ptr("supersecret"),
		HideSsid:     ptr(false),
		UapsdEnabled: ptr(true),
		MloEnabled:   ptr(true),
		L2Isolation:  ptr(false),
		IsGuest:      ptr(false),
		Priority:     ptr("high"),

		Wpa: &WlanWpa{
			Mode:       ptr("wpa2"),
			Enc:        ptr("gcmp-256"),
			PskRadius:  ptr("optional"),
			PmfMode:    ptr("required"),
			PmfCipher:  ptr("bip-gmac-256"),
			GroupRekey: ptr(3600),
		},
		Wpa3: &WlanWpa3{
			Support:     ptr(true),
			Transition:  ptr(true),
			Enhanced192: ptr(false),
			FastRoaming: ptr(true),
		},
		Wep: &WlanWep{
			Key:   ptr("wepkey"),
			Index: ptr(2),
		},
		Roaming: &WlanRoaming{
			FastRoamingEnabled: ptr(true),
			BssTransition:      ptr(true),
			IappKey:            ptr("abcdef0123456789abcdef0123456789"),
		},
		Multicast: &WlanMulticast{
			EnhanceEnabled:         ptr(true),
			ProxyArp:               ptr(true),
			BroadcastFilterEnabled: ptr(true),
			BroadcastFilterList:    []string{"00:11:22:33:44:55"},
		},
		ApGroups: &WlanApGroups{
			Ids:  []string{"ap1", "ap2"},
			Mode: ptr("groups"),
		},
		VlanTagging: &WlanVlanTagging{
			Vlan:    ptr(20),
			Enabled: ptr(true),
		},
		BandSteering: &WlanBandSteering{
			Band:      ptr("both"),
			Bands:     []string{"2g", "5g", "6g"},
			No2GhzOui: ptr(true),
		},
		Dtim: &WlanDtim{
			Mode: ptr("custom"),
			Na:   ptr(3),
			Ng:   ptr(1),
			SixE: ptr(3),
		},
		MinRate: &WlanMinrate{
			NaEnabled:         ptr(true),
			NaDataRateKbps:    ptr(6000),
			NgEnabled:         ptr(true),
			NgDataRateKbps:    ptr(1000),
			SettingPreference: ptr("manual"),
		},
		MacFilter: &WlanMacFilter{
			Enabled: ptr(true),
			Policy:  ptr("allow"),
			List:    []string{"aa:bb:cc:dd:ee:ff"},
		},
		Schedule: &WlanSchedule{
			Enabled: ptr(true),
			Legacy:  []string{"mon|0900-1700"},
			Entries: []WlanScheduleEntry{{
				DurationMinutes: 480,
				StartHour:       9,
				StartMinute:     ptr(30),
				StartDaysOfWeek: []string{"mon", "tue"},
				Name:            ptr("workday"),
			}},
		},
		Dpi: &WlanDpi{
			Enabled: ptr(true),
			GroupId: ptr("dpi1"),
		},
		Iot: &WlanIot{
			Enhanced:                 ptr(true),
			OptimizeWifiConnectivity: ptr(true),
		},
		P2p: &WlanP2p{
			Enabled:      ptr(true),
			CrossConnect: ptr(true),
		},
		Radius: &WlanRadius{
			ProfileId:           ptr("rad1"),
			MacAuthEnabled:      ptr(true),
			MacaclFormat:        ptr("colon_lower"),
			MacaclEmptyPassword: ptr(true),
			DasEnabled:          ptr(true),
			NasIdentifier:       ptr("nas1"),
			NasIdentifierType:   ptr("custom"),
		},
		PrivatePresharedKeys: &WlanPrivatePresharedKeys{
			Enabled: ptr(true),
			Keys: []WlanPrivatePsk{{
				Password:  "pskpass1",
				NetworkId: ptr("net456"),
			}},
		},
		Sae: &WlanSae{
			Psks: []WlanSaePsk{{
				Psk:  "saepass1",
				Id:   ptr("sae1"),
				Mac:  ptr("11:22:33:44:55:66"),
				Vlan: ptr(30),
			}},
			Groups:       []int{19, 20},
			AntiClogging: ptr(5),
			Sync:         ptr(5),
		},
	}

	u := args.toUnifi("wlan-id")
	if u.ID != "wlan-id" {
		t.Fatalf("ID = %q, want wlan-id", u.ID)
	}
	if u.XPassphrase != "supersecret" {
		t.Fatalf("XPassphrase = %q", u.XPassphrase)
	}
	if !u.WPA3Support || !u.MloEnabled || u.PMFMode != "required" {
		t.Fatalf("security/roaming flags not mapped: %+v", u)
	}
	if u.GroupRekey != 3600 {
		t.Fatalf("groupRekey not mapped: %d", u.GroupRekey)
	}
	if u.XIappKey != "abcdef0123456789abcdef0123456789" {
		t.Fatalf("iappKey secret not mapped: %q", u.XIappKey)
	}
	if u.XWEP != "wepkey" {
		t.Fatalf("wep key secret not mapped: %q", u.XWEP)
	}
	if u.VLAN != 20 || !u.VLANEnabled {
		t.Fatalf("vlan not mapped: %d enabled=%v", u.VLAN, u.VLANEnabled)
	}
	if len(u.WLANBands) != 3 || u.WLANBands[2] != "6g" {
		t.Fatalf("wlan bands not mapped: %v", u.WLANBands)
	}
	if u.RADIUSProfileID != "rad1" || u.NasIDentifier != "nas1" {
		t.Fatalf("radius not mapped: profile=%q nas=%q", u.RADIUSProfileID, u.NasIDentifier)
	}
	if len(u.PrivatePresharedKeys) != 1 || u.PrivatePresharedKeys[0].Password != "pskpass1" {
		t.Fatalf("private psk not mapped: %+v", u.PrivatePresharedKeys)
	}
	if len(u.SaePsk) != 1 || u.SaePsk[0].Psk != "saepass1" || u.SaePsk[0].VLAN != 30 {
		t.Fatalf("sae psk not mapped: %+v", u.SaePsk)
	}
	if len(u.ScheduleWithDuration) != 1 || u.ScheduleWithDuration[0].DurationMinutes != 480 {
		t.Fatalf("schedule not mapped: %+v", u.ScheduleWithDuration)
	}

	st := wlanStateFrom(u, args)
	if st.WlanId != "wlan-id" {
		t.Fatalf("WlanId = %q", st.WlanId)
	}
	if st.Name != "Enterprise" || st.NetworkId != "net123" {
		t.Fatalf("identity not preserved: %+v", st.WlanArgs)
	}
	if st.Passphrase == nil || *st.Passphrase != "supersecret" {
		t.Fatalf("passphrase secret not preserved: %v", st.Passphrase)
	}
	if st.UapsdEnabled == nil || !*st.UapsdEnabled {
		t.Fatalf("uapsdEnabled not round-tripped: %v", st.UapsdEnabled)
	}
	if st.Priority == nil || *st.Priority != "high" {
		t.Fatalf("priority not round-tripped: %v", st.Priority)
	}

	if st.Wpa == nil {
		t.Fatal("wpa group lost on round-trip")
	}
	if st.Wpa.PmfMode == nil || *st.Wpa.PmfMode != "required" {
		t.Fatalf("wpa.pmfMode not round-tripped: %v", st.Wpa.PmfMode)
	}
	if st.Wpa.GroupRekey == nil || *st.Wpa.GroupRekey != 3600 {
		t.Fatalf("wpa.groupRekey not round-tripped: %v", st.Wpa.GroupRekey)
	}

	if st.Wpa3 == nil {
		t.Fatal("wpa3 group lost on round-trip")
	}
	if st.Wpa3.Support == nil || !*st.Wpa3.Support {
		t.Fatalf("wpa3.support not round-tripped")
	}

	if st.Wep == nil {
		t.Fatal("wep group lost on round-trip")
	}
	if st.Wep.Key == nil || *st.Wep.Key != "wepkey" {
		t.Fatalf("wep.key secret not preserved: %v", st.Wep.Key)
	}

	if st.Roaming == nil {
		t.Fatal("roaming group lost on round-trip")
	}
	if st.Roaming.IappKey == nil || *st.Roaming.IappKey != "abcdef0123456789abcdef0123456789" {
		t.Fatalf("roaming.iappKey secret not preserved: %v", st.Roaming.IappKey)
	}

	if st.VlanTagging == nil {
		t.Fatal("vlanTagging group lost on round-trip")
	}
	if st.VlanTagging.Vlan == nil || *st.VlanTagging.Vlan != 20 {
		t.Fatalf("vlanTagging.vlan not round-tripped: %v", st.VlanTagging.Vlan)
	}

	if st.BandSteering == nil {
		t.Fatal("bandSteering group lost on round-trip")
	}
	if len(st.BandSteering.Bands) != 3 {
		t.Fatalf("bandSteering.bands not round-tripped: %v", st.BandSteering.Bands)
	}

	if st.Dtim == nil {
		t.Fatal("dtim group lost on round-trip")
	}
	if st.Dtim.SixE == nil || *st.Dtim.SixE != 3 {
		t.Fatalf("dtim.sixE not round-tripped: %v", st.Dtim.SixE)
	}

	if st.MinRate == nil {
		t.Fatal("minRate group lost on round-trip")
	}
	if st.MinRate.NaDataRateKbps == nil || *st.MinRate.NaDataRateKbps != 6000 {
		t.Fatalf("minRate.naDataRateKbps not round-tripped: %v", st.MinRate.NaDataRateKbps)
	}

	if st.MacFilter == nil {
		t.Fatal("macFilter group lost on round-trip")
	}
	if st.MacFilter.Policy == nil || *st.MacFilter.Policy != "allow" {
		t.Fatalf("macFilter.policy not round-tripped: %v", st.MacFilter.Policy)
	}

	if st.Multicast == nil {
		t.Fatal("multicast group lost on round-trip")
	}
	if len(st.Multicast.BroadcastFilterList) != 1 || st.Multicast.BroadcastFilterList[0] != "00:11:22:33:44:55" {
		t.Fatalf("multicast.broadcastFilterList not round-tripped: %v", st.Multicast.BroadcastFilterList)
	}

	if st.Schedule == nil {
		t.Fatal("schedule group lost on round-trip")
	}
	if len(st.Schedule.Entries) != 1 || st.Schedule.Entries[0].StartHour != 9 {
		t.Fatalf("schedule.entries not round-tripped: %+v", st.Schedule.Entries)
	}
	if len(st.Schedule.Legacy) != 1 || st.Schedule.Legacy[0] != "mon|0900-1700" {
		t.Fatalf("schedule.legacy not round-tripped: %v", st.Schedule.Legacy)
	}

	if st.ApGroups == nil {
		t.Fatal("apGroups group lost on round-trip")
	}
	if len(st.ApGroups.Ids) != 2 || st.ApGroups.Mode == nil || *st.ApGroups.Mode != "groups" {
		t.Fatalf("apGroups not round-tripped: %+v", st.ApGroups)
	}

	if st.Dpi == nil {
		t.Fatal("dpi group lost on round-trip")
	}
	if st.Dpi.GroupId == nil || *st.Dpi.GroupId != "dpi1" {
		t.Fatalf("dpi.groupId not round-tripped: %v", st.Dpi.GroupId)
	}

	if st.Iot == nil {
		t.Fatal("iot group lost on round-trip")
	}
	if st.Iot.Enhanced == nil || !*st.Iot.Enhanced {
		t.Fatalf("iot.enhanced not round-tripped")
	}

	if st.P2p == nil {
		t.Fatal("p2p group lost on round-trip")
	}
	if st.P2p.CrossConnect == nil || !*st.P2p.CrossConnect {
		t.Fatalf("p2p.crossConnect not round-tripped")
	}

	if st.Radius == nil {
		t.Fatal("radius group lost on round-trip")
	}
	if st.Radius.ProfileId == nil || *st.Radius.ProfileId != "rad1" {
		t.Fatalf("radius.profileId not round-tripped: %v", st.Radius.ProfileId)
	}

	if st.PrivatePresharedKeys == nil {
		t.Fatal("privatePresharedKeys group lost on round-trip")
	}
	if len(st.PrivatePresharedKeys.Keys) != 1 || st.PrivatePresharedKeys.Keys[0].Password != "pskpass1" {
		t.Fatalf("private psk secret not preserved: %+v", st.PrivatePresharedKeys.Keys)
	}

	if st.Sae == nil {
		t.Fatal("sae group lost on round-trip")
	}
	if len(st.Sae.Psks) != 1 || st.Sae.Psks[0].Psk != "saepass1" {
		t.Fatalf("sae psk secret not preserved: %+v", st.Sae.Psks)
	}
	if len(st.Sae.Groups) != 2 {
		t.Fatalf("sae.groups not round-tripped: %v", st.Sae.Groups)
	}
}

func TestWlanDefaults(t *testing.T) {
	args := WlanArgs{Name: "Basic", NetworkId: "n1"}
	u := args.toUnifi("")
	if !u.Enabled {
		t.Fatalf("enabled default should be true")
	}
	if u.Security != "wpapsk" {
		t.Fatalf("security default = %q, want wpapsk", u.Security)
	}
	if u.HideSSID {
		t.Fatalf("hideSsid default should be false")
	}
}

// TestWlanEmptyGroupsStayNil asserts that round-tripping a minimal WLAN does not
// synthesize non-nil facet groups out of controller zero values.
func TestWlanEmptyGroupsStayNil(t *testing.T) {
	u := WlanArgs{Name: "Basic", NetworkId: "n1"}.toUnifi("")
	u.ID = "wlan-min"
	st := wlanStateFrom(u, WlanArgs{Name: "Basic", NetworkId: "n1"})
	out := st.WlanArgs
	if out.Wpa != nil {
		t.Errorf("wpa should be nil, got %+v", out.Wpa)
	}
	if out.Wpa3 != nil {
		t.Errorf("wpa3 should be nil, got %+v", out.Wpa3)
	}
	if out.Sae != nil {
		t.Errorf("sae should be nil, got %+v", out.Sae)
	}
	if out.Wep != nil {
		t.Errorf("wep should be nil, got %+v", out.Wep)
	}
	if out.PrivatePresharedKeys != nil {
		t.Errorf("privatePresharedKeys should be nil, got %+v", out.PrivatePresharedKeys)
	}
	if out.Radius != nil {
		t.Errorf("radius should be nil, got %+v", out.Radius)
	}
	if out.VlanTagging != nil {
		t.Errorf("vlanTagging should be nil, got %+v", out.VlanTagging)
	}
	if out.BandSteering != nil {
		t.Errorf("bandSteering should be nil, got %+v", out.BandSteering)
	}
	if out.Dtim != nil {
		t.Errorf("dtim should be nil, got %+v", out.Dtim)
	}
	if out.MinRate != nil {
		t.Errorf("minRate should be nil, got %+v", out.MinRate)
	}
	if out.MacFilter != nil {
		t.Errorf("macFilter should be nil, got %+v", out.MacFilter)
	}
	if out.Multicast != nil {
		t.Errorf("multicast should be nil, got %+v", out.Multicast)
	}
	if out.Schedule != nil {
		t.Errorf("schedule should be nil, got %+v", out.Schedule)
	}
	if out.ApGroups != nil {
		t.Errorf("apGroups should be nil, got %+v", out.ApGroups)
	}
	if out.Dpi != nil {
		t.Errorf("dpi should be nil, got %+v", out.Dpi)
	}
	if out.Iot != nil {
		t.Errorf("iot should be nil, got %+v", out.Iot)
	}
	if out.P2p != nil {
		t.Errorf("p2p should be nil, got %+v", out.P2p)
	}
	if out.Roaming != nil {
		t.Errorf("roaming should be nil, got %+v", out.Roaming)
	}
}
