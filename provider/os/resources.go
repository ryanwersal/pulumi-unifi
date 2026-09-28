// SPDX-License-Identifier: Apache-2.0

// Package os manages resources in the private UniFi OS Users API.
package os

import (
	"context"
	"errors"
	"fmt"

	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/ryanwersal/pulumi-unifi/provider/config"
	"github.com/ryanwersal/pulumi-unifi/provider/internal/usersapi"
)

type CustomRole struct{}

type CustomRoleArgs struct {
	Name        string `pulumi:"name"`
	NetworkRole string `pulumi:"networkRole,optional"`
	ProtectRole string `pulumi:"protectRole,optional"`
	AccessRole  string `pulumi:"accessRole,optional"`
	TalkRole    string `pulumi:"talkRole,optional"`
}

type CustomRoleState struct {
	CustomRoleArgs
	RoleId string `pulumi:"roleId"`
}

func (s *CustomRoleState) Annotate(a infer.Annotator) {
	a.Describe(&s.RoleId, "Console-assigned role identifier.")
}

func (r *CustomRole) Annotate(a infer.Annotator) {
	a.Describe(&r, "A custom UniFi OS console role, managed through the private Users API.")
}

func (a *CustomRoleArgs) Annotate(an infer.Annotator) {
	an.Describe(&a.Name, "Role name.")
	an.Describe(&a.NetworkRole, `Network access: "admin", "hotspot", "view", "site_admin", or "none".`)
	an.SetDefault(&a.NetworkRole, "none")
	an.Describe(&a.ProtectRole, `Protect access: "admin", "live", "view", or "none".`)
	an.SetDefault(&a.ProtectRole, "none")
	an.Describe(&a.AccessRole, `Access access: "admin" or "none".`)
	an.SetDefault(&a.AccessRole, "none")
	an.Describe(&a.TalkRole, `Talk access: "admin", "view", or "none".`)
	an.SetDefault(&a.TalkRole, "none")
}

func rolePermissions(a CustomRoleArgs) []usersapi.Permission {
	out := []usersapi.Permission{{
		ManifestUniqueKey: "system.management.user",
		Formulas:          []string{"admin"},
	}}
	values := []struct{ key, formula string }{
		{"network.management", a.NetworkRole},
		{"protect.management", a.ProtectRole},
		{"access.management", a.AccessRole},
		{"talk.management", a.TalkRole},
	}
	for _, value := range values {
		if value.formula != "" && value.formula != "none" {
			out = append(out, usersapi.Permission{ManifestUniqueKey: value.key, Formulas: []string{value.formula}})
		}
	}
	return out
}

func roleState(role *usersapi.Role, args CustomRoleArgs) CustomRoleState {
	return CustomRoleState{CustomRoleArgs: args, RoleId: role.ID}
}

func (CustomRole) Create(ctx context.Context, req infer.CreateRequest[CustomRoleArgs]) (infer.CreateResponse[CustomRoleState], error) {
	if req.DryRun {
		return infer.CreateResponse[CustomRoleState]{Output: CustomRoleState{CustomRoleArgs: req.Inputs}}, nil
	}
	client, err := infer.GetConfig[config.Config](ctx).UsersAdmin()
	if err != nil {
		return infer.CreateResponse[CustomRoleState]{}, err
	}
	role, err := client.CreateRole(ctx, req.Inputs.Name, rolePermissions(req.Inputs))
	if err != nil {
		return infer.CreateResponse[CustomRoleState]{}, fmt.Errorf("create UniFi OS role %q: %w", req.Inputs.Name, err)
	}
	return infer.CreateResponse[CustomRoleState]{ID: role.ID, Output: roleState(role, req.Inputs)}, nil
}

