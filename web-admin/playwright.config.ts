import { devices, type PlaywrightTestConfig } from "@playwright/test";

const config: PlaywrightTestConfig = {
  globalSetup: "./tests/setup/globalSetup.ts",
  globalTeardown: "./tests/setup/globalTeardown.ts",
  webServer: {
    command: "npm run build && npm run preview",
    port: 3000,
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
  use: {
    baseURL: "http://localhost:3000",
    ...devices["Desktop Chrome"],
  },
  projects: [
    { name: "auth", testMatch: "authenticate.spec.ts" },
    {
      name: "e2e",
      use: {
        storageState: "playwright/.auth/user.json",
      },
      dependencies: ["auth"],
    },
  ],
};

export default config;
