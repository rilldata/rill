import { devices, type PlaywrightTestConfig } from "@playwright/test";
import { ADMIN_AUTH_FILE, GITHUB_AUTH_FILE } from "./tests/setup/constants";

const config: PlaywrightTestConfig = {
  webServer: {
    command: "npm run build && npm run preview",
    port: 3000,
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
  retries: 0,
  reporter: "html",
  use: {
    baseURL: "http://localhost:3000",
    ...devices["Desktop Chrome"],
    trace: "on-first-retry",
    video: "retain-on-failure",
  },
  projects: [
    ...(process.env.CI
      ? [] // skip in CI
      : [
          {
            // Whenever the GitHub auth cookies expire, run this project manually to renew them.
            // Commit the resultant `playwright/.auth/github.json` file to the repo.
            name: "save-github-cookies",
            testMatch: "auth.github.ts",
          },
        ]),
    {
      name: "setup",
      testMatch: "setup.ts",
      ...(process.env.E2E_NO_TEARDOWN || process.env.CI
        ? undefined
        : { teardown: "teardown" }),
      use: {
        storageState: GITHUB_AUTH_FILE,
      },
    },
    ...(process.env.CI
      ? [] // skip in CI, since the GitHub Action uses an ephemeral runner
      : [
          {
            name: "teardown",
            testMatch: "teardown.ts",
            use: {
              storageState: ADMIN_AUTH_FILE,
            },
          },
        ]),
    {
      name: "e2e",
      dependencies: process.env.E2E_NO_SETUP_OR_TEARDOWN ? [] : ["setup"],
      testIgnore: "/setup",
      use: {
        storageState: ADMIN_AUTH_FILE,
      },
    },
  ],
};

export default config;
