// SPDX-License-Identifier: Apache-2.0

package protect

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/filipowm/go-unifi/unifi"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/ryanwersal/pulumi-unifi/provider/config"
)

// CameraStream owns one RTSPS quality on an already-adopted camera.
// Other qualities and camera settings are never changed. URLs are deliberately
// not persisted: a stream URL grants access to camera video.
type CameraStream struct{}

type CameraStreamArgs struct {
	CameraId string `pulumi:"cameraId" provider:"replaceOnChanges"`
	Quality  string `pulumi:"quality" provider:"replaceOnChanges"`
}

type CameraStreamState struct{ CameraStreamArgs }

func (r *CameraStream) Annotate(a infer.Annotator) {
	a.Describe(&r, "Enable one RTSPS stream quality on an existing Protect camera through the public integration API. Imports use cameraId/quality. Delete disables only this quality; use retainOnDelete to leave it enabled. Stream URLs are not stored or exported. Refresh detects streams disabled outside Pulumi.")
}

func (a *CameraStreamArgs) Annotate(n infer.Annotator) {
	n.Describe(&a.CameraId, "ID of an already-adopted Protect camera.")
	n.Describe(&a.Quality, "RTSPS quality to enable: high, medium, low, or package (where supported by the camera). Changing camera or quality replaces the resource.")
}

var cameraStreamIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func validStreamQuality(q string) bool {
	switch q {
	case "high", "medium", "low", "package":
		return true
	}
	return false
}

func (CameraStream) Check(ctx context.Context, req infer.CheckRequest) (infer.CheckResponse[CameraStreamArgs], error) {
	args, failures, err := infer.DefaultCheck[CameraStreamArgs](ctx, req.NewInputs)
	if err != nil || len(failures) > 0 {
		return infer.CheckResponse[CameraStreamArgs]{Inputs: args, Failures: failures}, err
	}
	if !req.NewInputs.Get("cameraId").IsComputed() && !cameraStreamIDPattern.MatchString(args.CameraId) {
		failures = append(failures, p.CheckFailure{Property: "cameraId", Reason: "must be a nonempty camera ID containing only letters, digits, underscores, or hyphens"})
	}
	if !req.NewInputs.Get("quality").IsComputed() && !validStreamQuality(args.Quality) {
		failures = append(failures, p.CheckFailure{Property: "quality", Reason: "must be high, medium, low, or package"})
	}
	return infer.CheckResponse[CameraStreamArgs]{Inputs: args, Failures: failures}, nil
}

func parseCameraStreamID(id string) (CameraStreamArgs, error) {
	camera, quality, ok := strings.Cut(id, "/")
	if !ok || !cameraStreamIDPattern.MatchString(camera) || !validStreamQuality(quality) {
		return CameraStreamArgs{}, fmt.Errorf("stream ID must be cameraId/quality with quality high, medium, low, or package")
	}
	return CameraStreamArgs{CameraId: camera, Quality: quality}, nil
}

func streamClient(ctx context.Context) (unifi.Client, error) {
	cfg := infer.GetConfig[config.Config](ctx)
	// The public Protect API requires an API key, not a Network login session.
	if _, err := cfg.Protect(); err != nil {
		return nil, err
	}
	return cfg.Controller(), nil
}

func streamPath(a CameraStreamArgs) string { return cameraPatchPath + a.CameraId + "/rtsps-stream" }

func streamExists(ctx context.Context, client unifi.Client, a CameraStreamArgs) (bool, error) {
	var streams map[string]*string
	if err := client.Do(ctx, http.MethodGet, streamPath(a), nil, &streams); err != nil {
		return false, err
	}
	value := streams[a.Quality]
	return value != nil && *value != "", nil
}

func streamNotFound(err error) bool {
	var serverError *unifi.ServerError
	return errors.Is(err, unifi.ErrNotFound) || (errors.As(err, &serverError) && serverError.StatusCode == http.StatusNotFound)
}

func (CameraStream) Create(ctx context.Context, req infer.CreateRequest[CameraStreamArgs]) (infer.CreateResponse[CameraStreamState], error) {
	state := CameraStreamState{CameraStreamArgs: req.Inputs}
	if req.DryRun {
		return infer.CreateResponse[CameraStreamState]{Output: state}, nil
	}
	id := req.Inputs.CameraId + "/" + req.Inputs.Quality
	if _, err := parseCameraStreamID(id); err != nil {
		return infer.CreateResponse[CameraStreamState]{}, err
	}
	client, err := streamClient(ctx)
	if err != nil {
		return infer.CreateResponse[CameraStreamState]{}, err
	}
	active, err := streamExists(ctx, client, req.Inputs)
	if err != nil {
		return infer.CreateResponse[CameraStreamState]{}, wrap("read camera stream before enabling", err)
	}
	if !active {
		// Do not regenerate an existing URL. POST adds only the requested quality.
		body := map[string]any{"qualities": []string{req.Inputs.Quality}}
		if err := client.Do(ctx, http.MethodPost, streamPath(req.Inputs), body, nil); err != nil {
			return infer.CreateResponse[CameraStreamState]{}, wrap("enable camera stream", err)
		}
		active, err = streamExists(ctx, client, req.Inputs)
		if err != nil {
			return infer.CreateResponse[CameraStreamState]{}, wrap("verify camera stream", err)
		}
		if !active {
			return infer.CreateResponse[CameraStreamState]{}, fmt.Errorf("protect did not enable the requested %s stream on camera %s", req.Inputs.Quality, req.Inputs.CameraId)
		}
	}
	return infer.CreateResponse[CameraStreamState]{ID: id, Output: state}, nil
}

func (CameraStream) Read(ctx context.Context, req infer.ReadRequest[CameraStreamArgs, CameraStreamState]) (infer.ReadResponse[CameraStreamArgs, CameraStreamState], error) {
	args, err := parseCameraStreamID(req.ID)
	if err != nil {
		return infer.ReadResponse[CameraStreamArgs, CameraStreamState]{}, err
	}
	client, err := streamClient(ctx)
	if err != nil {
		return infer.ReadResponse[CameraStreamArgs, CameraStreamState]{}, err
	}
	active, err := streamExists(ctx, client, args)
	if streamNotFound(err) || (err == nil && !active) {
		return infer.ReadResponse[CameraStreamArgs, CameraStreamState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[CameraStreamArgs, CameraStreamState]{}, wrap("read camera stream", err)
	}
	return infer.ReadResponse[CameraStreamArgs, CameraStreamState]{ID: req.ID, Inputs: args, State: CameraStreamState{CameraStreamArgs: args}}, nil
}

func (CameraStream) Delete(ctx context.Context, req infer.DeleteRequest[CameraStreamState]) (infer.DeleteResponse, error) {
	args, err := parseCameraStreamID(req.ID)
	if err != nil {
		return infer.DeleteResponse{}, err
	}
	client, err := streamClient(ctx)
	if err != nil {
		return infer.DeleteResponse{}, err
	}
	active, err := streamExists(ctx, client, args)
	if streamNotFound(err) || (err == nil && !active) {
		return infer.DeleteResponse{}, nil
	}
	if err != nil {
		return infer.DeleteResponse{}, wrap("read camera stream before disabling", err)
	}
	query := url.Values{"qualities": []string{args.Quality}}
	err = client.Do(ctx, http.MethodDelete, streamPath(args)+"?"+query.Encode(), nil, nil)
	if streamNotFound(err) {
		return infer.DeleteResponse{}, nil
	}
	return infer.DeleteResponse{}, wrap("disable camera stream", err)
}
