export function endomyLanguage(lang: string): string {
  return new Intl.DisplayNames([lang], { type: 'language' }).of(lang) ?? lang
}
