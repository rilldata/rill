import { test as base, type Page } from "@playwright/test";

type MyFixtures = {
  anonPage: Page;
};

export const test = base.extend<MyFixtures>({
  // Fixtures
  anonPage: async ({ browser }, use) => {
    const context = await browser.newContext({
      storageState: { cookies: [], origins: [] },
    });
    const anonPage = await context.newPage();
    await use(anonPage);
    await context.close();
  },
});
