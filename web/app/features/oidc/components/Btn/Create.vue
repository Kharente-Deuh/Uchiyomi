<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
defineProps<{ disabled: boolean }>()
const showModal = ref(false)
</script>

<template>
  <OidcModalCreate v-if="!disabled" v-model="showModal" />

  <VBtn
    :disabled
    color="primary"
    variant="tonal"
    class="border-thin-primary"
    icon="fa6-solid:plus"
    density="comfortable"
    @click="showModal = true"
  />
</template>
