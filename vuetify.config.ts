// SPDX-License-Identifier: AGPL-3.0-or-later
import { defineVuetifyConfiguration } from 'vuetify-nuxt-module/custom-configuration'

/**
 * Emerald palette, translated from the design reference in
 * `docs/design/`. Vuetify's theme owns light/dark (ADR-0009).
 *
 * Surfaces stack background -> surface -> surface-bright -> surface-variant.
 * Primary is emerald; on-primary uses dark ink in dark mode (the accent is
 * light enough that dark text reads better) and white in light mode.
 */
export default defineVuetifyConfiguration({
  icons: {
    // Icons are configured at runtime via the `vuetify-icons` plugin
    // (Iconify custom set). 'custom' tells vuetify-nuxt-module to step back.
    defaultSet: 'custom',
  },
  theme: {
    defaultTheme: 'dark',
    themes: {
      light: {
        dark: false,
        colors: {
          'background': '#f5f6f8',
          'surface': '#ffffff',
          'surface-bright': '#ffffff',
          'surface-light': '#f3f5f8',
          'surface-variant': '#e9edf2',
          'on-background': '#1a1d23',
          'on-surface': '#1a1d23',
          'on-surface-variant': '#5a626e',
          'primary': '#0d9488',
          'primary-darken-1': '#0f766e',
          'on-primary': '#ffffff',
          'secondary': '#5a626e',
          'secondary-darken-1': '#475569',
          'error': '#d32f3a',
          'info': '#2f6fb3',
          'success': '#2ea043',
          'warning': '#b97e16',
        },
      },
      dark: {
        dark: true,
        colors: {
          'background': '#0e1014',
          'surface': '#151821',
          'surface-bright': '#1c212c',
          'surface-light': '#1c212c',
          'surface-variant': '#232936',
          'on-background': '#e8eaf0',
          'on-surface': '#e8eaf0',
          'on-surface-variant': '#9aa1ad',
          'primary': '#14b8a6',
          'primary-darken-1': '#0d9488',
          'on-primary': '#04211d',
          'secondary': '#64748b',
          'secondary-darken-1': '#475569',
          'error': '#e5484d',
          'info': '#4d96d9',
          'success': '#3fb950',
          'warning': '#d99a2b',
        },
      },
    },
  },
})
