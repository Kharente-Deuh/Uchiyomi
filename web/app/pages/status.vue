<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { safeRedirect } from '~/utils/redirect'
import { isStatusPath } from '~/utils/routes'
import { sleep } from '~/utils/sleep'

definePageMeta({ layout: 'auth' })

const POLL_INTERVAL_MS = 5 * 1000

const { getServerStatus } = createHealthApi()
const route = useRoute('status')
const router = useRouter()

const serverStatus = ref<ServerStatusResponse>()
let stopped = false

const report = computed(() => {
  const current = serverStatus.value

  return current && current.status !== 'unreachable' ? current : undefined
})

const isDown = computed(() => serverStatus.value?.status === 'failed' || serverStatus.value?.status === 'unreachable')

const failedComponents = computed(() => {
  return Object.entries(report.value?.components ?? {})
    .filter(([, component]) => component.status !== 'ok')
    .map(([name, component]) => ({ name, reason: component.reason ?? '' }))
})

const loaderValue = computed(() => {
  const components = Object.values(report.value?.components ?? {})

  if (components.length === 0) {
    return undefined
  }

  const started = components.filter(({ status }) => status === 'ok')

  return (started.length / components.length) * 100
})

function isKnownRoute(path: string): boolean {
  return router.resolve(path).matched.length > 0
}

async function statusCheckLoop(): Promise<void> {
  while (true) {
    if (stopped) {
      return
    }

    const status = await getServerStatus()

    if (stopped) {
      return
    }

    serverStatus.value = status

    if (status.status === 'ok') {
      await navigateTo(safeRedirect(route.query.redirect, isKnownRoute))
      if (!isStatusPath(route.fullPath)) {
        return
      }
    }

    await sleep(POLL_INTERVAL_MS)
  }
}

onScopeDispose(() => {
  stopped = true
})

onMounted(() => {
  statusCheckLoop()
})
</script>

<template>
  <VProgressLinear
    v-if="!isDown"
    :indeterminate="loaderValue === undefined"
    :model-value="loaderValue"
    color="primary"
    height="4"
    rounded
  />

  <div v-else class="text-error w-100">
    <h2 class="mx-auto font-weight-bold text-center">
      {{ $t('status.failedTitle') }}
    </h2>
    <p class="text-center">
      {{ $t('status.failedHint') }}
    </p>

    <ul v-if="failedComponents.length > 0" class="mt-2 ps-4">
      <li v-for="component in failedComponents" :key="component.name">
        {{ component.reason ? $t('status.componentFailure', component) : component.name }}
      </li>
    </ul>
  </div>
</template>
