// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { DOMWrapper, VueWrapper } from '@vue/test-utils'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h, ref } from 'vue'
import { VApp } from 'vuetify/components'
import Create from './Create.vue'

const { create } = vi.hoisted(() => ({ create: vi.fn() }))

vi.mock('~/features/oidc/composables/oidc.composable', () => ({
  useOidc: () => ({ create, loading: ref(false), testLoading: ref(false), providers: ref([]), getAll: vi.fn(), test: vi.fn() }),
}))

const ModalStub = defineComponent({
  name: 'OrganismModal',
  props: { modelValue: { type: Boolean, default: false }, isFormComplete: { type: Boolean, default: true } },
  emits: ['submit', 'cancel'],
  template: '<div><slot /><button data-test="submit" @click="$emit(\'submit\')" /><button data-test="cancel" @click="$emit(\'cancel\')" /></div>',
})

const TestBtnStub = defineComponent({
  name: 'OidcBtnTest',
  props: { issuerUrl: { type: String, default: '' } },
  template: '<div />',
})

interface Mounted {
  wrapper: VueWrapper
  show: ReturnType<typeof ref<boolean>>
}

async function mount(): Promise<Mounted> {
  const show = ref(true)
  const wrapper = await mountSuspended(
    {
      render: () => h(VApp, () => [h(Create, {
        'modelValue': show.value,
        'onUpdate:modelValue': (isShown: boolean) => {
          show.value = isShown
        },
      })]),
    },
    { global: { stubs: { OrganismModal: ModalStub, OidcBtnTest: TestBtnStub } } },
  )

  return { wrapper, show }
}

function fieldByLabel(wrapper: VueWrapper, label: string): DOMWrapper<HTMLInputElement> {
  const input = wrapper.findAll('.v-input')
    .find(i => i.find('label').exists() && i.find('label').text() === label)
    ?.find('input')

  if (!input?.exists()) {
    throw new Error(`no field labelled "${label}"`)
  }

  return input as DOMWrapper<HTMLInputElement>
}

async function addChip(wrapper: VueWrapper, label: string, value: string): Promise<void> {
  const input = fieldByLabel(wrapper, label)
  await input.setValue(value)
  await input.trigger('keydown', { key: 'Enter' })
}

async function fill(wrapper: VueWrapper): Promise<void> {
  await fieldByLabel(wrapper, 'Display name').setValue('PocketID')
  await fieldByLabel(wrapper, 'Client ID').setValue('client')
  await fieldByLabel(wrapper, 'Client secret').setValue('secret')
  await fieldByLabel(wrapper, 'Issuer URL').setValue('https://id.example.org')
  await fieldByLabel(wrapper, 'Username claim').setValue('preferred_username')
  await addChip(wrapper, 'Scopes', 'openid')
}

function isFormComplete(wrapper: VueWrapper): boolean {
  return wrapper.findComponent(ModalStub).props('isFormComplete') === true
}

beforeEach(() => {
  create.mockReset()
  create.mockResolvedValue(undefined)
})

describe('oidcModalCreate', () => {
  it('starts with every field empty', async () => {
    const { wrapper } = await mount()

    expect(wrapper.findAll('input[type="text"], input[type="password"]')
      .every(i => (i.element as HTMLInputElement).value === '')).toBe(true)
  })

  it('holds the submit button back while the form is incomplete', async () => {
    const { wrapper } = await mount()
    expect(isFormComplete(wrapper)).toBe(false)
  })

  it('releases the submit button once every required field is filled', async () => {
    const { wrapper } = await mount()

    await fill(wrapper)

    await vi.waitFor(() => expect(isFormComplete(wrapper)).toBe(true))
  })

  it('stays incomplete while the issuer url is not a url', async () => {
    const { wrapper } = await mount()

    await fill(wrapper)
    await fieldByLabel(wrapper, 'Issuer URL').setValue('not-a-url')

    await vi.waitFor(() => expect(isFormComplete(wrapper)).toBe(false))
  })

  it('stays incomplete without a scope', async () => {
    const { wrapper } = await mount()

    await fieldByLabel(wrapper, 'Display name').setValue('PocketID')
    await fieldByLabel(wrapper, 'Client ID').setValue('client')
    await fieldByLabel(wrapper, 'Client secret').setValue('secret')
    await fieldByLabel(wrapper, 'Issuer URL').setValue('https://id.example.org')
    await fieldByLabel(wrapper, 'Username claim').setValue('preferred_username')

    await vi.waitFor(() => expect(isFormComplete(wrapper)).toBe(false))
  })

  it('creates the provider from the filled values', async () => {
    const { wrapper } = await mount()

    await fill(wrapper)
    await vi.waitFor(() => expect(isFormComplete(wrapper)).toBe(true))
    await wrapper.find('[data-test="submit"]').trigger('click')

    await vi.waitFor(() => expect(create).toHaveBeenCalledWith(expect.objectContaining({
      displayName: 'PocketID',
      clientId: 'client',
      clientSecret: 'secret',
      issuerUrl: 'https://id.example.org',
      usernameClaim: 'preferred_username',
      scopes: ['openid'],
      autoProvision: false,
    })))
  })

  it('closes itself once the provider is created', async () => {
    const { wrapper, show } = await mount()

    await fill(wrapper)
    await vi.waitFor(() => expect(isFormComplete(wrapper)).toBe(true))
    await wrapper.find('[data-test="submit"]').trigger('click')

    await vi.waitFor(() => expect(show.value).toBe(false))
  })

  it('does not create while the form is incomplete', async () => {
    const { wrapper } = await mount()

    await wrapper.find('[data-test="submit"]').trigger('click')

    expect(create).not.toHaveBeenCalled()
  })

  it('disables the role value lists until a role claim is set', async () => {
    const { wrapper } = await mount()

    expect(fieldByLabel(wrapper, 'Admin role values').attributes('disabled')).toBeDefined()

    await fieldByLabel(wrapper, 'Role claim').setValue('groups')

    await vi.waitFor(() => expect(fieldByLabel(wrapper, 'Admin role values').attributes('disabled')).toBeUndefined())
  })

  it('clears both role value lists when the role claim is emptied', async () => {
    const { wrapper } = await mount()

    await fieldByLabel(wrapper, 'Role claim').setValue('groups')
    await addChip(wrapper, 'Admin role values', 'admins')
    await addChip(wrapper, 'Regular user role values', 'users')
    await vi.waitFor(() => {
      expect(wrapper.text()).toContain('admins')
      expect(wrapper.text()).toContain('users')
    })

    await fieldByLabel(wrapper, 'Role claim').setValue('')

    await vi.waitFor(() => {
      expect(wrapper.text()).not.toContain('admins')
      expect(wrapper.text()).not.toContain('users')
    })
  })

  it('closes without creating on cancel', async () => {
    const { wrapper, show } = await mount()

    await fill(wrapper)
    await wrapper.find('[data-test="cancel"]').trigger('click')

    expect(show.value).toBe(false)
    expect(create).not.toHaveBeenCalled()
  })

  it('resets the form when reopened', async () => {
    const { wrapper, show } = await mount()

    await fieldByLabel(wrapper, 'Display name').setValue('PocketID')
    show.value = false
    await vi.waitFor(() => expect(fieldByLabel(wrapper, 'Display name').element.value).toBe('PocketID'))

    show.value = true

    await vi.waitFor(() => expect(fieldByLabel(wrapper, 'Display name').element.value).toBe(''))
  })
})
