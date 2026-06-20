// SPDX-License-Identifier: AGPL-3.0-or-later
import process from 'node:process'
import { defineVitestProject } from '@nuxt/test-utils/config'
import { defineConfig } from 'vitest/config'

// Load `.env` so DB integration tests can read TEST_DATABASE_URL (same approach
// as prisma.config.ts; Prisma 7 / Vitest do not auto-load it). Workers inherit
// process.env from this config process.
try {
  process.loadEnvFile()
} catch {
  // No .env (e.g. CI provides env vars directly) — ignore.
}

// Two projects: a plain Node project for server/DB/util tests (no Nuxt Vite
// transforms, so native CJS deps like `pg` load correctly), and a Nuxt project
// for component/app tests.
export default defineConfig(async () => ({
  test: {
    projects: [
      {
        test: {
          name: 'node',
          environment: 'node',
          include: ['test/**/*.test.ts'],
        },
      },
      await defineVitestProject({
        test: {
          name: 'nuxt',
          environment: 'nuxt',
          include: ['app/**/*.{test,spec}.ts'],
        },
      }),
    ],
  },
}))
