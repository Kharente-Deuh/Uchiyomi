<script setup lang="ts">
import type { RouteLocationNormalized, RouteLocationRaw } from 'vue-router'

export interface BottomNavigationItemProps {
  icon: string
  to: RouteLocationRaw
  isActiveFn: (route: RouteLocationNormalized) => boolean
  baseRoute: RouteLocationRaw
}

const props = defineProps<BottomNavigationItemProps>()

const router = useRouter()

const active = computed(() => props.isActiveFn(router.currentRoute.value))
const isBaseRoute = computed(() => router.currentRoute.value.path === props.baseRoute)
</script>

<template>
  <AtomLink :to="isBaseRoute ? undefined : to">
    <div class="d-flex flec-column justify-center align-center w-fit h-100 px-4">
      <div class="d-flex justify-center w-fit">
        <VIcon
          :icon
          :color="active ? 'primary' : 'secondary'"
        />
      </div>
    </div>
  </AtomLink>
</template>
