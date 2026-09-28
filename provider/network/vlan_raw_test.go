// SPDX-License-Identifier: Apache-2.0

package network

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/filipowm/go-unifi/unifi"
)

func TestVlanWANUpdatePreservesLiveConfiguration(t *testing.T) {
	live := map[string]any{
		"_id": "wan-id", "name": "Internet 1", "purpose": "wan", "enabled": true,
		"wan_type": "dhcp", "wan_type_v6": "disabled", "wan_networkgroup": "WAN",
		"wan_dhcpv6_pd_size_auto": true, "ipv6_wan_delegation_type": "none",
		"routing_table_id": float64(201), "wan_failover_priority": float64(1),
		"firewall_zone_id": "external-zone", "attr_no_delete": true,
		"wan_provider_capabilities": map[string]any{"download_kilobits_per_second": float64(1203000)},
		"future_controller_setting": true,
	}
	state, err := vlanRawState(live, VlanArgs{})
	if err != nil {
		t.Fatal(err)
	}
	desired := state.VlanArgs
	wan := *desired.Wan
	wan.TypeV6 = ptr(VlanWanTypeV6("dhcpv6"))
	wan.Dhcpv6PdSize = ptr(64)
	wan.Dhcpv6PdSizeAuto = ptr(false)
	desired.Wan = &wan
	ipv6 := *desired.Ipv6
	ipv6.WanDelegationType = ptr(VlanIpv6WanDelegationType("pd"))
	desired.Ipv6 = &ipv6
	got, err := vlanUpdatePayload(live, state.VlanArgs, desired, "wan-id")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{}
	for key, value := range live {
		want[key] = value
	}
	want["wan_type_v6"] = "dhcpv6"
	want["wan_dhcpv6_pd_size"] = float64(64)
	want["wan_dhcpv6_pd_size_auto"] = false
	want["ipv6_wan_delegation_type"] = "pd"
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected WAN update:\ngot  %#v\nwant %#v", got, want)
	}
	if live["wan_type_v6"] != "disabled" {
		t.Fatal("mutated original live object")
	}
	// Explicit false survives Read; drift to true is visible rather than hidden
	// by a prior false input, and the next update corrects it.
	read, err := vlanRawState(got, desired)
	if err != nil {
		t.Fatal(err)
	}
	if read.Wan.Dhcpv6PdSizeAuto == nil || *read.Wan.Dhcpv6PdSizeAuto {
		t.Fatal("lost false on read")
	}
	got["wan_dhcpv6_pd_size_auto"] = true
	drift, err := vlanRawState(got, desired)
	if err != nil {
		t.Fatal(err)
	}
	fixed, err := vlanUpdatePayload(got, drift.VlanArgs, desired, "wan-id")
	if err != nil {
		t.Fatal(err)
	}
	if fixed["wan_dhcpv6_pd_size_auto"] != false {
		t.Fatal("did not correct refreshed drift")
	}
}

func TestVlanPrefixSizeAutoOptional(t *testing.T) {
	for _, value := range []*bool{nil, ptr(false), ptr(true)} {
		args := VlanArgs{Name: "WAN", Wan: &VlanWan{Dhcpv6PdSizeAuto: value}}
		body, err := vlanPayload(args, "")
		if err != nil {
			t.Fatal(err)
		}
		got, exists := body["wan_dhcpv6_pd_size_auto"]
		if value == nil && exists {
			t.Fatal("unset option sent")
		}
		if value != nil && (!exists || got != *value) {
			t.Fatal("explicit option lost")
		}
	}
}

type vlanRawClient struct {
	unifi.Client
	calls int
}

func (c *vlanRawClient) Do(_ context.Context, method, path string, body, response any) error {
	c.calls++
	if method != http.MethodGet || path != "s/default/rest/networkconf/wan-id" || body != nil {
		panic("unexpected request")
	}
	return json.Unmarshal([]byte(`{"meta":{"rc":"ok"},"data":[{"_id":"wan-id","wan_dhcpv6_pd_size_auto":false,"future_field":true}]}`), response)
}
func TestVlanRawRequestRetainsUnknownFields(t *testing.T) {
	client := &vlanRawClient{}
	got, err := vlanRequest(context.Background(), client, http.MethodGet, "default", "wan-id", nil)
	if err != nil {
		t.Fatal(err)
	}
	if got["future_field"] != true || got["wan_dhcpv6_pd_size_auto"] != false || client.calls != 1 {
		t.Fatalf("unexpected response: %#v", got)
	}
}
