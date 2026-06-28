package network

import (
	"context"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/ryanwersal/pulumi-unifi/provider/config"
)

// Wlan is the controlling (marker) struct for a UniFi wireless network (SSID).
type Wlan struct{}

// WlanPrivatePsk is a single private pre-shared key entry. Each PSK can place
// connecting clients onto a specific network/VLAN based on the key used.
type WlanPrivatePsk struct {
	// Password is the pre-shared key for this entry (8-255 chars). Secret.
	Password string `pulumi:"password" provider:"secret"`
	// NetworkId is the network/VLAN (`_id`) clients using this key are placed on.
	NetworkId *string `pulumi:"networkId,optional"`
}

func (d *WlanPrivatePsk) Annotate(a infer.Annotator) {
	a.Describe(&d.Password, "Password is the pre-shared key for this entry (8-255 chars). Secret.")
	a.Describe(&d.NetworkId, "NetworkId is the network/VLAN (`_id`) clients using this key are placed on.")
}

// WlanSaePsk is a single WPA3 SAE pre-shared key entry, optionally bound to a
// MAC address and placed onto a specific VLAN.
type WlanSaePsk struct {
	// Psk is the SAE pre-shared key (8-255 chars). Secret.
	Psk string `pulumi:"psk" provider:"secret"`
	// Id is an optional identifier for this SAE PSK entry.
	Id *string `pulumi:"id,optional"`
	// Mac optionally binds this key to a specific client MAC (XX:XX:XX:XX:XX:XX).
	Mac *string `pulumi:"mac,optional"`
	// Vlan optionally places clients using this key onto the given VLAN ID.
	Vlan *int `pulumi:"vlan,optional"`
}

func (d *WlanSaePsk) Annotate(a infer.Annotator) {
	a.Describe(&d.Psk, "Psk is the SAE pre-shared key (8-255 chars). Secret.")
	a.Describe(&d.Id, "Id is an optional identifier for this SAE PSK entry.")
	a.Describe(&d.Mac, "Mac optionally binds this key to a specific client MAC (XX:XX:XX:XX:XX:XX).")
	a.Describe(&d.Vlan, "Vlan optionally places clients using this key onto the given VLAN ID.")
}

// WlanScheduleEntry is a single entry of the duration-based broadcast schedule.
type WlanScheduleEntry struct {
	// DurationMinutes is how long, in minutes, the SSID stays active once started.
	DurationMinutes int `pulumi:"durationMinutes"`
	// StartHour is the start hour (0-23).
	StartHour int `pulumi:"startHour"`
	// StartMinute is the start minute (0-59). Defaults to 0.
	StartMinute *int `pulumi:"startMinute,optional"`
	// StartDaysOfWeek selects the days this entry applies to: sun|mon|tue|wed|thu|fri|sat.
	StartDaysOfWeek []string `pulumi:"startDaysOfWeek,optional"`
	// Name is an optional friendly label for the schedule entry.
	Name *string `pulumi:"name,optional"`
}

func (d *WlanScheduleEntry) Annotate(a infer.Annotator) {
	a.Describe(&d.DurationMinutes, "DurationMinutes is how long, in minutes, the SSID stays active once started.")
	a.Describe(&d.StartHour, "StartHour is the start hour (0-23).")
	a.Describe(&d.StartMinute, "StartMinute is the start minute (0-59). Defaults to 0.")
	a.Describe(&d.StartDaysOfWeek, "StartDaysOfWeek selects the days this entry applies to: sun|mon|tue|wed|thu|fri|sat.")
	a.Describe(&d.Name, "Name is an optional friendly label for the schedule entry.")
}

// WlanWpa groups the core WPA/encryption tuning applied on top of `security`.
type WlanWpa struct {
	// Mode is the WPA mode: auto | wpa1 | wpa2.
	Mode *string `pulumi:"mode,optional"`
	// Enc is the WPA encryption cipher: auto | ccmp | gcmp | ccmp-256 | gcmp-256.
	Enc *string `pulumi:"enc,optional"`
	// PskRadius controls RADIUS PSK auth: disabled | optional | required.
	// This is a mode enum, not a credential, so it is not marked secret.
	PskRadius *string `pulumi:"pskRadius,optional"`
	// PmfMode is Protected Management Frames mode: disabled | optional | required.
	PmfMode *string `pulumi:"pmfMode,optional"`
	// PmfCipher is the PMF cipher: auto | aes-128-cmac | bip-gmac-256.
	PmfCipher *string `pulumi:"pmfCipher,optional"`
	// GroupRekey is the group key rekey interval in seconds (0 disables).
	GroupRekey *int `pulumi:"groupRekey,optional"`
}

func (d *WlanWpa) Annotate(a infer.Annotator) {
	a.Describe(&d.Mode, "Mode is the WPA mode: auto | wpa1 | wpa2.")
	a.Describe(&d.Enc, "Enc is the WPA encryption cipher: auto | ccmp | gcmp | ccmp-256 | gcmp-256.")
	a.Describe(&d.PskRadius, "PskRadius controls RADIUS PSK auth: disabled | optional | required. This is a mode enum, not a credential, so it is not marked secret.")
	a.Describe(&d.PmfMode, "PmfMode is Protected Management Frames mode: disabled | optional | required.")
	a.Describe(&d.PmfCipher, "PmfCipher is the PMF cipher: auto | aes-128-cmac | bip-gmac-256.")
	a.Describe(&d.GroupRekey, "GroupRekey is the group key rekey interval in seconds (0 disables).")
}

// WlanWpa3 groups the WPA3 feature cluster.
type WlanWpa3 struct {
	// Support enables WPA3 (requires wpapsk security and PMF enabled).
	Support *bool `pulumi:"support,optional"`
	// Transition enables WPA3/WPA2 transition mode (requires support).
	Transition *bool `pulumi:"transition,optional"`
	// Enhanced192 enables WPA3 Enterprise 192-bit mode.
	Enhanced192 *bool `pulumi:"enhanced192,optional"`
	// FastRoaming enables 802.11r fast roaming for WPA3.
	FastRoaming *bool `pulumi:"fastRoaming,optional"`
}

func (d *WlanWpa3) Annotate(a infer.Annotator) {
	a.Describe(&d.Support, "Support enables WPA3 (requires wpapsk security and PMF enabled).")
	a.Describe(&d.Transition, "Transition enables WPA3/WPA2 transition mode (requires support).")
	a.Describe(&d.Enhanced192, "Enhanced192 enables WPA3 Enterprise 192-bit mode.")
	a.Describe(&d.FastRoaming, "FastRoaming enables 802.11r fast roaming for WPA3.")
}

// WlanSae groups the WPA3 SAE handshake configuration.
type WlanSae struct {
	// Psks is the list of WPA3 SAE pre-shared keys.
	Psks []WlanSaePsk `pulumi:"psks,optional"`
	// Groups are the SAE finite cyclic groups to allow.
	Groups []int `pulumi:"groups,optional"`
	// AntiClogging is the SAE anti-clogging threshold.
	AntiClogging *int `pulumi:"antiClogging,optional"`
	// Sync is the SAE sync value.
	Sync *int `pulumi:"sync,optional"`
}

func (d *WlanSae) Annotate(a infer.Annotator) {
	a.Describe(&d.Psks, "Psks is the list of WPA3 SAE pre-shared keys.")
	a.Describe(&d.Groups, "Groups are the SAE finite cyclic groups to allow.")
	a.Describe(&d.AntiClogging, "AntiClogging is the SAE anti-clogging threshold.")
	a.Describe(&d.Sync, "Sync is the SAE sync value.")
}

