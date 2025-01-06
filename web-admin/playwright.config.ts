import { devices, type PlaywrightTestConfig } from "@playwright/test";
import dotenv from "dotenv";
import path from "path";
import { fileURLToPath } from "url";
import { ADMIN_AUTH_FILE } from "./tests/setup/constants";

// Load environment variables from our root `.env` file
const __dirname = path.dirname(fileURLToPath(import.meta.url));
dotenv.config({ path: path.resolve(__dirname, "../.env") });

const config: PlaywrightTestConfig = {
  ...(process.env.E2E_SKIP_GLOBAL_SETUP
    ? {}
    : {
        globalSetup: "./tests/setup/global.setup.ts",
        globalTeardown: "./tests/setup/global.teardown.ts",
      }),
  webServer: {
    command: "npm run build && npm run preview",
    port: 3000,
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
  retries: 0,
  /* Reporter to use. See https://playwright.dev/docs/test-reporters */
  reporter: "html",
  use: {
    baseURL: "http://localhost:3000",
    ...devices["Desktop Chrome"],
    /* Collect trace when retrying the failed test. See https://playwright.dev/docs/trace-viewer */
    trace: "on-first-retry",
    video: "retain-on-failure",
  },
  projects: [
    {
      name: "data-setup",
      testMatch: "data.setup.ts",
      ...(process.env.E2E_SKIP_DATA_TEARDOWN
        ? {}
        : { teardown: "data-teardown" }),
    },
    {
      name: "data-teardown",
      testMatch: "data.teardown.ts",
      use: {
        storageState: ADMIN_AUTH_FILE,
      },
    },
    {
      name: "e2e",
      dependencies: process.env.E2E_SKIP_DATA_SETUP ? [] : ["data-setup"],
      testIgnore: "/setup",
      use: {
        storageState: ADMIN_AUTH_FILE,
      },
    },
  ],
};

export default config;
