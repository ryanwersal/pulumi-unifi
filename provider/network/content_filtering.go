// SPDX-License-Identifier: Apache-2.0

package network

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"

	"github.com/filipowm/go-unifi/unifi"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/ryanwersal/pulumi-unifi/provider/config"
)

// ContentFiltering owns a Network v2 content-filter policy.
type ContentFiltering struct{}

type ContentFilteringArgs struct {
	Name       string                    `pulumi:"name"`
	Enabled    *bool                     `pulumi:"enabled,optional"`
	NetworkIds []string                  `pulumi:"networkIds,optional"`
	ClientMacs []string                  `pulumi:"clientMacs,optional"`
	Categories []string                  `pulumi:"categories,optional"`
	AllowList  []string                  `pulumi:"allowList,optional"`
	BlockList  []string                  `pulumi:"blockList,optional"`
	SafeSearch []string                  `pulumi:"safeSearch,optional"`
	Schedule   *ContentFilteringSchedule `pulumi:"schedule,optional"`
}

// ContentFilteringScheduleMode selects when a policy is active.
type ContentFilteringScheduleMode string

const (
	ContentFilteringScheduleModeAlways      ContentFilteringScheduleMode = "ALWAYS"
	ContentFilteringScheduleModeEveryDay    ContentFilteringScheduleMode = "EVERY_DAY"
	ContentFilteringScheduleModeEveryWeek   ContentFilteringScheduleMode = "EVERY_WEEK"
	ContentFilteringScheduleModeOneTimeOnly ContentFilteringScheduleMode = "ONE_TIME_ONLY"
	ContentFilteringScheduleModeCustom      ContentFilteringScheduleMode = "CUSTOM"
)

func (ContentFilteringScheduleMode) Values() []infer.EnumValue[ContentFilteringScheduleMode] {
	return []infer.EnumValue[ContentFilteringScheduleMode]{
		{Name: "Always", Value: ContentFilteringScheduleModeAlways, Description: "Always active."},
		{Name: "EveryDay", Value: ContentFilteringScheduleModeEveryDay, Description: "Active every day."},
		{Name: "EveryWeek", Value: ContentFilteringScheduleModeEveryWeek, Description: "Active on selected weekdays."},
		{Name: "OneTimeOnly", Value: ContentFilteringScheduleModeOneTimeOnly, Description: "Active on a single date."},
		{Name: "Custom", Value: ContentFilteringScheduleModeCustom, Description: "Custom recurring schedule."},
	}
}

type ContentFilteringSchedule struct {
	Mode           ContentFilteringScheduleMode `pulumi:"mode"`
	Date           *string                      `pulumi:"date,optional"`
	DateStart      *string                      `pulumi:"dateStart,optional"`
	DateEnd        *string                      `pulumi:"dateEnd,optional"`
	RepeatOnDays   []string                     `pulumi:"repeatOnDays,optional"`
	TimeAllDay     *bool                        `pulumi:"timeAllDay,optional"`
	TimeRangeStart *string                      `pulumi:"timeRangeStart,optional"`
	TimeRangeEnd   *string                      `pulumi:"timeRangeEnd,optional"`
}

type ContentFilteringState struct {
	ContentFilteringArgs
	PolicyId string `pulumi:"policyId"`
}

