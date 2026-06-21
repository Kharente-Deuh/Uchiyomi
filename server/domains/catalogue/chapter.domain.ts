// SPDX-License-Identifier: AGPL-3.0-or-later
// Flat ESM module, consumed via `import * as Chapter`.
export type ModelProps = Model

export class Model {
  declare id: string
  declare name: string
  declare chapterNumber: number
  declare uploadDate: string
  declare isDownloaded: boolean

  constructor(data: ModelProps) {
    Object.assign<ModelProps, ModelProps>(this, data)
  }
}
