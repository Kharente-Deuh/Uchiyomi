// SPDX-License-Identifier: AGPL-3.0-or-later
import { describe, expect, it } from 'vitest'
import * as Extension from '../../server/domains/extensions/extension.domain'

describe('extension.domain', () => {
  it('available constructs from props', () => {
    const ext = new Extension.ExtensionModel({
      pkgName: 'eu.kanade.tachiyomi.extension.en.foo',
      name: 'Foo',
      lang: 'en',
      isNsfw: false,
      isInstalled: true,
      hasUpdate: false,
      versionName: '1.0.0',
    })
    expect(ext.name).toBe('Foo')
    expect(ext.iconUrl).toBeUndefined()
  })
})
