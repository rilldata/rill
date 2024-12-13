import { test as base, type Page } from "@playwright/test";
import { cliLogin, cliLogout } from "./cli";

type MyFixtures = {
  anonPage: Page;
  cli: void;
};

export const test = base.extend<MyFixtures>({
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
});
