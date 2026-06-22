<script setup lang="ts">
// SPDX-License-Identifier: AGPL-3.0-or-later
import { useDebounceFn } from '@vueuse/core'
import { displayNameRule } from '~/features/auth/utils/display-name'

const { user, updateDisplayName } = useAuth()
const { t } = useI18n()
const toast = useToast()

const showUpdatePassword = ref(false)

// Editable display name with debounced auto-save. The field mirrors the store;
// on a failed save we revert to the last persisted value.
const displayName = ref(user.value?.displayName ?? '')
watch(() => user.value?.displayName, (value) => {
  displayName.value = value ?? ''
})

const rule = displayNameRule(t('settings.account.displayName'))

const saveDisplayName = useDebounceFn(async (value: string): Promise<void> => {
  const current = user.value?.displayName ?? ''
  const next = value.trim()
  if (next === current) {
    return
  }

  if (!rule.isValidSync(next)) {
    toast.error(t('settings.account.displayNameError'))
    displayName.value = current

    return
  }

  const res = await updateDisplayName({ displayName: next })
  if (res.success) {
    toast.success(t('settings.account.displayNameUpdated'))
  } else {
    toast.error(t('settings.account.displayNameError'))
    displayName.value = current
  }
}, 500)
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
        v-model="displayName"
        style="max-width: 15rem !important;"
        hide-details
        density="compact"
        data-test="settings-account-display-name"
        @update:model-value="saveDisplayName"
      />
    </SettingsCardItem>
    <VDivider />
    <SettingsCardItem :title="$t('settings.account.accountName')">
      <VTextField
        :model-value="user.accountName"
        disabled
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
