// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/pulumi/pulumi-go-provider/integration"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
)

type fakeVlanNetwork struct {
	unifi.Client
	live map[string]any
}

func (f *fakeVlanNetwork) Do(_ context.Context, method, path string, body, response any) error {
	switch method {
	case http.MethodPost:
		f.live = body.(map[string]any)
		f.live["_id"] = "wan-id"
		f.live["firewall_zone_id"] = "external-zone"
		f.live["wan_failover_priority"] = float64(1)
	case http.MethodGet:
		if f.live == nil {
			return unifi.ErrNotFound
		}
	case http.MethodPut:
		next := body.(map[string]any)
		if next["firewall_zone_id"] != "external-zone" || next["wan_failover_priority"] != float64(1) {
			return fmt.Errorf("lost live WAN fields")
		}
		f.live = next
	default:
		return fmt.Errorf("unexpected request %s %s", method, path)
	}
	data, err := json.Marshal(map[string]any{"data": []map[string]any{f.live}})
	if err != nil {
		return err
	}
	return json.Unmarshal(data, response)
}
func (f *fakeVlanNetwork) DeleteNetwork(_ context.Context, _, _ string) error {
	f.live = nil
	return nil
}

func TestVlanWANLifecycle(t *testing.T) {
	client := &fakeVlanNetwork{}
	server := newLifecycleServer(t, client, nil)
	inputs := func(mode string, auto bool) property.Map {
		return pmap(map[string]property.Value{
			"name": property.New("Internet 1"), "purpose": property.New("wan"),
			"wan": property.New(pmap(map[string]property.Value{
				"type": property.New("dhcp"), "typeV6": property.New(mode),
				"dhcpv6PdSize": property.New(float64(64)), "dhcpv6PdSizeAuto": property.New(auto),
			})),
		})
	}
	integration.LifeCycleTest{
		Resource: "unifi:network:Vlan",
		Create:   integration.Operation{Inputs: inputs("disabled", true)},
		Updates: []integration.Operation{{Inputs: inputs("dhcpv6", false), Hook: func(_, out property.Map) {
			eq(t, out, "networkId", "wan-id")
			if client.live["network_group"] == "LAN" {
				t.Fatal("WAN received LAN group default")
			}
			if client.live["wan_type_v6"] != "dhcpv6" || client.live["wan_dhcpv6_pd_size_auto"] != false {
				t.Fatalf("WAN IPv6 not updated: %#v", client.live)
			}
		}}},
	}.Run(t, server)
}
