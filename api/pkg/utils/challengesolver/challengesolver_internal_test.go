// SPDX-License-Identifier: AGPL-3.0-or-later

package challengesolver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

const testBaseURL = "http://challenge-solver:8191"

func TestNewTrimsTrailingSlashesAndV1(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"http://localhost:8191", "http://localhost:8191"},
		{"http://localhost:8191/", "http://localhost:8191"},
		{"http://localhost:8191/v1", "http://localhost:8191"},
		{"http://localhost:8191/v1/", "http://localhost:8191"},
		{"http://localhost:8191///", "http://localhost:8191"},
	}
	for _, tc := range cases {
		c := New(tc.input)
		if c.baseURL != tc.want {
			t.Errorf("New(%q).baseURL = %q, want %q", tc.input, c.baseURL, tc.want)
		}
	}
}

func TestOptions(t *testing.T) {
	hc := &http.Client{Timeout: 5 * time.Second}
	c := New("http://x",
		WithHTTPClient(hc),
		WithMaxTimeout(10*time.Second),
		WithHeader("X-Custom", "val"),
		WithRetries(3),
		WithSecondsTimeout(),
	)
	if c.httpc != hc {
		t.Error("WithHTTPClient not applied")
	}
	if c.maxTimeout != 10*time.Second {
		t.Errorf("maxTimeout = %v, want 10s", c.maxTimeout)
	}
	if c.headers["X-Custom"] != "val" {
		t.Errorf("header = %q, want val", c.headers["X-Custom"])
	}
	if c.retries != 3 {
		t.Errorf("retries = %d, want 3", c.retries)
	}
	if !c.timeoutInSeconds {
		t.Error("timeoutInSeconds should be true")
	}
}

func TestCookieAsHTTP(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	c := Cookie{
		Name:     "cf_clearance",
		Value:    "secret",
		Domain:   ".example.com",
		Path:     "",
		Expires:  float64(now.Unix()),
		HTTPOnly: true,
		Secure:   true,
	}
	hc := c.AsHTTP()
	if hc.Name != "cf_clearance" || hc.Value != "secret" {
		t.Errorf("unexpected name/val: %s=%s", hc.Name, hc.Value)
	}
	if hc.Path != "/" {
		t.Errorf("Path = %q, want /", hc.Path)
	}
	if !hc.HttpOnly || !hc.Secure {
		t.Error("flags HttpOnly/Secure not preserved")
	}
	if !hc.Expires.Equal(now) {
		t.Errorf("Expires = %v, want %v", hc.Expires, now)
	}
}

func TestSolutionHelpers(t *testing.T) {
	sol := Solution{
		URL:       "https://example.com/page",
		UserAgent: "CustomUA/1.0",
		Cookies: []Cookie{
			{Name: "a", Value: "1"},
			{Name: "b", Value: "2"},
		},
	}

	if got := sol.CookieHeader(); got != "a=1; b=2" {
		t.Errorf("CookieHeader() = %q, want %q", got, "a=1; b=2")
	}

	m := sol.CookieMap()
	if len(m) != 2 || m["a"] != "1" || m["b"] != "2" {
		t.Errorf("CookieMap() = %v", m)
	}

	hc, err := sol.HTTPClient(nil)
	if err != nil {
		t.Fatalf("HTTPClient: %v", err)
	}
	if hc.Jar == nil {
		t.Fatal("expected non-nil cookie jar")
	}
	u, _ := url.Parse("https://example.com/page")
	jarCookies := hc.Jar.Cookies(u)
	if len(jarCookies) != 2 {
		t.Errorf("got %d cookies in jar, want 2", len(jarCookies))
	}
}

func TestSolutionHTTPClientRequiresURL(t *testing.T) {
	sol := Solution{}
	if _, err := sol.HTTPClient(nil); err == nil {
		t.Error("expected error for empty URL")
	}
}

func TestErrorFormatting(t *testing.T) {
	cases := []struct {
		err  *Error
		want string
	}{
		{
			err:  &Error{HTTPStatus: 500, Status: "error", Message: "challenge failed"},
			want: `challengesolver: challenge failed (http 500, status "error")`,
		},
		{
			err:  &Error{HTTPStatus: 502, Status: "", Message: ""},
			want: `challengesolver: unexpected response (http 502, status "")`,
		},
	}
	for _, tc := range cases {
		if got := tc.err.Error(); got != tc.want {
			t.Errorf("Error() = %q, want %q", got, tc.want)
		}
	}
}

func TestHealthSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Errorf("path = %q, want /health", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"msg":"FlareSolverr is ready!","version":"v3.3.21","userAgent":"Mozilla/5.0"}`))
	}))
	defer srv.Close()

	h, err := New(srv.URL).Health(context.Background())
	if err != nil {
		t.Fatalf("Health() err: %v", err)
	}
	if h.Version != "v3.3.21" {
		t.Errorf("Version = %q, want v3.3.21", h.Version)
	}
	if h.Msg != "FlareSolverr is ready!" {
		t.Errorf("Msg = %q", h.Msg)
	}
}

func TestHealthErrors(t *testing.T) {
	t.Run("server error with json", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"error":"browser crashed"}`))
		}))
		defer srv.Close()

		_, err := New(srv.URL).Health(context.Background())
		if err == nil {
			t.Fatal("expected error")
		}
		var apiErr *Error
		if !errors.As(err, &apiErr) {
			t.Fatalf("Health() = %v, want *challengesolver.Error", err)
		}
		if apiErr.HTTPStatus != 503 {
			t.Errorf("HTTPStatus = %d, want 503", apiErr.HTTPStatus)
		}
		if apiErr.Message != "browser crashed" {
			t.Errorf("Message = %q, want 'browser crashed'", apiErr.Message)
		}
	})

	t.Run("invalid json response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`not json`))
		}))
		defer srv.Close()

		_, err := New(srv.URL).Health(context.Background())
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "invalid json") {
			t.Errorf("err = %v, want invalid json mention", err)
		}
	})
}

func TestDoSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1" {
			t.Errorf("path = %q, want /v1", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		var reqBody map[string]any
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if reqBody["cmd"] != CmdGet {
			t.Errorf("cmd = %v, want %s", reqBody["cmd"], CmdGet)
		}
		if reqBody["url"] != "https://target.example.com" {
			t.Errorf("url = %v", reqBody["url"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"status": "ok",
			"message": "Challenge solved!",
			"solution": {
				"url": "https://target.example.com",
				"status": 200,
				"headers": {"content-type": "text/html"},
				"response": "<html>ok</html>",
				"userAgent": "TestUA/1.0",
				"cookies": [{"name":"cf_clearance","value":"xyz","domain":".example.com","path":"/"}]
			}
		}`))
	}))
	defer srv.Close()

	sol, err := New(srv.URL).Get(context.Background(), "https://target.example.com")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if sol.Status != 200 {
		t.Errorf("Status = %d, want 200", sol.Status)
	}
	if sol.Response != "<html>ok</html>" {
		t.Errorf("Response = %q", sol.Response)
	}
	if len(sol.Cookies) != 1 || sol.Cookies[0].Name != "cf_clearance" {
		t.Errorf("Cookies = %v", sol.Cookies)
	}
}

func TestDoPayloadVariants(t *testing.T) {
	t.Run("milliseconds timeout (default)", func(t *testing.T) {
		var received map[string]any
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewDecoder(r.Body).Decode(&received)
			_, _ = w.Write([]byte(`{"status":"ok","solution":{"url":"https://x","status":200}}`))
		}))
		defer srv.Close()

		_, err := New(srv.URL, WithMaxTimeout(5*time.Second)).Get(context.Background(), "https://x")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if received["maxTimeout"] != float64(5000) {
			t.Errorf("maxTimeout = %v, want 5000", received["maxTimeout"])
		}
		if _, ok := received["max_timeout"]; ok {
			t.Error("max_timeout should not be present")
		}
	})

	t.Run("seconds timeout (WithSecondsTimeout)", func(t *testing.T) {
		var received map[string]any
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewDecoder(r.Body).Decode(&received)
			_, _ = w.Write([]byte(`{"status":"ok","solution":{"url":"https://x","status":200}}`))
		}))
		defer srv.Close()

		_, err := New(srv.URL, WithMaxTimeout(5*time.Second), WithSecondsTimeout()).Get(context.Background(), "https://x")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if received["max_timeout"] != float64(5) {
			t.Errorf("max_timeout = %v, want 5", received["max_timeout"])
		}
	})

	t.Run("post request with data", func(t *testing.T) {
		var received map[string]any
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewDecoder(r.Body).Decode(&received)
			_, _ = w.Write([]byte(`{"status":"ok","solution":{"url":"https://x","status":200}}`))
		}))
		defer srv.Close()

		form := url.Values{"username": {"alice"}}
		_, err := New(srv.URL).Post(context.Background(), "https://x", form)
		if err != nil {
			t.Fatalf("Post: %v", err)
		}
		if received["cmd"] != CmdPost {
			t.Errorf("cmd = %v, want %s", received["cmd"], CmdPost)
		}
		if received["postData"] != "username=alice" {
			t.Errorf("postData = %v", received["postData"])
		}
	})

	t.Run("proxy and custom timeout in request", func(t *testing.T) {
		var received map[string]any
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewDecoder(r.Body).Decode(&received)
			_, _ = w.Write([]byte(`{"status":"ok","solution":{"url":"https://x","status":200}}`))
		}))
		defer srv.Close()

		req := Request{
			URL:        "https://x",
			Proxy:      "socks5://127.0.0.1:1080",
			MaxTimeout: 12 * time.Second,
		}
		_, err := New(srv.URL).Do(context.Background(), req)
		if err != nil {
			t.Fatalf("Do: %v", err)
		}
		if received["proxy"] != "socks5://127.0.0.1:1080" {
			t.Errorf("proxy = %v", received["proxy"])
		}
		if received["maxTimeout"] != float64(12000) {
			t.Errorf("maxTimeout = %v, want 12000", received["maxTimeout"])
		}
	})
}

func TestDoRequiresURL(t *testing.T) {
	_, err := New("http://x").Do(context.Background(), Request{})
	if err == nil {
		t.Error("expected error for missing URL")
	}
}

func TestDoErrors(t *testing.T) {
	t.Run("status error in json body", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"error","message":"challenge not solved"}`))
		}))
		defer srv.Close()

		_, err := New(srv.URL).Get(context.Background(), "https://x")
		if err == nil {
			t.Fatal("expected error")
		}
		var apiErr *Error
		if !errors.As(err, &apiErr) {
			t.Fatalf("Get() = %v, want *challengesolver.Error", err)
		}
		if apiErr.Status != "error" || apiErr.Message != "challenge not solved" {
			t.Errorf("unexpected error fields: %+v", apiErr)
		}
		if apiErr.HTTPStatus != 200 {
			t.Errorf("HTTPStatus = %d, want 200 (Challenge solver responds 200 with status error)", apiErr.HTTPStatus)
		}
	})

	t.Run("non-200 HTTP status with html error page", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`<html>502 Bad Gateway</html>`))
		}))
		defer srv.Close()

		_, err := New(srv.URL).Get(context.Background(), "https://x")
		if err == nil {
			t.Fatal("expected error")
		}
		var apiErr *Error
		if !errors.As(err, &apiErr) {
			t.Fatalf("Get() = %v, want *challengesolver.Error", err)
		}
		if apiErr.HTTPStatus != 502 {
			t.Errorf("HTTPStatus = %d, want 502", apiErr.HTTPStatus)
		}
	})
}

func TestDoRetriesOn5xx(t *testing.T) {
	var attempts int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"error":"temporarily unavailable"}`))
			return
		}
		_, _ = w.Write([]byte(`{"status":"ok","solution":{"url":"https://x","status":200,"response":"done"}}`))
	}))
	defer srv.Close()

	sol, err := New(srv.URL, WithRetries(3)).Get(context.Background(), "https://x")
	if err != nil {
		t.Fatalf("Get with retries: %v", err)
	}
	if sol.Response != "done" {
		t.Errorf("Response = %q, want done", sol.Response)
	}
	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}

func TestDoDoesNotRetryOn4xx(t *testing.T) {
	var attempts int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"invalid cmd"}`))
	}))
	defer srv.Close()

	_, err := New(srv.URL, WithRetries(3)).Get(context.Background(), "https://x")
	if err == nil {
		t.Fatal("expected error")
	}
	if attempts != 1 {
		t.Errorf("attempts = %d, want 1 (should not retry 4xx)", attempts)
	}
}

func TestDoContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(500 * time.Millisecond)
		_, _ = w.Write([]byte(`{"status":"ok","solution":{"url":"https://x","status":200}}`))
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := New(srv.URL).Get(ctx, "https://x")
	if err == nil {
		t.Fatal("expected error on cancelled context")
	}
}

