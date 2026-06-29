<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
const { mobile } = useDisplay()
</script>

<template>
  <OrganismPageLayout
    :title="$t('settings.title')"
    icon="fa6-solid:gear"
    :subtitle="$t('settings.subtitle')"
  >
    <div class="d-flex flex-column" :class="mobile ? 'ga-2' : 'ga-4'">
      <SettingsAccount />
      <SettingsAppearance />
      <SettingsSensitiveContent />
    </div>
  </OrganismPageLayout>
</template>
