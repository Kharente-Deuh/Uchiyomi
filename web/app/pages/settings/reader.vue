<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { ReaderSettings } from '~/features/reader/types'
import { AUTHENTICATED_ROUTE_GROUP } from '~/constants/auth'

definePageMeta({
  layout: 'default',
  authGroups: [AUTHENTICATED_ROUTE_GROUP],
})

const { t } = useI18n()
const toast = useToast()
const { getReaderSettings, updateReaderSettings } = createReaderSettingsApi()

const isLoading = ref(false)
const readerSettings = ref<ReaderSettings[]>([])

onMounted(() => {
  fetchReaderSettings()
})

async function fetchReaderSettings(): Promise<void> {
  isLoading.value = true
  const res = await getReaderSettings()
  if (res.success) {
    readerSettings.value = res.data
  } else {
    console.error('getReaderSettings', res.error)
    toast.error(t('error.unknown'))
  }

  isLoading.value = false
}
</script>

<template>
  <OrganismPageLayout :title="$t('settings.reader.title')" :loading="isLoading" />
</template>
