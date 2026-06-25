<script setup lang="ts">
// SPDX-License-Identifier: AGPL-3.0-or-later
import type { RouteLocationRaw } from 'vue-router'

const props = defineProps<{
  title: string
  subtitle?: string
  backRoute?: RouteLocationRaw
  loading?: boolean
}>()

useHead({ title: props.title })

const { mobile } = useDisplay()
</script>

<template>
  <div :class="!mobile ? 'pa-6' : ''" class="position-relative">
    <VProgressLinear
      v-if="loading"
      indeterminate
      class="position-absolute bottom-0 left-0 w-100"
    />
    <div
      class="d-flex flex-column ga-3"
      :class="mobile ? 'px-6 py-3' : 'mb-8'"
    >
      <div class="d-flex align-center ga-4">
        <AtomLink
          v-if="backRoute && mobile"
          :to="backRoute"
        >
          <VIcon icon="fa6-solid:chevron-left" />
        </AtomLink>
        <span class="text-display-small font-title text-truncate">{{ title }}</span>
      </div>
      <span v-if="subtitle && !mobile" class="text-title-small text-medium-emphasis">{{ subtitle }}</span>
    </div>

    <slot />
  </div>
</template>
