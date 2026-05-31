// Package config defines the provider-level configuration and builds the
// authenticated UniFi controller client once per provider process. Resources
// retrieve the configured client via infer.GetConfig.
package config

import (
	"context"
	"fmt"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/pulumi/pulumi-go-provider/infer"
)

// Config is the provider configuration. Secret fields are marked
// `provider:"secret"` so they are encrypted in state and redacted in output.
//
// Authentication is either an API key (preferred for headless use) OR a
// username/password pair — not both. Prefer a local-only API key generated on
// the UniFi OS console.
type Config struct {
	// URL is the base URL of the controller, e.g. https://192.168.1.1 (no /api suffix).
	URL string `pulumi:"url"`
	// APIKey authenticates via the UniFi OS local API key (X-API-Key).
	APIKey *string `pulumi:"apiKey,optional" provider:"secret"`
	// Username for username/password auth (alternative to APIKey).
	Username *string `pulumi:"username,optional"`
	// Password for username/password auth (alternative to APIKey).
	Password *string `pulumi:"password,optional" provider:"secret"`
	// Site is the UniFi site name. Defaults to "default".
	Site *string `pulumi:"site,optional"`
	// InsecureTLS skips TLS verification (self-signed controller certs).
	InsecureTLS *bool `pulumi:"insecureTls,optional"`

	// net is the configured client, populated by Configure. It is not a Pulumi
	// field (no struct tag) and never appears in state.
	net unifi.Client
}

// Annotate attaches human-readable descriptions to the config fields.
func (c *Config) Annotate(a infer.Annotator) {
	a.Describe(&c.URL, "Base URL of the UniFi controller, e.g. https://192.168.1.1 (omit any /api suffix).")
	a.Describe(&c.APIKey, "UniFi OS local API key. Preferred for headless automation. Mutually exclusive with username/password.")
	a.Describe(&c.Username, "Local admin username. Use with password when not using an API key.")
	a.Describe(&c.Password, "Local admin password.")
	a.Describe(&c.Site, `UniFi site name (defaults to "default").`)
	a.Describe(&c.InsecureTLS, "Skip TLS certificate verification (for self-signed controller certs).")
}

// Configure builds the authenticated UniFi client. Called once per provider
// process, after the receiver has been hydrated from inputs.
func (c *Config) Configure(_ context.Context) error {
	cc := &unifi.ClientConfig{URL: c.URL}
	cc.VerifySSL = !(c.InsecureTLS != nil && *c.InsecureTLS)

	switch {
	case c.APIKey != nil && *c.APIKey != "":
		cc.APIKey = *c.APIKey
	case c.Username != nil && *c.Username != "" && c.Password != nil:
		cc.User = *c.Username
		cc.Password = *c.Password
	default:
		return fmt.Errorf("unifi provider: set either `apiKey` or both `username` and `password`")
	}

	client, err := unifi.NewClient(cc)
	if err != nil {
		return fmt.Errorf("unifi provider: failed to create client for %s: %w", c.URL, err)
	}
	c.net = client
	return nil
}

// Network returns the configured UniFi Network client.
func (c Config) Network() unifi.Client { return c.net }

// ResolvedSite returns the configured site, defaulting to "default".
func (c Config) ResolvedSite() string {
	if c.Site != nil && *c.Site != "" {
		return *c.Site
	}
	return "default"
}
