// SPDX-License-Identifier: AGPL-3.0-or-later
import type { PreferenceDto } from '#shared/dto/extensions'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it, vi } from 'vitest'
import SourcePreferenceForm from './SourcePreferenceForm.vue'

const switchPref: PreferenceDto = {
  position: 0,
  type: 'switch',
  title: 'Enable feature',
  summary: 'Toggle this on',
  visible: true,
  booleanValue: false,
  booleanDefault: false,
}

const editTextPref: PreferenceDto = {
  position: 1,
  type: 'editText',
  title: 'API key',
  summary: 'Enter your key',
  visible: true,
  textValue: '',
  textDefault: '',
}

const listPref: PreferenceDto = {
  position: 2,
  type: 'list',
  title: 'Language',
  summary: 'Pick a language',
  visible: true,
  entries: ['English', 'French'],
  entryValues: ['en', 'fr'],
  textValue: 'en',
}

const checkboxPref: PreferenceDto = {
  position: 3,
  type: 'checkbox',
  title: 'Show NSFW',
  visible: true,
  booleanDefault: false,
}

const multiSelectPref: PreferenceDto = {
  position: 4,
  type: 'multiSelect',
  title: 'Tags',
  visible: true,
  entries: ['Action', 'Comedy'],
  entryValues: ['action', 'comedy'],
  multiValue: [],
}

const updatedPreferences: PreferenceDto[] = [{ ...switchPref, booleanValue: true }]

const getPreferences = vi.fn().mockResolvedValue({ success: true, data: [switchPref, editTextPref, listPref, checkboxPref, multiSelectPref] })
const updatePreference = vi.fn().mockResolvedValue({ success: true, data: updatedPreferences })

vi.mock('~/features/extensions/api/extensions.api', () => ({
  createExtensionsApi: () => ({
    getPreferences,
    updatePreference,
    listExtensions: vi.fn(),
    getExtension: vi.fn(),
    extensionAction: vi.fn(),
    listSources: vi.fn(),
    setSourceEnabled: vi.fn(),
  }),
}))

describe('sourcePreferenceForm', () => {
  it('renders a VSwitch for switch preference', async () => {
    const wrapper = await mountSuspended(SourcePreferenceForm, {
      props: { sourceId: 'src-1' },
    })
    expect(wrapper.find('.v-switch').exists()).toBe(true)
  })

  it('renders a VTextField for editText preference', async () => {
    const wrapper = await mountSuspended(SourcePreferenceForm, {
      props: { sourceId: 'src-1' },
    })
    expect(wrapper.find('.v-text-field').exists()).toBe(true)
  })

  it('renders a VSelect for list preference', async () => {
    const wrapper = await mountSuspended(SourcePreferenceForm, {
      props: { sourceId: 'src-1' },
    })
    expect(wrapper.find('.v-select').exists()).toBe(true)
  })

  it('calls updatePreference with position and booleanValue when switch changes', async () => {
    updatePreference.mockClear()
    const wrapper = await mountSuspended(SourcePreferenceForm, {
      props: { sourceId: 'src-1' },
    })
    const vswitch = wrapper.findComponent({ name: 'VSwitch' })
    expect(vswitch.exists()).toBe(true)
    await vswitch.vm.$emit('update:modelValue', true)
    expect(updatePreference).toHaveBeenCalledWith('src-1', { position: 0, booleanValue: true })
  })

  it('replaces preferences list with the response on successful update', async () => {
    updatePreference.mockClear()
    const wrapper = await mountSuspended(SourcePreferenceForm, {
      props: { sourceId: 'src-1' },
    })
    const vswitch = wrapper.findComponent({ name: 'VSwitch' })
    await vswitch.vm.$emit('update:modelValue', true)
    await wrapper.vm.$nextTick()
    // After update, preferences replaced — only 1 switch from updatedPreferences
    const switches = wrapper.findAllComponents({ name: 'VSwitch' })
    expect(switches).toHaveLength(1)
  })
})
