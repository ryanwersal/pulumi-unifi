// SPDX-License-Identifier: Apache-2.0

package protect

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
)

func TestCameraStreamCheck(t *testing.T) {
	for _, tc := range []struct {
		camera, quality string
		valid           bool
	}{
		{"cam-123", "high", true}, {"cam-123", "medium", true}, {"cam-123", "low", true}, {"cam-123", "package", true},
		{"", "high", false}, {"../cam", "high", false}, {"cam?x=y", "high", false}, {"cam-123", "", false}, {"cam-123", "ultra", false},
	} {
		in := property.NewMap(map[string]property.Value{"cameraId": property.New(tc.camera), "quality": property.New(tc.quality)})
		got, err := (CameraStream{}).Check(context.Background(), infer.CheckRequest{NewInputs: in})
		if err != nil {
			t.Fatal(err)
		}
		if (len(got.Failures) == 0) != tc.valid {
			t.Errorf("Check(%q,%q) failures=%v", tc.camera, tc.quality, got.Failures)
		}
	}
	in := property.NewMap(map[string]property.Value{"cameraId": property.New(property.Computed), "quality": property.New(property.Computed)})
	got, err := (CameraStream{}).Check(context.Background(), infer.CheckRequest{NewInputs: in})
	if err != nil || len(got.Failures) > 0 {
		t.Fatalf("unknown preview inputs rejected: %+v %v", got, err)
	}
}

func TestCameraStreamHTTPRead(t *testing.T) {
	status := http.StatusOK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path != "/proxy/protect/integration/v1/cameras/cam-123/rtsps-stream" || r.Header.Get("X-API-Key") != "test-key" {
			t.Error("incorrect request path or missing API key")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if status == http.StatusOK {
			_, _ = fmt.Fprint(w, `{"high":"rtsps://private-token","medium":null,"low":null,"package":null}`)
		} else {
			_, _ = fmt.Fprint(w, `{"message":"request failed"}`)
		}
	}))
	defer server.Close()
	client, err := unifi.NewBareClient(&unifi.ClientConfig{URL: server.URL, APIKey: "test-key", Logger: unifi.NewDefaultLogger(unifi.DisabledLevel)})
	if err != nil {
		t.Fatal(err)
	}
	active, err := streamExists(context.Background(), client, CameraStreamArgs{CameraId: "cam-123", Quality: "high"})
	if err != nil || !active {
		t.Fatalf("read high stream: %v %v", active, err)
	}
	active, err = streamExists(context.Background(), client, CameraStreamArgs{CameraId: "cam-123", Quality: "medium"})
	if err != nil || active {
		t.Fatalf("read disabled stream: %v %v", active, err)
	}
	for _, code := range []int{404, 403, 500} {
		status = code
		_, err := streamExists(context.Background(), client, CameraStreamArgs{CameraId: "cam-123", Quality: "high"})
		if err == nil || streamNotFound(err) != (code == 404) {
			t.Fatalf("incorrect error classification for %d: %v", code, err)
		}
	}
}
