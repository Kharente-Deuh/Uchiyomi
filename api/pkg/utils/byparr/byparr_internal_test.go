// SPDX-License-Identifier: AGPL-3.0-or-later

package byparr

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const testBaseURL = "http://byparr:8191"

const testCookieValue = "abc"

const cfClearanceCookie = "cf_clearance"

func TestNewNormalisesBaseURL(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		testBaseURL:          testBaseURL,
		testBaseURL + "/":    testBaseURL,
		testBaseURL + "/v1":  testBaseURL,
		testBaseURL + "/v1/": testBaseURL,
		testBaseURL + "///":  testBaseURL,
	}

	for in, want := range tests {
		t.Run(in, func(t *testing.T) {
			t.Parallel()

			if got := New(in).baseURL; got != want {
				t.Errorf("New(%q).baseURL = %q, want %q", in, got, want)
			}
		})
	}
}

func TestNewDefaults(t *testing.T) {
	t.Parallel()

	c := New("http://x")

	if c.maxTimeout != DefaultMaxTimeout {
		t.Errorf("maxTimeout = %v, want %v", c.maxTimeout, DefaultMaxTimeout)
	}

	if c.httpc == nil {
		t.Error("httpc = nil, want un client par défaut")
	}

	if c.retries != 0 {
		t.Errorf("retries = %d, want 0", c.retries)
	}
}

func TestOptionsIgnoreInvalidValues(t *testing.T) {
	t.Parallel()

	c := New("http://x",
		WithMaxTimeout(0),
		WithMaxTimeout(-time.Second),
		WithRetries(-3),
		WithHTTPClient(nil),
	)

	if c.maxTimeout != DefaultMaxTimeout {
		t.Errorf("maxTimeout = %v, un timeout <= 0 doit être ignoré", c.maxTimeout)
	}

	if c.retries != 0 {
		t.Errorf("retries = %d, une valeur négative doit être ignorée", c.retries)
	}

	if c.httpc == nil {
		t.Error("httpc = nil, WithHTTPClient(nil) doit être ignoré")
	}
}

func TestCookieAsHTTP(t *testing.T) {
	t.Parallel()

	t.Run("path par défaut", func(t *testing.T) {
		t.Parallel()

		got := Cookie{Name: "a", Value: "b"}.AsHTTP()
		if got.Path != "/" {
			t.Errorf("Path = %q, want %q", got.Path, "/")
		}

		if !got.Expires.IsZero() {
			t.Errorf("Expires = %v, want zéro pour un cookie de session", got.Expires)
		}
	})

	t.Run("expiration positive", func(t *testing.T) {
		t.Parallel()

		got := Cookie{Name: "a", Value: "b", Expires: 1735689600, Path: "/x", HTTPOnly: true, Secure: true}.AsHTTP()

		if got.Expires.Unix() != 1735689600 {
			t.Errorf("Expires = %v, want %v", got.Expires.Unix(), 1735689600)
		}

		if got.Path != "/x" || !got.HttpOnly || !got.Secure {
			t.Errorf("cookie = %+v", got)
		}
	})

	t.Run("cookie de session", func(t *testing.T) {
		t.Parallel()

		got := Cookie{Name: "a", Value: "b", Expires: -1}.AsHTTP()
		if !got.Expires.IsZero() {
			t.Errorf("Expires = %v, want zéro quand Expires vaut -1", got.Expires)
		}
	})
}

func TestSolutionCookieHelpers(t *testing.T) {
	t.Parallel()

	sol := &Solution{Cookies: []Cookie{
		{Name: cfClearanceCookie, Value: testCookieValue},
		{Name: "session", Value: "xyz"},
	}}

	if got, want := sol.CookieHeader(), "cf_clearance=abc; session=xyz"; got != want {
		t.Errorf("CookieHeader() = %q, want %q", got, want)
	}

	m := sol.CookieMap()
	if m[cfClearanceCookie] != testCookieValue || m["session"] != "xyz" || len(m) != 2 {
		t.Errorf("CookieMap() = %v", m)
	}

	empty := &Solution{}
	if got := empty.CookieHeader(); got != "" {
		t.Errorf("CookieHeader() sans cookie = %q, want vide", got)
	}

	if got := empty.CookieMap(); len(got) != 0 {
		t.Errorf("CookieMap() sans cookie = %v, want vide", got)
	}
}

