// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/blang/semver"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/integration"
)

// goldenPath is the committed full-schema snapshot. Regenerate intentionally
// with `UPDATE_SCHEMA=1 go test ./provider/`.
var goldenPath = filepath.Join("testdata", "schema.json")

// TestSchemaGolden pins the entire generated schema so any change — a renamed
// field, a dropped default, a new type — is surfaced in review instead of
// slipping out through a silent SDK regen.
func TestSchemaGolden(t *testing.T) {
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

	var indented bytes.Buffer
	if err := json.Indent(&indented, []byte(resp.Schema), "", "  "); err != nil {
		t.Fatalf("indent schema: %v", err)
	}
	got := append(indented.Bytes(), '\n')

	if os.Getenv("UPDATE_SCHEMA") != "" {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o750); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}
		if err := os.WriteFile(goldenPath, got, 0o600); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		return
	}

	want, err := os.ReadFile(goldenPath) //nolint:gosec // fixed in-repo test path
	if err != nil {
		t.Fatalf("read golden (run `UPDATE_SCHEMA=1 go test ./provider/` to create it): %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("schema differs from %s. If intentional, regenerate with `UPDATE_SCHEMA=1 go test ./provider/`", goldenPath)
	}
}