func (r *ContentFiltering) Annotate(a infer.Annotator) {
	a.Describe(&r, "A UniFi Network v2 content-filter policy. Import using its controller ID. Updates preserve unmodeled controller fields. Delete removes the entire policy; protect imported security policies or use retainOnDelete. Do not manage the same policy with multiple resources or writers.")
}
func (r *ContentFilteringArgs) Annotate(a infer.Annotator) {
	a.Describe(&r.Name, "Policy display name.")
	a.Describe(&r.Enabled, "Whether the policy is active. Defaults to true.")
	a.SetDefault(&r.Enabled, true)
	a.Describe(&r.NetworkIds, "Controller network IDs to filter. Set networks, client MACs, or both; an empty scope is rejected.")
	a.Describe(&r.ClientMacs, "Client MAC addresses to filter.")
	a.Describe(&r.Categories, "Controller category identifiers, including FAMILY and ADVERTISEMENT. Additional categories depend on firmware and subscriptions. An empty list clears category filtering.")
	a.Describe(&r.AllowList, "Domains allowed by this policy, using controller domain syntax (for example pinggy-free.link). Entries apply according to UniFi's domain matching rules. An empty list clears the allowlist.")
	a.Describe(&r.BlockList, "Additional domains blocked by this policy. An empty list clears the blocklist.")
	a.Describe(&r.SafeSearch, "Search providers with enforced safe search: GOOGLE, YOUTUBE, BING, or other controller-supported identifiers.")
	a.Describe(&r.Schedule, "Activation schedule. Defaults to ALWAYS. Times use the controller's timezone. Unknown schedule fields are preserved during updates.")
}
func (s *ContentFilteringSchedule) Annotate(a infer.Annotator) {
	a.Describe(&s.Mode, "Schedule mode: ALWAYS, EVERY_DAY, EVERY_WEEK, ONE_TIME_ONLY, or CUSTOM.")
	a.Describe(&s.Date, "Single activation date, YYYY-MM-DD.")
	a.Describe(&s.DateStart, "First activation date, YYYY-MM-DD.")
	a.Describe(&s.DateEnd, "Last activation date, YYYY-MM-DD.")
	a.Describe(&s.RepeatOnDays, "Recurring weekdays: mon, tue, wed, thu, fri, sat, sun.")
	a.Describe(&s.TimeAllDay, "Whether the schedule covers the full day.")
	a.Describe(&s.TimeRangeStart, "Start time, HH:MM.")
	a.Describe(&s.TimeRangeEnd, "End time, HH:MM.")
}
func (s *ContentFilteringState) Annotate(a infer.Annotator) {
	a.Describe(&s.PolicyId, "Controller-assigned content-filter policy ID.")
}

var contentFilterID = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func (ContentFiltering) Check(ctx context.Context, req infer.CheckRequest) (infer.CheckResponse[ContentFilteringArgs], error) {
	args, failures, err := infer.DefaultCheck[ContentFilteringArgs](ctx, req.NewInputs)
	if err != nil || len(failures) > 0 {
		return infer.CheckResponse[ContentFilteringArgs]{Inputs: args, Failures: failures}, err
	}
	if !req.NewInputs.Get("name").HasComputed() && strings.TrimSpace(args.Name) == "" {
		failures = append(failures, p.CheckFailure{Property: "name", Reason: "must not be blank"})
	}
	if !req.NewInputs.Get("networkIds").HasComputed() && !req.NewInputs.Get("clientMacs").HasComputed() && len(args.NetworkIds) == 0 && len(args.ClientMacs) == 0 {
		failures = append(failures, p.CheckFailure{Property: "networkIds", Reason: "specify at least one network ID or client MAC to avoid an ambiguous filtering scope"})
	}
	return infer.CheckResponse[ContentFilteringArgs]{Inputs: args, Failures: failures}, nil
}

// Relative to go-unifi's auto-detected /api or /proxy/network/api prefix.
func contentFilteringPath(site, id string) (string, error) {
	if !contentFilterID.MatchString(site) || (id != "" && !contentFilterID.MatchString(id)) {
		return "", fmt.Errorf("content-filter site and ID must contain only letters, digits, underscores, or hyphens")
	}
	path := "../v2/api/site/" + url.PathEscape(site) + "/content-filtering"
	if id != "" {
		path += "/" + url.PathEscape(id)
	}
	return path, nil
}

func contentFilteringPayload(args ContentFilteringArgs) (map[string]any, error) {
	schedule := map[string]any{"mode": "ALWAYS"}
	if a := args.Schedule; a != nil {
		schedule["mode"] = a.Mode
		setContentFilteringOptional(schedule, "date", a.Date)
		setContentFilteringOptional(schedule, "date_start", a.DateStart)
		setContentFilteringOptional(schedule, "date_end", a.DateEnd)
		setContentFilteringOptional(schedule, "time_all_day", a.TimeAllDay)
		setContentFilteringOptional(schedule, "time_range_start", a.TimeRangeStart)
		setContentFilteringOptional(schedule, "time_range_end", a.TimeRangeEnd)
		if len(a.RepeatOnDays) > 0 {
			schedule["repeat_on_days"] = a.RepeatOnDays
		}
	}
	raw := map[string]any{
		"name": args.Name, "enabled": derefOr(args.Enabled, true),
		"network_ids": args.NetworkIds, "client_macs": args.ClientMacs,
		"categories": args.Categories, "allow_list": args.AllowList, "block_list": args.BlockList,
		"safe_search": args.SafeSearch, "schedule": schedule,
	}
	// Normalize maps to their JSON representation for reliable delta comparisons.
	b, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err = json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	for _, k := range []string{"network_ids", "client_macs", "categories", "allow_list", "block_list", "safe_search"} {
		if out[k] == nil {
			out[k] = []any{}
		}
	}
	return out, nil
}
func setContentFilteringOptional[T any](m map[string]any, k string, v *T) {
	if v != nil {
		m[k] = *v
	}
}
func contentFilteringOptional[T any](m map[string]any, k string) *T {
	if v, ok := m[k].(T); ok {
		return &v
	}
	return nil
}

