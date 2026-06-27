<script setup lang="ts">
// SPDX-License-Identifier: AGPL-3.0-or-later
import { useDebounceFn } from '@vueuse/core'

const { user, updateShowNsfw, loading: authLoading } = useAuth()
const showNsfw = ref<boolean>(user.value?.showNsfw ?? false)
const loading = computed(() => !user.value && authLoading.value)
watch((loading), (value) => {
  if (!value) {
    showNsfw.value = user.value?.showNsfw ?? false
  }
})

const doUpdateNsfw = useDebounceFn(async (value: boolean | null): Promise<void> => {
  if (value === null || (value === user.value?.showNsfw && !authLoading.value)) {
    return
  }

  await updateShowNsfw(value)
}, 1000)
</script>

<template>
  <SettingsCard
    v-if="user"
    :title="$t('settings.sensitiveContent.title')"
  >
    <SettingsCardItem :title="$t('settings.sensitiveContent.nsfw')">
      <VSwitch
        v-model="showNsfw"
        theme="light"
        inset
        density="compact"
        hide-details
        color="primary"
        @update:model-value="doUpdateNsfw"
      />
    </SettingsCardItem>
  </SettingsCard>
</template>