// WlanWep groups the legacy WEP-only settings (used when security="wep").
type WlanWep struct {
	// Key is the WEP key. Secret. Only used when security is "wep".
	Key *string `pulumi:"key,optional" provider:"secret"`
	// Index is the WEP key index (1-4).
	Index *int `pulumi:"index,optional"`
}

func (d *WlanWep) Annotate(a infer.Annotator) {
	a.Describe(&d.Key, "Key is the WEP key. Secret. Only used when security is \"wep\".")
	a.Describe(&d.Index, "Index is the WEP key index (1-4).")
}

// WlanPrivatePresharedKeys groups per-key network placement (PPSK).
type WlanPrivatePresharedKeys struct {
	// Enabled enables per-key network placement.
	Enabled *bool `pulumi:"enabled,optional"`
	// Keys is the list of private pre-shared keys.
	Keys []WlanPrivatePsk `pulumi:"keys,optional"`
}

func (d *WlanPrivatePresharedKeys) Annotate(a infer.Annotator) {
	a.Describe(&d.Enabled, "Enabled enables per-key network placement.")
	a.Describe(&d.Keys, "Keys is the list of private pre-shared keys.")
}

// WlanRadius groups the RADIUS/802.1X plumbing for wpaeap and RADIUS-MAC auth.
type WlanRadius struct {
	// ProfileId is the RADIUS profile (`_id`) for wpaeap security.
	ProfileId *string `pulumi:"profileId,optional"`
	// MacAuthEnabled enables RADIUS-based MAC authentication.
	MacAuthEnabled *bool `pulumi:"macAuthEnabled,optional"`
	// MacaclFormat is the MAC ACL format: none_lower | hyphen_lower |
	// colon_lower | none_upper | hyphen_upper | colon_upper.
	MacaclFormat *string `pulumi:"macaclFormat,optional"`
	// MacaclEmptyPassword sends an empty password for MAC ACL auth.
	MacaclEmptyPassword *bool `pulumi:"macaclEmptyPassword,optional"`
	// DasEnabled enables RADIUS Dynamic Authorization (CoA/DM).
	DasEnabled *bool `pulumi:"dasEnabled,optional"`
	// NasIdentifier is the RADIUS NAS identifier value (0-48 chars).
	NasIdentifier *string `pulumi:"nasIdentifier,optional"`
	// NasIdentifierType: ap_name | ap_mac | bssid | site_name | custom.
	NasIdentifierType *string `pulumi:"nasIdentifierType,optional"`
}

func (d *WlanRadius) Annotate(a infer.Annotator) {
	a.Describe(&d.ProfileId, "ProfileId is the RADIUS profile (`_id`) for wpaeap security.")
	a.Describe(&d.MacAuthEnabled, "MacAuthEnabled enables RADIUS-based MAC authentication.")
	a.Describe(&d.MacaclFormat, "MacaclFormat is the MAC ACL format: none_lower | hyphen_lower | colon_lower | none_upper | hyphen_upper | colon_upper.")
	a.Describe(&d.MacaclEmptyPassword, "MacaclEmptyPassword sends an empty password for MAC ACL auth.")
	a.Describe(&d.DasEnabled, "DasEnabled enables RADIUS Dynamic Authorization (CoA/DM).")
	a.Describe(&d.NasIdentifier, "NasIdentifier is the RADIUS NAS identifier value (0-48 chars).")
	a.Describe(&d.NasIdentifierType, "NasIdentifierType: ap_name | ap_mac | bssid | site_name | custom.")
}

// WlanVlanTagging groups the VLAN tagging facet.
type WlanVlanTagging struct {
	// Vlan is the VLAN ID to tag client traffic with.
	Vlan *int `pulumi:"vlan,optional"`
	// Enabled enables VLAN tagging for this WLAN.
	Enabled *bool `pulumi:"enabled,optional"`
}

func (d *WlanVlanTagging) Annotate(a infer.Annotator) {
	a.Describe(&d.Vlan, "Vlan is the VLAN ID to tag client traffic with.")
	a.Describe(&d.Enabled, "Enabled enables VLAN tagging for this WLAN.")
}

// WlanBandSteering groups radio band selection and band steering.
type WlanBandSteering struct {
	// Band is the radio band: 2g | 5g | both.
	Band *string `pulumi:"band,optional"`
	// Bands are the radio bands to broadcast on: 2g | 5g | 6g.
	Bands []string `pulumi:"bands,optional"`
	// No2GhzOui steers high-performance clients to 5GHz only.
	No2GhzOui *bool `pulumi:"no2GhzOui,optional"`
}

func (d *WlanBandSteering) Annotate(a infer.Annotator) {
	a.Describe(&d.Band, "Band is the radio band: 2g | 5g | both.")
	a.Describe(&d.Bands, "Bands are the radio bands to broadcast on: 2g | 5g | 6g.")
	a.Describe(&d.No2GhzOui, "No2GhzOui steers high-performance clients to 5GHz only.")
}

// WlanDtim groups the DTIM interval control across bands.
type WlanDtim struct {
	// Mode controls DTIM interval handling: default | custom.
	Mode *string `pulumi:"mode,optional"`
	// Na is the DTIM interval for the 5GHz band (1-255).
	Na *int `pulumi:"na,optional"`
	// Ng is the DTIM interval for the 2.4GHz band (1-255).
	Ng *int `pulumi:"ng,optional"`
	// SixE is the DTIM interval for the 6GHz band (1-255).
	SixE *int `pulumi:"sixE,optional"`
}

func (d *WlanDtim) Annotate(a infer.Annotator) {
	a.Describe(&d.Mode, "Mode controls DTIM interval handling: default | custom.")
	a.Describe(&d.Na, "Na is the DTIM interval for the 5GHz band (1-255).")
	a.Describe(&d.Ng, "Ng is the DTIM interval for the 2.4GHz band (1-255).")
	a.Describe(&d.SixE, "SixE is the DTIM interval for the 6GHz band (1-255).")
}

// WlanMinrate groups the minimum-data-rate control.
type WlanMinrate struct {
	// NaEnabled enables the 5GHz minimum data rate control.
	NaEnabled *bool `pulumi:"naEnabled,optional"`
	// NaDataRateKbps is the minimum 5GHz data rate in Kbps.
	NaDataRateKbps *int `pulumi:"naDataRateKbps,optional"`
	// NaAdvertisingRates advertises only allowed 5GHz rates.
	NaAdvertisingRates *bool `pulumi:"naAdvertisingRates,optional"`
	// NgEnabled enables the 2.4GHz minimum data rate control.
	NgEnabled *bool `pulumi:"ngEnabled,optional"`
	// NgDataRateKbps is the minimum 2.4GHz data rate in Kbps.
	NgDataRateKbps *int `pulumi:"ngDataRateKbps,optional"`
	// NgAdvertisingRates advertises only allowed 2.4GHz rates.
	NgAdvertisingRates *bool `pulumi:"ngAdvertisingRates,optional"`
	// SettingPreference: auto | manual.
	SettingPreference *string `pulumi:"settingPreference,optional"`
}

