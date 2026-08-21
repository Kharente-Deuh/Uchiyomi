<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
const { messages } = useToast()
</script>

<template>
  <VApp>
    <VMain>
      <slot />
    </VMain>

    <VSnackbarQueue v-model="messages" />
  </VApp>
</template>
