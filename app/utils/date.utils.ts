// SPDX-License-Identifier: AGPL-3.0-or-later

export interface RelativeTimeOpts {
  locale: string
  now?: Date
  timeZone?: string // défaut: fuseau du navigateur
}

/** "2026-06-28" dans le fuseau donné (en-CA force le format ISO). */
function ymdInTz(date: Date, timeZone?: string): string {
  return new Intl.DateTimeFormat('en-CA', {
    timeZone,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).format(date)
}

/** Différence en jours calendaires, fuseau-aware (gère DST). */
function calendarDayDiff(from: Date, to: Date, timeZone?: string): number {
  const a = Date.parse(`${ymdInTz(from, timeZone)}T00:00:00Z`)
  const b = Date.parse(`${ymdInTz(to, timeZone)}T00:00:00Z`)

  return Math.round((b - a) / 86_400_000)
}

export function formatRelativeTime(value: Date | string | number, opts: RelativeTimeOpts): string {
  const date = value instanceof Date ? value : new Date(value)
  const now = opts.now ?? new Date()

  const auto = new Intl.RelativeTimeFormat(opts.locale, { numeric: 'auto' })
  const always = new Intl.RelativeTimeFormat(opts.locale, { numeric: 'always' })

  const days = calendarDayDiff(date, now, opts.timeZone)

  if (days <= 0) {
    // Aujourd'hui → heures / minutes
    const minutes = Math.round((now.getTime() - date.getTime()) / 60_000)
    if (minutes < 60) {
      return always.format(-Math.max(minutes, 0), 'minute')
    } // "il y a 5 minutes"

    return always.format(-Math.round(minutes / 60), 'hour') // "il y a 3 heures"
  }

  if (days === 1) {
    return auto.format(-1, 'day')
  } // "hier" / "yesterday"

  if (days < 7) {
    return always.format(-days, 'day')
  } // "il y a 3 jours"

  if (days < 30) {
    return always.format(-Math.floor(days / 7), 'week')
  } // "il y a 2 semaines"

  if (days < 365) {
    return always.format(-Math.floor(days / 30), 'month')
  } // "il y a 5 mois"

  return always.format(-Math.floor(days / 365), 'year') // "il y a 1 an"
}
