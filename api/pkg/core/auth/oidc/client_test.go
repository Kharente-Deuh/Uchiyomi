// SPDX-License-Identifier: AGPL-3.0-or-later

package oidc_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	jose "github.com/go-jose/go-jose/v4"
	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidc"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidcproviders"
)

const (
	testRedirectURI = "https://app.example.com/callback"
	testNonce       = "test-nonce"
	testSubject     = "user-123"
	testState       = "test-state"
	scopeOpenID     = "openid"
	scopeProfile    = "profile"
)

type testIdP struct {
	tokenHandler  func(w http.ResponseWriter, r *http.Request)
	srv           *httptest.Server
	key           *rsa.PrivateKey
	keyID         string
	mu            sync.Mutex
	discoveryHits int32
}

func newTestIdP(t *testing.T) *testIdP {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey: %v", err)
	}

	idp := &testIdP{key: key, keyID: "test-key-1"}

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	idp.srv = srv

	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&idp.discoveryHits, 1)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"issuer": %q,
			"authorization_endpoint": "%s/auth",
			"token_endpoint": "%s/token",
			"jwks_uri": "%s/jwks"
		}`, srv.URL, srv.URL, srv.URL, srv.URL)
	})

	mux.HandleFunc("/jwks", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		set := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{
			{Key: &key.PublicKey, KeyID: idp.keyID, Algorithm: "RS256", Use: "sig"},
		}}

		if err := json.NewEncoder(w).Encode(set); err != nil {
			t.Fatalf("json.NewEncoder.Encode: %v", err)
		}
	})

	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		idp.mu.Lock()
		h := idp.tokenHandler
		idp.mu.Unlock()

		if h == nil {
			t.Fatalf("token endpoint called without a handler set")

			return
		}

		h(w, r)
	})

	return idp
}

func (idp *testIdP) setTokenHandler(h func(w http.ResponseWriter, r *http.Request)) {
	idp.mu.Lock()
	idp.tokenHandler = h
	idp.mu.Unlock()
}

func jsonTokenResponse(rawIDToken string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w,
			`{"access_token":"test-access-token","token_type":"Bearer","expires_in":3600,"id_token":%q}`,
			rawIDToken)
	}
}

func assertingTokenResponse(
	t *testing.T,
	wantCode, wantVerifier, wantRedirectURI, rawIDToken string,
) func(w http.ResponseWriter, r *http.Request) {
	t.Helper()

	inner := jsonTokenResponse(rawIDToken)

	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("r.ParseForm: %v", err)
		}

		if got := r.PostFormValue("code"); got != wantCode {
			t.Errorf("token request code = %q, want %q", got, wantCode)
		}

		if got := r.PostFormValue("code_verifier"); got != wantVerifier {
			t.Errorf("token request code_verifier = %q, want %q", got, wantVerifier)
		}

		if got := r.PostFormValue("redirect_uri"); got != wantRedirectURI {
			t.Errorf("token request redirect_uri = %q, want %q", got, wantRedirectURI)
		}

		if got := r.PostFormValue("grant_type"); got != "authorization_code" {
			t.Errorf("token request grant_type = %q, want %q", got, "authorization_code")
		}

		inner(w, r)
	}
}

func invalidGrantResponse() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error":"invalid_grant","error_description":"the code is invalid"}`)
	}
}

