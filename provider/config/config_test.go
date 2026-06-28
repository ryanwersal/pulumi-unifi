// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"strings"
	"testing"
)

func TestResolvedSite(t *testing.T) {
	def := "default"
	custom := "branch-office"
	cases := []struct {
		name string
		site *string
		want string
	}{
		{"nil defaults", nil, "default"},
		{"empty defaults", strPtr(""), "default"},
		{"custom honored", &custom, "branch-office"},
		{"explicit default", &def, "default"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg := Config{Site: c.site}
			if got := cfg.ResolvedSite(); got != c.want {
				t.Errorf("ResolvedSite() = %q, want %q", got, c.want)
			}
		})
	}
}

func TestHostOf(t *testing.T) {
	cases := []struct {
		raw  string
		want string
	}{
		{"https://192.168.1.1", "192.168.1.1"},
		{"https://10.0.0.1:8443", "10.0.0.1:8443"},
		{"https://unifi.example.com", "unifi.example.com"},
		{"192.168.1.1", "192.168.1.1"}, // bare host, no scheme
	}
	for _, c := range cases {
		got, err := hostOf(c.raw)
		if err != nil {
			t.Errorf("hostOf(%q) error: %v", c.raw, err)
			continue
		}
		if got != c.want {
			t.Errorf("hostOf(%q) = %q, want %q", c.raw, got, c.want)
		}
	}
}

// TestConfigureRequiresAuth verifies the no-credential path errors clearly
// before any network client is constructed (the error is returned ahead of
// unifi.NewClient, so this needs no controller).
func TestConfigureRequiresAuth(t *testing.T) {
	cfg := &Config{URL: "https://192.168.1.1"}
	err := cfg.Configure(context.Background())
	if err == nil {
		t.Fatal("Configure with no credentials returned nil error")
	}
	if !strings.Contains(err.Error(), "apiKey") || !strings.Contains(err.Error(), "username") {
		t.Errorf("error should mention both auth options, got: %v", err)
	}
}

// TestConfigureRequiresURL verifies the empty-URL path errors clearly before any
// client is constructed (url is optional in the schema, supplied via config or
// the UNIFI_URL env var).
func TestConfigureRequiresURL(t *testing.T) {
	cfg := &Config{APIKey: strPtr("k")}
	err := cfg.Configure(context.Background())
	if err == nil || !strings.Contains(err.Error(), "url") {
		t.Fatalf("Configure with empty url should error mentioning url, got: %v", err)
	}
}

func strPtr(s string) *string { return &s }
