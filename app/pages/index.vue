<script setup lang="ts">
const { t } = useI18n()

useHead({
  title: () => t('welcome.title'),
})
</script>

<template>
  <VRow justify="center" class="py-12">
    <VCol
      cols="12"
      md="8"
      lg="6"
      class="text-center"
    >
      <VIcon
        size="64"
        color="primary"
        icon="fa6-solid:book-open"
        class="mb-4"
      />
      <h1 class="text-h4 mb-2">
        {{ t('welcome.title') }}
      </h1>
      <p class="text-medium-emphasis mb-1">
        {{ t('welcome.subtitle') }}
      </p>
      <p class="text-medium-emphasis">
        {{ t('app.tagline') }}
      </p>
    </VCol>
  </VRow>
</template>
