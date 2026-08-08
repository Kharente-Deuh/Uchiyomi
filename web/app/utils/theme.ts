// SPDX-License-Identifier: AGPL-3.0-or-later

export function resolveInitialTheme(
  saved: string | undefined,
  valid: readonly string[],
  fallback: string,
): string {
  return saved !== undefined && valid.includes(saved) ? saved : fallback
}
