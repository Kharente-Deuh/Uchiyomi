// SPDX-License-Identifier: AGPL-3.0-or-later

import process from 'node:process'
import { defineConfig, env } from 'prisma/config'

// Prisma 7 no longer auto-loads .env. Load it for Migrate/CLI commands
// (`prisma migrate`, `prisma db ...`). `prisma generate` does not need a URL.
try {
  process.loadEnvFile()
} catch {
  // No .env file (e.g. CI or `prisma generate`); env vars come from the environment.
}

export default defineConfig({
  schema: 'prisma/schema.prisma',
  datasource: {
    url: env('DATABASE_URL'),
  },
})