// Only apply the desired delta to the latest live policy. Recurse into schedule
// objects so future controller fields survive, including during an allowlist edit.
func contentFilteringDelta(live, old, next map[string]any) map[string]any {
	out := maps.Clone(live)
	if out == nil {
		out = map[string]any{}
	}
	for k, value := range next {
		before, exists := old[k]
		if exists && reflect.DeepEqual(before, value) {
			continue
		}
		if child, ok := value.(map[string]any); ok {
			prior, _ := before.(map[string]any)
			current, _ := live[k].(map[string]any)
			out[k] = contentFilteringDelta(current, prior, child)
		} else {
			out[k] = value
		}
	}
	for k := range old {
		if _, exists := next[k]; !exists {
			delete(out, k)
		}
	}
	return out
}

func contentFilteringState(raw map[string]any) (ContentFilteringState, error) {
	id, _ := raw["_id"].(string)
	if !contentFilterID.MatchString(id) {
		return ContentFilteringState{}, fmt.Errorf("controller returned content-filter policy without a valid ID")
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return ContentFilteringState{}, err
	}
	var wire unifi.ContentFiltering
	if err = json.Unmarshal(b, &wire); err != nil {
		return ContentFilteringState{}, err
	}
	args := ContentFilteringArgs{
		Name: wire.Name, Enabled: ptr(wire.Enabled), NetworkIds: wire.NetworkIDs, ClientMacs: wire.ClientMACs,
		Categories: wire.Categories, AllowList: wire.AllowList, BlockList: wire.BlockList, SafeSearch: wire.SafeSearch,
	}
	if schedule, ok := raw["schedule"].(map[string]any); ok {
		args.Schedule = &ContentFilteringSchedule{
			Mode: ContentFilteringScheduleMode(wire.Schedule.Mode), RepeatOnDays: wire.Schedule.RepeatOnDays,
			Date:           contentFilteringOptional[string](schedule, "date"),
			DateStart:      contentFilteringOptional[string](schedule, "date_start"),
			DateEnd:        contentFilteringOptional[string](schedule, "date_end"),
			TimeAllDay:     contentFilteringOptional[bool](schedule, "time_all_day"),
			TimeRangeStart: contentFilteringOptional[string](schedule, "time_range_start"),
			TimeRangeEnd:   contentFilteringOptional[string](schedule, "time_range_end"),
		}
	}
	return ContentFilteringState{ContentFilteringArgs: args, PolicyId: id}, nil
}

