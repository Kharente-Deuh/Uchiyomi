<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { PageScale, ReaderSettings } from '../../types.ts'
import type { PageMode } from '../Input/PageMode.vue'
import type { ComicType } from '~/features/comics/types'

const settings = defineModel<ReaderSettings>({ required: true })
const type = defineModel<ComicType>('type', { required: true })

const { t } = useI18n()
const toast = useToast()
const { smAndDown } = useDisplay()
const { updateReaderSettings } = createReaderSettingsApi()

const internalSettings = ref<ReaderSettings>({ ...settings.value })
watch(settings, () => {
  internalSettings.value = { ...settings.value }
}, { immediate: true })

const updateLoading = ref(false)
async function updateSettings(): Promise<void> {
  updateLoading.value = true

  const res = await updateReaderSettings(internalSettings.value)
  if (res.success) {
    settings.value = res.data
  } else {
    console.error('updateReaderSettings', res.error)
    toast.error(t('error.unknown'))
  }

  updateLoading.value = false
}

const canUpdateSettings = computed(() => settings.value.doublePage !== internalSettings.value.doublePage
  || settings.value.pageScale !== internalSettings.value.pageScale
  || settings.value.readingMode !== internalSettings.value.readingMode)

const unwantedPageScaleOptions = computed((): PageScale[] => {
  if (internalSettings.value.readingMode === 'webtoon') {
    return ['fit-height', 'fit-screen']
  }

  return []
})

const pageMode = computed({
  get(): PageMode {
    return internalSettings.value.doublePage ? 'double' : 'single'
  },
  set(value: PageMode) {
    internalSettings.value.doublePage = value === 'double'
  },
})

const unwantedPageModeOptions = computed((): PageMode[] => {
  if (internalSettings.value.readingMode === 'webtoon') {
    return ['double']
  }

  return []
})

const DEFAULT_OPTIONS: Record<ComicType, Omit<ReaderSettings, 'type'>> = {
  manhwa: {
    readingMode: 'webtoon',
    pageScale: 'fit-width',
    doublePage: false,
  },
  mangatoon: {
    readingMode: 'webtoon',
    pageScale: 'fit-width',
    doublePage: false,
  },
  manhua: {
    readingMode: 'paged-rtl',
    pageScale: 'fit-screen',
    doublePage: false,
  },
  manga: {
    readingMode: 'webtoon',
    pageScale: 'fit-width',
    doublePage: false,
  },
}

const canRestoreDefault = computed(() => {
  const defaultOptions = DEFAULT_OPTIONS[type.value]

  return settings.value.doublePage !== defaultOptions.doublePage
    || settings.value.pageScale !== defaultOptions.pageScale
    || settings.value.readingMode !== defaultOptions.readingMode
})

const restoreDefaultLoading = ref(false)
async function restoreDefault(): Promise<void> {
  restoreDefaultLoading.value = true

  const res = await updateReaderSettings({ ...DEFAULT_OPTIONS[type.value], type: type.value })
  if (res.success) {
    settings.value = res.data
  } else {
    console.error('updateReaderSettings', res.error)
    toast.error(t('error.unknown'))
  }

  restoreDefaultLoading.value = false
}
</script>

<template>
  <VCard class="pa-4 border-thin d-flex flex-column ga-6" style="border-radius: 12px;">
    <div class="d-flex ga-2 align-center justify-space-between">
      <span class="text-title-large">{{ $t('sources.asura.type.label') }}</span>
      <div class="w-fit">
        <ComicsInputType
          v-model="type"
          hide-label
          :clearable="false"
          :disabled="updateLoading || restoreDefaultLoading"
        />
      </div>
    </div>
    <VDivider />

    <ReaderInputMode v-model="internalSettings.readingMode" :disabled="updateLoading || restoreDefaultLoading" />

    <ReaderInputPageScale
      v-model="internalSettings.pageScale"
      :disabled-options="unwantedPageScaleOptions"
      :disabled="updateLoading || restoreDefaultLoading"
    />

    <ReaderInputPageMode
      v-model="pageMode"
      :disabled-options="unwantedPageModeOptions"
      :disabled="updateLoading || restoreDefaultLoading"
    />

    <div class="d-flex justify-end ga-4 w-100 flex-wrap">
      <VBtn
        v-if="canRestoreDefault"
        :text="$t('actions.reset')"
        color="secondary"
        size="large"
        variant="outlined"
        :class="{
          'w-100': smAndDown,
        }"
        :disabled="updateLoading"
        :loading="restoreDefaultLoading"
        @click="restoreDefault"
      />

      <VBtn
        :text="$t('actions.save')"
        color="primary"
        size="large"
        :class="{
          'border-thin-primary': canUpdateSettings,
          'w-100': smAndDown,
        }"
        :disabled="!canUpdateSettings || restoreDefaultLoading"
        :loading="updateLoading"
        @click="updateSettings"
      />
    </div>
  </VCard>
</template>
