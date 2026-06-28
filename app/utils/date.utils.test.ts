// SPDX-License-Identifier: AGPL-3.0-or-later

import { describe, expect, it } from 'vitest'
import { formatRelativeTime } from './date.utils'

// A fixed reference point so every case is deterministic and never reads the
// wall clock. We pin the timeZone to UTC so the calendar-day arithmetic is
// stable regardless of where the suite runs.
const NOW = new Date('2026-06-28T12:00:00Z')
const TZ = 'UTC'

describe('formatRelativeTime', () => {
  // --- same calendar day → minutes ---

  it('formats minutes ago on the same calendar day', () => {
    expect(formatRelativeTime(new Date('2026-06-28T11:55:00Z'), { locale: 'en', now: NOW, timeZone: TZ }))
      .toBe('5 minutes ago')
  })

  it('clamps a future (negative) minute delta to 0', () => {
    expect(formatRelativeTime(new Date('2026-06-28T12:30:00Z'), { locale: 'en', now: NOW, timeZone: TZ }))
      .toBe('0 minutes ago')
  })

  // --- same calendar day → hours ---

  it('formats hours ago on the same calendar day (minutes >= 60)', () => {
    expect(formatRelativeTime(new Date('2026-06-28T09:00:00Z'), { locale: 'en', now: NOW, timeZone: TZ }))
      .toBe('3 hours ago')
  })

  // --- days === 1 → "yesterday" (numeric: auto) ---

  it('formats "yesterday" for a one-day difference', () => {
    expect(formatRelativeTime(new Date('2026-06-27T12:00:00Z'), { locale: 'en', now: NOW, timeZone: TZ }))
      .toBe('yesterday')
  })

  // --- days < 7 → days ---

  it('formats days ago for a sub-week difference', () => {
    expect(formatRelativeTime(new Date('2026-06-25T12:00:00Z'), { locale: 'en', now: NOW, timeZone: TZ }))
      .toBe('3 days ago')
  })

  // --- days < 30 → weeks ---

  it('formats weeks ago for a sub-month difference (floor)', () => {
    // 8 days → floor(8 / 7) === 1 week
    expect(formatRelativeTime(new Date('2026-06-20T12:00:00Z'), { locale: 'en', now: NOW, timeZone: TZ }))
      .toBe('1 week ago')
  })

  it('formats multiple weeks ago', () => {
    // 21 days → floor(21 / 7) === 3 weeks
    expect(formatRelativeTime(new Date('2026-06-07T12:00:00Z'), { locale: 'en', now: NOW, timeZone: TZ }))
      .toBe('3 weeks ago')
  })

  // --- days < 365 → months ---

  it('formats months ago for a sub-year difference (floor)', () => {
    // 40 days → floor(40 / 30) === 1 month
    expect(formatRelativeTime(new Date('2026-05-19T12:00:00Z'), { locale: 'en', now: NOW, timeZone: TZ }))
      .toBe('1 month ago')
  })

  it('formats multiple months ago', () => {
    // 160 days → floor(160 / 30) === 5 months
    expect(formatRelativeTime(new Date('2026-01-19T12:00:00Z'), { locale: 'en', now: NOW, timeZone: TZ }))
      .toBe('5 months ago')
  })

  // --- days >= 365 → years ---

  it('formats years ago for a >= 365-day difference', () => {
    // 400 days → floor(400 / 365) === 1 year
    expect(formatRelativeTime(new Date('2025-05-24T12:00:00Z'), { locale: 'en', now: NOW, timeZone: TZ }))
      .toBe('1 year ago')
  })

  // --- input shapes: ISO string and epoch number ---

  it('accepts an ISO string value', () => {
    expect(formatRelativeTime('2026-06-25T12:00:00Z', { locale: 'en', now: NOW, timeZone: TZ }))
      .toBe('3 days ago')
  })

  it('accepts an epoch-millisecond number value', () => {
    const epoch = new Date('2026-06-25T12:00:00Z').getTime()
    expect(formatRelativeTime(epoch, { locale: 'en', now: NOW, timeZone: TZ }))
      .toBe('3 days ago')
  })

  // --- locale: 'fr' ---

  it('localises minutes to French', () => {
    expect(formatRelativeTime(new Date('2026-06-28T11:55:00Z'), { locale: 'fr', now: NOW, timeZone: TZ }))
      .toBe('il y a 5 minutes')
  })

  it('localises "yesterday" to French ("hier")', () => {
    expect(formatRelativeTime(new Date('2026-06-27T12:00:00Z'), { locale: 'fr', now: NOW, timeZone: TZ }))
      .toBe('hier')
  })

  it('localises days to French', () => {
    expect(formatRelativeTime(new Date('2026-06-25T12:00:00Z'), { locale: 'fr', now: NOW, timeZone: TZ }))
      .toBe('il y a 3 jours')
  })

  // --- timeZone option affects the calendar-day boundary ---

  it('honours the timeZone option when computing calendar days', () => {
    // Same two instants ~6h apart. In UTC they straddle midnight (different
    // calendar days → "yesterday"); in Asia/Tokyo they share a calendar day,
    // so the diff falls into the hours branch instead.
    const now = new Date('2026-06-28T02:00:00Z')
    const value = new Date('2026-06-27T20:00:00Z')

    expect(formatRelativeTime(value, { locale: 'en', now, timeZone: 'UTC' }))
      .toBe('yesterday')
    expect(formatRelativeTime(value, { locale: 'en', now, timeZone: 'Asia/Tokyo' }))
      .toBe('6 hours ago')
  })
})
