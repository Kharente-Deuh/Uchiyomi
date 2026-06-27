<script setup lang="ts">
defineProps<{
  hasSettings: boolean
  pkgName: string
}>()
defineEmits<{ showSettings: [] }>()
</script>

<template>
  <AtomLink
    v-if="hasSettings"
    :to="`/browse/extensions/${pkgName}/settings`"
  >
    <VBtn
      icon="fa6-solid:gear"
      size="x-small"
      class="border-thin-secondary"
      color="secondary"
    />
  </AtomLink>
  <VChip
    v-else
    color="grey"
    variant="tonal"
    size="small"
    :text="$t('sources.settings.noSettings')"
  />
</template>
