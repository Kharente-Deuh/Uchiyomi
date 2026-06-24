// SPDX-License-Identifier: AGPL-3.0-or-later
import type { ExtensionDto } from '#shared/dto/extensions'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it, vi } from 'vitest'
import { computed, ref } from 'vue'
import AdminExtensionManager from './AdminExtensionManager.vue'

const installedExt: ExtensionDto = {
  pkgName: 'eu.kanade.tachiyomi.source.mangadex',
  name: 'MangaDex',
  lang: 'en',
  iconUrl: 'https://example.com/mangadex.png',
  isNsfw: false,
  isInstalled: true,
  hasUpdate: false,
  versionName: '1.4.1',
  isHealthy: true,
}

const notInstalledExt: ExtensionDto = {
  pkgName: 'eu.kanade.tachiyomi.source.webtoons',
  name: 'Webtoons',
  lang: 'en',
  iconUrl: undefined,
  isNsfw: false,
  isInstalled: false,
  hasUpdate: false,
  versionName: '1.0.0',
  isHealthy: undefined,
}

const install = vi.fn()
const uninstall = vi.fn()

function makeExtensionsMock(): ReturnType<typeof useExtensions> {
  return {
    extensions: computed(() => [installedExt, notInstalledExt]),
    fetchLoading: ref(false),
    install,
    uninstall,
    installExtensionsLoading: ref([]),
    uninstallLoading: ref(false),
    maxPage: computed(() => 1),
    nsfwFilter: ref(undefined),
    isInstalledFilter: ref(undefined),
    isUpToDateFilter: ref(undefined),
    searchFilter: ref(undefined),
  } as unknown as ReturnType<typeof useExtensions>
}

mockNuxtImport('useExtensions', () => makeExtensionsMock)

describe('adminExtensionManager', () => {
  it('renders both extension names', async () => {
    const wrapper = await mountSuspended(AdminExtensionManager)
    const text = wrapper.text()
    expect(text).toContain('MangaDex')
    expect(text).toContain('Webtoons')
  })

  it('shows Uninstall button for installed extension and Install for non-installed', async () => {
    const wrapper = await mountSuspended(AdminExtensionManager)
    const buttons = wrapper.findAll('.v-btn')
    const buttonTexts = buttons.map(b => b.text())
    expect(buttonTexts.some(t => t.includes('Uninstall'))).toBe(true)
    expect(buttonTexts.some(t => t.includes('Install'))).toBe(true)
  })

  it('calls install for non-installed extension on click', async () => {
    install.mockClear()
    const wrapper = await mountSuspended(AdminExtensionManager)
    const buttons = wrapper.findAll('.v-btn')
    const installBtn = buttons.find(b => b.text().includes('Install'))
    expect(installBtn).toBeDefined()
    await installBtn!.trigger('click')
    expect(install).toHaveBeenCalledWith(notInstalledExt.pkgName)
  })

  it('calls uninstall for installed extension on click', async () => {
    uninstall.mockClear()
    const wrapper = await mountSuspended(AdminExtensionManager)
    const buttons = wrapper.findAll('.v-btn')
    const uninstallBtn = buttons.find(b => b.text().includes('Uninstall'))
    expect(uninstallBtn).toBeDefined()
    await uninstallBtn!.trigger('click')
    expect(uninstall).toHaveBeenCalledWith(installedExt.pkgName)
  })

  it('shows health badge for installed extension with isHealthy=true', async () => {
    const wrapper = await mountSuspended(AdminExtensionManager)
    const chips = wrapper.findAll('.v-chip')
    expect(chips.length).toBeGreaterThan(0)
    expect(chips[0]!.text()).toContain('Healthy')
  })

  it('renders four filter controls for nsfw/installed/upToDate/search', async () => {
    const wrapper = await mountSuspended(AdminExtensionManager)
    const selects = wrapper.findAll('.v-select')
    expect(selects.length).toBe(3)
    // VSelect renders inside .v-text-field too, so count those that are NOT also .v-select
    const allTextFields = wrapper.findAll('.v-text-field')
    const pureTextFields = allTextFields.filter(el => !el.classes('v-select'))
    expect(pureTextFields.length).toBe(1)
  })
})
