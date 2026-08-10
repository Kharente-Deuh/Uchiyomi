// SPDX-License-Identifier: AGPL-3.0-or-later

package auth

import "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidcproviders"

func (s *Service) mapClaims(
	provider oidcproviders.OIDCProvider,
	tokenSet *oidcproviders.TokenSet,
) (username string, isAdmin, allowed bool) {
	username, ok := usernameClaim(tokenSet.Claims, provider.UsernameClaim)
	if !ok {
		return "", false, false
	}

	if provider.RoleClaim == nil {
		return username, false, true
	}

	values := claimValues(tokenSet.Claims, *provider.RoleClaim)

	allowed = len(provider.AllowedValues) == 0 || intersects(values, provider.AllowedValues)
	isAdmin = mapsAdminRole(provider) && intersects(values, provider.AdminValues)

	return username, isAdmin, allowed
}

func mapsAdminRole(provider oidcproviders.OIDCProvider) bool {
	return provider.RoleClaim != nil && len(provider.AdminValues) > 0
}

func usernameClaim(claims map[string]any, key string) (string, bool) {
	raw, ok := claims[key]
	if !ok {
		return "", false
	}

	username, ok := raw.(string)
	if !ok || username == "" {
		return "", false
	}

	return username, true
}

func claimValues(claims map[string]any, key string) []string {
	switch v := claims[key].(type) {
	case string:
		return []string{v}
	case []string:
		return v
	case []any:
		values := make([]string, 0, len(v))

		for _, item := range v {
			if s, ok := item.(string); ok {
				values = append(values, s)
			}
		}

		return values
	default:
		return nil
	}
}

func intersects(a, b []string) bool {
	set := make(map[string]struct{}, len(b))
	for _, v := range b {
		set[v] = struct{}{}
	}

	for _, v := range a {
		if _, ok := set[v]; ok {
			return true
		}
	}

	return false
}
