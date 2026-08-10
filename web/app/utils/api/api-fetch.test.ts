// SPDX-License-Identifier: AGPL-3.0-or-later

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { initApi } from './api-fetch'

const { create } = vi.hoisted(() => ({ create: vi.fn(() => 'fetch-client') }))

vi.mock('#build/fetch.mjs', () => ({ $fetch: { create } }))

beforeEach(() => {
  create.mockClear()
})

describe('initApi', () => {
  it('roots every call at the /api prefix', () => {
    initApi()

    expect(create).toHaveBeenCalledWith({ baseURL: '/api' })
  })

  it('appends the endpoint to the /api prefix', () => {
    initApi('/oidc/providers')

    expect(create).toHaveBeenCalledWith({ baseURL: '/api/oidc/providers' })
  })

  it('returns the created client', () => {
    expect(initApi()).toBe('fetch-client')
  })
})
