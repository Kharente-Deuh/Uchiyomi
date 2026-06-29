<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
const props = defineProps<{ modelValue?: string }>()
const emits = defineEmits<{ 'update:modelValue': [string | undefined] }>()

const internalModelValue = ref(props.modelValue)
const debounced = useDebounce(internalModelValue, 400)

watch(debounced, value => emits('update:modelValue', value))
</script>

<template>
  <VTextField
    v-model.trim="internalModelValue"
    hide-details
    style="max-width: 15rem !important;"
    density="compact"
  />
</template>
