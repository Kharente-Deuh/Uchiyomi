<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
const props = defineProps<{
  mode: 'btn' | 'label'
}>()

const icon = computed(() => props.mode === 'btn' ? 'fa6-solid:trash' : 'fa6-solid:check')
const { smAndDown } = useDisplay()
</script>

<template>
  <div
    class="d-flex transition-smooth flex-column items-center justify-center rounded-lg transition-smooth remove-library-btn"
    :class="{
      'border-thin': mode === 'label',
      'border-thin-error': mode === 'btn',
      'remove-library-btn--btn': mode === 'btn',
      'remove-library-btn--label': mode === 'label',
    }"
  >
    <div class="d-flex align-center transition-smooth ga-1">
      <VIcon
        :icon
        size="x-small"
      />

      <VExpandXTransition v-if="!smAndDown">
        <span
          v-show="mode === 'label'"
          class="text-no-wrap text-body-small transition-smooth"
        >{{ $t('sources.asurascans.label.inLibrary') }}</span>
      </VExpandXTransition>
    </div>

    <VTooltip
      v-if="mode === 'btn'"
      :text="$t(`comics.remove.title`)"
      activator="parent"
    />
  </div>
</template>

<style lang="scss" scoped>
.remove-library-btn {
  padding: 0.4rem;
  backdrop-filter: blur(2px);
}

.remove-library-btn--label {
  background-color: rgba(var(--v-theme-primary), 1);
  color: rgb(var(--v-theme-surface));
}

.remove-library-btn--btn {
  background-color: rgba(var(--v-theme-surface), 0.7);
  color: rgb(var(--v-theme-error));

  &:hover {
    background-color: rgb(var(--v-theme-error));
    color: rgb(var(--v-theme-surface));
  }
}
</style>
