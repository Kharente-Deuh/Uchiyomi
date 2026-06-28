<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
const props = withDefaults(defineProps<{
  items: { title: string, value: string }[]
  noDataText?: string
  color?: string
  prependIcon?: string
}>(), {
  color: 'secondary',
})

const modelValue = defineModel<string>({ required: true })
const internalColor = computed(() => {
  if (props.items.length === 0) {
    return 'error'
  }

  return props.color
})

const modelValueTitle = computed(() => props.items.find(item => item.value === modelValue.value)?.title ?? props.noDataText)

const isDisabled = computed(() => props.items.length <= 1)
</script>

<template>
  <VChip
    :color="internalColor"
    :class="`border-thin-${internalColor} ${isDisabled ? '' : 'cursor-pointer'}`"
    :prepend-icon="prependIcon"
    :disabled="items.length <= 1"
    :append-icon="items.length > 1 ? 'fa6-solid:caret-down' : undefined"
    rounded="pill"
  >
    {{ modelValueTitle }}
    <VMenu activator="parent" location="bottom end">
      <VList
        density="compact"
        class="py-0 text-label-small list-override"
        style="max-height: 15rem !important; border-radius: 10px !important;"
      >
        <VListItem
          v-for="item in items"
          :key="item.value"
          class="text-label-small"
          color="primary"
          density="compact"
          :title="item.title"
          :active="modelValue === item.value"
          @click="modelValue = item.value"
        />
      </VList>
    </VMenu>
  </VChip>
</template>
