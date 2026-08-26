<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { ComicType } from '../../types'

const props = withDefaults(defineProps<{
  type: ComicType
  updatable?: boolean
  size?: 'default' | 'small'
}>(), { size: 'default' })

const emit = defineEmits<{ 'update:type': [ComicType] }>()

const { t } = useI18n()

const menuItems: { label: string, value: ComicType }[] = [
  {
    label: t('comics.type.manga'),
    value: 'manga',
  },
  {
    label: t('comics.type.manhua'),
    value: 'manhua',
  },
  {
    label: t('comics.type.manhwa'),
    value: 'manhwa',
  },
  {
    label: t('comics.type.mangatoon'),
    value: 'mangatoon',
  },
]

const showMenu = ref(false)

function toggleItem(item: ComicType): void {
  if (item !== props.type) {
    emit('update:type', item)
  }

  showMenu.value = false
}
</script>

<template>
  <span
    class="ga-3 w-fit h-fit px-2 py-1 border-thin text-truncate text-uppercase bg-background d-flex ga-1 align-center"
    :class="{
      'text-body-large': size === 'default',
      'text-body-medium': size === 'small',
      'cursor-pointer': updatable,
    }"
    :style="{ borderRadius: '12px' }"
  >
    <span>{{ $t(`comics.type.${type}`) }}</span>
    <template v-if="updatable">
      <VIcon
        icon="fa6-solid:pencil"
        class="text-label-small"
        size="x-small"
      />
      <VMenu
        v-model="showMenu"
        activator="parent"
        location="bottom center"
        offset="8"
      >
        <div
          class="d-flex flex-column bg-surface border-thin transition-smooth pa-2"
          style="border-radius: 12px;"
        >
          <div
            v-for="item in menuItems"
            :key="item.value"
            class="px-2 py-1 type-chip-item"
            :class="{ 'type-chip-item-active': type === item.value }"
            @click="toggleItem(item.value)"
          >
            <span class="transition-smooth">{{ $t(`comics.type.${item.value}`) }}</span>
          </div>
        </div>
      </VMenu>
    </template>
  </span>
</template>

<style scoped lang="scss">
.type-chip-item {
  cursor: pointer;
  border-radius: 8px;

  &:hover {
    color: rgb(var(--v-theme-primary));
  }
}

.type-chip-item-active {
  background-color: rgba(var(--v-theme-primary), 0.1);
  color: rgb(var(--v-theme-primary));
}
</style>
