// SPDX-License-Identifier: AGPL-3.0-or-later

import type { SourceDto } from '#shared/dto/extensions'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import DesktopList from './index.vue'
import DesktopItem from './Item.vue'

function makeSource(id: string, isEnabled = false): SourceDto {
  return {
    id,
    name: `Source ${id}`,
    lang: 'en',
    isNsfw: false,
    isConfigurable: false,
    isEnabled,
    supportsLatest: true,
  }
}

const sources: SourceDto[] = [makeSource('s1'), makeSource('s2', true), makeSource('s3')]

function wrap(
  props: Partial<InstanceType<typeof DesktopList>['$props']> & { sources: SourceDto[] },
): { render: () => ReturnType<typeof h> } {
  return {
    render: () => h(VApp, () => [
      h(DesktopList, {
        'sources': props.sources,
        'canManageExtensions': props.canManageExtensions ?? false,
        'sourceToggleLoading': props.sourceToggleLoading ?? new Set(),
        'enabledSourcesCount': props.enabledSourcesCount ?? 0,
        'totalSourcesCount': props.totalSourcesCount ?? props.sources.length,
        'hasSettings': props.hasSettings ?? false,
        'pkgName': props.pkgName ?? 'com.example.pkg',
        'showOnlyEnabledSources': props.showOnlyEnabledSources,
        'onUpdate:showOnlyEnabledSources': () => {},
        'onToggle': () => {},
      }),
    ]),
  }
}

describe('extensionsSourceListDesktop', () => {
  it('renders one item per source', async () => {
    const wrapper = await mountSuspended(wrap({ sources }))
    expect(wrapper.findAllComponents(DesktopItem)).toHaveLength(3)
  })

  it('passes the source prop through to each item', async () => {
    const wrapper = await mountSuspended(wrap({ sources }))
    const items = wrapper.findAllComponents(DesktopItem)
    expect(items[0]!.props('source')).toEqual(sources[0])
    expect(items[1]!.props('source')).toEqual(sources[1])
  })

  it('shows the enabled sources count summary text', async () => {
    const wrapper = await mountSuspended(
      wrap({ sources, enabledSourcesCount: 1, totalSourcesCount: 3 }),
    )
    // The text includes the i18n key interpolation; just verify both numbers appear
    expect(wrapper.text()).toMatch(/1/)
    expect(wrapper.text()).toMatch(/3/)
  })

  it('shows the filter button when canManageExtensions is true and there is more than one source', async () => {
    const wrapper = await mountSuspended(wrap({ sources, canManageExtensions: true }))
    // EnabledFilterBtn is rendered; the button with the filter icon should exist
    expect(wrapper.find('button').exists()).toBe(true)
  })

  it('emits toggle with the source id when a child item emits toggle', async () => {
    const emitted: string[] = []
    const wrapper = await mountSuspended({
      render: () => h(VApp, () => [
        h(DesktopList, {
          sources,
          'canManageExtensions': false,
          'sourceToggleLoading': new Set<string>(),
          'enabledSourcesCount': 0,
          'totalSourcesCount': sources.length,
          'hasSettings': false,
          'pkgName': 'com.example.pkg',
          'showOnlyEnabledSources': undefined,
          'onUpdate:showOnlyEnabledSources': () => {},
          'onToggle': (id: string) => emitted.push(id),
        }),
      ]),
    })
    const items = wrapper.findAllComponents(DesktopItem)
    await items[0]!.find('.source-card').trigger('click')
    expect(emitted).toContain('s1')
  })
})
