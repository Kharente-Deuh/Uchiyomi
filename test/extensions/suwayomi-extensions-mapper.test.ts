// SPDX-License-Identifier: AGPL-3.0-or-later
import { describe, expect, it } from 'vitest'
import { preferenceToDomain, sourceToDomain, toChangeInput } from '../../server/domains/extensions/infrastructure/transport/graphql/graphql-suwayomi-extensions.mapper'

describe('sourceToDomain', () => {
  it('maps supportsLatest: true', () => {
    const domain = sourceToDomain({ id: 's1', name: 'Source 1', lang: 'en', isNsfw: false, isConfigurable: false, supportsLatest: true })
    expect(domain).toMatchObject({ id: 's1', name: 'Source 1', lang: 'en', isNsfw: false, isConfigurable: false, supportsLatest: true })
  })

  it('maps supportsLatest: false', () => {
    const domain = sourceToDomain({ id: 's2', name: 'Source 2', lang: 'fr', isNsfw: true, isConfigurable: true, supportsLatest: false })
    expect(domain).toMatchObject({ id: 's2', supportsLatest: false })
  })
})

describe('preferenceToDomain', () => {
  it('maps a SwitchPreference', () => {
    const p = preferenceToDomain(
      { __typename: 'SwitchPreference', key: 'k', title: 't', summary: 's', currentValueBool: true, defaultBool: false, visible: true },
      0,
    )
    expect(p).toMatchObject({ position: 0, type: 'switch', booleanValue: true, booleanDefault: false })
  })

  it('maps a CheckBoxPreference with null currentValue', () => {
    const p = preferenceToDomain(
      { __typename: 'CheckBoxPreference', key: 'cb', title: 'CB', summary: null, currentValueBool: null, defaultBool: true, visible: true },
      1,
    )
    expect(p).toMatchObject({ position: 1, type: 'checkbox', booleanValue: undefined, booleanDefault: true })
  })

  it('maps an EditTextPreference', () => {
    const p = preferenceToDomain(
      { __typename: 'EditTextPreference', key: 'et', title: 'ET', summary: null, visible: true, dialogTitle: 'DT', dialogMessage: 'DM', currentValueStr: 'hello', defaultStr: 'world' },
      1,
    )
    expect(p).toMatchObject({ position: 1, type: 'editText', textValue: 'hello', textDefault: 'world', dialogTitle: 'DT', dialogMessage: 'DM' })
  })

  it('maps a ListPreference with entries', () => {
    const p = preferenceToDomain(
      { __typename: 'ListPreference', key: 'k', title: 't', summary: null, visible: true, entries: ['A', 'B'], entryValues: ['a', 'b'], currentValueStr: 'b', defaultStr: 'a' },
      2,
    )
    expect(p).toMatchObject({ position: 2, type: 'list', textValue: 'b', textDefault: 'a', entries: ['A', 'B'], entryValues: ['a', 'b'] })
  })

  it('maps a MultiSelectListPreference', () => {
    const p = preferenceToDomain(
      { __typename: 'MultiSelectListPreference', key: 'm', title: 'MS', summary: null, visible: false, entries: ['X', 'Y'], entryValues: ['x', 'y'], currentValueList: ['x'], defaultList: ['y'] },
      3,
    )
    expect(p).toMatchObject({ position: 3, type: 'multiSelect', multiValue: ['x'], multiDefault: ['y'], entries: ['X', 'Y'], entryValues: ['x', 'y'] })
  })

  it('maps null optional string fields to undefined', () => {
    const p = preferenceToDomain(
      { __typename: 'SwitchPreference', key: null, title: null, summary: null, currentValueBool: false, defaultBool: true, visible: false },
      0,
    )
    expect(p.key).toBeUndefined()
    expect(p.title).toBeUndefined()
    expect(p.summary).toBeUndefined()
  })

  it('throws on unknown __typename', () => {
    expect(() => preferenceToDomain(
      { __typename: 'UnknownPreference', key: null, title: null, summary: null, visible: true } as never,
      0,
    ),
    ).toThrow('Unknown preference typename')
  })
})

describe('toChangeInput', () => {
  it('builds switchState for a switch', () => {
    expect(toChangeInput({ position: 0, type: 'switch', booleanValue: true })).toEqual({ position: 0, switchState: true })
  })

  it('builds checkBoxState for a checkbox', () => {
    expect(toChangeInput({ position: 1, type: 'checkbox', booleanValue: false })).toEqual({ position: 1, checkBoxState: false })
  })

  it('builds editTextState for editText', () => {
    expect(toChangeInput({ position: 0, type: 'editText', textValue: 'hello' })).toEqual({ position: 0, editTextState: 'hello' })
  })

  it('builds listState for a list', () => {
    expect(toChangeInput({ position: 1, type: 'list', textValue: 'b' })).toEqual({ position: 1, listState: 'b' })
  })

  it('builds multiSelectState for a multiSelect', () => {
    expect(toChangeInput({ position: 2, type: 'multiSelect', multiValue: ['a', 'b'] })).toEqual({ position: 2, multiSelectState: ['a', 'b'] })
  })
})
