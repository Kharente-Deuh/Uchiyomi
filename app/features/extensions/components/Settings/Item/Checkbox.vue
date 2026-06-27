<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
const props = defineProps<{ modelValue?: boolean }>()
const emits = defineEmits<{ 'update:modelValue': [boolean | undefined] }>()

const internalModelValue = ref(props.modelValue ?? false)
const debounced = useDebounce(internalModelValue, 400)

watch(debounced, value => emits('update:modelValue', value))
</script>

<template>
  <VCheckbox
    v-model="internalModelValue"
    hide-details
  />
</template>
