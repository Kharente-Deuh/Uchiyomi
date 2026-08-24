// SPDX-License-Identifier: AGPL-3.0-or-later

const SLUG_RE = /^[a-z0-9]+(?:-[a-z0-9]+)*$/
const SLUG_MAX = 64

export function slugFromDisplayName(name: string): string {
  const slug = name
    .toLowerCase()
    .replaceAll(/[^a-z0-9]+/g, '-')
    .replaceAll(/^-+|-+$/g, '')

  if (slug.length === 0 || slug.length > SLUG_MAX || !SLUG_RE.test(slug)) {
    return ''
  }

  return slug
}
