<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { SourceDto } from '#shared/dto/extensions'

const props = defineProps<{
  canManageExtensions: boolean
  sources: SourceDto[]
  sourceToggleLoading: Set<string>
  enabledSourcesCount: number
  totalSourcesCount: number
  hasSettings: boolean
  pkgName: string
}>()

defineEmits<{ toggle: [sourceId: string] }>()

const showOnlyEnabledSources = defineModel<boolean | undefined>('showOnlyEnabledSources', { required: true })

watch(() => props.sources, () => {
  if (props.sources.some(({ isEnabled }) => isEnabled)) {
    showOnlyEnabledSources.value = true
  }
}, { once: true })
</script>

<template>
  <div class="source-list-mobile-header bg-background border-b-thin d-flex align-center justify-space-between ga-4 px-6 py-1">
    <span class=" text-medium-emphasis text-body-small text-uppercase py-1" style="letter-spacing: 0.08em;">
      {{ $t('sources.enabled', { count: enabledSourcesCount, total: totalSourcesCount }) }}
    </span>

    <div v-if="canManageExtensions" class="d-flex ga-3 align-center mb-2">
      <ExtensionsSourceEnabledFilterBtn
        v-if="sources.length > 1 && canManageExtensions"
        v-model="showOnlyEnabledSources"
      />

      <ExtensionsSettingsBtn
        :has-settings
        :pkg-name
      />
    </div>
  </div>

  <ExtensionsSourceListMobileItem
    v-for="(source, index) in sources"
    v-show="showOnlyEnabledSources === undefined || showOnlyEnabledSources === source.isEnabled"
    :key="index"
    class="border-b-thin"
    :can-manage="canManageExtensions"
    :source
    :loading="sourceToggleLoading.has(source.id)"
    @toggle="$emit('toggle', source.id)"
  />
</template>

<style lang="scss">
.source-list-mobile-item {
  &--active {
    background-color: rgb(var(--v-theme-primary), 0.05);
    border: 1px solid rgb(var(--v-theme-primary));

    .title {
      color: rgb(var(--v-theme-primary));
    }
  }
}

.source-list-mobile-header {
  position: sticky;
  top: calc(var(--page-header-height, 0px) + var(--tabs-height, 0px));
  z-index: 5;
}
</style>
