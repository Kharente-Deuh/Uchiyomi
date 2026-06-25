// SPDX-License-Identifier: AGPL-3.0-or-later
// Flat ESM module, consumed via `import * as Chapter`.
export class ChapterModel {
  declare id: string
  declare name: string
  declare chapterNumber: number
  declare uploadDate: string
  declare isDownloaded: boolean

  constructor(data: ChapterModel) {
    Object.assign<ChapterModel, ChapterModel>(this, data)
  }
}
