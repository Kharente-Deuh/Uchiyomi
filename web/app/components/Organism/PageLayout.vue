<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { RouteLocationRaw } from 'vue-router'

export interface PageLayoutBackRoute {
  to: RouteLocationRaw
  name: string
  image?: string
}

const props = defineProps<{
  title?: string
  backRoutes?: PageLayoutBackRoute[]
  subtitle?: string
  loading?: boolean
  prependImage?: string
  globalLoader?: boolean
  stickyHeader?: boolean
}>()

useHead({ title: props.title })

const { mobile, smAndDown } = useDisplay()
const slots = useSlots()
const debounceLoading = useDebounce(computed(() => props.loading), 500)

const headerRef = useTemplateRef<HTMLElement>('headerRef')
const { height: headerHeight } = useElementSize(headerRef, undefined, { box: 'border-box' })
</script>

<template>
  <VProgressLinear
    v-if="loading"
    indeterminate
    color="primary"
    class="position-absolute rounded-lg bottom-0 left-0 w-100"
  />

  <div
    v-if="globalLoader ? !debounceLoading : true"
    :class="{ 'px-8': !smAndDown }"
    :style="`--page-header-height: ${headerHeight}px`"
    class="position-relative h-screen h-100"
  >
    <div
      ref="headerRef"
      class="d-flex flex-column ga-2 bg-background page-layout__sticky-header"
    >
      <div
        v-if="title"
        class="d-flex ga-3 align-center justify-space-between"
        :class="[smAndDown ? 'px-6 pb-3' : 'mb-8']"
      >
        <div class="d-flex align-center ga-4 text-truncate">
          <div v-if="backRoutes?.length" class="d-flex ga-4 align-center">
            <AtomLink v-if="mobile" :to="backRoutes.at(-1)?.to">
              <VIcon icon="fa6-solid:angle-left" size="large" />
            </AtomLink>
            <template v-else>
              <template
                v-for="({ to, name, image }, i) of backRoutes"
                :key="i"
              >
                <AtomLink :to>
                  <div class="d-flex align-center ga-4 text-primary-hover text-underline-hover text-truncate">
                    <img
                      v-if="image"
                      aspect-ratio="1"
                      :src="image"
                      class="border-thin rounded-lg"
                      style="height: 42px;"
                    >
                    <span
                      class="font-title text-truncate"
                      :class="{
                        'text-title-large': !mobile,
                        'text-title-medium': mobile,
                      }"
                    >{{ name }}</span>
                  </div>
                </AtomLink>
                <span class="opacity-50 text-title-large">/</span>
              </template>
            </template>
          </div>

          <slot name="prepend-title" />

          <div class="d-flex flex-column ga-3 text-truncate">
            <span class="text-display-small font-title text-truncate">{{ title }}</span>
            <span v-if="subtitle && !mobile && !slots.subtitle" class="text-title-small text-medium-emphasis">{{ subtitle }}</span>
            <slot name="subtitle" />
          </div>
        </div>

        <slot name="append-title" />
      </div>
      <slot name="sub-header" />
    </div>

    <slot />

    <slot name="footer" />
  </div>
  <div v-else class="d-flex flex-column justify-center w-100 h-screen">
    <div class="d-flex justify-center w-100">
      <VProgressCircular
        indeterminate
        color="primary"
        size="48"
        class="mb-4"
      />
    </div>
  </div>
</template>

<style lang="scss" scoped>
.page-layout__sticky-header {
  position: sticky;
  top: 0;
  z-index: 20;
  padding-bottom: 12px;
  padding-top: var(--page-header-padding-top);
}
</style>
