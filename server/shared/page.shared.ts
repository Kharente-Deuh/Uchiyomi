// `shared/` is imported with a relative path here so the Vitest `node`
// project (which has no path aliases) can resolve it. See CLAUDE.md.
import type { PageDto } from '../../shared/dto/page.dto'

export interface Page<T> {
  items: T[]
  total: number
}

export function toPageDto<T, U>(page: Page<U>, mapFn?: (item: U) => T): PageDto<T> {
  return {
    items: mapFn
      ? page.items.map(item => mapFn(item))
      : page.items as unknown as T[],
    total: page.total,
  }
}
