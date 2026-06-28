// Package protectapi is a minimal client for UniFi Protect's PRIVATE Alarm
// Manager automations API (/proxy/protect/api/automations).
//
// Ubiquiti's official Protect integration API can only TRIGGER alarms
// (POST /v1/alarm-manager/webhook/{id}); rule CRUD exists only on this
// private, unversioned surface. Rather than hand-rolling session handling,
// we ride the configured go-unifi client: its login flow captures the session
// cookie and CSRF token (or sends X-API-Key), and its generic verbs accept
// absolute controller paths.
//
// API quirks, verified against the active community implementations
// (Hovborg/unifi-protect-bridge, sirkirby/unifi-mcp,
// JeffSteinbok/hass-uiprotectalarms):
//   - There is no per-item GET: `GET automations/{id}` 404s. Readers must
//     list and filter.
//   - `PATCH automations/{id}` requires the FULL rule body; partial bodies
//     are rejected. Callers must read-modify-write (see MergeManaged).
//   - `POST automations` parses strictly: unknown keys fail with
//     400 "Failed to parse 'request-body'".
//   - Consoles set to the "Global" Alarm Manager mode reject local rule
//     writes with 400.
package protectapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/filipowm/go-unifi/unifi"
)

const automationsPath = "/proxy/protect/api/automations"

// ErrNotFound is returned by Find when no automation has the requested ID.
var ErrNotFound = errors.New("automation not found")

// Automation is an Alarm Manager rule in the controller's wire format,
// restricted to the fields this provider manages plus the assigned ID.
// Controller-owned fields (userId, status, createdAt, editable, deleted) are
// intentionally absent; updates preserve them via MergeManaged.
type Automation struct {
	ID                string             `json:"id,omitempty"`
	Name              string             `json:"name"`
	Enable            bool               `json:"enable"`
	IsCreatedBySystem bool               `json:"isCreatedBySystem"`
	Sources           []Source           `json:"sources"`
	Conditions        []ConditionWrapper `json:"conditions"`
	HistoryConditions []json.RawMessage  `json:"historyConditions"`
	Schedules         []json.RawMessage  `json:"schedules"`
	Actions           []Action           `json:"actions"`
	Cooldown          Cooldown           `json:"cooldown"`
}

// Source scopes a rule to devices. An empty sources list means all devices.
type Source struct {
	// Device is the device MAC, uppercase hex without separators ("F4E2C6730625").
	Device string `json:"device"`
	// Type is "include" or "exclude".
	Type string `json:"type"`
}

// ConditionWrapper reflects the controller's nesting: each conditions entry
// wraps a single condition object.
type ConditionWrapper struct {
	Condition Condition `json:"condition"`
}

// Condition is a single trigger, e.g. {type: "is", source: "person"}.
type Condition struct {
	Type   string `json:"type"`
	Source string `json:"source"`
	Value  string `json:"value,omitempty"`
}

// Action is one rule action. Metadata stays raw because its shape depends on
// Type and Protect keeps adding action types.
type Action struct {
	Type     string          `json:"type"`
	Metadata json.RawMessage `json:"metadata"`
	Order    int             `json:"order"`
}

// HTTPRequestMetadata is the metadata payload for "HTTP_REQUEST" actions.
type HTTPRequestMetadata struct {
	URL     string   `json:"url"`
	Method  string   `json:"method"`
	Headers []Header `json:"headers"`
	// Timeout is in milliseconds.
	Timeout int `json:"timeout"`
	// UseThumbnail attaches the event thumbnail to the request.
	UseThumbnail bool `json:"useThumbnail"`
}

// Header is one entry of an HTTP_REQUEST action's headers list (the
// controller wants a list, not a map).
type Header struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Cooldown suppresses repeat fires within Timeout milliseconds.
type Cooldown struct {
	Enable  bool `json:"enable"`
	Timeout int  `json:"timeout"`
}