func TestSolutionApplyTo(t *testing.T) {
	t.Parallel()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar.New: %v", err)
	}

	sol := &Solution{Cookies: []Cookie{{Name: cfClearanceCookie, Value: testCookieValue}}}
	if err := sol.ApplyTo(jar, "https://asuracomic.net/series"); err != nil {
		t.Fatalf("ApplyTo: %v", err)
	}

	u, _ := url.Parse("https://asuracomic.net/anything")

	cookies := jar.Cookies(u)
	if len(cookies) != 1 || cookies[0].Name != cfClearanceCookie || cookies[0].Value != testCookieValue {
		t.Errorf("cookies dans le jar = %v", cookies)
	}
}

func TestSolutionApplyToInvalidURL(t *testing.T) {
	t.Parallel()

	jar, _ := cookiejar.New(nil)

	sol := &Solution{Cookies: []Cookie{{Name: "a", Value: "b"}}}
	if err := sol.ApplyTo(jar, "://pas-une-url"); err == nil {
		t.Error("ApplyTo sur une URL invalide = nil, want une erreur")
	}
}

func TestErrorMessage(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		err  *Error
		want string
	}{
		"message présent": {
			err:  &Error{HTTPStatus: 500, Status: "error", Message: "challenge failed"},
			want: `byparr: challenge failed (http 500, status "error")`,
		},
		"message absent": {
			err:  &Error{HTTPStatus: 502},
			want: `byparr: réponse inattendue (http 502, status "")`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := tc.err.Error(); got != tc.want {
				t.Errorf("Error() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestExtractMessage(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		body string
		want string
	}{
		"clé message":      {body: `{"message":"boom"}`, want: "boom"},
		"clé detail":       {body: `{"detail":"not found"}`, want: "not found"},
		"clé error":        {body: `{"error":"nope"}`, want: "nope"},
		"message prime":    {body: `{"detail":"d","message":"m"}`, want: "m"},
		"texte brut":       {body: "  502 Bad Gateway  ", want: "502 Bad Gateway"},
		"json sans clé":    {body: `{"other":1}`, want: `{"other":1}`},
		"corps vide":       {body: "", want: ""},
		"valeur numérique": {body: `{"message":42}`, want: "42"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := extractMessage([]byte(tc.body)); got != tc.want {
				t.Errorf("extractMessage(%q) = %q, want %q", tc.body, got, tc.want)
			}
		})
	}
}

func TestExtractMessageTruncatesLongBodies(t *testing.T) {
	t.Parallel()

	got := extractMessage([]byte(strings.Repeat("x", 400)))

	if !strings.HasSuffix(got, "…") {
		t.Errorf("un corps long doit être tronqué, longueur = %d", len(got))
	}

	if want := 300 + len("…"); len(got) != want {
		t.Errorf("len = %d, want %d", len(got), want)
	}
}

func TestHealth(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/health")
		}

		if accept := r.Header.Get("Accept"); accept != "application/json" {
			t.Errorf("Accept = %q, want %q", accept, "application/json")
		}

		_, _ = w.Write([]byte(`{"msg":"ok","version":"1.2.3","userAgent":"Mozilla/5.0"}`))
	}))
	defer srv.Close()

	h, err := New(srv.URL).Health(context.Background())
	if err != nil {
		t.Fatalf("Health: %v", err)
	}

	if h.Msg != "ok" || h.Version != "1.2.3" || h.UserAgent != "Mozilla/5.0" {
		t.Errorf("health = %+v", h)
	}
}

