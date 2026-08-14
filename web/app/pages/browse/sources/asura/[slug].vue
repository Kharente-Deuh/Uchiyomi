<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { AsuraComicChapter, AsuraComicInfos } from '~/features/asura/types'
import defaultCover from '~/assets/images/default/comic-cover.webp'
import asuraImg from '~/assets/images/sources/asurascans.webp'
import { AUTHENTICATED_ROUTE_GROUP } from '~/constants/auth'

definePageMeta({
  layout: 'default',
  authGroups: [AUTHENTICATED_ROUTE_GROUP],
})

const route = useRoute('browse-sources-asura-slug')
const toast = useToast()
const { t } = useI18n()
const { smAndDown } = useDisplay()
const api = createAsuraApi()

const fetchInfosLoading = ref(false)
const infos = ref<AsuraComicInfos>()

const fetchChaptersLoading = ref(false)
const chapters = ref<AsuraComicChapter[]>([])

const coverSrc = ref<string>()

onMounted(async () => {
  await Promise.all([
    fetchInfos(),
    fetchChapters(),
  ])
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
      ? t('sources.asura.comic.notFound')
      : t('error.unknown'))

    await navigateTo('/browse/sources/asura')
  }

  fetchInfosLoading.value = false
}

async function fetchChapters(): Promise<void> {
  fetchChaptersLoading.value = true

  const res = await api.getSeriesChapters(route.params.slug as string)

  if (res.success) {
    chapters.value = res.data
  } else {
    console.error('api.getSeriesChapters', res.error)
    toast.error(res.error.status === 404
      ? t('sources.asura.comic.chapters.error.fetch')
      : t('error.unknown'))
  }
}
</script>

<template>
  <OrganismPageLayout
    :title="infos?.title ?? ''"
    :loading="fetchInfosLoading"
    :global-loader="fetchInfosLoading"
    sticky-header
    :back-routes="[{
      to: '/browse/sources/asura',
      name: $t('sources.asurascans.title'),
      image: asuraImg }]"
  >
    <div
      v-if="infos"
      class="d-flex ga-4"
      :class="{ 'flex-column': smAndDown, 'px-6': smAndDown }"
    >
      <div class="d-flex flex-column ga-6 w-33">
        <VImg
          v-if="coverSrc"
          :src="coverSrc"
          rounded="lg"
          cover
          :lazy-src="defaultCover"
          width="100%"
          class="border-thin rounded-lg"
          @error="coverSrc = defaultCover"
        />

        <AsuraComicStatusInfos v-if="!smAndDown" v-model="infos" />
      </div>
    </div>
  </OrganismPageLayout>
</template>