func signClaims(t *testing.T, key *rsa.PrivateKey, keyID string, claims map[string]any) string {
	t.Helper()

	signer, err := jose.NewSigner(jose.SigningKey{
		Algorithm: jose.RS256,
		Key:       jose.JSONWebKey{Key: key, KeyID: keyID, Algorithm: "RS256", Use: "sig"},
	}, (&jose.SignerOptions{}).WithType("JWT"))
	if err != nil {
		t.Fatalf("jose.NewSigner: %v", err)
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	jws, err := signer.Sign(payload)
	if err != nil {
		t.Fatalf("signer.Sign: %v", err)
	}

	compact, err := jws.CompactSerialize()
	if err != nil {
		t.Fatalf("jws.CompactSerialize: %v", err)
	}

	return compact
}

func baseClaims(issuer, clientID, nonce string) map[string]any {
	now := time.Now()

	return map[string]any{
		"iss":   issuer,
		"aud":   clientID,
		"sub":   testSubject,
		"exp":   now.Add(time.Hour).Unix(),
		"iat":   now.Unix(),
		"nonce": nonce,
	}
}

func testProvider(issuer string) oidcproviders.OIDCProvider {
	return oidcproviders.OIDCProvider{
		ID:              uuid.New(),
		IssuerURL:       issuer,
		ClientID:        "test-client",
		ClientSecretEnc: []byte("encrypted-secret"),
		Scopes:          []string{scopeOpenID, scopeProfile},
	}
}

type stubDecrypter struct {
	err       error
	plaintext []byte
}

func (s stubDecrypter) Open(_ []byte) ([]byte, error) {
	return s.plaintext, s.err
}

func newTestClient(t *testing.T) *oidc.Client {
	t.Helper()

	c, err := oidc.NewClient(oidc.ClientConfig{Timeout: 5 * time.Second}, oidc.ClientDeps{
		HTTPClient: http.DefaultClient,
		Cipher:     stubDecrypter{plaintext: []byte("test-client-secret")},
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	return c
}

func TestAuthCodeURLIncludesPKCEStateNonceAndOfflineAccess(t *testing.T) {
	t.Parallel()

	idp := newTestIdP(t)
	c := newTestClient(t)
	provider := testProvider(idp.srv.URL)

	rawURL, err := c.AuthCodeURL(context.Background(), provider, oidcproviders.AuthCodeParams{
		RedirectURI: testRedirectURI,
		State:       testState,
		Nonce:       testNonce,
		Verifier:    "test-verifier-that-is-long-enough-for-pkce-1234567890",
	})
	if err != nil {
		t.Fatalf("AuthCodeURL: %v", err)
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("url.Parse: %v", err)
	}

	q := parsed.Query()

	if q.Get("state") != testState {
		t.Errorf("state = %q, want %q", q.Get("state"), testState)
	}

	if q.Get("nonce") != testNonce {
		t.Errorf("nonce = %q, want %q", q.Get("nonce"), testNonce)
	}

	if q.Get("code_challenge_method") != "S256" {
		t.Errorf("code_challenge_method = %q, want %q", q.Get("code_challenge_method"), "S256")
	}

	if q.Get("code_challenge") == "" {
		t.Error("code_challenge is empty")
	}

	scopes := strings.Fields(q.Get("scope"))

	var offlineCount int

	var hasOpenID, hasProfile bool

	for _, s := range scopes {
		switch s {
		case "offline_access":
			offlineCount++
		case scopeOpenID:
			hasOpenID = true
		case scopeProfile:
			hasProfile = true
		}
	}

	if offlineCount != 1 {
		t.Errorf("offline_access count in scope = %d, want 1 (scope=%q)", offlineCount, q.Get("scope"))
	}

	if !hasOpenID || !hasProfile {
		t.Errorf("scope = %q, want it to include openid and profile", q.Get("scope"))
	}
}

func TestAuthCodeURLAlwaysRequestsTheOpenIDScope(t *testing.T) {
	t.Parallel()

	tests := map[string][]string{
		"scopes sans openid": {scopeProfile, "email"},
		"scopes avec openid": {scopeOpenID, scopeProfile},
		"aucun scope":        nil,
	}

	for name, scopes := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			idp := newTestIdP(t)
			c := newTestClient(t)
			provider := testProvider(idp.srv.URL)
			provider.Scopes = scopes

			rawURL, err := c.AuthCodeURL(context.Background(), provider, oidcproviders.AuthCodeParams{
				RedirectURI: testRedirectURI,
				State:       testState,
				Nonce:       testNonce,
				Verifier:    "test-verifier-that-is-long-enough-for-pkce-1234567890",
			})
			if err != nil {
				t.Fatalf("AuthCodeURL: %v", err)
			}

			parsed, err := url.Parse(rawURL)
			if err != nil {
				t.Fatalf("url.Parse: %v", err)
			}

			var openIDCount int

			for _, s := range strings.Fields(parsed.Query().Get("scope")) {
				if s == scopeOpenID {
					openIDCount++
				}
			}

			if openIDCount != 1 {
				t.Errorf("openid count in scope = %d, want 1 (scope=%q)", openIDCount, parsed.Query().Get("scope"))
			}
		})
	}
}

