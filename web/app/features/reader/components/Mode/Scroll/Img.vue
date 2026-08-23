<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
const props = defineProps<{
  src: string
  root?: HTMLElement
}>()

const emit = defineEmits<{
  intersecting: [isIntersecting: boolean]
}>()

const el = useTemplateRef<HTMLElement>('el')

useIntersectionObserver(el, ([entry]) => {
  if (!entry) {
    return
  }

  emit('intersecting', entry.isIntersecting)
}, {
  root: () => props.root,
  threshold: 0,
})

onBeforeUnmount(() => {
  emit('intersecting', false)
})
</script>

<template>
  <div ref="el">
    <VImg
      eager
      :src="src"
      class="w-100"
    />
  </div>
</template>