func TestHealthError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"detail":"browser down"}`))
	}))
	defer srv.Close()

	_, err := New(srv.URL).Health(context.Background())

	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("Health() = %v, want *byparr.Error", err)
	}

	if apiErr.HTTPStatus != http.StatusServiceUnavailable {
		t.Errorf("HTTPStatus = %d, want %d", apiErr.HTTPStatus, http.StatusServiceUnavailable)
	}

	if apiErr.Message != "browser down" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "browser down")
	}
}

func TestHealthInvalidJSON(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{`))
	}))
	defer srv.Close()

	if _, err := New(srv.URL).Health(context.Background()); err == nil {
		t.Error("Health sur un JSON invalide = nil, want une erreur")
	}
}

func TestDoRejectsEmptyURL(t *testing.T) {
	t.Parallel()

	if _, err := New("http://x").Do(context.Background(), Request{}); err == nil {
		t.Error("Do sans URL = nil, want une erreur")
	}
}

func TestDoPayload(t *testing.T) {
	t.Parallel()

	var got map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/v1")
		}

		if r.Method != http.MethodPost {
			t.Errorf("méthode = %q, want POST", r.Method)
		}

		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/json")
		}

		if v := r.Header.Get("X-Auth"); v != "secret" {
			t.Errorf("X-Auth = %q, want %q (en-tête client)", v, "secret")
		}

		if v := r.Header.Get("X-Req"); v != "per-request" {
			t.Errorf("X-Req = %q, want %q (en-tête de requête)", v, "per-request")
		}

		_ = json.NewDecoder(r.Body).Decode(&got)

		_, _ = w.Write([]byte(`{"status":"ok","solution":{"url":"https://target/","status":200}}`))
	}))
	defer srv.Close()

	c := New(srv.URL, WithHeader("X-Auth", "secret"), WithMaxTimeout(30*time.Second))

	_, err := c.Do(context.Background(), Request{
		Cmd:      CmdPost,
		URL:      "https://target/",
		PostData: "a=1",
		Proxy:    "http://proxy:3128",
		Headers:  map[string]string{"X-Req": "per-request"},
	})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}

	if got["cmd"] != CmdPost {
		t.Errorf("cmd = %v, want %q", got["cmd"], CmdPost)
	}

	if got["url"] != "https://target/" {
		t.Errorf("url = %v", got["url"])
	}

	if got["postData"] != "a=1" {
		t.Errorf("postData = %v", got["postData"])
	}

	if got["proxy"] != "http://proxy:3128" {
		t.Errorf("proxy = %v", got["proxy"])
	}

	if got["maxTimeout"] != float64(30_000) {
		t.Errorf("maxTimeout = %v, want 30000 (millisecondes)", got["maxTimeout"])
	}

	if _, ok := got["max_timeout"]; ok {
		t.Error("max_timeout ne doit pas être envoyé sans WithSecondsTimeout")
	}
}

func TestDoSecondsTimeout(t *testing.T) {
	t.Parallel()

	var got map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		_, _ = w.Write([]byte(`{"status":"ok","solution":{"url":"https://t/","status":200}}`))
	}))
	defer srv.Close()

	c := New(srv.URL, WithSecondsTimeout(), WithMaxTimeout(45*time.Second))

	if _, err := c.Get(context.Background(), "https://t/"); err != nil {
		t.Fatalf("Get: %v", err)
	}

	if got["max_timeout"] != float64(45) {
		t.Errorf("max_timeout = %v, want 45 (secondes)", got["max_timeout"])
	}

	if _, ok := got["maxTimeout"]; ok {
		t.Error("maxTimeout ne doit pas être envoyé avec WithSecondsTimeout")
	}
}

func TestGetDefaultsToCmdGet(t *testing.T) {
	t.Parallel()

	var got map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		_, _ = w.Write([]byte(`{"status":"ok","solution":{"url":"https://t/","status":200,"response":"<html/>"}}`))
	}))
	defer srv.Close()

	sol, err := New(srv.URL).Get(context.Background(), "https://t/")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if got["cmd"] != CmdGet {
		t.Errorf("cmd = %v, want %q", got["cmd"], CmdGet)
	}

	if sol.Response != "<html/>" {
		t.Errorf("Response = %q", sol.Response)
	}
}

