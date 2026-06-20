// SPDX-License-Identifier: AGPL-3.0-or-later
import { writeFileSync } from 'node:fs'
import process from 'node:process'
import { buildClientSchema, getIntrospectionQuery, printSchema } from 'graphql'
import { request } from 'graphql-request'

const base = process.env.SUWAYOMI_URL ?? 'http://localhost:4567'
const endpoint = `${base}/api/graphql`
const outPath = 'server/utils/suwayomi/schema.graphql'

try {
  const data = await request<{ __schema: unknown }>(endpoint, getIntrospectionQuery())
  // buildClientSchema expects the introspection result wrapped as { __schema }.
  const schema = buildClientSchema(data as never)
  writeFileSync(outPath, `${printSchema(schema)}\n`)

  console.log(`Wrote ${outPath} from ${endpoint}`)
} catch (err) {
  console.error(`Failed to introspect ${endpoint}:`, err)
  process.exitCode = 1
}
