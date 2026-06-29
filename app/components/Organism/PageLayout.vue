<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { RouteLocationRaw } from 'vue-router'

const props = defineProps<{
  title: string
  subtitle?: string
  showBackRoute?: boolean
  backRoute?: RouteLocationRaw
  loading?: boolean
  prependImage?: string
  globalLoader?: boolean
  /**
   * Pin the (mobile) default header to the top of the scroll. Publishes the
   * measured header height as the `--page-header-height` CSS var so descendants
   * can stack their own sticky layers underneath it.
   */
  stickyHeader?: boolean
}>()

useHead({ title: props.title })

const { mobile } = useDisplay()
const slots = useSlots()
const debounceLoading = useDebounce(computed(() => props.loading), 500)

const headerRef = useTemplateRef<HTMLElement>('headerRef')
// border-box so the published offset includes the header's padding — otherwise
// the layers stacked underneath pin too high and overlap it.
const { height: headerHeight } = useElementSize(headerRef, undefined, { box: 'border-box' })
const isStickyHeader = computed(() => props.stickyHeader && mobile.value)
</script>

<template>
  <div
    v-if="globalLoader ? !debounceLoading : true"
    :class="!mobile ? 'pl-8 py-4 pr-4' : ''"
    :style="isStickyHeader ? { '--page-header-height': `${headerHeight}px` } : undefined"
    class="position-relative"
  >
    <VProgressLinear
      v-if="loading"
      indeterminate
      class="position-absolute bottom-0 left-0 w-100"
    />

    <slot name="header" />
    <div
      v-if="!slots.header"
      ref="headerRef"
      class="d-flex ga-3 align-center justify-space-between"
      :class="[mobile ? 'px-6 py-3' : 'mb-8', { 'page-layout__sticky-header bg-background': isStickyHeader }]"
    >
      <div class="d-flex align-center ga-6 text-truncate">
        <AtomLink
          v-if="backRoute && (showBackRoute || mobile)"
          :to="backRoute"
        >
          <VIcon icon="fa6-solid:chevron-left" />
        </AtomLink>

        <slot name="prepend-title" />

        <div class="d-flex flex-column ga-3 text-truncate">
          <span class="text-display-small font-title text-truncate">{{ title }}</span>
          <span v-if="subtitle && !mobile && !slots.subtitle" class="text-title-small text-medium-emphasis">{{ subtitle }}</span>
          <slot name="subtitle" />
        </div>
      </div>

      <slot name="append-title" />
    </div>

    <slot />
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
}
</style>
