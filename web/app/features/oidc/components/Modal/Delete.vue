<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
const show = defineModel<boolean>({ required: true })

const { deleteProvider, deleteLoading, provider } = useOidcProvider()
</script>

<template>
  <OrganismModalConfirmation
    v-model="show"
    :text="$t('settings.oidc.delete.confirmation', { name: provider?.displayName ?? '' })"
    :loading="deleteLoading"
    :confirm-text="$t('actions.delete')"
    @confirm="deleteProvider"
  />
</template>
