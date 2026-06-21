// SPDX-License-Identifier: AGPL-3.0-or-later
// Flat ESM module, consumed via `import * as Password`.
export interface HashParams {
  password: string
}

export interface VerifyParams {
  hash: string
  password: string
}

// The PORT — the scrypt adapter implements it.
export interface Hasher {
  hash: (p: HashParams) => Promise<string>
  verify: (p: VerifyParams) => Promise<boolean>
}
