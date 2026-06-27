// SPDX-License-Identifier: AGPL-3.0-or-later
import type { H3Event } from 'h3'
import type { z } from 'zod'

export async function parseBody<T>(event: H3Event, schema: z.ZodSchema<T>): Promise<T> {
  const parsed = await readValidatedBody(event, schema.safeParse)
  if (!parsed.success) {
    throw createError({ statusCode: 400, statusMessage: 'Invalid body' })
  }

  return parsed.data
}

export async function parseQuery<T>(event: H3Event, schema: z.ZodSchema<T>): Promise<T> {
  const parsed = await getValidatedQuery(event, schema.safeParse)
  if (!parsed.success) {
    throw createError({ statusCode: 400, statusMessage: 'Invalid query' })
  }

  return parsed.data
}