func TestPostEncodesForm(t *testing.T) {
	t.Parallel()

	var got map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		_, _ = w.Write([]byte(`{"status":"ok","solution":{"url":"https://t/","status":200}}`))
	}))
	defer srv.Close()

	form := url.Values{"user": {"bob"}, "pass": {"s3cr3t"}}

	if _, err := New(srv.URL).Post(context.Background(), "https://t/", form); err != nil {
		t.Fatalf("Post: %v", err)
	}

	if got["cmd"] != CmdPost {
		t.Errorf("cmd = %v, want %q", got["cmd"], CmdPost)
	}

	if got["postData"] != form.Encode() {
		t.Errorf("postData = %v, want %q", got["postData"], form.Encode())
	}
}

func TestDoStatusNotOK(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":"error","message":"challenge non résolu"}`))
	}))
	defer srv.Close()

	_, err := New(srv.URL).Get(context.Background(), "https://t/")

	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("Get() = %v, want *byparr.Error", err)
	}

	if apiErr.Status != "error" || apiErr.Message != "challenge non résolu" {
		t.Errorf("err = %+v", apiErr)
	}

	if apiErr.HTTPStatus != http.StatusOK {
		t.Errorf("HTTPStatus = %d, want 200 (Byparr répond 200 avec status error)", apiErr.HTTPStatus)
	}
}

func TestDoDoesNotRetryOn4xx(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"detail":"cmd inconnu"}`))
	}))
	defer srv.Close()

	_, err := New(srv.URL, WithRetries(3)).Get(context.Background(), "https://t/")
	if err == nil {
		t.Fatal("Get sur un 400 = nil, want une erreur")
	}

	if got := calls.Load(); got != 1 {
		t.Errorf("%d appels pour un 400, want 1 (pas de retry sur une erreur cliente)", got)
	}
}

func TestDoRetriesOn5xx(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`{"message":"upstream"}`))

			return
		}

		_, _ = w.Write([]byte(`{"status":"ok","solution":{"url":"https://t/","status":200}}`))
	}))
	defer srv.Close()

	sol, err := New(srv.URL, WithRetries(1)).Get(context.Background(), "https://t/")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if sol == nil {
		t.Fatal("solution nil après un retry réussi")
	}

	if got := calls.Load(); got != 2 {
		t.Errorf("%d appels, want 2 (un échec puis une réussite)", got)
	}
}

func TestDoStopsOnCancelledContext(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := New(srv.URL, WithRetries(5)).Get(ctx, "https://t/"); err == nil {
		t.Error("Get sur un contexte annulé = nil, want une erreur")
	}
}

func TestSolutionHTTPClientRequiresURL(t *testing.T) {
	t.Parallel()

	if _, err := (&Solution{}).HTTPClient(nil); err == nil {
		t.Error("HTTPClient sans URL = nil, want une erreur")
	}
}

func TestSessionCarriesCookiesAndUserAgent(t *testing.T) {
	t.Parallel()

	var (
		gotUA     string
		gotCookie string
	)

	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")

		if ck, err := r.Cookie(cfClearanceCookie); err == nil {
			gotCookie = ck.Value
		}

		_, _ = w.Write([]byte("ok"))
	}))
	defer target.Close()

	byparrSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":"ok","solution":{
			"url":"` + target.URL + `/",
			"status":200,
			"userAgent":"Mozilla/5.0 (test)",
			"cookies":[{"name":"cf_clearance","value":"abc123"}]
		}}`))
	}))
	defer byparrSrv.Close()

	hc, sol, err := New(byparrSrv.URL).Session(context.Background(), target.URL+"/")
	if err != nil {
		t.Fatalf("Session: %v", err)
	}

	if sol.UserAgent != "Mozilla/5.0 (test)" {
		t.Errorf("UserAgent = %q", sol.UserAgent)
	}

	resp, err := hc.Get(target.URL + "/api")
	if err != nil {
		t.Fatalf("requête directe: %v", err)
	}

	defer resp.Body.Close()

	if gotUA != "Mozilla/5.0 (test)" {
		t.Errorf("User-Agent envoyé = %q, want celui de la solution", gotUA)
	}

	if gotCookie != "abc123" {
		t.Errorf("cookie cf_clearance envoyé = %q, want %q", gotCookie, "abc123")
	}
}

