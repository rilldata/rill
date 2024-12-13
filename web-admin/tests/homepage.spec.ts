import { expect } from "@playwright/test";
import { test } from "./setup/base";

test.describe("Homepage", () => {
  test("Authenticated user should see the homepage", async ({ page }) => {
    await page.goto("/");
    await expect(page.getByText("Hi qa@rilldata.com!")).toBeVisible();
  });

  test("Unauthenticated user should be redirected to login", async ({
    anonPage,
  }) => {
    await anonPage.goto("/");
    await expect(anonPage.getByText("Log in to Rill")).toBeVisible();
  });
});
