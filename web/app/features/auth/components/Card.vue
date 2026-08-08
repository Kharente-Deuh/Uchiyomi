<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
interface Props {
  loading?: boolean
  title: string
  subtitle?: string
  error?: string
  onSubmit: () => void
  submitText: string
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
      v-if="error"
      class="rounded-lg border-thin-error text-body-medium"
      density="compact"
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
