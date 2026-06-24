<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { PreferenceDto } from '#shared/dto/extensions'
import { createExtensionsApi } from '../api/extensions.api'

const props = defineProps<{ sourceId: string }>()
const api = createExtensionsApi()

const preferences = ref<PreferenceDto[]>([])
const loading = ref(false)

async function load(): Promise<void> {
  loading.value = true
  const res = await api.getPreferences(props.sourceId)
  loading.value = false
  if (res.success) {
    preferences.value = res.data
  }
}

async function write(position: number, payload: { booleanValue?: boolean, textValue?: string, multiValue?: string[] }): Promise<void> {
  const res = await api.updatePreference(props.sourceId, { position, ...payload })
  if (res.success) {
    preferences.value = res.data
  }
}

function entryItems(p: PreferenceDto): { title: string, value: string }[] {
  return (p.entries ?? []).map((title, i) => ({ title, value: (p.entryValues ?? [])[i] ?? '' }))
}

onMounted(load)
</script>

<template>
  <MoleculeLoadingState v-if="loading" />
  <div v-else>
    <template v-for="p in preferences" :key="p.position">
      <VSwitch
        v-if="(p.type === 'switch' || p.type === 'checkbox') && p.visible"
        :model-value="p.booleanValue ?? p.booleanDefault ?? false"
        :label="p.title"
        :messages="p.summary"
        color="primary"
        @update:model-value="v => write(p.position, { booleanValue: !!v })"
      />
      <VTextField
        v-else-if="p.type === 'editText' && p.visible"
        :model-value="p.textValue ?? p.textDefault ?? ''"
        :label="p.title"
        :hint="p.summary"
        persistent-hint
        @change="(e: Event) => write(p.position, { textValue: (e.target as HTMLInputElement).value })"
      />
      <VSelect
        v-else-if="p.type === 'list' && p.visible"
        :model-value="p.textValue ?? p.textDefault"
        :items="entryItems(p)"
        :label="p.title"
        :hint="p.summary"
        persistent-hint
        @update:model-value="v => write(p.position, { textValue: v as string })"
      />
      <VSelect
        v-else-if="p.type === 'multiSelect' && p.visible"
        :model-value="p.multiValue ?? p.multiDefault ?? []"
        :items="entryItems(p)"
        :label="p.title"
        :hint="p.summary"
        multiple
        persistent-hint
        @update:model-value="v => write(p.position, { multiValue: v as string[] })"
      />
    </template>
  </div>
</template>
