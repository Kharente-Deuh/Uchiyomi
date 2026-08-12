<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
withDefaults(defineProps<{
  placeholder?: string
  readonly?: boolean
  bgColor?: string
  autofocus?: boolean
}>(), {
  bgColor: 'white',
  readonly: false,
})

const search = defineModel<string | undefined>({ required: true })
</script>

<template>
  <VTextField
    v-model.trim="search"
    :placeholder="placeholder || $t('fields.search')"
    hide-details
    density="compact"
    :autofocus
    clearable
    :readonly
  >
    <template #prepend-inner>
      <VIcon
        icon="fa6-solid:magnifying-glass"
        size="x-small"
        class="mr-2"
      />
    </template>
  </VTextField>
</template>
