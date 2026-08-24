// SPDX-License-Identifier: AGPL-3.0-or-later

package oidcproviders

import "regexp"

const MaxSlugLength = 64

var SlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func IsValidSlug(slug string) bool {
	return len(slug) > 0 && len(slug) <= MaxSlugLength && SlugPattern.MatchString(slug)
}
