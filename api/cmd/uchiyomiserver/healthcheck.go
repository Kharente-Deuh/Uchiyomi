// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	healthcheckTimeout = 3 * time.Second
	defaultPort        = "3000"
)

func readyzURL() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	return fmt.Sprintf("http://localhost:%s/readyz", port)
}

func healthcheck(ctx context.Context, url string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("http.NewRequestWithContext: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http.DefaultClient.Do: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("readyz a répondu %d", resp.StatusCode)
	}

	return nil
}

func runHealthcheck() int {
	ctx, cancel := context.WithTimeout(context.Background(), healthcheckTimeout)
	defer cancel()

	if err := healthcheck(ctx, readyzURL()); err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck: %s\n", err)

		return 1
	}

	return 0
}
