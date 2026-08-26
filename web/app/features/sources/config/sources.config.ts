// SPDX-License-Identifier: AGPL-3.0-or-later

import type { SourceConfig } from '../types'
import type { ComicSource } from '~/features/comics/types'
import asuraImg from '~/assets/images/sources/asurascans.webp'
import kingOfShojoImg from '~/assets/images/sources/kingofshojo.webp'
import { ASURA_SCANS_URL, ASURA_SOURCE_NAME, KING_OF_SHOJO_SOURCE_NAME, KING_OF_SHOJO_URL } from '~/constants'

export const SOURCES_CONFIG: Record<ComicSource, SourceConfig> = {
  [ASURA_SOURCE_NAME]: {
    id: ASURA_SOURCE_NAME,
    nameKey: 'sources.asurascans.title',
    url: ASURA_SCANS_URL,
    image: asuraImg,
    color: '#913fe2',
    allowedSorts: ['popular', 'latest', 'rating', 'title', 'newest'],
  },
  [KING_OF_SHOJO_SOURCE_NAME]: {
    id: KING_OF_SHOJO_SOURCE_NAME,
    nameKey: 'sources.kingofshojo.title',
    url: KING_OF_SHOJO_URL,
    image: kingOfShojoImg,
    color: '#2503e5',
    allowedSorts: ['popular', 'latest', 'title', 'newest'],
    enrichSearchFromSeries: true,
  },
}

export function getSourceConfig(source: string): SourceConfig | undefined {
  return (SOURCES_CONFIG as Record<string, SourceConfig>)[source]
}

export function getComicOriginUrl(config: SourceConfig, publicUrl: string): string {
  if (!publicUrl) {
    return ''
  }

  if (publicUrl.startsWith('http')) {
    return publicUrl
  }

  return `${config.url}${publicUrl}`
}