func (CustomRole) Read(ctx context.Context, req infer.ReadRequest[CustomRoleArgs, CustomRoleState]) (infer.ReadResponse[CustomRoleArgs, CustomRoleState], error) {
	client, err := infer.GetConfig[config.Config](ctx).UsersAdmin()
	if err != nil {
		return infer.ReadResponse[CustomRoleArgs, CustomRoleState]{}, err
	}
	role, err := client.GetRole(ctx, req.ID)
	if errors.Is(err, usersapi.ErrRoleNotFound) {
		return infer.ReadResponse[CustomRoleArgs, CustomRoleState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[CustomRoleArgs, CustomRoleState]{}, err
	}
	state := roleState(role, req.Inputs)
	return infer.ReadResponse[CustomRoleArgs, CustomRoleState]{ID: req.ID, Inputs: state.CustomRoleArgs, State: state}, nil
}

func (CustomRole) Update(ctx context.Context, req infer.UpdateRequest[CustomRoleArgs, CustomRoleState]) (infer.UpdateResponse[CustomRoleState], error) {
	client, err := infer.GetConfig[config.Config](ctx).UsersAdmin()
	if err != nil {
		return infer.UpdateResponse[CustomRoleState]{}, err
	}
	role, err := client.UpdateRole(ctx, req.ID, req.Inputs.Name, rolePermissions(req.Inputs))
	if err != nil {
		return infer.UpdateResponse[CustomRoleState]{}, err
	}
	return infer.UpdateResponse[CustomRoleState]{Output: roleState(role, req.Inputs)}, nil
}

func (CustomRole) Delete(ctx context.Context, req infer.DeleteRequest[CustomRoleState]) (infer.DeleteResponse, error) {
	client, err := infer.GetConfig[config.Config](ctx).UsersAdmin()
	if err != nil {
		return infer.DeleteResponse{}, err
	}
	return infer.DeleteResponse{}, client.DeleteRole(ctx, req.ID)
}

type LocalAdmin struct{}

type LocalAdminArgs struct {
	Username  string `pulumi:"username" provider:"replaceOnChanges"`
	Password  string `pulumi:"password" provider:"secret"`
	FirstName string `pulumi:"firstName" provider:"replaceOnChanges"`
	LastName  string `pulumi:"lastName" provider:"replaceOnChanges"`
	Email     string `pulumi:"email" provider:"replaceOnChanges"`
	RoleId    string `pulumi:"roleId" provider:"replaceOnChanges"`
}

type LocalAdminState struct {
	LocalAdminArgs
	UserId string `pulumi:"userId"`
}

func (s *LocalAdminState) Annotate(a infer.Annotator) {
	a.Describe(&s.UserId, "Console-assigned user identifier.")
}

func (r *LocalAdmin) Annotate(a infer.Annotator) {
	a.Describe(&r, "A local-only UniFi OS console administrator, managed through the private Users API.")
}

func (a *LocalAdminArgs) Annotate(an infer.Annotator) {
	an.Describe(&a.Username, "Local login username. Changing it replaces the user.")
	an.Describe(&a.Password, "Local login password. Stored as a Pulumi secret.")
	an.Describe(&a.FirstName, "First name.")
	an.Describe(&a.LastName, "Last name.")
	an.Describe(&a.Email, "Local account email label; it need not be deliverable.")
	an.Describe(&a.RoleId, "UniFi OS role ID.")
}

func userState(user *usersapi.LocalAdmin, args LocalAdminArgs) LocalAdminState {
	roleID := user.RoleID
	if roleID == "" && len(user.Roles) != 0 {
		roleID = user.Roles[0].ID
	}
	return LocalAdminState{LocalAdminArgs: LocalAdminArgs{
		Username:  user.Username,
		Password:  args.Password, // write-only; UniFi never returns it
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		RoleId:    roleID,
	}, UserId: user.ID}
}

func (LocalAdmin) Create(ctx context.Context, req infer.CreateRequest[LocalAdminArgs]) (infer.CreateResponse[LocalAdminState], error) {
	if req.DryRun {
		return infer.CreateResponse[LocalAdminState]{Output: LocalAdminState{LocalAdminArgs: req.Inputs}}, nil
	}
	client, err := infer.GetConfig[config.Config](ctx).UsersAdmin()
	if err != nil {
		return infer.CreateResponse[LocalAdminState]{}, err
	}
	user, err := client.CreateLocalAdmin(ctx, usersapi.LocalAdminSpec{
		Username: req.Inputs.Username, Password: req.Inputs.Password, FirstName: req.Inputs.FirstName,
		LastName: req.Inputs.LastName, Email: req.Inputs.Email, RoleID: req.Inputs.RoleId,
	})
	if err != nil {
		return infer.CreateResponse[LocalAdminState]{}, fmt.Errorf("create local UniFi OS admin %q: %w", req.Inputs.Username, err)
	}
	return infer.CreateResponse[LocalAdminState]{ID: user.ID, Output: userState(user, req.Inputs)}, nil
}

func (LocalAdmin) Read(ctx context.Context, req infer.ReadRequest[LocalAdminArgs, LocalAdminState]) (infer.ReadResponse[LocalAdminArgs, LocalAdminState], error) {
	client, err := infer.GetConfig[config.Config](ctx).UsersAdmin()
	if err != nil {
		return infer.ReadResponse[LocalAdminArgs, LocalAdminState]{}, err
	}
	user, err := client.GetLocalAdmin(ctx, req.ID)
	if errors.Is(err, usersapi.ErrUserNotFound) {
		return infer.ReadResponse[LocalAdminArgs, LocalAdminState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[LocalAdminArgs, LocalAdminState]{}, err
	}
	state := userState(user, req.Inputs)
	return infer.ReadResponse[LocalAdminArgs, LocalAdminState]{ID: req.ID, Inputs: state.LocalAdminArgs, State: state}, nil
}

// Update hydrates a write-only password during import. Every readable field is
// replace-only, so password is the only input that can reach this method. UniFi
// does not return password material; verifying the live user and then retaining
// the configured secret is the only non-destructive way to adopt an account.
func (LocalAdmin) Update(ctx context.Context, req infer.UpdateRequest[LocalAdminArgs, LocalAdminState]) (infer.UpdateResponse[LocalAdminState], error) {
	client, err := infer.GetConfig[config.Config](ctx).UsersAdmin()
	if err != nil {
		return infer.UpdateResponse[LocalAdminState]{}, err
	}
	user, err := client.GetLocalAdmin(ctx, req.ID)
	if err != nil {
		return infer.UpdateResponse[LocalAdminState]{}, err
	}
	state := userState(user, req.Inputs)
	return infer.UpdateResponse[LocalAdminState]{Output: state}, nil
}

func (LocalAdmin) Delete(ctx context.Context, req infer.DeleteRequest[LocalAdminState]) (infer.DeleteResponse, error) {
	client, err := infer.GetConfig[config.Config](ctx).UsersAdmin()
	if err != nil {
		return infer.DeleteResponse{}, err
	}
	return infer.DeleteResponse{}, client.DeleteLocalAdmin(ctx, req.ID)
}

type APIKey struct{}

type APIKeyArgs struct {
	Name     string `pulumi:"name" provider:"replaceOnChanges"`
	Username string `pulumi:"username" provider:"replaceOnChanges"`
	Password string `pulumi:"password" provider:"secret"`
}

type APIKeyState struct {
	APIKeyArgs
	KeyId  string `pulumi:"keyId"`
	APIKey string `pulumi:"apiKey" provider:"secret"`
}

func (a *APIKeyArgs) Annotate(an infer.Annotator) {
	an.Describe(&a.Name, "Display name for the integration key.")
	an.Describe(&a.Username, "Local account that will own the key.")
	an.Describe(&a.Password, "Password for the local account. Stored as a Pulumi secret.")
}

func (s *APIKeyState) Annotate(a infer.Annotator) {
	a.Describe(&s.Name, "Display name for the integration key.")
	a.Describe(&s.Username, "Local account that owns the key.")
	a.Describe(&s.Password, "Password for the local account. Stored as a Pulumi secret.")
	a.Describe(&s.KeyId, "Console-assigned API key identifier.")
	a.Describe(&s.APIKey, "Full API key. Returned only at creation and stored as a Pulumi secret.")
}

func (r *APIKey) Annotate(a infer.Annotator) {
	a.Describe(&r, "A UniFi OS integration API key owned by the supplied local user. The full key is returned once and stored as a Pulumi secret.")
}

func (APIKey) Create(ctx context.Context, req infer.CreateRequest[APIKeyArgs]) (infer.CreateResponse[APIKeyState], error) {
	if req.DryRun {
		return infer.CreateResponse[APIKeyState]{Output: APIKeyState{APIKeyArgs: req.Inputs}}, nil
	}
	cfg := infer.GetConfig[config.Config](ctx)
	client := cfg.UsersFor(req.Inputs.Username, req.Inputs.Password)
	key, err := client.CreateAPIKey(ctx, req.Inputs.Name)
	if err != nil {
		return infer.CreateResponse[APIKeyState]{}, fmt.Errorf("create UniFi OS API key %q: %w", req.Inputs.Name, err)
	}
	state := APIKeyState{APIKeyArgs: req.Inputs, KeyId: key.ID, APIKey: key.Key}
	return infer.CreateResponse[APIKeyState]{ID: key.ID, Output: state}, nil
}

func (APIKey) Read(ctx context.Context, req infer.ReadRequest[APIKeyArgs, APIKeyState]) (infer.ReadResponse[APIKeyArgs, APIKeyState], error) {
	cfg := infer.GetConfig[config.Config](ctx)
	keys, err := cfg.UsersFor(req.Inputs.Username, req.Inputs.Password).ListAPIKeys(ctx)
	if err != nil {
		return infer.ReadResponse[APIKeyArgs, APIKeyState]{}, err
	}
	for _, key := range keys {
		if key.ID == req.ID {
			state := req.State
			state.Name = key.Name
			return infer.ReadResponse[APIKeyArgs, APIKeyState]{ID: req.ID, Inputs: state.APIKeyArgs, State: state}, nil
		}
	}
	return infer.ReadResponse[APIKeyArgs, APIKeyState]{}, nil
}

func (APIKey) Delete(ctx context.Context, req infer.DeleteRequest[APIKeyState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[config.Config](ctx)
	return infer.DeleteResponse{}, cfg.UsersFor(req.State.Username, req.State.Password).DeleteAPIKey(ctx, req.ID)
}
