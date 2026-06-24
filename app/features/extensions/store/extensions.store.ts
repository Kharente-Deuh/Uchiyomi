import type { ExtensionDto } from '~~/shared/dto/extensions'

export const useExtensionsStore = defineStore('extensions', () => {
  const extensions = ref<ExtensionDto[]>([])

  function install(pkgName: string): void {
    const index = extensions.value.findIndex(extension => extension.pkgName === pkgName)
    if (index === -1) {
      return
    }

    extensions.value[index]!.isInstalled = true
  }

  function uninstall(pkgName: string): void {
    const index = extensions.value.findIndex(extension => extension.pkgName === pkgName)
    if (index === -1) {
      return
    }

    extensions.value[index]!.isInstalled = false
  }

  function setExtensions(value: ExtensionDto[]): void {
    extensions.value = value
  }

  return {
    extensions,
    uninstall,
    install,
    setExtensions,
  }
})
