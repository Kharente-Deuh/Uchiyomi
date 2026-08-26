<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { ComicSource } from '~/features/comics/types'
import type { SourceComicInfos } from '~/features/sources/types'
import defaultCover from '~/assets/images/default/comic-cover.webp'
import { AUTHENTICATED_ROUTE_GROUP } from '~/constants/auth'
import { getComicOriginUrl, getSourceConfig } from '~/features/sources/config/sources.config'

definePageMeta({
  layout: 'default',
  authGroups: [AUTHENTICATED_ROUTE_GROUP],
})

const route = useRoute('browse-sources-source-slug')
const sourceParam = route.params.source as string
const config = getSourceConfig(sourceParam)

if (!config) {
  await navigateTo('/browse/sources')
}

const sourceId = sourceParam as ComicSource
const toast = useToast()
const { t } = useI18n()
const { smAndDown } = useDisplay()
const api = createSourceApi(sourceId)

const fetchInfosLoading = ref(false)
const infos = ref<SourceComicInfos>()

const coverSrc = ref<string>()

onMounted(() => {
  fetchInfos()
})

async function fetchInfos(): Promise<void> {
  fetchInfosLoading.value = true

  const res = await api.getInfosBySlug(route.params.slug as string)

  if (res.success) {
    infos.value = res.data
    coverSrc.value = res.data.cover
  } else {
    console.error('api.getInfosBySlug', res.error)
    toast.error(res.error.status === 404
      ? t('sources.comic.notFound')
      : t('error.unknown'))

    await navigateTo(`/browse/sources/${sourceId}`)
  }

  fetchInfosLoading.value = false
}

const comicOriginUrl = computed(() => {
  if (!infos.value?.publicUrl || !config) {
    return ''
  }

  return getComicOriginUrl(config, infos.value.publicUrl)
})
</script>

<template>
  <OrganismPageLayout
    :title="infos?.title ?? ''"
    :loading="fetchInfosLoading"
    :global-loader="fetchInfosLoading"
    sticky-header
    :back-routes="[{
      to: `/browse/sources/${sourceId}`,
      name: $t(config!.nameKey),
      image: config!.image }]"
  >
    <div
      v-if="infos"
      :class="smAndDown ? 'd-flex flex-column ga-8 px-6' : 'comic-infos-grid'"
    >
      <div v-if="!smAndDown" class="d-flex flex-column ga-6">
        <VImg
          v-if="coverSrc"
          :src="coverSrc"
          rounded="lg"
          :aspect-ratio="2 / 3"
          cover
          :lazy-src="defaultCover"
          class="border-thin rounded-lg flex-grow-0"
          @error="coverSrc = defaultCover"
        />

        <SourcesComicStatusInfos
          v-if="!smAndDown"
          v-model="infos"
          :source="sourceId"
          :comic-origin-url="comicOriginUrl"
        />
      </div>

      <div class="d-flex flex-column ga-6 w-100">
        <div class="d-flex ga-4 align-center">
          <VImg
            v-if="coverSrc && smAndDown"
            :src="coverSrc"
            rounded="lg"
            :aspect-ratio="2 / 3"
            cover
            :width="120"
            :lazy-src="defaultCover"
            class="border-thin rounded-lg flex-0-0"
            @error="coverSrc = defaultCover"
          />

          <span class="font-title font-weight-bold" :class="{ 'text-display-medium': !smAndDown, 'text-title-large': smAndDown }">{{ infos.title }}</span>
        </div>
        <SourcesComicStatusInfos
          v-if="smAndDown"
          v-model="infos"
          :source="sourceId"
          :comic-origin-url="comicOriginUrl"
        />
        <SourcesComicGeneralInfos :comic="infos" />
        <SourcesComicChapters
          :source="sourceId"
          :slug="route.params.slug as string"
          :comic-origin-url="comicOriginUrl"
        />
      </div>
    </div>
  </OrganismPageLayout>
</template>

<style scoped lang="scss">
.comic-infos-grid {
  display: grid;
  align-items: start;
  gap: 3rem;
  grid-template-columns: minmax(0, 1fr) minmax(0, 2fr);
}
</style>
