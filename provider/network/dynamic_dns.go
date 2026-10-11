// SPDX-License-Identifier: Apache-2.0

package network

import (
	"context"
	"fmt"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/ryanwersal/pulumi-unifi/provider/config"
)

// DynamicDns configures the gateway's persistent dynamic DNS updater.
type DynamicDns struct{}

type DynamicDnsArgs struct {
	HostName      string   `pulumi:"hostName"`
	Service       string   `pulumi:"service"`
	Login         string   `pulumi:"login"`
	Password      string   `pulumi:"password" provider:"secret"`
	Interface     *string  `pulumi:"interface,optional"`
	Server        *string  `pulumi:"server,optional"`
	CustomService *string  `pulumi:"customService,optional"`
	Options       []string `pulumi:"options,optional"`
}

func (d *DynamicDnsArgs) Annotate(a infer.Annotator) {
	a.Describe(&d.HostName, "Fully qualified DNS hostname to update.")
	a.Describe(&d.Service, "UniFi DDNS provider identifier, such as cloudflare, dyndns, or custom.")
	a.Describe(&d.Login, "Provider login. For Cloudflare, use the zone name (for example example.com), not an email address.")
	a.Describe(&d.Password, "Provider password or API token. Stored as a Pulumi secret; retained when the controller omits it on read.")
	a.Describe(&d.Interface, "Logical WAN interface: wan for Internet 1, wan2 for Internet 2, and so on.")
	a.SetDefault(&d.Interface, "wan")
	a.Describe(&d.Server, "Optional DDNS server override. Empty uses the provider default.")
	a.Describe(&d.CustomService, "Custom provider identifier when service is custom.")
	a.Describe(&d.Options, "Additional controller DDNS options.")
}

type DynamicDnsState struct {
	DynamicDnsArgs
	DynamicDnsId string `pulumi:"dynamicDnsId"`
}

func (d *DynamicDnsState) Annotate(a infer.Annotator) {
	a.Describe(&d.DynamicDnsId, "Controller-assigned dynamic DNS entry ID, also used for import.")
}

func (d *DynamicDns) Annotate(a infer.Annotator) {
	a.Describe(&d, "A gateway dynamic DNS entry. The gateway updates the provider independently of Pulumi runs. Manage the public DNS record separately.")
}

func (a DynamicDnsArgs) toUnifi(id string) *unifi.DynamicDNS {
	return &unifi.DynamicDNS{
		ID: id, HostName: a.HostName, Service: a.Service, Login: a.Login,
		XPassword: a.Password, Interface: derefOr(a.Interface, "wan"),
		Server: derefOr(a.Server, ""), CustomService: derefOr(a.CustomService, ""),
		Options: a.Options,
	}
}

func dynamicDnsStateFrom(u *unifi.DynamicDNS, prior DynamicDnsArgs) DynamicDnsState {
	password := u.XPassword
	if password == "" || password == "********" {
		password = prior.Password
	}
	args := DynamicDnsArgs{
		HostName: u.HostName, Service: u.Service, Login: u.Login, Password: password,
		Interface: ptr(u.Interface), Options: u.Options,
	}
	if u.Server != "" || prior.Server != nil {
		args.Server = ptr(u.Server)
	}
	if u.CustomService != "" || prior.CustomService != nil {
		args.CustomService = ptr(u.CustomService)
	}
	return DynamicDnsState{DynamicDnsArgs: args, DynamicDnsId: u.ID}
}

// Create provisions a new dynamic DNS entry.
func (DynamicDns) Create(ctx context.Context, req infer.CreateRequest[DynamicDnsArgs]) (infer.CreateResponse[DynamicDnsState], error) {
	if req.DryRun {
		return infer.CreateResponse[DynamicDnsState]{Output: DynamicDnsState{DynamicDnsArgs: req.Inputs}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	created, err := cfg.Network().CreateDynamicDNS(ctx, cfg.ResolvedSite(), req.Inputs.toUnifi(""))
	if err != nil {
		return infer.CreateResponse[DynamicDnsState]{}, wrap(fmt.Sprintf("create dynamic DNS entry %q (site %q)", req.Inputs.HostName, cfg.ResolvedSite()), err)
	}
	if created.ID == "" {
		return infer.CreateResponse[DynamicDnsState]{}, infer.ProviderErrorf("created dynamic DNS entry but controller returned no ID")
	}
	return infer.CreateResponse[DynamicDnsState]{ID: created.ID, Output: dynamicDnsStateFrom(created, req.Inputs)}, nil
}

// Read recovers state from the controller, enabling `pulumi import`.
func (DynamicDns) Read(ctx context.Context, req infer.ReadRequest[DynamicDnsArgs, DynamicDnsState]) (infer.ReadResponse[DynamicDnsArgs, DynamicDnsState], error) {
	cfg := infer.GetConfig[config.Config](ctx)
	u, err := cfg.Network().GetDynamicDNS(ctx, cfg.ResolvedSite(), req.ID)
	if notFound(err) {
		return infer.ReadResponse[DynamicDnsArgs, DynamicDnsState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[DynamicDnsArgs, DynamicDnsState]{}, wrap(fmt.Sprintf("read dynamic DNS entry %q (site %q)", req.ID, cfg.ResolvedSite()), err)
	}
	st := dynamicDnsStateFrom(u, req.Inputs)
	return infer.ReadResponse[DynamicDnsArgs, DynamicDnsState]{ID: req.ID, Inputs: st.DynamicDnsArgs, State: st}, nil
}

// Update applies changed inputs in place.
func (DynamicDns) Update(ctx context.Context, req infer.UpdateRequest[DynamicDnsArgs, DynamicDnsState]) (infer.UpdateResponse[DynamicDnsState], error) {
	if req.DryRun {
		return infer.UpdateResponse[DynamicDnsState]{Output: DynamicDnsState{DynamicDnsArgs: req.Inputs, DynamicDnsId: req.ID}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	updated, err := cfg.Network().UpdateDynamicDNS(ctx, cfg.ResolvedSite(), req.Inputs.toUnifi(req.ID))
	if err != nil {
		return infer.UpdateResponse[DynamicDnsState]{}, wrap(fmt.Sprintf("update dynamic DNS entry %q (site %q)", req.ID, cfg.ResolvedSite()), err)
	}
	return infer.UpdateResponse[DynamicDnsState]{Output: dynamicDnsStateFrom(updated, req.Inputs)}, nil
}

// Delete removes the dynamic DNS entry.
func (DynamicDns) Delete(ctx context.Context, req infer.DeleteRequest[DynamicDnsState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[config.Config](ctx)
	err := cfg.Network().DeleteDynamicDNS(ctx, cfg.ResolvedSite(), req.ID)
	if notFound(err) {
		return infer.DeleteResponse{}, nil
	}
	return infer.DeleteResponse{}, wrap(fmt.Sprintf("delete dynamic DNS entry %q (site %q)", req.ID, cfg.ResolvedSite()), err)
}
