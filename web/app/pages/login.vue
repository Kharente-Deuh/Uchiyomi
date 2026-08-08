<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { NOT_AUTHENTICATED_ROUTE_GROUP } from '~/constants/auth'

definePageMeta({
  layout: 'auth',
  authGroups: [NOT_AUTHENTICATED_ROUTE_GROUP],
})
</script>

<template>
  <h1>LOGIN</h1>
</template>
