// SPDX-License-Identifier: AGPL-3.0-or-later

export function useRelativeTime() {
  const { locale } = useI18n()

  return (value: Date | string | number, timeZone?: string) => formatRelativeTime(value, { locale: locale.value, timeZone })
}
