// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	protecttypes "github.com/ClifHouck/unified/types"
	"github.com/blang/semver"
	"github.com/filipowm/go-unifi/unifi"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/integration"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/ryanwersal/pulumi-unifi/provider/config"
)

// newLifecycleServer builds the real provider with the given fake clients
// injected, configured against a dummy controller, ready for LifeCycleTest.
func newLifecycleServer(t *testing.T, net unifi.Client, protect protecttypes.ProtectV1) integration.Server {
	t.Helper()
	t.Cleanup(config.InjectClientsForTest(net, protect))
	prov, err := New()
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	server, err := integration.NewServer(
		context.Background(), Name, semver.MustParse("0.1.0"),
		integration.WithProvider(prov),
	)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	if err := server.Configure(p.ConfigureRequest{Args: property.NewMap(map[string]property.Value{
		"url":    property.New("https://unifi.test"),
		"apiKey": property.New("test-key"),
	})}); err != nil {
		t.Fatalf("Configure: %v", err)
	}
	return server
}

func pmap(kv map[string]property.Value) property.Map { return property.NewMap(kv) }

func parray(vals ...string) property.Value {
	out := make([]property.Value, len(vals))
	for i, v := range vals {
		out[i] = property.New(v)
	}
	return property.New(property.NewArray(out))
}

// eq asserts a string-valued output property.
func eq(t *testing.T, out property.Map, key, want string) {
	t.Helper()
	got := out.Get(key)
	if !got.IsString() || got.AsString() != want {
		t.Errorf("output %q = %v, want %q", key, got, want)
	}
}

func notEmpty(t *testing.T, out property.Map, key string) {
	t.Helper()
	got := out.Get(key)
	if !got.IsString() || got.AsString() == "" {
		t.Errorf("output %q is empty, want a value", key)
	}
}

// --- fake go-unifi Network client (only the methods the tested resources use) ---

type fakeNetwork struct {
	unifi.Client // embedded: unimplemented methods panic if ever called
	seq          int
	firewallGrp  map[string]*unifi.FirewallGroup
	dnsRecords   map[string]*unifi.DNSRecord
	userGroups   map[string]*unifi.UserGroup
	// cameras backs the raw Protect camera PATCH (Camera writes go through the
	// go-unifi client's Do, not the typed Protect client).
	cameras map[string]*protecttypes.Camera
}

func newFakeNetwork() *fakeNetwork {
	return &fakeNetwork{
		firewallGrp: map[string]*unifi.FirewallGroup{},
		dnsRecords:  map[string]*unifi.DNSRecord{},
		userGroups:  map[string]*unifi.UserGroup{},
	}
}

func (f *fakeNetwork) id(prefix string) string { f.seq++; return fmt.Sprintf("%s-%d", prefix, f.seq) }

// Do handles the raw Protect camera PATCH; it applies the map body (including
// explicit false toggles) to the shared camera and echoes it back.
func (f *fakeNetwork) Do(_ context.Context, method, apiPath string, reqBody, respBody any) error {
	if method != http.MethodPatch || !strings.Contains(apiPath, "/cameras/") {
		panic("fakeNetwork.Do: unexpected " + method + " " + apiPath)
	}
	id := apiPath[strings.LastIndex(apiPath, "/")+1:]
	cam, ok := f.cameras[id]
	if !ok {
		return unifi.ErrNotFound
	}
	body, _ := reqBody.(map[string]any)
	if v, ok := body["name"].(string); ok {
		cam.Name = v
	}
	if led, ok := body["ledSettings"].(map[string]any); ok {
		if v, ok := led["isEnabled"].(bool); ok {
			cam.LedSettings.IsEnabled = v
		}
	}
	if osd, ok := body["osdSettings"].(map[string]any); ok {
		if v, ok := osd["isNameEnabled"].(bool); ok {
			cam.OsdSettings.IsNameEnabled = v
		}
	}
	if out, ok := respBody.(*protecttypes.Camera); ok {
		*out = *cam
	}
	return nil
}

func (f *fakeNetwork) CreateFirewallGroup(_ context.Context, _ string, d *unifi.FirewallGroup) (*unifi.FirewallGroup, error) {
	cp := *d
	cp.ID = f.id("fg")
	f.firewallGrp[cp.ID] = &cp
	out := cp
	return &out, nil
}

