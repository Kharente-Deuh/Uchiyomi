import type { AsuraScansComicChapter } from '~/features/asurascans/types'
import type { KingOfShojoComicChapter } from '~/features/kingofshojo/types'

export function isChapterDownloadInProgress(chapter: KingOfShojoComicChapter | AsuraScansComicChapter): boolean {
  return chapter.download !== undefined && chapter.download >= 0 && chapter.download < 100
}
