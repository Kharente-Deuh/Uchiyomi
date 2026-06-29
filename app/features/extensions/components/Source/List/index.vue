<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { SourceDto } from '#shared/dto/extensions'

const props = defineProps<{
  canManageExtensions: boolean
  sources: SourceDto[]
  sourceToggleLoading: Set<string>
  hasSettings: boolean
  pkgName: string
}>()

defineEmits<{ toggle: [sourceId: string] }>()

const { mobile } = useDisplay()
const showOnlyEnabledSources = ref<boolean>()

const enabledSourcesCount = computed(() => props.sources.filter(({ isEnabled }) => isEnabled).length)

watch(() => props.sources, () => {
  if (props.sources.some(({ isEnabled }) => isEnabled)) {
    showOnlyEnabledSources.value = true
  }
}, { once: true })
</script>

<template>
  <ExtensionsSourceListMobile
    v-if="mobile"
    v-bind="props"
    v-model:show-only-enabled-sources="showOnlyEnabledSources"
    :enabled-sources-count="enabledSourcesCount"
    :total-sources-count="sources.length"
    @toggle="$emit('toggle', $event)"
  />
  <ExtensionsSourceListDesktop
    v-else
    v-bind="props"
    v-model:show-only-enabled-sources="showOnlyEnabledSources"
    :enabled-sources-count="enabledSourcesCount"
    :total-sources-count="sources.length"
    @toggle="$emit('toggle', $event)"
  />
</template>
