// SPDX-License-Identifier: AGPL-3.0-or-later

package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sessionshttp "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions/gateway/http"
)

const (
	cookieName = "uchiyomi_session"
	testToken  = "letoken"
	testUser   = "alice"
)

func newCookies(t *testing.T, secure bool) *sessionshttp.CookieManager {
	t.Helper()

	m, err := sessionshttp.NewCookieManager(sessionshttp.CookieConfig{Name: cookieName, Path: "/", Secure: secure})
	if err != nil {
		t.Fatalf("NewCookieManager: %v", err)
	}

	return m
}

func TestNewCookieManagerValidatesConfig(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		wantErr string
		cfg     sessionshttp.CookieConfig
	}{
		"empty name": {cfg: sessionshttp.CookieConfig{Path: "/"}, wantErr: "cfg.Validate: name is required"},
		"empty path": {cfg: sessionshttp.CookieConfig{Name: "s"}, wantErr: "cfg.Validate: path is required"},
		"path without slash": {
			cfg:     sessionshttp.CookieConfig{Name: "s", Path: "x"},
			wantErr: `cfg.Validate: path must start with '/', got "x"`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			m, err := sessionshttp.NewCookieManager(tc.cfg)
			if err == nil {
				t.Fatalf("NewCookieManager() = nil, want %q", tc.wantErr)
			}

			if m != nil {
				t.Error("NewCookieManager returned a manager in addition to the error")
			}

			if err.Error() != tc.wantErr {
				t.Errorf("NewCookieManager() = %q, want %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestSetProducesHardenedCookie(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	now := time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)

	newCookies(t, true).Set(rec, testToken, now.Add(2*time.Hour), now)

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("%d cookies set, want 1", len(cookies))
	}

	got := cookies[0]

	if got.Name != cookieName {
		t.Errorf("Name = %q, want %q", got.Name, cookieName)
	}

	if got.Value != testToken {
		t.Errorf("Value = %q, want %q", got.Value, testToken)
	}

	if !got.HttpOnly {
		t.Error("HttpOnly missing: XSS could read the token")
	}

	if !got.Secure {
		t.Error("Secure missing although config requires it")
	}

	if got.SameSite != http.SameSiteLaxMode {
		t.Errorf("SameSite = %v, want Lax", got.SameSite)
	}

	if got.Path != "/" {
		t.Errorf("Path = %q, want %q", got.Path, "/")
	}

	if got.MaxAge != 7200 {
		t.Errorf("MaxAge = %d, want 7200", got.MaxAge)
	}
}

func TestSetAllowsInsecureForLocalDevelopment(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	now := time.Now()

	newCookies(t, false).Set(rec, testToken, now.Add(time.Hour), now)

	got := rec.Result().Cookies()[0]

	if got.Secure {
		t.Error("Secure set although config does not require it")
	}

	if !got.HttpOnly {
		t.Error("HttpOnly lost when Secure is false: it must not follow config")
	}
}

func TestSetClampsExpiredDeadlineToOneSecond(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	now := time.Now()

	newCookies(t, true).Set(rec, testToken, now.Add(-time.Hour), now)

	if got := rec.Result().Cookies()[0].MaxAge; got != 1 {
		t.Errorf("MaxAge = %d, want 1", got)
	}
}

func TestReadReturnsTokenOrEmpty(t *testing.T) {
	t.Parallel()

	m := newCookies(t, true)

	withCookie := httptest.NewRequest(http.MethodGet, "/", nil)
	withCookie.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})

	if got := m.Read(withCookie); got != testToken {
		t.Errorf("Read() = %q, want %q", got, testToken)
	}

	if got := m.Read(httptest.NewRequest(http.MethodGet, "/", nil)); got != "" {
		t.Errorf("Read() = %q, want \"\" without cookie", got)
	}

	wrongName := httptest.NewRequest(http.MethodGet, "/", nil)
	wrongName.AddCookie(&http.Cookie{Name: "autre", Value: testToken})

	if got := m.Read(wrongName); got != "" {
		t.Errorf("Read() = %q, want \"\" for another cookie", got)
	}
}

func TestClearExpiresTheCookie(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()

	newCookies(t, true).Clear(rec)

	got := rec.Result().Cookies()[0]

	if got.Name != cookieName {
		t.Errorf("Name = %q, want %q", got.Name, cookieName)
	}

	if got.Value != "" {
		t.Errorf("Value = %q, want empty", got.Value)
	}

	if got.MaxAge != -1 {
		t.Errorf("MaxAge = %d, want -1", got.MaxAge)
	}
}
