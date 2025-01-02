import { devices, type PlaywrightTestConfig } from "@playwright/test";
import dotenv from "dotenv";
import path from "path";
import { fileURLToPath } from "url";
import { ADMIN_AUTH_FILE } from "./tests/setup/constants";

// Load environment variables from our root `.env` file
const __dirname = path.dirname(fileURLToPath(import.meta.url));
dotenv.config({ path: path.resolve(__dirname, "../.env") });

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
        storageState: ADMIN_AUTH_FILE,
      },
      dependencies: ["auth"],
    },
  ],
};

export default config;
