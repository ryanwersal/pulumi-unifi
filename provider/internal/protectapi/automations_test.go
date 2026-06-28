package protectapi

import (
	"encoding/json"
	"testing"
)

// controllerRule mimics a full rule body as returned by GET /automations,
// including read-only fields and unmodeled settings the provider must
// preserve across updates. Shape taken from the hass-uiprotectalarms fixture.
const controllerRule = `{
	"name": "Person at Front Door",
	"enable": true,
	"isCreatedBySystem": false,
	"editable": true,
	"sources": [{"device": "F4E2C6730625", "type": "include"}],
	"conditions": [{"condition": {"type": "is", "source": "person"}}],
	"schedules": [{"schedule": {"type": "is", "unit": "time", "values": ["night"]}}],
	"actions": [{"type": "SEND_NOTIFICATION", "metadata": {"receivers": []}, "order": -1}],
	"userId": "abc123",
	"status": {"lastExecutedAt": 1732777401247, "lastExecutedState": "ok", "total": 5},
	"cooldown": {"enable": true, "timeout": 900000},
	"id": "6746a0a203df5603e4001e3b",
	"deleted": false,
	"createdAt": 1732681890987
}`

func TestMergeManagedOverlaysManagedFields(t *testing.T) {
	desired := Automation{
		Name:              "Renamed",
		Enable:            false,
		Sources:           []Source{},
		Conditions:        []ConditionWrapper{{Condition: Condition{Type: "is", Source: "vehicle"}}},
		HistoryConditions: []json.RawMessage{},
		Schedules:         []json.RawMessage{},
		Actions: []Action{{
			Type:     "HTTP_REQUEST",
			Metadata: json.RawMessage(`{"url":"https://example.test/hook","method":"POST","headers":[],"timeout":30000,"useThumbnail":true}`),
			Order:    -1,
		}},
		Cooldown: Cooldown{Enable: false, Timeout: 600000},
	}

	merged, err := MergeManaged(json.RawMessage(controllerRule), desired)
	if err != nil {
		t.Fatalf("MergeManaged: %v", err)
	}

	if merged["name"] != "Renamed" {
		t.Errorf("name not overlaid: %v", merged["name"])
	}
	if merged["enable"] != false {
		t.Errorf("enable not overlaid: %v", merged["enable"])
	}
	if conds := merged["conditions"].([]any); len(conds) != 1 {
		t.Errorf("conditions not overlaid: %v", merged["conditions"])
	}
	if acts := merged["actions"].([]any); len(acts) != 1 {
		t.Fatalf("actions not overlaid: %v", merged["actions"])
	} else if acts[0].(map[string]any)["type"] != "HTTP_REQUEST" {
		t.Errorf("UI action survived overlay: %v", acts[0])
	}
}

func TestMergeManagedPreservesUnmanagedFields(t *testing.T) {
	merged, err := MergeManaged(json.RawMessage(controllerRule), Automation{Name: "x"})
	if err != nil {
		t.Fatalf("MergeManaged: %v", err)
	}

	// Read-only and unmodeled fields must survive the read-modify-write so the
	// full-body PATCH doesn't clobber them.
	for _, key := range []string{"id", "userId", "status", "createdAt", "deleted", "editable"} {
		if _, ok := merged[key]; !ok {
			t.Errorf("unmanaged field %q dropped by merge", key)
		}
	}
	scheds, ok := merged["schedules"].([]any)
	if !ok || len(scheds) != 1 {
		t.Errorf("UI-configured schedules not preserved: %v", merged["schedules"])
	}
}

func TestAutomationMarshalsSlicesAsArrays(t *testing.T) {
	// The private API's strict POST parser wants [] rather than null, and
	// camelCase keys only.
	a := Automation{
		Name:              "t",
		Sources:           []Source{},
		Conditions:        []ConditionWrapper{},
		HistoryConditions: []json.RawMessage{},
		Schedules:         []json.RawMessage{},
		Actions:           []Action{},
	}
	out, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"sources", "conditions", "historyConditions", "schedules", "actions"} {
		if _, ok := m[key].([]any); !ok {
			t.Errorf("%s should marshal as a JSON array, got %T", key, m[key])
		}
	}
	if _, ok := m["isCreatedBySystem"]; !ok {
		t.Error("isCreatedBySystem missing (POST parser is strict camelCase)")
	}
}
