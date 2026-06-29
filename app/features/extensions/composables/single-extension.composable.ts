// SPDX-License-Identifier: AGPL-3.0-or-later

import type { ExtensionDto, SourceDto } from '#shared/dto/extensions'
import type { ExtensionSettingsDto } from '#shared/dto/extensions/extension-settings.dto'
import { createExtensionsApi } from '../api/extensions.api'
import { useSingleExtensionStore } from '../store/single-extension.store'

interface SingleExtensionComposable {
  extension: ComputedRef<ExtensionDto | undefined>

  sources: ComputedRef<SourceDto[]>
  fetchSourcesLoading: Ref<boolean>
  toggleSourceEnabled: (sourceId: string) => Promise<void>
  sourceToggleLoading: Ref<Set<string>>

  settings: ComputedRef<ExtensionSettingsDto | undefined>
  hasSettings: ComputedRef<boolean>
  updateSettings: (settings: ExtensionSettingsDto) => Promise<void>
  uninstallExtension: () => Promise<boolean>
}

/**
 * Reads the single-extension store. The store is primed by the `single-extension`
 * route middleware before the page renders, so this composable does not fetch; it
 * only exposes the cached data and clears it when the page unmounts.
 */
export function useSingleExtension(pkgName: string): SingleExtensionComposable {
  const store = useSingleExtensionStore()
  const authStore = useAuthStore()
  const canManageExtensions = computed(() => authStore.capabilities.canManageExtensions)
  const api = createExtensionsApi()
  const toast = useToast()
  const { t } = useI18n()

  const extension = computed(() => store.extension)
  const sources = computed(() => store.sources)
  const settings = computed(() => store.settings)

  async function fetchExtension(): Promise<void> {
    const res = await api.getExtension(pkgName)
    if (!res.success) {
      toast.error(t('extension.errors.loadFailed'))

      return
    }

    store.setExtension(res.data.extension)
  }

  const fetchSourcesLoading = ref(false)
  async function fetchSources(): Promise<void> {
    fetchSourcesLoading.value = true
    const res = await api.listSources(pkgName)
    if (res.success) {
      store.setSources(res.data)
    } else {
      toast.error(t('sources.errors.loadFailed'))
      console.error(res.error)
    }

    fetchSourcesLoading.value = false
  }

  async function fetchSettings(): Promise<void> {
    if (!canManageExtensions.value) {
      return
    }

    const res = await api.getSettings(pkgName)
    if (res.success) {
      store.setSettings(res.data)
    } else {
      toast.error(t('sources.settings.errors.loadFailed'))
      console.error(res.error)
    }
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
    const res = await api.setSourceEnabled(pkgName, sourceId, !source.isEnabled)
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
    await Promise.all([
      ...(store.extension ? [] : [fetchExtension()]),
      ...(store.sources.length === 0 ? [fetchSources()] : []),
      ...(canManageExtensions.value ? [fetchSettings()] : []),
    ])
  })

  onBeforeUnmount(() => {
    store.clear()
  })

  const hasSettings = computed(() => {
    if (!settings.value) {
      return false
    }

    if (settings.value.common.length > 0) {
      return true
    }

    return settings.value.sources.some(s => s.preferences.length > 0)
  })

  async function updateSettings(settings: ExtensionSettingsDto): Promise<void> {
    const res = await api.updateSettings(pkgName, settings)
    if (res.success) {
      store.updateSettings(res.data)
    }
  }

  return {
    extension,
    sources,
    settings,
    fetchSourcesLoading,
    uninstallExtension,
    hasSettings,
    toggleSourceEnabled,
    sourceToggleLoading,
    updateSettings: useDebounceFn(updateSettings, 500),
  }
}
