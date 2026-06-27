import type { ExtensionModel, ExtensionSettings, ListedExtension, StoredExtensionSource } from '../extension.domain'
import type { GetExtensionByPkgNameUseCaseOpts, GetExtensionHealthUseCaseOpts, GetExtensionHealthUseCaseResult, GetExtensionSettingsUseCaseOpts, InstallExtensionUseCaseOpts, ListExtensionSourcesUseCaseOpts, ListExtensionsUseCaseOpts, ListExtensionsUseCaseResult, SetSourceEnabledUseCaseOpts, UninstallExtensionUseCaseOpts, UpdateExtensionSettingsUseCaseOpts, UpdateExtensionUseCaseOpts } from './usecases'
import { PrismaExtensionRepository } from '../infrastructure/persistence/prisma/prisma-extension.repository'
import { PrismaSourceRepository } from '../infrastructure/persistence/prisma/prisma-source.repository'
import { GraphqlSuwayomiExtensionsAdapter } from '../infrastructure/transport/graphql/graphql-suwayomi-extensions.adapter'
import { GetExtensionByPkgNameUseCase, GetExtensionHealthUseCase, GetExtensionSettingsUseCase, InstallExtensionUseCase, ListExtensionSourcesUseCase, ListExtensionsUseCase, SetSourceEnabledUseCase, UninstallExtensionUseCase, UpdateExtensionSettingsUseCase, UpdateExtensionUseCase } from './usecases'

export interface ExtensionsService {
  getExtensionByPkgName: (opts: GetExtensionByPkgNameUseCaseOpts) => Promise<ExtensionModel | undefined>
  listExtensions: (opts: ListExtensionsUseCaseOpts) => Promise<ListExtensionsUseCaseResult>
  installExtension: (opts: InstallExtensionUseCaseOpts) => Promise<ListedExtension>
  uninstallExtension: (opts: UninstallExtensionUseCaseOpts) => Promise<ListedExtension>
  updateExtension: (opts: UpdateExtensionUseCaseOpts) => Promise<ListedExtension>
  getExtensionHealth: (opts: GetExtensionHealthUseCaseOpts) => Promise<GetExtensionHealthUseCaseResult | undefined>
  listExtensionSources: (opts: ListExtensionSourcesUseCaseOpts) => Promise<StoredExtensionSource[]>
  setSourceEnabled: (opts: SetSourceEnabledUseCaseOpts) => Promise<StoredExtensionSource>
  getExtensionSettings: (opts: GetExtensionSettingsUseCaseOpts) => Promise<ExtensionSettings>
  updateExtensionSettings: (opts: UpdateExtensionSettingsUseCaseOpts) => Promise<ExtensionSettings>
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

function getExtensionHealth(opts: GetExtensionHealthUseCaseOpts): Promise<GetExtensionHealthUseCaseResult | undefined> {
  return new GetExtensionHealthUseCase(overlay).execute(opts)
}

function listExtensionSources(opts: ListExtensionSourcesUseCaseOpts): Promise<StoredExtensionSource[]> {
  return new ListExtensionSourcesUseCase(sourceOverlay).execute(opts)
}

function setSourceEnabled(opts: SetSourceEnabledUseCaseOpts): Promise<StoredExtensionSource> {
  return new SetSourceEnabledUseCase(sourceOverlay).execute(opts)
}

function getExtensionSettings(opts: GetExtensionSettingsUseCaseOpts): Promise<ExtensionSettings> {
  return new GetExtensionSettingsUseCase(sourceOverlay, suwayomiExtensionsPort).execute(opts)
}

function updateExtensionSettings(opts: UpdateExtensionSettingsUseCaseOpts): Promise<ExtensionSettings> {
  return new UpdateExtensionSettingsUseCase(sourceOverlay, suwayomiExtensionsPort).execute(opts)
}

export function extensionsService(): ExtensionsService {
  return {
    getExtensionByPkgName,
    listExtensions,
    installExtension,
    uninstallExtension,
    updateExtension,
    getExtensionHealth,
    listExtensionSources,
    setSourceEnabled,
    getExtensionSettings,
    updateExtensionSettings,
  }
}
