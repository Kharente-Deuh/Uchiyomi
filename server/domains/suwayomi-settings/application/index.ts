// SPDX-License-Identifier: AGPL-3.0-or-later

import { GraphqlSuwayomiSettingsAdapter } from '../infrastructure/transport/graphql/graphql-suwayomi-settings.adapter'
import { ReconcileSettingsUseCase } from './usecases'

const { suwayomiUrl } = useRuntimeConfig()

const suwayomiClient = createSuwayomiClient({ endpoint: `${suwayomiUrl}/api/graphql` })
const settingsPort = new GraphqlSuwayomiSettingsAdapter(suwayomiClient)

export const reconcileSuwayomiSettings = new ReconcileSettingsUseCase(settingsPort)
