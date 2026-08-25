// SPDX-License-Identifier: AGPL-3.0-or-later

package challengesolver

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
		return fmt.Errorf("challengesolver: invalid url %q: %w", rawURL, err)
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

	return fmt.Sprintf("challengesolver: %s (http %d, status %q)", msg, e.HTTPStatus, e.Status)
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
		return nil, fmt.Errorf("challengesolver: appel /health: %w", err)
	}
	defer drain(resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("challengesolver: lecture /health: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &Error{HTTPStatus: resp.StatusCode, Message: extractMessage(body)}
	}
	var h Health
	if err := json.Unmarshal(body, &h); err != nil {
		return nil, fmt.Errorf("challengesolver: /health invalid json: %w", err)
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
		return nil, errors.New("challengesolver: URL manquante")
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
				return nil, fmt.Errorf("challengesolver: context canceled during backoff: %w", ctx.Err())
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
		return nil, fmt.Errorf("challengesolver: appel /v1: %w", err)
	}
	defer drain(resp)

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("challengesolver: lecture /v1: %w", err)
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
		return nil, fmt.Errorf("challengesolver: /v1 invalid json: %w", jsonErr)
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
		return nil, errors.New("challengesolver: solution without URL")
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
		return nil, errors.New("challengesolver: Transport.Client is nil")
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
		return nil, fmt.Errorf("challengesolver: method %s not supported", req.Method)
	}

	sol, err := t.Client.Do(req.Context(), r)
	if err != nil {
		return nil, err
	}

	status := sol.Status
	if status == 0 {
		status = http.StatusOK
	}

	header := make(http.Header)
	for k, v := range sol.Headers {
		switch val := v.(type) {
		case string:
			header.Set(k, val)
		case []any:
			for _, item := range val {
				header.Add(k, fmt.Sprint(item))
			}
		default:
			header.Set(k, fmt.Sprint(v))
		}
	}
	if cookieHeader := sol.CookieHeader(); cookieHeader != "" && header.Get("Set-Cookie") == "" {
		header.Set("Set-Cookie", cookieHeader)
	}

	return &http.Response{
		Status:        http.StatusText(status),
		StatusCode:    status,
		Header:        header,
		Body:          io.NopCloser(strings.NewReader(sol.Response)),
		ContentLength: int64(len(sol.Response)),
		Request:       req,
	}, nil
}

func drain(resp *http.Response) {
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}

func extractMessage(b []byte) string {
	var m struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}
	if json.Unmarshal(b, &m) == nil {
		if m.Message != "" {
			return m.Message
		}
		if m.Error != "" {
			return m.Error
		}
	}

	return strings.TrimSpace(string(b))
}
