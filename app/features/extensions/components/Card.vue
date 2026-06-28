<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { ExtensionDto } from '#shared/dto/extensions'

const props = defineProps<{
  extension: ExtensionDto
  loading?: boolean
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
    icon: 'fa6-regular:trash-can',
    color: 'error',
    tooltip: t('actions.uninstall'),
    action: () => emits('uninstall'),
  }
})

const canSeeExtension = computed(() => props.extension.isInstalled && !props.extension.hasUpdate)
</script>

<template>
  <AtomLink :to="canSeeExtension ? `/browse/extensions/${extension.pkgName}` : undefined" class="text-truncate">
    <div
      class="d-flex justify-space-between ga-2 pa-2 bg-surface text-truncate"
      :class="{
        'elevation-down': !mobile,
        'border-thin': !mobile,
        'border-b-thin': mobile,
        'cursor-pointer': canSeeExtension,
      }"
      :style="{ borderRadius: !mobile ? '12px' : undefined }"
    >
      <div class="d-flex align-center ga-3 text-truncate">
        <ExtensionsAvatar :url="extension.iconUrl ?? ''" size="small" />
        <div class="d-flex flex-column ga-1 text-truncate">
          <span class="text-body-large font-weight-bold text-truncate">{{ extension.name }}</span>
          <div class="d-flex align-center ga-3">
            <AtomChipLang
              :lang="extension.lang"
              size="small"
            />
            <AtomChipVersion
              :version="extension.versionName"
              :has-update="extension.hasUpdate"
              size="small"
            />
            <AtomChipNsfw v-if="extension.isNsfw" size="small" />
          </div>
        </div>
      </div>

      <div class="d-flex flex-column justify-center pr-1">
        <VBtn
          v-if="actionBtn"
          v-tooltip="actionBtn.tooltip"
          :class="`border-thin-${actionBtn.color}`"
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