func (f *fakeNetwork) GetFirewallGroup(_ context.Context, _, id string) (*unifi.FirewallGroup, error) {
	g, ok := f.firewallGrp[id]
	if !ok {
		return nil, unifi.ErrNotFound
	}
	out := *g
	return &out, nil
}

func (f *fakeNetwork) UpdateFirewallGroup(_ context.Context, _ string, d *unifi.FirewallGroup) (*unifi.FirewallGroup, error) {
	if _, ok := f.firewallGrp[d.ID]; !ok {
		return nil, unifi.ErrNotFound
	}
	cp := *d
	f.firewallGrp[cp.ID] = &cp
	out := cp
	return &out, nil
}

func (f *fakeNetwork) DeleteFirewallGroup(_ context.Context, _, id string) error {
	delete(f.firewallGrp, id)
	return nil
}

func (f *fakeNetwork) CreateDNSRecord(_ context.Context, _ string, d *unifi.DNSRecord) (*unifi.DNSRecord, error) {
	cp := *d
	cp.ID = f.id("dns")
	f.dnsRecords[cp.ID] = &cp
	out := cp
	return &out, nil
}

func (f *fakeNetwork) GetDNSRecord(_ context.Context, _, id string) (*unifi.DNSRecord, error) {
	r, ok := f.dnsRecords[id]
	if !ok {
		return nil, unifi.ErrNotFound
	}
	out := *r
	return &out, nil
}

func (f *fakeNetwork) UpdateDNSRecord(_ context.Context, _ string, d *unifi.DNSRecord) (*unifi.DNSRecord, error) {
	if _, ok := f.dnsRecords[d.ID]; !ok {
		return nil, unifi.ErrNotFound
	}
	cp := *d
	f.dnsRecords[cp.ID] = &cp
	out := cp
	return &out, nil
}

func (f *fakeNetwork) DeleteDNSRecord(_ context.Context, _, id string) error {
	delete(f.dnsRecords, id)
	return nil
}

func (f *fakeNetwork) CreateUserGroup(_ context.Context, _ string, d *unifi.UserGroup) (*unifi.UserGroup, error) {
	cp := *d
	cp.ID = f.id("ug")
	f.userGroups[cp.ID] = &cp
	out := cp
	return &out, nil
}

func (f *fakeNetwork) GetUserGroup(_ context.Context, _, id string) (*unifi.UserGroup, error) {
	g, ok := f.userGroups[id]
	if !ok {
		return nil, unifi.ErrNotFound
	}
	out := *g
	return &out, nil
}

func (f *fakeNetwork) UpdateUserGroup(_ context.Context, _ string, d *unifi.UserGroup) (*unifi.UserGroup, error) {
	if _, ok := f.userGroups[d.ID]; !ok {
		return nil, unifi.ErrNotFound
	}
	cp := *d
	f.userGroups[cp.ID] = &cp
	out := cp
	return &out, nil
}

func (f *fakeNetwork) DeleteUserGroup(_ context.Context, _, id string) error {
	delete(f.userGroups, id)
	return nil
}

// --- fake Protect client ---

type fakeProtect struct {
	protecttypes.ProtectV1 // embedded
	cameras                map[string]*protecttypes.Camera
}

func (f *fakeProtect) CameraDetails(id protecttypes.CameraID) (*protecttypes.Camera, error) {
	cam, ok := f.cameras[string(id)]
	if !ok {
		return nil, fmt.Errorf("got unexpected http code 404 for camera %s", id)
	}
	out := *cam
	return &out, nil
}

// --- lifecycle tests ---

func TestFirewallGroupLifecycle(t *testing.T) {
	server := newLifecycleServer(t, newFakeNetwork(), nil)
	integration.LifeCycleTest{
		Resource: "unifi:network:FirewallGroup",
		Create: integration.Operation{
			Inputs: pmap(map[string]property.Value{
				"name":         property.New("web"),
				"groupMembers": parray("10.0.0.0/24"),
			}),
			Hook: func(_, out property.Map) {
				eq(t, out, "name", "web")
				eq(t, out, "groupType", "address-group") // SetDefault applied via Check
				notEmpty(t, out, "firewallGroupId")
			},
		},
		Updates: []integration.Operation{{
			Inputs: pmap(map[string]property.Value{
				"name":         property.New("web"),
				"groupType":    property.New("port-group"),
				"groupMembers": parray("80", "443"),
			}),
			Hook: func(_, out property.Map) {
				eq(t, out, "groupType", "port-group")
			},
		}},
	}.Run(t, server)
}

