// SPDX-License-Identifier: AGPL-3.0-or-later

import { defineVitestConfig } from '@nuxt/test-utils/config'

export default defineVitestConfig({
  test: {
    coverage: {
      reporter: ['text', 'json', 'json-summary'],
      include: ['app/**/*.{ts,vue}'],
    },
  },
})
