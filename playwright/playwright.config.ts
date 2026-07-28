import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: '.',
  testMatch: 'screenshots.spec.ts',
  fullyParallel: false,
  workers: 1,
  reporter: 'list',
  outputDir: 'output/test-results',
  use: {
    baseURL: process.env.PLAYWRIGHT_BASE_URL || 'http://127.0.0.1:3000',
    viewport: { width: 1440, height: 1000 },
    colorScheme: 'light'
  },
  webServer: process.env.PLAYWRIGHT_BASE_URL
    ? undefined
    : {
        command: 'NUXT_IGNORE_LOCK=1 NUXT_PLAYWRIGHT_SPA=true bun run dev --host 127.0.0.1 --port 3000',
        cwd: '../tinyschool-ui',
        url: 'http://127.0.0.1:3000',
        reuseExistingServer: false,
        timeout: 120_000
      }
})
