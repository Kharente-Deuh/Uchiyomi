import process from 'node:process'
import { pwa } from './app/config/pwa'
import { appDescription } from './app/constants/index'

// SPDX-License-Identifier: AGPL-3.0-or-later
export default defineNuxtConfig({
  modules: [
    '@vueuse/nuxt',
    '@pinia/nuxt',
    'vuetify-nuxt-module',
    '@nuxtjs/i18n',
    '@vite-pwa/nuxt',
    '@nuxt/eslint',
  ],

  devtools: {
    enabled: true,
  },

  app: {
    head: {
      viewport: 'width=device-width,initial-scale=1',
      link: [
        { rel: 'icon', href: '/favicon.ico', sizes: 'any' },
        { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' },
        { rel: 'apple-touch-icon', href: '/apple-touch-icon.png' },
      ],
      meta: [
        { name: 'description', content: appDescription },
        { name: 'apple-mobile-web-app-status-bar-style', content: 'black-translucent' },
        { name: 'theme-color', media: '(prefers-color-scheme: light)', content: 'white' },
        { name: 'theme-color', media: '(prefers-color-scheme: dark)', content: '#222222' },
      ],
    },
  },

  css: [
    '~/assets/styles/reset.scss',
    '~/assets/styles/global.scss',
    '~/assets/styles/vuetify-overrides.scss',
  ],

  runtimeConfig: {
    // Server-only base URL of the headless Suwayomi engine (ADR-0001). Never
    // exposed to the client bundle. Override at runtime with NUXT_SUWAYOMI_URL.
    suwayomiUrl: process.env.SUWAYOMI_URL || 'http://localhost:4567',
  },

  future: {
    compatibilityVersion: 4,
  },

  experimental: {
    payloadExtraction: false,
    renderJsonPayloads: true,
    typedPages: true,
  },

  compatibilityDate: '2025-06-20',

  nitro: {
    esbuild: {
      options: {
        target: 'esnext',
      },
    },
    prerender: {
      crawlLinks: false,
      routes: ['/'],
    },
  },

  vite: {
    optimizeDeps: {
      include: [
        '@iconify/vue',
        '@vue/devtools-core',
        '@vue/devtools-kit',
      ],
    },
  },

  eslint: {
    config: {
      standalone: false,
      nuxt: {
        sortConfigKeys: true,
      },
    },
  },

  i18n: {
    vueI18n: 'i18n.config.ts',
    strategy: 'no_prefix',
    defaultLocale: 'en',
    detectBrowserLanguage: {
      useCookie: true,
      cookieKey: 'uchiyomi_lang',
      redirectOn: 'root',
    },
    locales: [
      { code: 'en', language: 'en-US', name: 'English', file: 'en.json' },
      { code: 'fr', language: 'fr-FR', name: 'Français', file: 'fr.json' },
    ],
  },

  pwa,

  vuetify: {
    moduleOptions: {
      ssrClientHints: {
        reloadOnFirstRequest: false,
        viewportSize: true,
        prefersColorScheme: true,
        prefersColorSchemeOptions: {
          useBrowserThemeOnly: false,
        },
      },
    },
    vuetifyOptions: './vuetify.config.ts',
  },
})
