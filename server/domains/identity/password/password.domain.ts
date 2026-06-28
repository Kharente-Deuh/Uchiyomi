// SPDX-License-Identifier: AGPL-3.0-or-later

export interface PasswordHashParams {
  password: string
}

export interface PasswordVerifyParams {
  hash: string
  password: string
}

// The PORT — the scrypt adapter implements it.
export interface PasswordHasher {
  hash: (p: PasswordHashParams) => Promise<string>
  verify: (p: PasswordVerifyParams) => Promise<boolean>
}
