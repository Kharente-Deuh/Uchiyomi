// SPDX-License-Identifier: AGPL-3.0-or-later
import { describe, expect, it, vi } from 'vitest'
import * as GetExtensionSettings from '../../server/domains/extensions/application/usecases/get-extension-settings.use-case'
import * as InstallExtension from '../../server/domains/extensions/application/usecases/install-extension.use-case'
import * as ListExtensionSources from '../../server/domains/extensions/application/usecases/list-extension-sources.use-case'
import * as ListExtensions from '../../server/domains/extensions/application/usecases/list-extensions.use-case'
import * as SetSourceEnabled from '../../server/domains/extensions/application/usecases/set-source-enabled.use-case'
import * as UninstallExtension from '../../server/domains/extensions/application/usecases/uninstall-extension.use-case'
import * as UpdateExtensionSettings from '../../server/domains/extensions/application/usecases/update-extension-settings.use-case'
import * as UpdateExtension from '../../server/domains/extensions/application/usecases/update-extension.use-case'

function avail(pkgName: string, isNsfw: boolean): { pkgName: string, name: string, lang: string, isNsfw: boolean, isInstalled: boolean, hasUpdate: boolean, versionName: string } {
  return { pkgName, name: pkgName, lang: 'en', isNsfw, isInstalled: true, hasUpdate: false, versionName: '1' }
}

describe('listExtensions.UseCase', () => {
  const page = [
    { pkgName: 'a', name: 'A', lang: 'en', isNsfw: false, isInstalled: true, hasUpdate: false, versionName: '1' },
    { pkgName: 'b', name: 'B', lang: 'en', isNsfw: true, isInstalled: true, hasUpdate: true, versionName: '1' },
  ]
  function makeUseCase(): { uc: ListExtensions.ListExtensionsUseCase, listExtensions: ReturnType<typeof vi.fn> } {
    const listExtensions = vi.fn().mockResolvedValue({ items: page, total: 2 })
    const suwayomi = { listExtensions } as never

    return { uc: new ListExtensions.ListExtensionsUseCase(suwayomi), listExtensions }
  }

  it('passes user filters through and echoes pagination', async () => {
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
    expect(res).toMatchObject({ total: 2 })
    expect(res.items.map(e => e.pkgName)).toEqual(['a', 'b'])
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
    const overlay = { upsertInstalled: vi.fn().mockResolvedValue() } as never
    const sources = { syncForExtension: vi.fn().mockResolvedValue() } as never
    const useCase = new InstallExtension.InstallExtensionUseCase(suwayomi, overlay, sources)

    const res = await useCase.execute({ pkgName: 'p1', actorId: 'admin1' })

    expect(suwayomi.install).toHaveBeenCalledWith('p1')
    expect(overlay.upsertInstalled).toHaveBeenCalledWith(expect.objectContaining({ pkgName: 'p1', isNsfw: true, installedByUserId: 'admin1' }))
    expect(res).toMatchObject({ pkgName: 'p1', isInstalled: true })
  })

  it('rethrows when Suwayomi install throws', async () => {
    const suwayomi = { install: vi.fn().mockRejectedValue(new Error('boom')), getExtension: vi.fn().mockResolvedValue(avail('p1', false)) } as never
    const overlay = { upsertInstalled: vi.fn().mockResolvedValue() } as never
    const sources = { syncForExtension: vi.fn().mockResolvedValue() } as never
    const useCase = new InstallExtension.InstallExtensionUseCase(suwayomi, overlay, sources)

    await expect(useCase.execute({ pkgName: 'p1', actorId: 'admin1' })).rejects.toThrow('boom')
    expect(overlay.upsertInstalled).toHaveBeenCalledWith(expect.objectContaining({ pkgName: 'p1' }))
  })

  it('syncs source rows after a successful install', async () => {
    const calls: { pkgName: string, ids: string[] }[] = []
    const suwayomi = {
      getExtension: async () => ({ pkgName: 'p', name: 'P', lang: 'en', isNsfw: false, isInstalled: false, hasUpdate: false, versionName: '1' }),
      install: async () => {},
      listSources: async () => [{ id: 's1', name: 'S1', lang: 'en', isNsfw: false, isConfigurable: true }],
    } as any
    const overlay = { upsertInstalled: async () => {} } as any
    const sources = { syncForExtension: async (pkgName: string, s: any[]) => {
      calls.push({ pkgName, ids: s.map(x => x.id) })
    } } as any

    const uc = new InstallExtension.InstallExtensionUseCase(suwayomi, overlay, sources)
    await uc.execute({ pkgName: 'p', actorId: 'u1' })
    expect(calls).toEqual([{ pkgName: 'p', ids: ['s1'] }])
  })
})

