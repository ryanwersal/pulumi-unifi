// SPDX-License-Identifier: Apache-2.0
package network

import (
	"context"
	"testing"

	"github.com/pulumi/pulumi-go-provider/infer"
)

func TestDynamicDnsCredentialRoundTrip(t *testing.T) {
	args := DynamicDnsArgs{HostName: "atlas.example.com", Service: "cloudflare", Login: "example.com", Password: "test-token"}
	wire := args.toUnifi("ddns-id")
	if wire.Interface != "wan" || wire.Server != "" || wire.XPassword != args.Password {
		t.Fatal("incorrect DDNS defaults or missing credential")
	}
	for _, secret := range []string{"", "********", "test-token"} {
		wire.XPassword = secret
		state := dynamicDnsStateFrom(wire, args)
		if state.Password != args.Password || state.DynamicDnsId != "ddns-id" || state.Server != nil {
			t.Fatal("read lost a masked credential or introduced an optional field")
		}
	}
	wire.XPassword = "changed-token"
	if dynamicDnsStateFrom(wire, args).Password != "changed-token" {
		t.Fatal("read failed to report an unmasked changed credential")
	}
}

func TestDynamicDnsPreviewDoesNotConnect(t *testing.T) {
	resource := DynamicDns{}
	args := DynamicDnsArgs{HostName: "atlas.example.com", Service: "cloudflare", Login: "example.com", Password: "test-token"}
	if _, err := resource.Create(context.Background(), infer.CreateRequest[DynamicDnsArgs]{Inputs: args, DryRun: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := resource.Update(context.Background(), infer.UpdateRequest[DynamicDnsArgs, DynamicDnsState]{ID: "ddns-id", Inputs: args, DryRun: true}); err != nil {
		t.Fatal(err)
	}
}
