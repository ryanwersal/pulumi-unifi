// SPDX-License-Identifier: Apache-2.0
package provider

import (
	"context"
	"testing"

	"github.com/filipowm/go-unifi/unifi"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/integration"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
)

type fakeDynamicDnsNetwork struct {
	unifi.Client
	live *unifi.DynamicDNS
}

func (f *fakeDynamicDnsNetwork) CreateDynamicDNS(_ context.Context, _ string, v *unifi.DynamicDNS) (*unifi.DynamicDNS, error) {
	f.live = v
	f.live.ID = "ddns-id"
	return f.live, nil
}
func (f *fakeDynamicDnsNetwork) GetDynamicDNS(_ context.Context, _, _ string) (*unifi.DynamicDNS, error) {
	if f.live == nil {
		return nil, unifi.ErrNotFound
	}
	result := *f.live
	result.XPassword = ""
	return &result, nil
}
func (f *fakeDynamicDnsNetwork) UpdateDynamicDNS(_ context.Context, _ string, v *unifi.DynamicDNS) (*unifi.DynamicDNS, error) {
	f.live = v
	return f.GetDynamicDNS(context.Background(), "", v.ID)
}
func (f *fakeDynamicDnsNetwork) DeleteDynamicDNS(_ context.Context, _, _ string) error {
	if f.live == nil {
		return unifi.ErrNotFound
	}
	f.live = nil
	return nil
}

func TestDynamicDnsLifecycle(t *testing.T) {
	client := &fakeDynamicDnsNetwork{}
	server := newLifecycleServer(t, client, nil)
	inputs := func(host, token string) property.Map {
		return pmap(map[string]property.Value{
			"hostName": property.New(host), "service": property.New("cloudflare"),
			"login": property.New("example.com"), "password": property.New(token).WithSecret(true),
		})
	}
	integration.LifeCycleTest{
		Resource: "unifi:network:DynamicDns",
		Create:   integration.Operation{Inputs: inputs("atlas.example.com", "test-token")},
		Updates: []integration.Operation{{Inputs: inputs("vpn.example.com", "new-token"), Hook: func(_, out property.Map) {
			eq(t, out, "dynamicDnsId", "ddns-id")
			if !out.Get("password").Secret() {
				t.Fatal("credential is not secret")
			}
			if client.live.HostName != "vpn.example.com" || client.live.XPassword != "new-token" {
				t.Fatal("update did not reach controller")
			}
		}}},
	}.Run(t, server)
	const urn = "urn:pulumi:test::test::unifi:network:DynamicDns::ddns"
	state := inputs("vpn.example.com", "new-token").Set("dynamicDnsId", property.New("ddns-id"))
	if err := server.Delete(p.DeleteRequest{Urn: urn, ID: "ddns-id", Properties: state}); err != nil {
		t.Fatalf("delete missing entry: %v", err)
	}
	read, err := server.Read(p.ReadRequest{Urn: urn, ID: "ddns-id", Properties: state, Inputs: inputs("vpn.example.com", "new-token")})
	if err != nil || read.ID != "" {
		t.Fatalf("read missing entry: %v", err)
	}
}