// List returns every Alarm Manager rule as raw JSON. The private API has no
// filtering or pagination; the response is a flat array.
func List(ctx context.Context, c unifi.Client) ([]json.RawMessage, error) {
	var out []json.RawMessage
	if err := c.Get(ctx, automationsPath, nil, &out); err != nil {
		return nil, wrap("list alarm automations", err)
	}
	return out, nil
}

// Find returns the raw rule with the given ID, or ErrNotFound. The API 404s
// on per-item GET, so this lists and filters.
func Find(ctx context.Context, c unifi.Client, id string) (json.RawMessage, error) {
	rules, err := List(ctx, c)
	if err != nil {
		return nil, err
	}
	for _, raw := range rules {
		var probe struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(raw, &probe) == nil && probe.ID == id {
			return raw, nil
		}
	}
	return nil, fmt.Errorf("alarm automation %q: %w", id, ErrNotFound)
}

// Create posts a new rule and returns the controller's echo, including the
// assigned ID.
func Create(ctx context.Context, c unifi.Client, a Automation) (Automation, error) {
	var out Automation
	if err := c.Post(ctx, automationsPath, a, &out); err != nil {
		return Automation{}, wrap(fmt.Sprintf("create alarm automation %q", a.Name), err)
	}
	return out, nil
}

// Patch updates a rule and returns the controller's echo. full must be the
// COMPLETE rule body (build it with MergeManaged); Protect rejects partial
// PATCH bodies.
func Patch(ctx context.Context, c unifi.Client, id string, full map[string]any) (Automation, error) {
	var out Automation
	if err := c.Do(ctx, http.MethodPatch, automationsPath+"/"+id, full, &out); err != nil {
		return Automation{}, wrap(fmt.Sprintf("update alarm automation %q", id), err)
	}
	return out, nil
}

// Delete removes a rule. An already-absent rule is not an error.
func Delete(ctx context.Context, c unifi.Client, id string) error {
	err := c.Delete(ctx, automationsPath+"/"+id, nil, nil)
	if err == nil || errors.Is(err, unifi.ErrNotFound) {
		return nil
	}
	var se *unifi.ServerError
	if errors.As(err, &se) && se.StatusCode == http.StatusNotFound {
		return nil
	}
	return wrap(fmt.Sprintf("delete alarm automation %q", id), err)
}

// managedKeys are the rule fields this provider owns. Everything else on the
// controller's copy — read-only fields (userId, status, createdAt, ...) and
// settings we don't model (schedules, historyConditions) — is preserved
// across updates.
var managedKeys = []string{"name", "enable", "sources", "conditions", "actions", "cooldown"}

// MergeManaged overlays a's managed fields onto raw (the controller's current
// full rule body) and returns the merged body for Patch.
func MergeManaged(raw json.RawMessage, a Automation) (map[string]any, error) {
	var full map[string]any
	if err := json.Unmarshal(raw, &full); err != nil {
		return nil, fmt.Errorf("decode current alarm automation: %w", err)
	}
	ours, err := json.Marshal(a)
	if err != nil {
		return nil, fmt.Errorf("encode desired alarm automation: %w", err)
	}
	var desired map[string]any
	if err := json.Unmarshal(ours, &desired); err != nil {
		return nil, fmt.Errorf("re-decode desired alarm automation: %w", err)
	}
	for _, k := range managedKeys {
		full[k] = desired[k]
	}
	return full, nil
}

// wrap adds actionable hints for the private API's known failure modes.
func wrap(op string, err error) error {
	var se *unifi.ServerError
	if errors.As(err, &se) {
		switch se.StatusCode {
		case http.StatusUnauthorized, http.StatusForbidden:
			return fmt.Errorf("%s: %w (the private Protect automations API may not accept API-key auth; configure `username`/`password`)", op, err)
		case http.StatusBadRequest:
			return fmt.Errorf("%s: %w (consoles using the GLOBAL Alarm Manager mode reject local rule writes; check the Alarm Manager settings on the console)", op, err)
		}
	}
	return fmt.Errorf("%s: %w", op, err)
}
