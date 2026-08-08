// SPDX-License-Identifier: AGPL-3.0-or-later

import { describe, expect, it } from 'vitest'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { safeRedirect } from './redirect'
import { STATUS_PATH } from './routes'

const isEverythingKnown = (): boolean => true
const stub = defineComponent({ render: () => h('div') })

describe('safeRedirect', () => {
  it('accepte un chemin interne résolu par le router', () => {
    expect(safeRedirect('/library', isEverythingKnown)).toBe('/library')
  })

  it('conserve la query string du chemin cible', () => {
    expect(safeRedirect('/library?page=2', isEverythingKnown)).toBe('/library?page=2')
  })

  it('retombe sur / quand le paramètre est absent', () => {
    expect(safeRedirect(undefined, isEverythingKnown)).toBe('/')
  })

  it('retombe sur / quand la clé est répétée et vue-router rend un tableau', () => {
    expect(safeRedirect(['/a', '/b'], isEverythingKnown)).toBe('/')
  })

  it('rejette une URL absolue', () => {
    expect(safeRedirect('https://evil.com', isEverythingKnown)).toBe('/')
  })

  it('rejette une URL protocol-relative', () => {
    expect(safeRedirect('//evil.com', isEverythingKnown)).toBe('/')
  })

  it('rejette un antislash, que certains navigateurs normalisent en /', () => {
    expect(safeRedirect(String.raw`/\evil.com`, isEverythingKnown)).toBe('/')
  })

  it('rejette un chemin relatif sans / initial', () => {
    expect(safeRedirect('library', isEverythingKnown)).toBe('/')
  })

  it('rejette un chemin qui ne résout aucune route', () => {
    expect(safeRedirect('/nope', () => false)).toBe('/')
  })

  it('rejette /status, qui se renverrait sur lui-même', () => {
    expect(safeRedirect(STATUS_PATH, isEverythingKnown)).toBe('/')
  })

  it("rejette /status même porteur d'une query string", () => {
    expect(safeRedirect('/status?redirect=%2Flibrary', isEverythingKnown)).toBe('/')
  })

  it("rejette /status même porteur d'un fragment", () => {
    expect(safeRedirect('/status#bas', isEverythingKnown)).toBe('/')
  })

  it('accepte un chemin seulement préfixé par /status', () => {
    expect(safeRedirect('/statustique', isEverythingKnown)).toBe('/statustique')
  })

  it('rejette /status avec un slash final', () => {
    expect(safeRedirect('/status/', isEverythingKnown)).toBe('/')
  })

  it('rejette /status quelle que soit la casse', () => {
    expect(safeRedirect('/Status', isEverythingKnown)).toBe('/')
  })

  it("rejette /status porteur à la fois d'une query et d'un fragment", () => {
    expect(safeRedirect('/status?redirect=%2Fa#bas', isEverythingKnown)).toBe('/')
    expect(safeRedirect('/status#bas?redirect=%2Fa', isEverythingKnown)).toBe('/')
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

const isKnownRoute = (path: string): boolean => router.resolve(path).matched.length > 0

describe('safeRedirect adossé au router', () => {
  it('accepte une route à paramètre', () => {
    expect(safeRedirect('/source/one-piece', isKnownRoute)).toBe('/source/one-piece')
  })

  it('accepte une route à paramètres imbriqués', () => {
    expect(safeRedirect('/source/one-piece/chapter/42', isKnownRoute))
      .toBe('/source/one-piece/chapter/42')
  })

  it("accepte une route à paramètre porteuse d'une query string", () => {
    expect(safeRedirect('/source/one-piece?page=2', isKnownRoute)).toBe('/source/one-piece?page=2')
  })

  it('rejette un segment de paramètre manquant', () => {
    expect(safeRedirect('/source', isKnownRoute)).toBe('/')
  })

  it('rejette une URL absolue sans même interroger le router', () => {
    expect(safeRedirect('https://evil.com/x', isKnownRoute)).toBe('/')
  })
})
