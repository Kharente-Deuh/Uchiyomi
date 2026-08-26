<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
defineProps<{ loading?: boolean }>()
</script>

<template>
  <div
    v-tooltip="$t(`sources.add.title`)"
    class="d-flex flex-column items-center justify-center rounded-lg transition-smooth add-library-btn border-thin-primary"
  >
    <VProgressCircular
      v-show="loading"
      indeterminate
      width="2"
      size="x-small"
      color="primary"
    />

    <VIcon
      v-show="!loading"
      icon="fa6-solid:book"
      size="x-small"
    />
  </div>
</template>

<style lang="scss" scoped>
.add-library-btn {
  padding: 0.4rem;
  backdrop-filter: blur(2px);
  background-color: rgba(var(--v-theme-surface), 0.7);
  color: rgb(var(--v-theme-primary));

  &:hover {
    background-color: rgb(var(--v-theme-primary));
    color: rgb(var(--v-theme-surface));
  }
}
</style>
