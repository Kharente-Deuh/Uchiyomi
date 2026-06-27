<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
const props = defineProps<{ modelValue?: boolean }>()
const emits = defineEmits<{ 'update:modelValue': [boolean | undefined] }>()

const internalModelValue = ref(props.modelValue)
const debounced = useDebounce(internalModelValue, 400)

watch(debounced, value => emits('update:modelValue', value))
</script>

<template>
  <VSwitch
    v-model="internalModelValue"
    hide-details
    size="small"
    color="primary"
    class="d-inline-block"
    inset
  />
</template>
