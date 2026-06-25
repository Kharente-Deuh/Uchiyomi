// SPDX-License-Identifier: AGPL-3.0-or-later
import type { ReconcileSettingsUseCaseResult } from './usecases'
import { GraphqlSuwayomiSettingsAdapter } from '../infrastructure/transport/graphql/graphql-suwayomi-settings.adapter'
import { ReconcileSettingsUseCase } from './usecases'

export interface SuwayomiSettingsService {
  reconcile: () => Promise<ReconcileSettingsUseCaseResult>
}

const { suwayomiUrl } = useRuntimeConfig()

const suwayomiClient = createSuwayomiClient({ endpoint: `${suwayomiUrl}/api/graphql` })
const settingsPort = new GraphqlSuwayomiSettingsAdapter(suwayomiClient)

function reconcile(): Promise<ReconcileSettingsUseCaseResult> {
  return new ReconcileSettingsUseCase(settingsPort).execute()
}

export function suwayomiSettingsService(): SuwayomiSettingsService {
  return {
    reconcile,
  }
}
