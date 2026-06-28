// SPDX-License-Identifier: Apache-2.0

package protect

import (
	"testing"

	protecttypes "github.com/ClifHouck/unified/types"
)

func TestCameraToPatchRoundTrip(t *testing.T) {
	args := CameraArgs{
		CameraId:               "cam-123",
		Name:                   ptr("Front Door"),
		MicVolume:              ptr(80),
		VideoMode:              ptr("highFps"),
		HdrType:                ptr("auto"),
		LedEnabled:             ptr(true),
		OsdNameEnabled:         ptr(true),
		OsdDateEnabled:         ptr(false),
		OsdLogoEnabled:         ptr(true),
		OsdDebugEnabled:        ptr(false),
		LcdMessageType:         ptr("CUSTOM_MESSAGE"),
		LcdMessageText:         ptr("Be right there"),
		LcdMessageResetAt:      ptr(1700000000000),
		SmartDetectObjectTypes: []string{"person", "vehicle"},
		SmartDetectAudioTypes:  []string{"smoke"},
	}

	p := args.toPatch()

	if p.Name != "Front Door" {
		t.Errorf("Name = %q, want Front Door", p.Name)
	}
	if p.MicVolume != 80 {
		t.Errorf("MicVolume = %d, want 80", p.MicVolume)
	}
	if p.VideoMode != "highFps" {
		t.Errorf("VideoMode = %q, want highFps", p.VideoMode)
	}
	if p.HdrType != "auto" {
		t.Errorf("HdrType = %q, want auto", p.HdrType)
	}
	if !p.LedSettings.IsEnabled {
		t.Error("LedSettings.IsEnabled = false, want true")
	}
	if !p.OsdSettings.IsNameEnabled || !p.OsdSettings.IsLogoEnabled {
		t.Error("OsdSettings name/logo should be enabled")
	}
	if p.LcdMessage.Type != "CUSTOM_MESSAGE" || p.LcdMessage.Text != "Be right there" || p.LcdMessage.ResetAt != 1700000000000 {
		t.Errorf("LcdMessage = %+v, want CUSTOM_MESSAGE/Be right there/1700000000000", p.LcdMessage)
	}
	if len(p.SmartDetectSettings.ObjectTypes) != 2 || p.SmartDetectSettings.ObjectTypes[0] != "person" {
		t.Errorf("SmartDetectSettings.ObjectTypes = %v, want [person vehicle]", p.SmartDetectSettings.ObjectTypes)
	}
	if len(p.SmartDetectSettings.AudioTypes) != 1 || p.SmartDetectSettings.AudioTypes[0] != "smoke" {
		t.Errorf("SmartDetectSettings.AudioTypes = %v, want [smoke]", p.SmartDetectSettings.AudioTypes)
	}

	// Build a device echoing the patch and confirm stateFrom maps it back.
	cam := &protecttypes.Camera{}
	cam.ID = "cam-123"
	cam.ModelKey = "UVC G4 Doorbell"
	cam.State = "CONNECTED"
	cam.Name = "Front Door"
	cam.MicVolume = 80
	cam.VideoMode = "highFps"
	cam.HdrType = "auto"
	cam.LedSettings.IsEnabled = true
	cam.OsdSettings.IsNameEnabled = true
	cam.OsdSettings.IsLogoEnabled = true
	cam.LcdMessage.Type = "CUSTOM_MESSAGE"
	cam.LcdMessage.Text = "Be right there"
	cam.LcdMessage.ResetAt = 1700000000000
	cam.SmartDetectSettings.ObjectTypes = []string{"person", "vehicle"}
	cam.SmartDetectSettings.AudioTypes = []string{"smoke"}

	st := stateFrom(cam, args)

	if st.Type != "UVC G4 Doorbell" {
		t.Errorf("Type = %q, want UVC G4 Doorbell", st.Type)
	}
	if st.State != "CONNECTED" {
		t.Errorf("State = %q, want CONNECTED", st.State)
	}
	if st.CameraId != "cam-123" {
		t.Errorf("CameraId = %q, want cam-123", st.CameraId)
	}
	if derefOr(st.Name, "") != "Front Door" {
		t.Errorf("Name = %v, want Front Door", st.Name)
	}
	if derefOr(st.MicVolume, 0) != 80 {
		t.Errorf("MicVolume = %v, want 80", st.MicVolume)
	}
	if derefOr(st.VideoMode, "") != "highFps" {
		t.Errorf("VideoMode = %v, want highFps", st.VideoMode)
	}
	if derefOr(st.LedEnabled, false) != true {
		t.Errorf("LedEnabled = %v, want true", st.LedEnabled)
	}
	if derefOr(st.OsdNameEnabled, false) != true {
		t.Errorf("OsdNameEnabled = %v, want true", st.OsdNameEnabled)
	}
	if derefOr(st.LcdMessageText, "") != "Be right there" {
		t.Errorf("LcdMessageText = %v, want Be right there", st.LcdMessageText)
	}
	if len(st.SmartDetectObjectTypes) != 2 {
		t.Errorf("SmartDetectObjectTypes = %v, want 2 entries", st.SmartDetectObjectTypes)
	}
}

func TestCameraStateFromPreservesPriorForZeroValues(t *testing.T) {
	// User set ledEnabled=true and a mic volume, but the device returns zero
	// values (e.g. mid-reconnect). Prior inputs should be preserved.
	prior := CameraArgs{
		CameraId:   "cam-9",
		LedEnabled: ptr(true),
		MicVolume:  ptr(50),
		HdrType:    ptr("off"),
	}
	cam := &protecttypes.Camera{}
	cam.ID = "cam-9"
	cam.ModelKey = "UVC G4 Bullet"
	cam.State = "DISCONNECTED"
	cam.Name = "Yard"

	st := stateFrom(cam, prior)

	if derefOr(st.MicVolume, 0) != 50 {
		t.Errorf("MicVolume = %v, want preserved 50", st.MicVolume)
	}
	if derefOr(st.HdrType, "") != "off" {
		t.Errorf("HdrType = %v, want preserved off", st.HdrType)
	}
	if derefOr(st.LedEnabled, false) != true {
		t.Errorf("LedEnabled = %v, want preserved true", st.LedEnabled)
	}
	// An optional bool the user never set must stay nil despite the device's false.
	if st.OsdDebugEnabled != nil {
		t.Errorf("OsdDebugEnabled = %v, want nil", st.OsdDebugEnabled)
	}
}
