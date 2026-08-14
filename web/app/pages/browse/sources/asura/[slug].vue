<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { AsuraComicInfos } from '~/features/asura/types'
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
      ? t('sources.asura.comic.notFound')
      : t('error.unknown'))

    await navigateTo('/browse/sources/asura')
  }

  fetchInfosLoading.value = false
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
      class="d-flex ga-8"
      :class="{ 'comic-infos-grid': !smAndDown, 'px-6': smAndDown }"
    >
      <div v-if="!smAndDown" class="d-flex flex-column ga-6 w-100">
        <VImg
          v-if="coverSrc"
          :src="coverSrc"
          rounded="lg"
          cover
          :lazy-src="defaultCover"
          class="border-thin rounded-lg"
          @error="coverSrc = defaultCover"
        />

        <AsuraComicStatusInfos v-if="!smAndDown" v-model="infos" />
      </div>

      <div class="d-flex flex-column ga-6 w-100">
        <div class="d-flex ga-4 align-center">
          <VImg
            v-if="coverSrc && smAndDown"
            :src="coverSrc"
            rounded="lg"
            cover
            :lazy-src="defaultCover"
            width="120"
            class="border-thin rounded-lg"
            @error="coverSrc = defaultCover"
          />

          <span class="font-title font-weight-bold" :class="{ 'text-display-medium': !smAndDown, 'text-title-large': smAndDown }">{{ infos.title }}</span>
        </div>
        <AsuraComicStatusInfos v-if="smAndDown" v-model="infos" />
        <AsuraComicGeneralInfos :comic="infos" />
        <AsuraComicChapters :slug="route.params.slug" />
      </div>
    </div>
  </OrganismPageLayout>
</template>

<style scoped lang="scss">
.comic-infos-grid {
  display: grid;
  gap: 2rem;
  grid-template-columns: 33% 66%;
}
</style>
