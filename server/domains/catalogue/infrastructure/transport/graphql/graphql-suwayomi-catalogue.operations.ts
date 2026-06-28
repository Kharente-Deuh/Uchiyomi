// SPDX-License-Identifier: AGPL-3.0-or-later

import { graphql } from '../../../../../utils/suwayomi/generated'

// NOTE: field names match the committed schema.graphql SDL.
// SourceType.iconUrl is String! (non-nullable) in the SDL.
export const LIST_SOURCES = graphql(`
  query ListSources {
    sources {
      nodes {
        id
        name
        lang
        iconUrl
        isNsfw
      }
    }
  }
`)

// fetchSourceManga browses a source by type: SEARCH (uses query), POPULAR, or LATEST.
// query is optional — only SEARCH uses it; popular/latest pass null.
// sourceId is LongString! (a 64-bit int encoded as string).
export const SEARCH_SOURCE = graphql(`
  mutation SearchSource($sourceId: LongString!, $type: FetchSourceMangaType!, $query: String, $page: Int!) {
    fetchSourceManga(input: { source: $sourceId, type: $type, query: $query, page: $page }) {
      mangas {
        id
        title
        thumbnailUrl
        inLibrary
      }
      hasNextPage
    }
  }
`)

// NOTE: manga(id: Int!) — mangaId must be converted from string to number before calling.
// ChapterType.uploadDate is LongString! (non-nullable) in the SDL.
export const GET_MANGA_DETAILS = graphql(`
  query GetMangaDetails($mangaId: Int!) {
    manga(id: $mangaId) {
      id
      title
      thumbnailUrl
      inLibrary
      author
      description
      status
      chapters {
        nodes {
          id
          name
          chapterNumber
          uploadDate
          isDownloaded
        }
      }
    }
  }
`)

// Eager chapter enrichment for a search result. fetchChapters is a mutation that
// makes Suwayomi fetch the manga's chapter list from the remote source (it hits the
// site), then returns the full list. We derive count + last chapter from it.
// mangaId is Int! in the SDL — convert from the domain string id before calling.
export const FETCH_CHAPTERS = graphql(`
  mutation FetchSourceMangaChapters($mangaId: Int!) {
    fetchChapters(input: { mangaId: $mangaId }) {
      chapters {
        name
        uploadDate
      }
    }
  }
`)
