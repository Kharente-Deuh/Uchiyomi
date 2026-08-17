// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { RouteLocationRaw } from 'vue-router'
import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { DEFAULT_PAGE } from '~/constants'
import { safeRedirect } from './redirect'
import { STATUS_PATH } from './routes'

const { resolve } = vi.hoisted(() => ({
  resolve: { fn: (_: RouteLocationRaw) => ({ matched: [{}] }) },
}))

function unsubscribe(): void {}

function noop(): () => void {
  return unsubscribe
}

async function ready(): Promise<void> {}

function resolveRoute(to: RouteLocationRaw): { matched: object[] } {
  return resolve.fn(to)
}

const routerStub = {
  resolve: resolveRoute,
  beforeEach: noop,
  beforeResolve: noop,
  afterEach: noop,
  onError: noop,
  isReady: ready,
}

function useRouterStub(): typeof routerStub {
  return routerStub
}

mockNuxtImport('useRouter', () => useRouterStub)

const stub = defineComponent({ render: () => h('div') })

describe('safeRedirect', () => {
  beforeEach(() => {
    resolve.fn = () => ({ matched: [{}] })
  })

  it('accepts an internal path resolved by the router', () => {
    expect(safeRedirect('/library')).toBe('/library')
  })

  it('preserves the target path query string', () => {
    expect(safeRedirect('/library?page=2')).toBe('/library?page=2')
  })

  it('falls back to the default page when the parameter is absent', () => {
    expect(safeRedirect(undefined)).toBe(DEFAULT_PAGE)
  })

  it('falls back to the default page when the key is repeated and vue-router returns an array', () => {
    expect(safeRedirect(['/a', '/b'])).toBe(DEFAULT_PAGE)
  })

  it('rejects an absolute URL', () => {
    expect(safeRedirect('https://evil.com')).toBe(DEFAULT_PAGE)
  })

  it('rejects a protocol-relative URL', () => {
    expect(safeRedirect('//evil.com')).toBe(DEFAULT_PAGE)
  })

  it('rejects a backslash, which some browsers normalize to /', () => {
    expect(safeRedirect(String.raw`/\evil.com`)).toBe(DEFAULT_PAGE)
  })

  it('rejects a relative path without a leading /', () => {
    expect(safeRedirect('library')).toBe(DEFAULT_PAGE)
  })

  it('rejects a path that matches no route', () => {
    resolve.fn = () => ({ matched: [] })

    expect(safeRedirect('/nope')).toBe(DEFAULT_PAGE)
  })

  it('rejects /status, which would redirect to itself', () => {
    expect(safeRedirect(STATUS_PATH)).toBe(DEFAULT_PAGE)
  })

  it('rejects /status even when it carries a query string', () => {
    expect(safeRedirect('/status?redirect=%2Flibrary')).toBe(DEFAULT_PAGE)
  })

  it('rejects /status even when it carries a fragment', () => {
    expect(safeRedirect('/status#bas')).toBe(DEFAULT_PAGE)
  })

  it('accepts a path only prefixed by /status', () => {
    expect(safeRedirect('/statustique')).toBe('/statustique')
  })

  it('rejects /status with a trailing slash', () => {
    expect(safeRedirect('/status/')).toBe(DEFAULT_PAGE)
  })

  it('rejects /status regardless of case', () => {
    expect(safeRedirect('/Status')).toBe(DEFAULT_PAGE)
  })

  it('rejects /status carrying both a query and a fragment', () => {
    expect(safeRedirect('/status?redirect=%2Fa#bas')).toBe(DEFAULT_PAGE)
    expect(safeRedirect('/status#bas?redirect=%2Fa')).toBe(DEFAULT_PAGE)
  })
})

const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: '/', component: stub },
    { path: '/status', component: stub },
    { path: '/source/:slug', component: stub },
    { path: '/source/:slug/chapter/:id', component: stub },
  ],
})

describe('safeRedirect with router', () => {
  beforeEach(() => {
    resolve.fn = to => router.resolve(to)
  })

  it('accepts a parameterized route', () => {
    expect(safeRedirect('/source/one-piece')).toBe('/source/one-piece')
  })

  it('accepts a route with nested parameters', () => {
    expect(safeRedirect('/source/one-piece/chapter/42'))
      .toBe('/source/one-piece/chapter/42')
  })

  it('accepts a parameterized route carrying a query string', () => {
    expect(safeRedirect('/source/one-piece?page=2')).toBe('/source/one-piece?page=2')
  })

  it('rejects a missing parameter segment', () => {
    expect(safeRedirect('/source')).toBe(DEFAULT_PAGE)
  })

  it('rejects an absolute URL without even consulting the router', () => {
    expect(safeRedirect('https://evil.com/x')).toBe(DEFAULT_PAGE)
  })
})
