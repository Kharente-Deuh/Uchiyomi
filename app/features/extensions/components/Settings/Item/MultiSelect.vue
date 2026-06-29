<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
const props = defineProps<{
  modelValue?: string[]
  entries: string[]
  entryValues: string[]
}>()
const emits = defineEmits<{ 'update:modelValue': [string[] | undefined] }>()

const internalModelValue = ref(props.modelValue ?? [])
const items = ref<{ value: string, title: string }[]>([])

const debounced = useDebounce(internalModelValue, 400)
watch(debounced, value => emits('update:modelValue', value))

function toggle(value: string): void {
  // Reassign in both branches: useDebounce watches the ref, not its contents, so an
  // in-place push would not trigger the debounced emit (only removals would persist).
  if (internalModelValue.value.includes(value)) {
    internalModelValue.value = internalModelValue.value.filter(v => v !== value)
  } else {
    internalModelValue.value = [...internalModelValue.value, value]
  }
}

onMounted(() => {
  for (let i = 0; i < props.entries.length; i++) {
    items.value.push({
      value: props.entryValues[i] as string,
      title: props.entries[i] as string,
    })
  }
})
</script>

<template>
  <div class="d-flex ga-2 flex-wrap align-center">
    <VChip
      v-for="(value, index) in entryValues"
      :key="index"
      size="small"
      :class="`border-thin-${internalModelValue.includes(value) ? 'primary' : 'secondary'}`"
      :color="internalModelValue.includes(value) ? 'primary' : 'secondary'"
      :text="entries?.[index]"
      @click="toggle(value)"
    />
  </div>
</template>
