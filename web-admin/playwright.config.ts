import type { PlaywrightTestConfig } from "@playwright/test";

const config: PlaywrightTestConfig = {
  webServer: {
    command: "npm run build && npm run preview",
    port: 3000,
    timeout: 120_000,
  },
  use: {
    baseURL: "http://localhost:3000",
  },
};

export default config;
