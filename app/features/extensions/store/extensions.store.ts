// SPDX-License-Identifier: AGPL-3.0-or-later
import type { ExtensionDto } from '#shared/dto/extensions'

export const useExtensionsStore = defineStore('extensions', () => {
  const extensions = ref<ExtensionDto[]>([])
  const page = ref(1)
  const nsfwFilter = ref<boolean>()
  const isInstalledFilter = ref<boolean>()
  const isUpToDateFilter = ref<boolean | undefined>()
  const hasFiltersBeenSet = ref(false)

  function update(extension: ExtensionDto): void {
    const index = extensions.value.findIndex(e => e.pkgName === extension.pkgName)
    if (index === -1) {
      return
    }

    extensions.value[index] = extension
  }

  function setExtensions(value: ExtensionDto[]): void {
    extensions.value = value
  }

  function setPage(value: number): void {
    page.value = value
  }

  function setNsfwFilter(value: boolean | undefined): void {
    nsfwFilter.value = value
    hasFiltersBeenSet.value = true
  }

  function setIsInstalledFilter(value: boolean | undefined): void {
    isInstalledFilter.value = value
    hasFiltersBeenSet.value = true
  }

  function setIsUpToDateFilter(value: boolean | undefined): void {
    isUpToDateFilter.value = value
    hasFiltersBeenSet.value = true
  }

  function clear(): void {
    extensions.value = []
  }

  return {
    extensions,
    update,
    setExtensions,
    clear,

    page,
    nsfwFilter,
    isInstalledFilter,
    isUpToDateFilter,
    setPage,
    setNsfwFilter,
    setIsInstalledFilter,
    setIsUpToDateFilter,
    hasFiltersBeenSet,
  }
})
