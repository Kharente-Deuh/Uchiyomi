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
      ? t('sources.asurascans.comic.notFound')
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

watch(() => comic.value?.type, async (newType, oldType) => {
  if (!comic.value) {
    return
  }

  if (!!newType && !!oldType && newType !== oldType) {
    const response = await api.updateType(comic.value.id, newType)
    if (!response.success) {
      console.error('api.updateType', response.error)
      comic.value.type = oldType
    }
  }
})
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
      class="bg-background pb-8"
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

        <div class="d-flex justify-space-between ga-4 w-100 py-2">
          <VBtn
            v-if="canRefresh"
            :text="$t('comics.refresh.title')"
            prepend-icon="fa6-solid:repeat"
            color="secondary"
            class="border-thin-secondary"
            @click="refreshComic"
          />

          <VBtn
            :text="$t('comics.remove.titleShort')"
            class="border-thin-error"
            color="error"
            prepend-icon="fa6-solid:trash"
            @click="showDeleteModal = true"
          />
        </div>

        <ComicsStatusInfos
          v-if="!smAndDown"
          :comic
          @update:type="comic.type = $event"
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

          <div class="d-flex ga-2 flex-column text-truncate justify-space-between w-100">
            <span
              class="font-title font-weight-bold text-wrap"
              :class="{
                'text-display-medium': !smAndDown,
                'text-title-large': smAndDown,
              }"
            >{{ comic.title }}</span>

            <template v-if="smAndDown">
              <div class="d-flex flex-wrap ga-3 align-center my-1">
                <ComicsIconStatus :status="comic.status" with-background />
                <ComicsChipType
                  :type="comic.type"
                  size="small"
                  updatable
                  @update:type="comic.type = $event"
                />
                <ComicsChipSource
                  :source="comic.source"
                  size="small"
                />
              </div>

              <div v-if="comic.author" class="d-flex ga-2 justify-space-between">
                <span class="text-body-medium text-medium-emphasis text-uppercase text-truncate">{{ $t('comic.fields.author') }}</span>
                <span class="text-body-medium font-weight-bold text-truncate">{{ comic.author }}</span>
              </div>
              <div v-if="comic.artist" class="d-flex ga-2 justify-space-between">
                <span class="text-body-medium text-medium-emphasis text-uppercase text-truncate">{{ $t('comic.fields.artist') }}</span>
                <span class="text-body-medium font-weight-bold text-truncate">{{ comic.artist }}</span>
              </div>
            </template>
          </div>
        </div>

        <div
          v-if="smAndDown"
          class="d-flex justify-space-between ga-4 w-100"
        >
          <VBtn
            v-if="canRefresh"
            prepend-icon="fa6-solid:repeat"
            color="secondary"
            :text="$t('comics.refresh.title')"
            class="border-thin-secondary"
            @click="refreshComic"
          />

          <VBtn
            :text="$t('comics.remove.title')"
            class="border-thin-error"
            color="error"
            prepend-icon="fa6-solid:trash"
            @click="showDeleteModal = true"
          />
        </div>

        <ComicsGeneralInfos :comic />
        <ComicsChapters
          :id="route.params.id"
          :continue="continueProgress"
          @refetch-progress="fetchProgress"
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
