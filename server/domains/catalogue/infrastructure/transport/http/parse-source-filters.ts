// SPDX-License-Identifier: AGPL-3.0-or-later

import type { SourceFilterChange } from '../../../catalogue.domain'
import { z } from 'zod'

const triStateSchema = z.enum(['IGNORE', 'INCLUDE', 'EXCLUDE'])
const sortStateSchema = z.object({ ascending: z.boolean(), index: z.number().int() }).strict()

// Recursive: groupChange nests one level (Tachiyomi groups contain leaf filters only).
export const sourceFilterChangeSchema: z.ZodType<SourceFilterChange> = z.lazy(() => z.object({
  position: z.number().int().min(0),
  checkBoxState: z.boolean().optional(),
  triState: triStateSchema.optional(),
  selectState: z.number().int().optional(),
  textState: z.string().optional(),
  sortState: sortStateSchema.optional(),
  groupChange: sourceFilterChangeSchema.optional(),
}).strict())

export const sourceFiltersSchema = z.array(sourceFilterChangeSchema)

// Parse the JSON-encoded `filters` query param into validated domain changes.
// Returns [] when absent. Returns null when the param is present but invalid
// (malformed JSON or wrong shape) — the route maps null to a 400.
export function parseSourceFiltersParam(raw: string | undefined): SourceFilterChange[] | null {
  if (raw === undefined || raw === '') {
    return []
  }

  let json: unknown
  try {
    json = JSON.parse(raw)
  } catch {
    return null
  }

  const parsed = sourceFiltersSchema.safeParse(json)

  return parsed.success ? parsed.data : null
}

// SEARCH needs at least a query OR one filter; popular/latest never require a query.
export function searchQueryMissing(type: string, query: string, filters: SourceFilterChange[]): boolean {
  return type === 'search' && query.length === 0 && filters.length === 0
}

// Filters apply only to the SEARCH browse type; sending them with popular/latest is a client error.
export function filtersRequireSearch(type: string, filters: SourceFilterChange[]): boolean {
  return type !== 'search' && filters.length > 0
}
