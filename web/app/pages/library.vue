<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { RouteLocationNormalized, RouteLocationRaw } from 'vue-router'
import { AUTHENTICATED_ROUTE_GROUP } from '~/constants/auth'

definePageMeta({
  layout: 'default',
  authGroups: [AUTHENTICATED_ROUTE_GROUP],
})

const { smAndDown } = useDisplay()
const { resetFilters, page, maxPage, isLoading, comics } = useLibrarySearch({ doSearch: true })

const hasNextPage = computed(() => page.value < maxPage.value)
const loadMoreSentinel = useTemplateRef<HTMLElement>('loadMoreSentinel')
useIntersectionObserver(loadMoreSentinel, ([entry]) => {
  if (entry?.isIntersecting && smAndDown.value && !isLoading.value && hasNextPage.value) {
    page.value += 1
  }
})

onBeforeRouteLeave((to: RouteLocationNormalized) => {
  if (to.name !== 'comic-id' as unknown as RouteLocationRaw) {
    resetFilters()
  }
})
</script>

<template>
  <OrganismPageLayout :loading="isLoading">
    <template #sub-header>
      <LibraryHeader />
    </template>

    <div class="pt-6" :class="{ 'px-4': smAndDown, 'pt-3': smAndDown }">
      <div v-if="!isLoading && comics.length === 0" class="d-flex justify-center">
        <span class="text-medium-emphasis text-body-large"> {{ $t('errors.noResults') }} </span>
      </div>

      <div class="comics-grid">
        <LibraryComicCard
          v-for="(comic, i) in comics"
          :key="i"
          :comic
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
