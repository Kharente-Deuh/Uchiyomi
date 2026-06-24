// SPDX-License-Identifier: AGPL-3.0-or-later
import type { ComputedRef, Ref } from 'vue'
import type { ExtensionDto, ExtensionListResponseDto } from '#shared/dto/extensions'
import { createExtensionsApi } from '../api/extensions.api'

export interface ExtensionsFilters {
  nsfw?: boolean
  isHealthy?: boolean
  isInstalled?: boolean
  hasUpdate?: boolean
  search?: string
}

export interface ExtensionsComposable {
  extensions: ComputedRef<ExtensionDto[]>
  maxPage: ComputedRef<number>
  fetchLoading: Ref<boolean> | ComputedRef<boolean>

  install: (pkgName: string) => void
  installExtensionsLoading: Ref<string[]>

  uninstall: (pkgName: string) => void
  uninstallLoading: Ref<boolean> | ComputedRef<boolean>

  nsfwFilter: Ref<boolean | undefined> | ComputedRef<boolean | undefined>
  isInstalledFilter: Ref<boolean | undefined> | ComputedRef<boolean | undefined>
  isUpToDateFilter: Ref<boolean | undefined> | ComputedRef<boolean | undefined>
  searchFilter: Ref<string | undefined> | ComputedRef<string | undefined>
}

const PAGE_SIZE = 100

export function useExtensions(): ExtensionsComposable {
  const api = createExtensionsApi()
  const toast = useToast()
  const { t } = useI18n()
  const store = useExtensionsStore()

  const nsfwFilter = ref<boolean>()
  const isInstalledFilter = ref<boolean>()
  const isUpToDateFilter = ref<boolean>()
  const searchFilter = ref<string>()
  const page = ref<number>(1)

  const debouncedSearch = useDebounce(searchFilter, 300)

  const queryKey = computed(() => [
    'extensions',
    page.value,
    ...(debouncedSearch.value ? [debouncedSearch.value] : []),
    ...(isInstalledFilter.value ? [isInstalledFilter.value] : []),
    ...(isUpToDateFilter.value ? [isUpToDateFilter.value] : []),
    ...(nsfwFilter.value ? [nsfwFilter.value] : []),
  ])

  const { data, isLoading } = useQuery<ExtensionListResponseDto>({
    key: queryKey,
    query: async () => {
      const res = await api.listExtensions({
        page: page.value,
        pageSize: PAGE_SIZE,
        search: debouncedSearch.value,
        isInstalled: isInstalledFilter.value,
        ...(isUpToDateFilter.value !== undefined && { hasUpdate: !isUpToDateFilter.value }),
        nsfw: nsfwFilter.value,
      })

      if (!res.success) {
        throw new Error(res.error?.message)
      }

      return res.data
    },
  })

  const maxPage = computed(() => Math.ceil((data.value?.totalCount ?? 0) / PAGE_SIZE))
  watch(isLoading, (value) => {
    if (!value && data.value) {
      store.setExtensions(data.value.items)
    }
  })

  async function doAction(pkgName: string, action: 'install' | 'uninstall'): Promise<void> {
    const res = await api.extensionAction(pkgName, action)
    if (!res.success) {
      toast.error(t('extensions.errors.actionFailed'))

      return
    }

    if (action === 'install') {
      store.install(pkgName)
    } else {
      store.uninstall(pkgName)
    }
  }

  const installExtensionsLoading = ref<string[]>([])
  async function install(pkgName: string): Promise<void> {
    installExtensionsLoading.value.push(pkgName)
    await doAction(pkgName, 'install')
    installExtensionsLoading.value = installExtensionsLoading.value.filter(p => p !== pkgName)
  }

  const uninstallLoading = ref<boolean>(false)
  async function uninstall(pkgName: string): Promise<void> {
    uninstallLoading.value = true
    await doAction(pkgName, 'uninstall')
    uninstallLoading.value = false
  }

  return {
    maxPage,
    install,
    installExtensionsLoading,
    uninstall,
    uninstallLoading,
    fetchLoading: isLoading,
    nsfwFilter,
    isInstalledFilter,
    isUpToDateFilter,
    searchFilter,
    extensions: computed(() => store.extensions),
  }
}
