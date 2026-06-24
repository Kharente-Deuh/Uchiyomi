<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import ExtensionHealthBadge from './ExtensionHealthBadge.vue'

const {
  extensions,
  fetchLoading,
  install,
  uninstall,
  installExtensionsLoading,
  nsfwFilter,
  isInstalledFilter,
  isUpToDateFilter,
  searchFilter,
} = useExtensions()
</script>

<template>
  <div>
    <VRow class="mb-4">
      <VCol
        cols="12"
        sm="6"
        md="3"
      >
        <VSelect
          v-model="nsfwFilter"
          :label="$t('extensions.filters.nsfw')"
          :items="[
            { title: $t('extensions.filterOptions.all'), value: undefined },
            { title: $t('extensions.filterOptions.yes'), value: true },
            { title: $t('extensions.filterOptions.no'), value: false },
          ]"
          item-title="title"
          item-value="value"
          clearable
        />
      </VCol>
      <VCol
        cols="12"
        sm="6"
        md="3"
      >
        <VSelect
          v-model="isInstalledFilter"
          :label="$t('extensions.filters.installed')"
          :items="[
            { title: $t('extensions.filterOptions.all'), value: undefined },
            { title: $t('extensions.filterOptions.yes'), value: true },
            { title: $t('extensions.filterOptions.no'), value: false },
          ]"
          item-title="title"
          item-value="value"
          clearable
        />
      </VCol>
      <VCol
        cols="12"
        sm="6"
        md="3"
      >
        <VSelect
          v-model="isUpToDateFilter"
          :label="$t('extensions.filters.hasUpdate')"
          :items="[
            { title: $t('extensions.filterOptions.all'), value: undefined },
            { title: $t('extensions.filterOptions.yes'), value: true },
            { title: $t('extensions.filterOptions.no'), value: false },
          ]"
          item-title="title"
          item-value="value"
          clearable
        />
      </VCol>
      <VCol
        cols="12"
        sm="6"
        md="3"
      >
        <VTextField
          v-model="searchFilter"
          :label="$t('extensions.filters.search')"
          clearable
        />
      </VCol>
    </VRow>
    <MoleculeLoadingState v-if="fetchLoading" />
    <VList v-else>
      <VListItem
        v-for="ext in extensions"
        :key="ext.pkgName"
        :prepend-avatar="ext.iconUrl ?? ''"
        :title="ext.name"
        :subtitle="`${ext.lang} · ${ext.versionName}`"
      >
        <template #append>
          <ExtensionHealthBadge
            :is-healthy="ext.isHealthy"
            class="mr-2"
          />
          <VBtn
            v-if="ext.isInstalled"
            :loading="installExtensionsLoading.includes(ext.pkgName)"
            color="error"
            variant="text"
            @click="uninstall(ext.pkgName)"
          >
            {{ $t('extensions.actions.uninstall') }}
          </VBtn>
          <VBtn
            v-else
            :loading="installExtensionsLoading.includes(ext.pkgName)"
            color="primary"
            variant="text"
            @click="install(ext.pkgName)"
          >
            {{ $t('extensions.actions.install') }}
          </VBtn>
        </template>
      </VListItem>
    </VList>
  </div>
</template>
