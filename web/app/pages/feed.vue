<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { AUTHENTICATED_ROUTE_GROUP } from '~/constants/auth'

definePageMeta({
  layout: 'default',
  authGroups: [AUTHENTICATED_ROUTE_GROUP],
})

const { smAndDown } = useDisplay()
const { isLoading, items, page, maxPage, source, type } = useFeed()

const hasNextPage = computed(() => page.value < maxPage.value)
const loadMoreSentinel = useTemplateRef<HTMLElement>('loadMoreSentinel')
useIntersectionObserver(loadMoreSentinel, ([entry]) => {
  if (entry?.isIntersecting && smAndDown.value && !isLoading.value && hasNextPage.value) {
    page.value += 1
  }
})
</script>

<template>
  <OrganismPageLayout :loading="isLoading">
    <template #sub-header>
      <div
        v-if="smAndDown"
        class="d-flex ga-4 flex-wrap px-8 pt-3"
      >
        <ComicsInputSource v-model="source" :disabled="isLoading" />
        <ComicsInputType v-model="type" :disabled="isLoading" />
      </div>
      <div v-else class="d-flex ga-4 flex-wrap">
        <div :class="{ 'w-fit-content': !smAndDown }">
          <ComicsInputSource v-model="source" :disabled="isLoading" />
        </div>
        <div :class="{ 'w-fit-content': !smAndDown }">
          <ComicsInputType v-model="type" :disabled="isLoading" />
        </div>
      </div>
    </template>

    <div v-if="items.length === 0 && !isLoading" class="d-flex justify-center py-4 px-8">
      <span class="text-body-large font-weight-bold text-medium-emphasis">{{ $t('feed.noResults') }}</span>
    </div>
    <div
      v-else
      class="comics-grid w-100 pt-4"
      :class="{ 'px-8': smAndDown }"
    >
      <FeedComicCard
        v-for="(item, index) in items"
        :key="index"
        :item="item"
      />
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
  grid-template-columns: repeat(3, 1fr);
  gap: 2rem;
}

@media screen and (max-width: 1350px) {
  .comics-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media screen and (max-width: 850px) {
  .comics-grid {
    grid-template-columns: repeat(1, 1fr);
  }
}
</style>
