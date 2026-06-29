// SPDX-License-Identifier: Apache-2.0

package protect

import (
	"context"
	"fmt"
	"net/http"

	protecttypes "github.com/ClifHouck/unified/types"
	"github.com/filipowm/go-unifi/unifi"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/ryanwersal/pulumi-unifi/provider/config"
)

// Camera manages the settings of an EXISTING UniFi Protect camera.
//
// Cameras are physical hardware: the Protect API exposes no create/delete, only
// settings patches. This resource therefore follows an adoption model — it
// binds to a camera by its Protect ID and manages the subset of its settings
// that the Protect CameraPatchRequest supports. Create applies the desired
// settings to the already-adopted device, and Delete is a no-op (the camera is
// left in place, untouched).
//
// Field selection: the Terraform UniFi providers (filipowm, paultyng) model the
// network controller only and have NO Protect camera resource, so there is no
// upstream naming to mirror; the exposed surface is exactly the writable fields
// of the Protect CameraPatchRequest (name, OSD overlays, status LED, doorbell
// LCD message, microphone volume, video mode, HDR, and smart-detect types).
type Camera struct{}

// CameraArgs are the user-supplied inputs.
type CameraArgs struct {
	// CameraId is the Protect camera ID (from the Protect API / bootstrap).
	CameraId string `pulumi:"cameraId" provider:"replaceOnChanges"`
	// Name is the camera's display name.
	Name *string `pulumi:"name,optional"`
	// MicVolume sets the microphone volume (0-100). A value of 0 is not applied
	// by the Protect API (the field is omitted when zero).
	MicVolume *int `pulumi:"micVolume,optional"`
	// VideoMode selects the capture mode, e.g. "default", "highFps", "sport".
	VideoMode *string `pulumi:"videoMode,optional"`
	// HdrType selects HDR behaviour, e.g. "auto", "always", "off".
	HdrType *string `pulumi:"hdrType,optional"`
	// LedEnabled toggles the status LED (ledSettings.isEnabled).
	LedEnabled *bool `pulumi:"ledEnabled,optional"`

	// OsdNameEnabled overlays the camera name on the video (osdSettings.isNameEnabled).
	OsdNameEnabled *bool `pulumi:"osdNameEnabled,optional"`
	// OsdDateEnabled overlays the date/time on the video (osdSettings.isDateEnabled).
	OsdDateEnabled *bool `pulumi:"osdDateEnabled,optional"`
	// OsdLogoEnabled overlays the logo on the video (osdSettings.isLogoEnabled).
	OsdLogoEnabled *bool `pulumi:"osdLogoEnabled,optional"`
	// OsdDebugEnabled overlays debug telemetry on the video (osdSettings.isDebugEnabled).
	OsdDebugEnabled *bool `pulumi:"osdDebugEnabled,optional"`

	// LcdMessageType is the doorbell LCD message type, e.g. "CUSTOM_MESSAGE",
	// "DO_NOT_DISTURB", "LEAVE_PACKAGE_AT_DOOR" (doorbell cameras only).
	LcdMessageType *string `pulumi:"lcdMessageType,optional"`
	// LcdMessageText is the doorbell LCD custom message text (used with type CUSTOM_MESSAGE).
	LcdMessageText *string `pulumi:"lcdMessageText,optional"`
	// LcdMessageResetAt is the epoch-millisecond timestamp at which the LCD
	// message clears. 0 (or omitted) leaves the message until changed.
	LcdMessageResetAt *int `pulumi:"lcdMessageResetAt,optional"`

	// SmartDetectObjectTypes selects which objects trigger smart detections,
	// e.g. "person", "vehicle", "animal", "package", "licensePlate".
	SmartDetectObjectTypes []string `pulumi:"smartDetectObjectTypes,optional"`
	// SmartDetectAudioTypes selects which sounds trigger smart detections,
	// e.g. "smoke", "cmonitor" (CO alarm), "alrmSmoke", "alrmCmonitor".
	SmartDetectAudioTypes []string `pulumi:"smartDetectAudioTypes,optional"`
}

