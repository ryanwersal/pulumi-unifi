// SPDX-License-Identifier: Apache-2.0

package protect

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/ryanwersal/pulumi-unifi/provider/config"
	"github.com/ryanwersal/pulumi-unifi/provider/internal/protectapi"
)

// AlarmAutomation manages a UniFi Protect Alarm Manager rule (an
// "automation": device sources + trigger conditions -> actions).
//
// Rule CRUD only exists on Protect's PRIVATE API — the official integration
// API can merely trigger alarms — so this resource carries private-API
// caveats: the surface is unversioned, and consoles set to the "Global"
// Alarm Manager mode reject local rule writes. See provider/internal/protectapi.
//
// Scope: webhook (HTTP_REQUEST) actions only. The resource owns the rule's
// ENTIRE actions list — actions added in the Protect UI to a managed rule
// are removed on the next update. Schedules and other unmodeled settings are
// preserved across updates.
type AlarmAutomation struct{}

// AlarmSource scopes a rule to a device.
type AlarmSource struct {
	// Device is the device MAC address, uppercase hex without separators (e.g. "F4E2C6730625").
	Device string `pulumi:"device"`
	// Type is "include" or "exclude". Defaults to "include".
	Type *string `pulumi:"type,optional"`
}

func (s *AlarmSource) Annotate(a infer.Annotator) {
	a.Describe(&s.Device, `Device is the device MAC address, uppercase hex without separators (e.g. "F4E2C6730625").`)
	a.Describe(&s.Type, `Type is "include" or "exclude". Defaults to "include".`)
	a.SetDefault(&s.Type, "include")
}

// AlarmCondition is a single trigger condition. Conditions are ANDed.
type AlarmCondition struct {
	// Source is the detection trigger, e.g. "motion", "person", "vehicle",
	// "ring", "sensor_door_opened", "sensor_water_leak", "audio_alarm_smoke".
	Source string `pulumi:"source"`
	// Type is the match type. Defaults to "is".
	Type *string `pulumi:"type,optional"`
	// Value refines some sources (e.g. a crossing-line direction or a known license plate).
	Value *string `pulumi:"value,optional"`
}

func (c *AlarmCondition) Annotate(a infer.Annotator) {
	a.Describe(&c.Source, `Source is the detection trigger, e.g. "motion", "person", "vehicle", "ring", "sensor_door_opened", "sensor_water_leak", "audio_alarm_smoke".`)
	a.Describe(&c.Type, `Type is the match type. Defaults to "is".`)
	a.SetDefault(&c.Type, "is")
	a.Describe(&c.Value, `Value refines some sources (e.g. a crossing-line direction or a known license plate).`)
}

// AlarmWebhookAction sends an HTTP request when the rule fires.
type AlarmWebhookAction struct {
	// Url is the webhook target URL.
	Url string `pulumi:"url"`
	// Method is the HTTP method. Defaults to "POST".
	Method *string `pulumi:"method,optional"`
	// Headers are extra request headers.
	Headers map[string]string `pulumi:"headers,optional"`
	// TimeoutMs is the request timeout in milliseconds. Defaults to 30000.
	TimeoutMs *int `pulumi:"timeoutMs,optional"`
	// UseThumbnail attaches the event thumbnail to the request. Defaults to true.
	UseThumbnail *bool `pulumi:"useThumbnail,optional"`
}

func (w *AlarmWebhookAction) Annotate(a infer.Annotator) {
	a.Describe(&w.Url, "Url is the webhook target URL.")
	a.Describe(&w.Method, `Method is the HTTP method. Defaults to "POST".`)
	a.SetDefault(&w.Method, defaultWebhookMethod)
	a.Describe(&w.Headers, "Headers are extra request headers.")
	a.Describe(&w.TimeoutMs, "TimeoutMs is the request timeout in milliseconds. Defaults to 30000.")
	a.SetDefault(&w.TimeoutMs, defaultWebhookTimeoutMs)
	a.Describe(&w.UseThumbnail, "UseThumbnail attaches the event thumbnail to the request. Defaults to true.")
	a.SetDefault(&w.UseThumbnail, true)
}

// AlarmCooldown suppresses repeat fires of the rule.
type AlarmCooldown struct {
	// Enabled toggles the cooldown.
	Enabled bool `pulumi:"enabled"`
	// TimeoutMs is the suppression window in milliseconds.
	TimeoutMs int `pulumi:"timeoutMs"`
}

func (c *AlarmCooldown) Annotate(a infer.Annotator) {
	a.Describe(&c.Enabled, "Enabled toggles the cooldown.")
	a.Describe(&c.TimeoutMs, "TimeoutMs is the suppression window in milliseconds.")
}

// AlarmAutomationArgs are the user-supplied inputs.
type AlarmAutomationArgs struct {
	// Name is the rule's display name.
	Name string `pulumi:"name"`
	// Enabled controls whether the rule fires. Defaults to true.
	Enabled *bool `pulumi:"enabled,optional"`
	// Sources scopes the rule to devices. Empty means all devices.
	Sources []AlarmSource `pulumi:"sources,optional"`
	// Conditions are the trigger conditions (ANDed). At least one is required.
	Conditions []AlarmCondition `pulumi:"conditions"`
	// WebhookActions fire when the rule matches. At least one is required.
	WebhookActions []AlarmWebhookAction `pulumi:"webhookActions"`
	// Cooldown suppresses repeat fires. Defaults to disabled with a 10-minute window.
	Cooldown *AlarmCooldown `pulumi:"cooldown,optional"`
}

