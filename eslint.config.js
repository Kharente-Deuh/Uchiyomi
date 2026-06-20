// @ts-check
import antfu from '@antfu/eslint-config'
import nuxt from './.nuxt/eslint.config.mjs'

// Add your own options to the antfu() call below.
export default antfu(
  {
    ignores: ['prisma/generated'],
    vue: true,
    typescript: true,
    formatters: true,
    pnpm: true,
  },
)
  .append(nuxt())
