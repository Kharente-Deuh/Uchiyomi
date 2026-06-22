<script setup lang="ts">
// SPDX-License-Identifier: AGPL-3.0-or-later
const { t } = useI18n()

useHead({
  title: () => t('nav.settings'),
})
</script>

<template>
  <div>
    <div class="d-flex flex-column ga-3 mb-8">
      <span class="text-display-small font-title text-truncate">{{ $t('settings.title') }}</span>
      <span class="text-title-small text-medium-emphasis">{{ $t('settings.subtitle') }}</span>
    </div>

    <div class="d-flex flex-column ga-6">
      <SettingsAccount />
      <SettingsAppearance />
    </div>
  </div>
</template>
