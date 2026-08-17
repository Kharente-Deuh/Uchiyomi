<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<script setup lang="ts">
import { useLibrarySearch } from '../composables/library-search.composable'

const { search, sort, status, type, source, isLoading } = useLibrarySearch({ doSearch: false })

const { mdAndDown } = useDisplay()

const show = ref(!mdAndDown.value)
</script>

<template>
  <div
    class="d-flex flex-column ga-4"
    :class="{ 'px-4': mdAndDown }"
  >
    <div class="d-flex align-center justify-space-between ga-4">
      <AtomInputSearch v-model="search" :style="{ 'max-width': mdAndDown ? '100%' : '45rem' }" />
      <VIcon
        :icon="show ? 'fa6-solid:angle-up' : 'fa6-solid:angle-down'"
        color="primary"
        @click="show = !show"
      />
    </div>

    <VExpandTransition>
      <div
        v-show="show"
        class="d-flex ga-4 flex-wrap"
        :class="{ 'justify-space-between': mdAndDown }"
      >
        <LibraryInputSource v-model="source" :disabled="isLoading" />
        <LibraryInputSort v-model="sort" :disabled="isLoading" />
        <ComicsInputStatus v-model="status" :disabled="isLoading" />
        <ComicsInputType v-model="type" :disabled="isLoading" />
      </div>
    </VExpandTransition>
  </div>
</template>
