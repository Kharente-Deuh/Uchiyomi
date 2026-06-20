<script setup lang="ts">
const { locale, locales, setLocale } = useI18n()
const { t } = useI18n()

const available = computed(() => locales.value.map(l => (typeof l === 'string' ? { code: l, name: l } : l)))
</script>

<template>
  <VMenu>
    <template #activator="{ props }">
      <VBtn
        icon="fa6-solid:language"
        :aria-label="t('actions.switchLanguage')"
        variant="text"
        v-bind="props"
      />
    </template>
    <VList>
      <VListItem
        v-for="l in available"
        :key="l.code"
        :active="l.code === locale"
        :title="l.name"
        @click="setLocale(l.code)"
      />
    </VList>
  </VMenu>
</template>
