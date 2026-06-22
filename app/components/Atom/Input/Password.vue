<script setup lang="ts">
// SPDX-License-Identifier: AGPL-3.0-or-later
const showPassword = ref(false)
</script>

<template>
  <VTextField
    v-bind="$attrs"
    :type="showPassword ? 'text' : 'password'"
  >
    <template #append-inner>
      <VIcon
        size="x-small"
        :icon="showPassword ? 'fa6-solid:eye-slash' : 'fa6-solid:eye'"
        @click="showPassword = !showPassword"
      />
    </template>
  </VTextField>
</template>
