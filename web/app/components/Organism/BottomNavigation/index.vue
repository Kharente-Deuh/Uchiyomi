<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { BottomNavigationItemProps } from './Item.vue'

defineProps<{
  items: BottomNavigationItemProps[]
}>()
</script>

<template>
  <header class="position-fixed bottom-0 w-100 pa-4 nvagation-bottom">
    <div class="w-100 h-100 bg-surface rounded-lg border-thin elevation-down d-flex justify-space-evenly">
      <OrganismBottomNavigationItem
        v-for="(item, index) in items"
        :key="index"
        v-bind="item"
      />
    </div>
  </header>
</template>

<style lang="scss">
.nvagation-bottom {
  z-index: 1004;
  height: var(--bottom-navigation-height);
}
</style>
