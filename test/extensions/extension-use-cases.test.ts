// SPDX-License-Identifier: AGPL-3.0-or-later
import { describe, expect, it, vi } from 'vitest'
import * as InstallExtension from '../../server/domains/extensions/application/usecases/install-extension.use-case'
import * as ListExtensionSources from '../../server/domains/extensions/application/usecases/list-extension-sources.use-case'
import * as ListExtensions from '../../server/domains/extensions/application/usecases/list-extensions.use-case'
import * as SetSourceEnabled from '../../server/domains/extensions/application/usecases/set-source-enabled.use-case'
import * as UninstallExtension from '../../server/domains/extensions/application/usecases/uninstall-extension.use-case'
import * as UpdateExtension from '../../server/domains/extensions/application/usecases/update-extension.use-case'

function avail(pkgName: string, isNsfw: boolean): { pkgName: string, name: string, lang: string, isNsfw: boolean, isInstalled: boolean, hasUpdate: boolean, versionName: string } {
  return { pkgName, name: pkgName, lang: 'en', isNsfw, isInstalled: true, hasUpdate: false, versionName: '1' }
}

describe('listExtensions.UseCase', () => {
  const page = [
    { pkgName: 'a', name: 'A', lang: 'en', isNsfw: false, isInstalled: true, hasUpdate: false, versionName: '1' },
    { pkgName: 'b', name: 'B', lang: 'en', isNsfw: true, isInstalled: true, hasUpdate: true, versionName: '1' },
  ]
  const health = [
    { pkgName: 'a', health: 'OK', consecutiveFailures: 0 },
    { pkgName: 'b', health: 'ERROR', consecutiveFailures: 2 },
  ]

  function makeUseCase(): { uc: ListExtensions.ListExtensionsUseCase, listExtensions: ReturnType<typeof vi.fn> } {
    const listExtensions = vi.fn().mockResolvedValue({ items: page, totalCount: 2 })
    const suwayomi = { listExtensions } as never
    const overlay = { listHealthByPkgNames: async (_pkgNames: string[]) => health } as never

    return { uc: new ListExtensions.ListExtensionsUseCase(suwayomi, overlay), listExtensions }
  }

  it('passes user filters through and annotates isHealthy + pagination echo', async () => {
    const { uc, listExtensions } = makeUseCase()
    const res = await uc.execute({
      isAdmin: true,
      viewerCanSeeNsfw: true,
      page: 1,
      pageSize: 20,
      filters: { search: 'a', hasUpdate: true },
    })

    expect(listExtensions).toHaveBeenCalledWith({
      page: 1,
      pageSize: 20,
      filters: { search: 'a', pkgName: undefined, isInstalled: undefined, hasUpdate: true, isNsfw: undefined },
    })
    expect(res).toMatchObject({ page: 1, pageSize: 20, totalCount: 2 })
    expect(res.items.map(e => [e.pkgName, e.isHealthy])).toEqual([['a', true], ['b', false]])
  })

  it('forces isInstalled=true for non-admins', async () => {
    const { uc, listExtensions } = makeUseCase()
    await uc.execute({ isAdmin: false, viewerCanSeeNsfw: true, page: 1, pageSize: 20, filters: { isInstalled: false } })

    expect(listExtensions).toHaveBeenCalledWith(expect.objectContaining({
      filters: expect.objectContaining({ isInstalled: true }),
    }))
  })

  it('forces isNsfw=false and ignores the client nsfw flag when the viewer cannot see NSFW', async () => {
    const { uc, listExtensions } = makeUseCase()
    await uc.execute({ isAdmin: true, viewerCanSeeNsfw: false, page: 1, pageSize: 20, filters: { nsfw: true } })

    expect(listExtensions).toHaveBeenCalledWith(expect.objectContaining({
      filters: expect.objectContaining({ isNsfw: false }),
    }))
  })

  it('honours the client nsfw flag when the viewer can see NSFW', async () => {
    const { uc, listExtensions } = makeUseCase()
    await uc.execute({ isAdmin: true, viewerCanSeeNsfw: true, page: 1, pageSize: 20, filters: { nsfw: true } })

    expect(listExtensions).toHaveBeenCalledWith(expect.objectContaining({
      filters: expect.objectContaining({ isNsfw: true }),
    }))
  })
})

