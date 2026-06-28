package network

import "testing"

func TestDnsRecordRoundTrip(t *testing.T) {
	args := DnsRecordArgs{
		Key:        "_sip._tcp.example.com",
		RecordType: "SRV",
		Value:      "sip.example.com",
		Enabled:    ptr(true),
		Ttl:        ptr(3600),
		Priority:   ptr(10),
		Port:       ptr(5060),
		Weight:     ptr(5),
	}

	u := args.toUnifi("rec-123")
	if u.ID != "rec-123" {
		t.Fatalf("ID = %q, want rec-123", u.ID)
	}
	if u.Key != args.Key {
		t.Errorf("Key = %q, want %q", u.Key, args.Key)
	}
	if u.RecordType != "SRV" {
		t.Errorf("RecordType = %q, want SRV", u.RecordType)
	}
	if u.Value != "sip.example.com" {
		t.Errorf("Value = %q, want sip.example.com", u.Value)
	}
	if !u.Enabled {
		t.Errorf("Enabled = false, want true")
	}
	if u.Ttl != 3600 || u.Priority != 10 || u.Port != 5060 || u.Weight != 5 {
		t.Errorf("numeric fields mismatch: ttl=%d priority=%d port=%d weight=%d", u.Ttl, u.Priority, u.Port, u.Weight)
	}

	st := dnsRecordStateFrom(u, args)
	if st.DnsRecordId != "rec-123" {
		t.Errorf("DnsRecordId = %q, want rec-123", st.DnsRecordId)
	}
	if st.Key != args.Key || st.RecordType != args.RecordType || st.Value != args.Value {
		t.Errorf("identity fields did not survive round-trip: %+v", st.DnsRecordArgs)
	}
	if st.Enabled == nil || !*st.Enabled {
		t.Errorf("Enabled did not survive round-trip")
	}
	if st.Ttl == nil || *st.Ttl != 3600 {
		t.Errorf("Ttl did not survive round-trip")
	}
	if st.Priority == nil || *st.Priority != 10 {
		t.Errorf("Priority did not survive round-trip")
	}
	if st.Port == nil || *st.Port != 5060 {
		t.Errorf("Port did not survive round-trip")
	}
	if st.Weight == nil || *st.Weight != 5 {
		t.Errorf("Weight did not survive round-trip")
	}
}

func TestDnsRecordDefaultsAndPreserve(t *testing.T) {
	// Minimal A record: optional fields unset.
	args := DnsRecordArgs{
		Key:        "host.example.com",
		RecordType: "A",
		Value:      "192.168.1.10",
	}

	u := args.toUnifi("")
	if u.ID != "" {
		t.Errorf("ID = %q, want empty on create", u.ID)
	}
	if !u.Enabled {
		t.Errorf("Enabled default = false, want true")
	}
	if u.Ttl != 0 || u.Priority != 0 || u.Port != 0 || u.Weight != 0 {
		t.Errorf("unset optionals should be zero, got ttl=%d priority=%d port=%d weight=%d", u.Ttl, u.Priority, u.Port, u.Weight)
	}

	// Controller echoes zeros; prior had nils -> stay nil (no spurious diffs).
	st := dnsRecordStateFrom(u, args)
	if st.Ttl != nil || st.Priority != nil || st.Port != nil || st.Weight != nil {
		t.Errorf("zero controller values with nil prior should remain nil")
	}
}
