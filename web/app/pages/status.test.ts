// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { ServerStatusResponse } from '~/composables/health.api'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import StatusPage from './status.vue'

const { getServerStatus, navigateTo, query, fullPath } = vi.hoisted(() => ({
  getServerStatus: vi.fn(),
  navigateTo: vi.fn(),
  query: { redirect: undefined as string | undefined },
  fullPath: '/status',
}))

vi.mock('~/utils/sleep', () => ({
  sleep: () => new Promise(() => {}),
}))

function healthApiStub(): { getServerStatus: typeof getServerStatus } {
  return { getServerStatus }
}

function routeStub(): { query: typeof query, fullPath: string } {
  return { query, fullPath }
}

mockNuxtImport('createHealthApi', () => healthApiStub)
mockNuxtImport('useRoute', () => routeStub)
mockNuxtImport('navigateTo', () => navigateTo)

async function mount(): Promise<VueWrapper> {
  return mountSuspended({ render: () => h(VApp, () => [h(StatusPage)]) })
}

beforeEach(() => {
  getServerStatus.mockReset()
  navigateTo.mockReset()
  query.redirect = undefined
  getServerStatus.mockResolvedValue({ status: 'starting', components: {} } satisfies ServerStatusResponse)
})

describe('statusPage', () => {
  it('redirects when the server is ok', async () => {
    getServerStatus.mockResolvedValue({ status: 'ok', components: { db: { status: 'ok' } } })
    await mount()

    await vi.waitFor(() => expect(navigateTo).toHaveBeenCalledWith('/feed'))
  })

  it('shows failed components when the server is down', async () => {
    getServerStatus.mockResolvedValue({
      status: 'failed',
      components: { db: { status: 'failed', reason: 'connection refused' } },
    })
    const wrapper = await mount()

    await vi.waitFor(() => expect(wrapper.text()).toContain('The server is unavailable'))
    expect(wrapper.text()).toContain('db: connection refused')
    expect(navigateTo).not.toHaveBeenCalled()
  })

  it('treats unreachable as down', async () => {
    getServerStatus.mockResolvedValue({ status: 'unreachable' })
    const wrapper = await mount()

    await vi.waitFor(() => expect(wrapper.text()).toContain('The server is unavailable'))
  })
})
