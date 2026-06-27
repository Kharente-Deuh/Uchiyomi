<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
definePageMeta({ middleware: 'single-extension-middleware' })

const { pkgName } = useRoute().params
const {
  extension,
  uninstallExtension,
  sources,
  toggleSourceEnabled,
  sourceToggleLoading,
  hasSettings,
} = useSingleExtension(pkgName)
const authStore = useAuthStore()
const canManageExtensions = computed(() => authStore.capabilities.canManageExtensions)

const showUninstallConfirmationModal = ref(false)
const uninstallExtensionLoading = ref(false)

async function doUninstallExtension(): Promise<void> {
  uninstallExtensionLoading.value = true
  const ok = await uninstallExtension()
  if (!ok) {
    uninstallExtensionLoading.value = false

    return
  }

  navigateTo('/browse/extensions')
}

const { mobile } = useDisplay()

const selectedTab = ref<'sources' | 'series'>('sources')
</script>

<template>
  <OrganismPageLayout
    :title="extension?.name ?? ''"
    :subtitle="extension?.lang ?? ''"
    :prepend-image="extension?.iconUrl"
    back-route="/browse/extensions"
    show-back-route
  >
    <template v-if="extension && !mobile" #header>
      <div class="d-flex ga-4 align-center w-100 mb-8">
        <VCard class="border-thin elevation-down px-6 pb-6 pt-10 d-flex justify-space-between align-end w-100 ga-6" :style="{ borderRadius: '16px' }">
          <div class="d-flex align-center ga-6 text-truncate">
            <ExtensionsAvatar :url="extension.iconUrl ?? ''" size="large" />
            <div class="d-flex flex-column ga-2 justify-end text-truncate">
              <span class="font-weight-bold text-truncate font-title" style="font-size: 20pt;">{{ extension.name }}</span>
              <div class="d-flex ga-4 text-body-medium align-end">
                <AtomChipVersion
                  :version="extension.versionName"
                  :has-update="extension.hasUpdate"
                />

                <span class="text-medium-emphasis"> {{ $t('sources.count', { count: sources.length }) }}</span>

                <AtomChipNsfw v-if="extension.isNsfw" />
              </div>
            </div>
          </div>

          <template v-if="canManageExtensions">
            <OrganismModalConfirmation
              v-model="showUninstallConfirmationModal"
              :loading="uninstallExtensionLoading"
              :text="$t('extension.uninstall.confirmation')"
              @confirm="doUninstallExtension"
            />

            <VBtn
              color="error"
              text="Uninstall"
              class="border-thin-error"
              prepend-icon="fa6-regular:trash-can"
              variant="tonal"
              @click="showUninstallConfirmationModal = true"
            />
          </template>
        </VCard>
      </div>
    </template>

    <template v-if="extension && mobile" #subtitle>
      <div class="d-flex ga-4 text-body-medium align-end">
        <AtomChipVersion
          :version="extension.versionName"
          :has-update="extension.hasUpdate"
          size="small"
        />

        <span class="text-medium-emphasis text-body-small"> {{ $t('sources.count', { count: sources.length }) }}</span>

        <AtomChipNsfw v-if="extension.isNsfw" size="small" />
      </div>
    </template>

    <template v-if="canManageExtensions" #append-title>
      <OrganismModalConfirmation
        v-model="showUninstallConfirmationModal"
        :loading="uninstallExtensionLoading"
        :text="$t('extension.uninstall.confirmation')"
        @confirm="doUninstallExtension"
      />
      <VBtn
        class="border-thin-error"
        icon="fa6-solid:trash"
        color="error"
        variant="tonal"
        size="small"
        :text="$t('actions.install')"
        @click="showUninstallConfirmationModal = true"
      />
    </template>

    <ExtensionsMobileTabs
      v-if="mobile"
      v-model="selectedTab"
      :style="{ zIndex: 1000 }"
      class="mb-4"
    />
    <ExtensionsSourceList
      v-show="!mobile || selectedTab === 'sources'"
      :can-manage-extensions
      :sources
      :has-settings
      :pkg-name
      :source-toggle-loading
      @toggle="toggleSourceEnabled"
    />
  </OrganismPageLayout>
</template>
