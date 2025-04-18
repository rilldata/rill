import { defineConfig, devices } from "@playwright/test";

/**
 * See https://playwright.dev/docs/test-configuration.
 */
export default defineConfig({
  testDir: "./tests",
  /* Don't run tests in files in parallel in CI*/
  fullyParallel: !process.env.CI,
  /* Fail the build on CI if you accidentally left test.only in the source code. */
  forbidOnly: !!process.env.CI,
  retries: 0,
  /* Opt out of parallel testing in CI */
  workers: process.env.CI ? 1 : 8,
  /* Reporter to use. See https://playwright.dev/docs/test-reporters */
  reporter: "html",
  /* Shared settings for all the projects below. See https://playwright.dev/docs/api/class-testoptions. */
  use: {
    /* Base URL to use in actions like `await page.goto('/')`. */
    baseURL: "http://localhost:8083",
    /* Collect trace when tests fail. See https://playwright.dev/docs/trace-viewer */
    trace: "retain-on-failure",
    video: "retain-on-failure",
    launchOptions: {
      slowMo: parseInt(process.env.PLAYWRIGHT_SLOW_MO || "0"),
    },
  },
  /* Configure projects for major browsers */
  projects: [
    {
      name: "setup",
      testMatch: "setup.ts",
    },
    {
      name: "e2e-chrome",
      dependencies: ["setup"],
      use: { ...devices["Desktop Chrome"] },
    },
    {
      name: "e2e-safari",
      dependencies: ["setup"],
      use: { ...devices["Desktop Safari"] },
    },
    {
      name: "e2e-firefox",
      dependencies: ["setup"],
      use: { ...devices["Desktop Firefox"] },
    },
  ],
});
