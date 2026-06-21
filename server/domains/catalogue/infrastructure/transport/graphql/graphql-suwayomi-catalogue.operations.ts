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

// fetchSourceManga is a mutation but an idempotent fetch — re-fetching a search
// page re-upserts the same rows in Suwayomi with no client-visible side effects,
// so the client's transport/timeout retry policy is safe here.
// NOTE: Suwayomi exposes source search as a mutation (fetchSourceManga), not a query.
// The type field must be SEARCH to perform a keyword search.
// sourceId is LongString! in the SDL (a 64-bit int encoded as string).
export const SEARCH_SOURCE = graphql(`
  mutation SearchSource($sourceId: LongString!, $query: String!, $page: Int!) {
    fetchSourceManga(input: { source: $sourceId, query: $query, page: $page, type: SEARCH }) {
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