func (d *WlanMinrate) Annotate(a infer.Annotator) {
	a.Describe(&d.NaEnabled, "NaEnabled enables the 5GHz minimum data rate control.")
	a.Describe(&d.NaDataRateKbps, "NaDataRateKbps is the minimum 5GHz data rate in Kbps.")
	a.Describe(&d.NaAdvertisingRates, "NaAdvertisingRates advertises only allowed 5GHz rates.")
	a.Describe(&d.NgEnabled, "NgEnabled enables the 2.4GHz minimum data rate control.")
	a.Describe(&d.NgDataRateKbps, "NgDataRateKbps is the minimum 2.4GHz data rate in Kbps.")
	a.Describe(&d.NgAdvertisingRates, "NgAdvertisingRates advertises only allowed 2.4GHz rates.")
	a.Describe(&d.SettingPreference, "SettingPreference: auto | manual.")
}

// WlanMacFilter groups the MAC access-control list.
type WlanMacFilter struct {
	// Enabled enables MAC-based access control.
	Enabled *bool `pulumi:"enabled,optional"`
	// Policy: allow | deny.
	Policy *string `pulumi:"policy,optional"`
	// List is the list of MACs (XX:XX:XX:XX:XX:XX) the policy applies to.
	List []string `pulumi:"list,optional"`
}

func (d *WlanMacFilter) Annotate(a infer.Annotator) {
	a.Describe(&d.Enabled, "Enabled enables MAC-based access control.")
	a.Describe(&d.Policy, "Policy: allow | deny.")
	a.Describe(&d.List, "List is the list of MACs (XX:XX:XX:XX:XX:XX) the policy applies to.")
}

// WlanMulticast groups the multicast/broadcast traffic handling.
type WlanMulticast struct {
	// EnhanceEnabled converts multicast to unicast for reliability.
	EnhanceEnabled *bool `pulumi:"enhanceEnabled,optional"`
	// ProxyArp lets APs proxy common broadcast frames as unicast.
	ProxyArp *bool `pulumi:"proxyArp,optional"`
	// BroadcastFilterEnabled enables filtering of broadcast/multicast traffic.
	BroadcastFilterEnabled *bool `pulumi:"broadcastFilterEnabled,optional"`
	// BroadcastFilterList is the allow list of MACs exempt from filtering.
	BroadcastFilterList []string `pulumi:"broadcastFilterList,optional"`
}

func (d *WlanMulticast) Annotate(a infer.Annotator) {
	a.Describe(&d.EnhanceEnabled, "EnhanceEnabled converts multicast to unicast for reliability.")
	a.Describe(&d.ProxyArp, "ProxyArp lets APs proxy common broadcast frames as unicast.")
	a.Describe(&d.BroadcastFilterEnabled, "BroadcastFilterEnabled enables filtering of broadcast/multicast traffic.")
	a.Describe(&d.BroadcastFilterList, "BroadcastFilterList is the allow list of MACs exempt from filtering.")
}

// WlanSchedule groups the time-based broadcast scheduling.
type WlanSchedule struct {
	// Enabled enables time-based broadcast scheduling.
	Enabled *bool `pulumi:"enabled,optional"`
	// Legacy is the legacy schedule format entries (day|HHMM-HHMM).
	Legacy []string `pulumi:"legacy,optional"`
	// Entries is the duration-based broadcast schedule.
	Entries []WlanScheduleEntry `pulumi:"entries,optional"`
}

func (d *WlanSchedule) Annotate(a infer.Annotator) {
	a.Describe(&d.Enabled, "Enabled enables time-based broadcast scheduling.")
	a.Describe(&d.Legacy, "Legacy is the legacy schedule format entries (day|HHMM-HHMM).")
	a.Describe(&d.Entries, "Entries is the duration-based broadcast schedule.")
}

// WlanApGroups groups which AP groups/devices broadcast this SSID.
type WlanApGroups struct {
	// Ids are the AP groups that should broadcast this SSID.
	Ids []string `pulumi:"ids,optional"`
	// Mode controls AP selection: all | groups | devices.
	Mode *string `pulumi:"mode,optional"`
}

func (d *WlanApGroups) Annotate(a infer.Annotator) {
	a.Describe(&d.Ids, "Ids are the AP groups that should broadcast this SSID.")
	a.Describe(&d.Mode, "Mode controls AP selection: all | groups | devices.")
}

// WlanDpi groups the Deep Packet Inspection toggle and group reference.
type WlanDpi struct {
	// Enabled enables deep packet inspection for this WLAN.
	Enabled *bool `pulumi:"enabled,optional"`
	// GroupId is the DPI group to apply.
	GroupId *string `pulumi:"groupId,optional"`
}

func (d *WlanDpi) Annotate(a infer.Annotator) {
	a.Describe(&d.Enabled, "Enabled enables deep packet inspection for this WLAN.")
	a.Describe(&d.GroupId, "GroupId is the DPI group to apply.")
}

// WlanIot groups the IoT connectivity behaviors.
type WlanIot struct {
	// Enhanced enables enhanced IoT connectivity behaviors.
	Enhanced *bool `pulumi:"enhanced,optional"`
	// OptimizeWifiConnectivity optimizes connectivity for IoT devices.
	OptimizeWifiConnectivity *bool `pulumi:"optimizeWifiConnectivity,optional"`
}

func (d *WlanIot) Annotate(a infer.Annotator) {
	a.Describe(&d.Enhanced, "Enhanced enables enhanced IoT connectivity behaviors.")
	a.Describe(&d.OptimizeWifiConnectivity, "OptimizeWifiConnectivity optimizes connectivity for IoT devices.")
}

// WlanP2p groups the peer-to-peer (client-to-client) traffic settings.
type WlanP2p struct {
	// Enabled enables peer-to-peer (client-to-client) traffic.
	Enabled *bool `pulumi:"enabled,optional"`
	// CrossConnect allows P2P traffic across APs.
	CrossConnect *bool `pulumi:"crossConnect,optional"`
}

func (d *WlanP2p) Annotate(a infer.Annotator) {
	a.Describe(&d.Enabled, "Enabled enables peer-to-peer (client-to-client) traffic.")
	a.Describe(&d.CrossConnect, "CrossConnect allows P2P traffic across APs.")
}

// WlanRoaming groups the general client roaming/handoff settings.
type WlanRoaming struct {
	// FastRoamingEnabled enables 802.11r fast BSS transition.
	FastRoamingEnabled *bool `pulumi:"fastRoamingEnabled,optional"`
	// BssTransition enables 802.11v BSS transition management.
	BssTransition *bool `pulumi:"bssTransition,optional"`
	// IappKey is the inter-AP protocol key (32 hex chars). Secret.
	IappKey *string `pulumi:"iappKey,optional" provider:"secret"`
}

func (d *WlanRoaming) Annotate(a infer.Annotator) {
	a.Describe(&d.FastRoamingEnabled, "FastRoamingEnabled enables 802.11r fast BSS transition.")
	a.Describe(&d.BssTransition, "BssTransition enables 802.11v BSS transition management.")
	a.Describe(&d.IappKey, "IappKey is the inter-AP protocol key (32 hex chars). Secret.")
}

