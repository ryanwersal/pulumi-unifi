// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/filipowm/go-unifi/unifi"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/integration"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
)

type fakeStreams struct {
	unifi.Client
	live        map[string]*string
	calls       []string
	fail        error
	ignoreWrite bool
}

func (f *fakeStreams) Do(_ context.Context, method, path string, body, response any) error {
	f.calls = append(f.calls, method+" "+path)
	if f.fail != nil {
		return f.fail
	}
	u, err := url.Parse(path)
	if err != nil {
		return err
	}
	if u.Path != "/proxy/protect/integration/v1/cameras/cam-123/rtsps-stream" {
		return fmt.Errorf("unexpected path %s", path)
	}
	switch method {
	case http.MethodGet:
		data, err := json.Marshal(f.live)
		if err != nil {
			return err
		}
		return json.Unmarshal(data, response)
	case http.MethodPost:
		wire, _ := json.Marshal(body)
		if string(wire) != `{"qualities":["high"]}` {
			return fmt.Errorf("unexpected POST body %s", wire)
		}
		if !f.ignoreWrite {
			v := "rtsps://private-video-token"
			f.live["high"] = &v
		}
	case http.MethodDelete:
		if u.RawQuery != "qualities=high" {
			return fmt.Errorf("unsafe delete query: %s", u.RawQuery)
		}
		delete(f.live, "high")
	default:
		return fmt.Errorf("unexpected method %s", method)
	}
	return nil
}

func streamInputs() property.Map {
	return pmap(map[string]property.Value{"cameraId": property.New("cam-123"), "quality": property.New("high")})
}

const streamURN = resource.URN("urn:pulumi:test::test::unifi:protect:CameraStream::test")

func TestCameraStreamLifecycle(t *testing.T) {
	medium := "rtsps://existing-medium"
	client := &fakeStreams{live: map[string]*string{"medium": &medium, "high": nil}}
	server := newLifecycleServer(t, client, &fakeProtect{})
	integration.LifeCycleTest{
		Resource: "unifi:protect:CameraStream",
		Create: integration.Operation{Inputs: streamInputs(), Hook: func(_, out property.Map) {
			eq(t, out, "cameraId", "cam-123")
			eq(t, out, "quality", "high")
			if out.Len() != 2 {
				t.Error("stream URLs must not be exported")
			}
		}},
	}.Run(t, server)
	if client.live["high"] != nil || client.live["medium"] == nil || *client.live["medium"] != medium {
		t.Fatal("delete changed an unmanaged stream or failed to remove the owned quality")
	}
}

func TestCameraStreamReadAndDrift(t *testing.T) {
	high := "rtsps://existing-high"
	client := &fakeStreams{live: map[string]*string{"high": &high}}
	server := newLifecycleServer(t, client, &fakeProtect{})
	// Create adopts an existing stream without a POST, preserving the URL.
	created, err := server.Create(p.CreateRequest{Urn: streamURN, Properties: streamInputs()})
	if err != nil {
		t.Fatal(err)
	}
	if len(client.calls) != 1 || !strings.HasPrefix(client.calls[0], "GET ") {
		t.Fatal("create regenerated an existing stream")
	}
	// Import reconstructs both inputs from the compound ID alone.
	read, err := server.Read(p.ReadRequest{Urn: streamURN, ID: "cam-123/high"})
	if err != nil {
		t.Fatal(err)
	}
	eq(t, read.Inputs, "quality", "high")
	eq(t, read.Inputs, "cameraId", "cam-123")
	delete(client.live, "high")
	read, err = server.Read(p.ReadRequest{Urn: streamURN, ID: created.ID, Properties: created.Properties, Inputs: streamInputs()})
	if err != nil || read.ID != "" {
		t.Fatalf("disabled stream should be reported missing: %+v %v", read, err)
	}
	// A subsequent create restores the missing quality.
	if _, err := server.Create(p.CreateRequest{Urn: streamURN, Properties: streamInputs()}); err != nil {
		t.Fatal(err)
	}
	if client.live["high"] == nil {
		t.Fatal("drift was not corrected")
	}
	client.fail = &unifi.ServerError{StatusCode: http.StatusNotFound}
	read, err = server.Read(p.ReadRequest{Urn: streamURN, ID: created.ID, Properties: created.Properties, Inputs: streamInputs()})
	if err != nil || read.ID != "" {
		t.Fatalf("missing camera should be reported missing: %+v %v", read, err)
	}
	if err := server.Delete(p.DeleteRequest{Urn: streamURN, ID: created.ID, Properties: created.Properties}); err != nil {
		t.Fatal(err)
	}
}

func TestCameraStreamFailuresAndPreview(t *testing.T) {
	client := &fakeStreams{live: map[string]*string{}, ignoreWrite: true}
	server := newLifecycleServer(t, client, &fakeProtect{})
	if _, err := server.Create(p.CreateRequest{Urn: streamURN, Properties: streamInputs(), DryRun: true}); err != nil {
		t.Fatal(err)
	}
	if len(client.calls) != 0 {
		t.Fatal("preview contacted Protect")
	}
	if _, err := server.Create(p.CreateRequest{Urn: streamURN, Properties: streamInputs()}); err == nil {
		t.Fatal("unverified enable was accepted")
	}
	client.fail = fmt.Errorf("403 Forbidden")
	if _, err := server.Read(p.ReadRequest{Urn: streamURN, ID: "cam-123/high"}); err == nil {
		t.Fatal("permission failure was treated as missing")
	}
	if _, err := server.Create(p.CreateRequest{Urn: streamURN, Properties: streamInputs()}); err == nil {
		t.Fatal("permission failure was ignored")
	}
	if err := server.Delete(p.DeleteRequest{Urn: streamURN, ID: "cam-123/high", Properties: streamInputs()}); err == nil {
		t.Fatal("delete permission failure was ignored")
	}
	for _, id := range []string{"", "cam-123", "../high", "cam-123/bad", "cam-123/high/extra"} {
		if _, err := server.Read(p.ReadRequest{Urn: streamURN, ID: id}); err == nil {
			t.Errorf("invalid import ID %q accepted", id)
		}
	}
}

func TestCameraStreamReplacement(t *testing.T) {
	server := newLifecycleServer(t, &fakeStreams{}, &fakeProtect{})
	for _, field := range []string{"cameraId", "quality"} {
		inputs := streamInputs().Set(field, property.New("medium"))
		diff, err := server.Diff(p.DiffRequest{Urn: streamURN, ID: "cam-123/high", State: streamInputs(), OldInputs: streamInputs(), Inputs: inputs})
		if err != nil {
			t.Fatal(err)
		}
		if !diff.HasChanges || diff.DetailedDiff[field].Kind != p.UpdateReplace {
			t.Fatalf("%s must replace: %+v", field, diff)
		}
	}
}
