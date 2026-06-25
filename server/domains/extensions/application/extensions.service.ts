import type { ExtensionModel, ExtensionSourcePreferenceModel, ListedExtension, StoredExtensionSource, UpdatePreferenceParams } from '../extension.domain'
import type { GetExtensionByPkgNameUseCaseOpts, GetExtensionHealthUseCaseOpts, GetExtensionHealthUseCaseResult, InstallExtensionUseCaseOpts, ListExtensionSourcesUseCaseOpts, ListExtensionsUseCaseOpts, ListExtensionsUseCaseResult, ListSourcePreferencesUseCaseOpts, SetSourceEnabledUseCaseOpts, UninstallExtensionUseCaseOpts, UpdateExtensionUseCaseOpts } from './usecases'
import { PrismaExtensionRepository } from '../infrastructure/persistence/prisma/prisma-extension.repository'
import { PrismaSourceRepository } from '../infrastructure/persistence/prisma/prisma-source.repository'
import { GraphqlSuwayomiExtensionsAdapter } from '../infrastructure/transport/graphql/graphql-suwayomi-extensions.adapter'
import { GetExtensionByPkgNameUseCase, GetExtensionHealthUseCase, InstallExtensionUseCase, ListExtensionSourcesUseCase, ListExtensionsUseCase, ListSourcePreferencesUseCase, SetSourceEnabledUseCase, UninstallExtensionUseCase, UpdateExtensionUseCase, UpdateSourcePreferenceUseCase } from './usecases'

export interface ExtensionsService {
  getExtensionByPkgName: (opts: GetExtensionByPkgNameUseCaseOpts) => Promise<ExtensionModel | undefined>
  listExtensions: (opts: ListExtensionsUseCaseOpts) => Promise<ListExtensionsUseCaseResult>
  installExtension: (opts: InstallExtensionUseCaseOpts) => Promise<ListedExtension>
  uninstallExtension: (opts: UninstallExtensionUseCaseOpts) => Promise<ListedExtension>
  updateExtension: (opts: UpdateExtensionUseCaseOpts) => Promise<ListedExtension>
  listSourcePreferences: (opts: ListSourcePreferencesUseCaseOpts) => Promise<ExtensionSourcePreferenceModel[]>
  updateSourcePreference: (opts: UpdatePreferenceParams) => Promise<ExtensionSourcePreferenceModel[]>
  getExtensionHealth: (opts: GetExtensionHealthUseCaseOpts) => Promise<GetExtensionHealthUseCaseResult | undefined>
  listExtensionSources: (opts: ListExtensionSourcesUseCaseOpts) => Promise<StoredExtensionSource[]>
  setSourceEnabled: (opts: SetSourceEnabledUseCaseOpts) => Promise<StoredExtensionSource>
}

const { suwayomiUrl } = useRuntimeConfig()
const suwayomiClient = createSuwayomiClient({ endpoint: `${suwayomiUrl}/api/graphql` })
const suwayomiExtensionsPort = new GraphqlSuwayomiExtensionsAdapter(suwayomiClient)
const overlay = new PrismaExtensionRepository(prisma)
const sourceOverlay = new PrismaSourceRepository(prisma)

function getExtensionByPkgName(opts: GetExtensionByPkgNameUseCaseOpts): Promise<ExtensionModel | undefined> {
  return new GetExtensionByPkgNameUseCase(suwayomiExtensionsPort).execute(opts)
}

function listExtensions(opts: ListExtensionsUseCaseOpts): Promise<ListExtensionsUseCaseResult> {
  return new ListExtensionsUseCase(suwayomiExtensionsPort, overlay).execute(opts)
}

function installExtension(opts: InstallExtensionUseCaseOpts): Promise<ListedExtension> {
  return new InstallExtensionUseCase(suwayomiExtensionsPort, overlay, sourceOverlay).execute(opts)
}

function uninstallExtension(opts: UninstallExtensionUseCaseOpts): Promise<ListedExtension> {
  return new UninstallExtensionUseCase(suwayomiExtensionsPort, overlay).execute(opts)
}

function updateExtension(opts: UpdateExtensionUseCaseOpts): Promise<ListedExtension> {
  return new UpdateExtensionUseCase(suwayomiExtensionsPort, overlay, sourceOverlay).execute(opts)
}

function listSourcePreferences(opts: ListSourcePreferencesUseCaseOpts): Promise<ExtensionSourcePreferenceModel[]> {
  return new ListSourcePreferencesUseCase(suwayomiExtensionsPort).execute(opts)
}

function updateSourcePreference(opts: UpdatePreferenceParams): Promise<ExtensionSourcePreferenceModel[]> {
  return new UpdateSourcePreferenceUseCase(suwayomiExtensionsPort).execute(opts)
}

function getExtensionHealth(opts: GetExtensionHealthUseCaseOpts): Promise<GetExtensionHealthUseCaseResult | undefined> {
  return new GetExtensionHealthUseCase(overlay).execute(opts)
}

function listExtensionSources(opts: ListExtensionSourcesUseCaseOpts): Promise<StoredExtensionSource[]> {
  return new ListExtensionSourcesUseCase(sourceOverlay).execute(opts)
}

function setSourceEnabled(opts: SetSourceEnabledUseCaseOpts): Promise<StoredExtensionSource> {
  return new SetSourceEnabledUseCase(sourceOverlay).execute(opts)
}

export function extensionsService(): ExtensionsService {
  return {
    getExtensionByPkgName,
    listExtensions,
    installExtension,
    uninstallExtension,
    updateExtension,
    listSourcePreferences,
    updateSourcePreference,
    getExtensionHealth,
    listExtensionSources,
    setSourceEnabled,
  }
}
