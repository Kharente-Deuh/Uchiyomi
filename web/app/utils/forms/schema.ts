// SPDX-License-Identifier: AGPL-3.0-or-later

import type { Schema, SchemaDescription, SchemaObjectDescription } from 'yup'
import type { AnyObjectSchema } from './types'
import { reach, ValidationError } from 'yup'

function describeAtPath(schema: AnyObjectSchema, path: string): SchemaDescription {
  const parts = path.split('.')
  let desc: SchemaDescription = schema.describe()

  for (const part of parts) {
    const objectDesc = desc as SchemaObjectDescription
    if (!objectDesc.fields || !Object.hasOwn(objectDesc.fields, part)) {
      const sub = reach(schema, path) as { describe: () => SchemaDescription }

      return sub.describe()
    }

    desc = objectDesc.fields[part] as SchemaDescription
  }

  return desc
}

const emptyValues: Record<string, unknown> = { string: '', array: [] }

// yup describes `defined()` and `required()` identically on everything but strings,
// so the only way to tell them apart is to check whether the empty value validates.
function requiresNonEmptyValue(schema: AnyObjectSchema, path: string, type: string): boolean {
  try {
    return !(reach(schema, path) as Schema).isValidSync(emptyValues[type])
  } catch {
    return true
  }
}

export function getFieldMeta(schema: AnyObjectSchema, path: string): { label?: string, required: boolean } {
  const desc = describeAtPath(schema, path)
  const { label, nullable, optional, type } = desc as SchemaDescription & { nullable: boolean, optional: boolean }

  return {
    label: label || undefined,
    required: !nullable && !optional && requiresNonEmptyValue(schema, path, type),
  }
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