func TestExchangeVerifiesTheIDToken(t *testing.T) {
	t.Parallel()

	idp := newTestIdP(t)
	provider := testProvider(idp.srv.URL)

	rawIDToken := signClaims(t, idp.key, idp.keyID, baseClaims(idp.srv.URL, provider.ClientID, testNonce))
	idp.setTokenHandler(assertingTokenResponse(t, "test-code", "test-verifier", testRedirectURI, rawIDToken))

	c := newTestClient(t)

	ts, err := c.Exchange(context.Background(), provider, "test-code", "test-verifier", testNonce, testRedirectURI)
	if err != nil {
		t.Fatalf("Exchange: %v", err)
	}

	if ts.Subject != testSubject {
		t.Errorf("Subject = %q, want %q", ts.Subject, testSubject)
	}

	if ts.Claims["sub"] != testSubject {
		t.Errorf("Claims[sub] = %v, want %q", ts.Claims["sub"], testSubject)
	}
}

func TestExchangeRejectsANonceMismatch(t *testing.T) {
	t.Parallel()

	idp := newTestIdP(t)
	provider := testProvider(idp.srv.URL)

	rawIDToken := signClaims(t, idp.key, idp.keyID, baseClaims(idp.srv.URL, provider.ClientID, "correct-nonce"))
	idp.setTokenHandler(jsonTokenResponse(rawIDToken))

	c := newTestClient(t)

	_, err := c.Exchange(context.Background(), provider, "test-code", "test-verifier", "wrong-nonce", testRedirectURI)
	if !errors.Is(err, oidc.ErrNonceMismatch) {
		t.Errorf("Exchange = %v, want ErrNonceMismatch", err)
	}
}

func TestExchangeRejectsAnEmptyNonce(t *testing.T) {
	t.Parallel()

	idp := newTestIdP(t)
	provider := testProvider(idp.srv.URL)

	rawIDToken := signClaims(t, idp.key, idp.keyID, baseClaims(idp.srv.URL, provider.ClientID, ""))
	idp.setTokenHandler(jsonTokenResponse(rawIDToken))

	c := newTestClient(t)

	_, err := c.Exchange(context.Background(), provider, "test-code", "test-verifier", "", testRedirectURI)
	if !errors.Is(err, oidc.ErrNonceMismatch) {
		t.Errorf("Exchange = %v, want ErrNonceMismatch", err)
	}
}

func TestExchangeRejectsAnEmptyVerifier(t *testing.T) {
	t.Parallel()

	idp := newTestIdP(t)
	provider := testProvider(idp.srv.URL)
	c := newTestClient(t)

	_, err := c.Exchange(context.Background(), provider, "test-code", "", testNonce, testRedirectURI)
	if !errors.Is(err, oidc.ErrExchangeFailed) {
		t.Errorf("Exchange = %v, want ErrExchangeFailed", err)
	}
}

func TestExchangeRejectsAnExpiredIDToken(t *testing.T) {
	t.Parallel()

	idp := newTestIdP(t)
	provider := testProvider(idp.srv.URL)

	claims := baseClaims(idp.srv.URL, provider.ClientID, testNonce)
	claims["exp"] = time.Now().Add(-time.Hour).Unix()

	idp.setTokenHandler(jsonTokenResponse(signClaims(t, idp.key, idp.keyID, claims)))

	c := newTestClient(t)

	_, err := c.Exchange(context.Background(), provider, "test-code", "test-verifier", testNonce, testRedirectURI)
	if !errors.Is(err, oidc.ErrIDTokenInvalid) {
		t.Errorf("Exchange = %v, want ErrIDTokenInvalid", err)
	}
}

func TestExchangeRejectsABadSignature(t *testing.T) {
	t.Parallel()

	idp := newTestIdP(t)
	provider := testProvider(idp.srv.URL)

	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey: %v", err)
	}

	rawIDToken := signClaims(t, otherKey, idp.keyID, baseClaims(idp.srv.URL, provider.ClientID, testNonce))
	idp.setTokenHandler(jsonTokenResponse(rawIDToken))

	c := newTestClient(t)

	_, err = c.Exchange(context.Background(), provider, "test-code", "test-verifier", testNonce, testRedirectURI)
	if !errors.Is(err, oidc.ErrIDTokenInvalid) {
		t.Errorf("Exchange = %v, want ErrIDTokenInvalid", err)
	}
}

