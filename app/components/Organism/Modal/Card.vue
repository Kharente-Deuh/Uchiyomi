<script setup lang="ts">
// SPDX-License-Identifier: AGPL-3.0-or-later
export interface ModalCardProps {
  loading?: boolean
  prependIcon?: string
  title?: string
  isFormComplete?: boolean
  globalLoading?: boolean
  noActions?: boolean
  submitText?: string
  cancelText?: string
}

withDefaults(defineProps<ModalCardProps>(), { isFormComplete: true })
defineEmits<{ close: [], submit: [], cancel: [] }>()

const { mobile } = useDisplay()
</script>

<template>
  <VCard
    class="position-relative custom-scrollbar"
    :style="{ borderRadius: !mobile ? '12px' : '12px 12px 0 0' }"
  >
    <VCardTitle
      v-if="title"
      class="d-flex justify-space-between ga-4 text-truncate align-center position-sticky bg-surface top-0 text-raleway border-b-thin py-3 elevation-down"
      style="z-index: 100;"
    >
      <div class="d-flex ga-4 text-truncate align-center">
        <VIcon
          v-if="prependIcon"
          :icon="prependIcon"
          color="primary"
          class="text-body-large"
        />

        <h3 class="text-truncate text-title-large font-weight-bold">
          {{ title }}
        </h3>
      </div>

      <VIcon
        icon="fa6-solid:xmark"
        size="x-small"
        class="modal-card-close-btn"
        @click="$emit('close')"
      />
    </VCardTitle>

    <VProgressLinear
      v-if="globalLoading"
      indeterminate
      rounded
    />

    <VForm
      :disabled="globalLoading || loading"
      class="pt-2 position-relative"
      @submit.prevent="$emit('submit')"
    >
      <div class="d-flex flex-column ga-4 px-4 py-4">
        <slot />
      </div>

      <VCardActions
        v-if="!noActions"
        class="d-flex flex-column ga-4 position-sticky bottom-0 bg-surface pa-4 border-t-thin elevation-down"
        style="z-index: 100;"
      >
        <slot name="actions" />

        <div class="d-flex justify-space-between ga-2 w-100">
          <VBtn
            variant="text"
            :text="cancelText || $t('actions.cancel')"
            color="error"
            :disabled="globalLoading || loading"
            @click="$emit('cancel')"
          />

          <VBtn
            type="submit"
            :class="isFormComplete && !globalLoading && !loading ? 'border-thin-primary' : undefined"
            :variant="isFormComplete && !globalLoading && !loading ? 'tonal' : 'text'"
            :disabled="!isFormComplete || globalLoading"
            :loading
            :text="submitText || $t('actions.save')"
          />
        </div>
      </VCardActions>
    </VForm>
  </VCard>
</template>

<style lang="scss">
.modal-card-close-btn {
  transition: all 0.3s;

  &:hover {
    color: rgb(var(--v-theme-primary));
  }
}
</style>
