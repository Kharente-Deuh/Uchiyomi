<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
defineProps<{ loading: boolean }>()

const { locale } = useI18n()
const { provider } = useOidcProvider()
</script>

<template>
  <MoleculeCardComponent
    :disabled="loading"
    :loading
    :title="$t('settings.oidc.category.users.title', { count: provider?.users?.length })"
    icon="fa6-solid:users"
  >
    <div v-if="provider?.users?.length" class="d-flex flex-column">
      <div class="provider-users-grid pb-3 border-b-thin font-title text-title-medium">
        <span>{{ $t('settings.oidc.fields.users.username') }}</span>
        <div class="d-flex justify-center">
          <span>{{ $t('settings.oidc.fields.users.role') }}</span>
        </div>

        <div class="d-flex justify-end">
          <span>{{ $t('settings.oidc.fields.users.linkedAt') }}</span>
        </div>
      </div>

      <template v-for="(user, i) of provider.users" :key="i">
        <div class="provider-users-grid py-3" :class="{ 'border-b-thin': i < provider.users.length - 1 }">
          <span class="my-auto">{{ user.username }}</span>
          <div class="d-flex justify-center my-auto">
            <VChip
              :color="user.isAdmin ? 'primary' : undefined"
              :text="user.isAdmin ? $t('role.admin') : $t('role.user')"
              density="comfortable"
              size="small"
            />
          </div>

          <div class="d-flex justify-end">
            <span class="text-body-medium text-medium-emphasis my-auto">{{ user.linkedAt.toLocaleDateString(locale) }}</span>
          </div>
        </div>
      </template>
    </div>

    <div v-else-if="!loading" class="d-flex justify-center">
      <span class="text-medium-emphasis">{{ $t('settings.oidc.category.users.noUsers') }}</span>
    </div>
  </MoleculeCardComponent>
</template>

<style lang="scss">
.provider-users-grid {
  overflow-x: auto;
  display: grid;
  gap: 1.5rem;
  grid-template-columns: repeat(3, 1fr);
}
</style>
