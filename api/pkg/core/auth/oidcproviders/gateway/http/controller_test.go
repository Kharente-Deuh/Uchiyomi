// SPDX-License-Identifier: AGPL-3.0-or-later

package http_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidcproviders"
	providershttp "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidcproviders/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions"
	sessionshttp "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
)

const (
	endpoint        = "/oidc/providers"
	cookieName      = "uchiyomi_session"
	testToken       = "letoken"
	testDisplayName = "Keycloak"
	testIssuerURL   = "https://sso.example.com"
	testUsername    = "alice"

	//nolint:lll
	validBody = `{"displayName":"Keycloak","issuerUrl":"https://sso.example.com","clientId":"uchiyomi","clientSecret":"s3cr3t","usernameClaim":"preferred_username","scopes":["openid"],"roleClaim":null,"adminValues":null,"allowedValues":null,"autoProvision":true}`

	//nolint:lll
	validPutBody = `{"displayName":"Keycloak","issuerUrl":"https://sso.example.com","clientId":"uchiyomi","usernameClaim":"preferred_username","scopes":["openid"],"roleClaim":null,"adminValues":null,"allowedValues":null,"autoProvision":true}`
)

var errUnexpected = errors.New("boom")

type stubService struct {
	err       error
	provider  *oidcproviders.OIDCProvider
	probe     *oidcproviders.ProbeResult
	updateOpt *oidcproviders.UpdateOpts
	list      []oidcproviders.LightOIDCProvider
	users     []oidcproviders.OIDCProviderUser
}

func (s *stubService) List(context.Context) ([]oidcproviders.LightOIDCProvider, error) {
	return s.list, s.err
}

//nolint:lll
func (s *stubService) GetByID(context.Context, uuid.UUID) (*oidcproviders.OIDCProviderDetails, error) {
	if s.err != nil {
		return nil, s.err
	}

	return &oidcproviders.OIDCProviderDetails{Provider: *s.provider, Users: s.users}, nil
}

//nolint:lll
func (s *stubService) Create(context.Context, oidcproviders.CreateOpts) (*oidcproviders.OIDCProvider, error) {
	return s.provider, s.err
}

//nolint:lll
func (s *stubService) Update(_ context.Context, _ uuid.UUID, opts oidcproviders.UpdateOpts) (*oidcproviders.OIDCProvider, error) {
	s.updateOpt = &opts

	return s.provider, s.err
}

func (s *stubService) Delete(context.Context, uuid.UUID) error {
	return s.err
}

func (s *stubService) Probe(context.Context, string) (*oidcproviders.ProbeResult, error) {
	return s.probe, s.err
}

type stubSessionService struct {
	user *users.User
}

//nolint:lll
func (s *stubSessionService) Authenticate(_ context.Context, _ string) (*sessions.AuthenticatedSession, error) {
	if s.user == nil {
		return nil, sessions.ErrInvalidSession
	}

	return &sessions.AuthenticatedSession{
		User: s.user,
		Session: sessions.Session{
			ID:        uuid.New(),
			UserID:    s.user.ID,
			ExpiresAt: time.Now().Add(time.Hour),
		},
	}, nil
}

func adminMiddlewares(t *testing.T, user *users.User) chi.Middlewares {
	t.Helper()

	logger := slog.New(slog.DiscardHandler)

	cookies, err := sessionshttp.NewCookieManager(sessionshttp.CookieConfig{Name: cookieName, Path: "/"})
	if err != nil {
		t.Fatalf("NewCookieManager: %v", err)
	}

	a, err := sessionshttp.NewAuthenticator(sessionshttp.AuthenticatorDeps{
		SessionService: &stubSessionService{user: user},
		Cookies:        cookies,
		Logger:         logger,
	})
	if err != nil {
		t.Fatalf("NewAuthenticator: %v", err)
	}

	return chi.Middlewares{a.Middleware, a.RequireAdmin}
}

