// SPDX-License-Identifier: AGPL-3.0-or-later

package auth

import "testing"

func TestSafeRedirectPath(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		in   string
		want string
	}{
		"vide":                             {in: "", want: "/"},
		"pas de slash initial":             {in: "foo", want: "/"},
		"chemin local valide":              {in: "/ok", want: "/ok"},
		"double slash protocol-relative":   {in: "//x", want: "/"},
		"antislash":                        {in: "/a\\b", want: "/"},
		"URL avec schéma et hôte":          {in: "https://x", want: "/"},
		"tabulation avant un double slash": {in: "/\t/evil.example.com", want: "/"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := safeRedirectPath(tc.in); got != tc.want {
				t.Errorf("safeRedirectPath(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
