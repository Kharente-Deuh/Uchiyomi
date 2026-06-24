<script setup lang="ts">
// SPDX-License-Identifier: AGPL-3.0-or-later
import type { RouteLocationRaw } from 'vue-router'

export interface IntermediateListItemProps {
  title: string
  icon?: string
  subtitle?: string
  last?: boolean
  to: RouteLocationRaw
}

defineProps<IntermediateListItemProps>()

const { mobile } = useDisplay()
</script>

<template>
  <AtomLink :to>
    <MoleculeIntermediateListItemDesktop
      v-if="!mobile"
      :title
      :icon
      :subtitle
    />
    <MoleculeIntermediateListItemMobile
      v-else
      :title
      :icon
      :subtitle
      :last
    />
  </AtomLink>
</template>
