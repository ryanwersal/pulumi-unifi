package network

import "testing"

func TestUserRoundTrip(t *testing.T) {
	args := UserArgs{
		Mac:                      "00:11:22:33:44:55",
		Name:                     ptr("workstation"),
		Note:                     ptr("desk client"),
		UserGroupId:              ptr("ug-1"),
		NetworkId:                ptr("net-1"),
		FixedIp:                  ptr("192.168.20.10"),
		Blocked:                  ptr(true),
		LocalDnsRecord:           ptr("desk.local"),
		DevIdOverride:            ptr(44),
		VirtualNetworkOverrideId: ptr("vnet-1"),
		FixedApMac:               ptr("aa:bb:cc:dd:ee:ff"),
	}

	u := args.toUnifi("usr-1")
	if u.ID != "usr-1" {
		t.Fatalf("ID = %q, want usr-1", u.ID)
	}
	if u.MAC != "00:11:22:33:44:55" {
		t.Fatalf("MAC = %q", u.MAC)
	}
	if u.Name != "workstation" {
		t.Fatalf("Name = %q", u.Name)
	}
	if u.UserGroupID != "ug-1" {
		t.Fatalf("UserGroupID = %q", u.UserGroupID)
	}
	if u.NetworkID != "net-1" {
		t.Fatalf("NetworkID = %q", u.NetworkID)
	}
	// Setting FixedIp should auto-enable UseFixedIP.
	if u.FixedIP != "192.168.20.10" || !u.UseFixedIP {
		t.Fatalf("FixedIP = %q UseFixedIP = %v", u.FixedIP, u.UseFixedIP)
	}
	if !u.Blocked {
		t.Fatalf("Blocked = %v, want true", u.Blocked)
	}
	// Setting LocalDnsRecord should auto-enable LocalDNSRecordEnabled.
	if u.LocalDNSRecord != "desk.local" || !u.LocalDNSRecordEnabled {
		t.Fatalf("LocalDNSRecord = %q enabled = %v", u.LocalDNSRecord, u.LocalDNSRecordEnabled)
	}
	if u.DevIdOverride != 44 {
		t.Fatalf("DevIdOverride = %d, want 44", u.DevIdOverride)
	}
	if u.VirtualNetworkOverrideID != "vnet-1" || !u.VirtualNetworkOverrideEnabled {
		t.Fatalf("VirtualNetworkOverrideID = %q enabled = %v", u.VirtualNetworkOverrideID, u.VirtualNetworkOverrideEnabled)
	}
	if u.FixedApMAC != "aa:bb:cc:dd:ee:ff" || !u.FixedApEnabled {
		t.Fatalf("FixedApMAC = %q enabled = %v", u.FixedApMAC, u.FixedApEnabled)
	}

	// Simulate controller echoing back computed fields.
	u.ID = "usr-1"
	u.Hostname = "workstation.lan"
	u.IP = "192.168.20.10"
	u.LastSeen = 1700000000

	st := userStateFrom(u, args)
	if st.UserId != "usr-1" {
		t.Fatalf("UserId = %q, want usr-1", st.UserId)
	}
	if st.Mac != "00:11:22:33:44:55" {
		t.Fatalf("state Mac = %q", st.Mac)
	}
	if st.Name == nil || *st.Name != "workstation" {
		t.Fatalf("state Name = %v", st.Name)
	}
	if st.FixedIp == nil || *st.FixedIp != "192.168.20.10" {
		t.Fatalf("state FixedIp = %v", st.FixedIp)
	}
	if st.UseFixedIp == nil || !*st.UseFixedIp {
		t.Fatalf("state UseFixedIp = %v", st.UseFixedIp)
	}
	if st.Blocked == nil || !*st.Blocked {
		t.Fatalf("state Blocked = %v", st.Blocked)
	}
	if st.LocalDnsRecord == nil || *st.LocalDnsRecord != "desk.local" {
		t.Fatalf("state LocalDnsRecord = %v", st.LocalDnsRecord)
	}
	if st.DevIdOverride == nil || *st.DevIdOverride != 44 {
		t.Fatalf("state DevIdOverride = %v", st.DevIdOverride)
	}
	if st.VirtualNetworkOverrideId == nil || *st.VirtualNetworkOverrideId != "vnet-1" {
		t.Fatalf("state VirtualNetworkOverrideId = %v", st.VirtualNetworkOverrideId)
	}
	if st.FixedApMac == nil || *st.FixedApMac != "aa:bb:cc:dd:ee:ff" {
		t.Fatalf("state FixedApMac = %v", st.FixedApMac)
	}
	if st.Hostname != "workstation.lan" {
		t.Fatalf("state Hostname = %q", st.Hostname)
	}
	if st.Ip != "192.168.20.10" {
		t.Fatalf("state Ip = %q", st.Ip)
	}
	if st.LastSeen != 1700000000 {
		t.Fatalf("state LastSeen = %d", st.LastSeen)
	}
}

func TestUserMinimal(t *testing.T) {
	args := UserArgs{Mac: "de:ad:be:ef:00:01"}
	u := args.toUnifi("")
	if u.ID != "" {
		t.Fatalf("ID = %q, want empty on create", u.ID)
	}
	if u.MAC != "de:ad:be:ef:00:01" {
		t.Fatalf("MAC = %q", u.MAC)
	}
	if u.UseFixedIP {
		t.Fatalf("UseFixedIP = true, want false when FixedIp unset")
	}
}
