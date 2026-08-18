// SPDX-License-Identifier: AGPL-3.0-or-later

import { pwa } from './app/config/pwa'
import { APP_DESCRIPTION } from './app/constants/index'

export default defineNuxtConfig({
  modules: [
    '@vueuse/nuxt',
    '@pinia/nuxt',
    'vuetify-nuxt-module',
    '@nuxtjs/i18n',
    '@nuxt/eslint',
    '@vite-pwa/nuxt',
    '@nuxt/fonts',
    'nuxt-authorization',
  ],

  ssr: false,
  components: [
    { path: '~/components', pathPrefix: true },
    { path: '~/features/auth/components', prefix: 'Auth' },
    { path: '~/features/oidc/components', prefix: 'Oidc' },
    { path: '~/features/comics/components', prefix: 'Comics' },
    { path: '~/features/asura/components', prefix: 'Asura' },
    { path: '~/features/library/components', prefix: 'Library' },
    { path: '~/features/feed/components', prefix: 'Feed' },
  ],

  imports: {
    dirs: ['**/composables/**'],
  },

  devtools: {
    enabled: true,
  },

  app: {
    head: {
      viewport: 'width=device-width,initial-scale=1,viewport-fit=cover',
      link: [
        { rel: 'icon', href: '/favicon.ico', sizes: 'any' },
        { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' },
        { rel: 'apple-touch-icon', href: '/apple-touch-icon.png' },
      ],
      meta: [
        { name: 'description', content: APP_DESCRIPTION },
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
    '~/assets/styles/constants.scss',
  ],

  devServer: {
    port: 3001,
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
    devProxy: {
      '/api': 'http://localhost:3000/api',
    },
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
        'yup',
        'yup-locales',
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

  fonts: {
    defaults: {
      weights: [100, 300, 400, 500, 700, 900],
      styles: ['normal', 'italic'],
    },
    families: [
      { name: 'Inter', provider: 'google' },
      { name: 'Spectral', provider: 'google' },
    ],

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

  pinia: {
    storesDirs: ['features/**/stores', 'stores/**'],
  },

  pwa,

  vuetify: {
    moduleOptions: {
      prefixComposables: ['useLayout'],
    },
    vuetifyOptions: './vuetify.config.ts',
  },
})
