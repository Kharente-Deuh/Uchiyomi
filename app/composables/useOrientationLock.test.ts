// SPDX-License-Identifier: AGPL-3.0-or-later
import { describe, expect, it } from 'vitest'
import { shouldLockPortrait, shouldShowOrientationOverlay } from '~/composables/useOrientationLock'

describe('shouldLockPortrait', () => {
  it('locks when enabled and the API is supported', () => {
    expect(shouldLockPortrait(true, true)).toBe(true)
  })
  it('does not lock when disabled (e.g. desktop or opted-out)', () => {
    expect(shouldLockPortrait(false, true)).toBe(false)
  })
  it('does not lock when the API is unsupported', () => {
    expect(shouldLockPortrait(true, false)).toBe(false)
  })
})

describe('shouldShowOrientationOverlay', () => {
  it('shows the overlay when enabled and in landscape', () => {
    expect(shouldShowOrientationOverlay(true, 'landscape-primary')).toBe(true)
  })
  it('hides the overlay in portrait', () => {
    expect(shouldShowOrientationOverlay(true, 'portrait-primary')).toBe(false)
  })
  it('hides the overlay when disabled', () => {
    expect(shouldShowOrientationOverlay(false, 'landscape-primary')).toBe(false)
  })
  it('hides the overlay when orientation is unknown', () => {
    expect(shouldShowOrientationOverlay(true)).toBe(false)
  })
})
