<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
defineProps<{
  text: string
  loading?: boolean
}>()

defineEmits<{ confirm: [] }>()
const show = defineModel<boolean>()

const { mobile } = useDisplay()
</script>

<template>
  <VDialog
    v-if="!mobile"
    v-model="show"
    width="20rem"
    transition="dialog-bottom-transition"
    :persistent="loading"
  >
    <OrganismModalConfirmationCard
      :text
      :loading
      @confirm="$emit('confirm')"
      @cancel="show = false"
    >
      <slot />
    </OrganismModalConfirmationCard>
  </VDialog>

  <VBottomSheet
    v-else
    v-model="show"
    :persistent="loading"
  >
    <OrganismModalConfirmationCard
      :text
      :loading
      @confirm="$emit('confirm')"
      @cancel="show = false"
    >
      <slot />
    </OrganismModalConfirmationCard>
  </VBottomSheet>
</template>
