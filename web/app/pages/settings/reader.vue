<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { ComicType } from '~/features/comics/types'
import type { ReaderSettings } from '~/features/reader/types'
import { AUTHENTICATED_ROUTE_GROUP } from '~/constants/auth'

definePageMeta({
  layout: 'default',
  authGroups: [AUTHENTICATED_ROUTE_GROUP],
})

const { t } = useI18n()
const toast = useToast()
const { smAndDown } = useDisplay()
const { getReaderSettings } = createReaderSettingsApi()

const isLoading = ref(false)
const readerSettings = ref<ReaderSettings[]>([])

const comicType = ref<ComicType>('manhwa')
const currentSettings = computed({
  get: () => readerSettings.value.find(settings => settings.type === comicType.value),
  set: (value: ReaderSettings) => {
    if (value) {
      updateSettings(value)
    }
  },
})

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

async function updateSettings(settings: ReaderSettings): Promise<void> {
  const i = readerSettings.value.findIndex(settings => settings.type === comicType.value)
  if (i === -1) {
    return
  }

  readerSettings.value[i] = settings
}
</script>

<template>
  <OrganismPageLayout
    :title="$t('settings.reader.title')"
    :loading="isLoading"
    :global-loader="isLoading"
  >
    <div class="d-flex flex-column ga-6" :class="{ 'px-8': smAndDown }">
      <ReaderCard
        v-if="currentSettings"
        v-model="currentSettings"
        v-model:type="comicType"
      />
    </div>
  </OrganismPageLayout>
</template>
