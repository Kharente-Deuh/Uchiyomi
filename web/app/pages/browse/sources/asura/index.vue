<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { PageLayoutBackRoute } from '~/components/Organism/PageLayout.vue'
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

const { isLoading, series, page, hasNextPage, addComicInLibrary, addComicInLibraryLoading } = useAsuraSearch({ doSearch: true })

const showDeleteModal = ref(false)
const comicToDelete = ref<AsuraSearchItem>()

async function doToggleComic(comic: AsuraSearchItem): Promise<void> {
  if (comic.internalId) {
    comicToDelete.value = comic
    showDeleteModal.value = true
  } else {
    await addComicInLibrary(comic)
  }
}
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
      <AsuraHeader />
    </template>

    <AsuraModalDelete v-model="showDeleteModal" v-model:comic="comicToDelete" />

    <div class="pt-6" :class="{ 'px-4': smAndDown }">
      <div v-if="!isLoading && series.length === 0" class="d-flex justify-center">
        <span class="text-medium-emphasis text-body-large"> {{ $t('errors.noResults') }} </span>
      </div>

      <div class="comics-grid">
        <AsuraComicCard
          v-for="(comic, i) in series"
          :key="i"
          :comic
          :loading="addComicInLibraryLoading[comic.slug]"
          @toggle="doToggleComic(comic)"
        />
      </div>
    </div>

    <MoleculePaginationFooter
      v-model="page"
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
