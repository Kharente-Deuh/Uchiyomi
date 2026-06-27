import { describe, expect, it } from 'vitest'
// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment node
import { DownloadPriority, DownloadStatus, LibraryStatus, ReadingDirection, ReadingMode, SeriesStatus, SeriesType } from '../prisma/generated/enums'

// The `prisma-client` generator exposes enums as runtime objects but no runtime
// dmmf. Enum values/order are asserted here; model/field validity is covered by
// `prisma generate` and the constraint integration tests.
describe('overlay enums', () => {
  it('declares the series enums', () => {
    expect(Object.values(SeriesType)).toEqual(['MANGA', 'WEBTOON', 'COMIC'])
    expect(Object.values(SeriesStatus)).toEqual(['ONGOING', 'COMPLETED', 'HIATUS', 'UNKNOWN'])
  })

  it('declares the library and reading enums', () => {
    expect(Object.values(LibraryStatus)).toEqual(['READING', 'STOPPED'])
    expect(Object.values(ReadingMode)).toEqual(['PAGED', 'LONG_STRIP'])
    expect(Object.values(ReadingDirection)).toEqual(['LTR', 'RTL'])
  })

  it('declares the download enums', () => {
    expect(Object.values(DownloadStatus)).toEqual(['PENDING', 'RUNNING', 'DONE', 'FAILED'])
    expect(Object.values(DownloadPriority)).toEqual(['NEW_CHAPTER', 'BACKFILL'])
  })
})
