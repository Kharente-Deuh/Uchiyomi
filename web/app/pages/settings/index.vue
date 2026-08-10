<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { RouteLocationRaw } from 'vue-router'
import { ADMIN_ROUTE_GROUP, AUTHENTICATED_ROUTE_GROUP } from '~/constants/auth'

definePageMeta({
  layout: 'default',
  authGroups: [AUTHENTICATED_ROUTE_GROUP, ADMIN_ROUTE_GROUP],
})

const { isAdmin } = useAuth()
const { t } = useI18n()
const { smAndDown } = useDisplay()

interface SettingsModule {
  title: string
  to: RouteLocationRaw
  description?: string
  icon: string
}

const settingsModules = computed((): SettingsModule[] => [
  ...(isAdmin.value
    ? [
      {
        title: t('settings.oidc.titleShort'),
        description: t('settings.oidc.description'),
        icon: 'fa6-regular:address-card',
        to: '/settings/oidc',
      } satisfies SettingsModule,
      ]
    : []),
])
</script>

<template>
  <OrganismPageLayout :title="$t('settings.title')">
    <div class="settings-modules-grid" :class="{ 'px-6': smAndDown }">
      <AtomLink
        v-for="(m, i) of settingsModules"
        :key="i"
        :to="m.to"
      >
        <VCard class="pa-4 border-thin" style="border-radius: 12px;">
          <div class="d-flex ga-6 text-wrap align-center">
            <VIcon :icon="m.icon" size="large" />
            <div class="d-flex flex-column">
              <span class="font-title text-title-large text-truncate"> {{ m.title }}</span>
              <span v-if="m.description" class="text-body-medium text-medium-emphasis text-wrap">{{ m.description
              }}</span>
            </div>
          </div>
        </VCard>
      </AtomLink>
    </div>
  </OrganismPageLayout>
</template>

<style lang="scss" scoped>
.settings-modules-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1.5rem;
}

@media screen and (max-width: 900px) {
  .settings-modules-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media screen and (max-width: 900px) {
  .settings-modules-grid {
    grid-template-columns: 1fr;
  }
}
</style>
