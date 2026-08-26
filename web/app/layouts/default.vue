<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { RouteLocationNormalized } from 'vue-router'
import type { BottomNavigationItemProps } from '~/components/Organism/BottomNavigation/Item.vue'
import type { NavigationDrawerListProps } from '~/components/Organism/NavigationDrawer/List/index.vue'
import type { NavigationDrawerListItemProps } from '~/components/Organism/NavigationDrawer/List/Item.vue'
import { useDisplay } from 'vuetify'

const { mobile } = useDisplay()
const { logout, loading, isAdmin } = useAuth()
const { messages } = useToast()

async function onLogout(): Promise<void> {
  const endSessionUrl = await logout()

  if (endSessionUrl) {
    window.location.assign(endSessionUrl)
  } else {
    await navigateTo('/login')
  }
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
      isActiveFn: (route: RouteLocationNormalized) => route.path.startsWith('/settings/oidc'),
      baseRoute: '/settings/oidc',
    },
  ]
})

const navigationDrawerItems = computed((): NavigationDrawerListProps[] => [
  {
    title: t('reading.title'),
    items: [
      {
        title: t('feed.title'),
        to: '/feed',
        icon: 'fa6-solid:newspaper',
        isActiveFn: (route: RouteLocationNormalized) => route.path.startsWith('/feed'),
        baseRoute: '/feed',
      },
      {
        title: t('library.title'),
        to: '/library',
        icon: 'fa6-solid:book',
        isActiveFn: (route: RouteLocationNormalized) => route.path.startsWith('/library'),
        baseRoute: '/library',
      },
    ],
  },
  {
    title: t('browse.title'),
    items: [
      {
        title: t('sources.title'),
        to: '/browse/sources',
        icon: 'fa6-solid:boxes-stacked',
        isActiveFn: (route: RouteLocationNormalized) => route.path.startsWith('/browse/sources'),
        baseRoute: '/browse/sources',
      },
    ],
  },
  {
    title: t('nav.settings.title'),
    items: [
      {
        title: t('settings.reader.title'),
        to: '/settings/reader',
        icon: 'fa6-solid:book-open-reader',
        isActiveFn: (route: RouteLocationNormalized) => route.path.startsWith('/settings/reader'),
        baseRoute: '/settings/reader',
      },
      ...settingsNavItems.value,
    ],
  },
])

const bottomNavigationItems = computed((): BottomNavigationItemProps[] => [
  {
    to: '/feed',
    icon: 'fa6-solid:newspaper',
    isActiveFn: (route: RouteLocationNormalized) => route.path.startsWith('/feed') || route.query.from === 'feed',
    baseRoute: '/feed',
  },
  {
    to: '/library',
    icon: 'fa6-solid:book',
    isActiveFn: (route: RouteLocationNormalized) => route.path.startsWith('/library') || route.query.from === 'library',
    baseRoute: '/library',
  },
  {
    to: '/browse',
    icon: 'fa6-solid:compass',
    isActiveFn: (route: RouteLocationNormalized) => route.path.startsWith('/browse'),
    baseRoute: '/browse',
  },
  {
    to: '/settings',
    icon: 'fa6-solid:gear',
    isActiveFn: (route: RouteLocationNormalized) => route.path.startsWith('/settings'),
    baseRoute: '/settings',
  },
])

const { mainStyles } = useLayoutPadding()
</script>

<template>
  <VApp>
    <VMain :style="mainStyles">
      <OrganismNavigationDrawer
        v-if="!mobile"
        :items="navigationDrawerItems"
        :logout-loading="loading"
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

    <VSnackbarQueue
      v-model="messages"
      :timeout="2000"
      location="top right"
      closable
    />
  </VApp>
</template>
