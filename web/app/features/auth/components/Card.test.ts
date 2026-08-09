// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it, vi } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import Card from './Card.vue'

type Props = InstanceType<typeof Card>['$props']

const DEFAULT_SLOTS = { default: () => 'fields' }

function wrap(
  props: Props,
  slots: Record<string, () => unknown> = DEFAULT_SLOTS,
): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(Card, props, slots)]) }
}

const base = { title: 'Sign in', submitText: 'Log in', onSubmit: () => {} }

describe('authCard', () => {
  it('renders the title and the submit label', async () => {
    const wrapper = await mountSuspended(wrap(base))
    expect(wrapper.text()).toContain('Sign in')
    expect(wrapper.text()).toContain('Log in')
  })

  it('renders the subtitle when given one', async () => {
    const wrapper = await mountSuspended(wrap({ ...base, subtitle: 'with your account' }))
    expect(wrapper.text()).toContain('with your account')
  })

  it('renders the field slot', async () => {
    const wrapper = await mountSuspended(wrap(base))
    expect(wrapper.text()).toContain('fields')
  })

  it('renders the footer slot', async () => {
    const wrapper = await mountSuspended(wrap(base, { default: () => 'fields', footer: () => 'forgot?' }))
    expect(wrapper.text()).toContain('forgot?')
  })

  it('shows no error alert by default', async () => {
    const wrapper = await mountSuspended(wrap(base))
    expect(wrapper.find('[data-test="auth-form-error"]').exists()).toBe(false)
  })

  it('shows the error alert when an error is given', async () => {
    const wrapper = await mountSuspended(wrap({ ...base, error: 'Invalid credentials' }))
    expect(wrapper.find('[data-test="auth-form-error"]').text()).toContain('Invalid credentials')
  })

  it('calls onSubmit when the form is submitted', async () => {
    const onSubmit = vi.fn()
    const wrapper = await mountSuspended(wrap({ ...base, onSubmit }))

    await wrapper.find('form').trigger('submit')

    expect(onSubmit).toHaveBeenCalledOnce()
  })

  it('marks the submit button loading', async () => {
    const wrapper = await mountSuspended(wrap({ ...base, loading: true }))
    expect(wrapper.find('.v-btn--loading').exists()).toBe(true)
  })
})
