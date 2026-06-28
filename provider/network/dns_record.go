package network

import (
	"context"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/ryanwersal/pulumi-unifi/provider/config"
)

// DnsRecord is the controlling (marker) struct for a UniFi controller-local DNS
// record (static DNS), supported on gateways such as the UDM-SE.
type DnsRecord struct{}

// DnsRecordArgs are the user-supplied inputs for a controller-local DNS record.
type DnsRecordArgs struct {
	// Key is the hostname/record name the record answers for, e.g. "host.example.com".
	Key string `pulumi:"key"`
	// RecordType is the DNS record type: A | AAAA | CNAME | MX | NS | PTR | SOA | SRV | TXT.
	RecordType string `pulumi:"recordType"`
	// Value is the record payload (e.g. an IP for A/AAAA, a hostname for CNAME/MX/NS).
	Value string `pulumi:"value"`
	// Enabled controls whether the record is active. Defaults to true.
	Enabled *bool `pulumi:"enabled,optional"`
	// Ttl is the record time-to-live in seconds. 0 lets the controller pick a default.
	Ttl *int `pulumi:"ttl,optional"`
	// Priority is the preference value for MX and SRV records.
	Priority *int `pulumi:"priority,optional"`
	// Port is the target port for SRV records.
	Port *int `pulumi:"port,optional"`
	// Weight is the relative weight for SRV records.
	Weight *int `pulumi:"weight,optional"`
}

// DnsRecordState is the persisted state: inputs plus controller-assigned fields.
type DnsRecordState struct {
	DnsRecordArgs
	// DnsRecordId is the controller-assigned identifier (the UniFi `_id`).
	DnsRecordId string `pulumi:"dnsRecordId"`
}

// Annotate documents the resource. Must use a pointer receiver so the
// annotator can take the address of the resource and its fields.
func (d *DnsRecord) Annotate(a infer.Annotator) {
	a.Describe(&d, "A UniFi controller-local (static) DNS record. Supported on gateways such as the UDM-SE. "+
		"Maps to a controller static-dns object.")
}

// toUnifi builds a go-unifi DNSRecord from inputs. id is empty on create.
func (a DnsRecordArgs) toUnifi(id string) *unifi.DNSRecord {
	r := &unifi.DNSRecord{
		ID:         id,
		Key:        a.Key,
		RecordType: a.RecordType,
		Value:      a.Value,
		Enabled:    derefOr(a.Enabled, true),
	}
	if a.Ttl != nil {
		r.Ttl = *a.Ttl
	}
	if a.Priority != nil {
		r.Priority = *a.Priority
	}
	if a.Port != nil {
		r.Port = *a.Port
	}
	if a.Weight != nil {
		r.Weight = *a.Weight
	}
	return r
}

// dnsRecordIntPtr reflects a controller int, falling back to the prior input when zero.
func dnsRecordIntPtr(v int, prior *int) *int {
	if v != 0 {
		return ptr(v)
	}
	return prior
}

// dnsRecordStateFrom maps a controller DNSRecord back into resource state. prior
// holds the user inputs so unset optional fields are preserved across the round-trip.
func dnsRecordStateFrom(r *unifi.DNSRecord, prior DnsRecordArgs) DnsRecordState {
	args := DnsRecordArgs{
		Key:        r.Key,
		RecordType: r.RecordType,
		Value:      r.Value,
		Enabled:    ptr(r.Enabled),
		Ttl:        dnsRecordIntPtr(r.Ttl, prior.Ttl),
		Priority:   dnsRecordIntPtr(r.Priority, prior.Priority),
		Port:       dnsRecordIntPtr(r.Port, prior.Port),
		Weight:     dnsRecordIntPtr(r.Weight, prior.Weight),
	}
	return DnsRecordState{DnsRecordArgs: args, DnsRecordId: r.ID}
}

// Create provisions a new DNS record.
func (DnsRecord) Create(ctx context.Context, req infer.CreateRequest[DnsRecordArgs]) (infer.CreateResponse[DnsRecordState], error) {
	if req.DryRun {
		return infer.CreateResponse[DnsRecordState]{Output: DnsRecordState{DnsRecordArgs: req.Inputs}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	created, err := cfg.Network().CreateDNSRecord(ctx, cfg.ResolvedSite(), req.Inputs.toUnifi(""))
	if err != nil {
		return infer.CreateResponse[DnsRecordState]{}, err
	}
	return infer.CreateResponse[DnsRecordState]{ID: created.ID, Output: dnsRecordStateFrom(created, req.Inputs)}, nil
}

// Read recovers state from the controller, enabling `pulumi import`.
func (DnsRecord) Read(ctx context.Context, req infer.ReadRequest[DnsRecordArgs, DnsRecordState]) (infer.ReadResponse[DnsRecordArgs, DnsRecordState], error) {
	cfg := infer.GetConfig[config.Config](ctx)
	r, err := cfg.Network().GetDNSRecord(ctx, cfg.ResolvedSite(), req.ID)
	if err != nil {
		return infer.ReadResponse[DnsRecordArgs, DnsRecordState]{}, err
	}
	st := dnsRecordStateFrom(r, req.Inputs)
	return infer.ReadResponse[DnsRecordArgs, DnsRecordState]{ID: req.ID, Inputs: st.DnsRecordArgs, State: st}, nil
}

// Update applies changed inputs in place.
func (DnsRecord) Update(ctx context.Context, req infer.UpdateRequest[DnsRecordArgs, DnsRecordState]) (infer.UpdateResponse[DnsRecordState], error) {
	if req.DryRun {
		return infer.UpdateResponse[DnsRecordState]{Output: DnsRecordState{DnsRecordArgs: req.Inputs, DnsRecordId: req.ID}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	updated, err := cfg.Network().UpdateDNSRecord(ctx, cfg.ResolvedSite(), req.Inputs.toUnifi(req.ID))
	if err != nil {
		return infer.UpdateResponse[DnsRecordState]{}, err
	}
	return infer.UpdateResponse[DnsRecordState]{Output: dnsRecordStateFrom(updated, req.Inputs)}, nil
}

// Delete removes the DNS record.
func (DnsRecord) Delete(ctx context.Context, req infer.DeleteRequest[DnsRecordState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[config.Config](ctx)
	return infer.DeleteResponse{}, cfg.Network().DeleteDNSRecord(ctx, cfg.ResolvedSite(), req.ID)
}
