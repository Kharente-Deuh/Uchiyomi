// SPDX-License-Identifier: AGPL-3.0-or-later

import type { ExtensionDto, SourceDto } from '#shared/dto/extensions'
import type { ExtensionSettingsDto } from '#shared/dto/extensions/extension-settings.dto'

export const useSingleExtensionStore = defineStore('singleExtension', () => {
  const extension = ref<ExtensionDto>()
  const sources = ref<SourceDto[]>([])
  const settings = ref<ExtensionSettingsDto>()

  function setExtension(value: ExtensionDto): void {
    extension.value = value
  }

  function setSources(value: SourceDto[]): void {
    sources.value = value
  }

  function updateSource(value: SourceDto): void {
    const index = sources.value.findIndex(s => s.id === value.id)
    if (index === -1) {
      return
    }

    sources.value[index] = value
  }

  function setSettings(value: ExtensionSettingsDto): void {
    settings.value = value
  }

  function updateSettings(value: ExtensionSettingsDto): void {
    settings.value = value
  }

  function clear(): void {
    extension.value = undefined
    sources.value = []
    settings.value = undefined
  }

  return {
    extension,
    setExtension,
    setSources,
    updateSource,
    sources,
    settings,
    setSettings,
    updateSettings,
    clear,
  }
})
