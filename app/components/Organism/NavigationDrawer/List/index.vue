<script setup lang="ts">
import type { NavigationDrawerListItemProps } from './Item.vue'

export interface NavigationDrawerListProps {
  title: string
  items: NavigationDrawerListItemProps[]
}

defineProps<NavigationDrawerListProps>()
</script>

<template>
  <div class="d-flex flex-column ga-2">
    <span class="text-label-small text-medium-emphasis text-uppercase px-4 text-truncate w-100" style="letter-spacing: 0.08em!important;">{{ title }}</span>
    <OrganismNavigationDrawerListItem
      v-for="(item, i) in items"
      :key="i"
      v-bind="item"
    />
  </div>
</template>
