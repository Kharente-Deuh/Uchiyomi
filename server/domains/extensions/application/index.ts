// SPDX-License-Identifier: AGPL-3.0-or-later

import { PrismaExtensionRepository } from '../infrastructure/persistence/prisma/prisma-extension.repository'
import { PrismaSourceRepository } from '../infrastructure/persistence/prisma/prisma-source.repository'
import { GraphqlSuwayomiExtensionsAdapter } from '../infrastructure/transport/graphql/graphql-suwayomi-extensions.adapter'
import { GetExtensionHealthUseCase, InstallExtensionUseCase, ListExtensionSourcesUseCase, ListExtensionsUseCase, ListSourcePreferencesUseCase, SetSourceEnabledUseCase, UninstallExtensionUseCase, UpdateSourcePreferenceUseCase } from './usecases'

const { suwayomiUrl } = useRuntimeConfig()

const suwayomiClient = createSuwayomiClient({ endpoint: `${suwayomiUrl}/api/graphql` })
const suwayomiExtensionsPort = new GraphqlSuwayomiExtensionsAdapter(suwayomiClient)
const overlay = new PrismaExtensionRepository(prisma)
const sourceOverlay = new PrismaSourceRepository(prisma)

export const listExtensions = new ListExtensionsUseCase(suwayomiExtensionsPort, overlay)
export const installExtension = new InstallExtensionUseCase(suwayomiExtensionsPort, overlay, sourceOverlay)
export const uninstallExtension = new UninstallExtensionUseCase(suwayomiExtensionsPort, overlay)
export const listSourcePreferences = new ListSourcePreferencesUseCase(suwayomiExtensionsPort)
export const updateSourcePreference = new UpdateSourcePreferenceUseCase(suwayomiExtensionsPort)
export const getExtensionHealth = new GetExtensionHealthUseCase(overlay)
export const listExtensionSources = new ListExtensionSourcesUseCase(sourceOverlay)
export const setSourceEnabled = new SetSourceEnabledUseCase(sourceOverlay)

// Resolve the Suwayomi-internal icon path for a pkgName, honouring the same
// visibility rules as the listing (non-admins only see installed sources; NSFW
// is gated). The catalogue fetch is cached briefly so a page rendering N icons
// triggers a single Suwayomi fetch rather than N. Icons are immutable per
// version, so the proxied response itself is cached long-lived by the browser.
const cachedAvailable = defineCachedFunction(
  () => suwayomiExtensionsPort.listAll(),
  { maxAge: 60, name: 'suwayomi-catalogue', getKey: () => 'all' },
)

export async function resolveExtensionIconUrl(
  pkgName: string,
  opts: { isAdmin: boolean, viewerCanSeeNsfw: boolean },
): Promise<string | undefined> {
  const available = await cachedAvailable()
  const ext = available.find(e => e.pkgName === pkgName)
  if (!ext || !ext.iconUrl) {
    return
  }

  if (!opts.isAdmin && !ext.isInstalled) {
    return
  }

  if (!opts.viewerCanSeeNsfw && ext.isNsfw) {
    return
  }

  return ext.iconUrl
}
