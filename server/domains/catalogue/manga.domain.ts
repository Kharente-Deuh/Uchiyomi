// SPDX-License-Identifier: AGPL-3.0-or-later
// Flat ESM module, consumed via `import * as Manga`.
import type * as Chapter from './chapter.domain'

export type SummaryProps = Summary

export class Summary {
  declare id: string
  declare title: string
  declare thumbnailUrl?: string
  declare inLibrary: boolean

  constructor(data: SummaryProps) {
    Object.assign<SummaryProps, SummaryProps>(this, data)
  }
}

export type DetailsProps = Details

export class Details extends Summary {
  declare author?: string
  declare description?: string
  declare status: string
  declare chapters: Chapter.Model[]

  constructor(data: DetailsProps) {
    super(data)
    Object.assign<DetailsProps, DetailsProps>(this, data)
  }
}
