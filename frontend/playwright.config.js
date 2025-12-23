// @ts-check
const { defineConfig, devices } = require('@playwright/test');

/**
 * Playwright configuration for SmartCalc E2E tests
 * Tests run against the Wails dev server which provides both frontend and backend
 */
module.exports = defineConfig({
  testDir: './e2e',
  fullyParallel: false, // Run tests sequentially to avoid conflicts with single app instance
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1, // Single worker since we have one app instance
  reporter: [['html', { open: 'never' }]], // Don't auto-open report
  timeout: 30000, // 30 second timeout per test
  
  use: {
    baseURL: 'http://localhost:34115', // Wails dev server port
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],

  // Wails dev server is started externally by test-app.sh
  // We don't use webServer here because wails dev needs to run from project root
});
