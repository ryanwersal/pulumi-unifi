// SPDX-License-Identifier: Apache-2.0

package network

import (
	"reflect"
	"testing"
)

func TestContentFilteringScheduleDelta(t *testing.T) {
	live := map[string]any{"schedule": map[string]any{"mode": "EVERY_DAY", "time_range_start": "08:00", "time_range_end": "17:00", "future_timezone": "UTC"}, "future_flag": true}
	old := map[string]any{"schedule": map[string]any{"mode": "EVERY_DAY", "time_range_start": "08:00", "time_range_end": "17:00"}}
	next := map[string]any{"schedule": map[string]any{"mode": "ALWAYS"}}
	got := contentFilteringDelta(live, old, next)
	want := map[string]any{"schedule": map[string]any{"mode": "ALWAYS", "future_timezone": "UTC"}, "future_flag": true}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("schedule delta got %#v want %#v", got, want)
	}
	if live["schedule"].(map[string]any)["mode"] != "EVERY_DAY" {
		t.Fatal("mutated live snapshot")
	}
}
func TestContentFilteringPayloadDefaults(t *testing.T) {
	body, err := contentFilteringPayload(ContentFilteringArgs{Name: "Default", NetworkIds: []string{"lan"}, Enabled: ptr(false)})
	if err != nil {
		t.Fatal(err)
	}
	if body["enabled"] != false {
		t.Fatal("lost explicit false")
	}
	if body["schedule"].(map[string]any)["mode"] != "ALWAYS" {
		t.Fatal("missing schedule default")
	}
	for _, k := range []string{"allow_list", "block_list", "safe_search", "categories", "client_macs"} {
		if a, ok := body[k].([]any); !ok || len(a) != 0 {
			t.Fatalf("%s should be an empty array, got %#v", k, body[k])
		}
	}
}
