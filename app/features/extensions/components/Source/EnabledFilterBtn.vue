<script setup lang="ts">
const modelValue = defineModel<boolean | undefined>({ required: true })

function toggle(): void {
  switch (modelValue.value) {
    case true:
      modelValue.value = false
      break
    case false:
      modelValue.value = undefined
      break
    default:
      modelValue.value = true
  }
}

const color = computed(() => {
  switch (modelValue.value) {
    case true:
      return 'primary'
    case false:
      return 'error'
    default:
      return 'secondary'
  }
})
</script>

<template>
  <VBtn
    prepend-icon="fa6-solid:filter"
    variant="tonal"
    rounded="pill"
    size="small"
    :color
    text="Activées"
    :class="`border-thin-${color}`"
    @click.stop="toggle"
  />
</template>
