<script setup lang="ts">
defineProps<{
  label: string
  icon?: string
  disabled?: boolean
}>()

const modelValue = defineModel<boolean | undefined>({ required: true })
const { mobile } = useDisplay()
const color = computed(() => {
  if (modelValue.value === undefined) {
    return 'secondary'
  }

  return modelValue.value ? 'primary' : 'error'
})

function toggle(): void {
  if (modelValue.value === undefined) {
    modelValue.value = true
  } else if (modelValue.value === true) {
    modelValue.value = false
  } else {
    modelValue.value = undefined
  }
}

const borderClass = computed(() => {
  if (modelValue.value === undefined) {
    return 'border-thin-secondary'
  }

  return modelValue.value ? 'border-thin-primary' : 'border-thin-error'
})
</script>

<template>
  <VChip
    :disabled
    :size="mobile ? 'small' : undefined"
    variant="tonal"
    :color
    class="border-thin"
    :class="borderClass"
    :prepend-icon="icon"
    :text="label"
    @click="toggle"
  />
</template>
