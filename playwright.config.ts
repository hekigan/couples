import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright Configuration for Couple Card Game E2E Tests
 *
 * Quick-start configuration for basic smoke tests:
 * - Single worker for database isolation
 * - Chromium only (for speed)
 * - Screenshots/videos on failure
 * - Automatic Go server startup
 */
export default defineConfig({
  testDir: './tests/e2e',

  // Run tests sequentially (important for DB isolation)
  fullyParallel: false,

  // Forbid test.only in CI (we don't have CI yet, but good practice)
  forbidOnly: !!process.env.CI,

  // No retries for quick start (add later if flaky)
  retries: 0,

  // Single worker to avoid DB conflicts
  workers: 1,

  // Test timeout (30 seconds should be enough for smoke tests)
  timeout: 30000,

  // Reporter configuration
  reporter: [
    ['html', { outputFolder: 'playwright-report' }],
    ['list'], // Terminal output
  ],

  // Shared settings for all tests
  use: {
    // Base URL for the application
    baseURL: 'http://localhost:8188',

    // Collect trace on first retry (helps debugging)
    trace: 'on-first-retry',

    // Screenshots on failure
    screenshot: 'only-on-failure',

    // Video on failure
    video: 'retain-on-failure',

    // Action timeout (5 seconds for individual actions)
    actionTimeout: 5000,

    // Navigation timeout (10 seconds)
    navigationTimeout: 10000,
  },

  // Browser configuration - Chromium only for quick start
  projects: [
    {
      name: 'chromium',
      use: {
        ...devices['Desktop Chrome'],
        // Viewport size
        viewport: { width: 1280, height: 720 },
      },
    },

    // Uncomment to add Firefox and WebKit later:
    // {
    //   name: 'firefox',
    //   use: { ...devices['Desktop Firefox'] },
    // },
    // {
    //   name: 'webkit',
    //   use: { ...devices['Desktop Safari'] },
    // },
  ],

  // Web server configuration - automatically start Go server before tests
  webServer: {
    command: 'make run',
    url: 'http://localhost:8188/health',
    reuseExistingServer: true, // Use existing server if already running
    timeout: 120 * 1000, // 2 minutes to start server
    stdout: 'pipe', // Show server output in terminal
    stderr: 'pipe',
  },
});