func TestSessionOverrideRequestKeepsWarmupURL(t *testing.T) {
	t.Parallel()

	var got map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		_, _ = w.Write([]byte(`{"status":"ok","solution":{"url":"https://t/","status":200}}`))
	}))
	defer srv.Close()

	_, _, err := New(srv.URL).Session(context.Background(), "https://t/", Request{Cmd: CmdGet, Proxy: "http://p:1"})
	if err != nil {
		t.Fatalf("Session: %v", err)
	}

	if got["url"] != "https://t/" {
		t.Errorf("url = %v, want l'URL de warmup quand la Request n'en porte pas", got["url"])
	}

	if got["proxy"] != "http://p:1" {
		t.Errorf("proxy = %v, l'option de la Request doit être conservée", got["proxy"])
	}
}

func TestUATransportForcesUserAgent(t *testing.T) {
	t.Parallel()

	var got string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("User-Agent")
	}))
	defer srv.Close()

	hc := &http.Client{Transport: &uaTransport{base: http.DefaultTransport, ua: "forced-ua"}}

	req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
	req.Header.Set("User-Agent", "original")

	resp, err := hc.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}

	defer resp.Body.Close()

	if got != "forced-ua" {
		t.Errorf("User-Agent = %q, want %q", got, "forced-ua")
	}

	if req.Header.Get("User-Agent") != "original" {
		t.Errorf("la requête appelante a été mutée: %q", req.Header.Get("User-Agent"))
	}
}

func TestTransportRoundTrip(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":"ok","solution":{
			"url":"https://t/","status":201,"response":"<html>hi</html>",
			"cookies":[{"name":"a","value":"b"}]
		}}`))
	}))
	defer srv.Close()

	hc := &http.Client{Transport: &Transport{Client: New(srv.URL)}}

	resp, err := hc.Get("https://t/page")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusCreated)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "<html>hi</html>" {
		t.Errorf("body = %q", body)
	}

	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q", ct)
	}

	if len(resp.Header.Values("Set-Cookie")) != 1 {
		t.Errorf("Set-Cookie = %v", resp.Header.Values("Set-Cookie"))
	}
}

func TestTransportDefaultsStatusTo200(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":"ok","solution":{"url":"https://t/","response":"x"}}`))
	}))
	defer srv.Close()

	tr := &Transport{Client: New(srv.URL)}

	req, _ := http.NewRequest(http.MethodGet, "https://t/", nil)

	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want 200 quand la solution n'en porte pas", resp.StatusCode)
	}
}

func TestTransportErrors(t *testing.T) {
	t.Parallel()

	t.Run("client nil", func(t *testing.T) {
		t.Parallel()

		req, _ := http.NewRequest(http.MethodGet, "https://t/", nil)
		if _, err := (&Transport{}).RoundTrip(req); err == nil {
			t.Error("RoundTrip sans Client = nil, want une erreur")
		}
	})

	t.Run("méthode non supportée", func(t *testing.T) {
		t.Parallel()

		tr := &Transport{Client: New("http://x")}

		req, _ := http.NewRequest(http.MethodDelete, "https://t/", nil)
		if _, err := tr.RoundTrip(req); err == nil {
			t.Error("RoundTrip en DELETE = nil, want une erreur")
		}
	})
}

func TestTransportForwardsPostBody(t *testing.T) {
	t.Parallel()

	var got map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		_, _ = w.Write([]byte(`{"status":"ok","solution":{"url":"https://t/","status":200}}`))
	}))
	defer srv.Close()

	tr := &Transport{Client: New(srv.URL)}

	req, _ := http.NewRequest(http.MethodPost, "https://t/", strings.NewReader("a=1&b=2"))

	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}

	defer resp.Body.Close()

	if got["cmd"] != CmdPost {
		t.Errorf("cmd = %v, want %q", got["cmd"], CmdPost)
	}

	if got["postData"] != "a=1&b=2" {
		t.Errorf("postData = %v, want %q", got["postData"], "a=1&b=2")
	}
}
