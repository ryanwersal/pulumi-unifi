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

// Version is the provider version, stamped into the schema and SDK. It is a
// var (not a const) so release builds can override it via the linker:
//
//	-ldflags "-X github.com/ryanwersal/pulumi-unifi/provider.Version=<tag>"
//
// The default below is the local-dev/source version; goreleaser stamps the git
// tag at release time. Because the nodejs SDK is generated with
// respectSchemaVersion, this value also becomes the published npm package
// version.
var Version = "0.1.0"

// pluginDownloadURL tells Pulumi where to fetch the plugin binary when a
// consuming program references the SDK without a locally-installed plugin.
// The github:// scheme resolves to release assets on the repo's GitHub
// Releases (named pulumi-resource-unifi-v<ver>-<os>-<arch>.tar.gz), which the
// release workflow publishes via goreleaser.
const pluginDownloadURL = "github://api.github.com/ryanwersal/pulumi-unifi"

// npmPackageName overrides the default @pulumi/unifi name (we don't own the
// @pulumi npm scope) so the generated TypeScript SDK publishes under a scope
// we control.
const npmPackageName = "@ryanwersal/unifi"

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
			// respectSchemaVersion ties the SDK package version to the schema
			// (i.e. Version above), so a tagged release publishes a matching
			// npm version. packageName moves it off the @pulumi scope.
			"nodejs": map[string]any{
				"respectSchemaVersion": true,
				"packageName":          npmPackageName,
			},
		}).
		WithConfig(infer.Config(config.Config{})).
		WithResources(
			infer.Resource(network.Vlan{}),
			infer.Resource(network.Wlan{}),
			infer.Resource(protect.Camera{}),
		).
		Build()
}
