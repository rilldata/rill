import { test as base, type Page } from "@playwright/test";
import { ADMIN_AUTH_FILE, VIEWER_AUTH_FILE } from "./constants";
import { cliLogin, cliLogout } from "./fixtures/cli";
import { orgCreate, orgDelete } from "./fixtures/org";
import { projectDelete, projectDeploy } from "./fixtures/project";

type MyFixtures = {
  adminPage: Page;
  viewerPage: Page;
  anonPage: Page;
  cli: void;
  organization: void;
  project: void;
};

export const test = base.extend<MyFixtures>({
  adminPage: async ({ browser }, use) => {
    const context = await browser.newContext({
      storageState: ADMIN_AUTH_FILE,
    });
    const adminPage = await context.newPage();
    await use(adminPage);
    await context.close();
  },

  viewerPage: async ({ browser }, use) => {
    const context = await browser.newContext({
      storageState: VIEWER_AUTH_FILE,
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

  organization: async ({ cli: _ }, use) => {
    await orgCreate();
    await use();
    await orgDelete();
  },

  project: async ({ organization: _, page }, use) => {
    await projectDeploy(page);
    await use();
    await projectDelete();
  },
});
