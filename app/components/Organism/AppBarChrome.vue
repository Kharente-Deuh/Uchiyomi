<script setup lang="ts">
// SPDX-License-Identifier: AGPL-3.0-or-later
</script>

<template>
  <VAppBar flat density="comfortable">
    <VAppBarTitle>{{ $t('app.name') }}</VAppBarTitle>
    <template #append>
      <slot />
    </template>
  </VAppBar>
</template>
