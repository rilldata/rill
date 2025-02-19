import { test as base, type Page } from "@playwright/test";
import { ADMIN_STORAGE_STATE, VIEWER_STORAGE_STATE } from "./constants";
import { cliLogin, cliLogout } from "./fixtures/cli";
import path from "path";
import { fileURLToPath } from "url";
import {
  RILL_EMBED_SERVICE_TOKEN,
  RILL_ORG_NAME,
  RILL_PROJECT_NAME,
} from "./constants";
import fs from "fs";
import { generateEmbed } from "../utils/generate-embed";

type MyFixtures = {
  adminPage: Page;
  viewerPage: Page;
  anonPage: Page;
  cli: void;
  embedPage: Page;
};

export const test = base.extend<MyFixtures>({
  // Note: the `e2e` project uses the admin auth file by default, so it's likely that
  // this fixture won't be used often.
  adminPage: async ({ browser }, use) => {
    const context = await browser.newContext({
      storageState: ADMIN_STORAGE_STATE,
    });
    const adminPage = await context.newPage();
    await use(adminPage);
    await context.close();
  },

  viewerPage: async ({ browser }, use) => {
    const context = await browser.newContext({
      storageState: VIEWER_STORAGE_STATE,
    });
    const viewerPage = await context.newPage();
    await use(viewerPage);
    await context.close();
  },

  anonPage: async ({ browser }, use) => {
    const context = await browser.newContext({
      storageState: { cookies: [], origins: [] },
    });
    const anonPage = await context.newPage();
    await use(anonPage);
    await context.close();
  },

  cli: async ({ page }, use) => {
    await cliLogin(page);
    await use();
    await cliLogout();
  },

  embedPage: [
    async ({ browser }, use) => {
      const __dirname = path.dirname(fileURLToPath(import.meta.url));
      const readPath = path.join(
        __dirname,
        "..",
        "..",
        RILL_EMBED_SERVICE_TOKEN,
      );
      const rillServiceToken = fs.readFileSync(readPath, "utf-8");

      await generateEmbed(
        "bids_explore",
        rillServiceToken,
        RILL_ORG_NAME,
        RILL_PROJECT_NAME,
      );
      const filePath = "file://" + path.resolve(__dirname, "..", "embed.html");

      const context = await browser.newContext();
      const embedPage = await context.newPage();
      await embedPage.goto(filePath);

      await use(embedPage);

      await context.close();
    },
    { scope: "test" },
  ],
});
