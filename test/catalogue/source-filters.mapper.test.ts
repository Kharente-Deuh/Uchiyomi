// SPDX-License-Identifier: AGPL-3.0-or-later

// @vitest-environment node
import { describe, expect, it } from 'vitest'
import { filterChangeToInput, sourceFiltersToDomain } from '../../server/domains/catalogue/infrastructure/transport/graphql/graphql-suwayomi-catalogue-repository.mapper'

describe('sourceFiltersToDomain', () => {
  it('maps each union member, deriving position from the array index', () => {
    const result = sourceFiltersToDomain([
      { __typename: 'CheckBoxFilter', name: 'Completed', checkBoxDefault: false },
      { __typename: 'TriStateFilter', name: 'Action', triDefault: 'IGNORE' },
      { __typename: 'SelectFilter', name: 'Order', selectDefault: 1, values: ['ASC', 'DESC'] },
      { __typename: 'TextFilter', name: 'Author', textDefault: '' },
      { __typename: 'SortFilter', name: 'Sort', sortDefault: { ascending: true, index: 0 }, values: ['Title'] },
      { __typename: 'HeaderFilter', name: 'Section' },
      { __typename: 'SeparatorFilter', name: '' },
    ])

    expect(result).toEqual([
      { type: 'checkbox', position: 0, name: 'Completed', default: false },
      { type: 'tristate', position: 1, name: 'Action', default: 'IGNORE' },
      { type: 'select', position: 2, name: 'Order', default: 1, values: ['ASC', 'DESC'] },
      { type: 'text', position: 3, name: 'Author', default: '' },
      { type: 'sort', position: 4, name: 'Sort', default: { ascending: true, index: 0 }, values: ['Title'] },
      { type: 'header', position: 5, name: 'Section' },
      { type: 'separator', position: 6, name: '' },
    ])
  })

  it('maps a group with its child filters indexed within the group', () => {
    const result = sourceFiltersToDomain([
      {
        __typename: 'GroupFilter',
        name: 'Genres',
        filters: [
          { __typename: 'TriStateFilter', name: 'Comedy', triDefault: 'IGNORE' },
          { __typename: 'TriStateFilter', name: 'Drama', triDefault: 'INCLUDE' },
        ],
      },
    ])

    expect(result).toEqual([
      {
        type: 'group',
        position: 0,
        name: 'Genres',
        filters: [
          { type: 'tristate', position: 0, name: 'Comedy', default: 'IGNORE' },
          { type: 'tristate', position: 1, name: 'Drama', default: 'INCLUDE' },
        ],
      },
    ])
  })
})

describe('filterChangeToInput', () => {
  it('maps each leaf change to a FilterChangeInput with only its state field', () => {
    expect(filterChangeToInput({ position: 0, checkBoxState: true })).toEqual({ position: 0, checkBoxState: true })
    expect(filterChangeToInput({ position: 1, triState: 'EXCLUDE' })).toEqual({ position: 1, triState: 'EXCLUDE' })
    expect(filterChangeToInput({ position: 2, selectState: 3 })).toEqual({ position: 2, selectState: 3 })
    expect(filterChangeToInput({ position: 3, textState: 'naruto' })).toEqual({ position: 3, textState: 'naruto' })
    expect(filterChangeToInput({ position: 4, sortState: { ascending: false, index: 2 } }))
      .toEqual({ position: 4, sortState: { ascending: false, index: 2 } })
  })

  it('recurses into a group change', () => {
    expect(filterChangeToInput({ position: 5, groupChange: { position: 1, triState: 'INCLUDE' } }))
      .toEqual({ position: 5, groupChange: { position: 1, triState: 'INCLUDE' } })
  })
})