func (d *AlarmAutomationArgs) Annotate(a infer.Annotator) {
	a.Describe(&d.Name, "Name is the rule's display name.")
	a.Describe(&d.Enabled, "Enabled controls whether the rule fires. Defaults to true.")
	a.SetDefault(&d.Enabled, true)
	a.Describe(&d.Sources, "Sources scopes the rule to devices. Empty means all devices.")
	a.Describe(&d.Conditions, "Conditions are the trigger conditions (ANDed). At least one is required.")
	a.Describe(&d.WebhookActions, "WebhookActions fire when the rule matches. At least one is required.")
	a.Describe(&d.Cooldown, "Cooldown suppresses repeat fires. Defaults to disabled with a 10-minute window.")
}

// AlarmAutomationState is the persisted state: inputs plus the assigned ID.
type AlarmAutomationState struct {
	AlarmAutomationArgs
	// AutomationId is the controller-assigned rule identifier.
	AutomationId string `pulumi:"automationId"`
}

func (s *AlarmAutomationState) Annotate(a infer.Annotator) {
	a.Describe(&s.AutomationId, "AutomationId is the controller-assigned rule identifier.")
}

func (a *AlarmAutomation) Annotate(an infer.Annotator) {
	an.Describe(&a, "A UniFi Protect Alarm Manager rule (sources + conditions -> webhook actions). "+
		"Uses Protect's private automations API: requires the console's Alarm Manager to be in LOCAL mode, "+
		"and may require username/password auth. The resource owns the rule's entire actions list.")
}

const (
	defaultWebhookMethod    = "POST"
	defaultWebhookTimeoutMs = 30000
	defaultCooldownMs       = 600000
)

func (a AlarmAutomationArgs) validate() error {
	if len(a.Conditions) == 0 {
		return fmt.Errorf("alarm automation %q needs at least one condition", a.Name)
	}
	if len(a.WebhookActions) == 0 {
		return fmt.Errorf("alarm automation %q needs at least one webhook action", a.Name)
	}
	return nil
}

// toAutomation builds the wire-format rule. Slices are always non-nil: the
// private API's strict POST parser wants [] rather than null.
func (a AlarmAutomationArgs) toAutomation() protectapi.Automation {
	auto := protectapi.Automation{
		Name:              a.Name,
		Enable:            derefOr(a.Enabled, true),
		IsCreatedBySystem: false,
		Sources:           []protectapi.Source{},
		Conditions:        []protectapi.ConditionWrapper{},
		HistoryConditions: []json.RawMessage{},
		Schedules:         []json.RawMessage{},
		Actions:           []protectapi.Action{},
		Cooldown:          protectapi.Cooldown{Enable: false, Timeout: defaultCooldownMs},
	}
	for _, s := range a.Sources {
		auto.Sources = append(auto.Sources, protectapi.Source{
			Device: s.Device,
			Type:   derefOr(s.Type, "include"),
		})
	}
	for _, c := range a.Conditions {
		auto.Conditions = append(auto.Conditions, protectapi.ConditionWrapper{
			Condition: protectapi.Condition{
				Type:   derefOr(c.Type, "is"),
				Source: c.Source,
				Value:  derefOr(c.Value, ""),
			},
		})
	}
	for _, w := range a.WebhookActions {
		md, _ := json.Marshal(protectapi.HTTPRequestMetadata{
			URL:          w.Url,
			Method:       derefOr(w.Method, defaultWebhookMethod),
			Headers:      headerList(w.Headers),
			Timeout:      derefOr(w.TimeoutMs, defaultWebhookTimeoutMs),
			UseThumbnail: derefOr(w.UseThumbnail, true),
		})
		auto.Actions = append(auto.Actions, protectapi.Action{
			Type:     "HTTP_REQUEST",
			Metadata: md,
			Order:    -1,
		})
	}
	if a.Cooldown != nil {
		auto.Cooldown = protectapi.Cooldown{Enable: a.Cooldown.Enabled, Timeout: a.Cooldown.TimeoutMs}
	}
	return auto
}

// headerList converts the Pulumi-side header map to the controller's sorted
// {key,value} list (sorted for deterministic diffs).
func headerList(m map[string]string) []protectapi.Header {
	hs := make([]protectapi.Header, 0, len(m))
	for _, k := range slices.Sorted(maps.Keys(m)) {
		hs = append(hs, protectapi.Header{Key: k, Value: m[k]})
	}
	return hs
}

