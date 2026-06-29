<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { SourceDto } from '#shared/dto/extensions'

defineProps<{
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
</script>

<template>
  <div class="d-flex flex-column ga-6">
    <div class="d-flex align-center ga-8 justify-space-between w-100">
      <div class="d-flex ga-3 align-center transition-smooth justify-space-between w-100">
        <div class="d-flex ga-3 align-center">
          <span class="text-title-large font-weight-bold font-title">{{ $t('sources.title') }}</span>
          <span class="text-label-medium text-medium-emphasis pt-1" style="text-wrap: nowrap;">{{ $t('sources.enabled', { count: enabledSourcesCount, total: totalSourcesCount }) }}</span>
        </div>
        <div v-if="canManageExtensions" class="d-flex ga-3 align-center">
          <ExtensionsSourceEnabledFilterBtn v-if="sources.length > 1" v-model="showOnlyEnabledSources" />
          <ExtensionsSettingsBtn :has-settings :pkg-name />
        </div>
      </div>
    </div>

    <div class="d-flex ga-4 overflow-x-auto custom-scrollbar pb-4">
      <ExtensionsSourceListDesktopItem
        v-for="source in sources"
        v-show="showOnlyEnabledSources === undefined || showOnlyEnabledSources === source.isEnabled"
        :key="source.id"
        :can-manage="canManageExtensions"
        :source
        :loading="sourceToggleLoading.has(source.id)"
        @toggle="$emit('toggle', source.id)"
      />
    </div>
  </div>
</template>
