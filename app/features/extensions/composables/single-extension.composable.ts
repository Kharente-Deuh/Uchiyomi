import type { ExtensionDto, ExtensionHealthDto, SourceDto } from '#shared/dto/extensions'
import { createExtensionsApi } from '../api/extensions.api'
import { useSingleExtensionStore } from '../store/single-extension.store'

interface SingleExtensionComposable {
  extension: ComputedRef<ExtensionDto | undefined>
  health: ComputedRef<ExtensionHealthDto | undefined>
  sources: ComputedRef<SourceDto[]>
  fetchSourcesLoading: Ref<boolean>

  uninstallExtension: () => Promise<boolean>

  toggleSourceEnabled: (sourceId: string) => Promise<void>
  sourceToggleLoading: Ref<Set<string>>
}

/**
 * Reads the single-extension store. The store is primed by the `single-extension`
 * route middleware before the page renders, so this composable does not fetch; it
 * only exposes the cached data and clears it when the page unmounts.
 */
export function useSingleExtension(pkgName: string): SingleExtensionComposable {
  const store = useSingleExtensionStore()
  const api = createExtensionsApi()
  const toast = useToast()
  const { t } = useI18n()

  const extension = computed(() => store.extension)
  const sources = computed(() => store.sources)

  async function fetchExtension(): Promise<void> {
    const res = await api.getExtension(pkgName)
    if (!res.success) {
      toast.error(t('extension.errors.loadFailed'))

      return
    }

    store.setExtension(res.data.extension)
    if (res.data.health) {
      store.setHealth(res.data.health)
    }
  }

  const fetchSourcesLoading = ref(false)
  async function fetchSources(): Promise<void> {
    fetchSourcesLoading.value = true
    const res = await api.listSources(pkgName)
    if (res.success) {
      store.setSources(res.data)
    } else {
      toast.error(t('extension.errors.loadFailed'))
      console.error(res.error)
      fetchSourcesLoading.value = true

      return
    }

    fetchSourcesLoading.value = false
  }

  async function uninstallExtension(): Promise<boolean> {
    const res = await api.extensionAction(pkgName, 'uninstall')
    if (!res.success) {
      toast.error(t('extension.errors.uninstallFailed'))

      return false
    }

    return true
  }

  const sourceToggleLoading = ref<Set<string>>(new Set())
  async function toggleSourceEnabled(sourceId: string): Promise<void> {
    const source = sources.value.find(s => s.id === sourceId)
    if (!source) {
      return
    }

    sourceToggleLoading.value.add(sourceId)
    const res = await api.setSourceEnabled(sourceId, !source.isEnabled)
    if (res.success) {
      store.updateSource(res.data)
    } else {
      console.error(res.error)
      toast.error(source.isEnabled
        ? t('sources.errors.disableFailed')
        : t('sources.errors.enableFailed'))
    }

    sourceToggleLoading.value.delete(sourceId)
  }

  onMounted(async () => {
    if (!store.extension) {
      fetchExtension()
    }

    await Promise.all([
      ...(store.extension ? [] : [fetchExtension()]),
      ...(store.sources.length === 0 ? [fetchSources()] : []),
    ])
  })

  onBeforeUnmount(() => {
    store.clear()
  })

  return {
    extension,
    health: computed(() => store.health),
    sources,
    fetchSourcesLoading,
    uninstallExtension,

    toggleSourceEnabled,
    sourceToggleLoading,
  }
}
