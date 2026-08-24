// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import OriginalSite from './OriginalSite.vue'

describe('asuraScansComicLinkOriginalSite', () => {
  it('opens the source URL in a new tab', async () => {
    const wrapper = await mountSuspended({
      render: () => h(OriginalSite, { to: 'https://asurascans.com/series/solo-leveling' }),
    })
    const anchor = wrapper.find('a')

    expect(anchor.attributes('href')).toBe('https://asurascans.com/series/solo-leveling')
    expect(anchor.attributes('target')).toBe('_blank')
  })
})
