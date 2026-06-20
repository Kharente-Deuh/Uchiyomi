<script setup lang="ts">
import { useTheme } from 'vuetify'

const theme = useTheme()
const { t } = useI18n()

const isDark = computed(() => theme.global.current.value.dark)

function toggleDark() {
  theme.change(isDark.value ? 'light' : 'dark')
}
</script>

<template>
  <VBtn
    :icon="isDark ? 'mdi-weather-night' : 'mdi-weather-sunny'"
    :aria-label="t('actions.toggleTheme')"
    variant="text"
    @click="toggleDark"
  />
</template>