func TestExchangeSurfacesATokenEndpointError(t *testing.T) {
	t.Parallel()

	idp := newTestIdP(t)
	provider := testProvider(idp.srv.URL)
	idp.setTokenHandler(invalidGrantResponse())

	c := newTestClient(t)

	_, err := c.Exchange(context.Background(), provider, "bad-code", "test-verifier", testNonce, testRedirectURI)
	if !errors.Is(err, oidc.ErrExchangeFailed) {
		t.Errorf("Exchange = %v, want ErrExchangeFailed", err)
	}
}

func TestExchangeReadsTheSIDClaim(t *testing.T) {
	t.Parallel()

	idp := newTestIdP(t)
	provider := testProvider(idp.srv.URL)
	c := newTestClient(t)

	t.Run("with sid", func(t *testing.T) {
		claims := baseClaims(idp.srv.URL, provider.ClientID, testNonce)
		claims["sid"] = "session-abc"
		idp.setTokenHandler(jsonTokenResponse(signClaims(t, idp.key, idp.keyID, claims)))

		ts, err := c.Exchange(context.Background(), provider, "test-code", "test-verifier", testNonce, testRedirectURI)
		if err != nil {
			t.Fatalf("Exchange: %v", err)
		}

		if ts.SID != "session-abc" {
			t.Errorf("SID = %q, want %q", ts.SID, "session-abc")
		}
	})

	t.Run("without sid", func(t *testing.T) {
		claims := baseClaims(idp.srv.URL, provider.ClientID, testNonce)
		idp.setTokenHandler(jsonTokenResponse(signClaims(t, idp.key, idp.keyID, claims)))

		ts, err := c.Exchange(context.Background(), provider, "test-code", "test-verifier", testNonce, testRedirectURI)
		if err != nil {
			t.Fatalf("Exchange: %v", err)
		}

		if ts.SID != "" {
			t.Errorf("SID = %q, want empty", ts.SID)
		}
	})
}

func TestClientCachesTheProviderAcrossCalls(t *testing.T) {
	t.Parallel()

	idp := newTestIdP(t)
	provider := testProvider(idp.srv.URL)
	c := newTestClient(t)

	params := oidcproviders.AuthCodeParams{
		RedirectURI: testRedirectURI,
		State:       "state",
		Nonce:       testNonce,
		Verifier:    "verifier",
	}

	if _, err := c.AuthCodeURL(context.Background(), provider, params); err != nil {
		t.Fatalf("AuthCodeURL #1: %v", err)
	}

	if _, err := c.AuthCodeURL(context.Background(), provider, params); err != nil {
		t.Fatalf("AuthCodeURL #2: %v", err)
	}

	if got := atomic.LoadInt32(&idp.discoveryHits); got != 1 {
		t.Errorf("discoveryHits = %d, want 1", got)
	}
}

func TestEvictForcesARediscovery(t *testing.T) {
	t.Parallel()

	idp := newTestIdP(t)
	provider := testProvider(idp.srv.URL)
	c := newTestClient(t)

	params := oidcproviders.AuthCodeParams{
		RedirectURI: testRedirectURI,
		State:       "state",
		Nonce:       testNonce,
		Verifier:    "verifier",
	}

	if _, err := c.AuthCodeURL(context.Background(), provider, params); err != nil {
		t.Fatalf("AuthCodeURL #1: %v", err)
	}

	c.Evict(provider.ID)

	if _, err := c.AuthCodeURL(context.Background(), provider, params); err != nil {
		t.Fatalf("AuthCodeURL #2: %v", err)
	}

	if got := atomic.LoadInt32(&idp.discoveryHits); got != 2 {
		t.Errorf("discoveryHits = %d, want 2", got)
	}
}

func TestNewClientValidatesDeps(t *testing.T) {
	t.Parallel()

	if _, err := oidc.NewClient(oidc.ClientConfig{}, oidc.ClientDeps{Cipher: stubDecrypter{}}); err == nil {
		t.Error("NewClient without an HTTP client = nil, want an error")
	}

	if _, err := oidc.NewClient(oidc.ClientConfig{}, oidc.ClientDeps{HTTPClient: http.DefaultClient}); err == nil {
		t.Error("NewClient without a cipher = nil, want an error")
	}
}
