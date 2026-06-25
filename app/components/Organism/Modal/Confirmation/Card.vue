<script setup lang="ts">
// SPDX-License-Identifier: AGPL-3.0-or-later
defineProps<{
  text: string
  loading?: boolean
}>()

defineEmits<{ confirm: [], cancel: [] }>()

const { mobile } = useDisplay()
</script>

<template>
  <VCard
    class="d-flex flex-column ga-6 pa-4 border-thin"
    :style="{ borderRadius: !mobile ? '12px' : '12px 12px 0 0' }"
  >
    <span class="text-center">{{ text }}</span>

    <slot />
    <div class="d-flex justify-space-between w-100">
      <VBtn
        :disabled="loading"
        class="text-body-medium px-2"
        :text="$t('actions.cancel')"
        density="comfortable"
        color="error"
        variant="text"
        @click="$emit('cancel')"
      />

      <VBtn
        :loading
        density="comfortable"
        class="border-thin-primary text-body-medium px-2"
        :text="$t('actions.continue')"
        @click="$emit('confirm')"
      />
    </div>
  </VCard>
</template>
