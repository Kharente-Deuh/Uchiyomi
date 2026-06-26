import type { ExtensionDto, ExtensionHealthDto, SourceDto } from '#shared/dto/extensions'

export const useSingleExtensionStore = defineStore('singleExtension', () => {
  const extension = ref<ExtensionDto>()
  const health = ref<ExtensionHealthDto>()
  const sources = ref<SourceDto[]>([])

  function setExtension(value: ExtensionDto): void {
    extension.value = value
  }

  function setHealth(value: ExtensionHealthDto): void {
    health.value = value
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

  function clear(): void {
    extension.value = undefined
    health.value = undefined
    sources.value = []
  }

  return {
    extension,
    health,
    setExtension,
    setHealth,
    setSources,
    updateSource,
    sources,
    clear,
  }
})
