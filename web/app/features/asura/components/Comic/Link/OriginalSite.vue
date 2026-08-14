<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { ASURA_SCANS_URL } from '~/constants'

defineProps<{ publicUrl: string }>()
</script>

<template>
  <AtomLink
    :to="`${ASURA_SCANS_URL}${publicUrl}`"
    new-tab
    external
    class="original-site-link-icon"
    no-prefetch
  >
    <VIcon
      v-tooltip="$t('sources.comic.viewOnSource')"
      icon="fa6-solid:arrow-up-right-from-square"
      class="original-site-link-icon transition-smooth"
    />
  </AtomLink>
</template>

<style lang="scss" scoped>
.original-site-link-icon {
  &:hover {
    color: rgb(var(--v-theme-primary));
  }
}
</style>
