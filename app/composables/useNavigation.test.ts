// SPDX-License-Identifier: AGPL-3.0-or-later
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'

const { useNavigation } = await import('~/composables/useNavigation')

describe('useNavigation', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('exposes the home and settings destinations in order', () => {
    const { items } = useNavigation()
    expect(items.value.map(i => i.key)).toEqual(['home', 'settings'])
  })

  it('maps each destination to its route and i18n label key', () => {
    const { items } = useNavigation()
    expect(items.value).toEqual([
      { key: 'home', to: '/', icon: 'fa6-solid:house', labelKey: 'nav.home' },
      { key: 'settings', to: '/settings', icon: 'fa6-solid:gear', labelKey: 'nav.settings' },
    ])
  })
})
