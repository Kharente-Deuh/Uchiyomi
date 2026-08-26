// SPDX-License-Identifier: AGPL-3.0-or-later

import { describe, expect, it } from 'vitest'
import { getSourceConfig, SOURCES_CONFIG } from './sources.config'

describe('sources.config', () => {
  it('defines valid config for asurascans', () => {
    const config = getSourceConfig('asurascans')
    expect(config).toBeDefined()
    expect(config?.id).toBe('asurascans')
    expect(config?.allowedSorts).toContain('popular')
    expect(config?.allowedSorts).toContain('rating')
  })

  it('defines valid config for kingofshojo', () => {
    const config = getSourceConfig('kingofshojo')
    expect(config).toBeDefined()
    expect(config?.id).toBe('kingofshojo')
    expect(config?.allowedSorts).toContain('popular')
    expect(config?.allowedSorts).not.toContain('rating')
  })

  it('returns undefined for unknown source', () => {
    expect(getSourceConfig('unknown-source')).toBeUndefined()
  })

  it('defines configs for all registered sources', () => {
    expect(Object.keys(SOURCES_CONFIG)).toEqual(expect.arrayContaining(['asurascans', 'kingofshojo']))
  })
})
