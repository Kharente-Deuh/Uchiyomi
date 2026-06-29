// SPDX-License-Identifier: AGPL-3.0-or-later

// hashPassword / verifyPassword are nuxt-auth-utils server utils (scrypt) declared
// as ambient globals in .nuxt/types/nitro-imports.d.ts (via nitro.d.ts). No explicit
// import is needed or possible — vue-tsc project-references mode resolves them as globals.
import type { PasswordHasher, PasswordHashParams, PasswordVerifyParams } from '../../../password.domain'

export class ScryptPasswordHasher implements PasswordHasher {
  hash(p: PasswordHashParams): Promise<string> {
    return hashPassword(p.password)
  }

  verify(p: PasswordVerifyParams): Promise<boolean> {
    return verifyPassword(p.hash, p.password)
  }
}
