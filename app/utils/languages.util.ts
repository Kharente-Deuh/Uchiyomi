// SPDX-License-Identifier: AGPL-3.0-or-later
export function endomyLanguage(lang: string): string {
  return new Intl.DisplayNames([lang], { type: 'language' }).of(lang) ?? lang
}
