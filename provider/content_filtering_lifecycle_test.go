// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/filipowm/go-unifi/unifi"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/integration"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
)

const filterURN = resource.URN("urn:pulumi:test::test::unifi:network:ContentFiltering::test")

type fakeContentFiltering struct {
	unifi.Client
	live  map[string]any
	calls []string
	fail  error
}

func (f *fakeContentFiltering) Do(_ context.Context, method, path string, body, response any) error {
	f.calls = append(f.calls, method+" "+path)
	if f.fail != nil {
		return f.fail
	}
	expected := "../v2/api/site/default/content-filtering"
	if method != http.MethodPost && method != http.MethodGet {
		expected += "/policy-123"
	}
	if path != expected {
		return fmt.Errorf("unexpected path %s", path)
	}
	if method != http.MethodPost && method != http.MethodGet && f.live == nil {
		return &unifi.ServerError{StatusCode: 404}
	}
	switch method {
	case http.MethodGet:
		if body != nil {
			return fmt.Errorf("GET must have no body")
		}
	case http.MethodPost:
		f.live = body.(map[string]any)
		f.live["_id"] = "policy-123"
		f.live["future_security_setting"] = true
	case http.MethodPut:
		f.live = body.(map[string]any)
	case http.MethodDelete:
		f.live = nil
		return nil
	default:
		return fmt.Errorf("unexpected method %s", method)
	}
	var payload any = f.live
	if method == http.MethodGet {
		policies := []map[string]any{}
		if f.live != nil {
			policies = append(policies, f.live)
		}
		payload = policies
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, response)
}
func filterInputs(allow ...string) property.Map {
	return pmap(map[string]property.Value{
		"name": property.New("Default"), "enabled": property.New(true),
		"networkIds": parray("lan-id"), "clientMacs": parray(),
		"categories": parray("FAMILY", "ADVERTISEMENT"), "allowList": parray(allow...),
		"blockList": parray(), "safeSearch": parray(),
		"schedule": property.New(pmap(map[string]property.Value{"mode": property.New("ALWAYS")})),
	})
}
func liveFilter() map[string]any {
	return map[string]any{
		"_id": "policy-123", "name": "Default", "enabled": true,
		"network_ids": []any{"lan-id"}, "client_macs": []any{},
		"categories": []any{"FAMILY", "ADVERTISEMENT"}, "allow_list": []any{}, "block_list": []any{}, "safe_search": []any{},
		"schedule":                map[string]any{"mode": "ALWAYS", "future_timezone": "America/Phoenix"},
		"future_security_setting": true,
	}
}
func TestContentFilteringLifecycle(t *testing.T) {
	client := &fakeContentFiltering{}
	server := newLifecycleServer(t, client, nil)
	integration.LifeCycleTest{
		Resource: "unifi:network:ContentFiltering",
		Create:   integration.Operation{Inputs: filterInputs()},
		Updates: []integration.Operation{
			{Inputs: filterInputs("pinggy-free.link"), Hook: func(_, out property.Map) {
				eq(t, out, "policyId", "policy-123")
				if client.live["future_security_setting"] != true {
					t.Fatal("lost unmanaged security setting")
				}
			}},
			{Inputs: filterInputs()},
		},
	}.Run(t, server)
	if client.live != nil {
		t.Fatal("delete did not remove the policy")
	}
}
func TestContentFilteringImportAndAllowlistOnlyUpdate(t *testing.T) {
	client := &fakeContentFiltering{live: liveFilter()}
	server := newLifecycleServer(t, client, nil)
	read, err := server.Read(p.ReadRequest{Urn: filterURN, ID: "policy-123"})
	if err != nil {
		t.Fatal(err)
	}
	eq(t, read.Inputs, "name", "Default")
	// Concurrent edits to unmodified fields must survive the allowlist update.
	client.live["name"] = "Renamed in UI"
	want := liveFilter()
	want["name"] = "Renamed in UI"
	want["allow_list"] = []any{"pinggy-free.link"}
	out, err := server.Update(p.UpdateRequest{Urn: filterURN, ID: read.ID, State: read.Properties, Inputs: filterInputs("pinggy-free.link")})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(client.live, want) {
		t.Fatalf("changed fields outside allowlist:\ngot %#v\nwant %#v", client.live, want)
	}
	eq(t, out.Properties, "name", "Renamed in UI")
	// Refresh must surface out-of-band drift; a subsequent update restores it.
	client.live["allow_list"] = []any{"unexpected.example"}
	drift, err := server.Read(p.ReadRequest{Urn: filterURN, ID: read.ID, Properties: out.Properties, Inputs: filterInputs("pinggy-free.link")})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(drift.Inputs.Get("allowList"), parray("unexpected.example")) {
		t.Fatal("refresh hid allowlist drift")
	}
	_, err = server.Update(p.UpdateRequest{Urn: filterURN, ID: read.ID, State: drift.Properties, Inputs: filterInputs()})
	if err != nil {
		t.Fatal(err)
	}
	if len(client.live["allow_list"].([]any)) != 0 {
		t.Fatal("explicit empty allowlist did not clear entries")
	}
}
func TestContentFilteringPreviewMissingAndFailures(t *testing.T) {
	client := &fakeContentFiltering{live: liveFilter()}
	server := newLifecycleServer(t, client, nil)
	if _, err := server.Create(p.CreateRequest{Urn: filterURN, Properties: filterInputs(), DryRun: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := server.Update(p.UpdateRequest{Urn: filterURN, ID: "policy-123", State: filterInputs().Set("policyId", property.New("policy-123")), Inputs: filterInputs("pinggy-free.link"), DryRun: true}); err != nil {
		t.Fatal(err)
	}
	if len(client.calls) != 0 {
		t.Fatal("preview contacted controller")
	}
	client.fail = &unifi.ServerError{StatusCode: 403}
	if _, err := server.Read(p.ReadRequest{Urn: filterURN, ID: "policy-123"}); err == nil {
		t.Fatal("permission error treated as missing")
	}
	if _, err := server.Update(p.UpdateRequest{Urn: filterURN, ID: "policy-123", State: filterInputs().Set("policyId", property.New("policy-123")), Inputs: filterInputs("pinggy-free.link")}); err == nil {
		t.Fatal("update ignored failed read")
	}
	if err := server.Delete(p.DeleteRequest{Urn: filterURN, ID: "policy-123", Properties: filterInputs().Set("policyId", property.New("policy-123"))}); err == nil {
		t.Fatal("delete ignored permission error")
	}
	client.fail = &unifi.ServerError{StatusCode: 404}
	if _, err := server.Read(p.ReadRequest{Urn: filterURN, ID: "policy-123"}); err == nil {
		t.Fatal("missing collection endpoint treated as missing policy")
	}
	client.fail = nil
	client.live = nil
	read, err := server.Read(p.ReadRequest{Urn: filterURN, ID: "policy-123"})
	if err != nil || read.ID != "" {
		t.Fatalf("missing policy: %+v %v", read, err)
	}
	if err = server.Delete(p.DeleteRequest{Urn: filterURN, ID: "policy-123", Properties: filterInputs().Set("policyId", property.New("policy-123"))}); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"", "..", "../other", "bad/id", "bad?query"} {
		before := len(client.calls)
		if _, err := server.Read(p.ReadRequest{Urn: filterURN, ID: id}); err == nil {
			t.Errorf("accepted invalid import ID %q", id)
		}
		if len(client.calls) != before {
			t.Error("invalid ID reached controller")
		}
	}
}
func TestContentFilteringCheckScopeAndUnknowns(t *testing.T) {
	server := newLifecycleServer(t, &fakeContentFiltering{}, nil)
	for _, input := range []property.Map{
		filterInputs().Set("name", property.New(" ")),
		filterInputs().Set("networkIds", parray()),
	} {
		checked, err := server.Check(p.CheckRequest{Urn: filterURN, Inputs: input})
		if err != nil {
			t.Fatal(err)
		}
		if len(checked.Failures) == 0 {
			t.Fatal("unsafe policy accepted")
		}
	}
	computed := property.New(property.NewArray([]property.Value{property.New(property.Computed)}))
	checked, err := server.Check(p.CheckRequest{Urn: filterURN, Inputs: filterInputs().Set("networkIds", computed)})
	if err != nil || len(checked.Failures) > 0 {
		t.Fatalf("unknown network rejected: %+v %v", checked, err)
	}
}

func TestContentFilteringMinimalInputs(t *testing.T) {
	client := &fakeContentFiltering{}
	server := newLifecycleServer(t, client, nil)
	integration.LifeCycleTest{
		Resource: "unifi:network:ContentFiltering",
		Create: integration.Operation{Inputs: pmap(map[string]property.Value{"name": property.New("Minimal"), "clientMacs": parray("00:11:22:33:44:55")}), Hook: func(_, out property.Map) {
			if !out.Get("enabled").AsBool() {
				t.Fatal("policy was not enabled by default")
			}
			if client.live["schedule"].(map[string]any)["mode"] != "ALWAYS" {
				t.Fatal("schedule was not defaulted")
			}
		}},
	}.Run(t, server)
}
