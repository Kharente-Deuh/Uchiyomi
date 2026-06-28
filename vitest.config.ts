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
    // Native GitHub Actions annotations for test failures when running in CI.
    reporters: process.env.GITHUB_ACTIONS ? ['default', 'github-actions'] : ['default'],
    coverage: {
      provider: 'v8',
      // json-summary + json feed the PR coverage-comment action and the Gist
      // badge step; text shows in logs.
      reporter: ['text', 'json', 'json-summary'],
      // Still emit a report when tests fail so the PR comment stays populated.
      reportOnFailure: true,
      exclude: [
        'prisma/generated/**',
        '.nuxt/**',
        '**/*.config.{ts,mjs,js}',
        'test/**',
        '**/*.d.ts',
      ],
    },
    projects: [
      {
        test: {
          name: 'node',
          environment: 'node',
          include: ['test/**/*.test.ts'],
          // Integration/e2e tests share the single TEST_DATABASE_URL database and
          // each wipe `appUser`; running their files in parallel makes them clobber
          // one another. Serialise this project's files (the DB tests dominate
          // wall-clock anyway, so the cost is negligible).
          fileParallelism: false,
        },
      },
      await defineVitestProject({
        test: {
          name: 'nuxt',
          environment: 'nuxt',
          include: ['app/**/*.{test,spec}.ts'],
          // entry.mjs must come first: it calls setupNuxt() in a beforeAll,
          // which initialises the Nuxt app. setup.ts then installs Vuetify in
          // its own beforeAll, which runs after setupNuxt() completes.
          setupFiles: [
            'node_modules/@nuxt/test-utils/dist/runtime/entry.mjs',
            'app/test/setup.ts',
          ],
        },
      }),
    ],
  },
}))