// WlanArgs are the user-supplied inputs for a WLAN.
type WlanArgs struct {
	// Name is the SSID.
	Name string `pulumi:"name"`
	// NetworkId binds the WLAN to a network/VLAN (the network's `_id`).
	NetworkId string `pulumi:"networkId"`
	// WlanGroupId is the WLAN group to attach to. Required on many controllers.
	WlanGroupId *string `pulumi:"wlanGroupId,optional"`
	// UserGroupId is the user group (rate limiting / firewall) for clients.
	UserGroupId *string `pulumi:"userGroupId,optional"`
	// Enabled controls whether the SSID is broadcast. Defaults to true.
	Enabled *bool `pulumi:"enabled,optional"`
	// Security: open | wpapsk | wpaeap | wep | osen. Defaults to "wpapsk".
	Security *string `pulumi:"security,optional"`
	// Passphrase is the WPA pre-shared key (8-63 chars). Secret.
	Passphrase *string `pulumi:"passphrase,optional" provider:"secret"`
	// HideSsid hides the SSID from broadcast. Defaults to false.
	HideSsid *bool `pulumi:"hideSsid,optional"`
	// UapsdEnabled enables Unscheduled Automatic Power Save Delivery.
	UapsdEnabled *bool `pulumi:"uapsdEnabled,optional"`
	// MloEnabled enables Multi-Link Operation (WiFi 7).
	MloEnabled *bool `pulumi:"mloEnabled,optional"`
	// L2Isolation isolates wireless clients from each other at layer 2.
	L2Isolation *bool `pulumi:"l2Isolation,optional"`
	// IsGuest marks this WLAN as a guest network (enables guest behaviors).
	IsGuest *bool `pulumi:"isGuest,optional"`
	// Priority: medium | high | low.
	Priority *string `pulumi:"priority,optional"`

	// Wpa groups the core WPA/encryption tuning.
	Wpa *WlanWpa `pulumi:"wpa,optional"`
	// Wpa3 groups the WPA3 feature cluster.
	Wpa3 *WlanWpa3 `pulumi:"wpa3,optional"`
	// Sae groups the WPA3 SAE handshake configuration.
	Sae *WlanSae `pulumi:"sae,optional"`
	// Wep groups the legacy WEP-only settings.
	Wep *WlanWep `pulumi:"wep,optional"`
	// PrivatePresharedKeys groups per-key network placement (PPSK).
	PrivatePresharedKeys *WlanPrivatePresharedKeys `pulumi:"privatePresharedKeys,optional"`
	// Radius groups the RADIUS/802.1X plumbing.
	Radius *WlanRadius `pulumi:"radius,optional"`
	// VlanTagging groups the VLAN tagging facet.
	VlanTagging *WlanVlanTagging `pulumi:"vlanTagging,optional"`
	// BandSteering groups radio band selection and band steering.
	BandSteering *WlanBandSteering `pulumi:"bandSteering,optional"`
	// Dtim groups the DTIM interval control across bands.
	Dtim *WlanDtim `pulumi:"dtim,optional"`
	// MinRate groups the minimum-data-rate control.
	MinRate *WlanMinrate `pulumi:"minRate,optional"`
	// MacFilter groups the MAC access-control list.
	MacFilter *WlanMacFilter `pulumi:"macFilter,optional"`
	// Multicast groups the multicast/broadcast traffic handling.
	Multicast *WlanMulticast `pulumi:"multicast,optional"`
	// Schedule groups the time-based broadcast scheduling.
	Schedule *WlanSchedule `pulumi:"schedule,optional"`
	// ApGroups groups which AP groups/devices broadcast this SSID.
	ApGroups *WlanApGroups `pulumi:"apGroups,optional"`
	// Dpi groups the Deep Packet Inspection toggle and group reference.
	Dpi *WlanDpi `pulumi:"dpi,optional"`
	// Iot groups the IoT connectivity behaviors.
	Iot *WlanIot `pulumi:"iot,optional"`
	// P2p groups the peer-to-peer (client-to-client) traffic settings.
	P2p *WlanP2p `pulumi:"p2p,optional"`
	// Roaming groups the general client roaming/handoff settings.
	Roaming *WlanRoaming `pulumi:"roaming,optional"`
}

func (d *WlanArgs) Annotate(a infer.Annotator) {
	a.Describe(&d.Name, "Name is the SSID.")
	a.Describe(&d.NetworkId, "NetworkId binds the WLAN to a network/VLAN (the network's `_id`).")
	a.Describe(&d.WlanGroupId, "WlanGroupId is the WLAN group to attach to. Required on many controllers.")
	a.Describe(&d.UserGroupId, "UserGroupId is the user group (rate limiting / firewall) for clients.")
	a.Describe(&d.Enabled, "Enabled controls whether the SSID is broadcast. Defaults to true.")
	a.Describe(&d.Security, "Security: open | wpapsk | wpaeap | wep | osen. Defaults to \"wpapsk\".")
	a.Describe(&d.Passphrase, "Passphrase is the WPA pre-shared key (8-63 chars). Secret.")
	a.Describe(&d.HideSsid, "HideSsid hides the SSID from broadcast. Defaults to false.")
	a.Describe(&d.UapsdEnabled, "UapsdEnabled enables Unscheduled Automatic Power Save Delivery.")
	a.Describe(&d.MloEnabled, "MloEnabled enables Multi-Link Operation (WiFi 7).")
	a.Describe(&d.L2Isolation, "L2Isolation isolates wireless clients from each other at layer 2.")
	a.Describe(&d.IsGuest, "IsGuest marks this WLAN as a guest network (enables guest behaviors).")
	a.Describe(&d.Priority, "Priority: medium | high | low.")
	a.Describe(&d.Wpa, "Wpa groups the core WPA/encryption tuning.")
	a.Describe(&d.Wpa3, "Wpa3 groups the WPA3 feature cluster.")
	a.Describe(&d.Sae, "Sae groups the WPA3 SAE handshake configuration.")
	a.Describe(&d.Wep, "Wep groups the legacy WEP-only settings.")
	a.Describe(&d.PrivatePresharedKeys, "PrivatePresharedKeys groups per-key network placement (PPSK).")
	a.Describe(&d.Radius, "Radius groups the RADIUS/802.1X plumbing.")
	a.Describe(&d.VlanTagging, "VlanTagging groups the VLAN tagging facet.")
	a.Describe(&d.BandSteering, "BandSteering groups radio band selection and band steering.")
	a.Describe(&d.Dtim, "Dtim groups the DTIM interval control across bands.")
	a.Describe(&d.MinRate, "MinRate groups the minimum-data-rate control.")
	a.Describe(&d.MacFilter, "MacFilter groups the MAC access-control list.")
	a.Describe(&d.Multicast, "Multicast groups the multicast/broadcast traffic handling.")
	a.Describe(&d.Schedule, "Schedule groups the time-based broadcast scheduling.")
	a.Describe(&d.ApGroups, "ApGroups groups which AP groups/devices broadcast this SSID.")
	a.Describe(&d.Dpi, "Dpi groups the Deep Packet Inspection toggle and group reference.")
	a.Describe(&d.Iot, "Iot groups the IoT connectivity behaviors.")
	a.Describe(&d.P2p, "P2p groups the peer-to-peer (client-to-client) traffic settings.")
	a.Describe(&d.Roaming, "Roaming groups the general client roaming/handoff settings.")
}

// WlanState is the persisted state: inputs plus controller-assigned fields.
type WlanState struct {
	WlanArgs
	// WlanId is the controller-assigned identifier (the UniFi `_id`).
	WlanId string `pulumi:"wlanId"`
}

func (d *WlanState) Annotate(a infer.Annotator) {
	a.Describe(&d.WlanId, "WlanId is the controller-assigned identifier (the UniFi `_id`).")
}

