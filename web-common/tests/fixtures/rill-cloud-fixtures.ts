import { test as base, type Page } from "@playwright/test";
import {
  cliLogin,
  cliLogout,
} from "@rilldata/web-common/tests/fixtures/cli.ts";
import path from "path";
import {
  ADMIN_STORAGE_STATE,
  VIEWER_STORAGE_STATE,
  RILL_EMBED_SERVICE_TOKEN_FILE,
  RILL_ORG_NAME,
  RILL_PROJECT_NAME,
  RILL_EMBED_HTML_FILE,
} from "@rilldata/web-integration/tests/constants.ts";
import fs from "fs";
import { generateEmbed } from "@rilldata/web-common/tests/utils/generate-embed.ts";

type MyFixtures = {
  adminPage: Page;
  viewerPage: Page;
  anonPage: Page;
  cli: void;
  embedPage: Page;
  embeddedInitialState: string | null;
  /**
   * Resource to embed. Should be from the openrtb project.
   * Defaults to "bids_explore"
   */
  embeddedResourceName: string;
  embeddedResourceType: string;
};

export const rillCloud = base.extend<MyFixtures>({
  embeddedInitialState: [null, { option: true }],
  embeddedResourceName: ["bids_explore", { option: true }],
  embeddedResourceType: ["rill.runtime.v1.Explore", { option: true }],

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
    async (
      {
        browser,
        embeddedResourceName,
        embeddedResourceType,
        embeddedInitialState,
      },
      use,
    ) => {
      const readPath = path.join(process.cwd(), RILL_EMBED_SERVICE_TOKEN_FILE);
      const rillServiceToken = fs.readFileSync(readPath, "utf-8");

      await generateEmbed({
        organization: RILL_ORG_NAME,
        project: RILL_PROJECT_NAME,
        resourceName: embeddedResourceName,
        resourceType: embeddedResourceType,
        serviceToken: rillServiceToken,
        initialState: embeddedInitialState,
      });
      const filePath =
        "file://" + path.resolve(process.cwd(), RILL_EMBED_HTML_FILE);

      const context = await browser.newContext();
      const embedPage = await context.newPage();
      await embedPage.goto(filePath);

      await use(embedPage);

      await context.close();
    },
    { scope: "test" },
  ],
});
