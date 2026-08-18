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
}
