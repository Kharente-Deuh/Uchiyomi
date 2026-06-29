<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { PreferenceDto } from '#shared/dto/extensions'

defineProps<{ item: PreferenceDto }>()

defineEmits<{ 'update:model-value': [any | undefined] }>()
</script>

<template>
  <ExtensionsSettingsItemCheckbox
    v-if="item.type === 'checkbox'"
    :model-value="item.booleanValue"
    @update:model-value="$emit('update:model-value', $event)"
  />
  <ExtensionsSettingsItemSwitch
    v-else-if="item.type === 'switch'"
    :model-value="item.booleanValue"
    @update:model-value="$emit('update:model-value', $event)"
  />
  <ExtensionsSettingsItemText
    v-else-if="item.type === 'editText'"
    :model-value="item.textValue"
    @update:model-value="$emit('update:model-value', $event)"
  />
  <ExtensionsSettingsItemList
    v-else-if="item.type === 'list'"
    :model-value="item.textValue"
    :entries="item.entries"
    :entry-values="item.entryValues"
    @update:model-value="$emit('update:model-value', $event)"
  />
  <ExtensionsSettingsItemMultiSelect
    v-else-if="item.type === 'multiSelect'"
    :model-value="item.multiValue"
    :entries="item.entries"
    :entry-values="item.entryValues"
    @update:model-value="$emit('update:model-value', $event)"
  />
</template>
