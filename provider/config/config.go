// SPDX-License-Identifier: Apache-2.0

// Package config defines the provider-level configuration and builds the
// authenticated UniFi controller client once per provider process. Resources
// retrieve the configured client via infer.GetConfig.
package config

import (
	"context"
	"fmt"
	"io"
	"net/url"

	unifiprotect "github.com/ClifHouck/unified/client"
	protecttypes "github.com/ClifHouck/unified/types"
	"github.com/filipowm/go-unifi/unifi"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/sirupsen/logrus"

	"github.com/ryanwersal/pulumi-unifi/provider/internal/driveapi"
)

// Config is the provider configuration. Secret fields are marked
// `provider:"secret"` so they are encrypted in state and redacted in output.
//
// Authentication is either an API key (preferred for headless use) OR a
// username/password pair — not both. Prefer a local-only API key generated on
// the UniFi OS console.
type Config struct {
	// URL is the base URL of the controller, e.g. https://192.168.1.1 (no /api suffix).
	URL string `pulumi:"url,optional"`
	// APIKey authenticates via the UniFi OS local API key (X-API-Key).
	APIKey *string `pulumi:"apiKey,optional" provider:"secret"`
	// Username for username/password auth (alternative to APIKey).
	Username *string `pulumi:"username,optional"`
	// Password for username/password auth (alternative to APIKey).
	Password *string `pulumi:"password,optional" provider:"secret"`
	// Site is the UniFi site name. Defaults to "default".
	Site *string `pulumi:"site,optional"`
	// InsecureTLS skips TLS verification (self-signed controller certs). Applies
	// to both the controller and the UNAS appliance.
	InsecureTLS *bool `pulumi:"insecureTls,optional"`

	// UnasURL is the base URL of the UNAS appliance that hosts UniFi Drive, e.g.
	// https://192.168.1.20. This is a SEPARATE host from `url`: the UNAS is its
	// own UniFi OS console, not reachable through the main controller.
	UnasURL *string `pulumi:"unasUrl,optional"`
	// UnasUsername is a LOCAL UniFi OS admin account on the UNAS appliance (Drive
	// has no API-key auth; cloud/SSO logins do not work for the local API).
	UnasUsername *string `pulumi:"unasUsername,optional"`
	// UnasPassword is the password for unasUsername.
	UnasPassword *string `pulumi:"unasPassword,optional" provider:"secret"`

	// net is the configured Network client, populated by Configure. It is not a
	// Pulumi field (no struct tag) and never appears in state.
	net unifi.Client
	// protect is the configured Protect client. Only built when an API key is
	// supplied (the official Protect integration API is API-key only). nil
	// otherwise; Protect resources error with a clear message in that case.
	protect protecttypes.ProtectV1
	// drive is the configured UniFi Drive client for the UNAS appliance. Only
	// built when unasUrl + credentials are supplied. nil otherwise; Drive
	// resources error with a clear message in that case.
	drive driveapi.Client
}

// Annotate attaches descriptions, defaults, and env-var fallbacks to the config.
func (c *Config) Annotate(a infer.Annotator) {
	a.Describe(&c.URL, "Base URL of the UniFi controller, e.g. https://192.168.1.1 (omit any /api suffix).")
	a.SetDefault(&c.URL, nil, "UNIFI_URL")
	a.Describe(&c.APIKey, "UniFi OS local API key. Preferred for headless automation. Mutually exclusive with username/password.")
	a.SetDefault(&c.APIKey, nil, "UNIFI_API_KEY")
	a.Describe(&c.Username, "Local admin username. Use with password when not using an API key.")
	a.SetDefault(&c.Username, nil, "UNIFI_USERNAME")
	a.Describe(&c.Password, "Local admin password.")
	a.SetDefault(&c.Password, nil, "UNIFI_PASSWORD")
	a.Describe(&c.Site, `UniFi site name (defaults to "default").`)
	a.SetDefault(&c.Site, "default", "UNIFI_SITE")
	a.Describe(&c.InsecureTLS, "Skip TLS certificate verification (for self-signed controller and UNAS certs).")
	a.SetDefault(&c.InsecureTLS, nil, "UNIFI_INSECURE_TLS")
	a.Describe(&c.UnasURL, "Base URL of the UNAS appliance hosting UniFi Drive, e.g. https://192.168.1.20. A SEPARATE host from `url`; required to manage `unifi:drive:*` resources.")
	a.SetDefault(&c.UnasURL, nil, "UNIFI_UNAS_URL")
	a.Describe(&c.UnasUsername, "Local UniFi OS admin username on the UNAS appliance (UniFi Drive has no API-key auth).")
	a.SetDefault(&c.UnasUsername, nil, "UNIFI_UNAS_USERNAME")
	a.Describe(&c.UnasPassword, "Password for unasUsername.")
	a.SetDefault(&c.UnasPassword, nil, "UNIFI_UNAS_PASSWORD")
}

