// SPDX-License-Identifier: Apache-2.0

package protect

import (
	"encoding/json"
	"testing"

	"github.com/ryanwersal/pulumi-unifi/provider/internal/protectapi"
)

func TestToAutomationDefaults(t *testing.T) {
	args := AlarmAutomationArgs{
		Name:           "Person at door",
		Sources:        []AlarmSource{{Device: "F4E2C6730625"}},
		Conditions:     []AlarmCondition{{Source: "person"}},
		WebhookActions: []AlarmWebhookAction{{Url: "https://example.test/hook"}},
	}
	auto := args.toAutomation()

	if !auto.Enable {
		t.Error("Enable should default to true")
	}
	if auto.Sources[0].Type != "include" {
		t.Errorf("source type should default to include, got %q", auto.Sources[0].Type)
	}
	if auto.Conditions[0].Condition.Type != "is" {
		t.Errorf("condition type should default to is, got %q", auto.Conditions[0].Condition.Type)
	}
	if auto.Cooldown != (protectapi.Cooldown{Enable: false, Timeout: 600000}) {
		t.Errorf("unexpected default cooldown: %+v", auto.Cooldown)
	}

	var md protectapi.HTTPRequestMetadata
	if err := json.Unmarshal(auto.Actions[0].Metadata, &md); err != nil {
		t.Fatal(err)
	}
	if md.Method != "POST" || md.Timeout != 30000 || !md.UseThumbnail {
		t.Errorf("unexpected webhook defaults: %+v", md)
	}
	if md.Headers == nil {
		t.Error("headers should marshal as [] not null")
	}
	if auto.Actions[0].Order != -1 {
		t.Errorf("order should default to -1, got %d", auto.Actions[0].Order)
	}
}

func TestToAutomationSortsHeaders(t *testing.T) {
	args := AlarmAutomationArgs{
		Name:       "h",
		Conditions: []AlarmCondition{{Source: "motion"}},
		WebhookActions: []AlarmWebhookAction{{
			Url:     "https://example.test",
			Headers: map[string]string{"X-B": "2", "X-A": "1"},
		}},
	}
	var md protectapi.HTTPRequestMetadata
	if err := json.Unmarshal(args.toAutomation().Actions[0].Metadata, &md); err != nil {
		t.Fatal(err)
	}
	if len(md.Headers) != 2 || md.Headers[0].Key != "X-A" || md.Headers[1].Key != "X-B" {
		t.Errorf("headers should be sorted by key for deterministic diffs: %+v", md.Headers)
	}
}

func TestAlarmStateRoundTrip(t *testing.T) {
	args := AlarmAutomationArgs{
		Name:       "Vehicle arrival",
		Enabled:    ptr(true),
		Sources:    []AlarmSource{{Device: "AABBCC001122", Type: ptr("include")}},
		Conditions: []AlarmCondition{{Source: "smartDetectLine", Type: ptr("is"), Value: ptr("Arrival - down")}},
		WebhookActions: []AlarmWebhookAction{{
			Url:          "https://example.test/hook",
			Method:       ptr("POST"),
			Headers:      map[string]string{"X-Token": "s"},
			TimeoutMs:    ptr(30000),
			UseThumbnail: ptr(false),
		}},
		Cooldown: &AlarmCooldown{Enabled: true, TimeoutMs: 900000},
	}

	auto := args.toAutomation()
	auto.ID = "6746a0a203df5603e4001e3b"
	st := alarmStateFrom(auto)

	if st.AutomationId != auto.ID {
		t.Errorf("AutomationId = %q", st.AutomationId)
	}
	got, want := st.AlarmAutomationArgs, args
	gj, _ := json.Marshal(got)
	wj, _ := json.Marshal(want)
	if string(gj) != string(wj) {
		t.Errorf("round trip mismatch:\n got %s\nwant %s", gj, wj)
	}
}

func TestAlarmStateFromSkipsNonHTTPActions(t *testing.T) {
	auto := protectapi.Automation{
		Name:       "n",
		Conditions: []protectapi.ConditionWrapper{{Condition: protectapi.Condition{Type: "is", Source: "person"}}},
		Actions: []protectapi.Action{
			{Type: "SEND_NOTIFICATION", Metadata: json.RawMessage(`{"receivers":[]}`), Order: -1},
			{Type: "HTTP_REQUEST", Metadata: json.RawMessage(`{"url":"https://example.test","method":"POST","headers":[],"timeout":30000,"useThumbnail":true}`), Order: -1},
		},
	}
	st := alarmStateFrom(auto)
	if len(st.WebhookActions) != 1 || st.WebhookActions[0].Url != "https://example.test" {
		t.Errorf("expected only the HTTP_REQUEST action mapped: %+v", st.WebhookActions)
	}
}

func TestValidateRequiresConditionsAndActions(t *testing.T) {
	base := AlarmAutomationArgs{
		Name:           "v",
		Conditions:     []AlarmCondition{{Source: "motion"}},
		WebhookActions: []AlarmWebhookAction{{Url: "https://example.test"}},
	}
	if err := base.validate(); err != nil {
		t.Errorf("valid args rejected: %v", err)
	}
	noCond := base
	noCond.Conditions = nil
	if err := noCond.validate(); err == nil {
		t.Error("missing conditions should fail validation")
	}
	noAct := base
	noAct.WebhookActions = nil
	if err := noAct.validate(); err == nil {
		t.Error("missing webhook actions should fail validation")
	}
}
