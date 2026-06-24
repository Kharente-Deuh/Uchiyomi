// SPDX-License-Identifier: AGPL-3.0-or-later
import { graphql } from '../../../../../utils/suwayomi/generated'

// Managed subset only (ADR-0012). Extend alongside the allowlist in M6.2.
export const GET_SETTINGS = graphql(`
  query GetManagedSettings {
    settings {
      autoDownloadNewChapters
      extensionRepos
    }
  }
`)

export const SET_SETTINGS = graphql(`
  mutation SetManagedSettings($settings: PartialSettingsTypeInput!) {
    setSettings(input: { settings: $settings }) {
      settings {
        autoDownloadNewChapters
        extensionRepos
      }
    }
  }
`)
