<script setup lang="ts">
const props = defineProps<{
  modelValue?: string
  entries: string[]
  entryValues: string[]
}>()

const emits = defineEmits<{ 'update:modelValue': [string | undefined] }>()
const internalModelValue = ref<string | null>(props.modelValue || null)
const debounced = useDebounce(internalModelValue, 400)
watch(debounced, value => emits('update:modelValue', value || undefined))
const items = computed(() => props.entries?.map((entry, i) => ({
  value: props.entryValues?.[i],
  title: entry,
})) ?? [])
</script>

<template>
  <VSelect
    v-model="internalModelValue"
    hide-details
    :items
    style="max-width: 15rem !important;"
    density="compact"
  />
</template>
