<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { RouteLocationNormalized } from 'vue-router'
import type { PageLayoutBackRoute } from '~/components/Organism/PageLayout.vue'
import type { AsuraScansSearchItem } from '~/features/asurascans/types'
import asuraImg from '~/assets/images/sources/asurascans.webp'
import { AUTHENTICATED_ROUTE_GROUP } from '~/constants/auth'

definePageMeta({
  layout: 'default',
  authGroups: [AUTHENTICATED_ROUTE_GROUP],
})

const { t } = useI18n()
const { smAndDown } = useDisplay()

const backRoutes = computed((): PageLayoutBackRoute[] => [
  ...(smAndDown.value
    ? [
        {
          name: t('browse.title'),
          to: '/browse',
        },
      ]
    : []),
  {
    name: t('sources.title'),
    to: '/browse/sources',
  },
])

const {
  isLoading,
  series,
  page,
  maxPage,
  addComicInLibrary,
  addComicInLibraryLoading,
  resetFilters,
} = useAsuraScansSearch({ doSearch: true })
const hasNextPage = computed(() => page.value < maxPage.value)
const loadMoreSentinel = useTemplateRef<HTMLElement>('loadMoreSentinel')
useIntersectionObserver(loadMoreSentinel, ([entry]) => {
  if (entry?.isIntersecting && smAndDown.value && !isLoading.value && hasNextPage.value) {
    page.value += 1
  }
})

const showDeleteModal = ref(false)
const comicToDelete = ref<AsuraScansSearchItem>()

async function doToggleComic(comic: AsuraScansSearchItem): Promise<void> {
  if (comic.internalId) {
    comicToDelete.value = comic
    showDeleteModal.value = true
  } else {
    await addComicInLibrary(comic)
  }
}

onBeforeRouteLeave((to: RouteLocationNormalized) => {
  if (to.name !== 'browse-sources-asurascans-slug') {
    resetFilters()
  }
})
</script>

<template>
  <OrganismPageLayout
    :title="$t('sources.asurascans.title')"
    :back-routes
    :loading="isLoading"
  >
    <template #prepend-title>
      <img
        aspect-ratio="1"
        :src="asuraImg"
        class="border-thin rounded-lg"
        style="height: 48px;"
      >
    </template>

    <template #sub-header>
      <AsuraScansHeader />
    </template>

    <AsuraScansModalDelete v-model="showDeleteModal" v-model:comic="comicToDelete" />

    <div class="pt-6" :class="{ 'px-4': smAndDown }">
      <div v-if="!isLoading && series.length === 0" class="d-flex justify-center">
        <span class="text-medium-emphasis text-body-large"> {{ $t('errors.noResults') }} </span>
      </div>

      <div class="comics-grid">
        <AsuraScansComicCard
          v-for="(comic, i) in series"
          :key="i"
          :comic
          :loading="addComicInLibraryLoading[comic.slug]"
          @toggle="doToggleComic(comic)"
        />
      </div>
    </div>

    <template v-if="smAndDown" #footer>
      <div
        ref="loadMoreSentinel"
        class="d-flex justify-center py-4"
      >
        <VProgressCircular
          v-if="isLoading && page > 1"
          color="secondary"
          indeterminate
        />
      </div>
    </template>

    <MoleculePaginationFooter
      v-if="!smAndDown"
      v-model="page"
      :pages-total="maxPage"
      :has-next-page
      fixed
    />
  </OrganismPageLayout>
</template>

<style lang="scss" scoped>
.comics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 2rem;
}

@media screen and (max-width: 768px) {
  .comics-grid {
    grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  }
}

@media screen and (max-width: 400px) {
  .comics-grid {
    grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
  }
}
</style>
