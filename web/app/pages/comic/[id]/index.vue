<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { RouteRecordName } from 'vue-router'
import type { PageLayoutBackRoute } from '~/components/Organism/PageLayout.vue'
import type { Comic, ComicProgressContinue } from '~/features/comics/types'
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
const continueProgress = ref<ComicProgressContinue>()
const coverSrc = ref<string>()

const from = computed(() => route.query.from as string)

onMounted(async () => {
  isLoading.value = true

  await Promise.all([
    fetchComic(),
    fetchProgress(),
  ])

  isLoading.value = false
})

async function fetchComic(): Promise<void> {
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
}

async function fetchProgress(): Promise<void> {
  const res = await api.getProgress(route.params.id)
  if (!res.success) {
    console.error('api.getProgress', res.error)
    toast.error(t('error.unknown'))

    return
  }

  continueProgress.value = res.data.continue
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

const canRefresh = computed(() => comic.value
  && comic.value.status !== 'hiatus'
  && comic.value.status !== 'completed'
  && comic.value.status !== 'cancelled')

const refreshLoading = ref(false)
async function refreshComic(): Promise<void> {
  if (!comic.value) {
    return
  }

  refreshLoading.value = true

  const response = await api.refreshById(comic.value.id)
  if (response.success) {
    comic.value = response.data
  } else {
    console.error(response.error)
    toast.error(t('error.unknown'))
  }

  refreshLoading.value = false
}

const showDeleteModal = ref(false)
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
      <ComicsModalDelete
        v-model="showDeleteModal"
        :comic
        @deleted="onDelete"
      />

      <div v-if="!smAndDown" class="d-flex flex-column ga-4">
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

        <div class="d-flex justify-space-between ga-4 w-100 pa-2">
          <VBtn
            v-if="canRefresh"
            v-tooltip="$t('comics.refresh.title')"
            icon="fa6-solid:repeat"
            color="secondary"
            size="small"
            class="border-thin-secondary"
            @click="refreshComic"
          />

          <VBtn
            v-tooltip="$t('comics.remove.title')"
            class="border-thin-error"
            color="error"
            icon="fa6-solid:trash"
            size="small"
            @click="showDeleteModal = true"
          />
        </div>

        <ComicsStatusInfos
          v-if="!smAndDown"
          :comic
        />
      </div>

      <div class="d-flex flex-column w-100" :class="smAndDown ? 'ga-4' : 'ga-6'">
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

          <span
            class="font-title font-weight-bold"
            :class="{
              'text-display-medium': !smAndDown,
              'text-title-large': smAndDown,
            }"
          >{{ comic.title }}</span>
        </div>

        <div
          v-if="smAndDown"
          class="d-flex justify-space-between ga-4 w-100 pa-2"
        >
          <VBtn
            v-if="canRefresh"
            v-tooltip="$t('comics.refresh.title')"
            icon="fa6-solid:repeat"
            color="secondary"
            size="small"
            class="border-thin-secondary"
            @click="refreshComic"
          />

          <VBtn
            v-tooltip="$t('comics.remove.title')"
            class="border-thin-error"
            color="error"
            icon="fa6-solid:trash"
            size="small"
            @click="showDeleteModal = true"
          />
        </div>

        <ComicsStatusInfos
          v-if="smAndDown"
          :comic
        />
        <ComicsGeneralInfos :comic />
        <ComicsChapters :id="route.params.id" :continue="continueProgress" />
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
