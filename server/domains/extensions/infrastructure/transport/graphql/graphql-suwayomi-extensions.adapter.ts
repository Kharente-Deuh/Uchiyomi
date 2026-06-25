// SPDX-License-Identifier: AGPL-3.0-or-later
import type { SuwayomiClient } from '../../../../../utils/suwayomi/client'
import type { ExtensionFilterInput } from '../../../../../utils/suwayomi/generated/graphql'
import type { ExtensionModel, ExtensionSource, ExtensionSourcePreferenceModel, ListExtensionsFilters, ListExtensionsQuery, Page, SuwayomiExtensionsPort, UpdatePreferenceParams } from '../../../extension.domain'
import { extensionToDomain, preferenceToDomain, sourceToDomain, toChangeInput } from './graphql-suwayomi-extensions.mapper'
import { FETCH_EXTENSIONS, GET_EXTENSION_SOURCES, GET_SOURCE_PREFERENCES, LIST_EXTENSIONS, LIST_EXTENSIONS_PAGE, UPDATE_EXTENSION, UPDATE_SOURCE_PREFERENCE } from './graphql-suwayomi-extensions.operations'

export class GraphqlSuwayomiExtensionsAdapter implements SuwayomiExtensionsPort {
  constructor(private readonly client: SuwayomiClient) {}

  async listExtensions({ filters, page, pageSize }: ListExtensionsQuery): Promise<Page<ExtensionModel>> {
    const filter = buildExtensionFilter(filters)
    const data = await this.client.execute(LIST_EXTENSIONS_PAGE, {
      filter,
      order: [{ by: 'NAME', byType: 'ASC' }],
      first: pageSize,
      offset: (page - 1) * pageSize,
    })

    return {
      items: data.extensions.nodes.map(n => extensionToDomain(n)),
      totalCount: data.extensions.totalCount,
    }
  }

  async listAll(): Promise<ExtensionModel[]> {
    await this.client.execute(FETCH_EXTENSIONS)
    const data = await this.client.execute(LIST_EXTENSIONS)

    return data.extensions.nodes.map(n => extensionToDomain(n))
  }

  async getExtension(pkgName: string): Promise<ExtensionModel | undefined> {
    const { items } = await this.listExtensions({ filters: { pkgName }, page: 1, pageSize: 1 })

    return items[0]
  }

  async install(pkgName: string): Promise<void> {
    await this.client.execute(UPDATE_EXTENSION, { id: pkgName, patch: { install: true } })
  }

  async uninstall(pkgName: string): Promise<void> {
    await this.client.execute(UPDATE_EXTENSION, { id: pkgName, patch: { uninstall: true } })
  }

  async listSources(pkgName: string): Promise<ExtensionSource[]> {
    const data = await this.client.execute(GET_EXTENSION_SOURCES, { pkgName })

    return data.extension.source.nodes.map(n => sourceToDomain(n))
  }

  async listSourcePreferences(sourceId: string): Promise<ExtensionSourcePreferenceModel[]> {
    const data = await this.client.execute(GET_SOURCE_PREFERENCES, { sourceId })

    return data.source.preferences.map((n, i) => preferenceToDomain(n, i))
  }

  async updateSourcePreference(p: UpdatePreferenceParams): Promise<ExtensionSourcePreferenceModel[]> {
    // Read current prefs to learn the type at `position` (the change input is type-specific).
    const current = await this.listSourcePreferences(p.sourceId)
    const target = current[p.position]
    if (!target) {
      throw new Error(`No preference at position ${p.position}`)
    }

    const source = await this.client.execute(UPDATE_SOURCE_PREFERENCE, {
      source: p.sourceId,
      change: toChangeInput(target.type, p),
    })

    // updateSourcePreference returns null if the source is not found.
    if (!source.updateSourcePreference) {
      throw new Error(`updateSourcePreference returned null for source ${p.sourceId}`)
    }

    return source.updateSourcePreference.preferences.map((n, i) => preferenceToDomain(n, i))
  }
}

function buildExtensionFilter(filters?: ListExtensionsFilters): ExtensionFilterInput | undefined {
  if (!filters) {
    return
  }

  const and: ExtensionFilterInput[] = []
  if (filters.pkgName !== undefined) {
    and.push({ pkgName: { equalTo: filters.pkgName } })
  }

  if (filters.search !== undefined) {
    and.push({ name: { includesInsensitive: filters.search } })
  }

  if (filters.isInstalled !== undefined) {
    and.push({ isInstalled: { equalTo: filters.isInstalled } })
  }

  if (filters.hasUpdate !== undefined) {
    and.push({ hasUpdate: { equalTo: filters.hasUpdate } })
  }

  if (filters.isNsfw !== undefined) {
    and.push({ isNsfw: { equalTo: filters.isNsfw } })
  }

  return and.length > 0 ? { and } : undefined
}
