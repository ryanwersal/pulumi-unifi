// SPDX-License-Identifier: Apache-2.0

// Package usersapi is a minimal client for the private UniFi OS Users API.
package usersapi

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"
)

var (
	ErrRoleNotFound   = errors.New("usersapi: role not found")
	ErrUserNotFound   = errors.New("usersapi: user not found")
	ErrAPIKeyNotFound = errors.New("usersapi: API key not found")
)

type Permission struct {
	ManifestUniqueKey string   `json:"manifest_unique_key"`
	Formulas          []string `json:"formulas"`
}

type Role struct {
	ID          string       `json:"unique_id"`
	Name        string       `json:"name"`
	SystemRole  bool         `json:"system_role"`
	Permissions []Permission `json:"permissions"`
}

type LocalAdmin struct {
	ID        string `json:"unique_id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"user_email"`
	RoleID    string `json:"role_id"`
	Roles     []Role `json:"roles"`
}

type LocalAdminSpec struct {
	Username  string
	Password  string
	FirstName string
	LastName  string
	Email     string
	RoleID    string
}

type APIKey struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Key  string `json:"full_api_key,omitempty"`
}

type Client interface {
	ListRoles(context.Context) ([]Role, error)
	GetRole(context.Context, string) (*Role, error)
	CreateRole(context.Context, string, []Permission) (*Role, error)
	UpdateRole(context.Context, string, string, []Permission) (*Role, error)
	DeleteRole(context.Context, string) error
	ListLocalAdmins(context.Context) ([]LocalAdmin, error)
	GetLocalAdmin(context.Context, string) (*LocalAdmin, error)
	CreateLocalAdmin(context.Context, LocalAdminSpec) (*LocalAdmin, error)
	DeleteLocalAdmin(context.Context, string) error
	ListAPIKeys(context.Context) ([]APIKey, error)
	CreateAPIKey(context.Context, string) (*APIKey, error)
	DeleteAPIKey(context.Context, string) error
}

type Config struct {
	Host               string
	APIKey             string
	Username           string
	Password           string
	InsecureSkipVerify bool
	Timeout            time.Duration
}

type httpClient struct {
	base string
	user string
	pass string
	key  string
	http *http.Client
	mu   sync.Mutex
	csrf string
	auth bool
}

const (
	rolesPath    = "/proxy/users/api/v2/custom_roles"
	rolePath     = "/proxy/users/api/v2/custom_role"
	usersPath    = "/proxy/users/api/v2/user"
	keysPath     = "/proxy/users/api/v2/keys"
	selfKeysPath = "/proxy/users/api/v2/user/self/keys"
)

func New(cfg Config) Client {
	base := strings.TrimRight(cfg.Host, "/")
	if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "https://" + base
	}
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	jar, _ := cookiejar.New(nil)
	return &httpClient{
		base: base,
		key:  cfg.APIKey,
		user: cfg.Username,
		pass: cfg.Password,
		http: &http.Client{
			Timeout: timeout,
			Jar:     jar,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.InsecureSkipVerify}, //nolint:gosec
			},
		},
	}
}

type envelope[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

func (c *httpClient) login(ctx context.Context) error {
	payload, _ := json.Marshal(map[string]string{"username": c.user, "password": c.pass})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/api/auth/login", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("usersapi login: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	c.captureCSRF(resp)
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("usersapi login: authentication failed (HTTP %d); use a local UniFi OS administrator", resp.StatusCode)
	}
	if resp.StatusCode >= 400 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("usersapi login: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	c.auth = true
	return nil
}

func (c *httpClient) captureCSRF(resp *http.Response) {
	for _, name := range []string{"X-Csrf-Token", "X-Updated-Csrf-Token"} {
		if value := resp.Header.Get(name); value != "" {
			c.csrf = value
		}
	}
}

func (c *httpClient) request(ctx context.Context, method, path string, body any) (int, []byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for attempt := 0; attempt < 2; attempt++ {
		if c.key == "" && !c.auth {
			if err := c.login(ctx); err != nil {
				return 0, nil, err
			}
		}
		var reader io.Reader
		if body != nil {
			payload, err := json.Marshal(body)
			if err != nil {
				return 0, nil, err
			}
			reader = bytes.NewReader(payload)
		}
		req, err := http.NewRequestWithContext(ctx, method, c.base+path, reader)
		if err != nil {
			return 0, nil, err
		}
		req.Header.Set("Accept", "application/json")
		if c.key != "" {
			req.Header.Set("X-API-Key", c.key)
		}
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if c.csrf != "" {
			req.Header.Set("X-Csrf-Token", c.csrf)
		}
		resp, err := c.http.Do(req)
		if err != nil {
			return 0, nil, err
		}
		c.captureCSRF(resp)
		raw, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			return resp.StatusCode, nil, readErr
		}
		if c.key == "" && resp.StatusCode == http.StatusUnauthorized && attempt == 0 {
			c.auth = false
			continue
		}
		return resp.StatusCode, raw, nil
	}
	return 0, nil, errors.New("usersapi: authentication retry exhausted")
}

func decode[T any](method, path string, status int, raw []byte) (T, error) {
	var zero T
	if status >= 400 {
		return zero, fmt.Errorf("usersapi: %s %s: HTTP %d: %s", method, path, status, strings.TrimSpace(string(raw)))
	}
	if len(raw) == 0 {
		return zero, nil
	}
	var env envelope[T]
	if err := json.Unmarshal(raw, &env); err != nil {
		return zero, fmt.Errorf("usersapi: %s %s: decode: %w", method, path, err)
	}
	if env.Code != 0 && env.Code != 1 {
		return zero, fmt.Errorf("usersapi: %s %s: %s (code %d)", method, path, env.Msg, env.Code)
	}
	return env.Data, nil
}

