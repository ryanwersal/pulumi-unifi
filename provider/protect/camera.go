package protect

import (
	"context"
	"fmt"

	protecttypes "github.com/ClifHouck/unified/types"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/ryanwersal/pulumi-unifi/provider/config"
)

// Camera manages the settings of an EXISTING UniFi Protect camera.
//
// Cameras are physical hardware: the Protect API exposes no create/delete, only
// settings patches. This resource therefore follows an adoption model — it
// binds to a camera by its Protect ID and manages a subset of its settings.
// Create applies the desired settings to the already-adopted device, and Delete
// is a no-op (the camera is left in place, untouched).
type Camera struct{}

// CameraArgs are the user-supplied inputs.
type CameraArgs struct {
	// CameraId is the Protect camera ID (from the Protect API / bootstrap).
	CameraId string `pulumi:"cameraId"`
	// Name is the camera's display name.
	Name *string `pulumi:"name,optional"`
	// MicVolume sets the microphone volume (0-100).
	MicVolume *int `pulumi:"micVolume,optional"`
	// VideoMode, e.g. "default", "highFps".
	VideoMode *string `pulumi:"videoMode,optional"`
	// HdrType, e.g. "auto", "always", "off".
	HdrType *string `pulumi:"hdrType,optional"`
	// LedEnabled toggles the status LED.
	LedEnabled *bool `pulumi:"ledEnabled,optional"`
}

// CameraState is the persisted state: inputs plus read-only device facts.
type CameraState struct {
	CameraArgs
	// Type is the camera model/type (read-only).
	Type string `pulumi:"type"`
	// State is the connection state, e.g. "CONNECTED" (read-only).
	State string `pulumi:"state"`
}

func (c *Camera) Annotate(a infer.Annotator) {
	a.Describe(&c, "Manage settings of an existing UniFi Protect camera (adoption model; cameras are not created or deleted via the API).")
}

func (a CameraArgs) toPatch() *protecttypes.CameraPatchRequest {
	p := &protecttypes.CameraPatchRequest{}
	if a.Name != nil {
		p.Name = *a.Name
	}
	if a.MicVolume != nil {
		p.MicVolume = *a.MicVolume
	}
	if a.VideoMode != nil {
		p.VideoMode = *a.VideoMode
	}
	if a.HdrType != nil {
		p.HdrType = *a.HdrType
	}
	if a.LedEnabled != nil {
		p.LedSettings.IsEnabled = *a.LedEnabled
	}
	return p
}

// stateFrom maps a device into state. Write-mostly settings (mic/video/hdr/led)
// are preserved from prior inputs rather than re-read, since the Protect Camera
// payload nests them differently; Name and read-only facts come from the device.
func stateFrom(cam *protecttypes.Camera, prior CameraArgs) CameraState {
	args := prior
	args.CameraId = cam.ID
	args.Name = ptr(cam.Name)
	// These settings are top-level on the device, so round-trip them.
	if cam.VideoMode != "" {
		args.VideoMode = ptr(cam.VideoMode)
	}
	if cam.HdrType != "" {
		args.HdrType = ptr(cam.HdrType)
	}
	return CameraState{CameraArgs: args, Type: cam.ModelKey, State: cam.State}
}

func (Camera) Create(ctx context.Context, req infer.CreateRequest[CameraArgs]) (infer.CreateResponse[CameraState], error) {
	if req.DryRun {
		return infer.CreateResponse[CameraState]{Output: CameraState{CameraArgs: req.Inputs}}, nil
	}
	pc, err := cfgProtect(ctx)
	if err != nil {
		return infer.CreateResponse[CameraState]{}, err
	}
	id := protecttypes.CameraID(req.Inputs.CameraId)
	if _, err := pc.CameraDetails(id); err != nil {
		return infer.CreateResponse[CameraState]{}, fmt.Errorf("camera %q must already be adopted in Protect: %w", req.Inputs.CameraId, err)
	}
	cam, err := pc.CameraPatch(id, req.Inputs.toPatch())
	if err != nil {
		return infer.CreateResponse[CameraState]{}, err
	}
	return infer.CreateResponse[CameraState]{ID: req.Inputs.CameraId, Output: stateFrom(cam, req.Inputs)}, nil
}

func (Camera) Read(ctx context.Context, req infer.ReadRequest[CameraArgs, CameraState]) (infer.ReadResponse[CameraArgs, CameraState], error) {
	pc, err := cfgProtect(ctx)
	if err != nil {
		return infer.ReadResponse[CameraArgs, CameraState]{}, err
	}
	cam, err := pc.CameraDetails(protecttypes.CameraID(req.ID))
	if err != nil {
		return infer.ReadResponse[CameraArgs, CameraState]{}, err
	}
	st := stateFrom(cam, req.State.CameraArgs)
	return infer.ReadResponse[CameraArgs, CameraState]{ID: req.ID, Inputs: st.CameraArgs, State: st}, nil
}

func (Camera) Update(ctx context.Context, req infer.UpdateRequest[CameraArgs, CameraState]) (infer.UpdateResponse[CameraState], error) {
	if req.DryRun {
		return infer.UpdateResponse[CameraState]{Output: CameraState{CameraArgs: req.Inputs}}, nil
	}
	pc, err := cfgProtect(ctx)
	if err != nil {
		return infer.UpdateResponse[CameraState]{}, err
	}
	cam, err := pc.CameraPatch(protecttypes.CameraID(req.ID), req.Inputs.toPatch())
	if err != nil {
		return infer.UpdateResponse[CameraState]{}, err
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
