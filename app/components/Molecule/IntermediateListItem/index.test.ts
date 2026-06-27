// SPDX-License-Identifier: AGPL-3.0-or-later
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { ref } from 'vue'
import Desktop from './Desktop.vue'
import IntermediateListItem from './index.vue'
import Mobile from './Mobile.vue'

// We test the display-branching logic by providing two useDisplay mocks.
// mockNuxtImport requires the factory to be stable per file, so we use a
// module-level mutable ref that both mocks share.
const mobileRef = ref(false)

function useDisplayStub(): { mobile: typeof mobileRef } {
  return { mobile: mobileRef }
}

mockNuxtImport('useDisplay', () => useDisplayStub)

describe('moleculeIntermediateListItem', () => {
  it('renders the Desktop sub-component on large screens', async () => {
    mobileRef.value = false
    const wrapper = await mountSuspended(IntermediateListItem, {
      props: { title: 'Profile', to: '/settings/profile' },
    })
    expect(wrapper.findComponent(Desktop).exists()).toBe(true)
    expect(wrapper.findComponent(Mobile).exists()).toBe(false)
  })

  it('renders the Mobile sub-component on small screens', async () => {
    mobileRef.value = true
    const wrapper = await mountSuspended(IntermediateListItem, {
      props: { title: 'Profile', to: '/settings/profile' },
    })
    expect(wrapper.findComponent(Mobile).exists()).toBe(true)
    expect(wrapper.findComponent(Desktop).exists()).toBe(false)
  })

  it('forwards the title prop to the active sub-component (desktop)', async () => {
    mobileRef.value = false
    const wrapper = await mountSuspended(IntermediateListItem, {
      props: { title: 'Account', to: '/settings/account' },
    })
    expect(wrapper.text()).toContain('Account')
  })
})
