// SPDX-License-Identifier: AGPL-3.0-or-later

import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import { useLayoutStore } from '~/store/layout.store'
import PaginationFooter from './PaginationFooter.vue'

type FooterProps = InstanceType<typeof PaginationFooter>['$props']

function wrap(props: FooterProps): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(PaginationFooter, props)]) }
}

describe('moleculePaginationFooter', () => {
  it('renders the current page and the total', async () => {
    const wrapper = await mountSuspended(wrap({ modelValue: 2, pagesTotal: 5 }))
    expect(wrapper.text()).toContain('2 / 5')
  })

  it('disables the previous button on the first page', async () => {
    const wrapper = await mountSuspended(wrap({ modelValue: 1, pagesTotal: 5 }))
    const [prev, next] = wrapper.findAll('button')
    expect((prev!.element as HTMLButtonElement).disabled).toBe(true)
    expect((next!.element as HTMLButtonElement).disabled).toBe(false)
  })

  it('disables the next button on the last page', async () => {
    const wrapper = await mountSuspended(wrap({ modelValue: 5, pagesTotal: 5 }))
    const [prev, next] = wrapper.findAll('button')
    expect((prev!.element as HTMLButtonElement).disabled).toBe(false)
    expect((next!.element as HTMLButtonElement).disabled).toBe(true)
  })

  it('disables both buttons when disabled', async () => {
    const wrapper = await mountSuspended(wrap({ modelValue: 2, pagesTotal: 5, disabled: true }))
    const [prev, next] = wrapper.findAll('button')
    expect((prev!.element as HTMLButtonElement).disabled).toBe(true)
    expect((next!.element as HTMLButtonElement).disabled).toBe(true)
  })

  it('decrements the page when clicking previous', async () => {
    const wrapper = await mountSuspended(wrap({ modelValue: 3, pagesTotal: 5 }))
    await wrapper.findAll('button')[0]!.trigger('click')
    expect(wrapper.findComponent(PaginationFooter).emitted('update:modelValue')![0]).toEqual([2])
  })

  it('increments the page when clicking next', async () => {
    const wrapper = await mountSuspended(wrap({ modelValue: 3, pagesTotal: 5 }))
    await wrapper.findAll('button')[1]!.trigger('click')
    expect(wrapper.findComponent(PaginationFooter).emitted('update:modelValue')![0]).toEqual([4])
  })

  it('enables pagination in the layout store while mounted and disables it on unmount', async () => {
    const wrapper = await mountSuspended(wrap({ modelValue: 1, pagesTotal: 5 }))
    const store = useLayoutStore()
    expect(store.paginationEnabled).toBe(true)
    wrapper.unmount()
    expect(store.paginationEnabled).toBe(false)
  })
})
