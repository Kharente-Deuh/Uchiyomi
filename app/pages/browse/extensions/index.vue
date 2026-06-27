<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { ExtensionDto } from '#shared/dto/extensions'

const { mobile } = useDisplay()
const { capabilities } = useAuthStore()
const {
  extensions,
  fetchLoading,
  nsfwFilter,
  isInstalledFilter,
  isUpToDateFilter,
  searchFilter,
  install,
  installExtensionsLoading,
  update,
  updateExtensionsLoading,
  uninstall,
  fetchExtensions,
  maxPage,
  page,
} = useExtensions()

const showUninstallConfirmationModal = ref(false)
const uninstallLoading = ref<boolean>(false)
const toUninstall = ref<ExtensionDto>()
watch(toUninstall, (value) => {
  if (value) {
    showUninstallConfirmationModal.value = true
  }
})

async function uninstallExtension(pkgName: string): Promise<void> {
  uninstallLoading.value = true
  await uninstall(pkgName)
  uninstallLoading.value = false
  showUninstallConfirmationModal.value = false
}

const loadMoreSentinel = useTemplateRef<HTMLElement>('loadMoreSentinel')
useIntersectionObserver(loadMoreSentinel, ([entry]) => {
  if (entry?.isIntersecting && mobile.value && !fetchLoading.value && page.value < maxPage.value) {
    page.value += 1
  }
})

onMounted(() => {
  fetchExtensions()
})
</script>

<template>
  <OrganismPageLayout
    :title="$t('extensions.title')"
    icon="fa6-solid:puzzle-piece"
    :subtitle="$t('browse.extensions.subtitle')"
    :loading="fetchLoading"
    back-route="/browse"
  >
    <OrganismModalConfirmation
      v-if="capabilities.canManageExtensions && !mobile"
      v-model="showUninstallConfirmationModal"
      :text="$t('extensions.uninstall.confirmation')"
      :loading="uninstallLoading"
      @confirm="uninstallExtension(toUninstall!.pkgName)"
    >
      <div class="d-flex ga-2 w-100 text-truncate align-center justify-center">
        <VAvatar v-if="toUninstall!.iconUrl" :image="toUninstall?.iconUrl" />
        <span class="text-truncate">{{ toUninstall!.name }}</span>
      </div>
    </OrganismModalConfirmation>

    <div
      class="extensions-header-grid w-100 py-3 bg-background"
      :class="{ 'px-4': mobile, 'border-b-thin': mobile, 'mb-3': !mobile }"
    >
      <AtomInputSearch
        v-model="searchFilter"
        :max-width="mobile ? undefined : '25rem'"
      />
      <div class="d-flex ga-4 flex-wrap" :class="{ 'justify-space-between': mobile }">
        <AtomFilter
          v-if="capabilities.showNsfw"
          v-model="nsfwFilter"
          :label="$t('extensions.filters.nsfw')"
          icon="fa6-solid:ban"
          :disabled="fetchLoading"
        />
        <AtomFilter
          v-model="isInstalledFilter"
          icon="fa6-solid:floppy-disk"
          :label="$t('extensions.filters.installed')"
          :disabled="fetchLoading"
        />
        <AtomFilter
          v-model="isUpToDateFilter"
          icon="fa6-solid:exclamation"
          :label="$t('extensions.filters.hasUpdate')"
          :disabled="fetchLoading"
        />
      </div>
    </div>

    <div class="w-100" :class="{ 'extensions-grid': !mobile }">
      <ExtensionsCard
        v-for="(item, i) in extensions"
        :key="i"
        :extension="item"
        :loading="installExtensionsLoading.has(item.pkgName) || updateExtensionsLoading.has(item.pkgName) || (uninstallLoading && toUninstall?.pkgName === item.pkgName)"
        @install="install(item.pkgName)"
        @update="update(item.pkgName)"
        @uninstall="toUninstall = item"
      />
    </div>

    <div
      v-if="mobile"
      ref="loadMoreSentinel"
      class="d-flex justify-center py-4"
    >
      <VProgressCircular
        v-if="fetchLoading && page > 1"
        color="secondary"
        indeterminate
      />
    </div>

    <MoleculePaginationFooter
      v-if="!mobile"
      v-model="page"
      :pages-total="maxPage"
      :disabled="fetchLoading || !maxPage"
    />
  </OrganismPageLayout>
</template>

<style scoped>
.extensions-header-grid {
  position: sticky;
  top: 0;
  z-index: 10;
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 1rem;
}

.extensions-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
}

@media screen and (max-width: 1110px) {
  .extensions-header-grid {
    grid-template-columns: 1fr;
  }

  .extensions-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media screen and (max-width: 790px) {
  .extensions-header-grid {
    grid-template-columns: 1fr;
  }

  .extensions-grid {
    grid-template-columns: 1fr;
  }
}
</style>
