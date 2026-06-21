// SPDX-License-Identifier: AGPL-3.0-or-later
// Flat ESM module, consumed via `import * as Source`.
export type ModelProps = Model

export class Model {
  declare id: string
  declare name: string
  declare lang: string
  declare iconUrl: string
  declare isNsfw: boolean

  constructor(data: ModelProps) {
    Object.assign<ModelProps, ModelProps>(this, data)
  }
}
