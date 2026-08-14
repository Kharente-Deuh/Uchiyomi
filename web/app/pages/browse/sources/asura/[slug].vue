<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { AsuraComicChapter, AsuraComicInfos } from '~/features/asura/types'
import defaultCover from '~/assets/images/default/comic-cover.webp'
import asuraImg from '~/assets/images/sources/asurascans.webp'
import { ASURA_SCANS_URL, ASURA_SOURCE_NAME } from '~/constants'
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
const comicsApi = createComicsApi()

const showDeleteModal = ref(false)

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

function toggleLibraryAction(): void {
  if (!infos.value) {
    return
  }

  if (infos.value.internalId) {
    showDeleteModal.value = true
  } else {
    addComicToLibrary()
  }
}

const addToLibraryLoading = ref(false)

async function addComicToLibrary(): Promise<void> {
  if (!infos.value) {
    return
  }

  addToLibraryLoading.value = true

  const res = await comicsApi.create({ source: ASURA_SOURCE_NAME, slug: infos.value.slug })
  if (res.success) {
    infos.value.internalId = res.data.id
  } else {
    console.error('comicsApi.create', res.error)
    toast.error(t('error.unknown'))
  }

  addToLibraryLoading.value = false
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
    <AsuraModalDelete
      v-if="infos"
      v-model="showDeleteModal"
      :comic="infos"
      @update:comic="infos = { ...infos, internalId: $event?.internalId }"
    />

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

        <div
          v-if="!smAndDown"
          class="d-flex flex-column ga-4 pa-4 bg-surface border-thin"
          style="border-radius: 12px"
        >
          <div class="d-flex ga-4 flex-wrap align-center justify-space-between">
            <div class="d-flex ga-4 flex-wrap">
              <ComicsChipStatus :status="infos.status" />
              <ComicsChipType :type="infos.type" />
            </div>
            <AtomLink
              :to="`${ASURA_SCANS_URL}${infos.publicUrl}`"
              new-tab
              external
              no-prefetch
            >
              <VIcon
                v-tooltip="$t('sources.comic.viewOnSource')"
                icon="fa6-solid:arrow-up-right-from-square"
              />
            </AtomLink>
          </div>

          <VBtn
            variant="tonal"
            class="w-100"
            :color="infos.internalId ? 'error' : 'primary'"
            :class="{
              'border-thin-error': infos.internalId,
              'border-thin-primary': !infos.internalId,
            }"
            :prepend-icon="infos.internalId ? 'fa6-solid:trash' : 'fa6-solid:book'"
            size="large"
            :text="infos.internalId ? $t('comics.remove.title') : $t('sources.add.title')"
            :loading="addToLibraryLoading"
            @click="toggleLibraryAction"
          />
          <div v-if="infos.author" class="d-flex ga-2 justify-space-between">
            <span class="text-body-large text-medium-emphasis text-uppercase text-truncate">{{ $t('comic.fields.author') }}</span>
            <span class="text-body-large font-weight-bold text-truncate">{{ infos.author }}</span>
          </div>
          <div v-if="infos.artist" class="d-flex ga-2 justify-space-between">
            <span class="text-body-large text-medium-emphasis text-uppercase text-truncate">{{ $t('comic.fields.artist') }}</span>
            <span class="text-body-large font-weight-bold text-truncate">{{ infos.artist }}</span>
          </div>
          <template v-if="infos.genres.length">
            <VDivider />
            <div class="d-flex flex-wrap ga-3">
              <span
                v-for="(genre, i) in infos.genres"
                :key="i"
                class="text-capitalize px-2 py-1 bg-background border-thin"
                style="border-radius: 12px;"
              >{{ genre }}</span>
            </div>
          </template>
        </div>
      </div>
    </div>
  </OrganismPageLayout>
</template>
