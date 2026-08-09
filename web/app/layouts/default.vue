<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { RouteLocationNormalized } from 'vue-router'
import type { BottomNavigationItemProps } from '~/components/Organism/BottomNavigation/Item.vue'
import type { NavigationDrawerListProps } from '~/components/Organism/NavigationDrawer/List/index.vue'
import type { NavigationDrawerListItemProps } from '~/components/Organism/NavigationDrawer/List/Item.vue'
import { useDisplay } from 'vuetify'

const { mobile } = useDisplay()
const { logout, logoutLoading, isAdmin } = useAuth()
const { messages } = useToast()

async function onLogout(): Promise<void> {
  await logout()
  await navigateTo('/login')
}

const { t } = useI18n()

const settingsNavItems = computed((): NavigationDrawerListItemProps[] => {
  if (!isAdmin.value) {
    return []
  }

  return [
    {
      title: t('settings.oidc.title'),
      to: '/settings/oidc',
      icon: 'fa6-regular:address-card',
      isActiveFn: (route: RouteLocationNormalized) => route.path === '/settings/oidc',
      baseRoute: '/settings/oidc',
    },
  ]
})

const navigationDrawerItems = computed((): NavigationDrawerListProps[] => [
  {
    title: t('nav.settings.title'),
    items: settingsNavItems.value,
  },
])

const bottomNavigationItems = computed((): BottomNavigationItemProps[] => [
  {
    to: '/settings',
    icon: 'fa6-solid:gear',
    isActiveFn: (route: RouteLocationNormalized) => route.path.startsWith('/settings'),
    baseRoute: '/settings',
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
        :logout-loading
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