func (c *httpClient) ListRoles(ctx context.Context) ([]Role, error) {
	status, raw, err := c.request(ctx, http.MethodGet, rolesPath, nil)
	if err != nil {
		return nil, err
	}
	return decode[[]Role](http.MethodGet, rolesPath, status, raw)
}

func (c *httpClient) GetRole(ctx context.Context, id string) (*Role, error) {
	roles, err := c.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	for i := range roles {
		if roles[i].ID == id {
			return &roles[i], nil
		}
	}
	return nil, ErrRoleNotFound
}

func (c *httpClient) CreateRole(ctx context.Context, name string, permissions []Permission) (*Role, error) {
	body := map[string]any{"name": name, "permissions": permissions, "permission_resources": []any{}}
	status, raw, err := c.request(ctx, http.MethodPost, rolePath, body)
	if err != nil {
		return nil, err
	}
	role, err := decode[Role](http.MethodPost, rolePath, status, raw)
	return &role, err
}

func (c *httpClient) UpdateRole(ctx context.Context, id, name string, permissions []Permission) (*Role, error) {
	path := rolePath + "/" + url.PathEscape(id)
	status, raw, err := c.request(ctx, http.MethodPut, path, map[string]any{"name": name, "permissions": permissions, "permission_resources": []any{}})
	if err != nil {
		return nil, err
	}
	role, err := decode[Role](http.MethodPut, path, status, raw)
	return &role, err
}

func (c *httpClient) DeleteRole(ctx context.Context, id string) error {
	path := rolePath + "/" + url.PathEscape(id)
	status, raw, err := c.request(ctx, http.MethodDelete, path, nil)
	if err != nil || status == http.StatusNotFound {
		return err
	}
	_, err = decode[any](http.MethodDelete, path, status, raw)
	return err
}

func (c *httpClient) ListLocalAdmins(ctx context.Context) ([]LocalAdmin, error) {
	path := "/proxy/users/api/v2/users/admin/uos?page_num=1&page_size=999"
	status, raw, err := c.request(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	data, err := decode[json.RawMessage](http.MethodGet, path, status, raw)
	if err != nil {
		return nil, err
	}
	var direct []LocalAdmin
	if json.Unmarshal(data, &direct) == nil {
		return direct, nil
	}
	var paged struct {
		Items []LocalAdmin `json:"items"`
		Users []LocalAdmin `json:"users"`
		Data  []LocalAdmin `json:"data"`
	}
	if err := json.Unmarshal(data, &paged); err != nil {
		return nil, fmt.Errorf("usersapi: decode local administrator list: %w", err)
	}
	if len(paged.Items) != 0 {
		return paged.Items, nil
	}
	if len(paged.Users) != 0 {
		return paged.Users, nil
	}
	return paged.Data, nil
}

func (c *httpClient) GetLocalAdmin(ctx context.Context, id string) (*LocalAdmin, error) {
	users, err := c.ListLocalAdmins(ctx)
	if err != nil {
		return nil, err
	}
	for i := range users {
		if users[i].ID == id {
			return &users[i], nil
		}
	}
	return nil, ErrUserNotFound
}

func (c *httpClient) CreateLocalAdmin(ctx context.Context, spec LocalAdminSpec) (*LocalAdmin, error) {
	body := map[string]any{
		"first_name": spec.FirstName, "last_name": spec.LastName, "user_email": spec.Email,
		"force_add_nfc": true, "nfc_token": "", "group_ids": []string{}, "pin_code": "",
		"role_id": spec.RoleID, "username": spec.Username, "password": spec.Password,
		"only_local_account": true,
	}
	status, raw, err := c.request(ctx, http.MethodPost, usersPath, body)
	if err != nil {
		return nil, err
	}
	user, err := decode[LocalAdmin](http.MethodPost, usersPath, status, raw)
	return &user, err
}

func (c *httpClient) DeleteLocalAdmin(ctx context.Context, id string) error {
	path := usersPath + "/" + url.PathEscape(id)
	status, raw, err := c.request(ctx, http.MethodDelete, path, nil)
	if err != nil || status == http.StatusNotFound {
		return err
	}
	_, err = decode[any](http.MethodDelete, path, status, raw)
	return err
}

func (c *httpClient) ListAPIKeys(ctx context.Context) ([]APIKey, error) {
	status, raw, err := c.request(ctx, http.MethodGet, selfKeysPath, nil)
	if err != nil {
		return nil, err
	}
	return decode[[]APIKey](http.MethodGet, selfKeysPath, status, raw)
}

func (c *httpClient) CreateAPIKey(ctx context.Context, name string) (*APIKey, error) {
	status, raw, err := c.request(ctx, http.MethodPost, selfKeysPath, map[string]string{"name": name})
	if err != nil {
		return nil, err
	}
	key, err := decode[APIKey](http.MethodPost, selfKeysPath, status, raw)
	return &key, err
}

func (c *httpClient) DeleteAPIKey(ctx context.Context, id string) error {
	path := keysPath + "/" + url.PathEscape(id)
	status, raw, err := c.request(ctx, http.MethodDelete, path, nil)
	if err != nil || status == http.StatusNotFound {
		return err
	}
	_, err = decode[any](http.MethodDelete, path, status, raw)
	return err
}

var _ Client = (*httpClient)(nil)
