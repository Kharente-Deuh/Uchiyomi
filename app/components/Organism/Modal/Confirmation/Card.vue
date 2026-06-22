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
    class="d-flex flex-column ga-6 pa-4"
    :style="{ borderRadius: !mobile ? '12px' : '12px 12px 0 0' }"
  >
    <span class="text-center">{{ text }}</span>

    <slot />
    <div class="d-flex justify-space-between w-100">
      <VBtn
        :disabled="loading"
        :text="$t('actions.cancel')"
        color="error"
        variant="text"
        @click="$emit('cancel')"
      />

      <VBtn
        :loading
        class="border-thin-primary"
        :text="$t('actions.continue')"
        @click="$emit('confirm')"
      />
    </div>
  </VCard>
</template>
