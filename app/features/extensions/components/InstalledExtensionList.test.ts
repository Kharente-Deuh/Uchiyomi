// SPDX-License-Identifier: AGPL-3.0-or-later
import type { ExtensionDto } from '#shared/dto/extensions'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it, vi } from 'vitest'
import { computed, ref } from 'vue'
import InstalledExtensionList from './InstalledExtensionList.vue'

const installedItems: ExtensionDto[] = [
  {
    pkgName: 'eu.kanade.tachiyomi.source.mangadex',
    name: 'MangaDex',
    lang: 'en',
    iconUrl: 'https://example.com/mangadex.png',
    isNsfw: false,
    isInstalled: true,
    hasUpdate: false,
    versionName: '1.4.1',
    isHealthy: true,
  },
  {
    pkgName: 'eu.kanade.tachiyomi.source.nhentai',
    name: 'NHentai',
    lang: 'en',
    iconUrl: undefined,
    isNsfw: true,
    isInstalled: true,
    hasUpdate: false,
    versionName: '1.0.0',
    isHealthy: false,
  },
]

function makeExtensionsMock(): ReturnType<typeof useExtensions> {
  return {
    extensions: computed(() => installedItems),
    fetchLoading: ref(false),
    install: vi.fn(),
    uninstall: vi.fn(),
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

describe('installedExtensionList', () => {
  it('renders the name of each installed extension', async () => {
    const wrapper = await mountSuspended(InstalledExtensionList)
    const text = wrapper.text()
    expect(text).toContain('MangaDex')
    expect(text).toContain('NHentai')
  })

  it('renders one health badge per extension', async () => {
    const wrapper = await mountSuspended(InstalledExtensionList)
    const chips = wrapper.findAll('.v-chip')
    expect(chips).toHaveLength(2)
  })

  it('shows healthy badge for extension with isHealthy=true', async () => {
    const wrapper = await mountSuspended(InstalledExtensionList)
    const chips = wrapper.findAll('.v-chip')
    expect(chips[0]!.text()).toContain('Healthy')
  })

  it('shows error badge for extension with isHealthy=false', async () => {
    const wrapper = await mountSuspended(InstalledExtensionList)
    const chips = wrapper.findAll('.v-chip')
    expect(chips[1]!.text()).toContain('Error')
  })
})
