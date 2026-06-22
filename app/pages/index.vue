<script setup lang="ts">
// SPDX-License-Identifier: AGPL-3.0-or-later
const { t } = useI18n()

useHead({
  title: () => t('nav.home'),
})
</script>

<template>
  <MoleculeEmptyState
    icon="fa6-solid:house"
    :title="$t('nav.home')"
    :description="$t('states.empty')"
  />
</template>
