// SPDX-License-Identifier: AGPL-3.0-or-later

import { describe, expect, it, vi } from 'vitest'
import { GetVisibleSourceUseCase } from '../../server/domains/extensions/application/usecases/get-visible-source.use-case'

function source(overrides: Partial<{ pkgName: string, isEnabled: boolean, isNsfw: boolean }> = {}): { id: string, name: string, lang: string, isNsfw: boolean, isConfigurable: boolean, pkgName: string, isEnabled: boolean } {
  return { id: 's1', name: 'S', lang: 'en', isNsfw: false, isConfigurable: false, pkgName: 'pkg', isEnabled: true, ...overrides }
}

function makeUseCase(found?: ReturnType<typeof source>): { uc: GetVisibleSourceUseCase, findById: ReturnType<typeof vi.fn> } {
  const findById = vi.fn().mockResolvedValue(found)

  return { uc: new GetVisibleSourceUseCase({ findById } as never), findById }
}

describe('getVisibleSourceUseCase', () => {
  it('returns the source when it exists, belongs to the pkg, is enabled and SFW', async () => {
    const { uc } = makeUseCase(source())
    const res = await uc.execute({ pkgName: 'pkg', sourceId: 's1', isAdmin: false, canSeeNsfw: false })
    expect(res?.id).toBe('s1')
  })

  it('returns undefined when the source does not exist', async () => {
    const { uc } = makeUseCase()
    expect(await uc.execute({ pkgName: 'pkg', sourceId: 's1', isAdmin: true, canSeeNsfw: true })).toBeUndefined()
  })

  it('returns undefined when the source belongs to another extension', async () => {
    const { uc } = makeUseCase(source({ pkgName: 'other' }))
    expect(await uc.execute({ pkgName: 'pkg', sourceId: 's1', isAdmin: true, canSeeNsfw: true })).toBeUndefined()
  })

  it('hides a disabled source from non-admins but shows it to admins', async () => {
    const { uc: ucUser } = makeUseCase(source({ isEnabled: false }))
    expect(await ucUser.execute({ pkgName: 'pkg', sourceId: 's1', isAdmin: false, canSeeNsfw: true })).toBeUndefined()

    const { uc: ucAdmin } = makeUseCase(source({ isEnabled: false }))
    expect(await ucAdmin.execute({ pkgName: 'pkg', sourceId: 's1', isAdmin: true, canSeeNsfw: true })).toBeDefined()
  })

  it('hides an NSFW source when the viewer cannot see NSFW, even as admin', async () => {
    const { uc } = makeUseCase(source({ isNsfw: true }))
    expect(await uc.execute({ pkgName: 'pkg', sourceId: 's1', isAdmin: true, canSeeNsfw: false })).toBeUndefined()
  })
})
