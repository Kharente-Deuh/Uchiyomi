<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { ExtensionDto } from '~~/shared/dto/extensions'

const props = defineProps<{
  extension: ExtensionDto
  loading?: boolean
  firstItem?: boolean
}>()

const emits = defineEmits<{
  install: []
  uninstall: []
  update: []
}>()
const { t } = useI18n()
interface ActionBtn {
  icon: string
  color: string
  tooltip: string
  action: () => void
}

const { mobile } = useDisplay()

const actionBtn = computed((): ActionBtn | undefined => {
  if (!props.extension.isInstalled) {
    return {
      icon: 'fa6-solid:download',
      color: 'primary',
      tooltip: t('actions.install'),
      action: () => emits('install'),
    }
  }

  if (props.extension.hasUpdate) {
    return {
      icon: 'fa6-solid:rotate-right',
      color: 'warning',
      tooltip: t('actions.update'),
      action: () => emits('update'),
    }
  }

  if (mobile.value) {
    return
  }

  return {
    icon: 'fa6-solid:trash',
    color: 'error',
    tooltip: t('actions.uninstall'),
    action: () => emits('uninstall'),
  }
})
</script>

<template>
  <AtomLink :to="extension.isInstalled ? `/extensions/${extension.pkgName}` : undefined">
    <div
      class="d-flex justify-space-between ga-2 pa-2 bg-surface"
      :class="{
        'border-thin': !mobile,
        'border-b-thin': mobile,
        'border-t-thin': firstItem,
      }"
      :style="{ borderRadius: !mobile ? '12px' : undefined }"
    >
      <div class="d-flex align-center ga-2 ">
        <ExtensionsAvatar :url="extension.iconUrl ?? ''" />
        <div class="d-flex flex-column ga-1">
          <span class="text-body-large font-weight-bold text-truncate">{{ extension.name }}</span>
          <div class="d-flex align-center ga-2">
            <span class="text-label-small text-medium-emphasis text-truncate">{{ extension.lang }}</span>
            <VChip
              :text="extension.versionName"
              size="x-small"
              density="comfortable"
              class="text-label-small px-1"
              :color="extension.hasUpdate ? 'warning' : 'secondary'"
            />
            <AtomChipNsfw v-if="extension.isNsfw" />
          </div>
        </div>
      </div>

      <div class="d-flex flex-column justify-center pr-1">
        <VBtn
          v-if="actionBtn"
          v-tooltip="actionBtn.tooltip"
          size="small"
          :loading
          :icon="actionBtn.icon"
          :color="actionBtn.color"
          @click.prevent="actionBtn.action"
        />

        <VIcon
          v-else
          icon="fa6-solid:chevron-right"
          class="mr-2"
        />
      </div>
    </div>
  </AtomLink>
</template>
