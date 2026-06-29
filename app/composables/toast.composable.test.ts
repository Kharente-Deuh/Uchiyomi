// SPDX-License-Identifier: AGPL-3.0-or-later

import { beforeEach, describe, expect, it } from 'vitest'
import { useToast } from '~/composables/toast.composable'

describe('useToast', () => {
  beforeEach(() => {
    // The queue is module-level (shared); clear it between tests.
    useToast().messages.value.splice(0)
  })

  it('success pushes a success-coloured message onto the shared queue', () => {
    const { messages, success } = useToast()
    success('Saved')
    expect(messages.value).toEqual([{ text: 'Saved', color: 'success' }])
  })

  it('error pushes an error-coloured message', () => {
    const { messages, error } = useToast()
    error('Boom')
    expect(messages.value).toEqual([{ text: 'Boom', color: 'error' }])
  })

  it('shares one queue across calls (singleton)', () => {
    useToast().success('a')
    expect(useToast().messages.value).toHaveLength(1)
  })
})
