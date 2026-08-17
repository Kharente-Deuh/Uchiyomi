<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { BottomNavigationItemProps } from './Item.vue'

defineProps<{
  items: BottomNavigationItemProps[]
}>()
</script>

<template>
  <footer
    class="position-fixed bottom-0 w-100 nvagation-bottom h-100 bg-surface border-thin elevation-down d-flex justify-space-evenly"
    style="border-top-left-radius: 12px; border-top-right-radius: 12px;"
  >
    <OrganismBottomNavigationItem
      v-for="(item, index) in items"
      :key="index"
      v-bind="item"
    />
  </footer>
</template>

<style lang="scss">
.nvagation-bottom {
  z-index: 1004;
  height: var(--bottom-navigation-height);
}
</style>
