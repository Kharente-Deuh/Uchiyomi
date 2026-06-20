<script setup lang="ts">
import { useTheme } from 'vuetify'

const theme = useTheme()

const isDark = computed(() => theme.global.current.value.dark)

function toggleDark(): void {
  theme.change(isDark.value ? 'light' : 'dark')
}
</script>

<template>
  <VBtn
    :icon="isDark ? 'mdi-weather-night' : 'mdi-weather-sunny'"
    :aria-label="$t('actions.toggleTheme')"
    variant="text"
    @click="toggleDark"
  />
</template>
