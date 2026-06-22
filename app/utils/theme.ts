// SPDX-License-Identifier: AGPL-3.0-or-later

/**
 * Pick the theme to apply on boot: the persisted choice when it is a known
 * theme, otherwise the fallback (Vuetify's current/default theme). Keeps the
 * client-only SPA from resetting to `defaultTheme` on every reload.
 */
export function resolveInitialTheme(
  saved: string | undefined,
  valid: readonly string[],
  fallback: string,
): string {
  return saved !== undefined && valid.includes(saved) ? saved : fallback
}
