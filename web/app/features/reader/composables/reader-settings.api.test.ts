// SPDX-License-Identifier: AGPL-3.0-or-later

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createReaderSettingsApi } from './reader-settings.api'

const call = vi.fn()

vi.mock('~/utils/api', async importOriginal => ({
  ...(await importOriginal<typeof import('~/utils/api')>()),
  initApi: () => call,
}))

const settings = {
  type: 'manhwa',
  readingMode: 'webtoon',
  pageScale: 'fit-width',
  doublePage: false,
} as const

beforeEach(() => {
  call.mockReset()
})

describe('createReaderSettingsApi().getReaderSettings', () => {
  it('unwraps the items list', async () => {
    call.mockResolvedValue({ items: [settings] })

    const res = await createReaderSettingsApi().getReaderSettings()

    expect(call).toHaveBeenCalledWith('/')
    expect(res).toEqual({ success: true, data: [settings] })
  })

  it('returns an empty list untouched', async () => {
    call.mockResolvedValue({ items: [] })

    await expect(createReaderSettingsApi().getReaderSettings()).resolves.toEqual({
      success: true,
      data: [],
    })
  })

  it('surfaces a failure with its status', async () => {
    call.mockRejectedValue({ statusCode: 500, data: {} })

    const res = await createReaderSettingsApi().getReaderSettings()

    expect(res.success === false && res.error.status).toBe(500)
  })
})

describe('createReaderSettingsApi().updateReaderSettings', () => {
  it('puts the body on the type route', async () => {
    call.mockResolvedValue(settings)

    const res = await createReaderSettingsApi().updateReaderSettings({ ...settings })

    expect(call).toHaveBeenCalledWith('/manhwa', {
      method: 'PUT',
      body: { readingMode: 'webtoon', pageScale: 'fit-width', doublePage: false },
    })
    expect(res).toEqual({ success: true, data: settings })
  })

  it('surfaces a 404 with its status', async () => {
    call.mockRejectedValue({ statusCode: 404, data: {} })

    const res = await createReaderSettingsApi().updateReaderSettings({ ...settings })

    expect(res.success === false && res.error.status).toBe(404)
  })
})
