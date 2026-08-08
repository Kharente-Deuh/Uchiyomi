// SPDX-License-Identifier: AGPL-3.0-or-later

export interface RelativeTimeOpts {
  locale: string
  now?: Date
  timeZone?: string
}

function ymdInTz(date: Date, timeZone?: string): string {
  return new Intl.DateTimeFormat('en-CA', {
    timeZone,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).format(date)
}

function calendarDayDiff(from: Date, to: Date, timeZone?: string): number {
  const a = Date.parse(`${ymdInTz(from, timeZone)}T00:00:00Z`)
  const b = Date.parse(`${ymdInTz(to, timeZone)}T00:00:00Z`)

  return Math.round((b - a) / 86_400_000)
}

export function formatRelativeTime(value: Date | string | number, options: RelativeTimeOpts): string {
  const date = value instanceof Date ? value : new Date(value)
  const now = options.now ?? new Date()

  const auto = new Intl.RelativeTimeFormat(options.locale, { numeric: 'auto' })
  const always = new Intl.RelativeTimeFormat(options.locale, { numeric: 'always' })

  const days = calendarDayDiff(date, now, options.timeZone)

  if (days <= 0) {
    const minutes = Math.round((now.getTime() - date.getTime()) / 60_000)
    if (minutes < 60) {
      return always.format(-Math.max(minutes, 0), 'minute')
    }

    return always.format(-Math.round(minutes / 60), 'hour')
  }

  if (days === 1) {
    return auto.format(-1, 'day')
  }

  if (days < 7) {
    return always.format(-days, 'day')
  }

  if (days < 30) {
    return always.format(-Math.floor(days / 7), 'week')
  }

  if (days < 365) {
    return always.format(-Math.floor(days / 30), 'month')
  }

  return always.format(-Math.floor(days / 365), 'year')
}
