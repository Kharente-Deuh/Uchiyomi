<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
defineProps<{ data: TestResponse | undefined }>()
const show = defineModel<boolean>({ required: true })
</script>

<template>
  <OrganismModal
    v-model="show"
    prepend-icon="fa6-solid:vial"
    :title="$t('settings.oidc.test.title')"
    no-actions
    @close="show = false"
  >
    <div v-if="data" class="d-flex flex-column ga-4">
      <template
        v-for="[k, v] of Object.entries(data)"
        :key="k"
      >
        <VTextField
          v-if="typeof v === 'string'"
          :model-value="v"
          readonly
          :label="$t(`settings.oidc.test.fields.${k}.label`)"
          hide-details
        />
        <VCheckbox
          v-else-if="typeof v === 'boolean'"
          :model-value="v"
          readonly
          hide-details
          :label="$t(`settings.oidc.test.fields.${k}.label`)"
          density="comfortable"
        />
      </template>
    </div>
  </OrganismModal>
</template>
