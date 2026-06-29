// SPDX-License-Identifier: AGPL-3.0-or-later

import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

// useDisplay mock — mobile is mutable so individual tests can flip it.
const { displayStub } = vi.hoisted(() => ({
  displayStub: { mobile: { value: false } as { value: boolean } },
}))

function useDisplayStub(): typeof displayStub {
  return displayStub
}

mockNuxtImport('useDisplay', () => useDisplayStub)

const { useLayoutPadding } = await import('~/composables/layoutPadding.composable')
const { useLayoutStore } = await import('~/store/layout.store')

describe('useLayoutPadding', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    displayStub.mobile.value = false
  })

  describe('bottomLayout', () => {
    it('returns the bottom-navigation CSS variable on mobile', () => {
      displayStub.mobile.value = true
      const { bottomLayout } = useLayoutPadding()
      expect(bottomLayout.value).toBe('var(--bottom-navigation-height)')
    })

    it('returns the bottom-pagination CSS variable on desktop when pagination is enabled', () => {
      displayStub.mobile.value = false
      const store = useLayoutStore()
      store.setPaginationEnabled(true)
      const { bottomLayout } = useLayoutPadding()
      expect(bottomLayout.value).toBe('var(--bottom-pagination-height)')
    })

    it('returns undefined on desktop when pagination is disabled', () => {
      displayStub.mobile.value = false
      const store = useLayoutStore()
      store.setPaginationEnabled(false)
      const { bottomLayout } = useLayoutPadding()
      expect(bottomLayout.value).toBeUndefined()
    })

    it('mobile takes priority over pagination: returns bottom-navigation even when pagination is enabled', () => {
      displayStub.mobile.value = true
      const store = useLayoutStore()
      store.setPaginationEnabled(true)
      const { bottomLayout } = useLayoutPadding()
      expect(bottomLayout.value).toBe('var(--bottom-navigation-height)')
    })
  })

  describe('leftLayout', () => {
    it('returns undefined on mobile', () => {
      displayStub.mobile.value = true
      const { leftLayout } = useLayoutPadding()
      expect(leftLayout.value).toBeUndefined()
    })

    it('returns the compact-width CSS variable on desktop when drawer is compact', () => {
      displayStub.mobile.value = false
      const store = useLayoutStore()
      store.setNavigationDrawerCompact(true)
      const { leftLayout } = useLayoutPadding()
      expect(leftLayout.value).toBe('var(--navigation-drawer-compact-width)')
    })

    it('returns the full-width CSS variable on desktop when drawer is not compact', () => {
      displayStub.mobile.value = false
      const store = useLayoutStore()
      store.setNavigationDrawerCompact(false)
      const { leftLayout } = useLayoutPadding()
      expect(leftLayout.value).toBe('var(--navigation-drawer-width)')
    })
  })
})
