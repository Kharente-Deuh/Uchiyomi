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

  display: {
    mobileBreakpoint: 'sm',
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
  defaults: {
    VTextField: {
      density: 'comfortable',
      bgColor: 'surface-variant',
      class: 'text-field-override',
      clearIcon: 'fa6-solid:xmark',
      color: 'primary',
      variant: 'outlined',
    },
    VSelect: {
      clearIcon: 'fa6-solid:xmark',
      density: 'comfortable',
      bgColor: 'surface-variant',
      class: 'text-field-override',
      color: 'primary',
      variant: 'outlined',
      menuIcon: 'fa6-solid:caret-down',
    },
    VAutocomplete: {
      density: 'comfortable',
      bgColor: 'surface-variant',
      class: 'text-field-override',
      color: 'primary',
      clearIcon: 'fa6-solid:xmark',
      menuIcon: 'fa6-solid:caret-down',
      variant: 'outlined',
    },
    VTextarea: {
      clearIcon: 'fa6-solid:xmark',
      density: 'comfortable',
      bgColor: 'surface-variant',
      class: 'text-field-override',
      color: 'primary',
      variant: 'outlined',
      maxRows: 3,
      noResize: true,
    },
    VBtn: {
      variant: 'tonal',
      rounded: 'lg',
      color: 'primary',
      class: 'w-fit-content text-capitalize font-weight-bold',
    },
    VNumberInput: {
      density: 'comfortable',
      class: 'text-field-override',
      bgColor: 'surface-variant',
      clearIcon: 'fa6-solid:xmark',
      color: 'primary',
      variant: 'outlined',
      iconColor: 'primary',
      hideDetails: true,
      controlVariant: 'stacked',
      inset: true,
    },
    VDataTable: {
      sortDescIcon: 'fa6-solid:caret-down',
      sortAscIcon: 'fa6-solid:caret-up',
    },
    VIcon: {
      size: 'small',
    },
    VCheckbox: {
      color: 'primary',
      density: 'comfortable',
      trueIcon: 'fa6-solid:square-check',
      falseIcon: 'fa6-regular:square',
      indeterminateIcon: 'fa6-regular:square',
    },
    VChip: {
      closeIcon: 'fa6-solid:xmark',
    },
  },
})
