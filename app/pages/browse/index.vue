<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { IntermediateListItemProps } from '~/components/Molecule/IntermediateListItem/index.vue'

const { t } = useI18n()

const items: IntermediateListItemProps[] = [
  {
    title: t('series.title'),
    icon: 'fa6-solid:book',
    subtitle: t('browse.series.subtitle'),
    to: '/browse/series',
  },
  {
    title: t('extensions.title'),
    icon: 'fa6-solid:puzzle-piece',
    subtitle: t('browse.extensions.subtitle'),
    to: '/browse/extensions',
  },
]
</script>

<template>
  <OrganismPageLayout
    :title="$t('browse.title')"
    :subtitle="$t('browse.subtitle')"
  >
    <div class="browse-grid w-100">
      <MoleculeIntermediateListItem
        v-for="(item, i) in items"
        :key="i"
        v-bind="item"
        :last="i === items.length - 1"
      />
    </div>
  </OrganismPageLayout>
</template>

<style scoped>
.browse-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
}

@media screen and (max-width: 1080px) {
  .browse-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media screen and (max-width: 790px) {
  .browse-grid {
    grid-template-columns: 1fr;
    gap: 0rem;
  }
}
</style>