// CameraState is the persisted state: inputs plus read-only device facts.
type CameraState struct {
	CameraArgs
	// Type is the camera model/type (read-only; the Protect modelKey).
	Type string `pulumi:"type"`
	// State is the connection state, e.g. "CONNECTED" (read-only).
	State string `pulumi:"state"`
}

func (c *Camera) Annotate(a infer.Annotator) {
	a.Describe(&c, "Manage settings of an existing UniFi Protect camera (adoption model; cameras are not created or deleted via the API). "+
		"Exposes the writable fields of the Protect CameraPatchRequest.")
}

func (d *CameraArgs) Annotate(a infer.Annotator) {
	a.Describe(&d.CameraId, "CameraId is the Protect camera ID (from the Protect API / bootstrap).")
	a.Describe(&d.Name, "Name is the camera's display name.")
	a.Describe(&d.MicVolume, "MicVolume sets the microphone volume (0-100). A value of 0 is not applied by the Protect API (the field is omitted when zero).")
	a.Describe(&d.VideoMode, "VideoMode selects the capture mode, e.g. \"default\", \"highFps\", \"sport\".")
	a.Describe(&d.HdrType, "HdrType selects HDR behaviour, e.g. \"auto\", \"always\", \"off\".")
	a.Describe(&d.LedEnabled, "LedEnabled toggles the status LED (ledSettings.isEnabled).")
	a.Describe(&d.OsdNameEnabled, "OsdNameEnabled overlays the camera name on the video (osdSettings.isNameEnabled).")
	a.Describe(&d.OsdDateEnabled, "OsdDateEnabled overlays the date/time on the video (osdSettings.isDateEnabled).")
	a.Describe(&d.OsdLogoEnabled, "OsdLogoEnabled overlays the logo on the video (osdSettings.isLogoEnabled).")
	a.Describe(&d.OsdDebugEnabled, "OsdDebugEnabled overlays debug telemetry on the video (osdSettings.isDebugEnabled).")
	a.Describe(&d.LcdMessageType, "LcdMessageType is the doorbell LCD message type, e.g. \"CUSTOM_MESSAGE\", \"DO_NOT_DISTURB\", \"LEAVE_PACKAGE_AT_DOOR\" (doorbell cameras only).")
	a.Describe(&d.LcdMessageText, "LcdMessageText is the doorbell LCD custom message text (used with type CUSTOM_MESSAGE).")
	a.Describe(&d.LcdMessageResetAt, "LcdMessageResetAt is the epoch-millisecond timestamp at which the LCD message clears. 0 (or omitted) leaves the message until changed.")
	a.Describe(&d.SmartDetectObjectTypes, "SmartDetectObjectTypes selects which objects trigger smart detections, e.g. \"person\", \"vehicle\", \"animal\", \"package\", \"licensePlate\".")
	a.Describe(&d.SmartDetectAudioTypes, "SmartDetectAudioTypes selects which sounds trigger smart detections, e.g. \"smoke\", \"cmonitor\" (CO alarm), \"alrmSmoke\", \"alrmCmonitor\".")
}

func (s *CameraState) Annotate(a infer.Annotator) {
	a.Describe(&s.Type, "Type is the camera model/type (read-only; the Protect modelKey).")
	a.Describe(&s.State, "State is the connection state, e.g. \"CONNECTED\" (read-only).")
}

// cameraPatchPath is the Protect integration API camera resource.
const cameraPatchPath = "/proxy/protect/integration/v1/cameras/"

