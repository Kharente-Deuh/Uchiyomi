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
  <SettingsCard sticky-title :title="$t('sources.enabled', { count: enabledSourcesCount, total: totalSourcesCount })">
    <template v-if="canManageExtensions" #append-title>
      <div class="d-flex ga-3 align-center mb-2">
        <ExtensionsSourceEnabledFilterBtn
          v-if="sources.length > 1 && canManageExtensions"
          v-model="showOnlyEnabledSources"
        />

        <ExtensionsSettingsBtn
          :has-settings
          :pkg-name
        />
      </div>
    </template>

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
  </SettingsCard>
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
</style>
