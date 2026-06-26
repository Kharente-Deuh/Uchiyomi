<script setup lang="ts">
// SPDX-License-Identifier: AGPL-3.0-or-later
import type { NavigationDrawerListProps } from './List/index.vue'
import { useLayoutStore } from '~/store/layout.store'

defineProps<{
  items: NavigationDrawerListProps[]
}>()

defineEmits<{ logout: [] }>()

const layoutStore = useLayoutStore()
const { mdAndDown: compact } = useDisplay()

watch(compact, (value) => {
  layoutStore.setNavigationDrawerCompact(value)
}, { immediate: true })
</script>

<template>
  <nav
    class="left-0 position-fixed h-screen py-4 pl-4 transition-smooth"
    style="z-index: 1004"
    :style="{ width: compact ? 'var(--navigation-drawer-compact-width)' : 'var(--navigation-drawer-width)' }"
  >
    <div class="d-flex flex-column h-100 justify-space-between border-thin bg-surface rounded-lg elevation-down" :class="{ 'pa-4': !compact, 'pa-3': compact }">
      <div class="d-flex flex-column" :class="{ 'ga-6': !compact, 'ga-2': compact }">
        <AtomLink to="/">
          <AtomProjectLogo :compact :class="{ 'mb-6': compact }" />
        </AtomLink>

        <OrganismNavigationDrawerList
          v-for="(item, i) in items"
          :key="i"
          :compact
          :class="{ 'mt-2': i === 0 && !compact }"
          v-bind="item"
        />
      </div>

      <VBtn
        variant="tonal"
        class="border-thin"
        color="secondary"
        :size="compact ? 'small' : undefined"
        :prepend-icon="compact ? undefined : 'fa6-solid:right-from-bracket'"
        :icon="compact ? 'fa6-solid:right-from-bracket' : undefined"
        :text="compact ? undefined : $t('auth.signOut')"
        data-test="user-menu-logout"
        @click="$emit('logout')"
      />
    </div>
  </nav>
</template>
