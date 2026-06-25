<script setup lang="ts">
// SPDX-License-Identifier: AGPL-3.0-or-later
import type { RouteLocationNormalized } from 'vue-router'
import type { BottomNavigationItemProps } from '~/components/Organism/BottomNavigation/Item.vue'
import type { NavigationDrawerListProps } from '~/components/Organism/NavigationDrawer/List/index.vue'
import { useDisplay } from 'vuetify'

const { mobile } = useDisplay()
const { logout } = useAuth()
const { messages } = useToast()

const authStore = useAuthStore()
const { capabilities } = storeToRefs(authStore)

async function onLogout(): Promise<void> {
  const res = await logout()
  if (res.success) {
    await navigateTo('/login')
  }
}

const { t } = useI18n()
const navigationDrawerItems = computed((): NavigationDrawerListProps[] => [
  {
    title: t('browse.title'),
    items: [
      {
        title: t('series.title'),
        to: '/browse/series',
        icon: 'fa6-solid:book',
        isActiveFn: (route: RouteLocationNormalized) => route.path.startsWith('/browse/series'),
      },
      {
        title: t('extensions.title'),
        to: '/browse/extensions',
        icon: 'fa6-solid:puzzle-piece',
        isActiveFn: (route: RouteLocationNormalized) => route.path.startsWith('/browse/extensions'),
      },
    ],
  },
  {
    title: t('nav.system.title'),
    items: [
      {
        title: t('settings.title'),
        to: '/settings',
        icon: 'fa6-solid:gear',
        isActiveFn: (route: RouteLocationNormalized) => route.path === '/settings',
      },
      ...(capabilities.value.canManageExtensions
        ? [
            {
              title: t('extensions.admin.title'),
              to: '/admin/extensions',
              icon: 'fa6-solid:screwdriver-wrench',
              isActiveFn: (route: RouteLocationNormalized) => route.path.startsWith('/admin/extensions'),
            },
          ]
        : []),
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
    to: '/browse',
    icon: 'fa6-brands:safari',
    isActiveFn: (route: RouteLocationNormalized) => route.path.startsWith('/browse'),
  },
  {
    to: '/settings',
    icon: 'fa6-solid:gear',
    isActiveFn: (route: RouteLocationNormalized) => route.path.startsWith('/settings'),
  },
])

const { bottomLayout, leftLayout } = useLayoutPadding()
</script>

<template>
  <VApp>
    <VMain
      :style="{
        '--v-layout-bottom': bottomLayout,
        '--v-layout-left': leftLayout,
      }"
    >
      <OrganismNavigationDrawer
        v-if="!mobile"
        :items="navigationDrawerItems"
        @logout="onLogout"
      />

      <div class="mx-auto" style="max-width: 80rem;">
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