// toPatchBody builds the PATCH body as a map so explicit `false` toggles and
// empty smart-detect lists survive marshaling. The typed CameraPatchRequest
// can't: its bool/slice fields are `omitempty` (and the nested settings structs
// `omitzero`), so a false/empty value is dropped and the toggle can never be
// turned off. Strings/ints keep the "omit zero" behavior (the API ignores a 0
// micVolume / empty string).
func (a CameraArgs) toPatchBody() map[string]any {
	body := map[string]any{}
	if a.Name != nil && *a.Name != "" {
		body["name"] = *a.Name
	}
	if a.MicVolume != nil && *a.MicVolume != 0 {
		body["micVolume"] = *a.MicVolume
	}
	if a.VideoMode != nil && *a.VideoMode != "" {
		body["videoMode"] = *a.VideoMode
	}
	if a.HdrType != nil && *a.HdrType != "" {
		body["hdrType"] = *a.HdrType
	}
	if a.LedEnabled != nil {
		body["ledSettings"] = map[string]any{"isEnabled": *a.LedEnabled}
	}
	osd := map[string]any{}
	if a.OsdNameEnabled != nil {
		osd["isNameEnabled"] = *a.OsdNameEnabled
	}
	if a.OsdDateEnabled != nil {
		osd["isDateEnabled"] = *a.OsdDateEnabled
	}
	if a.OsdLogoEnabled != nil {
		osd["isLogoEnabled"] = *a.OsdLogoEnabled
	}
	if a.OsdDebugEnabled != nil {
		osd["isDebugEnabled"] = *a.OsdDebugEnabled
	}
	if len(osd) > 0 {
		body["osdSettings"] = osd
	}
	lcd := map[string]any{}
	if a.LcdMessageType != nil && *a.LcdMessageType != "" {
		lcd["type"] = *a.LcdMessageType
	}
	if a.LcdMessageText != nil && *a.LcdMessageText != "" {
		lcd["text"] = *a.LcdMessageText
	}
	if a.LcdMessageResetAt != nil && *a.LcdMessageResetAt != 0 {
		lcd["resetAt"] = *a.LcdMessageResetAt
	}
	if len(lcd) > 0 {
		body["lcdMessage"] = lcd
	}
	smart := map[string]any{}
	if a.SmartDetectObjectTypes != nil {
		smart["objectTypes"] = a.SmartDetectObjectTypes
	}
	if a.SmartDetectAudioTypes != nil {
		smart["audioTypes"] = a.SmartDetectAudioTypes
	}
	if len(smart) > 0 {
		body["smartDetectSettings"] = smart
	}
	return body
}

// cameraPatch PATCHes the camera via the go-unifi client (which carries the
// X-API-Key) with a raw body, then decodes the updated camera.
func cameraPatch(ctx context.Context, nc unifi.Client, id string, body map[string]any) (*protecttypes.Camera, error) {
	var cam protecttypes.Camera
	if err := nc.Do(ctx, http.MethodPatch, cameraPatchPath+id, body, &cam); err != nil {
		return nil, err
	}
	return &cam, nil
}

// cameraStrPtr reflects a device string, falling back to the prior input when empty.
func cameraStrPtr(v string, prior *string) *string {
	if v != "" {
		return ptr(v)
	}
	return prior
}

// cameraIntPtr reflects a device int, falling back to the prior input when zero.
func cameraIntPtr(v int, prior *int) *int {
	if v != 0 {
		return ptr(v)
	}
	return prior
}

// cameraBoolPtr reflects a device bool when the user set it or when it is true,
// otherwise leaves the optional input unset to avoid spurious diffs.
func cameraBoolPtr(v bool, prior *bool) *bool {
	if v {
		return ptr(v)
	}
	return prior
}

// cameraStrSlice reflects a device string slice, falling back to the prior
// input when the device returns nothing.
func cameraStrSlice(v []string, prior []string) []string {
	if len(v) > 0 {
		return v
	}
	return prior
}

// stateFrom maps a device into state. Settings the Camera payload echoes are
// round-tripped from the device; the prior inputs preserve any optional fields
// the device leaves at their zero value so they do not produce spurious diffs.
func stateFrom(cam *protecttypes.Camera, prior CameraArgs) CameraState {
	args := prior
	args.CameraId = cam.ID
	args.Name = ptr(cam.Name)
	args.MicVolume = cameraIntPtr(cam.MicVolume, prior.MicVolume)
	args.VideoMode = cameraStrPtr(cam.VideoMode, prior.VideoMode)
	args.HdrType = cameraStrPtr(cam.HdrType, prior.HdrType)
	args.LedEnabled = cameraBoolPtr(cam.LedSettings.IsEnabled, prior.LedEnabled)
	args.OsdNameEnabled = cameraBoolPtr(cam.OsdSettings.IsNameEnabled, prior.OsdNameEnabled)
	args.OsdDateEnabled = cameraBoolPtr(cam.OsdSettings.IsDateEnabled, prior.OsdDateEnabled)
	args.OsdLogoEnabled = cameraBoolPtr(cam.OsdSettings.IsLogoEnabled, prior.OsdLogoEnabled)
	args.OsdDebugEnabled = cameraBoolPtr(cam.OsdSettings.IsDebugEnabled, prior.OsdDebugEnabled)
	args.LcdMessageType = cameraStrPtr(cam.LcdMessage.Type, prior.LcdMessageType)
	args.LcdMessageText = cameraStrPtr(cam.LcdMessage.Text, prior.LcdMessageText)
	args.LcdMessageResetAt = cameraIntPtr(cam.LcdMessage.ResetAt, prior.LcdMessageResetAt)
	args.SmartDetectObjectTypes = cameraStrSlice(cam.SmartDetectSettings.ObjectTypes, prior.SmartDetectObjectTypes)
	args.SmartDetectAudioTypes = cameraStrSlice(cam.SmartDetectSettings.AudioTypes, prior.SmartDetectAudioTypes)
	return CameraState{CameraArgs: args, Type: cam.ModelKey, State: cam.State}
}

