<script setup lang="ts">
import type { SourceDto } from '#shared/dto/extensions'

defineProps<{
  canManageExtensions: boolean
  sources: SourceDto[]
  sourceToggleLoading: Set<string>
  enabledSourcesCount: number
  totalSourcesCount: number
}>()

defineEmits<{
  toggle: [sourceId: string]
}>()

const showOnlyEnabledSources = defineModel<boolean | undefined>('showOnlyEnabledSources', { required: true })
</script>

<template>
  <div class="d-flex flex-column ga-6">
    <div class="d-flex align-center ga-8 justify-space-between cursor-pointer w-100">
      <div class="d-flex ga-3 align-center transition-smooth">
        <span class="text-title-large font-weight-bold font-title">{{ $t('sources.title') }}</span>
        <span class="text-label-medium text-medium-emphasis pt-1" style="text-wrap: nowrap;">{{ $t('sources.enabled', { count: enabledSourcesCount, total: totalSourcesCount }) }}</span>
        <ExtensionsSourceEnabledFilterBtn v-if="sources.length > 1 && canManageExtensions" v-model="showOnlyEnabledSources" />
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