func TestSessionWarmupFlow(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")
		cookie, err := r.Cookie("cf_clearance")
		if err != nil || cookie.Value != "valid-clearance" || ua != "Camoufox/1.0" {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte("blocked"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("protected content"))
	}))
	defer target.Close()

	targetURL, _ := url.Parse(target.URL)

	solverSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(fmt.Appendf(nil, `{
			"status": "ok",
			"solution": {
				"url": %q,
				"status": 200,
				"userAgent": "Camoufox/1.0",
				"cookies": [{"name":"cf_clearance","value":"valid-clearance","domain":%q,"path":"/"}]
			}
		}`, target.URL+"/", targetURL.Hostname()))
	}))
	defer solverSrv.Close()

	hc, sol, err := New(solverSrv.URL).Session(context.Background(), target.URL+"/")
	if err != nil {
		t.Fatalf("Session: %v", err)
	}
	if sol.UserAgent != "Camoufox/1.0" {
		t.Errorf("UserAgent = %q", sol.UserAgent)
	}

	resp, err := hc.Get(target.URL + "/data")
	if err != nil {
		t.Fatalf("hc.Get: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "protected content" {
		t.Errorf("body = %q, want 'protected content'", string(body))
	}
}

func TestUATransportForcesUserAgent(t *testing.T) {
	var seenUA string
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	hc := &http.Client{Transport: &uaTransport{base: http.DefaultTransport, ua: "forced-ua"}}
	req, _ := http.NewRequest(http.MethodGet, backend.URL, nil)
	req.Header.Set("User-Agent", "original-ua")
	resp, err := hc.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	_ = resp.Body.Close()

	if seenUA != "forced-ua" {
		t.Errorf("seenUA = %q, want forced-ua", seenUA)
	}
}

func TestTransportRoundTrip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"status": "ok",
			"solution": {
				"url": "https://example.com/api",
				"status": 200,
				"headers": {"x-custom-res": "value1"},
				"response": "{\"data\":\"ok\"}",
				"userAgent": "SolverUA/1.0",
				"cookies": [{"name":"session","value":"abc","domain":".example.com","path":"/"}]
			}
		}`))
	}))
	defer srv.Close()

	hc := &http.Client{Transport: &Transport{Client: New(srv.URL)}}
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/api", nil)
	resp, err := hc.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
	if resp.Header.Get("x-custom-res") != "value1" {
		t.Errorf("header x-custom-res = %q", resp.Header.Get("x-custom-res"))
	}
	if !strings.Contains(resp.Header.Get("Set-Cookie"), "session=abc") {
		t.Errorf("Set-Cookie = %q", resp.Header.Get("Set-Cookie"))
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != `{"data":"ok"}` {
		t.Errorf("body = %q", string(body))
	}
}

func TestTransportDefaultsStatusTo200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
			"status": "ok",
			"solution": {
				"url": "https://example.com",
				"response": "ok"
			}
		}`))
	}))
	defer srv.Close()

	tr := &Transport{Client: New(srv.URL)}
	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
}

func TestTransportErrors(t *testing.T) {
	t.Run("nil client", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
		if _, err := (&Transport{}).RoundTrip(req); err == nil {
			t.Error("expected error for nil client")
		}
	})

	t.Run("unsupported method", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
		defer srv.Close()

		tr := &Transport{Client: New(srv.URL)}
		req, _ := http.NewRequest(http.MethodDelete, "https://example.com", nil)
		if _, err := tr.RoundTrip(req); err == nil {
			t.Error("expected error for DELETE method")
		}
	})
}

func TestTransportForwardsPostBody(t *testing.T) {
	var receivedPostData string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]any
		_ = json.NewDecoder(r.Body).Decode(&reqBody)
		if pd, ok := reqBody["postData"].(string); ok {
			receivedPostData = pd
		}
		_, _ = w.Write([]byte(`{"status":"ok","solution":{"url":"https://x","status":200,"response":"done"}}`))
	}))
	defer srv.Close()

	tr := &Transport{Client: New(srv.URL)}
	req, _ := http.NewRequest(http.MethodPost, "https://x", strings.NewReader("payload=123"))
	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	defer resp.Body.Close()

	if receivedPostData != "payload=123" {
		t.Errorf("postData = %q, want payload=123", receivedPostData)
	}
}
