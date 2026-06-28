// SPDX-License-Identifier: AGPL-3.0-or-later

import type { SuwayomiClient } from '../../../../../utils/suwayomi/client'
import type { ManagedSettings, SettingsPatch, SettingsPort } from '../../../suwayomi-settings.domain'
import { GET_SETTINGS, SET_SETTINGS } from './graphql-suwayomi-settings.operations'

export class GraphqlSuwayomiSettingsAdapter implements SettingsPort {
  constructor(private readonly client: SuwayomiClient) {}

  async read(): Promise<ManagedSettings> {
    const data = await this.client.execute(GET_SETTINGS)

    return {
      autoDownloadNewChapters: data.settings.autoDownloadNewChapters ?? null,
      extensionRepos: data.settings.extensionRepos ?? null,
    }
  }

  async apply(patch: SettingsPatch): Promise<void> {
    await this.client.execute(SET_SETTINGS, { settings: patch })
  }
}
