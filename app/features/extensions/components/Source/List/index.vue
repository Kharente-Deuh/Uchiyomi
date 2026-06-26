<script setup lang="ts">
import type { SourceDto } from '#shared/dto/extensions'

const props = defineProps<{
  canManageExtensions: boolean
  sources: SourceDto[]
  sourceToggleLoading: Set<string>
}>()

defineEmits<{
  toggle: [sourceId: string]
}>()

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
    v-model:show-only-enabled-sources="showOnlyEnabledSources"
    :sources
    :source-toggle-loading="sourceToggleLoading"
    :enabled-sources-count="enabledSourcesCount"
    :total-sources-count="sources.length"
    :can-manage-extensions="canManageExtensions"
    @toggle="$emit('toggle', $event)"
  />
  <ExtensionsSourceListDesktop
    v-else
    v-model:show-only-enabled-sources="showOnlyEnabledSources"
    :sources
    :source-toggle-loading="sourceToggleLoading"
    :enabled-sources-count="enabledSourcesCount"
    :total-sources-count="sources.length"
    :can-manage-extensions="canManageExtensions"
    @toggle="$emit('toggle', $event)"
  />
</template>
