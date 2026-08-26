<script setup lang="ts">
import type { ComicSource } from '~/features/comics/types'
import { ASURA_SOURCE_NAME, KING_OF_SHOJO_SOURCE_NAME } from '~/constants'

defineProps<{ disabled?: boolean }>()
const source = defineModel<ComicSource | undefined>({ required: true })
const { getSourceDetails } = useSources()
const asuraDetails = getSourceDetails('asurascans')
const kingOfShojoDetails = getSourceDetails('kingofshojo')

const items = computed((): { value: ComicSource, title: string, img: string }[] => [
  {
    value: ASURA_SOURCE_NAME,
    title: asuraDetails.name,
    img: asuraDetails.image,
  },
  {
    value: KING_OF_SHOJO_SOURCE_NAME,
    title: kingOfShojoDetails.name,
    img: kingOfShojoDetails.image,
  },
])
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
