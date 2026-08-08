// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { RouteLocationRaw } from 'vue-router'
import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
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

  it('accepte un chemin interne résolu par le router', () => {
    expect(safeRedirect('/library')).toBe('/library')
  })

  it('conserve la query string du chemin cible', () => {
    expect(safeRedirect('/library?page=2')).toBe('/library?page=2')
  })

  it('retombe sur / quand le paramètre est absent', () => {
    expect(safeRedirect(undefined)).toBe('/')
  })

  it('retombe sur / quand la clé est répétée et vue-router rend un tableau', () => {
    expect(safeRedirect(['/a', '/b'])).toBe('/')
  })

  it('rejette une URL absolue', () => {
    expect(safeRedirect('https://evil.com')).toBe('/')
  })

  it('rejette une URL protocol-relative', () => {
    expect(safeRedirect('//evil.com')).toBe('/')
  })

  it('rejette un antislash, que certains navigateurs normalisent en /', () => {
    expect(safeRedirect(String.raw`/\evil.com`)).toBe('/')
  })

  it('rejette un chemin relatif sans / initial', () => {
    expect(safeRedirect('library')).toBe('/')
  })

  it('rejette un chemin qui ne résout aucune route', () => {
    resolve.fn = () => ({ matched: [] })

    expect(safeRedirect('/nope')).toBe('/')
  })

  it('rejette /status, qui se renverrait sur lui-même', () => {
    expect(safeRedirect(STATUS_PATH)).toBe('/')
  })

  it("rejette /status même porteur d'une query string", () => {
    expect(safeRedirect('/status?redirect=%2Flibrary')).toBe('/')
  })

  it("rejette /status même porteur d'un fragment", () => {
    expect(safeRedirect('/status#bas')).toBe('/')
  })

  it('accepte un chemin seulement préfixé par /status', () => {
    expect(safeRedirect('/statustique')).toBe('/statustique')
  })

  it('rejette /status avec un slash final', () => {
    expect(safeRedirect('/status/')).toBe('/')
  })

  it('rejette /status quelle que soit la casse', () => {
    expect(safeRedirect('/Status')).toBe('/')
  })

  it("rejette /status porteur à la fois d'une query et d'un fragment", () => {
    expect(safeRedirect('/status?redirect=%2Fa#bas')).toBe('/')
    expect(safeRedirect('/status#bas?redirect=%2Fa')).toBe('/')
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

describe('safeRedirect adossé au router', () => {
  beforeEach(() => {
    resolve.fn = to => router.resolve(to)
  })

  it('accepte une route à paramètre', () => {
    expect(safeRedirect('/source/one-piece')).toBe('/source/one-piece')
  })

  it('accepte une route à paramètres imbriqués', () => {
    expect(safeRedirect('/source/one-piece/chapter/42'))
      .toBe('/source/one-piece/chapter/42')
  })

  it("accepte une route à paramètre porteuse d'une query string", () => {
    expect(safeRedirect('/source/one-piece?page=2')).toBe('/source/one-piece?page=2')
  })

  it('rejette un segment de paramètre manquant', () => {
    expect(safeRedirect('/source')).toBe('/')
  })

  it('rejette une URL absolue sans même interroger le router', () => {
    expect(safeRedirect('https://evil.com/x')).toBe('/')
  })
})
