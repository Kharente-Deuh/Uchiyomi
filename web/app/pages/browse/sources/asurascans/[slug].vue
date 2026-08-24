<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type AsuraScansComicChapters from '~/features/asurascans/components/Comic/Chapters/index.vue'
import type { AsuraScansComicInfos } from '~/features/asurascans/types'
import defaultCover from '~/assets/images/default/comic-cover.webp'
import asuraImg from '~/assets/images/sources/asurascans.webp'
import { ASURA_SCANS_URL, ASURA_SOURCE_NAME } from '~/constants'
import { AUTHENTICATED_ROUTE_GROUP } from '~/constants/auth'

definePageMeta({
  layout: 'default',
  authGroups: [AUTHENTICATED_ROUTE_GROUP],
})

const route = useRoute('browse-sources-asurascans-slug')
const toast = useToast()
const { t } = useI18n()
const { smAndDown } = useDisplay()
const api = createAsuraScansApi()

const fetchInfosLoading = ref(false)
const infos = ref<AsuraScansComicInfos>()

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
      ? t('sources.asurascans.comic.notFound')
      : t('error.unknown'))

    await navigateTo(`/browse/sources/${ASURA_SOURCE_NAME}`)
  }

  fetchInfosLoading.value = false
}

const comicOriginUrl = computed(() => {
  if (!infos.value?.publicUrl) {
    return ''
  }

  return `${ASURA_SCANS_URL}${infos.value?.publicUrl}`
})
</script>

<template>
  <OrganismPageLayout
    :title="infos?.title ?? ''"
    :loading="fetchInfosLoading"
    :global-loader="fetchInfosLoading"
    sticky-header
    :back-routes="[{
      to: `/browse/sources/${ASURA_SOURCE_NAME}`,
      name: $t('sources.asurascans.title'),
      image: asuraImg }]"
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

        <AsuraScansComicStatusInfos
          v-if="!smAndDown"
          v-model="infos"
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
        <AsuraScansComicStatusInfos
          v-if="smAndDown"
          v-model="infos"
          :comic-origin-url="comicOriginUrl"
        />
        <AsuraScansComicGeneralInfos :comic="infos" />
        <AsuraScansComicChapters
          :slug="route.params.slug"
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