// Annotate documents the resource and its most important fields.
func (w *Wlan) Annotate(a infer.Annotator) {
	a.Describe(&w, "A UniFi wireless network (SSID). Maps to a controller wlanconf object. "+
		"Security, RADIUS, roaming, band/rate, scheduling, and other settings are grouped into "+
		"nested objects (wpa, wpa3, sae, wep, radius, roaming, etc.).")
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
	if a.UserGroupId != nil {
		w.UserGroupID = *a.UserGroupId
	}
	if a.Passphrase != nil {
		w.XPassphrase = *a.Passphrase
	}
	if a.UapsdEnabled != nil {
		w.UapsdEnabled = *a.UapsdEnabled
	}
	if a.MloEnabled != nil {
		w.MloEnabled = *a.MloEnabled
	}
	if a.L2Isolation != nil {
		w.L2Isolation = *a.L2Isolation
	}
	if a.IsGuest != nil {
		w.IsGuest = *a.IsGuest
	}
	if a.Priority != nil {
		w.Priority = *a.Priority
	}

	// WPA / encryption.
	if g := a.Wpa; g != nil {
		if g.Mode != nil {
			w.WPAMode = *g.Mode
		}
		if g.Enc != nil {
			w.WPAEnc = *g.Enc
		}
		if g.PskRadius != nil {
			w.WPAPskRADIUS = *g.PskRadius
		}
		if g.PmfMode != nil {
			w.PMFMode = *g.PmfMode
		}
		if g.PmfCipher != nil {
			w.PMFCipher = *g.PmfCipher
		}
		if g.GroupRekey != nil {
			w.GroupRekey = *g.GroupRekey
		}
	}

	// WPA3.
	if g := a.Wpa3; g != nil {
		if g.Support != nil {
			w.WPA3Support = *g.Support
		}
		if g.Transition != nil {
			w.WPA3Transition = *g.Transition
		}
		if g.Enhanced192 != nil {
			w.WPA3Enhanced192 = *g.Enhanced192
		}
		if g.FastRoaming != nil {
			w.WPA3FastRoaming = *g.FastRoaming
		}
	}

	// SAE.
	if g := a.Sae; g != nil {
		if g.Psks != nil {
			out := make([]unifi.WLANSaePsk, len(g.Psks))
			for i, p := range g.Psks {
				e := unifi.WLANSaePsk{Psk: p.Psk}
				if p.Id != nil {
					e.ID = *p.Id
				}
				if p.Mac != nil {
					e.MAC = *p.Mac
				}
				if p.Vlan != nil {
					e.VLAN = *p.Vlan
				}
				out[i] = e
			}
			w.SaePsk = out
		}
		if g.Groups != nil {
			w.SaeGroups = g.Groups
		}
		if g.AntiClogging != nil {
			w.SaeAntiClogging = *g.AntiClogging
		}
		if g.Sync != nil {
			w.SaeSync = *g.Sync
		}
	}

	// WEP.
	if g := a.Wep; g != nil {
		if g.Key != nil {
			w.XWEP = *g.Key
		}
		if g.Index != nil {
			w.WEPIDX = *g.Index
		}
	}

	// Private pre-shared keys (PPSK).
	if g := a.PrivatePresharedKeys; g != nil {
		if g.Enabled != nil {
			w.PrivatePresharedKeysEnabled = *g.Enabled
		}
		if g.Keys != nil {
			out := make([]unifi.WLANPrivatePresharedKeys, len(g.Keys))
			for i, p := range g.Keys {
				out[i] = unifi.WLANPrivatePresharedKeys{
					Password:  p.Password,
					NetworkID: derefOr(p.NetworkId, ""),
				}
			}
			w.PrivatePresharedKeys = out
		}
	}

	// RADIUS.
	if g := a.Radius; g != nil {
		if g.ProfileId != nil {
			w.RADIUSProfileID = *g.ProfileId
		}
		if g.MacAuthEnabled != nil {
			w.RADIUSMACAuthEnabled = *g.MacAuthEnabled
		}
		if g.MacaclFormat != nil {
			w.RADIUSMACaclFormat = *g.MacaclFormat
		}
		if g.MacaclEmptyPassword != nil {
			w.RADIUSMACaclEmptyPassword = *g.MacaclEmptyPassword
		}
		if g.DasEnabled != nil {
			w.RADIUSDasEnabled = *g.DasEnabled
		}
		if g.NasIdentifier != nil {
			w.NasIDentifier = *g.NasIdentifier
		}
		if g.NasIdentifierType != nil {
			w.NasIDentifierType = *g.NasIdentifierType
		}
	}

	// VLAN tagging.
	if g := a.VlanTagging; g != nil {
		if g.Vlan != nil {
			w.VLAN = *g.Vlan
		}
		if g.Enabled != nil {
			w.VLANEnabled = *g.Enabled
		}
	}

	// Band steering.
	if g := a.BandSteering; g != nil {
		if g.Band != nil {
			w.WLANBand = *g.Band
		}
		if g.Bands != nil {
			w.WLANBands = g.Bands
		}
		if g.No2GhzOui != nil {
			w.No2GhzOui = *g.No2GhzOui
		}
	}

	// DTIM.
	if g := a.Dtim; g != nil {
		if g.Mode != nil {
			w.DTIMMode = *g.Mode
		}
		if g.Na != nil {
			w.DTIMNa = *g.Na
		}
		if g.Ng != nil {
			w.DTIMNg = *g.Ng
		}
		if g.SixE != nil {
			w.DTIM6E = *g.SixE
		}
	}

	// Minimum data rate.
	if g := a.MinRate; g != nil {
		if g.NaEnabled != nil {
			w.MinrateNaEnabled = *g.NaEnabled
		}
		if g.NaDataRateKbps != nil {
			w.MinrateNaDataRateKbps = *g.NaDataRateKbps
		}
		if g.NaAdvertisingRates != nil {
			w.MinrateNaAdvertisingRates = *g.NaAdvertisingRates
		}
		if g.NgEnabled != nil {
			w.MinrateNgEnabled = *g.NgEnabled
		}
		if g.NgDataRateKbps != nil {
			w.MinrateNgDataRateKbps = *g.NgDataRateKbps
		}
		if g.NgAdvertisingRates != nil {
			w.MinrateNgAdvertisingRates = *g.NgAdvertisingRates
		}
		if g.SettingPreference != nil {
			w.MinrateSettingPreference = *g.SettingPreference
		}
	}

	// MAC filter.
	if g := a.MacFilter; g != nil {
		if g.Enabled != nil {
			w.MACFilterEnabled = *g.Enabled
		}
		if g.Policy != nil {
			w.MACFilterPolicy = *g.Policy
		}
		if g.List != nil {
			w.MACFilterList = g.List
		}
	}

	// Multicast / broadcast.
	if g := a.Multicast; g != nil {
		if g.EnhanceEnabled != nil {
			w.MulticastEnhanceEnabled = *g.EnhanceEnabled
		}
		if g.ProxyArp != nil {
			w.ProxyArp = *g.ProxyArp
		}
		if g.BroadcastFilterEnabled != nil {
			w.BroadcastFilterEnabled = *g.BroadcastFilterEnabled
		}
		if g.BroadcastFilterList != nil {
			w.BroadcastFilterList = g.BroadcastFilterList
		}
	}

	// Schedule.
	if g := a.Schedule; g != nil {
		if g.Enabled != nil {
			w.ScheduleEnabled = *g.Enabled
		}
		if g.Legacy != nil {
			w.Schedule = g.Legacy
		}
		if g.Entries != nil {
			out := make([]unifi.WLANScheduleWithDuration, len(g.Entries))
			for i, s := range g.Entries {
				e := unifi.WLANScheduleWithDuration{
					DurationMinutes: s.DurationMinutes,
					StartHour:       s.StartHour,
				}
				if s.StartMinute != nil {
					e.StartMinute = *s.StartMinute
				}
				if s.Name != nil {
					e.Name = *s.Name
				}
				if s.StartDaysOfWeek != nil {
					e.StartDaysOfWeek = s.StartDaysOfWeek
				}
				out[i] = e
			}
			w.ScheduleWithDuration = out
		}
	}

	// AP groups.
	if g := a.ApGroups; g != nil {
		if g.Ids != nil {
			w.ApGroupIDs = g.Ids
		}
		if g.Mode != nil {
			w.ApGroupMode = *g.Mode
		}
	}

	// DPI.
	if g := a.Dpi; g != nil {
		if g.Enabled != nil {
			w.DPIEnabled = *g.Enabled
		}
		if g.GroupId != nil {
			w.DPIgroupID = *g.GroupId
		}
	}

	// IoT.
	if g := a.Iot; g != nil {
		if g.Enhanced != nil {
			w.EnhancedIot = *g.Enhanced
		}
		if g.OptimizeWifiConnectivity != nil {
			w.OptimizeIotWifiConnectivity = *g.OptimizeWifiConnectivity
		}
	}

	// Peer-to-peer.
	if g := a.P2p; g != nil {
		if g.Enabled != nil {
			w.P2P = *g.Enabled
		}
		if g.CrossConnect != nil {
			w.P2PCrossConnect = *g.CrossConnect
		}
	}

	// Roaming.
	if g := a.Roaming; g != nil {
		if g.FastRoamingEnabled != nil {
			w.FastRoamingEnabled = *g.FastRoamingEnabled
		}
		if g.BssTransition != nil {
			w.BssTransition = *g.BssTransition
		}
		if g.IappKey != nil {
			w.XIappKey = *g.IappKey
		}
	}

	return w
}

