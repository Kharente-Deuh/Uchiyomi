<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { RouteRecordName } from 'vue-router'
import type { PageLayoutBackRoute } from '~/components/Organism/PageLayout.vue'
import type { Comic } from '~/features/comics/types'
import defaultCover from '~/assets/images/default/comic-cover.webp'
import { AUTHENTICATED_ROUTE_GROUP } from '~/constants/auth'

const FEED_ROUTE_NAME: RouteRecordName = 'feed'
const LIBRARY_ROUTE_NAME: RouteRecordName = 'library'

definePageMeta({
  layout: 'default',
  authGroups: [AUTHENTICATED_ROUTE_GROUP],
})

const route = useRoute('comic-id')
const { t } = useI18n()
const toast = useToast()
const { smAndDown } = useDisplay()
const api = createComicsApi()

const isLoading = ref(true)
const comic = ref<Comic>()
const coverSrc = ref<string>()

const from = computed(() => route.query.from as string)

onMounted(() => {
  fetchComic()
})

async function fetchComic(): Promise<void> {
  isLoading.value = true

  const res = await api.getById(route.params.id)
  if (!res.success) {
    console.error('api.getById', res.error)
    toast.error(res.error.status === 404
      ? t('sources.asura.comic.notFound')
      : t('error.unknown'))

    if (from.value === FEED_ROUTE_NAME) {
      await navigateTo({ name: FEED_ROUTE_NAME })
    } else {
      await navigateTo({ name: LIBRARY_ROUTE_NAME })
    }

    return
  }

  comic.value = res.data
  coverSrc.value = res.data.cover

  isLoading.value = false
}

function onDelete(): void {
  navigateTo({ name: from.value === FEED_ROUTE_NAME ? FEED_ROUTE_NAME : LIBRARY_ROUTE_NAME })
}

const backRoutes = computed((): PageLayoutBackRoute[] => from.value === FEED_ROUTE_NAME
  ? [{
      to: '/feed',
      name: t('feed.title'),
    }]
  : [{
      to: '/library',
      name: t('library.title'),
    }])
</script>

<template>
  <OrganismPageLayout
    :title="comic?.title ?? ''"
    :loading="isLoading"
    :global-loader="isLoading"
    sticky-header
    :back-routes
  >
    <div
      v-if="comic"
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

        <ComicsStatusInfos
          v-if="!smAndDown"
          :comic
          @deleted="onDelete"
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

          <span class="font-title font-weight-bold" :class="{ 'text-display-medium': !smAndDown, 'text-title-large': smAndDown }">{{ comic.title }}</span>
        </div>
        <ComicsStatusInfos
          v-if="smAndDown"
          :comic
          @deleted="onDelete"
        />
        <ComicsGeneralInfos :comic />
        <ComicsChapters :id="route.params.id" />
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
