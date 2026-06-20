// SPDX-License-Identifier: AGPL-3.0-or-later
import type { TypedDocumentNode } from '@graphql-typed-document-node/core'
import { ClientError, GraphQLClient } from 'graphql-request'
import { classifySuwayomiError } from './errors'
import { withResilience } from './resilience'

const DEFAULT_TIMEOUT_MS = 15_000
const DEFAULT_RETRIES = 2
const DEFAULT_BACKOFF_MS = 250

export interface SuwayomiClient {
  execute: <T, V extends Record<string, unknown> = Record<string, never>>(
    document: TypedDocumentNode<T, V>,
    variables?: V,
  ) => Promise<T>
}

export interface SuwayomiClientOptions {
  endpoint: string
  timeoutMs?: number
  retries?: number
  backoffMs?: number
}

export function createSuwayomiClient(opts: SuwayomiClientOptions): SuwayomiClient {
  const gql = new GraphQLClient(opts.endpoint)
  const resilience = {
    timeoutMs: opts.timeoutMs ?? DEFAULT_TIMEOUT_MS,
    retries: opts.retries ?? DEFAULT_RETRIES,
    backoffMs: opts.backoffMs ?? DEFAULT_BACKOFF_MS,
    // GraphQL errors mean the request reached the server — never retry them.
    isRetryable: (err: unknown) => !(err instanceof ClientError),
  }

  return {
    async execute<T, V extends Record<string, unknown> = Record<string, never>>(
      document: TypedDocumentNode<T, V>,
      variables?: V,
    ): Promise<T> {
      try {
        return await withResilience(
          // graphql-request's request() overloads use a conditional rest-args type
          // that cannot be satisfied generically. We cast only the function signature
          // so argument and return types remain checked; the public execute() signature
          // on SuwayomiClient is correctly typed, and gql.request accepts the same
          // (document, variables?) positional form at runtime.
          () => (gql.request as (doc: TypedDocumentNode<T, V>, vars?: V) => Promise<T>)(document, variables),
          resilience,
        )
      } catch (err) {
        throw classifySuwayomiError(err)
      }
    },
  }
}
