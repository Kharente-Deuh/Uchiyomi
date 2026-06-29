// SPDX-License-Identifier: AGPL-3.0-or-later

import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it, vi } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import AuthCard from './Card.vue'

type CardProps = InstanceType<typeof AuthCard>['$props']

const base: CardProps = { title: 'Title', submitText: 'Submit', onSubmit: () => {} }

function wrap(
  props: CardProps,
  slot = 'slot-content',
): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(AuthCard, props, { default: () => slot })]) }
}

describe('authCard', () => {
  it('renders the title and subtitle', async () => {
    const wrapper = await mountSuspended(wrap({ ...base, title: 'My Title', subtitle: 'My Subtitle' }))
    expect(wrapper.text()).toContain('My Title')
    expect(wrapper.text()).toContain('My Subtitle')
  })

  it('shows the notice text in the notice alert', async () => {
    const wrapper = await mountSuspended(wrap({ ...base, notice: 'Session expired' }))
    const notice = wrapper.find('[data-test="auth-form-notice"]')
    expect(notice.exists()).toBe(true)
    expect(notice.text()).toContain('Session expired')
  })

  it('shows the error text in the error alert', async () => {
    const wrapper = await mountSuspended(wrap({ ...base, error: 'Bad credentials' }))
    const error = wrapper.find('[data-test="auth-form-error"]')
    expect(error.exists()).toBe(true)
    expect(error.text()).toContain('Bad credentials')
  })

  it('renders the default slot (the form fields)', async () => {
    const wrapper = await mountSuspended(wrap(base, 'FIELD_SLOT'))
    expect(wrapper.text()).toContain('FIELD_SLOT')
  })

  it('calls onSubmit when the form is submitted', async () => {
    const onSubmit = vi.fn()
    const wrapper = await mountSuspended(wrap({ ...base, onSubmit }))
    await wrapper.find('form').trigger('submit.prevent')
    expect(onSubmit).toHaveBeenCalledTimes(1)
  })

  it('keeps the submit button active (validation happens on submit, not via disabling)', async () => {
    const wrapper = await mountSuspended(wrap(base))
    const button = wrapper.find('button[type="submit"]')
    expect(button.exists()).toBe(true)
    expect(button.attributes('disabled')).toBeUndefined()
  })
})
