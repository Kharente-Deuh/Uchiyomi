// SPDX-License-Identifier: AGPL-3.0-or-later

import { describe, expect, it } from 'vitest'
import { DEFAULT_PAGE } from '~/constants'
import { routeStub } from './route.stub'
import { requiresSetupCheck, resolveSetupGuard } from './setup-guard'

describe('requiresSetupCheck', () => {
  it('exempts /status, which must remain reachable on an uninitialized instance', () => {
    expect(requiresSetupCheck(routeStub({ name: 'status' }))).toBe(false)
  })

  it('requires the check on every other route, including /setup', () => {
    expect(requiresSetupCheck(routeStub({ name: 'index' }))).toBe(true)
    expect(requiresSetupCheck(routeStub({ name: 'library' }))).toBe(true)
    expect(requiresSetupCheck(routeStub({ name: 'setup' }))).toBe(true)
  })
})

describe('resolveSetupGuard', () => {
  const library = routeStub({ name: 'library', fullPath: '/library' })
  const setup = routeStub({ name: 'setup', fullPath: '/setup' })

  it('redirects uninitialized instance to /setup', () => {
    expect(resolveSetupGuard(library, 'required')).toBe('/setup')
  })

  it('allows an initialized instance through', () => {
    expect(resolveSetupGuard(library, 'done')).toBeUndefined()
  })

  it('allows the setup form to display when required', () => {
    expect(resolveSetupGuard(setup, 'required')).toBeUndefined()
  })

  it('redirects completed setup to home', () => {
    expect(resolveSetupGuard(setup, 'done')).toBe(DEFAULT_PAGE)
  })

  it('redirects to /status when status is indeterminate', () => {
    expect(resolveSetupGuard(library, 'unknown')).toBe('/status?redirect=%2Flibrary')
  })

  it('blocks the setup form while status is indeterminate', () => {
    expect(resolveSetupGuard(setup, 'unknown')).toBe('/status?redirect=%2Fsetup')
  })

  it('encodes the target path query string', () => {
    const paged = routeStub({ name: 'library', fullPath: '/library?page=2' })

    expect(resolveSetupGuard(paged, 'unknown')).toBe('/status?redirect=%2Flibrary%3Fpage%3D2')
  })
})
