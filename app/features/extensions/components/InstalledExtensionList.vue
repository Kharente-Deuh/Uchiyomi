<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import ExtensionHealthBadge from './ExtensionHealthBadge.vue'

const { extensions, fetchLoading } = useExtensions()
</script>

<template>
  <div>
    <MoleculeLoadingState v-if="fetchLoading" />
    <MoleculeEmptyState v-else-if="extensions.length === 0" :title="$t('extensions.empty.title')" />
    <VList v-else>
      <VListItem
        v-for="ext in extensions"
        :key="ext.pkgName"
        :to="`/extensions/${ext.pkgName}`"
        :prepend-avatar="ext.iconUrl ?? ''"
        :title="ext.name"
        :subtitle="ext.lang"
      >
        <template #append>
          <ExtensionHealthBadge :is-healthy="ext.isHealthy" />
        </template>
      </VListItem>
    </VList>
  </div>
</template>