func contentFilteringRequest(ctx context.Context, method, id string, body map[string]any) (map[string]any, error) {
	cfg := infer.GetConfig[config.Config](ctx)
	path, err := contentFilteringPath(cfg.ResolvedSite(), id)
	if err != nil {
		return nil, err
	}
	var request any
	if body != nil {
		request = body
	}
	if method == http.MethodGet {
		// Current Network firmware exposes collection GET only (item GET is 405).
		// A missing endpoint is an error, not evidence that the policy was deleted.
		path, err = contentFilteringPath(cfg.ResolvedSite(), "")
		if err != nil {
			return nil, err
		}
		var policies []map[string]any
		if err = cfg.Network().Do(ctx, method, path, nil, &policies); err != nil {
			return nil, err
		}
		for _, policy := range policies {
			if policy["_id"] == id {
				return policy, nil
			}
		}
		return nil, unifi.ErrNotFound
	}
	if method == http.MethodDelete {
		return nil, cfg.Network().Do(ctx, method, path, request, nil)
	}
	var raw map[string]any
	err = cfg.Network().Do(ctx, method, path, request, &raw)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func (ContentFiltering) Create(ctx context.Context, req infer.CreateRequest[ContentFilteringArgs]) (infer.CreateResponse[ContentFilteringState], error) {
	if req.DryRun {
		return infer.CreateResponse[ContentFilteringState]{Output: ContentFilteringState{ContentFilteringArgs: req.Inputs}}, nil
	}
	payload, err := contentFilteringPayload(req.Inputs)
	if err != nil {
		return infer.CreateResponse[ContentFilteringState]{}, err
	}
	raw, err := contentFilteringRequest(ctx, http.MethodPost, "", payload)
	if err != nil {
		return infer.CreateResponse[ContentFilteringState]{}, wrap("create content-filter policy", err)
	}
	state, err := contentFilteringState(raw)
	return infer.CreateResponse[ContentFilteringState]{ID: state.PolicyId, Output: state}, err
}
func (ContentFiltering) Read(ctx context.Context, req infer.ReadRequest[ContentFilteringArgs, ContentFilteringState]) (infer.ReadResponse[ContentFilteringArgs, ContentFilteringState], error) {
	if req.ID == "" {
		return infer.ReadResponse[ContentFilteringArgs, ContentFilteringState]{}, fmt.Errorf("content-filter import requires a policy ID")
	}
	raw, err := contentFilteringRequest(ctx, http.MethodGet, req.ID, nil)
	if errors.Is(err, unifi.ErrNotFound) {
		return infer.ReadResponse[ContentFilteringArgs, ContentFilteringState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[ContentFilteringArgs, ContentFilteringState]{}, wrap("read content-filter policy", err)
	}
	state, err := contentFilteringState(raw)
	if err == nil && state.PolicyId != req.ID {
		err = fmt.Errorf("controller returned a different content-filter policy ID")
	}
	return infer.ReadResponse[ContentFilteringArgs, ContentFilteringState]{ID: req.ID, Inputs: state.ContentFilteringArgs, State: state}, err
}
func (ContentFiltering) Update(ctx context.Context, req infer.UpdateRequest[ContentFilteringArgs, ContentFilteringState]) (infer.UpdateResponse[ContentFilteringState], error) {
	if req.DryRun {
		return infer.UpdateResponse[ContentFilteringState]{Output: ContentFilteringState{ContentFilteringArgs: req.Inputs, PolicyId: req.ID}}, nil
	}
	live, err := contentFilteringRequest(ctx, http.MethodGet, req.ID, nil)
	if err != nil {
		return infer.UpdateResponse[ContentFilteringState]{}, wrap("read content-filter policy before update", err)
	}
	liveState, err := contentFilteringState(live)
	if err != nil {
		return infer.UpdateResponse[ContentFilteringState]{}, err
	}
	if liveState.PolicyId != req.ID {
		return infer.UpdateResponse[ContentFilteringState]{}, fmt.Errorf("controller returned a different content-filter policy ID")
	}
	old, err := contentFilteringPayload(req.State.ContentFilteringArgs)
	if err != nil {
		return infer.UpdateResponse[ContentFilteringState]{}, err
	}
	next, err := contentFilteringPayload(req.Inputs)
	if err != nil {
		return infer.UpdateResponse[ContentFilteringState]{}, err
	}
	raw, err := contentFilteringRequest(ctx, http.MethodPut, req.ID, contentFilteringDelta(live, old, next))
	if err != nil {
		return infer.UpdateResponse[ContentFilteringState]{}, wrap("update content-filter policy", err)
	}
	state, err := contentFilteringState(raw)
	if err == nil && state.PolicyId != req.ID {
		err = fmt.Errorf("controller returned a different content-filter policy ID")
	}
	return infer.UpdateResponse[ContentFilteringState]{Output: state}, err
}
func (ContentFiltering) Delete(ctx context.Context, req infer.DeleteRequest[ContentFilteringState]) (infer.DeleteResponse, error) {
	if req.ID == "" {
		return infer.DeleteResponse{}, fmt.Errorf("delete content-filter policy requires an ID")
	}
	_, err := contentFilteringRequest(ctx, http.MethodDelete, req.ID, nil)
	if notFound(err) {
		return infer.DeleteResponse{}, nil
	}
	return infer.DeleteResponse{}, wrap("delete content-filter policy", err)
}
