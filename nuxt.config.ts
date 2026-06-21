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
    'nuxt-auth-utils',
    'nuxt-authorization',
  ],

  // Feature-colocated data-access layer (ADR-0013): auto-import the per-feature
  // composables (`features/<Feature>/composables/*.composable.ts`) so components
  // consume `useX()` without manual imports. Pinia stores are wired separately
  // via `pinia.storesDirs` below.
  imports: {
    dirs: ['features/**/composables'],
  },

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

    // Pin sealed-session cookie flags explicitly so a future nuxt-auth-utils or
    // h3 default change cannot silently downgrade them.
    // Shape: runtimeConfig.session (SessionConfig from h3) with a cookie sub-object
    // (CookieSerializeOptions from cookie-es). maxAge and password are siblings of
    // cookie, not nested inside it. Password is intentionally omitted here — it is
    // injected at runtime from NUXT_SESSION_PASSWORD by nuxt-auth-utils.
    // secure is false in dev/test (NODE_ENV !== 'production') so that the e2e suite,
    // which runs a plain-HTTP test server, still receives and threads the cookie.
    session: {
      cookie: {
        httpOnly: true,
        sameSite: 'lax' as const,
        path: '/',
        secure: process.env.NODE_ENV === 'production',
      },
    },

    // Server-only authentication policy constants. Domain methods receive these
    // as params (never read config themselves); routes read minPasswordLength for
    // zod validation. Override at runtime with NUXT_AUTH_* env vars.
    auth: {
      sessionTtlMs: 30 * 24 * 60 * 60 * 1000,
      sessionRefreshThresholdMs: 7 * 24 * 60 * 60 * 1000,
      minPasswordLength: 10,
      loginRateLimitMaxAttempts: 10,
      loginRateLimitWindowMs: 15 * 60 * 1000,
    },
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

  // Pinia auto-import for the per-feature stores
  // (`features/<Feature>/store/*.store.ts`), exposing `useXStore()` app-wide.
  pinia: {
    storesDirs: ['features/**/store'],
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
