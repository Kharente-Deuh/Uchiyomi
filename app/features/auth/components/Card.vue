<script setup lang="ts">
// SPDX-License-Identifier: AGPL-3.0-or-later
interface Props {
  loading?: boolean
  title: string
  subtitle?: string
  error?: string
  onSubmit: () => void
  submitText: string
  notice?: string
}

defineProps<Props>()

const { mobile } = useDisplay()
</script>

<template>
  <div
    :class="!mobile ? ['rounded-lg', 'border-thin', 'bg-surface', 'rounded-xl'] : []"
    class="pa-4 d-flex flex-column ga-6"
  >
    <div class="text-center d-flex flex-column">
      <span class="text-display-small font-weight-bold font-title">
        {{ title }}
      </span>
      <span class="text-label-large text-medium-emphasis">
        {{ subtitle }}
      </span>
    </div>

    <VAlert
      v-if="notice"
      type="info"
      class="rounded-lg"
      variant="tonal"
      data-test="auth-form-notice"
    >
      {{ notice }}
    </VAlert>
    <VAlert
      v-if="error"
      class="rounded-lg"
      type="error"
      variant="tonal"
      data-test="auth-form-error"
    >
      {{ error }}
    </VAlert>

    <VForm class="d-flex flex-column ga-6" @submit.prevent="onSubmit">
      <slot />
      <VBtn
        type="submit"
        size="large"
        class="w-100 border-thin"
        style="border-color: rgba(var(--v-theme-primary), 0.3) !important;"
        :loading
        :text="submitText"
        color="primary"
      />
    </VForm>

    <slot name="footer" />
  </div>
</template>
