<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { ExtensionDto } from '~~/shared/dto/extensions'

defineProps<ExtensionDto>()
</script>

<template>
  <AtomLink :to="isInstalled ? `/extensions/${pkgName}` : undefined">
    <div
      class="d-flex flex-column ga-2 border-thin pa-2 bg-surface"
      style="border-radius: 12px"
    >
      <div class="d-flex align-center ga-2">
        <ExtensionsAvatar :url="iconUrl ?? ''" />
        <div class="d-flex flex-column">
          <span class="text-body-medium font-weight-bold text-truncate">{{ name }}</span>
          <span class="text-label-small text-medium-emphasis text-truncate">{{ lang }}</span>
        </div>
      </div>
    </div>
  </AtomLink>
</template>