// buildNetworkClient and buildProtectClient construct the real clients. They are
// package vars so InjectClientsForTest can swap in fakes for hermetic tests.
var (
	buildNetworkClient = unifi.NewClient
	buildProtectClient = func(apiKey, host string, insecureSkipVerify bool) protecttypes.ProtectV1 {
		pc := unifiprotect.NewDefaultConfig(apiKey)
		pc.Hostname = host
		pc.InsecureSkipVerify = insecureSkipVerify
		quiet := logrus.New()
		quiet.SetOutput(io.Discard)
		return unifiprotect.NewClient(context.Background(), pc, quiet).Protect
	}
	// buildDriveClient returns the UNAS Drive client, or (nil, nil) when Drive is
	// not configured, or an error when unasUrl is set without credentials.
	buildDriveClient = func(c *Config) (driveapi.Client, error) {
		if c.UnasURL == nil || *c.UnasURL == "" {
			return nil, nil
		}
		if c.UnasUsername == nil || *c.UnasUsername == "" || c.UnasPassword == nil || *c.UnasPassword == "" {
			return nil, fmt.Errorf("unifi provider: `unasUrl` is set but `unasUsername`/`unasPassword` are required to manage UniFi Drive")
		}
		return driveapi.New(driveapi.Config{
			Host:               *c.UnasURL,
			Username:           *c.UnasUsername,
			Password:           *c.UnasPassword,
			InsecureSkipVerify: c.InsecureTLS != nil && *c.InsecureTLS,
		}), nil
	}
)

// InjectClientsForTest swaps in fake clients and returns a restore func. It lets
// tests in other packages drive CRUD without a live controller. Test-only.
func InjectClientsForTest(net unifi.Client, protect protecttypes.ProtectV1, drive driveapi.Client) func() {
	origNet, origProtect, origDrive := buildNetworkClient, buildProtectClient, buildDriveClient
	buildNetworkClient = func(*unifi.ClientConfig) (unifi.Client, error) { return net, nil }
	buildProtectClient = func(string, string, bool) protecttypes.ProtectV1 { return protect }
	buildDriveClient = func(*Config) (driveapi.Client, error) { return drive, nil }
	return func() { buildNetworkClient, buildProtectClient, buildDriveClient = origNet, origProtect, origDrive }
}

// Configure builds the authenticated UniFi client. Called once per provider
// process, after the receiver has been hydrated from inputs.
func (c *Config) Configure(_ context.Context) error {
	if c.URL == "" {
		return fmt.Errorf("unifi provider: set `url` (or the UNIFI_URL env var)")
	}
	cc := &unifi.ClientConfig{URL: c.URL}
	cc.VerifySSL = c.InsecureTLS == nil || !*c.InsecureTLS
	// Silence the go-unifi logger; provider diagnostics go through Pulumi.
	cc.Logger = unifi.NewDefaultLogger(unifi.DisabledLevel)

	switch {
	case c.APIKey != nil && *c.APIKey != "":
		cc.APIKey = *c.APIKey
	case c.Username != nil && *c.Username != "" && c.Password != nil:
		cc.User = *c.Username
		cc.Password = *c.Password
	default:
		return fmt.Errorf("unifi provider: set either `apiKey` or both `username` and `password`")
	}

	client, err := buildNetworkClient(cc)
	if err != nil {
		return fmt.Errorf("unifi provider: failed to create client for %s: %w", c.URL, err)
	}
	c.net = client

	// Build the Protect client when an API key is available. Protect uses the
	// official integration API, which is API-key only.
	if cc.APIKey != "" {
		host, err := hostOf(c.URL)
		if err != nil {
			return fmt.Errorf("unifi provider: invalid url %q: %w", c.URL, err)
		}
		c.protect = buildProtectClient(cc.APIKey, host, !cc.VerifySSL)
	}

	// Build the UniFi Drive client when the UNAS appliance is configured. The
	// UNAS is a separate host with its own local credentials.
	drive, err := buildDriveClient(c)
	if err != nil {
		return err
	}
	c.drive = drive
	return nil
}

// hostOf extracts host[:port] from a controller URL.
func hostOf(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	if u.Host != "" {
		return u.Host, nil
	}
	return u.Path, nil // tolerate a bare host with no scheme
}

// Network returns the configured UniFi Network client.
func (c Config) Network() unifi.Client { return c.net }

// Controller returns the session-authenticated UniFi OS client for endpoints
// outside the Network application (e.g. the private Protect automations API).
// It is the same underlying client as Network(): go-unifi resolves absolute
// paths (leading "/") against the controller base URL and carries the session
// cookie + CSRF token (or X-API-Key) on every request.
func (c Config) Controller() unifi.Client { return c.net }

// Protect returns the configured UniFi Protect client, or an error if no API
// key was supplied (Protect requires API-key authentication).
func (c Config) Protect() (protecttypes.ProtectV1, error) {
	if c.protect == nil {
		return nil, fmt.Errorf("unifi provider: Protect resources require `apiKey` to be set")
	}
	return c.protect, nil
}

// Drive returns the configured UniFi Drive client for the UNAS appliance, or an
// error if the UNAS endpoint was not configured. Drive runs on the UNAS host —
// a separate UniFi OS console — so it needs its own `unasUrl` and local
// `unasUsername`/`unasPassword`.
func (c Config) Drive() (driveapi.Client, error) {
	if c.drive == nil {
		return nil, fmt.Errorf("unifi provider: Drive resources require `unasUrl`, `unasUsername`, and `unasPassword` to be set (the UNAS appliance is a separate host from the main controller)")
	}
	return c.drive, nil
}

// ResolvedSite returns the configured site, defaulting to "default".
func (c Config) ResolvedSite() string {
	if c.Site != nil && *c.Site != "" {
		return *c.Site
	}
	return "default"
}
