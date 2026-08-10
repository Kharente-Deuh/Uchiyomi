// SPDX-License-Identifier: AGPL-3.0-or-later

package byparr

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

const (
	DefaultMaxTimeout = 60 * time.Second
	grace             = 20 * time.Second
)

const (
	CmdGet  = "request.get"
	CmdPost = "request.post"
)

const statusError = "error"

type Cookie struct {
	Name     string  `json:"name"`
	Value    string  `json:"value"`
	Domain   string  `json:"domain"`
	Path     string  `json:"path"`
	SameSite string  `json:"sameSite"`
	Expires  float64 `json:"expires"`
	HTTPOnly bool    `json:"httpOnly"`
	Secure   bool    `json:"secure"`
}

func (c Cookie) AsHTTP() *http.Cookie {
	hc := &http.Cookie{
		Name:     c.Name,
		Value:    c.Value,
		Domain:   c.Domain,
		Path:     c.Path,
		HttpOnly: c.HTTPOnly,
		Secure:   c.Secure,
	}
	if hc.Path == "" {
		hc.Path = "/"
	}
	if c.Expires > 0 {
		hc.Expires = time.Unix(int64(c.Expires), 0)
	}

	return hc
}

type Solution struct {
	Headers   map[string]any `json:"headers"`
	URL       string         `json:"url"`
	UserAgent string         `json:"userAgent"`
	Response  string         `json:"response"`
	Cookies   []Cookie       `json:"cookies"`
	Status    int            `json:"status"`
}

func (s *Solution) CookieHeader() string {
	parts := make([]string, 0, len(s.Cookies))
	for _, c := range s.Cookies {
		parts = append(parts, c.Name+"="+c.Value)
	}

	return strings.Join(parts, "; ")
}

func (s *Solution) CookieMap() map[string]string {
	m := make(map[string]string, len(s.Cookies))
	for _, c := range s.Cookies {
		m[c.Name] = c.Value
	}

	return m
}

func (s *Solution) ApplyTo(jar http.CookieJar, rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("byparr: invalid url %q: %w", rawURL, err)
	}
	cookies := make([]*http.Cookie, 0, len(s.Cookies))
	for _, c := range s.Cookies {
		cookies = append(cookies, c.AsHTTP())
	}
	jar.SetCookies(u, cookies)

	return nil
}

type Health struct {
	Msg       string `json:"msg"`
	Version   string `json:"version"`
	UserAgent string `json:"userAgent"`
}

type linkResponse struct {
	Status         string   `json:"status"`
	Message        string   `json:"message"`
	Version        string   `json:"version"`
	Solution       Solution `json:"solution"`
	StartTimestamp int64    `json:"startTimestamp"`
	EndTimestamp   int64    `json:"endTimestamp"`
}

type Error struct {
	Status     string
	Message    string
	HTTPStatus int
}

func (e *Error) Error() string {
	msg := e.Message
	if msg == "" {
		msg = "unexpected response"
	}

	return fmt.Sprintf("byparr: %s (http %d, status %q)", msg, e.HTTPStatus, e.Status)
}

type Request struct {
	Headers    map[string]string
	Cmd        string
	URL        string
	PostData   string
	Proxy      string
	MaxTimeout time.Duration
}

type Client struct {
	httpc            *http.Client
	headers          map[string]string
	baseURL          string
	maxTimeout       time.Duration
	retries          int
	timeoutInSeconds bool
}

type Option func(*Client)

func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		if hc != nil {
			c.httpc = hc
		}
	}
}

func WithMaxTimeout(d time.Duration) Option {
	return func(c *Client) {
		if d > 0 {
			c.maxTimeout = d
		}
	}
}

func WithHeader(k, v string) Option {
	return func(c *Client) {
		if c.headers == nil {
			c.headers = map[string]string{}
		}
		c.headers[k] = v
	}
}

func WithRetries(n int) Option {
	return func(c *Client) {
		if n >= 0 {
			c.retries = n
		}
	}
}

func WithSecondsTimeout() Option {
	return func(c *Client) { c.timeoutInSeconds = true }
}

func New(baseURL string, opts ...Option) *Client {
	c := &Client{
		baseURL:    strings.TrimRight(strings.TrimSuffix(strings.TrimRight(baseURL, "/"), "/v1"), "/"),
		httpc:      &http.Client{},
		maxTimeout: DefaultMaxTimeout,
	}
	for _, o := range opts {
		o(c)
	}

	return c
}

func (c *Client) Health(ctx context.Context) (*Health, error) {
	ctx, cancel := context.WithTimeout(ctx, c.maxTimeout+grace)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}
	c.applyHeaders(req, nil)

	resp, err := c.httpc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("byparr: appel /health: %w", err)
	}
	defer drain(resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("byparr: lecture /health: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &Error{HTTPStatus: resp.StatusCode, Message: extractMessage(body)}
	}
	var h Health
	if err := json.Unmarshal(body, &h); err != nil {
		return nil, fmt.Errorf("byparr: /health invalid json: %w", err)
	}

	return &h, nil
}

func (c *Client) Get(ctx context.Context, targetURL string) (*Solution, error) {
	return c.Do(ctx, Request{Cmd: CmdGet, URL: targetURL})
}

func (c *Client) Post(ctx context.Context, targetURL string, form url.Values) (*Solution, error) {
	return c.Do(ctx, Request{Cmd: CmdPost, URL: targetURL, PostData: form.Encode()})
}

