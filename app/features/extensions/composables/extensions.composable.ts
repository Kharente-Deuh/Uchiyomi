// SPDX-License-Identifier: AGPL-3.0-or-later
import type { ComputedRef, Ref } from 'vue'
import type { ExtensionDto } from '#shared/dto/extensions'
import { createExtensionsApi } from '../api/extensions.api'

export interface ExtensionsFilters {
  nsfw?: boolean
  isInstalled?: boolean
  hasUpdate?: boolean
  search?: string
}

export interface ExtensionsComposable {
  fetchExtensions: () => Promise<void>
  extensions: ComputedRef<ExtensionDto[]>
  maxPage: Ref<number>
  page: Ref<number>
  fetchLoading: Ref<boolean> | ComputedRef<boolean>

  install: (pkgName: string) => void
  installExtensionsLoading: Ref<Set<string>>

  uninstall: (pkgName: string) => Promise<void>

  update: (pkgName: string) => void
  updateExtensionsLoading: Ref<Set<string>>

  nsfwFilter: Ref<boolean | undefined> | ComputedRef<boolean | undefined>
  isInstalledFilter: Ref<boolean | undefined> | ComputedRef<boolean | undefined>
  isUpToDateFilter: Ref<boolean | undefined> | ComputedRef<boolean | undefined>
  searchFilter: Ref<string | undefined> | ComputedRef<string | undefined>
}

const PAGE_SIZE = 50

export function useExtensions(): ExtensionsComposable {
  const api = createExtensionsApi()
  const toast = useToast()
  const { t } = useI18n()
  const store = useExtensionsStore()
  const { mobile } = useDisplay()

  const page = computed({
    get: () => store.page,
    set: (value: number) => store.setPage(value),
  })

  const isInstalledFilter = computed({
    get: () => store.isInstalledFilter,
    set: (value: boolean) => store.setIsInstalledFilter(value),
  })

  const nsfwFilter = computed({
    get: () => store.nsfwFilter,
    set: (value: boolean) => store.setNsfwFilter(value),
  })

  const isUpToDateFilter = computed({
    get: () => store.isUpToDateFilter,
    set: (value: boolean) => store.setIsUpToDateFilter(value),
  })

  const searchFilter = ref<string>()
  const maxPage = ref<number>(0)
  const isLoading = ref<boolean>(false)

  const debouncedSearch = useDebounce(searchFilter, 300)

  async function fetchExtensions(resetData?: boolean): Promise<void> {
    isLoading.value = true
    const res = await api.listExtensions({
      page: page.value,
      pageSize: PAGE_SIZE,
      ...(debouncedSearch.value !== undefined && { search: debouncedSearch.value }),
      isInstalled: isInstalledFilter.value,
      ...(isUpToDateFilter.value !== undefined && { hasUpdate: !isUpToDateFilter.value }),
      nsfw: nsfwFilter.value,
    })

    isLoading.value = false

    if (res.success) {
      maxPage.value = Math.ceil(res.data.total / PAGE_SIZE)
      store.setExtensions(mobile.value && !resetData ? [...store.extensions, ...res.data.items] : res.data.items)
    } else {
      toast.error(t('extensions.errors.loadFailed'))
    }
  }

  watch([debouncedSearch, isInstalledFilter, isUpToDateFilter, nsfwFilter], () => {
    page.value = 1
    fetchExtensions(true)
  })

  watch(page, () => {
    fetchExtensions()
  })

  async function doAction(pkgName: string, action: 'install' | 'uninstall' | 'update'): Promise<void> {
    const res = await api.extensionAction(pkgName, action)
    if (!res.success) {
      toast.error(t('extensions.errors.actionFailed'))

      return
    }

    store.update(res.data)
  }

  onMounted(() => {
    if (!store.hasFiltersBeenSet) {
      isInstalledFilter.value = true
    }
  })

  onBeforeUnmount(() => {
    store.clear()
  })

  const installExtensionsLoading = ref<Set<string>>(new Set())
  async function install(pkgName: string): Promise<void> {
    installExtensionsLoading.value.add(pkgName)
    await doAction(pkgName, 'install')
    installExtensionsLoading.value.delete(pkgName)
  }

  async function uninstall(pkgName: string): Promise<void> {
    await doAction(pkgName, 'uninstall')
  }

  const updateExtensionsLoading = ref<Set<string>>(new Set())
  async function update(pkgName: string): Promise<void> {
    updateExtensionsLoading.value.add(pkgName)
    await doAction(pkgName, 'update')
    updateExtensionsLoading.value.delete(pkgName)
  }

  return {
    fetchExtensions,
    maxPage,
    page,
    install,
    installExtensionsLoading,
    uninstall,
    update,
    updateExtensionsLoading,
    fetchLoading: isLoading,
    nsfwFilter,
    isInstalledFilter,
    isUpToDateFilter,
    searchFilter,
    extensions: computed(() => store.extensions),
  }
}
