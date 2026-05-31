// Package provider wires the inferred UniFi provider: its configuration and the
// set of resources it manages. Resource implementations live in subpackages
// (network/, protect/); this package only assembles them.
package provider

import (
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/ryanwersal/pulumi-unifi/provider/config"
	"github.com/ryanwersal/pulumi-unifi/provider/network"
	"github.com/ryanwersal/pulumi-unifi/provider/protect"
)

// Name is the Pulumi package/plugin name. It must match the
// `pulumi-resource-<Name>` binary and the generated SDK package.
const Name = "unifi"

// Version is the provider version, stamped into the schema and SDK. Keep in
// sync with git tags once releases are cut.
const Version = "0.1.0"

// New builds the inferred provider. The infer layer derives the Pulumi schema
// and gRPC server from the Go types referenced here.
func New() (p.Provider, error) {
	return infer.NewProviderBuilder().
		WithDisplayName("UniFi").
		WithDescription("Manage a UniFi Dream Machine's Network and Protect applications via the local controller API.").
		WithHomepage("https://github.com/ryanwersal/pulumi-unifi").
		WithRepository("https://github.com/ryanwersal/pulumi-unifi").
		WithLicense("Apache-2.0").
		WithConfig(infer.Config(config.Config{})).
		WithResources(
			infer.Resource(network.Vlan{}),
			infer.Resource(network.Wlan{}),
			infer.Resource(protect.Camera{}),
		).
		Build()
}
