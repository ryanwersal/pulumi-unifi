package provider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/blang/semver"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/integration"
)

// TestProviderSchemaBuilds is the provider-wide regression guard: it constructs
// the inferred provider and asks it for its schema. Because the schema is
// derived from the Go types and their Annotate methods, this fails the build if
// any resource has a malformed Annotate (e.g. describing a field on the marker
// struct), an unsupported field type, a duplicate token, or a registration
// mistake — the whole class of error that otherwise only surfaces at
// `pulumi package get-schema` time.
func TestProviderSchemaBuilds(t *testing.T) {
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
	resp, err := server.GetSchema(p.GetSchemaRequest{})
	if err != nil {
		t.Fatalf("GetSchema: %v", err)
	}
	if resp.Schema == "" {
		t.Fatal("empty schema")
	}

	type propSet map[string]struct {
		Secret           bool   `json:"secret"`
		Description      string `json:"description"`
		ReplaceOnChanges bool   `json:"replaceOnChanges"`
	}
	var schema struct {
		Resources map[string]struct {
			Description     string  `json:"description"`
			InputProperties propSet `json:"inputProperties"`
			Properties      propSet `json:"properties"`
		} `json:"resources"`
		Types map[string]struct {
			Properties propSet `json:"properties"`
		} `json:"types"`
	}
	if err := json.Unmarshal([]byte(resp.Schema), &schema); err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}

	// Every registered resource must be present in the schema.
	want := []string{
		"unifi:network:Vlan",
		"unifi:network:Wlan",
		"unifi:network:Device",
		"unifi:network:PortProfile",
		"unifi:network:PortForward",
		"unifi:network:FirewallGroup",
		"unifi:network:FirewallRule",
		"unifi:network:FirewallZonePolicy",
		"unifi:network:StaticRoute",
		"unifi:network:User",
		"unifi:network:UserGroup",
		"unifi:network:DnsRecord",
		"unifi:protect:Camera",
		"unifi:protect:AlarmAutomation",
	}
	for _, tok := range want {
		if _, ok := schema.Resources[tok]; !ok {
			t.Errorf("resource %q missing from schema", tok)
		}
	}

	// Guard against a silent regression in field coverage. The wide resources now
	// group their fields into nested facet objects, so the top-level surface is
	// small; assert the facet groups are present rather than a flat field count.
	facetGroups := map[string][]string{
		"unifi:network:Vlan":               {"dhcp", "dhcpV6", "ipv6", "igmp", "wan", "nat"},
		"unifi:network:Wlan":               {"wpa", "wpa3", "sae", "radius", "minRate", "schedule"},
		"unifi:network:PortProfile":        {"vlan", "link", "stormControl", "dot1x", "priorityQueues"},
		"unifi:network:Device":             {"led", "snmp", "switching", "outlet", "lcm"},
		"unifi:network:FirewallRule":       {"protocolMatch", "source", "destination", "connectionState"},
		"unifi:network:FirewallZonePolicy": {"matching"},
	}
	for tok, groups := range facetGroups {
		inputs := schema.Resources[tok].InputProperties
		for _, group := range groups {
			if _, ok := inputs[group]; !ok {
				t.Errorf("%s is missing the %q facet group input", tok, group)
			}
		}
	}
	// The nested facet types must keep their broad field coverage.
	typeFloors := map[string]int{
		"unifi:network:VlanWan":                 30,
		"unifi:network:VlanDhcp":                25,
		"unifi:network:PortProfileStormControl": 10,
	}
	for tok, floor := range typeFloors {
		if n := len(schema.Types[tok].Properties); n < floor {
			t.Errorf("type %s exposes only %d fields; expected >=%d", tok, n, floor)
		}
	}

	// Credentials must be marked secret so they are encrypted in state. The Vlan
	// PPPoE password now lives on the nested VlanWan type.
	resourceSecrets := map[string]string{
		"unifi:network:Wlan": "passphrase",
	}
	for tok, field := range resourceSecrets {
		prop, ok := schema.Resources[tok].InputProperties[field]
		if !ok {
			t.Errorf("%s.%s missing", tok, field)
			continue
		}
		if !prop.Secret {
			t.Errorf("%s.%s is not marked secret", tok, field)
		}
	}
	typeSecrets := map[string]string{
		"unifi:network:VlanWan":        "password", // PPPoE password
		"unifi:network:WlanWep":        "key",      // WEP key (moved into nested group)
		"unifi:network:WlanRoaming":    "iappKey",  // IAPP key (moved into nested group)
		"unifi:network:WlanPrivatePsk": "password", // PPSK
		"unifi:network:WlanSaePsk":     "psk",      // WPA3 SAE PSK
	}
	for tok, field := range typeSecrets {
		prop, ok := schema.Types[tok].Properties[field]
		if !ok {
			t.Errorf("type %s.%s missing", tok, field)
			continue
		}
		if !prop.Secret {
			t.Errorf("type %s.%s is not marked secret", tok, field)
		}
	}

	// Every consumer-facing property must carry a description, else the
	// generated SDK and Registry docs ship blank.
	for tok, r := range schema.Resources {
		if r.Description == "" {
			t.Errorf("resource %s has no description", tok)
		}
		for name, p := range r.InputProperties {
			if p.Description == "" {
				t.Errorf("input %s.%s has no description", tok, name)
			}
		}
		for name, p := range r.Properties {
			if p.Description == "" {
				t.Errorf("output %s.%s has no description", tok, name)
			}
		}
	}
	for tok, ty := range schema.Types {
		for name, p := range ty.Properties {
			if p.Description == "" {
				t.Errorf("type %s.%s has no description", tok, name)
			}
		}
	}

	// Identity/bind-key fields must force replacement, not an in-place update
	// that would patch the wrong object.
	replaceOnChanges := map[string]string{
		"unifi:network:Device": "mac",
		"unifi:protect:Camera": "cameraId",
		"unifi:network:User":   "mac",
	}
	for tok, field := range replaceOnChanges {
		prop, ok := schema.Resources[tok].InputProperties[field]
		if !ok {
			t.Errorf("%s.%s missing", tok, field)
			continue
		}
		if !prop.ReplaceOnChanges {
			t.Errorf("%s.%s is not marked replaceOnChanges", tok, field)
		}
	}
}
