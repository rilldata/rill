import { devices, type PlaywrightTestConfig } from "@playwright/test";
import { ADMIN_STORAGE_STATE } from "./tests/setup/constants";

const config: PlaywrightTestConfig = {
  webServer: {
    command: "npm run build && npm run preview",
    port: 3000,
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
  /* Retry on CI so genuine infra noise is absorbed and surfaced as "flaky" rather than failing the run. */
  retries: process.env.CI ? 2 : 0,
  reporter: process.env.CI
    ? [["github"], ["blob"], ["html", { open: "never" }]]
    : [["html"]],
  use: {
    baseURL: "http://localhost:3000",
    ...devices["Desktop Chrome"],
    trace: "retain-on-failure",
    video: "retain-on-failure",
  },
  testDir: "tests",
  projects: [
    {
      name: "setup",
      testMatch: "setup.ts",
      ...(process.env.E2E_NO_TEARDOWN ? undefined : { teardown: "teardown" }),
    },
    {
      name: "teardown",
      testMatch: "teardown.ts",
      use: {
        storageState: ADMIN_STORAGE_STATE,
      },
    },
    {
      name: "e2e",
      dependencies: process.env.E2E_NO_SETUP_OR_TEARDOWN ? [] : ["setup"],
      testIgnore: "/setup",
      use: {
        storageState: ADMIN_STORAGE_STATE,
      },
    },
    {
      // Responsive smoke checks run on an emulated phone against the same
      // seeded stack. Limited to mobile-smoke so the desktop suites aren't
      // re-run here.
      name: "mobile",
      dependencies: process.env.E2E_NO_SETUP_OR_TEARDOWN ? [] : ["setup"],
      testMatch: "mobile-smoke.spec.ts",
      use: {
        ...devices["iPhone 13"],
        storageState: ADMIN_STORAGE_STATE,
      },
    },
  ],
};

export default config;
