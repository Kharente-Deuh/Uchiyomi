<script setup lang="ts">
import type { RouteLocationRaw } from 'vue-router'

export interface LinkProps {
  to?: RouteLocationRaw
  newTab?: boolean
  external?: boolean
  noPrefetch?: boolean
  enableTextDecoration?: boolean
}

defineProps<LinkProps>()
</script>

<template>
  <NuxtLink
    v-if="to"
    :to
    :external
    :target="newTab ? '_blank' : undefined"
    :no-prefetch
    :class="{ 'text-decoration-none': !enableTextDecoration }"
    active-class="nuxt-link-active"
    class="cursor-pointer"
  >
    <slot />
  </NuxtLink>

  <slot v-else />
</template>

<style>
/* The following class is used to override the default NuxtLink active class with nothing */
.nuxt-link-active {
}

a:any-link {
  color: initial;
}
</style>
