<script setup lang="ts">
// SPDX-License-Identifier: AGPL-3.0-or-later
// Stub login route so the auth guard has a redirect target. M3.3 replaces this
// with the real sign-in form and its own auth layout.
definePageMeta({ layout: false })
</script>

<template>
  <VApp>
    <VMain>
      <div class="d-flex align-center justify-center" style="min-height: 100dvh">
        <p class="text-medium-emphasis">
          {{ $t('login.placeholder') }}
        </p>
      </div>
    </VMain>
  </VApp>
</template>
