<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
const showModal = ref(false)
</script>

<template>
  <OidcModalCreate v-model="showModal" />

  <VBtn
    color="primary"
    variant="tonal"
    class="border-thin-primary"
    icon="fa6-solid:plus"
    density="comfortable"
    @click="showModal = true"
  />
</template>
