// SPDX-License-Identifier: AGPL-3.0-or-later

import type { StoredExtensionSource } from '../../../extension.domain'

export interface RawSource {
  id: string
  pkgName: string
  name: string
  lang: string
  isNsfw: boolean
  isConfigurable: boolean
  isEnabled: boolean
}

export function sourceToDomain(row: RawSource): StoredExtensionSource {
  return {
    id: row.id,
    pkgName: row.pkgName,
    name: row.name,
    lang: row.lang,
    isNsfw: row.isNsfw,
    isConfigurable: row.isConfigurable,
    isEnabled: row.isEnabled,
  }
}
