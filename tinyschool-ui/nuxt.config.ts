// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  modules: ['@nuxt/eslint', '@nuxt/ui'],

  // Browser-level API mocks only see client requests. Keep production SSR
  // unchanged and opt into SPA rendering only for the screenshot harness.
  ssr: import.meta.env.NUXT_PLAYWRIGHT_SPA !== 'true',

  devtools: {
    enabled: false
  },

  css: ['~/assets/css/main.css'],

  colorMode: {
    preference: 'system'
  },

  runtimeConfig: {
    public: {
      apiBase: 'http://localhost:8080/api/v1'
    }
  },

  compatibilityDate: '2026-06-30',

  eslint: {
    config: {
      stylistic: {
        commaDangle: 'never',
        braceStyle: '1tbs'
      }
    }
  },

  icon: {
    provider: 'none',
    clientBundle: {
      scan: true
    }
  }
})
