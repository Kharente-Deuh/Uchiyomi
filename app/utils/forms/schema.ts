// SPDX-License-Identifier: AGPL-3.0-or-later

import type { SchemaDescription, SchemaObjectDescription } from 'yup'
import type { AnyObjectSchema } from './types'
import { reach, ValidationError } from 'yup'

/**
 * Walk describe().fields along a dotted path to retrieve the leaf SchemaDescription.
 * This is more reliable than `reach()` for nested schemas because yup's `reach()` may
 * not expose `nullable`/`optional` correctly on nested sub-schemas.
 */
function describeAtPath(schema: AnyObjectSchema, path: string): SchemaDescription {
  const parts = path.split('.')
  let desc: SchemaDescription = schema.describe()

  for (const part of parts) {
    const objDesc = desc as SchemaObjectDescription
    if (!objDesc.fields || !(part in objDesc.fields)) {
      // Fallback: use reach() and describe() directly
      const sub = reach(schema, path) as { describe: () => SchemaDescription }

      return sub.describe()
    }

    desc = objDesc.fields[part] as SchemaDescription
  }

  return desc
}

export function getFieldMeta(schema: AnyObjectSchema, path: string): { label?: string, required: boolean } {
  const desc = describeAtPath(schema, path)
  const { label, nullable, optional } = desc as SchemaDescription & { nullable: boolean, optional: boolean }

  return { label: label || undefined, required: !nullable && !optional }
}

export async function validateValues(schema: AnyObjectSchema, values: unknown): Promise<Record<string, string[]>> {
  try {
    await schema.validate(values, { abortEarly: false })

    return {}
  } catch (error) {
    if (!(error instanceof ValidationError)) {
      throw error
    }

    const result: Record<string, string[]> = {}
    for (const issue of error.inner) {
      const key = issue.path
      if (key && result[key] === undefined) {
        result[key] = issue.errors
      }
    }

    return result
  }
}
