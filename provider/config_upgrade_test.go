// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"testing"

	p "github.com/pulumi/pulumi-go-provider"
	rpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/ryanwersal/pulumi-unifi/provider/config"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestProviderUpgradePreservesResources(t *testing.T) {
	defer config.InjectClientsForTest(&fakeVlanNetwork{}, nil, nil)()
	prov, err := New()
	if err != nil {
		t.Fatal(err)
	}
	server, err := p.RawServer(Name, "0.2.1", prov)(nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	inputs := func(version, url string) *structpb.Struct {
		m, err := structpb.NewStruct(map[string]any{
			"url": url, "apiKey": "test-key", "site": "default",
			"version": version, "__pulumi-go-provider-infer": true,
			"__pulumi-go-provider-version": "v1.3.2",
		})
		if err != nil {
			t.Fatal(err)
		}
		return m
	}
	old := inputs("0.1.0", "https://controller.invalid")
	_, err = server.Configure(ctx, &rpc.ConfigureRequest{Args: old, SendsOldInputs: true, SendsOldInputsToDelete: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, changed := range []bool{false, true} {
		url := "https://controller.invalid"
		if changed {
			url = "https://different-controller.invalid"
		}
		next := inputs("0.2.1", url)
		delete(next.Fields, "__pulumi-go-provider-version")
		resp, err := server.DiffConfig(ctx, &rpc.DiffRequest{Olds: old, OldInputs: old, News: next})
		if err != nil {
			t.Fatal(err)
		}
		if changed {
			if len(resp.Replaces) != 1 || resp.Replaces[0] != "url" {
				t.Fatalf("controller change must replace: %v", resp)
			}
		} else if resp.Changes != rpc.DiffResponse_DIFF_NONE || len(resp.Replaces) != 0 {
			t.Fatalf("version-only upgrade must preserve resources: %v", resp)
		}
	}
}