// wlanBoolState reflects a controller bool, preferring a true controller value
// but falling back to the prior input when the controller reports false.
func wlanBoolState(v bool, prior *bool) *bool {
	if v {
		return ptr(v)
	}
	return prior
}

// wlanStringState reflects a controller string, falling back to the prior input
// when the controller reports an empty value.
func wlanStringState(v string, prior *string) *string {
	if v != "" {
		return ptr(v)
	}
	return prior
}

// wlanIntState reflects a controller int, falling back to the prior input when
// the controller reports zero.
func wlanIntState(v int, prior *int) *int {
	if v != 0 {
		return ptr(v)
	}
	return prior
}

// wlanStringsState reflects a controller string slice, falling back to the prior
// input when the controller returns nothing.
func wlanStringsState(v []string, prior []string) []string {
	if len(v) > 0 {
		return v
	}
	return prior
}

// wlanIntsState reflects a controller int slice, falling back to the prior input
// when the controller returns nothing.
func wlanIntsState(v []int, prior []int) []int {
	if len(v) > 0 {
		return v
	}
	return prior
}

// wlanPsksState reflects the controller's private PSK list, preserving the
// user-provided (secret) password by index when the controller masks it.
func wlanPsksState(u []unifi.WLANPrivatePresharedKeys, prior []WlanPrivatePsk) []WlanPrivatePsk {
	if len(u) == 0 {
		return prior
	}
	out := make([]WlanPrivatePsk, len(u))
	for i, p := range u {
		pw := p.Password
		if pw == "" && i < len(prior) {
			pw = prior[i].Password
		}
		out[i] = WlanPrivatePsk{Password: pw}
		if p.NetworkID != "" {
			out[i].NetworkId = ptr(p.NetworkID)
		}
	}
	return out
}

// wlanSaePsksState reflects the controller's SAE PSK list, preserving the
// user-provided (secret) key by index when the controller masks it.
func wlanSaePsksState(u []unifi.WLANSaePsk, prior []WlanSaePsk) []WlanSaePsk {
	if len(u) == 0 {
		return prior
	}
	out := make([]WlanSaePsk, len(u))
	for i, p := range u {
		psk := p.Psk
		if psk == "" && i < len(prior) {
			psk = prior[i].Psk
		}
		e := WlanSaePsk{Psk: psk}
		if p.ID != "" {
			e.Id = ptr(p.ID)
		}
		if p.MAC != "" {
			e.Mac = ptr(p.MAC)
		}
		if p.VLAN != 0 {
			e.Vlan = ptr(p.VLAN)
		}
		out[i] = e
	}
	return out
}

// wlanSchedulesState reflects the controller's duration-based schedule.
func wlanSchedulesState(u []unifi.WLANScheduleWithDuration, prior []WlanScheduleEntry) []WlanScheduleEntry {
	if len(u) == 0 {
		return prior
	}
	out := make([]WlanScheduleEntry, len(u))
	for i, s := range u {
		e := WlanScheduleEntry{
			DurationMinutes: s.DurationMinutes,
			StartHour:       s.StartHour,
		}
		if s.StartMinute != 0 {
			e.StartMinute = ptr(s.StartMinute)
		}
		if s.Name != "" {
			e.Name = ptr(s.Name)
		}
		if len(s.StartDaysOfWeek) > 0 {
			e.StartDaysOfWeek = s.StartDaysOfWeek
		}
		out[i] = e
	}
	return out
}

// isZero reports whether no SAE member is set (so the group round-trips as nil).
func (g WlanSae) isZero() bool {
	return g.Psks == nil && g.Groups == nil && g.AntiClogging == nil && g.Sync == nil
}

// isZero reports whether no PPSK member is set (so the group round-trips as nil).
func (g WlanPrivatePresharedKeys) isZero() bool {
	return g.Enabled == nil && g.Keys == nil
}

// isZero reports whether no band-steering member is set.
func (g WlanBandSteering) isZero() bool {
	return g.Band == nil && g.Bands == nil && g.No2GhzOui == nil
}

// isZero reports whether no MAC-filter member is set.
func (g WlanMacFilter) isZero() bool {
	return g.Enabled == nil && g.Policy == nil && g.List == nil
}

// isZero reports whether no multicast member is set.
func (g WlanMulticast) isZero() bool {
	return g.EnhanceEnabled == nil && g.ProxyArp == nil &&
		g.BroadcastFilterEnabled == nil && g.BroadcastFilterList == nil
}

// isZero reports whether no schedule member is set.
func (g WlanSchedule) isZero() bool {
	return g.Enabled == nil && g.Legacy == nil && g.Entries == nil
}

// isZero reports whether no AP-groups member is set.
func (g WlanApGroups) isZero() bool {
	return g.Ids == nil && g.Mode == nil
}

// wlanWpaFrom reconstructs the wpa group, returning nil when no member is set.
func wlanWpaFrom(w *unifi.WLAN, prior *WlanWpa) *WlanWpa {
	var p WlanWpa
	if prior != nil {
		p = *prior
	}
	g := WlanWpa{
		Mode:       wlanStringState(w.WPAMode, p.Mode),
		Enc:        wlanStringState(w.WPAEnc, p.Enc),
		PskRadius:  wlanStringState(w.WPAPskRADIUS, p.PskRadius),
		PmfMode:    wlanStringState(w.PMFMode, p.PmfMode),
		PmfCipher:  wlanStringState(w.PMFCipher, p.PmfCipher),
		GroupRekey: wlanIntState(w.GroupRekey, p.GroupRekey),
	}
	if g == (WlanWpa{}) {
		return nil
	}
	return &g
}

