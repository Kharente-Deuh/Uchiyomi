import type { ExtensionDto } from '~~/shared/dto/extensions'

export const useExtensionsStore = defineStore('extensions', () => {
  const extensions = ref<ExtensionDto[]>([])

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

  return {
    extensions,
    update,
    setExtensions,
  }
})
