import type { ComicSource } from '~/features/comics/types'
import asuraImg from '~/assets/images/sources/asurascans.webp'
import kingOfShojoImg from '~/assets/images/sources/kingofshojo.webp'
import { ASURA_SCANS_URL, KING_OF_SHOJO_URL } from '~/constants'

export interface SourceDetails {
  url: string
  name: string
  image: string
  color: string
}

export interface SourcesComposable {
  getSourceDetails: (source: ComicSource) => SourceDetails
}

export function useSources(): SourcesComposable {
  const { t } = useI18n()

  const sourceDetails: Record<ComicSource, SourceDetails> = {
    asurascans: {
      url: ASURA_SCANS_URL,
      name: t('sources.asurascans.title'),
      image: asuraImg,
      color: '#913fe2',
    },
    kingofshojo: {
      url: KING_OF_SHOJO_URL,
      name: t('sources.kingofshojo.title'),
      image: kingOfShojoImg,
      color: '#2503e5',
    },
  }

  function getSourceDetails(source: ComicSource): SourceDetails {
    return sourceDetails[source]
  }

  return {
    getSourceDetails,
  }
}
