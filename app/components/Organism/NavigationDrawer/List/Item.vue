<script setup lang="ts">
import type { RouteLocationNormalized, RouteLocationRaw } from 'vue-router'

export interface NavigationDrawerListItemProps {
  compact?: boolean
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
      class="d-flex align-center w-100 navigation-drawer-list-item transition-smooth"
      :class="{
        'text-truncate': !compact,
        'px-4': !compact,
        'py-2': !compact,
        'ga-4': !compact,
        'px-2': compact,
        'py-3': compact,
        'justify-center': compact,
        'navigation-drawer-list-item--active': active,
        'border-thin-primary': active,
      }"
    >
      <VIcon :icon size="x-small" />
      <span v-show="!compact" class="text-label-large text-truncate ">{{ title }}</span>
    </div>
  </AtomLink>
</template>

<style lang="scss" scoped>
.navigation-drawer-list-item {
  transition: all 0.3s ease-in-out;
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
  }
}
</style>