describe('install-extension', () => {
  it('installs in Suwayomi, upserts the overlay row and returns the installed extension', async () => {
    const suwayomi = { install: vi.fn().mockResolvedValue(), getExtension: vi.fn().mockResolvedValue(avail('p1', true)), listSources: vi.fn().mockResolvedValue([]) } as never
    const overlay = { upsertInstalled: vi.fn().mockResolvedValue(), recordSuccess: vi.fn(), recordFailure: vi.fn(), findHealth: vi.fn().mockResolvedValue() } as never
    const sources = { syncForExtension: vi.fn().mockResolvedValue() } as never
    const useCase = new InstallExtension.InstallExtensionUseCase(suwayomi, overlay, sources)

    const res = await useCase.execute({ pkgName: 'p1', actorId: 'admin1' })

    expect(suwayomi.install).toHaveBeenCalledWith('p1')
    expect(overlay.upsertInstalled).toHaveBeenCalledWith(expect.objectContaining({ pkgName: 'p1', isNsfw: true, installedByUserId: 'admin1' }))
    expect(res).toMatchObject({ pkgName: 'p1', isInstalled: true, isHealthy: true })
  })

  it('records a failure when Suwayomi install throws', async () => {
    const suwayomi = { install: vi.fn().mockRejectedValue(new Error('boom')), getExtension: vi.fn().mockResolvedValue(avail('p1', false)) } as never
    const overlay = { upsertInstalled: vi.fn().mockResolvedValue(), recordFailure: vi.fn().mockResolvedValue() } as never
    const sources = { syncForExtension: vi.fn().mockResolvedValue() } as never
    const useCase = new InstallExtension.InstallExtensionUseCase(suwayomi, overlay, sources)

    await expect(useCase.execute({ pkgName: 'p1', actorId: 'admin1' })).rejects.toThrow('boom')
    // upsert the row first so the failure has somewhere to attach, then record.
    expect(overlay.recordFailure).toHaveBeenCalledWith(expect.objectContaining({ pkgName: 'p1', context: 'install' }))
  })

  it('syncs source rows after a successful install', async () => {
    const calls: { pkgName: string, ids: string[] }[] = []
    const suwayomi = {
      getExtension: async () => ({ pkgName: 'p', name: 'P', lang: 'en', isNsfw: false, isInstalled: false, hasUpdate: false, versionName: '1' }),
      install: async () => {},
      listSources: async () => [{ id: 's1', name: 'S1', lang: 'en', isNsfw: false, isConfigurable: true }],
    } as any
    const overlay = { upsertInstalled: async () => {}, recordSuccess: async () => {}, recordFailure: async () => {}, findHealth: async () => {} } as any
    const sources = { syncForExtension: async (pkgName: string, s: any[]) => {
      calls.push({ pkgName, ids: s.map(x => x.id) })
    } } as any

    const uc = new InstallExtension.InstallExtensionUseCase(suwayomi, overlay, sources)
    await uc.execute({ pkgName: 'p', actorId: 'u1' })
    expect(calls).toEqual([{ pkgName: 'p', ids: ['s1'] }])
  })
})

describe('update-extension', () => {
  it('updates in Suwayomi, records success, re-syncs sources and returns the post-update extension', async () => {
    const before = { ...avail('p1', false), hasUpdate: true } // installed + outdated
    const after = { ...avail('p1', false), hasUpdate: false } // installed + up-to-date after the update
    const suwayomi = {
      getExtension: vi.fn().mockResolvedValueOnce(before).mockResolvedValueOnce(after),
      update: vi.fn().mockResolvedValue(),
      listSources: vi.fn().mockResolvedValue([{ id: 's1', name: 'S1', lang: 'en', isNsfw: false, isConfigurable: true }]),
    } as never
    const overlay = {
      recordSuccess: vi.fn().mockResolvedValue(),
      recordFailure: vi.fn(),
      findHealth: vi.fn().mockResolvedValue({ pkgName: 'p1', health: 'OK', consecutiveFailures: 0 }),
    } as never
    const sources = { syncForExtension: vi.fn().mockResolvedValue() } as never
    const uc = new UpdateExtension.UpdateExtensionUseCase(suwayomi, overlay, sources)

    const res = await uc.execute({ pkgName: 'p1' })

    expect(suwayomi.update).toHaveBeenCalledWith('p1')
    expect(overlay.recordSuccess).toHaveBeenCalledWith('p1')
    expect(sources.syncForExtension).toHaveBeenCalledWith('p1', [expect.objectContaining({ id: 's1' })])
    expect(res).toMatchObject({ pkgName: 'p1', hasUpdate: false, isHealthy: true })
  })

  it('records a failure with context "update" and rethrows when Suwayomi update throws', async () => {
    const suwayomi = { getExtension: vi.fn().mockResolvedValue(avail('p1', false)), update: vi.fn().mockRejectedValue(new Error('boom')), listSources: vi.fn() } as never
    const overlay = { recordSuccess: vi.fn(), recordFailure: vi.fn().mockResolvedValue() } as never
    const sources = { syncForExtension: vi.fn() } as never
    const uc = new UpdateExtension.UpdateExtensionUseCase(suwayomi, overlay, sources)

    await expect(uc.execute({ pkgName: 'p1' })).rejects.toThrow('boom')
    expect(overlay.recordFailure).toHaveBeenCalledWith(expect.objectContaining({ pkgName: 'p1', context: 'update' }))
    expect(overlay.recordSuccess).not.toHaveBeenCalled()
  })

  it('throws and does not call update when the extension is not installed', async () => {
    const notInstalled = { pkgName: 'p1', name: 'P1', lang: 'en', isNsfw: false, isInstalled: false, hasUpdate: true, versionName: '1' }
    const suwayomi = { getExtension: vi.fn().mockResolvedValue(notInstalled), update: vi.fn() } as never
    const uc = new UpdateExtension.UpdateExtensionUseCase(suwayomi, { recordFailure: vi.fn() } as never, { syncForExtension: vi.fn() } as never)

    await expect(uc.execute({ pkgName: 'p1' })).rejects.toThrow('not installed')
    expect(suwayomi.update).not.toHaveBeenCalled()
  })

  it('throws and does not call update when the extension is not found', async () => {
    const suwayomi = { getExtension: vi.fn().mockResolvedValue(), update: vi.fn() } as never
    const uc = new UpdateExtension.UpdateExtensionUseCase(suwayomi, { recordFailure: vi.fn() } as never, { syncForExtension: vi.fn() } as never)

    await expect(uc.execute({ pkgName: 'nope' })).rejects.toThrow('not found')
    expect(suwayomi.update).not.toHaveBeenCalled()
  })
})

