// SPDX-License-Identifier: AGPL-3.0-or-later

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createSetupApi } from './setup.api'

const call = vi.fn()

vi.mock('~/utils/api', async importOriginal => ({
  ...(await importOriginal<typeof import('~/utils/api')>()),
  initApi: () => call,
}))

describe('createSetupApi().getSetupStatus', () => {
  beforeEach(() => {
    call.mockReset()
  })

  it('maps required: true to required', async () => {
    call.mockResolvedValue({ required: true })

    await expect(createSetupApi().getSetupStatus()).resolves.toBe('required')
  })

  it('maps required: false to done', async () => {
    call.mockResolvedValue({ required: false })

    await expect(createSetupApi().getSetupStatus()).resolves.toBe('done')
  })

  it('falls back to unknown on a payload without the required flag', async () => {
    call.mockResolvedValue({ status: 'ok' })

    await expect(createSetupApi().getSetupStatus()).resolves.toBe('unknown')
  })

  it('falls back to unknown when required is not a boolean', async () => {
    call.mockResolvedValue({ required: 'yes' })

    await expect(createSetupApi().getSetupStatus()).resolves.toBe('unknown')
  })

  it('falls back to unknown on a null payload', async () => {
    call.mockResolvedValue(null)

    await expect(createSetupApi().getSetupStatus()).resolves.toBe('unknown')
  })

  it('falls back to unknown when the call throws', async () => {
    call.mockRejectedValue(new TypeError('Failed to fetch'))

    await expect(createSetupApi().getSetupStatus()).resolves.toBe('unknown')
  })
})

describe('createSetupApi().doSetup', () => {
  beforeEach(() => {
    call.mockReset()
  })

  const body = { username: 'alice', password: 'hunter2' }

  it('posts the credentials to the setup root', async () => {
    call.mockResolvedValue(undefined)

    await expect(createSetupApi().doSetup(body)).resolves.toEqual({ success: true, data: undefined })
    expect(call).toHaveBeenCalledWith('/', { method: 'POST', body })
  })

  it('surfaces a 409 as a failed response carrying the status', async () => {
    call.mockRejectedValue({ statusCode: 409, data: {} })

    const res = await createSetupApi().doSetup(body)

    expect(res.success).toBe(false)
    expect(res.success === false && res.error.status).toBe(409)
  })

  it('surfaces the server message when the payload carries one', async () => {
    call.mockRejectedValue({ statusCode: 400, data: { message: 'username already taken' } })

    const res = await createSetupApi().doSetup(body)

    expect(res.success === false && res.error.message).toBe('username already taken')
  })
})
