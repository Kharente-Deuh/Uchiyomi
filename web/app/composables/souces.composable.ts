import type { ComicSource } from '~/features/comics/types'
import { getSourceConfig } from '~/features/sources/config/sources.config'

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

  function getSourceDetails(source: ComicSource): SourceDetails {
    const config = getSourceConfig(source)
    if (!config) {
      throw new Error(`Unknown source: ${source}`)
    }

    return {
      url: config.url,
      name: t(config.nameKey),
      image: config.image,
      color: config.color,
    }
  }

  return {
    getSourceDetails,
  }
}
