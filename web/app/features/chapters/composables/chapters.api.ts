import type { Chapter, ChapterProgress, DetailedChapter, SaveChapterProgressRequest } from '../types'
import type { ApiResponse } from '~/utils/api'
import { ApiError, initApi } from '~/utils/api'

export interface ChaptersApi {
  retryDownload: (chapterId: string) => Promise<ApiResponse<void>>
  getByIds: (ids: string[]) => Promise<ApiResponse<Chapter[]>>
  getByComicId: (comicId: string) => Promise<ApiResponse<Chapter[]>>
  getById: (id: string) => Promise<ApiResponse<DetailedChapter>>
  saveProgress: (req: SaveChapterProgressRequest) => Promise<ApiResponse<ChapterProgress>>
}

function chapterFromHTTP({ progress, ...chapter }: Chapter): Chapter {
  return {
    ...chapter,
    progress: progress
      ? {
          updatedAt: new Date(progress.updatedAt),
          page: progress.page,
        }
      : undefined,
  }
}

export function createChaptersApi(): ChaptersApi {
  const api = initApi('/chapters')

  async function retryDownload(chapterId: string): Promise<ApiResponse<void>> {
    try {
      await api(`/${chapterId}/retry`, { method: 'POST' })

      return { success: true, data: undefined }
    } catch (error) {
      return { success: false, error: ApiError.fromFetchError(error) }
    }
  }

  async function getByIds(ids: string[]): Promise<ApiResponse<Chapter[]>> {
    try {
      const response = await api<Chapter[]>(`/list`, { method: 'POST', body: { ids } })

      return {
        success: true,
        data: response.map(chapter => chapterFromHTTP(chapter)),
      }
    } catch (error) {
      return { success: false, error: ApiError.fromFetchError(error) }
    }
  }

  async function getByComicId(comicId: string): Promise<ApiResponse<Chapter[]>> {
    try {
      const response = await api<Chapter[]>(`/`, { params: { comicId } })

      return {
        success: true,
        data: response.map(chapter => chapterFromHTTP(chapter)),
      }
    } catch (error) {
      return { success: false, error: ApiError.fromFetchError(error) }
    }
  }

  async function getById(id: string): Promise<ApiResponse<DetailedChapter>> {
    try {
      const { next, previous, pageUrls, ...response } = await api<DetailedChapter>(`/${id}`)

      return {
        success: true,
        data: {
          ...chapterFromHTTP(response),
          pageUrls: pageUrls ?? [],
          next,
          previous,
        },
      }
    } catch (error) {
      return { success: false, error: ApiError.fromFetchError(error) }
    }
  }

  async function saveProgress({ id, ...body }: SaveChapterProgressRequest): Promise<ApiResponse<ChapterProgress>> {
    try {
      const response = await api<ChapterProgress>(`/${id}/progress`, { method: 'PUT', body })

      return { success: true, data: response }
    } catch (error) {
      return { success: false, error: ApiError.fromFetchError(error) }
    }
  }

  return {
    retryDownload,
    getByIds,
    getByComicId,
    getById,
    saveProgress,
  }
}
