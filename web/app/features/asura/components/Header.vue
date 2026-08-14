<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
const { status, type, sort, search, isLoading } = useAsuraSearch({ doSearch: false })

const { mdAndDown } = useDisplay()
</script>

<template>
  <div
    class="d-flex ga-4"
    :class="{
      'flex-column': mdAndDown,
      'px-4': mdAndDown,
      'align-center': !mdAndDown,
      'justify-space-between': !mdAndDown,
    }"
  >
    <AtomInputSearch v-model="search" :style="{ 'max-width': mdAndDown ? '100%' : '25rem' }" />

    <div class="d-flex ga-4" :class="{ 'justify-space-between': mdAndDown }">
      <AsuraInputSort v-model="sort" :disabled="isLoading" />
      <ComicsInputStatus v-model="status" :disabled="isLoading" />
      <ComicsInputType v-model="type" :disabled="isLoading" />
    </div>
  </div>
</template>
