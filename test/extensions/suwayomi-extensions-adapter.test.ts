// SPDX-License-Identifier: AGPL-3.0-or-later
import { describe, expect, it, vi } from 'vitest'
import { GraphqlSuwayomiExtensionsAdapter } from '../../server/domains/extensions/infrastructure/transport/graphql/graphql-suwayomi-extensions.adapter'

function makeClient(totalCount: number, nodes: unknown[]): { client: never, execute: ReturnType<typeof vi.fn> } {
  const execute = vi.fn().mockResolvedValue({ extensions: { totalCount, nodes } })

  return { client: { execute } as never, execute }
}

const node = {
  pkgName: 'a',
  name: 'A',
  lang: 'en',
  iconUrl: '',
  isNsfw: false,
  isInstalled: true,
  hasUpdate: false,
  versionName: '1',
}

describe('graphqlSuwayomiExtensionsAdapter.listExtensions', () => {
  it('translates filters/sort/pagination into GraphQL variables and maps the result', async () => {
    const { client, execute } = makeClient(42, [node])
    const adapter = new GraphqlSuwayomiExtensionsAdapter(client)

    const res = await adapter.listExtensions({
      page: 2,
      pageSize: 20,
      filters: { search: 'naruto', isInstalled: true, hasUpdate: false, isNsfw: false },
    })

    expect(execute).toHaveBeenCalledTimes(1)
    const [, vars] = execute.mock.calls[0]
    expect(vars).toEqual({
      filter: {
        and: [
          { name: { includesInsensitive: 'naruto' } },
          { isInstalled: { equalTo: true } },
          { hasUpdate: { equalTo: false } },
          { isNsfw: { equalTo: false } },
        ],
      },
      order: [{ by: 'NAME', byType: 'ASC' }],
      first: 20,
      offset: 20,
    })
    expect(res.total).toBe(42)
    expect(res.items.map(i => i.pkgName)).toEqual(['a'])
  })

  it('omits the filter key entirely when no filters are provided', async () => {
    const { client, execute } = makeClient(3, [])
    const adapter = new GraphqlSuwayomiExtensionsAdapter(client)

    await adapter.listExtensions({ page: 1, pageSize: 20 })

    const [, vars] = execute.mock.calls[0]
    expect(vars.filter).toBeUndefined()
    expect(vars.offset).toBe(0)
    expect(vars.first).toBe(20)
  })

  it('maps the pkgName filter to an equalTo predicate (single-extension lookup)', async () => {
    const { client, execute } = makeClient(1, [node])
    const adapter = new GraphqlSuwayomiExtensionsAdapter(client)

    await adapter.listExtensions({ page: 1, pageSize: 1, filters: { pkgName: 'com.foo.bar' } })

    const [, vars] = execute.mock.calls[0]
    expect(vars.filter).toEqual({ and: [{ pkgName: { equalTo: 'com.foo.bar' } }] })
    expect(vars.first).toBe(1)
    expect(vars.offset).toBe(0)
  })
})

describe('graphqlSuwayomiExtensionsAdapter.getExtension', () => {
  it('delegates to listExtensions with pkgName equalTo filter and first:1, returning the single item', async () => {
    const { client, execute } = makeClient(1, [node])
    const adapter = new GraphqlSuwayomiExtensionsAdapter(client)

    const result = await adapter.getExtension('a')

    expect(execute).toHaveBeenCalledTimes(1)
    const [, vars] = execute.mock.calls[0]
    expect(vars.filter).toEqual({ and: [{ pkgName: { equalTo: 'a' } }] })
    expect(vars.first).toBe(1)
    expect(vars.offset).toBe(0)
    expect(result?.pkgName).toBe('a')
  })

  it('returns undefined when no matching extension is found', async () => {
    const { client } = makeClient(0, [])
    const adapter = new GraphqlSuwayomiExtensionsAdapter(client)

    const result = await adapter.getExtension('com.nonexistent.pkg')

    expect(result).toBeUndefined()
  })
})

describe('updateSourcePreferences (batch)', () => {
  it('reads the source once, then posts one change per item with the type from the change', async () => {
    const prefs = {
      source: { preferences: [
        { __typename: 'SwitchPreference', key: 's', title: null, summary: null, visible: true, currentValueBool: false, defaultBool: false },
        { __typename: 'ListPreference', key: 'l', title: null, summary: null, visible: true, currentValueStr: 'a', defaultStr: 'a', entries: ['A'], entryValues: ['a'] },
      ] },
    }
    const updated = { updateSourcePreference: { preferences: prefs.source.preferences } }
    const execute = vi.fn()
      .mockResolvedValueOnce(prefs) // GET_SOURCE_PREFERENCES
      .mockResolvedValue(updated) // each UPDATE_SOURCE_PREFERENCE
    const adapter = new GraphqlSuwayomiExtensionsAdapter({ execute } as never)

    await adapter.updateSourcePreferences('src1', [
      { position: 0, type: 'switch', booleanValue: true },
      { position: 1, type: 'list', textValue: 'b' },
    ])

    // 1 read + 2 writes
    expect(execute).toHaveBeenCalledTimes(3)
    expect(execute.mock.calls[1][1]).toMatchObject({ source: 'src1', change: { position: 0, switchState: true } })
    expect(execute.mock.calls[2][1]).toMatchObject({ source: 'src1', change: { position: 1, listState: 'b' } })
  })

  it('throws when a change targets a position with no preference', async () => {
    const execute = vi.fn().mockResolvedValueOnce({ source: { preferences: [] } })
    const adapter = new GraphqlSuwayomiExtensionsAdapter({ execute } as never)
    await expect(adapter.updateSourcePreferences('src1', [{ position: 0, type: 'switch', booleanValue: true }])).rejects.toThrow('No preference at position 0')
  })

  it('throws when the change type does not match the live preference type at that position (type-drift guard)', async () => {
    const prefs = {
      source: { preferences: [
        { __typename: 'SwitchPreference', key: 's', title: null, summary: null, visible: true, currentValueBool: false, defaultBool: false },
      ] },
    }
    const execute = vi.fn().mockResolvedValueOnce(prefs)
    const adapter = new GraphqlSuwayomiExtensionsAdapter({ execute } as never)

    // Sending 'checkbox' for a position that holds a 'switch' → type-drift error.
    await expect(adapter.updateSourcePreferences('src1', [{ position: 0, type: 'checkbox', booleanValue: true }]))
      .rejects
      .toThrow('Preference type drift at position 0')
  })
})
