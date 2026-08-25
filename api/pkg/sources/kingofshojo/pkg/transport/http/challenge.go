// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/domain"
)

var challengeMarkers = []string{
	"cf-browser-verification",
	"Just a moment",
	"challenge-platform",
}

func isChallenge(status int, body []byte) bool {
	if status == http.StatusForbidden || status == http.StatusServiceUnavailable {
		return true
	}

	text := strings.ToLower(string(body))
	for _, marker := range challengeMarkers {
		if strings.Contains(text, strings.ToLower(marker)) {
			return true
		}
	}

	return false
}

func (c *Client) get(ctx context.Context, targetURL string) (int, []byte, error) {
	status, body, err := c.doGet(ctx, c.httpClient(), targetURL)
	if err != nil {
		return 0, nil, fmt.Errorf("c.doGet: %w", err)
	}

	if !isChallenge(status, body) {
		return status, body, nil
	}

	if c.deps.Solver == nil {
		return 0, nil, domain.ErrChallenge
	}

	retryClient, _, err := c.deps.Solver.Session(ctx, targetURL)
	if err != nil {
		return 0, nil, fmt.Errorf("c.deps.Solver.Session: %w", err)
	}

	c.setRetryClient(retryClient)

	status, body, err = c.doGet(ctx, c.httpClient(), targetURL)
	if err != nil {
		return 0, nil, fmt.Errorf("c.doGet: %w", err)
	}

	if isChallenge(status, body) {
		return 0, nil, domain.ErrChallenge
	}

	return status, body, nil
}

func (c *Client) doGet(ctx context.Context, client *http.Client, targetURL string) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return 0, nil, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}

	res, err := client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("client.Do: %w", err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, nil, fmt.Errorf("io.ReadAll: %w", err)
	}

	if isChallenge(res.StatusCode, body) {
		return res.StatusCode, body, nil
	}

	if res.StatusCode == http.StatusNotFound {
		return res.StatusCode, body, nil
	}

	if res.StatusCode != http.StatusOK {
		return 0, nil, fmt.Errorf("unexpected status %d: %s", res.StatusCode, bytes.TrimSpace(body))
	}

	return res.StatusCode, body, nil
}
