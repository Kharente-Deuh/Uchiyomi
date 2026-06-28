// SPDX-License-Identifier: AGPL-3.0-or-later

import { graphql } from '../../../../../utils/suwayomi/generated'

// Repo catalogue. ExtensionType: pkgName/name/lang/iconUrl/isNsfw/isInstalled/hasUpdate/versionName.
export const LIST_EXTENSIONS = graphql(`
  query ListExtensions {
    extensions {
      nodes {
        pkgName
        name
        lang
        iconUrl
        isNsfw
        isInstalled
        hasUpdate
        versionName
      }
    }
  }
`)

// Filtered + paginated catalogue read. Filters/sort/pagination are pushed into
// Suwayomi; totalCount reflects the applied filter. Does NOT fetch the repo —
// listAvailable() still owns the refresh path.
export const LIST_EXTENSIONS_PAGE = graphql(`
  query ListExtensionsPage(
    $filter: ExtensionFilterInput
    $order: [ExtensionOrderInput!]
    $first: Int
    $offset: Int
  ) {
    extensions(filter: $filter, order: $order, first: $first, offset: $offset) {
      totalCount
      nodes {
        pkgName
        name
        lang
        iconUrl
        isNsfw
        isInstalled
        hasUpdate
        versionName
      }
    }
  }
`)

// Refresh the repo list before listing (idempotent).
export const FETCH_EXTENSIONS = graphql(`
  mutation FetchExtensions {
    fetchExtensions(input: {}) {
      extensions { pkgName }
    }
  }
`)

// install/uninstall via patch. id is the pkgName string.
export const UPDATE_EXTENSION = graphql(`
  mutation UpdateExtension($id: String!, $patch: UpdateExtensionPatchInput!) {
    updateExtension(input: { id: $id, patch: $patch }) {
      extension { pkgName isInstalled }
    }
  }
`)

// Sources of an installed extension. SourceType.id is LongString! (string).
export const GET_EXTENSION_SOURCES = graphql(`
  query GetExtensionSources($pkgName: String!) {
    extension(pkgName: $pkgName) {
      source {
        nodes {
          id
          name
          lang
          isNsfw
          isConfigurable
          supportsLatest
        }
      }
    }
  }
`)

// Preferences union; __typename drives the tagged DTO mapping.
// NOTE: currentValue and default have conflicting scalar types across union members
// (Boolean for Switch/CheckBox, String for EditText/List, [String!] for MultiSelectList).
// SDL-required fix: alias the conflicting fields per concrete type so the document is valid.
export const GET_SOURCE_PREFERENCES = graphql(`
  query GetSourcePreferences($sourceId: LongString!) {
    source(id: $sourceId) {
      id
      isConfigurable
      preferences {
        __typename
        ... on SwitchPreference {
          key title summary visible
          currentValueBool: currentValue
          defaultBool: default
        }
        ... on CheckBoxPreference {
          key title summary visible
          currentValueBool: currentValue
          defaultBool: default
        }
        ... on EditTextPreference {
          key title summary visible dialogTitle dialogMessage
          currentValueStr: currentValue
          defaultStr: default
        }
        ... on ListPreference {
          key title summary visible entries entryValues
          currentValueStr: currentValue
          defaultStr: default
        }
        ... on MultiSelectListPreference {
          key title summary visible entries entryValues
          currentValueList: currentValue
          defaultList: default
        }
      }
    }
  }
`)

export const UPDATE_SOURCE_PREFERENCE = graphql(`
  mutation UpdateSourcePreference($source: LongString!, $change: SourcePreferenceChangeInput!) {
    updateSourcePreference(input: { source: $source, change: $change }) {
      preferences {
        __typename
        ... on SwitchPreference {
          key title summary visible
          currentValueBool: currentValue
          defaultBool: default
        }
        ... on CheckBoxPreference {
          key title summary visible
          currentValueBool: currentValue
          defaultBool: default
        }
        ... on EditTextPreference {
          key title summary visible dialogTitle dialogMessage
          currentValueStr: currentValue
          defaultStr: default
        }
        ... on ListPreference {
          key title summary visible entries entryValues
          currentValueStr: currentValue
          defaultStr: default
        }
        ... on MultiSelectListPreference {
          key title summary visible entries entryValues
          currentValueList: currentValue
          defaultList: default
        }
      }
    }
  }
`)
