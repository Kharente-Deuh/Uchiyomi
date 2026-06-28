// SPDX-License-Identifier: AGPL-3.0-or-later

import type { H3Event } from 'h3'
import type { ExtensionModel } from '../../server/domains/extensions/extension.domain'
import type { UserModel } from '../../server/domains/identity/users/user.domain'
import { createError } from 'h3'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

// Replace the service module before importing the guard: the real factory wires
// Suwayomi adapters from runtime config at module scope, which is unavailable in
// the plain node test project.
const { getExtensionByPkgName } = vi.hoisted(() => ({ getExtensionByPkgName: vi.fn() }))

vi.mock('~~/server/domains/extensions/application/extensions.service', () => ({
  extensionsService: () => ({ getExtensionByPkgName }),
}))

const { extensionGuard, requireExtension } = await import(
  '../../server/domains/extensions/infrastructure/transport/http/guards/extension.guard',
)

// The guard uses Nitro's auto-imported createError/getRouterParam; provide the
// real createError and a minimal getRouterParam that reads event.context.params.
beforeEach(() => {
  vi.stubGlobal('createError', createError)
  vi.stubGlobal('getRouterParam', (event: H3Event, name: string) => event?.context?.params?.[name])
})

afterEach(() => {
  vi.unstubAllGlobals()
  vi.clearAllMocks()
})

function user(over: Partial<UserModel> = {}): UserModel {
  return { id: 'u1', canManageExtensions: false, allowNsfw: false, ...over } as UserModel
}

function extension(over: Partial<ExtensionModel> = {}): ExtensionModel {
  return { pkgName: 'pkg', isInstalled: true, hasUpdate: false, isNsfw: false, ...over } as ExtensionModel
}

function eventWith(authUser: UserModel | undefined, pkgName: string | undefined = 'pkg'): H3Event {
  return { context: { authUser, params: pkgName === undefined ? {} : { pkgName } } } as unknown as H3Event
}

describe('extensionGuard', () => {
  it('rejects with 401 when unauthenticated (before any extension load)', async () => {
    await expect(extensionGuard(eventWith())).rejects.toMatchObject({ statusCode: 401 })
    expect(getExtensionByPkgName).not.toHaveBeenCalled()
  })

  it('rejects with 400 when the pkgName route param is missing', async () => {
    const event = { context: { authUser: user(), params: {} } } as unknown as H3Event
    await expect(extensionGuard(event)).rejects.toMatchObject({ statusCode: 400 })
  })

  it('rejects with 404 when the extension does not exist', async () => {
    getExtensionByPkgName.mockResolvedValue()
    await expect(extensionGuard(eventWith(user()))).rejects.toMatchObject({ statusCode: 404 })
  })

  it('rejects with 403 when an installed-only route hits a non-installed extension', async () => {
    getExtensionByPkgName.mockResolvedValue(extension({ isInstalled: false }))
    await expect(extensionGuard(eventWith(user()), { installationStatus: 'installed' }))
      .rejects
      .toMatchObject({ statusCode: 403 })
  })

  it('rejects with 403 on an installed extension that has a pending update', async () => {
    getExtensionByPkgName.mockResolvedValue(extension({ hasUpdate: true }))
    await expect(extensionGuard(eventWith(user()), { installationStatus: 'installed' }))
      .rejects
      .toMatchObject({ statusCode: 403 })
  })

  it('allows a pending update through when byPassUpdateCheck is set', async () => {
    getExtensionByPkgName.mockResolvedValue(extension({ hasUpdate: true }))
    await expect(extensionGuard(eventWith(user()), { installationStatus: 'installed', byPassUpdateCheck: true }))
      .resolves
      .toMatchObject({ extension: { pkgName: 'pkg' } })
  })

  it('lets a manager bypass the install gate when adminBypassesInstallCheck is set', async () => {
    getExtensionByPkgName.mockResolvedValue(extension({ isInstalled: false }))
    await expect(
      extensionGuard(eventWith(user({ canManageExtensions: true })), {
        installationStatus: 'installed',
        adminBypassesInstallCheck: true,
      }),
    ).resolves.toMatchObject({ extension: { isInstalled: false } })
  })

  it('rejects with 403 when a non-NSFW user requests an NSFW extension', async () => {
    getExtensionByPkgName.mockResolvedValue(extension({ isNsfw: true }))
    await expect(extensionGuard(eventWith(user({ allowNsfw: false })))).rejects.toMatchObject({ statusCode: 403 })
  })

  it('returns the authUser and extension on success', async () => {
    const authUser = user()
    const ext = extension()
    getExtensionByPkgName.mockResolvedValue(ext)
    await expect(extensionGuard(eventWith(authUser))).resolves.toEqual({ authUser, extension: ext })
  })
})

describe('requireExtension', () => {
  it('rejects with 403 when an NSFW extension is loaded for a non-NSFW user', async () => {
    getExtensionByPkgName.mockResolvedValue(extension({ isNsfw: true }))
    await expect(requireExtension(eventWith(user()), user({ allowNsfw: false })))
      .rejects
      .toMatchObject({ statusCode: 403 })
  })

  it('returns the extension when all gates pass', async () => {
    const ext = extension()
    getExtensionByPkgName.mockResolvedValue(ext)
    await expect(requireExtension(eventWith(user()), user())).resolves.toBe(ext)
  })
})
