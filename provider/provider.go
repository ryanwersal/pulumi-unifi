// SPDX-License-Identifier: Apache-2.0

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

// Version is a var (not a const) so release builds can stamp the git tag via
// -ldflags; the default is the local-dev version. respectSchemaVersion (below)
// carries it into the published npm package version too.
var Version = "0.1.0"

// pluginDownloadURL points Pulumi at this repo's GitHub Releases to fetch the
// plugin binary when a consuming program has no local install.
const pluginDownloadURL = "github://api.github.com/ryanwersal/pulumi-unifi"

// npmPackageName moves the SDK off the @pulumi scope (which we don't own). The
// bare @ryanwersal scope carries no "pulumi" signal, so keep it in the name.
const npmPackageName = "@ryanwersal/pulumi-unifi"

// New builds the inferred provider. The infer layer derives the Pulumi schema
// and gRPC server from the Go types referenced here.
func New() (p.Provider, error) {
	return infer.NewProviderBuilder().
		WithDisplayName("UniFi").
		WithDescription("Manage a UniFi Dream Machine's Network and Protect applications via the local controller API.").
		WithHomepage("https://github.com/ryanwersal/pulumi-unifi").
		WithRepository("https://github.com/ryanwersal/pulumi-unifi").
		WithPublisher("ryanwersal").
		WithLicense("Apache-2.0").
		WithPluginDownloadURL(pluginDownloadURL).
		WithLanguageMap(map[string]any{
			"nodejs": map[string]any{
				"respectSchemaVersion": true,
				"packageName":          npmPackageName,
			},
		}).
		WithConfig(infer.Config(&config.Config{})).
		WithResources(
			infer.Resource(network.Vlan{}),
			infer.Resource(network.Wlan{}),
			infer.Resource(network.Device{}),
			infer.Resource(network.PortProfile{}),
			infer.Resource(network.PortForward{}),
			infer.Resource(network.FirewallGroup{}),
			infer.Resource(network.FirewallRule{}),
			infer.Resource(network.FirewallZonePolicy{}),
			infer.Resource(network.StaticRoute{}),
			infer.Resource(network.User{}),
			infer.Resource(network.UserGroup{}),
			infer.Resource(network.DnsRecord{}),
			infer.Resource(protect.Camera{}),
			infer.Resource(protect.AlarmAutomation{}),
		).
		Build()
}