// wlanWpa3From reconstructs the wpa3 group, returning nil when no member is set.
func wlanWpa3From(w *unifi.WLAN, prior *WlanWpa3) *WlanWpa3 {
	var p WlanWpa3
	if prior != nil {
		p = *prior
	}
	g := WlanWpa3{
		Support:     wlanBoolState(w.WPA3Support, p.Support),
		Transition:  wlanBoolState(w.WPA3Transition, p.Transition),
		Enhanced192: wlanBoolState(w.WPA3Enhanced192, p.Enhanced192),
		FastRoaming: wlanBoolState(w.WPA3FastRoaming, p.FastRoaming),
	}
	if g == (WlanWpa3{}) {
		return nil
	}
	return &g
}

// wlanSaeFrom reconstructs the sae group, returning nil when no member is set.
func wlanSaeFrom(w *unifi.WLAN, prior *WlanSae) *WlanSae {
	var p WlanSae
	if prior != nil {
		p = *prior
	}
	g := WlanSae{
		Psks:         wlanSaePsksState(w.SaePsk, p.Psks),
		Groups:       wlanIntsState(w.SaeGroups, p.Groups),
		AntiClogging: wlanIntState(w.SaeAntiClogging, p.AntiClogging),
		Sync:         wlanIntState(w.SaeSync, p.Sync),
	}
	if g.isZero() {
		return nil
	}
	return &g
}

// wlanWepFrom reconstructs the wep group, returning nil when no member is set.
func wlanWepFrom(w *unifi.WLAN, prior *WlanWep) *WlanWep {
	var p WlanWep
	if prior != nil {
		p = *prior
	}
	g := WlanWep{
		Key:   wlanStringState(w.XWEP, p.Key),
		Index: wlanIntState(w.WEPIDX, p.Index),
	}
	if g == (WlanWep{}) {
		return nil
	}
	return &g
}

// wlanPrivatePresharedKeysFrom reconstructs the privatePresharedKeys group,
// returning nil when no member is set. Secrets are preserved by index.
func wlanPrivatePresharedKeysFrom(w *unifi.WLAN, prior *WlanPrivatePresharedKeys) *WlanPrivatePresharedKeys {
	var p WlanPrivatePresharedKeys
	if prior != nil {
		p = *prior
	}
	g := WlanPrivatePresharedKeys{
		Enabled: wlanBoolState(w.PrivatePresharedKeysEnabled, p.Enabled),
		Keys:    wlanPsksState(w.PrivatePresharedKeys, p.Keys),
	}
	if g.isZero() {
		return nil
	}
	return &g
}

// wlanRadiusFrom reconstructs the radius group, returning nil when no member is set.
func wlanRadiusFrom(w *unifi.WLAN, prior *WlanRadius) *WlanRadius {
	var p WlanRadius
	if prior != nil {
		p = *prior
	}
	g := WlanRadius{
		ProfileId:           wlanStringState(w.RADIUSProfileID, p.ProfileId),
		MacAuthEnabled:      wlanBoolState(w.RADIUSMACAuthEnabled, p.MacAuthEnabled),
		MacaclFormat:        wlanStringState(w.RADIUSMACaclFormat, p.MacaclFormat),
		MacaclEmptyPassword: wlanBoolState(w.RADIUSMACaclEmptyPassword, p.MacaclEmptyPassword),
		DasEnabled:          wlanBoolState(w.RADIUSDasEnabled, p.DasEnabled),
		NasIdentifier:       wlanStringState(w.NasIDentifier, p.NasIdentifier),
		NasIdentifierType:   wlanStringState(w.NasIDentifierType, p.NasIdentifierType),
	}
	if g == (WlanRadius{}) {
		return nil
	}
	return &g
}

// wlanVlanTaggingFrom reconstructs the vlanTagging group, returning nil when no
// member is set.
func wlanVlanTaggingFrom(w *unifi.WLAN, prior *WlanVlanTagging) *WlanVlanTagging {
	var p WlanVlanTagging
	if prior != nil {
		p = *prior
	}
	g := WlanVlanTagging{
		Vlan:    wlanIntState(w.VLAN, p.Vlan),
		Enabled: wlanBoolState(w.VLANEnabled, p.Enabled),
	}
	if g == (WlanVlanTagging{}) {
		return nil
	}
	return &g
}

// wlanBandSteeringFrom reconstructs the bandSteering group, returning nil when no
// member is set.
func wlanBandSteeringFrom(w *unifi.WLAN, prior *WlanBandSteering) *WlanBandSteering {
	var p WlanBandSteering
	if prior != nil {
		p = *prior
	}
	g := WlanBandSteering{
		Band:      wlanStringState(w.WLANBand, p.Band),
		Bands:     wlanStringsState(w.WLANBands, p.Bands),
		No2GhzOui: wlanBoolState(w.No2GhzOui, p.No2GhzOui),
	}
	if g.isZero() {
		return nil
	}
	return &g
}

// wlanDtimFrom reconstructs the dtim group, returning nil when no member is set.
func wlanDtimFrom(w *unifi.WLAN, prior *WlanDtim) *WlanDtim {
	var p WlanDtim
	if prior != nil {
		p = *prior
	}
	g := WlanDtim{
		Mode: wlanStringState(w.DTIMMode, p.Mode),
		Na:   wlanIntState(w.DTIMNa, p.Na),
		Ng:   wlanIntState(w.DTIMNg, p.Ng),
		SixE: wlanIntState(w.DTIM6E, p.SixE),
	}
	if g == (WlanDtim{}) {
		return nil
	}
	return &g
}

// wlanMinRateFrom reconstructs the minRate group, returning nil when no member is set.
func wlanMinRateFrom(w *unifi.WLAN, prior *WlanMinrate) *WlanMinrate {
	var p WlanMinrate
	if prior != nil {
		p = *prior
	}
	g := WlanMinrate{
		NaEnabled:          wlanBoolState(w.MinrateNaEnabled, p.NaEnabled),
		NaDataRateKbps:     wlanIntState(w.MinrateNaDataRateKbps, p.NaDataRateKbps),
		NaAdvertisingRates: wlanBoolState(w.MinrateNaAdvertisingRates, p.NaAdvertisingRates),
		NgEnabled:          wlanBoolState(w.MinrateNgEnabled, p.NgEnabled),
		NgDataRateKbps:     wlanIntState(w.MinrateNgDataRateKbps, p.NgDataRateKbps),
		NgAdvertisingRates: wlanBoolState(w.MinrateNgAdvertisingRates, p.NgAdvertisingRates),
		SettingPreference:  wlanStringState(w.MinrateSettingPreference, p.SettingPreference),
	}
	if g == (WlanMinrate{}) {
		return nil
	}
	return &g
}

// wlanMacFilterFrom reconstructs the macFilter group, returning nil when no
// member is set.
func wlanMacFilterFrom(w *unifi.WLAN, prior *WlanMacFilter) *WlanMacFilter {
	var p WlanMacFilter
	if prior != nil {
		p = *prior
	}
	g := WlanMacFilter{
		Enabled: wlanBoolState(w.MACFilterEnabled, p.Enabled),
		Policy:  wlanStringState(w.MACFilterPolicy, p.Policy),
		List:    wlanStringsState(w.MACFilterList, p.List),
	}
	if g.isZero() {
		return nil
	}
	return &g
}

