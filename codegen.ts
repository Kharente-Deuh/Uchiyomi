// SPDX-License-Identifier: AGPL-3.0-or-later
import type { CodegenConfig } from '@graphql-codegen/cli'

// Generate from the COMMITTED SDL snapshot (never a live server), so codegen runs
// in CI and postinstall without Suwayomi. Refresh the snapshot with `pnpm suwayomi:schema`.
const config: CodegenConfig = {
  schema: 'server/utils/suwayomi/schema.graphql',
  documents: ['server/domains/**/infrastructure/operations.ts'],
  ignoreNoDocuments: true,
  generates: {
    'server/utils/suwayomi/generated/': {
      preset: 'client',
      presetConfig: { fragmentMasking: false },
      config: { useTypeImports: true, scalars: { LongString: 'string' } },
    },
  },
}

export default config
