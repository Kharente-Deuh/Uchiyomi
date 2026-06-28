<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts" generic="T extends 'on-off' | 'intermediate'">
type ModelValue<T extends 'on-off' | 'intermediate'> = T extends 'on-off' ? boolean : boolean | undefined
const props = defineProps<{
  text: string
  prependIcon?: string
  disabled?: boolean
  type: T
}>()

const modelValue = defineModel<ModelValue<T>>({ required: true })

const color = computed(() => {
  if (props.type === 'on-off') {
    return modelValue.value ? 'primary' : 'secondary'
  }

  switch (modelValue.value) {
    case true:
      return 'primary'
    case false:
      return 'error'
    default:
      return 'secondary'
  }
})

function toggle(): void {
  if (props.type === 'on-off') {
    modelValue.value = !modelValue.value

    return
  }

  switch (modelValue.value) {
    case undefined:
      modelValue.value = true
      break
    case true:
      modelValue.value = false
      break
    default:
      modelValue.value = undefined as ModelValue<T>
  }
}
</script>

<template>
  <VBtn
    :disabled
    :color
    size="small"
    rounded="pill"
    :text
    :class="`border-thin-${color}`"
    :prepend-icon
    @click.stop="toggle"
  />
</template>
