<script setup lang="ts">
const props = defineProps<{
  enabled?: boolean
  lang: string
}>()

const code = ref(props.lang === 'all' ? '*' : props.lang.split('-')[0] ?? '')
</script>

<template>
  <div
    class="d-flex flex-column justify-center align-center source-lang-code rounded-lg"
    :class="{ 'source-lang-code--active': enabled }"
  >
    <div
      :style="{ paddingTop: lang === 'all' ? '10px' : undefined }"
      class="d-flex justify-center align-center h-100 text-uppercase font-weight-bold"
      :class="{
        'text-label-medium': lang !== 'all',
        'text-title-large': lang === 'all',
      }"
    >
      {{ code }}
    </div>
  </div>
</template>

<style lang="scss">
.source-lang-code {
  transition: all 0.2s ease-in-out;
  height: 28px;
  width: 28px;
  min-width: 28px;

  color: rgb(var(--v-theme-secondary));
  background-color: rgb(var(--v-theme-secondary), 0.15);

  &--active {
    color: rgb(var(--v-theme-primary)) !important;
    background-color: rgb(var(--v-theme-primary), 0.15) !important;
  }
}
</style>
