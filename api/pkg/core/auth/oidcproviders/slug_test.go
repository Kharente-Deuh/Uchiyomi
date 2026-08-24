// SPDX-License-Identifier: AGPL-3.0-or-later

package oidcproviders_test

import (
	"strings"
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidcproviders"
)

func TestIsValidSlug(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		slug string
		ok   bool
	}{
		{"keycloak", "keycloak", true},
		{"acme-sso", "acme-sso", true},
		{"a", "a", true},
		{"max-length", strings.Repeat("a", 64), true},
		{"empty", "", false},
		{"Keycloak", "Keycloak", false},
		{"-acme", "-acme", false},
		{"acme-", "acme-", false},
		{"acme--sso", "acme--sso", false},
		{"acme_sso", "acme_sso", false},
		{"over-max-length", strings.Repeat("a", 65), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := oidcproviders.IsValidSlug(tc.slug); got != tc.ok {
				t.Errorf("IsValidSlug(%q) = %v, want %v", tc.slug, got, tc.ok)
			}
		})
	}
}
