<script setup lang="ts">
// SPDX-License-Identifier: AGPL-3.0-or-later
import type { NavigationDrawerListProps } from './List/index.vue'
import LogoHorizontalDark from '~/assets/images/logo-horizontal-dark.webp'
import LogoHorizontalLight from '~/assets/images/logo-horizontal.webp'

defineProps<{
  items: NavigationDrawerListProps[]
}>()

defineEmits<{ logout: [] }>()

const theme = useTheme()
const isDark = computed(() => theme.global.current.value.dark)
</script>

<template>
  <nav class="left-0 position-fixed h-screen pa-4" style="width: var(--navigation-drawer-width); z-index: 1004;">
    <div class="d-flex flex-column h-100 justify-space-between border-thin bg-surface rounded-lg elevation-down pa-4">
      <div class="d-flex flex-column ga-6">
        <AtomLink to="/">
          <VImg :src="isDark ? LogoHorizontalDark : LogoHorizontalLight" class="w-100" />
        </AtomLink>

        <OrganismNavigationDrawerList
          v-for="(item, i) in items"
          :key="i"
          :class="{ 'mt-2': i === 0 }"
          v-bind="item"
        />
      </div>

      <VBtn
        variant="tonal"
        class="border-thin"
        color="secondary"
        prepend-icon="fa6-solid:right-from-bracket"
        :text="$t('auth.signOut')"
        data-test="user-menu-logout"
        @click="$emit('logout')"
      />
    </div>
  </nav>
</template>
