<script setup lang="ts">
// SPDX-License-Identifier: AGPL-3.0-or-later
import type { ModalCardProps } from './Card.vue'

type Props = ModalCardProps & {
  width?: string
  persistent?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  isFormComplete: true,
  width: '30rem',
})

defineEmits<{ submit: [], cancel: [] }>()
const show = defineModel<boolean>({ required: true })
const { mobile } = useDisplay()

const modalCardProps = computed((): ModalCardProps => {
  const { width: _width, persistent: _persistent, ...rest } = props

  return rest
})
</script>

<template>
  <VDialog
    v-if="!mobile"
    v-model="show"
    :width
    max-height="70%"
    transition="dialog-bottom-transition"
    :persistent="loading || globalLoading || persistent"
  >
    <OrganismModalCard
      v-bind="modalCardProps"
      @close="show = false"
      @cancel="$emit('cancel')"
      @submit="$emit('submit')"
    >
      <slot />
      <template #actions>
        <slot name="actions" />
      </template>
    </OrganismModalCard>
  </VDialog>

  <VBottomSheet
    v-else
    v-model="show"
    :persistent="loading || globalLoading || persistent"
  >
    <OrganismModalCard
      class="border-thin"
      v-bind="modalCardProps"
      @close="show = false"
      @submit="$emit('submit')"
      @cancel="$emit('cancel')"
    >
      <slot />
      <template #actions>
        <slot name="actions" />
      </template>
    </OrganismModalCard>
  </VBottomSheet>
</template>
