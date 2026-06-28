// SPDX-License-Identifier: AGPL-3.0-or-later

export type TriStateValue = 'IGNORE' | 'INCLUDE' | 'EXCLUDE'

export interface SortStateDto {
  ascending: boolean
  index: number
}

// A source's filter definitions, mirroring Suwayomi's `Filter` union. `position`
// is the filter's index in the source's filter list (the union itself carries no
// position — it is derived during mapping and used to apply changes).
export type SourceFilterDto
  = | { type: 'checkbox', position: number, name: string, default: boolean }
    | { type: 'tristate', position: number, name: string, default: TriStateValue }
    | { type: 'select', position: number, name: string, default: number, values: string[] }
    | { type: 'text', position: number, name: string, default: string }
    | { type: 'sort', position: number, name: string, default: SortStateDto | null, values: string[] }
    | { type: 'group', position: number, name: string, filters: SourceFilterDto[] }
    | { type: 'header', position: number, name: string }
    | { type: 'separator', position: number, name: string }

// An applied filter change, mirroring Suwayomi's `FilterChangeInput`. Exactly one
// state field is set (or `groupChange` for a change to a filter nested in a group).
export interface SourceFilterChangeDto {
  position: number
  checkBoxState?: boolean
  triState?: TriStateValue
  selectState?: number
  textState?: string
  sortState?: SortStateDto
  groupChange?: SourceFilterChangeDto
}
