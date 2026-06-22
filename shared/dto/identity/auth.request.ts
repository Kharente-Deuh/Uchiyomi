// SPDX-License-Identifier: AGPL-3.0-or-later
// Request wire contracts for the auth routes. Route zod schemas are typed
// `satisfies z.ZodType<...>` against these, guaranteeing the schema matches.
export interface SetupRequestDto {
  accountName: string
  displayName: string
  password: string
}

export interface LoginRequestDto {
  accountName: string
  password: string
}

export interface SetupStatusDto {
  required: boolean
  minPasswordLength: number
}
