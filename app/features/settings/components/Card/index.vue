<script setup lang="ts">
// SPDX-License-Identifier: AGPL-3.0-or-later
const props = defineProps<{
  title: string
  subtitle?: string
  adminOnly?: boolean
  loading?: boolean
}>()

const { mobile } = useDisplay()
</script>

<template>
  <SettingsCardMobile
    v-if="mobile"
    :loading
    :disabled="loading"
    :title="title"
    :admin-only="adminOnly"
  >
    <template #append-title>
      <slot name="append-title" />
    </template>
    <slot />
  </SettingsCardMobile>
  <SettingsCardDesktop
    v-else
    v-bind="props"
  >
    <template #append-title>
      <slot name="append-title" />
    </template>

    <slot />
  </SettingsCardDesktop>
</template>
