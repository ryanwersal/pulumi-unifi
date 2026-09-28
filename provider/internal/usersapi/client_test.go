// SPDX-License-Identifier: Apache-2.0

package usersapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRoleAndAPIKeyRequests(t *testing.T) {
	var authenticated bool
	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		authenticated = true
		w.Header().Set("X-Csrf-Token", "csrf")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	mux.HandleFunc(rolesPath, func(w http.ResponseWriter, r *http.Request) {
		if !authenticated || r.Header.Get("X-Csrf-Token") != "csrf" {
			t.Errorf("request was not session authenticated")
		}
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 1, "data": []map[string]any{{
					"unique_id": "role-1", "name": "HA", "permissions": []map[string]any{{
						"manifest_unique_key": "protect.management", "formulas": []string{"admin"},
					}},
				}},
			})
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	})
	mux.HandleFunc(rolePath, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Name        string       `json:"name"`
			Permissions []Permission `json:"permissions"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.Name != "HA" || len(body.Permissions) != 1 || body.Permissions[0].ManifestUniqueKey != "protect.management" {
			t.Errorf("unexpected create role body: %#v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"code": 1, "data": map[string]any{"unique_id": "role-1", "name": "HA"}})
	})
	mux.HandleFunc(selfKeysPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected keys method %s", r.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 1, "data": map[string]any{"id": "key-1", "name": "Home Assistant", "full_api_key": "secret-key"},
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := New(Config{Host: server.URL, Username: "admin", Password: "password"})
	roles, err := client.ListRoles(context.Background())
	if err != nil || len(roles) != 1 || roles[0].ID != "role-1" {
		t.Fatalf("ListRoles() = %#v, %v", roles, err)
	}
	role, err := client.CreateRole(context.Background(), "HA", []Permission{{
		ManifestUniqueKey: "protect.management", Formulas: []string{"admin"},
	}})
	if err != nil || role.ID != "role-1" {
		t.Fatalf("CreateRole() = %#v, %v", role, err)
	}
	key, err := client.CreateAPIKey(context.Background(), "Home Assistant")
	if err != nil || key.ID != "key-1" || key.Key != "secret-key" {
		t.Fatalf("CreateAPIKey() = %#v, %v", key, err)
	}
}

func TestCreateLocalAdminPayload(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/login", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	mux.HandleFunc(usersPath, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["username"] != "homeassistant" || body["only_local_account"] != true || body["role_id"] != "role-1" {
			t.Errorf("unexpected local admin body: %#v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 1, "data": map[string]any{"unique_id": "user-1", "username": "homeassistant"},
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := New(Config{Host: server.URL, Username: "admin", Password: "password"})
	user, err := client.CreateLocalAdmin(context.Background(), LocalAdminSpec{
		Username: "homeassistant", Password: "pw", FirstName: "Home", LastName: "Assistant",
		Email: "homeassistant@local.invalid", RoleID: "role-1",
	})
	if err != nil || user.ID != "user-1" {
		t.Fatalf("CreateLocalAdmin() = %#v, %v", user, err)
	}
}
