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
    expect(res.totalCount).toBe(42)
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
