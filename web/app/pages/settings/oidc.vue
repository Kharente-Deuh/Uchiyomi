<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { ADMIN_ROUTE_GROUP, AUTHENTICATED_ROUTE_GROUP } from '~/constants/auth'

definePageMeta({
  layout: 'default',
  groups: [AUTHENTICATED_ROUTE_GROUP, ADMIN_ROUTE_GROUP],
})
</script>

<template>
  <OrganismPageLayout
    :title="$t('settings.oidc.title')"
    icon="fa6-solid:gear"
  />
</template>
