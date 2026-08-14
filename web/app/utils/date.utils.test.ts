// SPDX-License-Identifier: AGPL-3.0-or-later

import { describe, expect, it } from 'vitest'
import { formatRelativeTime } from './date.utils'

const NOW = new Date('2026-06-28T12:00:00Z')
const TZ = 'UTC'

describe('formatRelativeTime', () => {
  it('formats minutes ago on the same calendar day', () => {
    expect(formatRelativeTime(new Date('2026-06-28T11:55:00Z'), { locale: 'en', now: NOW, timeZone: TZ }))
      .toBe('5 minutes ago')
  })

  it('clamps a future (negative) minute delta to 0', () => {
    expect(formatRelativeTime(new Date('2026-06-28T12:30:00Z'), { locale: 'en', now: NOW, timeZone: TZ }))
      .toBe('0 minutes ago')
  })

  it('formats hours ago on the same calendar day (minutes >= 60)', () => {
    expect(formatRelativeTime(new Date('2026-06-28T09:00:00Z'), { locale: 'en', now: NOW, timeZone: TZ }))
      .toBe('3 hours ago')
  })

  it('formats "yesterday" for a one-day difference', () => {
    expect(formatRelativeTime(new Date('2026-06-27T12:00:00Z'), { locale: 'en', now: NOW, timeZone: TZ }))
      .toBe('yesterday')
  })

  it('formats days ago for a sub-week difference', () => {
    expect(formatRelativeTime(new Date('2026-06-25T12:00:00Z'), { locale: 'en', now: NOW, timeZone: TZ }))
      .toBe('3 days ago')
  })

  it('formats weeks ago for a sub-month difference (floor)', () => {
    expect(formatRelativeTime(new Date('2026-06-20T12:00:00Z'), { locale: 'en', now: NOW, timeZone: TZ }))
      .toBe('1 week ago')
  })

  it('formats multiple weeks ago', () => {
    expect(formatRelativeTime(new Date('2026-06-07T12:00:00Z'), { locale: 'en', now: NOW, timeZone: TZ }))
      .toBe('3 weeks ago')
  })

  it('formats months ago for a sub-year difference (floor)', () => {
    expect(formatRelativeTime(new Date('2026-05-19T12:00:00Z'), { locale: 'en', now: NOW, timeZone: TZ }))
      .toBe('1 month ago')
  })

  it('formats multiple months ago', () => {
    expect(formatRelativeTime(new Date('2026-01-19T12:00:00Z'), { locale: 'en', now: NOW, timeZone: TZ }))
      .toBe('5 months ago')
  })

  it('formats years ago for a >= 365-day difference', () => {
    expect(formatRelativeTime(new Date('2025-05-24T12:00:00Z'), { locale: 'en', now: NOW, timeZone: TZ }))
      .toBe('1 year ago')
  })

  it('accepts an ISO string value', () => {
    expect(formatRelativeTime('2026-06-25T12:00:00Z', { locale: 'en', now: NOW, timeZone: TZ }))
      .toBe('3 days ago')
  })

  it('accepts an epoch-millisecond number value', () => {
    const epoch = new Date('2026-06-25T12:00:00Z').getTime()
    expect(formatRelativeTime(epoch, { locale: 'en', now: NOW, timeZone: TZ }))
      .toBe('3 days ago')
  })

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

  it('honours the timeZone option when computing calendar days', () => {
    const now = new Date('2026-06-28T02:00:00Z')
    const value = new Date('2026-06-27T20:00:00Z')

    expect(formatRelativeTime(value, { locale: 'en', now, timeZone: 'UTC' }))
      .toBe('yesterday')
    expect(formatRelativeTime(value, { locale: 'en', now, timeZone: 'Asia/Tokyo' }))
      .toBe('6 hours ago')
  })

  describe('direction: future', () => {
    it('formats minutes until on the same calendar day', () => {
      expect(formatRelativeTime(new Date('2026-06-28T12:05:00Z'), { locale: 'en', now: NOW, timeZone: TZ, direction: 'future' }))
        .toBe('in 5 minutes')
    })

    it('clamps a past (negative) minute delta to 0', () => {
      expect(formatRelativeTime(new Date('2026-06-28T11:55:00Z'), { locale: 'en', now: NOW, timeZone: TZ, direction: 'future' }))
        .toBe('in 0 minutes')
    })

    it('formats hours until on the same calendar day', () => {
      expect(formatRelativeTime(new Date('2026-06-28T15:00:00Z'), { locale: 'en', now: NOW, timeZone: TZ, direction: 'future' }))
        .toBe('in 3 hours')
    })

    it('formats "tomorrow" for a one-day difference', () => {
      expect(formatRelativeTime(new Date('2026-06-29T12:00:00Z'), { locale: 'en', now: NOW, timeZone: TZ, direction: 'future' }))
        .toBe('tomorrow')
    })

    it('formats days until for a sub-week difference', () => {
      expect(formatRelativeTime(new Date('2026-07-01T12:00:00Z'), { locale: 'en', now: NOW, timeZone: TZ, direction: 'future' }))
        .toBe('in 3 days')
    })

    it('localises minutes to French', () => {
      expect(formatRelativeTime(new Date('2026-06-28T12:05:00Z'), { locale: 'fr', now: NOW, timeZone: TZ, direction: 'future' }))
        .toBe('dans 5 minutes')
    })

    it('localises "tomorrow" to French ("demain")', () => {
      expect(formatRelativeTime(new Date('2026-06-29T12:00:00Z'), { locale: 'fr', now: NOW, timeZone: TZ, direction: 'future' }))
        .toBe('demain')
    })
  })
})
