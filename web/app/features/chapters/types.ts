export interface Chapter {
  id: string
  comicId: string
  publishedAt: Date
  earlyAccessUntil?: Date
  sourceChapterSlug: string
  title: string
  number: number
  pagesNb: number
  download: number
  progress?: ChapterProgress
}

export interface ChapterProgress {
  updatedAt: Date
  page: number
}

export interface SaveChapterProgressRequest {
  id: string
  page: number
}

export type DetailedChapter = Chapter & {
  pageUrls: string[]
  next?: AdjacentChapter
  previous?: AdjacentChapter
}

export interface AdjacentChapter {
  id: string
  title: string
  number: number
}
