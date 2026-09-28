// SPDX-License-Identifier: Apache-2.0

package network

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/url"
	"reflect"

	"github.com/filipowm/go-unifi/unifi"
)

// Keep fields absent from go-unifi's typed Network (including current WAN
// routing and failover settings) intact. Do not round-trip the live object
// through that struct before writing it back.
func vlanPayload(args VlanArgs, id string) (map[string]any, error) {
	data, err := json.Marshal(args.toUnifi(id))
	if err != nil {
		return nil, fmt.Errorf("encode network inputs: %w", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	if args.Wan != nil && args.Wan.Dhcpv6PdSizeAuto != nil {
		payload["wan_dhcpv6_pd_size_auto"] = *args.Wan.Dhcpv6PdSizeAuto
	} else {
		delete(payload, "wan_dhcpv6_pd_size_auto")
	}
	return payload, nil
}

// Apply the input delta, not all go-unifi zero values, to the latest live
// object. Comparing with prior state also lets a refreshed drifted value be
// corrected while preserving unrelated controller-managed fields.
func vlanUpdatePayload(live map[string]any, prior, desired VlanArgs, id string) (map[string]any, error) {
	old, err := vlanPayload(prior, id)
	if err != nil {
		return nil, err
	}
	next, err := vlanPayload(desired, id)
	if err != nil {
		return nil, err
	}
	result := maps.Clone(live)
	for key, value := range next {
		before, exists := old[key]
		if !exists || !reflect.DeepEqual(before, value) {
			result[key] = value
		}
	}
	for key := range old {
		if _, exists := next[key]; !exists {
			delete(result, key)
		}
	}
	return result, nil
}

func vlanRawState(raw map[string]any, prior VlanArgs) (VlanState, error) {
	data, err := json.Marshal(raw)
	if err != nil {
		return VlanState{}, err
	}
	var n unifi.Network
	if err := json.Unmarshal(data, &n); err != nil {
		return VlanState{}, fmt.Errorf("decode network: %w", err)
	}
	state := vlanStateFrom(&n, prior)
	if value, ok := raw["wan_dhcpv6_pd_size_auto"].(bool); ok {
		if state.Wan == nil {
			state.Wan = &VlanWan{}
		}
		state.Wan.Dhcpv6PdSizeAuto = ptr(value)
	}
	return state, nil
}

func vlanRequest(ctx context.Context, client unifi.Client, method, site, id string, body map[string]any) (map[string]any, error) {
	path := fmt.Sprintf("s/%s/rest/networkconf", url.PathEscape(site))
	if id != "" {
		path += "/" + url.PathEscape(id)
	}
	var response struct {
		Data []map[string]any `json:"data"`
	}
	// A nil interface avoids a JSON null body on GET.
	var request any
	if body != nil {
		request = body
	}
	if err := client.Do(ctx, method, path, request, &response); err != nil {
		return nil, err
	}
	if len(response.Data) != 1 {
		return nil, unifi.ErrNotFound
	}
	return response.Data[0], nil
}
