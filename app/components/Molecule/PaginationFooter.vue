<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { useLayoutStore } from '~/store/layout.store'

defineProps<{
  pagesTotal: number
  disabled?: boolean
}>()

const layoutStore = useLayoutStore()

const page = defineModel<number>({ required: true })

onMounted(() => {
  layoutStore.setPaginationEnabled(true)
})

onBeforeUnmount(() => {
  layoutStore.setPaginationEnabled(false)
})
</script>

<template>
  <div
    class="d-flex justify-center w-100 position-fixed bottom-0 left-0 pb-4"
    style="padding-left: var(--v-layout-left); z-index: 100;"
  >
    <div class="d-flex ga-2 w-fit align-center bg-surface border-thin-primary rounded-lg">
      <VBtn
        style="border-top-right-radius: 0; border-bottom-right-radius: 0;"
        :disabled="page === 1 || disabled"
        size="small"
        :variant="page === 1 || disabled ? 'text' : 'tonal'"
        icon="fa6-solid:chevron-left"
        @click="page--"
      />
      <span class="px-1">{{ page }} / {{ pagesTotal }}</span>
      <VBtn
        style="border-top-left-radius: 0; border-bottom-left-radius: 0;"
        icon="fa6-solid:chevron-right"
        size="small"
        :disabled="page === pagesTotal || disabled"
        :variant="page === pagesTotal || disabled ? 'text' : 'tonal'"
        @click="page++"
      />
    </div>
  </div>
</template>