func (Camera) Create(ctx context.Context, req infer.CreateRequest[CameraArgs]) (infer.CreateResponse[CameraState], error) {
	if req.DryRun {
		return infer.CreateResponse[CameraState]{Output: CameraState{CameraArgs: req.Inputs}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	pc, err := cfg.Protect()
	if err != nil {
		return infer.CreateResponse[CameraState]{}, err
	}
	if _, err := pc.CameraDetails(protecttypes.CameraID(req.Inputs.CameraId)); err != nil {
		return infer.CreateResponse[CameraState]{}, fmt.Errorf("camera %q must already be adopted in Protect: %w", req.Inputs.CameraId, err)
	}
	cam, err := cameraPatch(ctx, cfg.Controller(), req.Inputs.CameraId, req.Inputs.toPatchBody())
	if err != nil {
		return infer.CreateResponse[CameraState]{}, wrap(fmt.Sprintf("create camera %q", req.Inputs.CameraId), err)
	}
	return infer.CreateResponse[CameraState]{ID: req.Inputs.CameraId, Output: stateFrom(cam, req.Inputs)}, nil
}

func (Camera) Read(ctx context.Context, req infer.ReadRequest[CameraArgs, CameraState]) (infer.ReadResponse[CameraArgs, CameraState], error) {
	pc, err := cfgProtect(ctx)
	if err != nil {
		return infer.ReadResponse[CameraArgs, CameraState]{}, err
	}
	cam, err := pc.CameraDetails(protecttypes.CameraID(req.ID))
	if isProtectNotFound(err) {
		return infer.ReadResponse[CameraArgs, CameraState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[CameraArgs, CameraState]{}, wrap(fmt.Sprintf("read camera %q", req.ID), err)
	}
	st := stateFrom(cam, req.State.CameraArgs)
	return infer.ReadResponse[CameraArgs, CameraState]{ID: req.ID, Inputs: st.CameraArgs, State: st}, nil
}

func (Camera) Update(ctx context.Context, req infer.UpdateRequest[CameraArgs, CameraState]) (infer.UpdateResponse[CameraState], error) {
	if req.DryRun {
		return infer.UpdateResponse[CameraState]{Output: CameraState{CameraArgs: req.Inputs}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	// Protect requires an API key; gate on it for a clear error before patching.
	if _, err := cfg.Protect(); err != nil {
		return infer.UpdateResponse[CameraState]{}, err
	}
	cam, err := cameraPatch(ctx, cfg.Controller(), req.ID, req.Inputs.toPatchBody())
	if err != nil {
		return infer.UpdateResponse[CameraState]{}, wrap(fmt.Sprintf("update camera %q", req.ID), err)
	}
	return infer.UpdateResponse[CameraState]{Output: stateFrom(cam, req.Inputs)}, nil
}

// Delete is a no-op: the physical camera is not removed, only released from
// Pulumi's management. Its current settings are left in place.
func (Camera) Delete(_ context.Context, _ infer.DeleteRequest[CameraState]) (infer.DeleteResponse, error) {
	return infer.DeleteResponse{}, nil
}

func cfgProtect(ctx context.Context) (protecttypes.ProtectV1, error) {
	return infer.GetConfig[config.Config](ctx).Protect()
}
