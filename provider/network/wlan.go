package network

import (
	"context"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/ryanwersal/pulumi-unifi/provider/config"
)

// Wlan is the controlling (marker) struct for a UniFi wireless network (SSID).
type Wlan struct{}

// WlanArgs are the user-supplied inputs for a WLAN.
type WlanArgs struct {
	// Name is the SSID.
	Name string `pulumi:"name"`
	// NetworkId binds the WLAN to a network/VLAN (the network's `_id`).
	NetworkId string `pulumi:"networkId"`
	// WlanGroupId is the WLAN group to attach to. Required on many controllers.
	WlanGroupId *string `pulumi:"wlanGroupId,optional"`
	// Enabled controls whether the SSID is broadcast. Defaults to true.
	Enabled *bool `pulumi:"enabled,optional"`
	// Security: open | wpapsk | wpaeap | wep | osen. Defaults to "wpapsk".
	Security *string `pulumi:"security,optional"`
	// Passphrase is the WPA pre-shared key (8-63 chars). Secret.
	Passphrase *string `pulumi:"passphrase,optional" provider:"secret"`
	// HideSsid hides the SSID from broadcast. Defaults to false.
	HideSsid *bool `pulumi:"hideSsid,optional"`
	// WpaMode: auto | wpa1 | wpa2.
	WpaMode *string `pulumi:"wpaMode,optional"`
}

// WlanState is the persisted state: inputs plus controller-assigned fields.
type WlanState struct {
	WlanArgs
	// WlanId is the controller-assigned identifier (the UniFi `_id`).
	WlanId string `pulumi:"wlanId"`
}

func (a WlanArgs) toUnifi(id string) *unifi.WLAN {
	w := &unifi.WLAN{
		ID:        id,
		Name:      a.Name,
		NetworkID: a.NetworkId,
		Enabled:   derefOr(a.Enabled, true),
		Security:  derefOr(a.Security, "wpapsk"),
		HideSSID:  derefOr(a.HideSsid, false),
	}
	if a.WlanGroupId != nil {
		w.WLANGroupID = *a.WlanGroupId
	}
	if a.Passphrase != nil {
		w.XPassphrase = *a.Passphrase
	}
	if a.WpaMode != nil {
		w.WPAMode = *a.WpaMode
	}
	return w
}

func wlanStateFrom(w *unifi.WLAN, passphrase *string) WlanState {
	args := WlanArgs{
		Name:      w.Name,
		NetworkId: w.NetworkID,
		Enabled:   ptr(w.Enabled),
		Security:  ptr(w.Security),
		HideSsid:  ptr(w.HideSSID),
		// Preserve the user-provided passphrase; the controller may not echo it back.
		Passphrase: passphrase,
	}
	if w.WLANGroupID != "" {
		args.WlanGroupId = ptr(w.WLANGroupID)
	}
	if w.WPAMode != "" {
		args.WpaMode = ptr(w.WPAMode)
	}
	return WlanState{WlanArgs: args, WlanId: w.ID}
}

func (Wlan) Create(ctx context.Context, req infer.CreateRequest[WlanArgs]) (infer.CreateResponse[WlanState], error) {
	if req.DryRun {
		return infer.CreateResponse[WlanState]{Output: WlanState{WlanArgs: req.Inputs}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	created, err := cfg.Network().CreateWLAN(ctx, cfg.ResolvedSite(), req.Inputs.toUnifi(""))
	if err != nil {
		return infer.CreateResponse[WlanState]{}, err
	}
	return infer.CreateResponse[WlanState]{ID: created.ID, Output: wlanStateFrom(created, req.Inputs.Passphrase)}, nil
}

func (Wlan) Read(ctx context.Context, req infer.ReadRequest[WlanArgs, WlanState]) (infer.ReadResponse[WlanArgs, WlanState], error) {
	cfg := infer.GetConfig[config.Config](ctx)
	w, err := cfg.Network().GetWLAN(ctx, cfg.ResolvedSite(), req.ID)
	if err != nil {
		return infer.ReadResponse[WlanArgs, WlanState]{}, err
	}
	st := wlanStateFrom(w, req.Inputs.Passphrase)
	return infer.ReadResponse[WlanArgs, WlanState]{ID: req.ID, Inputs: st.WlanArgs, State: st}, nil
}

func (Wlan) Update(ctx context.Context, req infer.UpdateRequest[WlanArgs, WlanState]) (infer.UpdateResponse[WlanState], error) {
	if req.DryRun {
		return infer.UpdateResponse[WlanState]{Output: WlanState{WlanArgs: req.Inputs, WlanId: req.ID}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	updated, err := cfg.Network().UpdateWLAN(ctx, cfg.ResolvedSite(), req.Inputs.toUnifi(req.ID))
	if err != nil {
		return infer.UpdateResponse[WlanState]{}, err
	}
	return infer.UpdateResponse[WlanState]{Output: wlanStateFrom(updated, req.Inputs.Passphrase)}, nil
}

func (Wlan) Delete(ctx context.Context, req infer.DeleteRequest[WlanState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[config.Config](ctx)
	return infer.DeleteResponse{}, cfg.Network().DeleteWLAN(ctx, cfg.ResolvedSite(), req.ID)
}