func TestDnsRecordLifecycle(t *testing.T) {
	server := newLifecycleServer(t, newFakeNetwork(), nil)
	integration.LifeCycleTest{
		Resource: "unifi:network:DnsRecord",
		Create: integration.Operation{
			Inputs: pmap(map[string]property.Value{
				"key":        property.New("host.example.com"),
				"recordType": property.New("A"),
				"value":      property.New("10.0.0.5"),
			}),
			Hook: func(_, out property.Map) {
				eq(t, out, "recordType", "A")
				eq(t, out, "value", "10.0.0.5")
				notEmpty(t, out, "dnsRecordId")
			},
		},
		Updates: []integration.Operation{{
			Inputs: pmap(map[string]property.Value{
				"key":        property.New("host.example.com"),
				"recordType": property.New("A"),
				"value":      property.New("10.0.0.6"),
			}),
			Hook: func(_, out property.Map) {
				eq(t, out, "value", "10.0.0.6")
			},
		}},
	}.Run(t, server)
}

func TestUserGroupLifecycle(t *testing.T) {
	server := newLifecycleServer(t, newFakeNetwork(), nil)
	integration.LifeCycleTest{
		Resource: "unifi:network:UserGroup",
		Create: integration.Operation{
			Inputs: pmap(map[string]property.Value{
				"name": property.New("limited"),
			}),
			Hook: func(_, out property.Map) {
				eq(t, out, "name", "limited")
				notEmpty(t, out, "userGroupId")
				if v := out.Get("qosRateMaxDown"); !v.IsNumber() || v.AsNumber() != -1 {
					t.Errorf("qosRateMaxDown = %v, want -1 (default)", v)
				}
			},
		},
		Updates: []integration.Operation{{
			Inputs: pmap(map[string]property.Value{
				"name":           property.New("limited"),
				"qosRateMaxDown": property.New(1000.0),
			}),
			Hook: func(_, out property.Map) {
				if v := out.Get("qosRateMaxDown"); !v.IsNumber() || v.AsNumber() != 1000 {
					t.Errorf("qosRateMaxDown = %v, want 1000", v)
				}
			},
		}},
	}.Run(t, server)
}

func TestCameraLifecycle(t *testing.T) {
	cam := &protecttypes.Camera{ID: "cam-1", Name: "old", ModelKey: "UVC-G4", State: "CONNECTED"}
	cam.LedSettings.IsEnabled = true
	cam.OsdSettings.IsNameEnabled = true
	cameras := map[string]*protecttypes.Camera{"cam-1": cam}
	fp := &fakeProtect{cameras: cameras}
	fn := newFakeNetwork()
	fn.cameras = cameras // the Protect read fake and the raw-PATCH fake share state
	server := newLifecycleServer(t, fn, fp)
	integration.LifeCycleTest{
		Resource: "unifi:protect:Camera",
		Create: integration.Operation{
			Inputs: pmap(map[string]property.Value{
				"cameraId":       property.New("cam-1"),
				"name":           property.New("Front Door"),
				"ledEnabled":     property.New(false), // turning a toggle OFF must land (the bug fix)
				"osdNameEnabled": property.New(false),
			}),
			Hook: func(_, out property.Map) {
				eq(t, out, "cameraId", "cam-1")
				eq(t, out, "name", "Front Door")
				eq(t, out, "type", "UVC-G4")
				eq(t, out, "state", "CONNECTED")
				if v := out.Get("ledEnabled"); !v.IsBool() || v.AsBool() {
					t.Errorf("ledEnabled = %v, want false (off toggle must transmit)", v)
				}
				if v := out.Get("osdNameEnabled"); !v.IsBool() || v.AsBool() {
					t.Errorf("osdNameEnabled = %v, want false", v)
				}
			},
		},
		Updates: []integration.Operation{{
			Inputs: pmap(map[string]property.Value{
				"cameraId": property.New("cam-1"),
				"name":     property.New("Back Door"),
			}),
			Hook: func(_, out property.Map) {
				eq(t, out, "name", "Back Door")
			},
		}},
	}.Run(t, server)
}
