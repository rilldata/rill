import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./tests",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? [["github"], ["list"]] : [["list"]],
  timeout: 60_000,
  use: {
    trace: "retain-on-failure",
    viewport: { width: 800, height: 700 },
  },
  projects: [{ name: "chromium", use: { ...devices["Desktop Chrome"] } }],
});
