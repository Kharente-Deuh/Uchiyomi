<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { RouteLocationRaw } from 'vue-router'
import { AUTHENTICATED_ROUTE_GROUP } from '~/constants/auth'

definePageMeta({
  layout: 'default',
  authGroups: [AUTHENTICATED_ROUTE_GROUP],
})

const { t } = useI18n()

interface Page {
  title: string
  to: RouteLocationRaw
  icon: string
}

const pages: Page[] = [
  {
    title: t('sources.title'),
    to: '/browse/sources',
    icon: 'fa6-solid:boxes-stacked',
  },
]

const { smAndDown } = useDisplay()
</script>

<template>
  <OrganismPageLayout
    :title="$t('browse.title')"
    global-loader
  >
    <div class="browse-grid" :class="{ 'px-4': smAndDown }">
      <AtomLink
        v-for="(page, i) of pages"
        :key="i"
        :to="page.to"
      >
        <VCard class="border-thin d-flex ga-6 text-truncate pa-4 align-center" style="border-radius: 12px;">
          <VIcon :icon="page.icon" size="large" />
          <span class="text-title-large text-truncate">{{ page.title }}</span>
        </VCard>
      </AtomLink>
    </div>
  </OrganismPageLayout>
</template>

<style scoped lang="scss">
.browse-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
}

@media screen and (max-width: 1100px) {
  .browse-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media screen and (max-width: 650px) {
  .browse-grid {
    grid-template-columns: 1fr;
  }
}
</style>
