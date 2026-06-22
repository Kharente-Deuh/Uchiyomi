<script setup lang="ts">
import type { RouteLocationNormalized, RouteLocationRaw } from 'vue-router'

export interface NavigationDrawerListItemProps {
  icon: string
  title: string
  to: RouteLocationRaw
  isActiveFn: (route: RouteLocationNormalized) => boolean
}

const props = defineProps<NavigationDrawerListItemProps>()

const router = useRouter()

const active = computed(() => props.isActiveFn(router.currentRoute.value))
</script>

<template>
  <AtomLink :to="active ? undefined : to">
    <div
      class="d-flex align-center ga-4 text-truncate w-100 px-4 py-2 navigation-drawer-list-item"
      :class="{ 'navigation-drawer-list-item--active': active }"
    >
      <VIcon :icon size="x-small" />
      <span class="text-label-large text-truncate ">{{ title }}</span>
    </div>
  </AtomLink>
</template>

<style lang="scss" scoped>
.navigation-drawer-list-item {
  transition: all 0.3s;
  border-radius: 10px;
  opacity: 0.7;
  cursor: pointer;

  &:hover {
    opacity: 1;
    background-color: rgba(var(--v-theme-surface-variant));
  }

  &--active {
    opacity: 1;
    background-color: rgba(var(--v-theme-primary), 0.1) !important;
    color: rgb(var(--v-theme-primary)) !important;
    cursor: default;
  }
}
</style>
