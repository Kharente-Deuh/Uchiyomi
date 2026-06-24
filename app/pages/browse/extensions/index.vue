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
  uninstall,
  uninstallLoading,
} = useExtensions()

const showUninstallConfirmationModal = ref(false)
const toUninstall = ref<ExtensionDto>()
watch(toUninstall, (value) => {
  if (value) {
    showUninstallConfirmationModal.value = true
  }
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
      v-if="capabilities.canManageExtensions"
      v-model="showUninstallConfirmationModal"
      :text="$t('extensions.uninstall.confirmation')"
      :loading="uninstallLoading"
      @confirm="uninstall(toUninstall!.pkgName)"
    >
      <div class="d-flex ga-2 w-100 text-truncate">
        <VAvatar v-if="toUninstall!.iconUrl" :image="toUninstall?.iconUrl" />
        <span class="text-truncate">{{ toUninstall!.name }}</span>
      </div>
    </OrganismModalConfirmation>

    <div
      class="extensions-header-grid w-100"
      :class="{ 'px-4': mobile }"
    >
      <AtomInputSearch
        v-model="searchFilter"
        :max-width="mobile ? undefined : '25rem'"
        :disabled="fetchLoading"
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

    <VVirtualScroll :items="extensions">
      <template #default="{ item }">
        <ExtensionsCard
          v-bind="item"
          :loading="installExtensionsLoading.includes(item.pkgName)"
          @install="install(item.pkgName)"
          @uninstall="toUninstall = item"
        />
      </template>
    </VVirtualScroll>
    <!-- <div class="extension-grid w-100">
      <ExtensionsCard
        v-for="(item, i) in extensions"
        :key="i"
        v-bind="item"
        :loading="installExtensionsLoading.includes(item.pkgName)"
        @install="install(item.pkgName)"
        @uninstall="toUninstall = item"
      />
    </div> -->
  </OrganismPageLayout>
</template>

<style scoped>
.extensions-header-grid {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 1rem;
}

@media screen and (max-width: 1110px) {
  .extensions-header-grid {
    grid-template-columns: 1fr;
  }
}
</style>
