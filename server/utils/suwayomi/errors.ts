// SPDX-License-Identifier: AGPL-3.0-or-later

import { ClientError } from 'graphql-request'

export type SuwayomiErrorKind = 'transport' | 'graphql' | 'timeout'

/** Unified error for every Suwayomi client failure. */
export class SuwayomiError extends Error {
  readonly kind: SuwayomiErrorKind
  override readonly cause?: unknown

  constructor(kind: SuwayomiErrorKind, message: string, cause?: unknown) {
    super(message)
    this.name = 'SuwayomiError'
    this.kind = kind
    this.cause = cause
  }
}

/**
 * Map any thrown value to a typed SuwayomiError. A ClientError (the query reached
 *  the server and returned `errors`) is a GraphQL error; everything else is transport.
 */
export function classifySuwayomiError(err: unknown): SuwayomiError {
  if (err instanceof SuwayomiError) {
    return err
  }

  if (err instanceof ClientError) {
    return new SuwayomiError('graphql', err.message, err)
  }

  const message = err instanceof Error ? err.message : String(err)

  return new SuwayomiError('transport', message, err)
}
