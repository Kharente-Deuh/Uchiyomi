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
    :icon="isDark ? 'fa6-solid:moon' : 'fa6-solid:sun'"
    :aria-label="$t('actions.toggleTheme')"
    variant="text"
    @click="toggleDark"
  />
</template>