func newRouter(t *testing.T, svc *stubService, mws chi.Middlewares) chi.Router {
	t.Helper()

	c, err := providershttp.New(
		providershttp.Config{Endpoint: endpoint, Middlewares: mws},
		providershttp.Deps{Logger: slog.New(slog.DiscardHandler), Service: svc},
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	r := chi.NewRouter()
	c.InitRouter(r)

	return r
}

func do(r chi.Router, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	return rec
}

func sampleProvider() *oidcproviders.OIDCProvider {
	return &oidcproviders.OIDCProvider{
		ID:            uuid.New(),
		DisplayName:   testDisplayName,
		IssuerURL:     testIssuerURL,
		ClientID:      "uchiyomi",
		UsernameClaim: "preferred_username",
		Scopes:        []string{"openid"},
		AutoProvision: true,
	}
}

func admin() *users.User {
	return &users.User{ID: uuid.New(), Name: "root", IsAdmin: true}
}

func TestListReturnsOnlyTheLightFields(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	svc := &stubService{list: []oidcproviders.LightOIDCProvider{
		{ID: id, DisplayName: testDisplayName, CreatedAt: time.Now(), UserCount: 3},
	}}
	r := newRouter(t, svc, adminMiddlewares(t, admin()))

	rec := do(r, http.MethodGet, endpoint, "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	var got []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("undecodable body (%q): %v", rec.Body.String(), err)
	}

	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}

	want := map[string]bool{"id": true, "displayName": true, "createdAt": true, "userCount": true}
	for key := range got[0] {
		if !want[key] {
			t.Errorf("the list exposes %q, which belongs to the detail response", key)
		}
	}

	if len(got[0]) != len(want) {
		t.Errorf("%d fields in the list entry, want %d: %s", len(got[0]), len(want), rec.Body.String())
	}

	if got[0]["id"] != id.String() {
		t.Errorf("id = %v, want %s", got[0]["id"], id)
	}

	if got[0]["userCount"] != float64(3) {
		t.Errorf("userCount = %v, want 3", got[0]["userCount"])
	}
}

func TestGetReturnsTheProviderWithoutASecret(t *testing.T) {
	t.Parallel()

	svc := &stubService{provider: sampleProvider()}
	r := newRouter(t, svc, adminMiddlewares(t, admin()))

	rec := do(r, http.MethodGet, endpoint+"/"+uuid.New().String(), "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("undecodable body (%q): %v", rec.Body.String(), err)
	}

	for key := range got {
		if strings.Contains(strings.ToLower(key), "secret") {
			t.Errorf("the response exposes a secret-looking field %q", key)
		}
	}

	if got["issuerUrl"] != testIssuerURL {
		t.Errorf("issuerUrl = %v, want the full provider", got["issuerUrl"])
	}
}

