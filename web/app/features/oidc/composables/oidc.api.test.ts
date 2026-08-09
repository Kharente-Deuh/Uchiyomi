// SPDX-License-Identifier: AGPL-3.0-or-later

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createOidcApi } from './oidc.api'

const call = vi.fn()

vi.mock('~/utils/api', async importOriginal => ({
  ...(await importOriginal<typeof import('~/utils/api')>()),
  initApi: () => call,
}))

const CREATED_AT = '2026-01-02T03:04:05.000Z'
const UPDATED_AT = '2026-02-03T04:05:06.000Z'

const payload = {
  id: 'p1',
  displayName: 'PocketID',
  issuerUrl: 'https://id.example.org',
  clientId: 'client',
  usernameClaim: 'preferred_username',
  scopes: ['openid', 'email'],
  roleClaim: 'groups',
  adminValues: ['admins'],
  allowedValues: ['users'],
  autoProvision: true,
  createdAt: CREATED_AT,
  updatedAt: UPDATED_AT,
}

const request = {
  displayName: 'PocketID',
  issuerUrl: 'https://id.example.org',
  clientId: 'client',
  clientSecret: 'secret',
  usernameClaim: 'preferred_username',
  scopes: ['openid'],
  autoProvision: false,
}

beforeEach(() => {
  call.mockReset()
})

describe('createOidcApi().getAll', () => {
  it('turns createdAt into a Date on every provider', async () => {
    call.mockResolvedValue([{ id: 'p1', displayName: 'PocketID', createdAt: CREATED_AT, userCount: 3 }])

    const res = await createOidcApi().getAll()

    expect(res).toEqual({
      success: true,
      data: [{ id: 'p1', displayName: 'PocketID', createdAt: new Date(CREATED_AT), userCount: 3 }],
    })
  })

  it('returns an empty list untouched', async () => {
    call.mockResolvedValue([])

    await expect(createOidcApi().getAll()).resolves.toEqual({ success: true, data: [] })
  })

  it('surfaces a failure with its status', async () => {
    call.mockRejectedValue({ statusCode: 500, data: {} })

    const res = await createOidcApi().getAll()

    expect(res.success === false && res.error.status).toBe(500)
  })
})

describe('createOidcApi().getById', () => {
  it('requests the provider by id', async () => {
    call.mockResolvedValue(payload)
    await createOidcApi().getById('p1')

    expect(call).toHaveBeenCalledWith('/p1')
  })

  it('exposes createdAt and updatedAt as Dates', async () => {
    call.mockResolvedValue(payload)

    const res = await createOidcApi().getById('p1')

    expect(res.success === true && res.data.createdAt).toEqual(new Date(CREATED_AT))
    expect(res.success === true && res.data.updatedAt).toEqual(new Date(UPDATED_AT))
  })

  it('surfaces a 404 with its status', async () => {
    call.mockRejectedValue({ statusCode: 404, data: {} })

    const res = await createOidcApi().getById('nope')

    expect(res.success === false && res.error.status).toBe(404)
  })
})