func (c *Client) Do(ctx context.Context, r Request) (*Solution, error) {
	if r.URL == "" {
		return nil, errors.New("byparr: URL manquante")
	}
	if r.Cmd == "" {
		r.Cmd = CmdGet
	}
	timeout := r.MaxTimeout
	if timeout <= 0 {
		timeout = c.maxTimeout
	}

	payload := map[string]any{"cmd": r.Cmd, "url": r.URL}
	if c.timeoutInSeconds {
		payload["max_timeout"] = int(timeout.Seconds())
	} else {
		payload["maxTimeout"] = timeout.Milliseconds()
	}
	if r.PostData != "" {
		payload["postData"] = r.PostData
	}
	if r.Proxy != "" {
		payload["proxy"] = r.Proxy
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= c.retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("byparr: context canceled during backoff: %w", ctx.Err())
			case <-time.After(time.Duration(attempt) * 2 * time.Second):
			}
		}
		sol, err := c.solveOnce(ctx, body, timeout, r.Headers)
		if err == nil {
			return sol, nil
		}
		lastErr = err
		if ctx.Err() != nil {
			return nil, err
		}
		var apiErr *Error
		if errors.As(err, &apiErr) && apiErr.HTTPStatus >= 400 && apiErr.HTTPStatus < 500 {
			return nil, err
		}
	}

	return nil, lastErr
}

func (c *Client) solveOnce(
	ctx context.Context, body []byte, timeout time.Duration, extra map[string]string,
) (*Solution, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout+grace)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	c.applyHeaders(req, extra)

	resp, err := c.httpc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("byparr: appel /v1: %w", err)
	}
	defer drain(resp)

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("byparr: lecture /v1: %w", err)
	}

	var lr linkResponse
	jsonErr := json.Unmarshal(raw, &lr)

	if resp.StatusCode != http.StatusOK || !strings.EqualFold(lr.Status, "ok") {
		msg := lr.Message
		if msg == "" {
			msg = extractMessage(raw)
		}

		return nil, &Error{HTTPStatus: resp.StatusCode, Status: lr.Status, Message: msg}
	}
	if jsonErr != nil {
		return nil, fmt.Errorf("byparr: /v1 invalid json: %w", jsonErr)
	}
	sol := lr.Solution

	return &sol, nil
}

func (c *Client) applyHeaders(req *http.Request, extra map[string]string) {
	req.Header.Set("Accept", "application/json")
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	for k, v := range extra {
		req.Header.Set(k, v)
	}
}

func (c *Client) Session(ctx context.Context, warmupURL string, opts ...Request) (*http.Client, *Solution, error) {
	r := Request{Cmd: CmdGet, URL: warmupURL}
	if len(opts) > 0 {
		r = opts[0]
		if r.URL == "" {
			r.URL = warmupURL
		}
	}
	sol, err := c.Do(ctx, r)
	if err != nil {
		return nil, nil, err
	}
	hc, err := sol.HTTPClient(nil)
	if err != nil {
		return nil, nil, err
	}

	return hc, sol, nil
}

func (s *Solution) HTTPClient(base http.RoundTripper) (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("cookiejar.New: %w", err)
	}
	target := s.URL
	if target == "" {
		return nil, errors.New("byparr: solution without URL")
	}
	if err := s.ApplyTo(jar, target); err != nil {
		return nil, err
	}
	if base == nil {
		base = http.DefaultTransport
	}

	return &http.Client{
		Jar:       jar,
		Transport: &uaTransport{base: base, ua: s.UserAgent},
		Timeout:   30 * time.Second,
	}, nil
}

type uaTransport struct {
	base http.RoundTripper
	ua   string
}

func (t *uaTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.ua == "" || req.Header.Get("User-Agent") == t.ua {
		//nolint:wrapcheck // passthrough RoundTripper: callers expect the transport error unchanged.
		return t.base.RoundTrip(req)
	}
	clone := req.Clone(req.Context())
	clone.Header.Set("User-Agent", t.ua)

	//nolint:wrapcheck // passthrough RoundTripper: callers expect the transport error unchanged.
	return t.base.RoundTrip(clone)
}

type Transport struct {
	Client *Client
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Client == nil {
		return nil, errors.New("byparr: Transport.Client is nil")
	}
	r := Request{URL: req.URL.String()}
	switch req.Method {
	case http.MethodGet, "":
		r.Cmd = CmdGet
	case http.MethodPost:
		r.Cmd = CmdPost
		if req.Body != nil {
			b, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, fmt.Errorf("io.ReadAll: %w", err)
			}
			_ = req.Body.Close()
			r.PostData = string(b)
		}
	default:
		return nil, fmt.Errorf("byparr: method %s not supported", req.Method)
	}

	sol, err := t.Client.Do(req.Context(), r)
	if err != nil {
		return nil, err
	}

	resp := &http.Response{
		StatusCode:    sol.Status,
		Status:        fmt.Sprintf("%d %s", sol.Status, http.StatusText(sol.Status)),
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        http.Header{},
		Body:          io.NopCloser(strings.NewReader(sol.Response)),
		ContentLength: int64(len(sol.Response)),
		Request:       req,
	}
	if resp.StatusCode == 0 {
		resp.StatusCode = http.StatusOK
		resp.Status = "200 OK"
	}
	resp.Header.Set("Content-Type", "text/html; charset=utf-8")
	for _, ck := range sol.Cookies {
		if v := ck.AsHTTP().String(); v != "" {
			resp.Header.Add("Set-Cookie", v)
		}
	}

	return resp, nil
}

func extractMessage(body []byte) string {
	var m map[string]any
	if err := json.Unmarshal(body, &m); err == nil {
		for _, k := range []string{"message", "detail", statusError} {
			if v, ok := m[k]; ok {
				return fmt.Sprint(v)
			}
		}
	}
	s := strings.TrimSpace(string(body))
	if len(s) > 300 {
		s = s[:300] + "…"
	}

	return s
}

func drain(resp *http.Response) {
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	_ = resp.Body.Close()
}
