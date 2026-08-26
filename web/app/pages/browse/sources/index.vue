<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { RouteLocationRaw } from 'vue-router'
import type { PageLayoutBackRoute } from '~/components/Organism/PageLayout.vue'
import { ASURA_SOURCE_NAME, KING_OF_SHOJO_SOURCE_NAME } from '~/constants'
import { AUTHENTICATED_ROUTE_GROUP } from '~/constants/auth'

definePageMeta({
  layout: 'default',
  authGroups: [AUTHENTICATED_ROUTE_GROUP],
})

const { t } = useI18n()
const { smAndDown } = useDisplay()

const backRoutes = computed((): PageLayoutBackRoute[] => smAndDown.value
  ? [
      {
        name: t('browse.title'),
        to: '/browse',
      },
    ]
  : [])

interface Source {
  to: RouteLocationRaw
  title: string
  src: string
}

const { getSourceDetails } = useSources()
const asuraDetails = getSourceDetails('asurascans')
const kingOfShojoDetails = getSourceDetails('kingofshojo')

const sources: Source[] = [
  {
    to: `/browse/sources/${ASURA_SOURCE_NAME}`,
    title: asuraDetails.name,
    src: asuraDetails.image,
  },
  {
    to: `/browse/sources/${KING_OF_SHOJO_SOURCE_NAME}`,
    title: kingOfShojoDetails.name,
    src: kingOfShojoDetails.image,
  },
]
</script>

<template>
  <OrganismPageLayout
    :title="$t('sources.title')"
    :back-routes
    global-loader
  >
    <div class="sources-grid" :class="{ 'px-4': smAndDown }">
      <AtomLink
        v-for="(source, i) of sources"
        :key="i"
        :to="source.to"
      >
        <VCard class="border-thin d-flex ga-6 text-truncate pa-4 align-center" style="border-radius: 12px;">
          <img
            aspect-ratio="1"
            :src="source.src"
            class="border-thin rounded-lg"
            style="height: 48px;"
          >
          <span class="text-title-large text-truncate">{{ source.title }}</span>
        </VCard>
      </AtomLink>
    </div>
  </OrganismPageLayout>
</template>

<style scoped lang="scss">
.sources-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
}

@media screen and (max-width: 1100px) {
  .sources-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media screen and (max-width: 650px) {
  .sources-grid {
    grid-template-columns: 1fr;
  }
}
</style>
