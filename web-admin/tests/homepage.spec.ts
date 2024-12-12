import { expect } from "@playwright/test";
import { test } from "./setup/test";

test("Authenticated user sees the homepage", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByText("Hi qa@rilldata.com!")).toBeVisible();
});

test("Unauthenticated user gets redirected to login", async ({ browser }) => {
  const anonContext = await browser.newContext({
    storageState: { cookies: [], origins: [] },
  });
  const page = await anonContext.newPage();
  await page.goto("/");
  await expect(page.getByText("Log in to Rill")).toBeVisible();
});
