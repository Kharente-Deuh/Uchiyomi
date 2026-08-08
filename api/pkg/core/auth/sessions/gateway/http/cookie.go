// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type CookieConfig struct {
	Name   string
	Path   string
	Secure bool
}

func (cfg *CookieConfig) Validate() error {
	if cfg.Name == "" {
		return errors.New("name is required")
	}

	if cfg.Path == "" {
		return errors.New("path is required")
	}

	if !strings.HasPrefix(cfg.Path, "/") {
		return fmt.Errorf("path must start with '/', got %q", cfg.Path)
	}

	return nil
}

type CookieManager struct {
	cfg CookieConfig
}

func NewCookieManager(cfg CookieConfig) (*CookieManager, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("cfg.Validate: %w", err)
	}

	return &CookieManager{cfg: cfg}, nil
}

func (m *CookieManager) Set(w http.ResponseWriter, token string, expiresAt, now time.Time) {
	maxAge := int(expiresAt.Sub(now).Seconds())
	if maxAge < 1 {
		maxAge = 1
	}

	http.SetCookie(w, &http.Cookie{
		Name:     m.cfg.Name,
		Value:    token,
		Path:     m.cfg.Path,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   m.cfg.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (m *CookieManager) Read(r *http.Request) string {
	c, err := r.Cookie(m.cfg.Name)
	if err != nil {
		return ""
	}

	return c.Value
}

func (m *CookieManager) Clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     m.cfg.Name,
		Value:    "",
		Path:     m.cfg.Path,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   m.cfg.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}