func TestGetReturnsTheLinkedUsers(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	linkedAt := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	svc := &stubService{
		provider: sampleProvider(),
		users: []oidcproviders.OIDCProviderUser{
			{ID: userID, Username: testUsername, LinkedAt: linkedAt, IsAdmin: true},
		},
	}
	r := newRouter(t, svc, adminMiddlewares(t, admin()))

	rec := do(r, http.MethodGet, endpoint+"/"+uuid.New().String(), "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	var got struct {
		Users []map[string]any `json:"users"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("undecodable body (%q): %v", rec.Body.String(), err)
	}

	if len(got.Users) != 1 {
		t.Fatalf("users = %+v, want one entry", got.Users)
	}

	if got.Users[0]["id"] != userID.String() || got.Users[0]["username"] != testUsername ||
		got.Users[0]["isAdmin"] != true || got.Users[0]["linkedAt"] != "2026-01-02T03:04:05Z" {
		t.Errorf("users[0] = %+v", got.Users[0])
	}
}

func TestGetReturnsAnEmptyUserListWhenNobodyIsLinked(t *testing.T) {
	t.Parallel()

	r := newRouter(t, &stubService{provider: sampleProvider()}, adminMiddlewares(t, admin()))

	rec := do(r, http.MethodGet, endpoint+"/"+uuid.New().String(), "")

	if !strings.Contains(rec.Body.String(), `"users":[]`) {
		t.Errorf("body = %s, want an empty users array", rec.Body.String())
	}
}

func TestGetNotFound(t *testing.T) {
	t.Parallel()

	r := newRouter(t, &stubService{err: domain.ErrNotFound}, adminMiddlewares(t, admin()))

	rec := do(r, http.MethodGet, endpoint+"/"+uuid.New().String(), "")

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d (body: %s)", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestGetRejectsAMalformedID(t *testing.T) {
	t.Parallel()

	r := newRouter(t, &stubService{provider: sampleProvider()}, adminMiddlewares(t, admin()))

	rec := do(r, http.MethodGet, endpoint+"/not-a-uuid", "")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d (body: %s)", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestListForbiddenForANonAdmin(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: testUsername, IsAdmin: false}
	r := newRouter(t, &stubService{}, adminMiddlewares(t, user))

	rec := do(r, http.MethodGet, endpoint, "")

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d (body: %s)", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestEveryRouteIsForbiddenForANonAdmin(t *testing.T) {
	t.Parallel()

	id := uuid.New().String()

	tests := map[string]struct {
		method string
		path   string
		body   string
	}{
		"list":   {method: http.MethodGet, path: endpoint},
		"get":    {method: http.MethodGet, path: endpoint + "/" + id},
		"create": {method: http.MethodPost, path: endpoint, body: validBody},
		"update": {method: http.MethodPut, path: endpoint + "/" + id, body: validPutBody},
		"delete": {method: http.MethodDelete, path: endpoint + "/" + id},
		"probe":  {method: http.MethodPost, path: endpoint + "/probe", body: `{"issuerUrl":"https://sso.example.com"}`},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			user := &users.User{ID: uuid.New(), Name: testUsername}
			r := newRouter(t, &stubService{}, adminMiddlewares(t, user))

			rec := do(r, tc.method, tc.path, tc.body)

			if rec.Code != http.StatusForbidden {
				t.Errorf("%s %s = %d, want %d", tc.method, tc.path, rec.Code, http.StatusForbidden)
			}
		})
	}
}

func TestCreateReturnsCreated(t *testing.T) {
	t.Parallel()

	svc := &stubService{provider: sampleProvider()}
	r := newRouter(t, svc, adminMiddlewares(t, admin()))

	rec := do(r, http.MethodPost, endpoint, validBody)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var got providershttp.ProviderResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("undecodable body (%q): %v", rec.Body.String(), err)
	}

	if got.DisplayName != testDisplayName {
		t.Errorf("displayName = %q", got.DisplayName)
	}
}

func TestCreateAcceptsEmptyValueListsWithoutARoleClaim(t *testing.T) {
	t.Parallel()

	r := newRouter(t, &stubService{provider: sampleProvider()}, adminMiddlewares(t, admin()))

	//nolint:lll
	body := `{"displayName":"K","issuerUrl":"https://s.example.com","clientId":"c","clientSecret":"s","usernameClaim":"u","scopes":["openid"],"adminValues":[],"allowedValues":[]}`

	rec := do(r, http.MethodPost, endpoint, body)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d (body: %s)", rec.Code, http.StatusCreated, rec.Body.String())
	}
}

func TestUpdateRejectsValuesWithoutARoleClaim(t *testing.T) {
	t.Parallel()

	r := newRouter(t, &stubService{provider: sampleProvider()}, adminMiddlewares(t, admin()))

	//nolint:lll
	body := `{"displayName":"K","issuerUrl":"https://s.example.com","clientId":"c","clientSecret":null,"usernameClaim":"u","scopes":["openid"],"adminValues":["admins"]}`

	rec := do(r, http.MethodPut, endpoint+"/"+uuid.New().String(), body)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d (body: %s)", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestCreateRejectsAMalformedBody(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"not json":         `{`,
		"missing fields":   `{"displayName":"Keycloak"}`,
		"bad issuer":       `{"displayName":"K","issuerUrl":"nope","clientId":"c","clientSecret":"s","usernameClaim":"u","scopes":["openid"]}`,                           //nolint:lll
		"unknown field":    `{"displayName":"K","issuerUrl":"https://s.example.com","clientId":"c","clientSecret":"s","usernameClaim":"u","scopes":["openid"],"nope":1}`, //nolint:lll
		"no scope":         `{"displayName":"K","issuerUrl":"https://s.example.com","clientId":"c","clientSecret":"s","usernameClaim":"u","scopes":[]}`,                  //nolint:lll
		"no client secret": `{"displayName":"K","issuerUrl":"https://s.example.com","clientId":"c","usernameClaim":"u","scopes":["openid"]}`,                             //nolint:lll
		//nolint:lll
		"admin values without a role claim": `{"displayName":"K","issuerUrl":"https://s.example.com","clientId":"c","clientSecret":"s","usernameClaim":"u","scopes":["openid"],"adminValues":["admins"]}`,
		//nolint:lll
		"allowed values with a blank role claim": `{"displayName":"K","issuerUrl":"https://s.example.com","clientId":"c","clientSecret":"s","usernameClaim":"u","scopes":["openid"],"roleClaim":"  ","allowedValues":["users"]}`,
	}

	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			r := newRouter(t, &stubService{provider: sampleProvider()}, adminMiddlewares(t, admin()))

			rec := do(r, http.MethodPost, endpoint, body)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d (body: %s)", rec.Code, http.StatusBadRequest, rec.Body.String())
			}
		})
	}
}

func TestCreateConflictOnADuplicateIssuer(t *testing.T) {
	t.Parallel()

	r := newRouter(t, &stubService{err: domain.ErrAlreadyExists}, adminMiddlewares(t, admin()))

	rec := do(r, http.MethodPost, endpoint, validBody)

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d (body: %s)", rec.Code, http.StatusConflict, rec.Body.String())
	}
}

func TestCreateBadRequestOnAnUnreachableIssuer(t *testing.T) {
	t.Parallel()

	r := newRouter(t, &stubService{err: oidcproviders.ErrUnreachableIssuer}, adminMiddlewares(t, admin()))

	rec := do(r, http.MethodPost, endpoint, validBody)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d (body: %s)", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestCreateBadRequestOnAnIncompleteDiscoveryDocument(t *testing.T) {
	t.Parallel()

	r := newRouter(t, &stubService{err: oidcproviders.ErrIncompleteIssuer}, adminMiddlewares(t, admin()))

	rec := do(r, http.MethodPost, endpoint, validBody)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d (body: %s)", rec.Code, http.StatusBadRequest, rec.Body.String())
	}

	if strings.Contains(rec.Body.String(), "unreachable") {
		t.Errorf("an incomplete discovery document must not be reported as unreachable: %s", rec.Body.String())
	}
}

func TestUpdateForwardsTheRequest(t *testing.T) {
	t.Parallel()

	svc := &stubService{provider: sampleProvider()}
	r := newRouter(t, svc, adminMiddlewares(t, admin()))

	rec := do(r, http.MethodPut, endpoint+"/"+uuid.New().String(), validPutBody)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	if svc.updateOpt == nil {
		t.Fatal("the service was never called")
	}

	if svc.updateOpt.IssuerURL != testIssuerURL {
		t.Errorf("IssuerURL = %q, want %q", svc.updateOpt.IssuerURL, testIssuerURL)
	}
}

func TestUpdateRejectsAClientSecret(t *testing.T) {
	t.Parallel()

	svc := &stubService{provider: sampleProvider()}
	r := newRouter(t, svc, adminMiddlewares(t, admin()))

	//nolint:lll
	body := `{"displayName":"Keycloak","issuerUrl":"https://sso.example.com","clientId":"uchiyomi","clientSecret":"rotated","usernameClaim":"preferred_username","scopes":["openid"],"roleClaim":null,"adminValues":null,"allowedValues":null,"autoProvision":true}`

	rec := do(r, http.MethodPut, endpoint+"/"+uuid.New().String(), body)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusBadRequest, rec.Body.String())
	}

	if svc.updateOpt != nil {
		t.Error("the service was called despite the rejected body")
	}
}

func TestUpdateNotFound(t *testing.T) {
	t.Parallel()

	r := newRouter(t, &stubService{err: domain.ErrNotFound}, adminMiddlewares(t, admin()))

	rec := do(r, http.MethodPut, endpoint+"/"+uuid.New().String(), validPutBody)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d (body: %s)", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestUpdateRejectsAMalformedID(t *testing.T) {
	t.Parallel()

	r := newRouter(t, &stubService{provider: sampleProvider()}, adminMiddlewares(t, admin()))

	rec := do(r, http.MethodPut, endpoint+"/not-a-uuid", validPutBody)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d (body: %s)", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestDeleteReturnsNoContent(t *testing.T) {
	t.Parallel()

	r := newRouter(t, &stubService{}, adminMiddlewares(t, admin()))

	rec := do(r, http.MethodDelete, endpoint+"/"+uuid.New().String(), "")

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusNoContent, rec.Body.String())
	}

	if rec.Body.Len() != 0 {
		t.Errorf("body = %q, want empty", rec.Body.String())
	}
}

func TestDeleteNotFound(t *testing.T) {
	t.Parallel()

	r := newRouter(t, &stubService{err: domain.ErrNotFound}, adminMiddlewares(t, admin()))

	rec := do(r, http.MethodDelete, endpoint+"/"+uuid.New().String(), "")

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestProbeReturnsTheDiscoveredEndpoints(t *testing.T) {
	t.Parallel()

	svc := &stubService{probe: &oidcproviders.ProbeResult{
		Issuer:                    testIssuerURL,
		AuthorizationEndpoint:     "https://sso.example.com/auth",
		TokenEndpoint:             "https://sso.example.com/token",
		EndSessionEndpoint:        "https://sso.example.com/logout",
		RedirectURI:               "https://manga.example.com/api/auth/oidc/callback",
		SupportsRPInitiatedLogout: true,
	}}
	r := newRouter(t, svc, adminMiddlewares(t, admin()))

	rec := do(r, http.MethodPost, endpoint+"/probe", `{"issuerUrl":"https://sso.example.com"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	var got providershttp.ProbeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("undecodable body (%q): %v", rec.Body.String(), err)
	}

	if !got.SupportsRPInitiatedLogout {
		t.Error("supportsRpInitiatedLogout = false, want true")
	}

	if got.TokenEndpoint != "https://sso.example.com/token" {
		t.Errorf("tokenEndpoint = %q", got.TokenEndpoint)
	}
}

func TestProbeBadRequestOnAnUnreachableIssuer(t *testing.T) {
	t.Parallel()

	r := newRouter(t, &stubService{err: oidcproviders.ErrUnreachableIssuer}, adminMiddlewares(t, admin()))

	rec := do(r, http.MethodPost, endpoint+"/probe", `{"issuerUrl":"https://sso.example.com"}`)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d (body: %s)", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestProbeBadRequestOnAnIncompleteDiscoveryDocument(t *testing.T) {
	t.Parallel()

	r := newRouter(t, &stubService{err: oidcproviders.ErrIncompleteIssuer}, adminMiddlewares(t, admin()))

	rec := do(r, http.MethodPost, endpoint+"/probe", `{"issuerUrl":"https://sso.example.com"}`)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d (body: %s)", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestInternalErrorIsNotLeaked(t *testing.T) {
	t.Parallel()

	r := newRouter(t, &stubService{err: errUnexpected}, adminMiddlewares(t, admin()))

	rec := do(r, http.MethodGet, endpoint, "")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	if strings.Contains(rec.Body.String(), "boom") {
		t.Errorf("the internal error leaked in the body: %s", rec.Body.String())
	}
}
