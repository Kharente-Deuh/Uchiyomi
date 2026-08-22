import type { ComicSource } from '~/features/comics/types'
import asuraImg from '~/assets/images/sources/asurascans.webp'

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
      url: 'https://asurascans.com',
      name: t('sources.asurascans.title'),
      image: asuraImg,
      color: '#913fe2',
    },
  }

  function getSourceDetails(source: ComicSource): SourceDetails {
    return sourceDetails[source]
  }

  return {
    getSourceDetails,
  }
}
