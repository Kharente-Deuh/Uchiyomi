<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
const props = defineProps<{
  size?: 'small' | 'default'
  lang: string
}>()

const isAll = computed(() => props.lang === 'all')
</script>

<template>
  <AtomChip
    :text="isAll ? $t('extensions.multiLanguages') : lang.split('-')[0]!"
    :class="{ 'text-uppercase': !isAll }"
    color="secondary"
    :size
  />
</template>
