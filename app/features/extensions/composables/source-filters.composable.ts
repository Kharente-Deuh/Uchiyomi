// SPDX-License-Identifier: AGPL-3.0-or-later

import type { ComputedRef, Ref } from 'vue'
import type { SortStateDto, SourceFilterChangeDto, SourceFilterDto } from '#shared/dto/catalogue/source-filters.dto'
import { createExtensionsApi } from '../api/extensions.api'

// A single editable filter value. Groups/headers/separators have no value.
export type FilterDraftValue = boolean | number | string | SortStateDto

export interface UseSourceFilters {
  definitions: Ref<SourceFilterDto[]>
  draft: Ref<Record<string, FilterDraftValue>>
  filtersLoading: Ref<boolean>
  appliedChanges: Ref<SourceFilterChangeDto[]>
  activeCount: ComputedRef<number>
  apply: () => void
  reset: () => void
}

// Stable key for a filter's draft value: top-level filters key on their own
// position; a filter nested in a group keys on "<groupPosition>.<position>".
export function filterPath(position: number, groupPosition?: number): string {
  return groupPosition === undefined ? `${position}` : `${groupPosition}.${position}`
}

function defaultOf(filter: SourceFilterDto): FilterDraftValue | undefined {
  switch (filter.type) {
    case 'checkbox': return filter.default
    case 'tristate': return filter.default
    case 'select': return filter.default
    case 'text': return filter.default
    case 'sort': return filter.default ?? { ascending: false, index: 0 }
    default: return undefined // group / header / separator carry no own value
  }
}

function buildDefaults(defs: SourceFilterDto[]): Record<string, FilterDraftValue> {
  const out: Record<string, FilterDraftValue> = {}

  for (const filter of defs) {
    if (filter.type === 'group') {
      for (const child of filter.filters) {
        const value = defaultOf(child)
        if (value !== undefined) {
          out[filterPath(child.position, filter.position)] = value
        }
      }
    } else {
      const value = defaultOf(filter)
      if (value !== undefined) {
        out[filterPath(filter.position)] = value
      }
    }
  }

  return out
}

function valuesEqual(a: FilterDraftValue, b: FilterDraftValue): boolean {
  if (typeof a === 'object' || typeof b === 'object') {
    return JSON.stringify(a) === JSON.stringify(b)
  }

  return a === b
}

// Build the SourceFilterChangeDto for an editable leaf when its draft differs
// from its default; null otherwise. `kind` selects the wire state field.
function leafChange(
  filter: SourceFilterDto,
  draft: Record<string, FilterDraftValue>,
  defaults: Record<string, FilterDraftValue>,
  groupPosition?: number,
): SourceFilterChangeDto | null {
  const key = filterPath(filter.position, groupPosition)
  if (!(key in defaults)) {
    return null
  }

  const value = draft[key]
  if (value === undefined || valuesEqual(value, defaults[key]!)) {
    return null
  }

  switch (filter.type) {
    case 'checkbox': return { position: filter.position, checkBoxState: value as boolean }
    case 'tristate': return { position: filter.position, triState: value as SourceFilterChangeDto['triState'] }
    case 'select': return { position: filter.position, selectState: value as number }
    case 'text': return { position: filter.position, textState: value as string }
    case 'sort': return { position: filter.position, sortState: value as SortStateDto }
    default: return null
  }
}

export function useSourceFilters(opts: { pkgName: Ref<string | undefined>, sourceId: Ref<string | undefined>, enabled: Ref<boolean> }): UseSourceFilters {
  const api = createExtensionsApi()
  const toast = useToast()
  const { t } = useI18n()

  const cache = new Map<string, SourceFilterDto[]>()
  const definitions = ref<SourceFilterDto[]>([])
  const draft = ref<Record<string, FilterDraftValue>>({})
  const filtersLoading = ref(false)
  const appliedChanges = ref<SourceFilterChangeDto[]>([])
  let defaults: Record<string, FilterDraftValue> = {}
  let initialized = false

  function setDefinitions(defs: SourceFilterDto[]): void {
    definitions.value = defs
    defaults = buildDefaults(defs)
    if (!initialized) {
      draft.value = { ...defaults }
      appliedChanges.value = []
      initialized = true
    }
  }

  async function load(sourceId: string): Promise<void> {
    const cached = cache.get(sourceId)
    if (cached) {
      setDefinitions(cached)

      return
    }

    filtersLoading.value = true
    const res = await api.getSourceFilters(opts.pkgName.value ?? '', sourceId)
    filtersLoading.value = false

    if (!res.success) {
      toast.error(t('sources.series.filters.errors.loadFailed'))
      setDefinitions([])

      return
    }

    cache.set(sourceId, res.data)
    setDefinitions(res.data)
  }

  watch([opts.sourceId, opts.enabled], ([sourceId, enabled]) => {
    if (sourceId && enabled) {
      void load(sourceId)
    }
  }, { immediate: true })

  function reset(): void {
    draft.value = { ...defaults }
  }

  function apply(): void {
    const changes: SourceFilterChangeDto[] = []

    for (const filter of definitions.value) {
      if (filter.type === 'group') {
        for (const child of filter.filters) {
          const change = leafChange(child, draft.value, defaults, filter.position)
          if (change) {
            changes.push({ position: filter.position, groupChange: change })
          }
        }
      } else {
        const change = leafChange(filter, draft.value, defaults)
        if (change) {
          changes.push(change)
        }
      }
    }

    appliedChanges.value = changes
  }

  const activeCount = computed(() => appliedChanges.value.length)

  return { definitions, draft, filtersLoading, appliedChanges, activeCount, apply, reset }
}
