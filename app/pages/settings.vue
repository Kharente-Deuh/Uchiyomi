<script setup lang="ts">
// SPDX-License-Identifier: AGPL-3.0-or-later
const { t } = useI18n()

useHead({
  title: () => t('nav.settings'),
})
</script>

<template>
  <EmptyState
    icon="fa6-solid:gear"
    :title="$t('nav.settings')"
    :description="$t('states.empty')"
  />
</template>
