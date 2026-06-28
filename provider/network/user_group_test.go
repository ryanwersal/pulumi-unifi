// SPDX-License-Identifier: Apache-2.0

package network

import "testing"

func TestUserGroupRoundTrip(t *testing.T) {
	args := UserGroupArgs{
		Name:           "limited",
		QosRateMaxDown: ptr(10000),
		QosRateMaxUp:   ptr(2000),
	}

	u := args.toUnifi("ug-1")
	if u.ID != "ug-1" {
		t.Fatalf("ID = %q, want ug-1", u.ID)
	}
	if u.Name != "limited" {
		t.Fatalf("Name = %q, want limited", u.Name)
	}
	if u.QOSRateMaxDown != 10000 {
		t.Fatalf("QOSRateMaxDown = %d, want 10000", u.QOSRateMaxDown)
	}
	if u.QOSRateMaxUp != 2000 {
		t.Fatalf("QOSRateMaxUp = %d, want 2000", u.QOSRateMaxUp)
	}

	st := userGroupStateFrom(u, args)
	if st.UserGroupId != "ug-1" {
		t.Fatalf("UserGroupId = %q, want ug-1", st.UserGroupId)
	}
	if st.Name != "limited" {
		t.Fatalf("state Name = %q, want limited", st.Name)
	}
	if st.QosRateMaxDown == nil || *st.QosRateMaxDown != 10000 {
		t.Fatalf("state QosRateMaxDown = %v, want 10000", st.QosRateMaxDown)
	}
	if st.QosRateMaxUp == nil || *st.QosRateMaxUp != 2000 {
		t.Fatalf("state QosRateMaxUp = %v, want 2000", st.QosRateMaxUp)
	}
}

func TestUserGroupDefaults(t *testing.T) {
	args := UserGroupArgs{Name: "default"}
	u := args.toUnifi("")
	if u.QOSRateMaxDown != -1 {
		t.Fatalf("QOSRateMaxDown default = %d, want -1", u.QOSRateMaxDown)
	}
	if u.QOSRateMaxUp != -1 {
		t.Fatalf("QOSRateMaxUp default = %d, want -1", u.QOSRateMaxUp)
	}
}
