// SPDX-License-Identifier: AGPL-3.0-or-later

import type { SourceComicChapter } from '../types'

export function isChapterDownloadInProgress(chapter: SourceComicChapter): boolean {
  return chapter.download !== undefined && chapter.download >= 0 && chapter.download < 100
}