// alarmStateFrom maps a wire-format rule back into resource state. Non-HTTP
// actions (e.g. notifications added in the UI) are not representable and are
// skipped; this resource owns the actions list, so they disappear on the next
// update anyway.
func alarmStateFrom(auto protectapi.Automation) AlarmAutomationState {
	args := AlarmAutomationArgs{
		Name:    auto.Name,
		Enabled: ptr(auto.Enable),
		Cooldown: &AlarmCooldown{
			Enabled:   auto.Cooldown.Enable,
			TimeoutMs: auto.Cooldown.Timeout,
		},
	}
	for _, s := range auto.Sources {
		args.Sources = append(args.Sources, AlarmSource{Device: s.Device, Type: ptr(s.Type)})
	}
	for _, cw := range auto.Conditions {
		c := AlarmCondition{Source: cw.Condition.Source, Type: ptr(cw.Condition.Type)}
		if cw.Condition.Value != "" {
			c.Value = ptr(cw.Condition.Value)
		}
		args.Conditions = append(args.Conditions, c)
	}
	for _, act := range auto.Actions {
		if act.Type != "HTTP_REQUEST" {
			continue
		}
		var md protectapi.HTTPRequestMetadata
		if err := json.Unmarshal(act.Metadata, &md); err != nil {
			continue
		}
		w := AlarmWebhookAction{
			Url:          md.URL,
			Method:       ptr(md.Method),
			TimeoutMs:    ptr(md.Timeout),
			UseThumbnail: ptr(md.UseThumbnail),
		}
		if len(md.Headers) > 0 {
			w.Headers = make(map[string]string, len(md.Headers))
			for _, h := range md.Headers {
				w.Headers[h.Key] = h.Value
			}
		}
		args.WebhookActions = append(args.WebhookActions, w)
	}
	return AlarmAutomationState{AlarmAutomationArgs: args, AutomationId: auto.ID}
}

func (AlarmAutomation) Create(ctx context.Context, req infer.CreateRequest[AlarmAutomationArgs]) (infer.CreateResponse[AlarmAutomationState], error) {
	if err := req.Inputs.validate(); err != nil {
		return infer.CreateResponse[AlarmAutomationState]{}, err
	}
	if req.DryRun {
		return infer.CreateResponse[AlarmAutomationState]{Output: AlarmAutomationState{AlarmAutomationArgs: req.Inputs}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	created, err := protectapi.Create(ctx, cfg.Controller(), req.Inputs.toAutomation())
	if err != nil {
		return infer.CreateResponse[AlarmAutomationState]{}, err
	}
	return infer.CreateResponse[AlarmAutomationState]{ID: created.ID, Output: alarmStateFrom(created)}, nil
}

func (AlarmAutomation) Read(ctx context.Context, req infer.ReadRequest[AlarmAutomationArgs, AlarmAutomationState]) (infer.ReadResponse[AlarmAutomationArgs, AlarmAutomationState], error) {
	cfg := infer.GetConfig[config.Config](ctx)
	raw, err := protectapi.Find(ctx, cfg.Controller(), req.ID)
	if errors.Is(err, protectapi.ErrNotFound) {
		// Empty response (no ID) marks the resource as deleted out-of-band.
		return infer.ReadResponse[AlarmAutomationArgs, AlarmAutomationState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[AlarmAutomationArgs, AlarmAutomationState]{}, err
	}
	var auto protectapi.Automation
	if err := json.Unmarshal(raw, &auto); err != nil {
		return infer.ReadResponse[AlarmAutomationArgs, AlarmAutomationState]{}, fmt.Errorf("decode alarm automation %q: %w", req.ID, err)
	}
	st := alarmStateFrom(auto)
	return infer.ReadResponse[AlarmAutomationArgs, AlarmAutomationState]{ID: req.ID, Inputs: st.AlarmAutomationArgs, State: st}, nil
}

func (AlarmAutomation) Update(ctx context.Context, req infer.UpdateRequest[AlarmAutomationArgs, AlarmAutomationState]) (infer.UpdateResponse[AlarmAutomationState], error) {
	if err := req.Inputs.validate(); err != nil {
		return infer.UpdateResponse[AlarmAutomationState]{}, err
	}
	if req.DryRun {
		return infer.UpdateResponse[AlarmAutomationState]{Output: AlarmAutomationState{AlarmAutomationArgs: req.Inputs, AutomationId: req.ID}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	// PATCH requires the full rule body: read the controller's copy, overlay
	// the fields we manage, and send the merged result back.
	raw, err := protectapi.Find(ctx, cfg.Controller(), req.ID)
	if err != nil {
		return infer.UpdateResponse[AlarmAutomationState]{}, err
	}
	full, err := protectapi.MergeManaged(raw, req.Inputs.toAutomation())
	if err != nil {
		return infer.UpdateResponse[AlarmAutomationState]{}, err
	}
	updated, err := protectapi.Patch(ctx, cfg.Controller(), req.ID, full)
	if err != nil {
		return infer.UpdateResponse[AlarmAutomationState]{}, err
	}
	return infer.UpdateResponse[AlarmAutomationState]{Output: alarmStateFrom(updated)}, nil
}

func (AlarmAutomation) Delete(ctx context.Context, req infer.DeleteRequest[AlarmAutomationState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[config.Config](ctx)
	return infer.DeleteResponse{}, protectapi.Delete(ctx, cfg.Controller(), req.ID)
}
