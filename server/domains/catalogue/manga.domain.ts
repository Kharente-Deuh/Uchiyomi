// SPDX-License-Identifier: AGPL-3.0-or-later
// Flat ESM module, consumed via `import * as Manga`.

import type { ChapterModel } from './chapter.domain'

export class MangaSummaryModel {
  declare id: string
  declare title: string
  declare thumbnailUrl?: string
  declare inLibrary: boolean

  constructor(data: MangaSummaryModel) {
    Object.assign<MangaSummaryModel, MangaSummaryModel>(this, data)
  }
}

export class MangaDetailsModel extends MangaSummaryModel {
  declare author?: string
  declare description?: string
  declare status: string
  declare chapters: ChapterModel[]

  constructor(data: MangaDetailsModel) {
    super(data)
    Object.assign<MangaDetailsModel, MangaDetailsModel>(this, data)
  }
}
