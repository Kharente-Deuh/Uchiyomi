<script setup lang="ts">
import type { ComicSource } from '~/features/comics/types'
import asuraImg from '~/assets/images/sources/asurascans.webp'
import { ASURA_SOURCE_NAME } from '~/constants'

defineProps<{ disabled?: boolean }>()
const source = defineModel<ComicSource | undefined>({ required: true })
const { t } = useI18n()

const items: { value: ComicSource, title: string, img: string }[] = [
  {
    value: ASURA_SOURCE_NAME,
    title: t('sources.asurascans.title'),
    img: asuraImg,
  },
]
</script>

<template>
  <VSelect
    v-model="source"
    density="compact"
    class="border-thin"
    style="border-radius: 12px"
    width="13rem"
    :disabled
    hide-details
    clearable
    :items="items"
    :label="$t('library.source.label')"
    @update:model-value="source = $event || undefined"
  >
    <template v-if="source" #selection="{ item }">
      <div class="d-flex ga-4 align-center text-truncate">
        <VImg
          :src="item.img"
          aspect-ratio="1"
          width="24"
          class="rounded-pill"
        />
        <span class="text-body-medium text-truncate">{{ item.title }}</span>
      </div>
    </template>

    <template #item="{ item, props: itemProps }">
      <VListItem density="compact" v-bind="itemProps">
        <template #prepend>
          <VImg
            :src="item.img"
            aspect-ratio="1"
            width="24"
            class="rounded-pill mr-2"
          />
        </template>
      </VListItem>
    </template>
  </VSelect>
</template>