// wlanMulticastFrom reconstructs the multicast group, returning nil when no
// member is set.
func wlanMulticastFrom(w *unifi.WLAN, prior *WlanMulticast) *WlanMulticast {
	var p WlanMulticast
	if prior != nil {
		p = *prior
	}
	g := WlanMulticast{
		EnhanceEnabled:         wlanBoolState(w.MulticastEnhanceEnabled, p.EnhanceEnabled),
		ProxyArp:               wlanBoolState(w.ProxyArp, p.ProxyArp),
		BroadcastFilterEnabled: wlanBoolState(w.BroadcastFilterEnabled, p.BroadcastFilterEnabled),
		BroadcastFilterList:    wlanStringsState(w.BroadcastFilterList, p.BroadcastFilterList),
	}
	if g.isZero() {
		return nil
	}
	return &g
}

// wlanScheduleFrom reconstructs the schedule group, returning nil when no member
// is set.
func wlanScheduleFrom(w *unifi.WLAN, prior *WlanSchedule) *WlanSchedule {
	var p WlanSchedule
	if prior != nil {
		p = *prior
	}
	g := WlanSchedule{
		Enabled: wlanBoolState(w.ScheduleEnabled, p.Enabled),
		Legacy:  wlanStringsState(w.Schedule, p.Legacy),
		Entries: wlanSchedulesState(w.ScheduleWithDuration, p.Entries),
	}
	if g.isZero() {
		return nil
	}
	return &g
}

// wlanApGroupsFrom reconstructs the apGroups group, returning nil when no member
// is set.
func wlanApGroupsFrom(w *unifi.WLAN, prior *WlanApGroups) *WlanApGroups {
	var p WlanApGroups
	if prior != nil {
		p = *prior
	}
	g := WlanApGroups{
		Ids:  wlanStringsState(w.ApGroupIDs, p.Ids),
		Mode: wlanStringState(w.ApGroupMode, p.Mode),
	}
	if g.isZero() {
		return nil
	}
	return &g
}

// wlanDpiFrom reconstructs the dpi group, returning nil when no member is set.
func wlanDpiFrom(w *unifi.WLAN, prior *WlanDpi) *WlanDpi {
	var p WlanDpi
	if prior != nil {
		p = *prior
	}
	g := WlanDpi{
		Enabled: wlanBoolState(w.DPIEnabled, p.Enabled),
		GroupId: wlanStringState(w.DPIgroupID, p.GroupId),
	}
	if g == (WlanDpi{}) {
		return nil
	}
	return &g
}

// wlanIotFrom reconstructs the iot group, returning nil when no member is set.
func wlanIotFrom(w *unifi.WLAN, prior *WlanIot) *WlanIot {
	var p WlanIot
	if prior != nil {
		p = *prior
	}
	g := WlanIot{
		Enhanced:                 wlanBoolState(w.EnhancedIot, p.Enhanced),
		OptimizeWifiConnectivity: wlanBoolState(w.OptimizeIotWifiConnectivity, p.OptimizeWifiConnectivity),
	}
	if g == (WlanIot{}) {
		return nil
	}
	return &g
}

// wlanP2pFrom reconstructs the p2p group, returning nil when no member is set.
func wlanP2pFrom(w *unifi.WLAN, prior *WlanP2p) *WlanP2p {
	var p WlanP2p
	if prior != nil {
		p = *prior
	}
	g := WlanP2p{
		Enabled:      wlanBoolState(w.P2P, p.Enabled),
		CrossConnect: wlanBoolState(w.P2PCrossConnect, p.CrossConnect),
	}
	if g == (WlanP2p{}) {
		return nil
	}
	return &g
}

// wlanRoamingFrom reconstructs the roaming group, returning nil when no member is
// set. The iappKey secret is preserved from prior when the controller masks it.
func wlanRoamingFrom(w *unifi.WLAN, prior *WlanRoaming) *WlanRoaming {
	var p WlanRoaming
	if prior != nil {
		p = *prior
	}
	g := WlanRoaming{
		FastRoamingEnabled: wlanBoolState(w.FastRoamingEnabled, p.FastRoamingEnabled),
		BssTransition:      wlanBoolState(w.BssTransition, p.BssTransition),
		IappKey:            wlanStringState(w.XIappKey, p.IappKey),
	}
	if g == (WlanRoaming{}) {
		return nil
	}
	return &g
}

func wlanStateFrom(w *unifi.WLAN, prior WlanArgs) WlanState {
	args := WlanArgs{
		Name:      w.Name,
		NetworkId: w.NetworkID,
		Enabled:   ptr(w.Enabled),
		Security:  ptr(w.Security),
		HideSsid:  ptr(w.HideSSID),
		// Preserve user-provided secrets; the controller may not echo them back.
		Passphrase: prior.Passphrase,
	}

	args.WlanGroupId = wlanStringState(w.WLANGroupID, prior.WlanGroupId)
	args.UserGroupId = wlanStringState(w.UserGroupID, prior.UserGroupId)
	args.UapsdEnabled = wlanBoolState(w.UapsdEnabled, prior.UapsdEnabled)
	args.MloEnabled = wlanBoolState(w.MloEnabled, prior.MloEnabled)
	args.L2Isolation = wlanBoolState(w.L2Isolation, prior.L2Isolation)
	args.IsGuest = wlanBoolState(w.IsGuest, prior.IsGuest)
	args.Priority = wlanStringState(w.Priority, prior.Priority)

	// Nested facets.
	args.Wpa = wlanWpaFrom(w, prior.Wpa)
	args.Wpa3 = wlanWpa3From(w, prior.Wpa3)
	args.Sae = wlanSaeFrom(w, prior.Sae)
	args.Wep = wlanWepFrom(w, prior.Wep)
	args.PrivatePresharedKeys = wlanPrivatePresharedKeysFrom(w, prior.PrivatePresharedKeys)
	args.Radius = wlanRadiusFrom(w, prior.Radius)
	args.VlanTagging = wlanVlanTaggingFrom(w, prior.VlanTagging)
	args.BandSteering = wlanBandSteeringFrom(w, prior.BandSteering)
	args.Dtim = wlanDtimFrom(w, prior.Dtim)
	args.MinRate = wlanMinRateFrom(w, prior.MinRate)
	args.MacFilter = wlanMacFilterFrom(w, prior.MacFilter)
	args.Multicast = wlanMulticastFrom(w, prior.Multicast)
	args.Schedule = wlanScheduleFrom(w, prior.Schedule)
	args.ApGroups = wlanApGroupsFrom(w, prior.ApGroups)
	args.Dpi = wlanDpiFrom(w, prior.Dpi)
	args.Iot = wlanIotFrom(w, prior.Iot)
	args.P2p = wlanP2pFrom(w, prior.P2p)
	args.Roaming = wlanRoamingFrom(w, prior.Roaming)

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
	return infer.CreateResponse[WlanState]{ID: created.ID, Output: wlanStateFrom(created, req.Inputs)}, nil
}

func (Wlan) Read(ctx context.Context, req infer.ReadRequest[WlanArgs, WlanState]) (infer.ReadResponse[WlanArgs, WlanState], error) {
	cfg := infer.GetConfig[config.Config](ctx)
	w, err := cfg.Network().GetWLAN(ctx, cfg.ResolvedSite(), req.ID)
	if notFound(err) {
		return infer.ReadResponse[WlanArgs, WlanState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[WlanArgs, WlanState]{}, err
	}
	st := wlanStateFrom(w, req.Inputs)
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
	return infer.UpdateResponse[WlanState]{Output: wlanStateFrom(updated, req.Inputs)}, nil
}

func (Wlan) Delete(ctx context.Context, req infer.DeleteRequest[WlanState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[config.Config](ctx)
	return infer.DeleteResponse{}, cfg.Network().DeleteWLAN(ctx, cfg.ResolvedSite(), req.ID)
}
