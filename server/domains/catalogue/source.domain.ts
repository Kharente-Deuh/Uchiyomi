// SPDX-License-Identifier: AGPL-3.0-or-later

// Flat ESM module, consumed via `import * as Source`.
export class SourceModel {
  declare id: string
  declare name: string
  declare lang: string
  declare iconUrl: string
  declare isNsfw: boolean

  constructor(data: SourceModel) {
    Object.assign<SourceModel, SourceModel>(this, data)
  }
}
