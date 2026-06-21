<script setup lang="ts">
// SPDX-License-Identifier: AGPL-3.0-or-later
import type { NavItem } from '~/composables/useNavigation'

defineProps<{
  variant: 'bottom' | 'rail'
  items: NavItem[]
  active: string
}>()
defineEmits<{ navigate: [to: string] }>()
</script>

<template>
  <VBottomNavigation
    v-if="variant === 'bottom'"
    :model-value="active"
    grow
    color="primary"
  >
    <VBtn
      v-for="item in items"
      :key="item.key"
      :value="item.to"
      variant="text"
      :data-test="`nav-${item.key}`"
      @click="$emit('navigate', item.to)"
    >
      <VIcon :icon="item.icon" />
      <span>{{ $t(item.labelKey) }}</span>
    </VBtn>
  </VBottomNavigation>

  <VNavigationDrawer
    v-else
    :model-value="true"
    rail
    permanent
  >
    <VList :selected="[active]" nav>
      <VListItem
        v-for="item in items"
        :key="item.key"
        :value="item.to"
        :prepend-icon="item.icon"
        :title="$t(item.labelKey)"
        :data-test="`nav-${item.key}`"
        @click="$emit('navigate', item.to)"
      />
    </VList>
  </VNavigationDrawer>
</template>