describe('update-extension', () => {
  it('updates in Suwayomi, re-syncs sources and returns the post-update extension', async () => {
    const before = { ...avail('p1', false), hasUpdate: true } // installed + outdated
    const after = { ...avail('p1', false), hasUpdate: false } // installed + up-to-date after the update
    const suwayomi = {
      getExtension: vi.fn().mockResolvedValueOnce(before).mockResolvedValueOnce(after),
      update: vi.fn().mockResolvedValue(),
      listSources: vi.fn().mockResolvedValue([{ id: 's1', name: 'S1', lang: 'en', isNsfw: false, isConfigurable: true }]),
    } as never
    const sources = { syncForExtension: vi.fn().mockResolvedValue() } as never
    const uc = new UpdateExtension.UpdateExtensionUseCase(suwayomi, sources)

    const res = await uc.execute({ pkgName: 'p1' })

    expect(suwayomi.update).toHaveBeenCalledWith('p1')
    expect(sources.syncForExtension).toHaveBeenCalledWith('p1', [expect.objectContaining({ id: 's1' })])
    expect(res).toMatchObject({ pkgName: 'p1', hasUpdate: false })
  })

  it('rethrows when Suwayomi update throws', async () => {
    const suwayomi = { getExtension: vi.fn().mockResolvedValue(avail('p1', false)), update: vi.fn().mockRejectedValue(new Error('boom')), listSources: vi.fn() } as never
    const sources = { syncForExtension: vi.fn() } as never
    const uc = new UpdateExtension.UpdateExtensionUseCase(suwayomi, sources)

    await expect(uc.execute({ pkgName: 'p1' })).rejects.toThrow('boom')
    expect(sources.syncForExtension).not.toHaveBeenCalled()
  })

  it('throws and does not call update when the extension is not installed', async () => {
    const notInstalled = { pkgName: 'p1', name: 'P1', lang: 'en', isNsfw: false, isInstalled: false, hasUpdate: true, versionName: '1' }
    const suwayomi = { getExtension: vi.fn().mockResolvedValue(notInstalled), update: vi.fn() } as never
    const uc = new UpdateExtension.UpdateExtensionUseCase(suwayomi, { syncForExtension: vi.fn() } as never)

    await expect(uc.execute({ pkgName: 'p1' })).rejects.toThrow('not installed')
    expect(suwayomi.update).not.toHaveBeenCalled()
  })

  it('throws and does not call update when the extension is not found', async () => {
    const suwayomi = { getExtension: vi.fn().mockResolvedValue(), update: vi.fn() } as never
    const uc = new UpdateExtension.UpdateExtensionUseCase(suwayomi, { syncForExtension: vi.fn() } as never)

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
      update: async (id: string, data: { isEnabled: boolean }) => ({ id, pkgName: 'p', name: 'A', lang: 'en', isNsfw: false, isConfigurable: true, ...data }),
    } as any
    const uc = new SetSourceEnabled.SetSourceEnabledUseCase(sources)
    const res = await uc.execute({ sourceId: '1', isEnabled: false })
    expect(res.isEnabled).toBe(false)
  })

  it('throws NotFound when the source is unknown', async () => {
    const sources = { findById: async (): Promise<undefined> => {} } as any
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

  // The use case pushes filtering into findMany; the fake honours the params it receives.
  function makeFakeSources(): any {
    return {
      findMany: async (params: { pkgName: string, isEnabled?: boolean, isNsfw?: boolean }) => {
        return stored.filter((s) => {
          if (params.pkgName !== undefined && s.pkgName !== params.pkgName) {
            return false
          }

          if (params.isEnabled !== undefined && s.isEnabled !== params.isEnabled) {
            return false
          }

          if (params.isNsfw !== undefined && s.isNsfw !== params.isNsfw) {
            return false
          }

          return true
        })
      },
    }
  }

  it('admin sees all sources (including disabled and NSFW when canSeeNsfw)', async () => {
    const uc = new ListExtensionSources.ListExtensionSourcesUseCase(makeFakeSources())
    const res = await uc.execute({ pkgName: 'p', isAdmin: true, canSeeNsfw: true })
    expect(res.map(s => s.id)).toEqual(['1', '2', '3'])
  })

  it('non-admin sees only enabled, NSFW-gated sources', async () => {
    const uc = new ListExtensionSources.ListExtensionSourcesUseCase(makeFakeSources())
    const res = await uc.execute({ pkgName: 'p', isAdmin: false, canSeeNsfw: false })
    expect(res.map(s => s.id)).toEqual(['1']) // 2 disabled, 3 nsfw-gated
  })
})

describe('getExtensionSettings.UseCase', () => {
  it('lists configurable sources, reads each source prefs, and merges', async () => {
    const sources = { findMany: vi.fn().mockResolvedValue([
      { id: 'a', pkgName: 'p', name: 'A', lang: 'fr', isNsfw: false, isConfigurable: true, isEnabled: true },
      { id: 'b', pkgName: 'p', name: 'B', lang: 'en', isNsfw: false, isConfigurable: true, isEnabled: true },
    ]) } as any
    const suwayomi = { listSourcePreferences: vi.fn(async (id: string) => [
      { position: 0, type: 'list', key: 'lang', visible: true, textValue: id === 'a' ? 'fr' : 'en' },
    ]) } as any

    const res = await new GetExtensionSettings.GetExtensionSettingsUseCase(sources, suwayomi).execute({ pkgName: 'p' })

    expect(sources.findMany).toHaveBeenCalledWith({ pkgName: 'p', isConfigurable: true })
    expect(res.common.map(p => p.key)).toEqual(['lang'])
    expect(res.sources.map(s => s.id)).toEqual(['a', 'b'])
  })

  it('returns empty settings without reading prefs when there are no configurable sources', async () => {
    const sources = { findMany: vi.fn().mockResolvedValue([]) } as any
    const suwayomi = { listSourcePreferences: vi.fn() } as any
    const res = await new GetExtensionSettings.GetExtensionSettingsUseCase(sources, suwayomi).execute({ pkgName: 'p' })
    expect(res).toEqual({ common: [], sources: [] })
    expect(suwayomi.listSourcePreferences).not.toHaveBeenCalled()
  })
})

function makeDepsForUpdateExtensionSettings(prefsByCall: Record<string, any[]>): { sources: any, suwayomi: any } {
  const sources = {
    findMany: vi.fn().mockResolvedValue([
      { id: 'a', pkgName: 'p', name: 'A', lang: 'fr', isNsfw: false, isConfigurable: true, isEnabled: true },
      { id: 'b', pkgName: 'p', name: 'B', lang: 'en', isNsfw: false, isConfigurable: true, isEnabled: true },
    ]),
  } as any
  const suwayomi = {
    listSourcePreferences: vi.fn(async (id: string) => prefsByCall[id]),
    updateSourcePreferences: vi.fn().mockResolvedValue([]),
  } as any

  return { sources, suwayomi }
}

describe('updateExtensionSettings.UseCase', () => {
  it('applies common changes to every source, resolving the position per source', async () => {
    const { sources, suwayomi } = makeDepsForUpdateExtensionSettings({
      a: [{ position: 0, type: 'list', key: 'lang', visible: true, textValue: 'fr' }],
      b: [{ position: 7, type: 'list', key: 'lang', visible: true, textValue: 'en' }],
    })
    const settings = { common: [{ position: 0, type: 'list', key: 'lang', visible: true, textValue: 'de' }], sources: [] }

    await new UpdateExtensionSettings.UpdateExtensionSettingsUseCase(sources, suwayomi).execute({ pkgName: 'p', settings })

    expect(suwayomi.updateSourcePreferences).toHaveBeenCalledWith('a', [{ position: 0, type: 'list', textValue: 'de' }])
    expect(suwayomi.updateSourcePreferences).toHaveBeenCalledWith('b', [{ position: 7, type: 'list', textValue: 'de' }])
  })

  it('does not call update for a source with no computed changes', async () => {
    const { sources, suwayomi } = makeDepsForUpdateExtensionSettings({
      a: [{ position: 0, type: 'list', key: 'lang', visible: true, textValue: 'fr' }],
      b: [{ position: 0, type: 'switch', key: 'other', visible: true, booleanValue: true }],
    })
    const settings = { common: [{ position: 0, type: 'list', key: 'lang', visible: true, textValue: 'de' }], sources: [] }

    await new UpdateExtensionSettings.UpdateExtensionSettingsUseCase(sources, suwayomi).execute({ pkgName: 'p', settings })

    expect(suwayomi.updateSourcePreferences).toHaveBeenCalledWith('a', [{ position: 0, type: 'list', textValue: 'de' }])
    expect(suwayomi.updateSourcePreferences).not.toHaveBeenCalledWith('b', expect.anything())
  })

  it('logs and continues when one source write throws, then returns re-merged truth', async () => {
    const warn = vi.spyOn(console, 'warn').mockImplementation(() => {})
    const { sources, suwayomi } = makeDepsForUpdateExtensionSettings({
      a: [{ position: 0, type: 'switch', key: 'k', visible: true, booleanValue: false }],
      b: [{ position: 0, type: 'switch', key: 'k', visible: true, booleanValue: false }],
    })
    suwayomi.updateSourcePreferences = vi.fn(async (id: string) => {
      if (id === 'a') {
        throw new Error('boom')
      }

      return []
    })
    const settings = { common: [{ position: 0, type: 'switch', key: 'k', visible: true, booleanValue: true }], sources: [] }

    const res = await new UpdateExtensionSettings.UpdateExtensionSettingsUseCase(sources, suwayomi).execute({ pkgName: 'p', settings })

    expect(warn).toHaveBeenCalled()
    expect(res.common.map(p => p.key)).toEqual(['k']) // returned re-merged state, no throw
    warn.mockRestore()
  })
})
