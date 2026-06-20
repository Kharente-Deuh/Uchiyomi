<script setup lang="ts">
const router = useRouter()
</script>

<template>
  <VRow justify="center" class="py-12">
    <VCol cols="12" md="6" class="text-center">
      <VIcon size="64" color="warning" icon="mdi-alert-circle-outline" class="mb-4" />
      <h1 class="text-h5 mb-4">
        404
      </h1>
      <VBtn color="primary" variant="tonal" prepend-icon="mdi-arrow-left" @click="router.back()">
        Back
      </VBtn>
    </VCol>
  </VRow>
</template>
