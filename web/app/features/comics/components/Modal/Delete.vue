<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { Comic } from '../../types'

const props = defineProps<{ comic: Comic }>()
const emits = defineEmits<{ deleted: [] }>()
const show = defineModel<boolean>({ required: true })

const toast = useToast()
const { t } = useI18n()
const api = createComicsApi()

const loading = ref(false)

watch(show, (value) => {
  if (!value) {
    loading.value = false
  }
})

async function handleDelete(): Promise<void> {
  loading.value = true

  const res = await api.deleteById(props.comic.id)
  if (!res.success) {
    console.error('api.deleteById failed', res.error)
    toast.error(res.error.status === 404
      ? t('sources.asura.comic.notFound')
      : t('error.unknown'))

    return
  } else {
    emits('deleted')
  }

  show.value = false
}
</script>

<template>
  <OrganismModalConfirmation
    v-model="show"
    :loading
    :text="$t('sources.asura.delete.confirmation', { name: comic?.title ?? '' })"
    @confirm="handleDelete"
  />
</template>