describe('uninstall-extension', () => {
  it('uninstalls in Suwayomi, deletes the overlay row and returns the now-uninstalled extension', async () => {
    const suwayomi = { getExtension: vi.fn().mockResolvedValue(avail('p1', false)), uninstall: vi.fn().mockResolvedValue() } as never
    const overlay = { deleteByPkgName: vi.fn().mockResolvedValue() } as never
    const uc = new UninstallExtension.UninstallExtensionUseCase(suwayomi, overlay)

    const res = await uc.execute({ pkgName: 'p1' })

    expect(suwayomi.uninstall).toHaveBeenCalledWith('p1')
    expect(overlay.deleteByPkgName).toHaveBeenCalledWith('p1')
    expect(res).toMatchObject({ pkgName: 'p1', isInstalled: false })
  })

  it('throws and does not call uninstall when the extension is not found', async () => {
    const suwayomi = { getExtension: vi.fn().mockResolvedValue(), uninstall: vi.fn() } as never
    const uc = new UninstallExtension.UninstallExtensionUseCase(suwayomi, { deleteByPkgName: vi.fn() } as never)

    await expect(uc.execute({ pkgName: 'nope' })).rejects.toThrow('not found')
    expect(suwayomi.uninstall).not.toHaveBeenCalled()
  })
})

describe('setSourceEnabled.UseCase', () => {
  it('sets isEnabled when the source exists', async () => {
    const sources = {
      findById: async () => ({ id: '1', pkgName: 'p', name: 'A', lang: 'en', isNsfw: false, isConfigurable: true, isEnabled: true }),
      setEnabled: async (id: string, isEnabled: boolean) => ({ id, pkgName: 'p', name: 'A', lang: 'en', isNsfw: false, isConfigurable: true, isEnabled }),
    } as any
    const uc = new SetSourceEnabled.SetSourceEnabledUseCase(sources)
    const res = await uc.execute({ sourceId: '1', isEnabled: false })
    expect(res.isEnabled).toBe(false)
  })

  it('throws NotFound when the source is unknown', async () => {
    const sources = { findById: async () => {} } as any
    const uc = new SetSourceEnabled.SetSourceEnabledUseCase(sources)
    await expect(uc.execute({ sourceId: 'x', isEnabled: false })).rejects.toThrow('Source not found')
  })
})

describe('listExtensionSources.UseCase', () => {
  const stored = [
    { id: '1', pkgName: 'p', name: 'A', lang: 'en', isNsfw: false, isConfigurable: true, isEnabled: true },
    { id: '2', pkgName: 'p', name: 'B', lang: 'en', isNsfw: false, isConfigurable: true, isEnabled: false },
    { id: '3', pkgName: 'p', name: 'C', lang: 'en', isNsfw: true, isConfigurable: true, isEnabled: true },
  ]
  const sources = { listByPkg: async () => stored } as any

  it('admin sees all sources', async () => {
    const uc = new ListExtensionSources.ListExtensionSourcesUseCase(sources)
    const res = await uc.execute({ pkgName: 'p', isAdmin: true, viewerCanSeeNsfw: false })
    expect(res.map(s => s.id)).toEqual(['1', '2', '3'])
  })

  it('non-admin sees only enabled, NSFW-gated sources', async () => {
    const uc = new ListExtensionSources.ListExtensionSourcesUseCase(sources)
    const res = await uc.execute({ pkgName: 'p', isAdmin: false, viewerCanSeeNsfw: false })
    expect(res.map(s => s.id)).toEqual(['1']) // 2 disabled, 3 nsfw-gated
  })
})
