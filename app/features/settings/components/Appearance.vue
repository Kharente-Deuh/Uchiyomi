<script setup lang="ts">
// SPDX-License-Identifier: AGPL-3.0-or-later
import { useTheme } from 'vuetify'

const theme = useTheme()
const { t, locale, setLocale } = useI18n()

const themeModelValue = computed({
  get: (): 'light' | 'dark' => theme.global.current.value.dark ? 'dark' : 'light',
  set(value: 'light' | 'dark'): void {
    theme.change(value)
  },
})

const themeItems = computed(() => [
  { title: t('settings.appearance.theme.dark'), value: 'dark' },
  { title: t('settings.appearance.theme.light'), value: 'light' },
])

const LOCALE_OBJECT: Record<typeof locale.value, string> = {
  en: 'English',
  fr: 'Français',
}

const localeItems = Object.entries(LOCALE_OBJECT).map(([value, title]) => ({ value, title }))
</script>

<template>
  <SettingsCard :title="$t('settings.appearance.title')" :subtitle="$t('settings.appearance.subtitle')">
    <SettingsCardItem :title="$t('settings.appearance.theme.title')">
      <VSelect
        v-model="themeModelValue"
        :items="themeItems"
        style="max-width: 8rem !important;"
        class="pa-0"
        density="compact"
        hide-details
      />
    </SettingsCardItem>
    <VDivider />
    <SettingsCardItem :title="$t('settings.appearance.language')">
      <VSelect
        :model-value="locale"
        :items="localeItems"
        style="max-width: 8rem !important;"
        hide-details
        density="compact"
        @update:model-value="setLocale($event)"
      />
    </SettingsCardItem>
  </SettingsCard>
</template>
