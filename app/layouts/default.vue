<script setup lang="ts">
// SPDX-License-Identifier: AGPL-3.0-or-later
import type { RouteLocationNormalized } from 'vue-router'
import type { BottomNavigationItemProps } from '~/components/Organism/BottomNavigation/Item.vue'
import type { NavigationDrawerListProps } from '~/components/Organism/NavigationDrawer/List/index.vue'
import { useDisplay } from 'vuetify'

const { mobile } = useDisplay()
const { logout } = useAuth()
const { messages } = useToast()

async function onLogout(): Promise<void> {
  const res = await logout()
  if (res.success) {
    await navigateTo('/login')
  }
}

const { t } = useI18n()
const navigationDrawerItems = computed((): NavigationDrawerListProps[] => [
  {
    title: t('nav.system.title'),
    items: [
      {
        title: t('settings.title'),
        to: '/settings',
        icon: 'fa6-solid:gear',
        isActiveFn: (route: RouteLocationNormalized) => route.path === '/settings',
      },
    ],
  },
])

const bottomNavigationItems = computed((): BottomNavigationItemProps[] => [
  {
    to: '/',
    icon: 'fa6-solid:house',
    isActiveFn: (route: RouteLocationNormalized) => route.path === '/',
  },
  {
    to: '/settings',
    icon: 'fa6-solid:gear',
    isActiveFn: (route: RouteLocationNormalized) => route.path.startsWith('/settings'),
  },
])
</script>

<template>
  <VApp>
    <VMain
      :style="{
        ...(mobile
          ? { '--v-layout-bottom': 'var(--bottom-navigation-height)' }
          : { '--v-layout-left': 'var(--navigation-drawer-width)' }),
      }"
    >
      <OrganismNavigationDrawer
        v-if="!mobile"
        :items="navigationDrawerItems"
        @logout="onLogout"
      />

      <div class="pa-6 mx-auto" style="max-width: 80rem;">
        <slot />
      </div>
    </VMain>

    <OrganismBottomNavigation
      v-if="mobile"
      :items="bottomNavigationItems"
    />

    <VSnackbarQueue v-model="messages" />
  </VApp>
</template>
