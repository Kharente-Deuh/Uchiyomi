<script setup lang="ts">
// SPDX-License-Identifier: AGPL-3.0-or-later
const { user } = useAuth()
const showUpdatePassword = ref(false)
</script>

<template>
  <SettingsAccountModalUpdatePassword v-model="showUpdatePassword" />
  <SettingsCard
    v-if="user"
    :title="$t('settings.account.title')"
    :subtitle="$t('settings.account.subtitle')"
  >
    <SettingsCardItem :title="$t('settings.account.displayName')">
      <VTextField
        :model-value="user.displayName"
        style="max-width: 15rem !important;"
        hide-details
        density="compact"
      />
    </SettingsCardItem>
    <VDivider />
    <SettingsCardItem :title="$t('settings.account.email')">
      <VTextField
        :model-value="user.email"
        style="max-width: 15rem !important;"
        hide-details
        density="compact"
      />
    </SettingsCardItem>
    <VDivider />

    <SettingsCardItem :title="$t('settings.account.password.title')">
      <VBtn
        class="border-thin-secondary"
        variant="tonal"
        color="secondary"
        :text="$t('settings.account.password.change')"
        @click="showUpdatePassword = true"
      />
    </SettingsCardItem>
  </SettingsCard>
</template>
