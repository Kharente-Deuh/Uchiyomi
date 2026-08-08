// SPDX-License-Identifier: AGPL-3.0-or-later

import { describe, expect, it } from 'vitest'
import { routeStub } from './route.stub'
import { requiresSetupCheck, resolveSetupGuard } from './setup-guard'

describe('requiresSetupCheck', () => {
  it('exempte /status, qui doit rester atteignable sur une instance non initialisée', () => {
    expect(requiresSetupCheck(routeStub({ name: 'status' }))).toBe(false)
  })

  it('exige le check sur toute autre route, /setup compris', () => {
    expect(requiresSetupCheck(routeStub({ name: 'index' }))).toBe(true)
    expect(requiresSetupCheck(routeStub({ name: 'library' }))).toBe(true)
    expect(requiresSetupCheck(routeStub({ name: 'setup' }))).toBe(true)
  })
})

describe('resolveSetupGuard', () => {
  const library = routeStub({ name: 'library', fullPath: '/library' })
  const setup = routeStub({ name: 'setup', fullPath: '/setup' })

  it('envoie sur /setup une instance non initialisée', () => {
    expect(resolveSetupGuard(library, 'required')).toBe('/setup')
  })

  it('laisse passer une instance initialisée', () => {
    expect(resolveSetupGuard(library, 'done')).toBeUndefined()
  })

  it('laisse le formulaire de setup s\'afficher quand il est requis', () => {
    expect(resolveSetupGuard(setup, 'required')).toBeUndefined()
  })

  it('renvoie à l\'accueil un setup déjà fait', () => {
    expect(resolveSetupGuard(setup, 'done')).toBe('/')
  })

  it('envoie sur /status quand le statut est indéterminable', () => {
    expect(resolveSetupGuard(library, 'unknown')).toBe('/status?redirect=%2Flibrary')
  })

  it('refuse le formulaire de setup tant que le statut est indéterminable', () => {
    expect(resolveSetupGuard(setup, 'unknown')).toBe('/status?redirect=%2Fsetup')
  })

  it('encode la query string du chemin cible', () => {
    const paged = routeStub({ name: 'library', fullPath: '/library?page=2' })

    expect(resolveSetupGuard(paged, 'unknown')).toBe('/status?redirect=%2Flibrary%3Fpage%3D2')
  })
})