describe('createOidcApi().create', () => {
  it('omits the role block entirely when no roleClaim is given', async () => {
    call.mockResolvedValue(payload)
    await createOidcApi().create(request)

    expect(call).toHaveBeenCalledWith('/', { method: 'POST', body: request })
  })

  it('sends the role claim with its value lists', async () => {
    call.mockResolvedValue(payload)
    await createOidcApi().create({ ...request, roleClaim: 'groups', adminValues: ['admins'], allowedValues: ['users'] })

    expect(call).toHaveBeenCalledWith('/', {
      method: 'POST',
      body: { ...request, roleClaim: 'groups', adminValues: ['admins'], allowedValues: ['users'] },
    })
  })

  it('drops empty value lists but keeps the role claim', async () => {
    call.mockResolvedValue(payload)
    await createOidcApi().create({ ...request, roleClaim: 'groups', adminValues: [], allowedValues: [] })

    expect(call).toHaveBeenCalledWith('/', { method: 'POST', body: { ...request, roleClaim: 'groups' } })
  })

  it('drops the value lists when the role claim is empty', async () => {
    call.mockResolvedValue(payload)
    await createOidcApi().create({ ...request, roleClaim: '', adminValues: ['admins'] })

    expect(call).toHaveBeenCalledWith('/', { method: 'POST', body: request })
  })

  it('normalizes the response into a provider', async () => {
    call.mockResolvedValue(payload)

    const res = await createOidcApi().create(request)

    expect(res).toEqual({
      success: true,
      data: {
        id: 'p1',
        displayName: 'PocketID',
        issuerUrl: 'https://id.example.org',
        clientId: 'client',
        usernameClaim: 'preferred_username',
        scopes: ['openid', 'email'],
        roleClaim: 'groups',
        adminValues: ['admins'],
        allowedValues: ['users'],
        autoProvision: true,
        createdAt: new Date(CREATED_AT),
        updatedAt: new Date(UPDATED_AT),
      },
    })
  })

  it('collapses empty value lists to undefined', async () => {
    call.mockResolvedValue({ ...payload, roleClaim: null, adminValues: [], allowedValues: [] })

    const res = await createOidcApi().create(request)

    expect(res.success === true && res.data.roleClaim).toBeUndefined()
    expect(res.success === true && res.data.adminValues).toBeUndefined()
    expect(res.success === true && res.data.allowedValues).toBeUndefined()
  })

  it('surfaces a 409 with its status', async () => {
    call.mockRejectedValue({ statusCode: 409, data: {} })

    const res = await createOidcApi().create(request)

    expect(res.success === false && res.error.status).toBe(409)
  })
})

describe('createOidcApi().updateById', () => {
  it('puts the payload on the provider route', async () => {
    const { clientSecret: _clientSecret, ...body } = request
    call.mockResolvedValue(payload)

    await createOidcApi().updateById('p1', body)

    expect(call).toHaveBeenCalledWith('/p1', { method: 'PUT', body })
  })

  it('normalizes the response the same way create does', async () => {
    const { clientSecret: _clientSecret, ...body } = request
    call.mockResolvedValue(payload)

    const res = await createOidcApi().updateById('p1', body)

    expect(res.success === true && res.data.allowedValues).toEqual(['users'])
    expect(res.success === true && res.data.createdAt).toEqual(new Date(CREATED_AT))
    expect(res.success === true && res.data.updatedAt).toEqual(new Date(UPDATED_AT))
  })

  it('surfaces a 404 with its status', async () => {
    const { clientSecret: _clientSecret, ...body } = request
    call.mockRejectedValue({ statusCode: 404, data: {} })

    const res = await createOidcApi().updateById('nope', body)

    expect(res.success === false && res.error.status).toBe(404)
  })
})

describe('createOidcApi().testByIssuerUrl', () => {
  it('probes the issuer url', async () => {
    call.mockResolvedValue({ issuer: 'https://id.example.org' })
    await createOidcApi().testByIssuerUrl('https://id.example.org')

    expect(call).toHaveBeenCalledWith('/probe', { method: 'POST', body: { issuerUrl: 'https://id.example.org' } })
  })

  it('surfaces an unreachable issuer with its message', async () => {
    call.mockRejectedValue({ statusCode: 400, data: { message: 'issuer is unreachable' } })

    const res = await createOidcApi().testByIssuerUrl('https://nope.example.org')

    expect(res.success === false && res.error.status).toBe(400)
    expect(res.success === false && res.error.message).toBe('issuer is unreachable')
  })
})

describe('createOidcApi().deleteById', () => {
  it('deletes the provider route', async () => {
    call.mockResolvedValue(undefined)

    await expect(createOidcApi().deleteById('p1')).resolves.toEqual({ success: true, data: undefined })
    expect(call).toHaveBeenCalledWith('/p1', { method: 'DELETE' })
  })

  it('surfaces a 404 with its status', async () => {
    call.mockRejectedValue({ statusCode: 404, data: {} })

    const res = await createOidcApi().deleteById('nope')

    expect(res.success === false && res.error.status).toBe(404)
  })
})
